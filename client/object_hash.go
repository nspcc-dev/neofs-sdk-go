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
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

// HashObjectPayloadRangesOptions groups optional parameters of
// [Client.HashObjectPayloadRanges].
type HashObjectPayloadRangesOptions struct {
	local bool

	sessionSet bool
	session    session.Object

	bearerTokenSet bool
	bearerToken    bearer.Token

	salt []byte
}

// PreventForwarding disables request forwarding to container nodes and
// instructs the server to hash object payload stored locally.
func (x *HashObjectPayloadRangesOptions) PreventForwarding() {
	x.local = true
}

// WithinSession specifies token of the session preliminary issued by some user
// with the client signer. Session must include [session.VerbObjectRangeHash]
// action. The token must be signed and target the subject authenticated by
// signer passed to [Client.HashObjectPayloadRanges]. If set, the session issuer
// will be treated as the original request sender.
//
// Note that sessions affect access control only indirectly: they just replace
// request originator.
//
// With session, [Client.HashObjectPayloadRanges] can also return
// [apistatus.ErrSessionTokenExpired] if the token has expired: this usually
// requires re-issuing the session.
//
// Note that it makes no sense to start session with the server via
// [Client.StartSession] like for [Client.DeleteObject] or [Client.PutObject].
func (x *HashObjectPayloadRangesOptions) WithinSession(s session.Object) {
	x.session, x.sessionSet = s, true
}

// WithBearerToken attaches bearer token carrying extended ACL rules that
// replace eACL of the object's container. The token must be issued by the
// container owner and target the subject authenticated by signer passed to
// [Client.HashObjectPayloadRanges]. In practice, bearer token makes sense only
// if it grants hashing rights to the subject.
func (x *HashObjectPayloadRangesOptions) WithBearerToken(t bearer.Token) {
	x.bearerToken, x.bearerTokenSet = t, true
}

// WithSalt attaches salt to XOR the object's payload range before hashing.
func (x *HashObjectPayloadRangesOptions) WithSalt(salt []byte) {
	x.salt = salt
}

// HashObjectPayloadRanges requests checksum of the referenced object's payload
// ranges. Checksum type must not be zero, range set must not be empty and
// contain zero-length element. Returns a list of checksums in raw form: the
// format of hashes and their number is left for the caller to check. Client
// preserves the order of the server's response.
//
// When only object payload's checksums are needed, HashObjectPayloadRanges
// should be used instead of hashing the [Client.GetObjectPayloadRange] or
// [Client.GetObject] result as much more efficient.
//
// HashObjectPayloadRanges returns:
//   - [apistatus.ErrContainerNotFound] if referenced container is missing
//   - [apistatus.ErrObjectNotFound] if referenced object is missing
//   - [apistatus.ErrObjectAccessDenied] if signer has no access to hash the payload
//   - [apistatus.ErrObjectOutOfRange] if at least one range is out of bounds
func (c *Client) HashObjectPayloadRanges(ctx context.Context, cnr cid.ID, obj oid.ID, typ checksum.Type, signer neofscrypto.Signer,
	opts HashObjectPayloadRangesOptions, ranges []object.Range) ([][]byte, error) {
	if signer == nil {
		return nil, errMissingSigner
	} else if typ == 0 {
		return nil, errors.New("zero checksum type")
	} else if len(ranges) == 0 {
		return nil, errors.New("missing ranges")
	}
	for i := range ranges {
		if ranges[i].Length == 0 {
			return nil, fmt.Errorf("zero length of range #%d", i)
		}
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodObjectHash, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apiobject.GetRangeHashRequest{
		Body: &apiobject.GetRangeHashRequest_Body{
			Address: &refs.Address{
				ContainerId: new(refs.ContainerID),
				ObjectId:    new(refs.ObjectID),
			},
			Ranges: make([]*apiobject.Range, len(ranges)),
			Salt:   opts.salt,
			Type:   refs.ChecksumType(typ),
		},
		MetaHeader: new(apisession.RequestMetaHeader),
	}
	cnr.WriteToV2(req.Body.Address.ContainerId)
	obj.WriteToV2(req.Body.Address.ObjectId)
	for i := range ranges {
		req.Body.Ranges[i] = &apiobject.Range{Offset: ranges[i].Offset, Length: ranges[i].Length}
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
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return nil, err
	}

	// send request
	resp, err := c.transport.object.GetRangeHash(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return nil, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return nil, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return nil, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return nil, err
	}
	if sts != nil {
		err = sts // for closure above
		return nil, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return nil, err
	}
	return resp.Body.HashList, nil
}
