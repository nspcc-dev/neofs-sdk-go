package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

// DeleteObjectOptions groups optional parameters of [Client.DeleteObject].
type DeleteObjectOptions struct {
	sessionSet bool
	session    session.Object

	bearerTokenSet bool
	bearerToken    bearer.Token
}

// WithinSession specifies token of the session preliminary opened with the
// remote server. Session tokens grants user-to-user power of attorney: remote
// server still creates tombstone objects, but they are owned by the session
// issuer. Session must include [session.VerbObjectDelete] action. The token
// must be signed by the user passed to [Client.DeleteObject].
//
// With session, [Client.DeleteObject] can also return
// [apistatus.ErrSessionTokenNotFound] if the session is missing on the server
// or [apistatus.ErrSessionTokenExpired] if it has expired: this usually
// requires re-issuing the session.
//
// To start a session, use [Client.StartSession].
func (x *DeleteObjectOptions) WithinSession(s session.Object) {
	x.session, x.sessionSet = s, true
}

// WithBearerToken attaches bearer token carrying extended ACL rules that
// replace eACL of the object's container. The token must be issued by the
// container owner and target the subject authenticated by signer passed to
// [Client.DeleteObject]. In practice, bearer token makes sense only if it
// grants deletion rights to the subject.
func (x *DeleteObjectOptions) WithBearerToken(t bearer.Token) {
	x.bearerToken, x.bearerTokenSet = t, true
}

// DeleteObject sends request to remove the referenced object. If the request is
// accepted, a special marker called a tombstone is created by the remote server
// and placed in the container. The tombstone confirms the user's intent to
// delete the object, and is itself a system server-owned object in the
// container: DeleteObject returns its ID. The tombstone has limited lifetime
// depending on the server configuration. Explicit deletion is done
// asynchronously, and is generally not guaranteed. Created tombstone is owned
// by specified user.
//
// DeleteObject returns:
//   - [apistatus.ErrContainerNotFound] if referenced container is missing
//   - [apistatus.ErrObjectAccessDenied] if signer has no access to remove the object
//   - [apistatus.ErrObjectLocked] if referenced objects is locked (meaning
//     protection from the removal while the lock is active)
func (c *Client) DeleteObject(ctx context.Context, cnr cid.ID, obj oid.ID, signer neofscrypto.Signer, opts DeleteObjectOptions) (oid.ID, error) {
	var res oid.ID
	if signer == nil {
		return res, errMissingSigner
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodObjectDelete, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apiobject.DeleteRequest{
		Body: &apiobject.DeleteRequest_Body{
			Address: &refs.Address{
				ContainerId: new(refs.ContainerID),
				ObjectId:    new(refs.ObjectID),
			},
		},
		MetaHeader: &apisession.RequestMetaHeader{Ttl: 2},
	}
	cnr.WriteToV2(req.Body.Address.ContainerId)
	obj.WriteToV2(req.Body.Address.ObjectId)
	if opts.sessionSet {
		req.MetaHeader.SessionToken = new(apisession.SessionToken)
		opts.session.WriteToV2(req.MetaHeader.SessionToken)
	}
	if opts.bearerTokenSet {
		req.MetaHeader.BearerToken = new(apiacl.BearerToken)
		opts.bearerToken.WriteToV2(req.MetaHeader.BearerToken)
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.object.Delete(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return res, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return res, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return res, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return res, err
	}
	if sts != nil {
		err = sts // for closure above
		return res, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldTombstone = "tombstone"
	if resp.Body.Tombstone == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldTombstone) // for closure above
		return res, err
	} else if resp.Body.Tombstone.ObjectId == nil {
		err = fmt.Errorf("%s (%s): missing ID field", errInvalidResponseBodyField, fieldTombstone) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.Tombstone.ObjectId); err != nil {
		err = fmt.Errorf("%s (%s): invalid ID: %w", errInvalidResponseBodyField, fieldTombstone, err) // for closure above
		return res, err
	}
	return res, nil
}
