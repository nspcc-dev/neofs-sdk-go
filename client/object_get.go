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
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

//
// // shared parameters of GET/HEAD/RANGE.
// type prmObjectRead struct {
// 	sessionContainer
//
// 	raw bool
// }
//
// // WithXHeaders specifies list of extended headers (string key-value pairs)
// // to be attached to the request. Must have an even length.
// //
// // Slice must not be mutated until the operation completes.
// func (x *prmObjectRead) WithXHeaders(hs ...string) {
// 	writeXHeadersToMeta(hs, &x.meta)
// }
//
// // MarkRaw marks an intent to read physically stored object.
// func (x *prmObjectRead) MarkRaw() {
// 	x.raw = true
// }
//
// // MarkLocal tells the server to execute the operation locally.
// func (x *prmObjectRead) MarkLocal() {
// 	x.meta.SetTTL(1)
// }
//
// // WithBearerToken attaches bearer token to be used for the operation.
// //
// // If set, underlying eACL rules will be used in access control.
// //
// // Must be signed.
// func (x *prmObjectRead) WithBearerToken(t bearer.Token) {
// 	var v2token acl.BearerToken
// 	t.WriteToV2(&v2token)
// 	x.meta.SetBearerToken(&v2token)
// }
//
// // PrmObjectGet groups optional parameters of ObjectGetInit operation.
// type PrmObjectGet struct {
// 	prmObjectRead
// }
//
// // PayloadReader is a data stream of the particular NeoFS object. Implements
// // [io.ReadCloser].
// //
// // Must be initialized using Client.ObjectGetInit, any other
// // usage is unsafe.
// type PayloadReader struct {
// 	cancelCtxStream context.CancelFunc
//
// 	client *Client
// 	stream interface {
// 		Read(resp *v2object.GetResponse) error
// 	}
//
// 	err error
//
// 	tailPayload []byte
//
// 	remainingPayloadLen int
//
// 	statisticCallback shortStatisticCallback
// }
//
// // readHeader reads header of the object. Result means success.
// // Failure reason can be received via Close.
// func (x *PayloadReader) readHeader(dst *object.Object) bool {
// 	var resp v2object.GetResponse
// 	x.err = x.stream.Read(&resp)
// 	if x.err != nil {
// 		return false
// 	}
//
// 	x.err = x.client.processResponse(&resp)
// 	if x.err != nil {
// 		return false
// 	}
//
// 	var partInit *v2object.GetObjectPartInit
//
// 	switch v := resp.GetBody().GetObjectPart().(type) {
// 	default:
// 		x.err = fmt.Errorf("unexpected message instead of heading part: %T", v)
// 		return false
// 	case *v2object.SplitInfo:
// 		x.err = object.NewSplitInfoError(object.NewSplitInfoFromV2(v))
// 		return false
// 	case *v2object.GetObjectPartInit:
// 		partInit = v
// 	}
//
// 	var objv2 v2object.Object
//
// 	objv2.SetObjectID(partInit.GetObjectID())
// 	objv2.SetHeader(partInit.GetHeader())
// 	objv2.SetSignature(partInit.GetSignature())
//
// 	x.remainingPayloadLen = int(objv2.GetHeader().GetPayloadLength())
//
// 	*dst = *object.NewFromV2(&objv2) // need smth better
//
// 	return true
// }
//
// func (x *PayloadReader) readChunk(buf []byte) (int, bool) {
// 	var read int
//
// 	// read remaining tail
// 	read = copy(buf, x.tailPayload)
//
// 	x.tailPayload = x.tailPayload[read:]
//
// 	if len(buf) == read {
// 		return read, true
// 	}
//
// 	var chunk []byte
// 	var lastRead int
//
// 	for {
// 		var resp v2object.GetResponse
// 		x.err = x.stream.Read(&resp)
// 		if x.err != nil {
// 			return read, false
// 		}
//
// 		x.err = x.client.processResponse(&resp)
// 		if x.err != nil {
// 			return read, false
// 		}
//
// 		part := resp.GetBody().GetObjectPart()
// 		partChunk, ok := part.(*v2object.GetObjectPartChunk)
// 		if !ok {
// 			x.err = fmt.Errorf("unexpected message instead of chunk part: %T", part)
// 			return read, false
// 		}
//
// 		// read new chunk
// 		chunk = partChunk.GetChunk()
// 		if len(chunk) == 0 {
// 			// just skip empty chunks since they are not prohibited by protocol
// 			continue
// 		}
//
// 		lastRead = copy(buf[read:], chunk)
//
// 		read += lastRead
//
// 		if read == len(buf) {
// 			// save the tail
// 			x.tailPayload = append(x.tailPayload, chunk[lastRead:]...)
//
// 			return read, true
// 		}
// 	}
// }
//
// func (x *PayloadReader) close(ignoreEOF bool) error {
// 	defer x.cancelCtxStream()
//
// 	if errors.Is(x.err, io.EOF) {
// 		if ignoreEOF {
// 			return nil
// 		}
// 		if x.remainingPayloadLen > 0 {
// 			return io.ErrUnexpectedEOF
// 		}
// 	}
// 	return x.err
// }
//
// // Close ends reading the object payload. Must be called after using the
// // PayloadReader.
// func (x *PayloadReader) Close() error {
// 	var err error
// 	if x.statisticCallback != nil {
// 		defer func() {
// 			x.statisticCallback(err)
// 		}()
// 	}
// 	err = x.close(true)
// 	return err
// }
//
// // Read implements io.Reader of the object payload.
// func (x *PayloadReader) Read(p []byte) (int, error) {
// 	n, ok := x.readChunk(p)
//
// 	x.remainingPayloadLen -= n
//
// 	if !ok {
// 		if err := x.close(false); err != nil {
// 			return n, err
// 		}
//
// 		return n, x.err
// 	}
//
// 	if x.remainingPayloadLen < 0 {
// 		return n, errors.New("payload size overflow")
// 	}
//
// 	return n, nil
// }
//
// // ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
// // Returns header of the requested object and stream of its payload separately.
// //
// // Exactly one return value is non-nil. Resulting PayloadReader must be finally closed.
// //
// // Context is required and must not be nil. It is used for network communication.
// //
// // Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// // the specified Signer, which is taken into account, in particular, for access control.
// //
// // Return errors:
// //   - global (see Client docs)
// //   - [ErrMissingSigner]
// //   - *[object.SplitInfoError] (returned on virtual objects with PrmObjectGet.MakeRaw)
// //   - [apistatus.ErrContainerNotFound]
// //   - [apistatus.ErrObjectNotFound]
// //   - [apistatus.ErrObjectAccessDenied]
// //   - [apistatus.ErrObjectAlreadyRemoved]
// //   - [apistatus.ErrSessionTokenExpired]
// func (c *Client) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectGet) (object.Object, *PayloadReader, error) {
// 	var (
// 		addr  v2refs.Address
// 		cidV2 v2refs.ContainerID
// 		oidV2 v2refs.ObjectID
// 		body  v2object.GetRequestBody
// 		hdr   object.Object
// 		err   error
// 	)
//
// 	defer func() {
// 		c.sendStatistic(stat.MethodObjectGet, err)()
// 	}()
//
// 	if signer == nil {
// 		return hdr, nil, ErrMissingSigner
// 	}
//
// 	containerID.WriteToV2(&cidV2)
// 	addr.SetContainerID(&cidV2)
//
// 	objectID.WriteToV2(&oidV2)
// 	addr.SetObjectID(&oidV2)
//
// 	body.SetRaw(prm.raw)
// 	body.SetAddress(&addr)
//
// 	// form request
// 	var req v2object.GetRequest
//
// 	req.SetBody(&body)
// 	c.prepareRequest(&req, &prm.meta)
// 	buf := c.buffers.Get().(*[]byte)
// 	err = signServiceMessage(signer, &req, *buf)
// 	c.buffers.Put(buf)
// 	if err != nil {
// 		err = fmt.Errorf("sign request: %w", err)
// 		return hdr, nil, err
// 	}
//
// 	ctx, cancel := context.WithCancel(ctx)
//
// 	stream, err := rpcAPIGetObject(&c.c, &req, client.WithContext(ctx))
// 	if err != nil {
// 		cancel()
// 		err = fmt.Errorf("open stream: %w", err)
// 		return hdr, nil, err
// 	}
//
// 	var r PayloadReader
// 	r.cancelCtxStream = cancel
// 	r.stream = stream
// 	r.client = c
// 	r.statisticCallback = func(err error) {
// 		c.sendStatistic(stat.MethodObjectGetStream, err)
// 	}
//
// 	if !r.readHeader(&hdr) {
// 		err = fmt.Errorf("header: %w", r.Close())
// 		return hdr, nil, err
// 	}
//
// 	return hdr, &r, nil
// }
//

// GetObjectHeaderOptions groups optional parameters of
// [Client.GetObjectHeader].
type GetObjectHeaderOptions struct {
	local bool
	raw   bool

	sessionSet bool
	session    session.Object

	bearerTokenSet bool
	bearerToken    bearer.Token
}

// PreventForwarding disables request forwarding to container nodes and
// instructs the server to read object header from the local storage.
func (x *GetObjectHeaderOptions) PreventForwarding() {
	x.local = true
}

// PreventAssembly disables assembly of object's header if the object is stored
// as split-chain of smaller objects. If PreventAssembly is given and requested
// object is actually split, [Client.GetObjectHeader] will return
// [object.SplitInfoError] carrying [object.SplitInfo] at least with last or
// linker split-chain part. For atomic objects option is no-op.
//
// PreventAssembly allows to optimize object assembly utilities and is unlikely
// needed when working with objects regularly.
func (x *GetObjectHeaderOptions) PreventAssembly() {
	x.raw = true
}

// WithinSession specifies token of the session preliminary issued by some user
// with the client signer. Session must include [session.VerbObjectHead] action.
// The token must be signed and target the subject authenticated by signer
// passed to [Client.GetObjectHeader]. If set, the session issuer will
// be treated as the original request sender.
//
// Note that sessions affect access control only indirectly: they just replace
// request originator.
//
// With session, [Client.GetObjectHeader] can also return
// [apistatus.ErrSessionTokenExpired] if the token has expired: this usually
// requires re-issuing the session.
//
// Note that it makes no sense to start session with the server via
// [Client.StartSession] like for [Client.DeleteObject] or [Client.PutObject].
func (x *GetObjectHeaderOptions) WithinSession(s session.Object) {
	x.session, x.sessionSet = s, true
}

// WithBearerToken attaches bearer token carrying extended ACL rules that
// replace eACL of the object's container. The token must be issued by the
// container owner and target the subject authenticated by signer passed to
// [Client.GetObjectHeader]. In practice, bearer token makes sense only if it
// grants "heading" rights to the subject.
func (x *GetObjectHeaderOptions) WithBearerToken(t bearer.Token) {
	x.bearerToken, x.bearerTokenSet = t, true
}

// GetObjectHeader requests header of the referenced object. When object's
// payload is not needed, GetObjectHeader should be used instead of
// [Client.GetObject] as much more efficient.
//
// GetObjectHeader returns:
//   - [apistatus.ErrContainerNotFound] if referenced container is missing
//   - [apistatus.ErrObjectNotFound] if the object is missing
//   - [apistatus.ErrObjectAccessDenied] if signer has no access to hash the payload
//   - [apistatus.ErrObjectAlreadyRemoved] if the object has already been removed
func (c *Client) GetObjectHeader(ctx context.Context, cnr cid.ID, obj oid.ID, signer neofscrypto.Signer, opts GetObjectHeaderOptions) (object.Header, error) {
	var res object.Header
	if signer == nil {
		return res, errMissingSigner
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodObjectHead, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apiobject.HeadRequest{
		Body: &apiobject.HeadRequest_Body{
			Address: &refs.Address{
				ContainerId: new(refs.ContainerID),
				ObjectId:    new(refs.ObjectID),
			},
			Raw: opts.raw,
		},
		MetaHeader: new(apisession.RequestMetaHeader),
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
		return res, err
	}

	// send request
	resp, err := c.transport.object.Head(ctx, req)
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
	switch f := resp.Body.Head.(type) {
	default:
		err = fmt.Errorf("%s: unknown/invalid oneof field (%T)", errInvalidResponseBodyField, f) // for closure above
		return res, err
	case *apiobject.HeadResponse_Body_Header:
		const fieldHeader = "header"
		if f == nil || f.Header == nil || f.Header.Header == nil {
			err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldHeader) // for closure above
			return res, err
		} else if err = res.ReadFromV2(f.Header.Header); err != nil {
			err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldHeader, err) // for closure above
			return res, err
		}
		return res, nil
	case *apiobject.HeadResponse_Body_SplitInfo:
		if !opts.raw {
			err = fmt.Errorf("%s: server responded with split info which was not requested", errInvalidResponseBody) // for closure above
			return res, err
		}
		const fieldSplitInfo = "split info"
		if f == nil || f.SplitInfo == nil {
			err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldSplitInfo) // for closure above
			return res, err
		}
		var splitInfo object.SplitInfo
		if err = splitInfo.ReadFromV2(f.SplitInfo); err != nil {
			err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldSplitInfo, err) // for closure above
			return res, err
		}
		err = object.SplitInfoError(splitInfo) // for closure above
		return res, err
	}
}

//
// // PrmObjectRange groups optional parameters of ObjectRange operation.
// type PrmObjectRange struct {
// 	prmObjectRead
// }
//
// // ObjectRangeReader is designed to read payload range of one object
// // from NeoFS system. Implements [io.ReadCloser].
// //
// // Must be initialized using Client.ObjectRangeInit, any other
// // usage is unsafe.
// type ObjectRangeReader struct {
// 	cancelCtxStream context.CancelFunc
//
// 	client *Client
//
// 	err error
//
// 	stream interface {
// 		Read(resp *v2object.GetRangeResponse) error
// 	}
//
// 	tailPayload []byte
//
// 	remainingPayloadLen int
//
// 	statisticCallback shortStatisticCallback
// }
//
// func (x *ObjectRangeReader) readChunk(buf []byte) (int, bool) {
// 	var read int
//
// 	// read remaining tail
// 	read = copy(buf, x.tailPayload)
//
// 	x.tailPayload = x.tailPayload[read:]
//
// 	if len(buf) == read {
// 		return read, true
// 	}
//
// 	var partChunk *v2object.GetRangePartChunk
// 	var chunk []byte
// 	var lastRead int
//
// 	for {
// 		var resp v2object.GetRangeResponse
// 		x.err = x.stream.Read(&resp)
// 		if x.err != nil {
// 			return read, false
// 		}
//
// 		x.err = x.client.processResponse(&resp)
// 		if x.err != nil {
// 			return read, false
// 		}
//
// 		// get chunk message
// 		switch v := resp.GetBody().GetRangePart().(type) {
// 		default:
// 			x.err = fmt.Errorf("unexpected message received: %T", v)
// 			return read, false
// 		case *v2object.SplitInfo:
// 			x.err = object.NewSplitInfoError(object.NewSplitInfoFromV2(v))
// 			return read, false
// 		case *v2object.GetRangePartChunk:
// 			partChunk = v
// 		}
//
// 		chunk = partChunk.GetChunk()
// 		if len(chunk) == 0 {
// 			// just skip empty chunks since they are not prohibited by protocol
// 			continue
// 		}
//
// 		lastRead = copy(buf[read:], chunk)
//
// 		read += lastRead
//
// 		if read == len(buf) {
// 			// save the tail
// 			x.tailPayload = append(x.tailPayload, chunk[lastRead:]...)
//
// 			return read, true
// 		}
// 	}
// }
//
// func (x *ObjectRangeReader) close(ignoreEOF bool) error {
// 	defer x.cancelCtxStream()
//
// 	if errors.Is(x.err, io.EOF) {
// 		if ignoreEOF {
// 			return nil
// 		}
// 		if x.remainingPayloadLen > 0 {
// 			return io.ErrUnexpectedEOF
// 		}
// 	}
// 	return x.err
// }
//
// // Close ends reading the payload range and returns the result of the operation
// // along with the final results. Must be called after using the ObjectRangeReader.
// //
// // Any client's internal or transport errors are returned as Go built-in error.
// // If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// // codes are returned as error.
// //
// // Return errors:
// //   - global (see Client docs)
// //   - *[object.SplitInfoError] (returned on virtual objects with PrmObjectRange.MakeRaw)
// //   - [apistatus.ErrContainerNotFound]
// //   - [apistatus.ErrObjectNotFound]
// //   - [apistatus.ErrObjectAccessDenied]
// //   - [apistatus.ErrObjectAlreadyRemoved]
// //   - [apistatus.ErrObjectOutOfRange]
// //   - [apistatus.ErrSessionTokenExpired]
// func (x *ObjectRangeReader) Close() error {
// 	var err error
// 	if x.statisticCallback != nil {
// 		defer func() {
// 			x.statisticCallback(err)
// 		}()
// 	}
// 	err = x.close(true)
// 	return err
// }
//
// // Read implements io.Reader of the object payload.
// func (x *ObjectRangeReader) Read(p []byte) (int, error) {
// 	n, ok := x.readChunk(p)
//
// 	x.remainingPayloadLen -= n
//
// 	if !ok {
// 		err := x.close(false)
// 		if err != nil {
// 			return n, err
// 		}
//
// 		return n, x.err
// 	}
//
// 	if x.remainingPayloadLen < 0 {
// 		return n, errors.New("payload range size overflow")
// 	}
//
// 	return n, nil
// }
//
// // ObjectRangeInit initiates reading an object's payload range through a remote
// // server using NeoFS API protocol.
// //
// // The call only opens the transmission channel, explicit fetching is done using the ObjectRangeReader.
// // Exactly one return value is non-nil. Resulting reader must be finally closed.
// //
// // Context is required and must not be nil. It is used for network communication.
// //
// // Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// // the specified Signer, which is taken into account, in particular, for access control.
// //
// // Return errors:
// //   - [ErrZeroRangeLength]
// //   - [ErrMissingSigner]
// func (c *Client) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer user.Signer, prm PrmObjectRange) (*ObjectRangeReader, error) {
// 	var (
// 		addr  v2refs.Address
// 		cidV2 v2refs.ContainerID
// 		oidV2 v2refs.ObjectID
// 		rngV2 v2object.Range
// 		body  v2object.GetRangeRequestBody
// 		err   error
// 	)
//
// 	defer func() {
// 		c.sendStatistic(stat.MethodObjectRange, err)()
// 	}()
//
// 	if length == 0 {
// 		err = ErrZeroRangeLength
// 		return nil, err
// 	}
//
// 	if signer == nil {
// 		return nil, ErrMissingSigner
// 	}
//
// 	containerID.WriteToV2(&cidV2)
// 	addr.SetContainerID(&cidV2)
//
// 	objectID.WriteToV2(&oidV2)
// 	addr.SetObjectID(&oidV2)
//
// 	rngV2.SetOffset(offset)
// 	rngV2.SetLength(length)
//
// 	// form request body
// 	body.SetRaw(prm.raw)
// 	body.SetAddress(&addr)
// 	body.SetRange(&rngV2)
//
// 	// form request
// 	var req v2object.GetRangeRequest
//
// 	req.SetBody(&body)
// 	c.prepareRequest(&req, &prm.meta)
//
// 	buf := c.buffers.Get().(*[]byte)
// 	err = signServiceMessage(signer, &req, *buf)
// 	c.buffers.Put(buf)
// 	if err != nil {
// 		err = fmt.Errorf("sign request: %w", err)
// 		return nil, err
// 	}
//
// 	ctx, cancel := context.WithCancel(ctx)
//
// 	stream, err := rpcAPIGetObjectRange(&c.c, &req, client.WithContext(ctx))
// 	if err != nil {
// 		cancel()
// 		err = fmt.Errorf("open stream: %w", err)
// 		return nil, err
// 	}
//
// 	var r ObjectRangeReader
// 	r.remainingPayloadLen = int(length)
// 	r.cancelCtxStream = cancel
// 	r.stream = stream
// 	r.client = c
// 	r.statisticCallback = func(err error) {
// 		c.sendStatistic(stat.MethodObjectRangeStream, err)()
// 	}
//
// 	return &r, nil
// }
