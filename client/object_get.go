package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)

// shared parameters of GET/HEAD/RANGE.
type prmObjectRead struct {
	raw bool

	local bool

	sessionSet bool
	session    session.Token

	bearerSet bool
	bearer    token.BearerToken

	cnrSet bool
	cnr    cid.ID

	objSet bool
	obj    oid.ID
}

// MarkRaw marks an intent to read physically stored object.
func (x *prmObjectRead) MarkRaw() {
	x.raw = true
}

// MarkLocal tells the server to execute the operation locally.
func (x *prmObjectRead) MarkLocal() {
	x.local = true
}

// WithinSession specifies session within which object should be read.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
func (x *prmObjectRead) WithinSession(t session.Token) {
	x.session = t
	x.sessionSet = true
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *prmObjectRead) WithBearerToken(t token.BearerToken) {
	x.bearer = t
	x.bearerSet = true
}

// FromContainer specifies NeoFS container of the object.
// Required parameter.
func (x *prmObjectRead) FromContainer(id cid.ID) {
	x.cnr = id
	x.cnrSet = true
}

// ByID specifies identifier of the requested object.
// Required parameter.
func (x *prmObjectRead) ByID(id oid.ID) {
	x.obj = id
	x.objSet = true
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

	// receive next message
	ok := x.ctxCall.readResponse()
	if !ok {
		return read, false
	}

	// get chunk part message
	part := x.bodyResp.GetObjectPart()

	var partChunk *v2object.GetObjectPartChunk

	partChunk, ok = part.(*v2object.GetObjectPartChunk)
	if !ok {
		x.ctxCall.err = fmt.Errorf("unexpected message instead of chunk part: %T", part)
		return read, false
	}

	// read new chunk
	chunk := partChunk.GetChunk()

	tailOffset := copy(buf[read:], chunk)

	read += tailOffset

	// save the tail
	x.tailPayload = append(x.tailPayload, chunk[tailOffset:]...)

	return read, true
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
		} else if x.remainingPayloadLen > 0 {
			return nil, io.ErrUnexpectedEOF
		} else if !ignoreEOF {
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
//   *object.SplitInfoError (returned on virtual objects with PrmObjectGet.MakeRaw).
//
// Return statuses:
//   global (see Client docs).
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
	case !prm.cnrSet:
		panic(panicMsgMissingContainer)
	case !prm.objSet:
		panic("missing object")
	}

	var addr v2refs.Address

	addr.SetContainerID(prm.cnr.ToV2())
	addr.SetObjectID(prm.obj.ToV2())

	// form request body
	var body v2object.GetRequestBody

	body.SetRaw(prm.raw)
	body.SetAddress(&addr)

	// form meta header
	var meta v2session.RequestMetaHeader

	if prm.local {
		meta.SetTTL(1)
	}

	if prm.bearerSet {
		meta.SetBearerToken(prm.bearer.ToV2())
	}

	if prm.sessionSet {
		meta.SetSessionToken(prm.session.ToV2())
	}

	// form request
	var req v2object.GetRequest

	req.SetBody(&body)
	req.SetMetaHeader(&meta)

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

		stream, err = rpcapi.GetObject(c.Raw(), &req, client.WithContext(ctx))
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

// GetFullObject reads object header and full payload to the memory and returns
// the result.
//
// Function behavior equivalent to Client.ObjectGetInit, see documentation for details.
func GetFullObject(ctx context.Context, c *Client, idCnr cid.ID, idObj oid.ID) (*object.Object, error) {
	var prm PrmObjectGet

	prm.FromContainer(idCnr)
	prm.ByID(idObj)

	rdr, err := c.ObjectGetInit(ctx, prm)
	if err != nil {
		return nil, fmt.Errorf("init object reading: %w", err)
	}

	var obj object.Object

	if rdr.ReadHeader(&obj) {
		payload, err := io.ReadAll(rdr)
		if err != nil {
			return nil, fmt.Errorf("read payload: %w", err)
		}

		object.NewRawFrom(&obj).SetPayload(payload)
	}

	res, err := rdr.Close()
	if err == nil {
		err = apistatus.ErrFromStatus(res.Status())
	}

	if err != nil {
		return nil, fmt.Errorf("finish object reading: %w", err)
	}

	return &obj, nil
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

	raw := object.NewRawFromV2(&objv2)
	raw.SetID(&x.idObj)

	*dst = *raw.Object()

	return true
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmObjectHead docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   *object.SplitInfoError (returned on virtual objects with PrmObjectHead.MakeRaw).
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) ObjectHead(ctx context.Context, prm PrmObjectHead) (*ResObjectHead, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.cnrSet:
		panic(panicMsgMissingContainer)
	case !prm.objSet:
		panic("missing object")
	}

	var addr v2refs.Address

	addr.SetContainerID(prm.cnr.ToV2())
	addr.SetObjectID(prm.obj.ToV2())

	// form request body
	var body v2object.HeadRequestBody

	body.SetRaw(prm.raw)
	body.SetAddress(&addr)

	// form meta header
	var meta v2session.RequestMetaHeader

	if prm.local {
		meta.SetTTL(1)
	}

	if prm.bearerSet {
		meta.SetBearerToken(prm.bearer.ToV2())
	}

	if prm.sessionSet {
		meta.SetSessionToken(prm.session.ToV2())
	}

	// form request
	var req v2object.HeadRequest

	req.SetBody(&body)
	req.SetMetaHeader(&meta)

	// init call context

	var (
		cc  contextCall
		res ResObjectHead
	)

	res.idObj = prm.obj

	if prm.keySet {
		c.initCallContextWithoutKey(&cc)
		cc.key = prm.key
	} else {
		c.initCallContext(&cc)
	}

	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.HeadObject(c.Raw(), &req, client.WithContext(ctx))
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

	off, ln uint64
}

// SetOffset sets offset of the payload range to be read.
// Zero by default.
func (x *PrmObjectRange) SetOffset(off uint64) {
	x.off = off
}

// SetLength sets length of the payload range to be read.
// Must be positive.
func (x *PrmObjectRange) SetLength(ln uint64) {
	x.ln = ln
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

	// receive next message
	ok := x.ctxCall.readResponse()
	if !ok {
		return read, false
	}

	// get chunk message
	var partChunk *v2object.GetRangePartChunk

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

	// read new chunk
	chunk := partChunk.GetChunk()

	tailOffset := copy(buf[read:], chunk)

	read += tailOffset

	// save the tail
	x.tailPayload = append(x.tailPayload, chunk[tailOffset:]...)

	return read, true
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
		} else if x.remainingPayloadLen > 0 {
			return nil, io.ErrUnexpectedEOF
		} else if !ignoreEOF {
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
//   *object.SplitInfoError (returned on virtual objects with PrmObjectRange.MakeRaw).
//
// Return statuses:
//   global (see Client docs).
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
	case !prm.cnrSet:
		panic(panicMsgMissingContainer)
	case !prm.objSet:
		panic("missing object")
	case prm.ln == 0:
		panic("zero range length")
	}

	var addr v2refs.Address

	addr.SetContainerID(prm.cnr.ToV2())
	addr.SetObjectID(prm.obj.ToV2())

	var rng v2object.Range

	rng.SetOffset(prm.off)
	rng.SetLength(prm.ln)

	// form request body
	var body v2object.GetRangeRequestBody

	body.SetRaw(prm.raw)
	body.SetAddress(&addr)
	body.SetRange(&rng)

	// form meta header
	var meta v2session.RequestMetaHeader

	if prm.local {
		meta.SetTTL(1)
	}

	if prm.bearerSet {
		meta.SetBearerToken(prm.bearer.ToV2())
	}

	if prm.sessionSet {
		meta.SetSessionToken(prm.session.ToV2())
	}

	// form request
	var req v2object.GetRangeRequest

	req.SetBody(&body)
	req.SetMetaHeader(&meta)

	// init reader
	var (
		r      ObjectRangeReader
		resp   v2object.GetRangeResponse
		stream *rpcapi.ObjectRangeResponseReader
	)

	r.remainingPayloadLen = int(prm.ln)

	ctx, r.cancelCtxStream = context.WithCancel(ctx)

	resp.SetBody(&r.bodyResp)

	// init call context
	c.initCallContext(&r.ctxCall)
	r.ctxCall.req = &req
	r.ctxCall.statusRes = new(ResObjectRange)
	r.ctxCall.resp = &resp
	r.ctxCall.wReq = func() error {
		var err error

		stream, err = rpcapi.GetObjectRange(c.Raw(), &req, client.WithContext(ctx))
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
