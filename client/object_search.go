package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

const fieldObjectIDList = "object ID list"

// SelectObjectsOptions groups optional parameters of [Client.SelectObjects].
type SelectObjectsOptions struct {
	local bool

	sessionSet bool
	session    session.Object

	bearerTokenSet bool
	bearerToken    bearer.Token
}

// PreventForwarding disables request forwarding to container nodes and
// instructs the server to select objects from the local storage only.
func (x *SelectObjectsOptions) PreventForwarding() {
	x.local = true
}

// WithinSession specifies token of the session preliminary issued by some user
// with the client signer. Session must include [session.VerbObjectSearch]
// action. The token must be signed and target the subject authenticated by
// signer passed to [Client.SelectObjects]. If set, the session issuer will be
// treated as the original request sender.
//
// Note that sessions affect access control only indirectly: they just replace
// request originator.
//
// With session, [Client.SelectObjects] can also return
// [apistatus.ErrSessionTokenExpired] if the token has expired: this usually
// requires re-issuing the session.
//
// Note that it makes no sense to start session with the server via
// [Client.StartSession] like for [Client.DeleteObject] or [Client.PutObject].
func (x *SelectObjectsOptions) WithinSession(s session.Object) {
	x.session, x.sessionSet = s, true
}

// WithBearerToken attaches bearer token carrying extended ACL rules that
// replace eACL of the container. The token must be issued by the container
// owner and target the subject authenticated by signer passed to
// [Client.SelectObjects]. In practice, bearer token makes sense only if it
// grants selecting rights to the subject.
func (x *SelectObjectsOptions) WithBearerToken(t bearer.Token) {
	x.bearerToken, x.bearerTokenSet = t, true
}

// AllObjectsQuery returns search query to select all objects in particular
// container.
func AllObjectsQuery() []object.SearchFilter { return nil }

var errBreak = errors.New("break")

// allows to share code b/w various methods calling object search. ID list
// passed to f is always non-empty. If f returns errBreak, this method breaks
// with no error.
func (c *Client) forEachSelectedObjectsSet(ctx context.Context, cnr cid.ID, signer neofscrypto.Signer, opts SelectObjectsOptions,
	filters []object.SearchFilter, f func(nResp int, ids []*refs.ObjectID) error) (err error) {
	if signer == nil {
		return errMissingSigner
	}

	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodObjectSearch, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apiobject.SearchRequest{
		Body: &apiobject.SearchRequest_Body{
			ContainerId: new(refs.ContainerID),
		},
		MetaHeader: new(apisession.RequestMetaHeader),
	}
	cnr.WriteToV2(req.Body.ContainerId)
	if len(filters) > 0 {
		req.Body.Filters = make([]*apiobject.SearchRequest_Body_Filter, len(filters))
		for i := range filters {
			req.Body.Filters[i] = new(apiobject.SearchRequest_Body_Filter)
			filters[i].WriteToV2(req.Body.Filters[i])
		}
	}
	if opts.sessionSet {
		req.MetaHeader.SessionToken = new(apisession.SessionToken)
		opts.session.WriteToV2(req.MetaHeader.SessionToken)
	}
	if opts.bearerTokenSet {
		req.MetaHeader.BearerToken = new(apiacl.BearerToken)
		opts.bearerToken.WriteToV2(req.MetaHeader.BearerToken)
	}
	if opts.local {
		req.MetaHeader.Ttl = 1
	} else {
		req.MetaHeader.Ttl = 2
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(signer, req, req.Body, *buf); err != nil {
		return fmt.Errorf("%s: %w", errSignRequest, err)
	}

	// send request
	ctx, cancelStream := context.WithCancel(ctx)
	defer cancelStream()
	stream, err := c.transport.object.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", errTransport, err)
	}

	// read the stream
	var resp apiobject.SearchResponse
	var lastStatus apistatus.StatusV2
	mustFin := false
	for n := 0; ; n++ {
		err = stream.RecvMsg(&resp)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if n > 0 { // at least 1 message carrying status is required
					return lastStatus
				}
				return errors.New("stream ended without a status response")
			}
			return fmt.Errorf("%s while reading response #%d: %w", errTransport, n, err)
		} else if mustFin {
			return fmt.Errorf("stream is not completed after the message #%d which must be the last one", n-1)
		}
		// intercept response info
		if c.interceptAPIRespInfo != nil && n == 0 {
			if err = c.interceptAPIRespInfo(ResponseMetaInfo{
				key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
				epoch: resp.GetMetaHeader().GetEpoch(),
			}); err != nil {
				return fmt.Errorf("%s: %w", errInterceptResponseInfo, err)
			}
		}
		if err = neofscrypto.VerifyResponse(&resp, resp.Body); err != nil {
			return fmt.Errorf("invalid response #%d: %s: %w", n, errResponseSignature, err)
		}
		lastStatus, err = apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
		if err != nil {
			return fmt.Errorf("invalid response #%d: %s: %w", n, errInvalidResponseStatus, err)
		} else if lastStatus != nil {
			mustFin = true
			continue
		}
		if resp.Body == nil || len(resp.Body.IdList) == 0 {
			if n == 0 {
				mustFin = true
				continue
			}
			// technically, we can continue. But if the server is malicious/buggy, it may
			// return zillion of such messages, and the only thing that could save us is the
			// context. So, it's safer to fail immediately.
			return fmt.Errorf("invalid response #%d: empty %s is only allowed in the first stream message", n, fieldObjectIDList)
		}
		if err = f(n, resp.Body.IdList); err != nil {
			if errors.Is(err, errBreak) {
				return nil
			}
			return err
		}
	}
}

// SelectObjects selects objects from the referenced container that match all
// specified search filters and returns their IDs. In particular, the empty set
// of filters matches all container objects ([AllObjectsQuery] may be used for
// this to make code clearer). if no matching objects are found, SelectObjects
// returns an empty result without an error. SelectObjects returns buffered
// objects regardless of error so that the caller can process the partial result
// if needed.
//
// SelectObjects returns:
//   - [apistatus.ErrContainerNotFound] if referenced container is missing
//   - [apistatus.ErrObjectAccessDenied] if signer has no access to select objects
//
// The method places the identifiers of all selected objects in a memory buffer,
// which can be quite large for some containers/queries. If full buffering is
// not required, [Client.ForEachSelectedObject] may be used to increase resource
// efficiency.
func (c *Client) SelectObjects(ctx context.Context, cnr cid.ID, signer neofscrypto.Signer, opts SelectObjectsOptions, filters []object.SearchFilter) ([]oid.ID, error) {
	var res []oid.ID
	return res, c.forEachSelectedObjectsSet(ctx, cnr, signer, opts, filters, func(nResp int, ids []*refs.ObjectID) error {
		off := len(res)
		res = append(res, make([]oid.ID, len(ids))...)
		for i := range ids {
			if ids[i] == nil {
				return fmt.Errorf("invalid respone #%d: invalid body: invalid field (%s): nil element #%d", nResp, fieldObjectIDList, i)
			} else if err := res[off+i].ReadFromV2(ids[i]); err != nil {
				res = res[:off+i]
				return fmt.Errorf("invalid response #%d: invalid body: invalid field (%s): invalid element #%d: %w", nResp, fieldObjectIDList, i, err)
			}
		}
		return nil
	})
}

// ForEachSelectedObject works like [Client.SelectObjects] but passes each
// select object's ID to f. If f returns false, ForEachSelectedObject breaks
// without an error. ForEachSelectedObject, like [Client.SelectObjects], returns
// no error if no matching objects are found (this case can be detected by the
// caller via f closure).
func (c *Client) ForEachSelectedObject(ctx context.Context, cnr cid.ID, signer neofscrypto.Signer, opts SelectObjectsOptions,
	filters []object.SearchFilter, f func(oid.ID) bool) error {
	return c.forEachSelectedObjectsSet(ctx, cnr, signer, opts, filters, func(nResp int, ids []*refs.ObjectID) error {
		var id oid.ID
		for i := range ids {
			if ids[i] == nil {
				return fmt.Errorf("invalid respone #%d: invalid body: invalid field (%s): nil element #%d", nResp, fieldObjectIDList, i)
			} else if err := id.ReadFromV2(ids[i]); err != nil {
				return fmt.Errorf("invalid response #%d: invalid body: invalid field (%s): invalid element #%d: %w", nResp, fieldObjectIDList, i, err)
			} else if !f(id) {
				return errBreak
			}
		}
		return nil
	})
}
