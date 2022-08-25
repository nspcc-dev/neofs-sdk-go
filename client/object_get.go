package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// shared parameters of GET/HEAD/RANGE.
type prmObjectRead struct {
	meta v2session.RequestMetaHeader

	raw bool

	addr v2refs.Address
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *prmObjectRead) WithXHeaders(hs ...string) {
	if len(hs)%2 != 0 {
		panic("slice of X-Headers with odd length")
	}

	writeXHeadersToMeta(hs, &x.meta)
}

// MarkRaw marks an intent to read physically stored object.
func (x *prmObjectRead) MarkRaw() {
	x.raw = true
}

// MarkLocal tells the server to execute the operation locally.
func (x *prmObjectRead) MarkLocal() {
	x.meta.SetTTL(1)
}

// WithinSession specifies session within which object should be read.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
func (x *prmObjectRead) WithinSession(t session.Object) {
	var tokv2 v2session.Token
	t.WriteToV2(&tokv2)
	x.meta.SetSessionToken(&tokv2)
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *prmObjectRead) WithBearerToken(t bearer.Token) {
	var v2token acl.BearerToken
	t.WriteToV2(&v2token)
	x.meta.SetBearerToken(&v2token)
}

// FromContainer specifies NeoFS container of the object.
// Required parameter.
func (x *prmObjectRead) FromContainer(id cid.ID) {
	var cnrV2 v2refs.ContainerID
	id.WriteToV2(&cnrV2)
	x.addr.SetContainerID(&cnrV2)
}

// ByID specifies identifier of the requested object.
// Required parameter.
func (x *prmObjectRead) ByID(id oid.ID) {
	var objV2 v2refs.ObjectID
	id.WriteToV2(&objV2)
	x.addr.SetObjectID(&objV2)
}

// PrmObjectGet groups parameters of ObjectGetInit operation.
type PrmObjectGet struct {
	prmObjectRead
}

// ResObjectGet groups the final result values of ObjectGetInit operation.
type ResObjectGet struct {
	statusRes
}

// ObjectReader is designed to read one object from NeoFS system.
//
// Must be initialized using Client.ObjectGetInit, any other
// usage is unsafe.
type ObjectReader struct {
	cancelCtxStream context.CancelFunc

	ctxCall contextCall

	// initially bound to contextCall
	bodyResp v2object.GetResponseBody

	tailPayload []byte

	remainingPayloadLen int
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Client default key is used.
func (x *ObjectReader) UseKey(key ecdsa.PrivateKey) {
	x.ctxCall.key = key
}

func handleSplitInfo(ctx *contextCall, i *v2object.SplitInfo) {
	ctx.err = object.NewSplitInfoError(object.NewSplitInfoFromV2(i))
}

// ReadHeader reads header of the object. Result means success.
// Failure reason can be received via Close.
func (x *ObjectReader) ReadHeader(dst *object.Object) bool {
	if !x.ctxCall.writeRequest() || !x.ctxCall.readResponse() {
		return false
	}

	var partInit *v2object.GetObjectPartInit

	switch v := x.bodyResp.GetObjectPart().(type) {
	default:
		x.ctxCall.err = fmt.Errorf("unexpected message instead of heading part: %T", v)
		return false
	case *v2object.SplitInfo:
		handleSplitInfo(&x.ctxCall, v)
		return false
	case *v2object.GetObjectPartInit:
		partInit = v
	}

	var objv2 v2object.Object

	objv2.SetObjectID(partInit.GetObjectID())
	objv2.SetHeader(partInit.GetHeader())
	objv2.SetSignature(partInit.GetSignature())

	x.remainingPayloadLen = int(objv2.GetHeader().GetPayloadLength())

	*dst = *object.NewFromV2(&objv2) // need smth better

	return true
}

func (x *ObjectReader) readChunk(buf []byte) (int, bool) {
	var read int

	// read remaining tail
	read = copy(buf, x.tailPayload)

	x.tailPayload = x.tailPayload[read:]

	if len(buf) == read {
		return read, true
	}

	var ok bool
	var part v2object.GetObjectPart
	var chunk []byte
	var lastRead int

	for {
		// receive next message
		ok = x.ctxCall.readResponse()
		if !ok {
			return read, false
		}

		// get chunk part message
		part = x.bodyResp.GetObjectPart()

		var partChunk *v2object.GetObjectPartChunk

		partChunk, ok = part.(*v2object.GetObjectPartChunk)
		if !ok {
			x.ctxCall.err = fmt.Errorf("unexpected message instead of chunk part: %T", part)
			return read, false
		}

		// read new chunk
		chunk = partChunk.GetChunk()
		if len(chunk) == 0 {
			// just skip empty chunks since they are not prohibited by protocol
			continue
		}

		lastRead = copy(buf[read:], chunk)

		read += lastRead

		if read == len(buf) {
			// save the tail
			x.tailPayload = append(x.tailPayload, chunk[lastRead:]...)

			return read, true
		}
	}
}

// ReadChunk reads another chunk of the object payload. Works similar to
// io.Reader.Read but returns success flag instead of error.
//
// Failure reason can be received via Close.
func (x *ObjectReader) ReadChunk(buf []byte) (int, bool) {
	return x.readChunk(buf)
}

func (x *ObjectReader) close(ignoreEOF bool) (*ResObjectGet, error) {
	defer x.cancelCtxStream()

	if x.ctxCall.err != nil {
		if !errors.Is(x.ctxCall.err, io.EOF) {
			return nil, x.ctxCall.err
		} else if !ignoreEOF {
			if x.remainingPayloadLen > 0 {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, io.EOF
		}
	}

	return x.ctxCall.statusRes.(*ResObjectGet), nil
}

// Close ends reading the object and returns the result of the operation
// along with the final results. Must be called after using the ObjectReader.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return errors:
//
//	*object.SplitInfoError (returned on virtual objects with PrmObjectGet.MakeRaw).
//
// Return statuses:
//   - global (see Client docs);
//   - *apistatus.ContainerNotFound;
//   - *apistatus.ObjectNotFound;
//   - *apistatus.ObjectAccessDenied;
//   - *apistatus.ObjectAlreadyRemoved;
//   - *apistatus.SessionTokenExpired.
func (x *ObjectReader) Close() (*ResObjectGet, error) {
	return x.close(true)
}

// Read implements io.Reader of the object payload.
func (x *ObjectReader) Read(p []byte) (int, error) {
	n, ok := x.readChunk(p)

	x.remainingPayloadLen -= n

	if !ok {
		res, err := x.close(false)
		if err != nil {
			return n, err
		}

		return n, apistatus.ErrFromStatus(res.Status())
	}

	if x.remainingPayloadLen < 0 {
		return n, errors.New("payload size overflow")
	}

	return n, nil
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching is done using the ObjectReader.
// Exactly one return value is non-nil. Resulting reader must be finally closed.
//
// Immediately panics if parameters are set incorrectly (see PrmObjectGet docs).
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ObjectGetInit(ctx context.Context, prm PrmObjectGet) (*ObjectReader, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.addr.GetContainerID() == nil:
		panic(panicMsgMissingContainer)
	case prm.addr.GetObjectID() == nil:
		panic(panicMsgMissingObject)
	}

	// form request body
	var body v2object.GetRequestBody

	body.SetRaw(prm.raw)
	body.SetAddress(&prm.addr)

	// form request
	var req v2object.GetRequest

	req.SetBody(&body)
	req.SetMetaHeader(&prm.meta)

	// init reader
	var (
		r      ObjectReader
		resp   v2object.GetResponse
		stream *rpcapi.GetResponseReader
	)

	ctx, r.cancelCtxStream = context.WithCancel(ctx)

	resp.SetBody(&r.bodyResp)

	// init call context
	c.initCallContext(&r.ctxCall)
	r.ctxCall.req = &req
	r.ctxCall.statusRes = new(ResObjectGet)
	r.ctxCall.resp = &resp
	r.ctxCall.wReq = func() error {
		var err error

		stream, err = rpcapi.GetObject(&c.c, &req, client.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("open stream: %w", err)
		}

		return nil
	}
	r.ctxCall.rResp = func() error {
		return stream.Read(&resp)
	}

	return &r, nil
}

// PrmObjectHead groups parameters of ObjectHead operation.
type PrmObjectHead struct {
	prmObjectRead

	keySet bool
	key    ecdsa.PrivateKey
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Client default key is used.
func (x *PrmObjectHead) UseKey(key ecdsa.PrivateKey) {
	x.keySet = true
	x.key = key
}

// ResObjectHead groups resulting values of ObjectHead operation.
type ResObjectHead struct {
	statusRes

	// requested object (response doesn't carry the ID)
	idObj oid.ID

	hdr *v2object.HeaderWithSignature
}

// ReadHeader reads header of the requested object.
// Returns false if header is missing in the response (not read).
func (x *ResObjectHead) ReadHeader(dst *object.Object) bool {
	if x.hdr == nil {
		return false
	}

	var objv2 v2object.Object

	objv2.SetHeader(x.hdr.GetHeader())
	objv2.SetSignature(x.hdr.GetSignature())

	obj := object.NewFromV2(&objv2)
	obj.SetID(x.idObj)

	*dst = *obj

	return true
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// If PrmInit.ResolveNeoFSFailures has been called, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmObjectHead docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//
//	*object.SplitInfoError (returned on virtual objects with PrmObjectHead.MakeRaw).
//
// Return statuses:
//   - global (see Client docs);
//   - *apistatus.ContainerNotFound;
//   - *apistatus.ObjectNotFound;
//   - *apistatus.ObjectAccessDenied;
//   - *apistatus.ObjectAlreadyRemoved;
//   - *apistatus.SessionTokenExpired.
func (c *Client) ObjectHead(ctx context.Context, prm PrmObjectHead) (*ResObjectHead, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.addr.GetContainerID() == nil:
		panic(panicMsgMissingContainer)
	case prm.addr.GetObjectID() == nil:
		panic(panicMsgMissingObject)
	}

	// form request body
	var body v2object.HeadRequestBody

	body.SetRaw(prm.raw)
	body.SetAddress(&prm.addr)

	// form request
	var req v2object.HeadRequest

	req.SetBody(&body)
	req.SetMetaHeader(&prm.meta)

	// init call context

	var (
		cc  contextCall
		res ResObjectHead
	)

	_ = res.idObj.ReadFromV2(*prm.addr.GetObjectID())

	c.initCallContext(&cc)
	if prm.keySet {
		cc.key = prm.key
	}

	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.HeadObject(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		switch v := r.(*v2object.HeadResponse).GetBody().GetHeaderPart().(type) {
		default:
			cc.err = fmt.Errorf("unexpected header type %T", v)
		case *v2object.SplitInfo:
			handleSplitInfo(&cc, v)
		case *v2object.HeaderWithSignature:
			res.hdr = v
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// PrmObjectRange groups parameters of ObjectRange operation.
type PrmObjectRange struct {
	prmObjectRead

	rng v2object.Range
}

// SetOffset sets offset of the payload range to be read.
// Zero by default.
func (x *PrmObjectRange) SetOffset(off uint64) {
	x.rng.SetOffset(off)
}

// SetLength sets length of the payload range to be read.
// Must be positive.
func (x *PrmObjectRange) SetLength(ln uint64) {
	x.rng.SetLength(ln)
}

// ResObjectRange groups the final result values of ObjectRange operation.
type ResObjectRange struct {
	statusRes
}

// ObjectRangeReader is designed to read payload range of one object
// from NeoFS system.
//
// Must be initialized using Client.ObjectRangeInit, any other
// usage is unsafe.
type ObjectRangeReader struct {
	cancelCtxStream context.CancelFunc

	ctxCall contextCall

	reqWritten bool

	// initially bound to contextCall
	bodyResp v2object.GetRangeResponseBody

	tailPayload []byte

	remainingPayloadLen int
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Client default key is used.
func (x *ObjectRangeReader) UseKey(key ecdsa.PrivateKey) {
	x.ctxCall.key = key
}

func (x *ObjectRangeReader) readChunk(buf []byte) (int, bool) {
	if !x.reqWritten {
		if !x.ctxCall.writeRequest() {
			return 0, false
		}

		x.reqWritten = true
	}

	var read int

	// read remaining tail
	read = copy(buf, x.tailPayload)

	x.tailPayload = x.tailPayload[read:]

	if len(buf) == read {
		return read, true
	}

	var ok bool
	var partChunk *v2object.GetRangePartChunk
	var chunk []byte
	var lastRead int

	for {
		// receive next message
		ok = x.ctxCall.readResponse()
		if !ok {
			return read, false
		}

		// get chunk message
		switch v := x.bodyResp.GetRangePart().(type) {
		default:
			x.ctxCall.err = fmt.Errorf("unexpected message received: %T", v)
			return read, false
		case *v2object.SplitInfo:
			handleSplitInfo(&x.ctxCall, v)
			return read, false
		case *v2object.GetRangePartChunk:
			partChunk = v
		}

		chunk = partChunk.GetChunk()
		if len(chunk) == 0 {
			// just skip empty chunks since they are not prohibited by protocol
			continue
		}

		lastRead = copy(buf[read:], chunk)

		read += lastRead

		if read == len(buf) {
			// save the tail
			x.tailPayload = append(x.tailPayload, chunk[lastRead:]...)

			return read, true
		}
	}
}

// ReadChunk reads another chunk of the object payload range.
// Works similar to io.Reader.Read but returns success flag instead of error.
//
// Failure reason can be received via Close.
func (x *ObjectRangeReader) ReadChunk(buf []byte) (int, bool) {
	return x.readChunk(buf)
}

func (x *ObjectRangeReader) close(ignoreEOF bool) (*ResObjectRange, error) {
	defer x.cancelCtxStream()

	if x.ctxCall.err != nil {
		if !errors.Is(x.ctxCall.err, io.EOF) {
			return nil, x.ctxCall.err
		} else if !ignoreEOF {
			if x.remainingPayloadLen > 0 {
				return nil, io.ErrUnexpectedEOF
			}

			return nil, io.EOF
		}
	}

	return x.ctxCall.statusRes.(*ResObjectRange), nil
}

// Close ends reading the payload range and returns the result of the operation
// along with the final results. Must be called after using the ObjectRangeReader.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return errors:
//
//	*object.SplitInfoError (returned on virtual objects with PrmObjectRange.MakeRaw).
//
// Return statuses:
//   - global (see Client docs);
//   - *apistatus.ContainerNotFound;
//   - *apistatus.ObjectNotFound;
//   - *apistatus.ObjectAccessDenied;
//   - *apistatus.ObjectAlreadyRemoved;
//   - *apistatus.ObjectOutOfRange;
//   - *apistatus.SessionTokenExpired.
func (x *ObjectRangeReader) Close() (*ResObjectRange, error) {
	return x.close(true)
}

// Read implements io.Reader of the object payload.
func (x *ObjectRangeReader) Read(p []byte) (int, error) {
	n, ok := x.readChunk(p)

	x.remainingPayloadLen -= n

	if !ok {
		res, err := x.close(false)
		if err != nil {
			return n, err
		}

		return n, apistatus.ErrFromStatus(res.Status())
	}

	if x.remainingPayloadLen < 0 {
		return n, errors.New("payload range size overflow")
	}

	return n, nil
}

// ObjectRangeInit initiates reading an object's payload range through a remote
// server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching is done using the ObjectRangeReader.
// Exactly one return value is non-nil. Resulting reader must be finally closed.
//
// Immediately panics if parameters are set incorrectly (see PrmObjectRange docs).
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ObjectRangeInit(ctx context.Context, prm PrmObjectRange) (*ObjectRangeReader, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.addr.GetContainerID() == nil:
		panic(panicMsgMissingContainer)
	case prm.addr.GetObjectID() == nil:
		panic(panicMsgMissingObject)
	case prm.rng.GetLength() == 0:
		panic("zero range length")
	}

	// form request body
	var body v2object.GetRangeRequestBody

	body.SetRaw(prm.raw)
	body.SetAddress(&prm.addr)
	body.SetRange(&prm.rng)

	// form request
	var req v2object.GetRangeRequest

	req.SetBody(&body)
	req.SetMetaHeader(&prm.meta)

	// init reader
	var (
		r      ObjectRangeReader
		resp   v2object.GetRangeResponse
		stream *rpcapi.ObjectRangeResponseReader
	)

	r.remainingPayloadLen = int(prm.rng.GetLength())

	ctx, r.cancelCtxStream = context.WithCancel(ctx)

	resp.SetBody(&r.bodyResp)

	// init call context
	c.initCallContext(&r.ctxCall)
	r.ctxCall.req = &req
	r.ctxCall.statusRes = new(ResObjectRange)
	r.ctxCall.resp = &resp
	r.ctxCall.wReq = func() error {
		var err error

		stream, err = rpcapi.GetObjectRange(&c.c, &req, client.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("open stream: %w", err)
		}

		return nil
	}
	r.ctxCall.rResp = func() error {
		return stream.Read(&resp)
	}

	return &r, nil
}
