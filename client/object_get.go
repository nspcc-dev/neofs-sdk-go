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

// PrmObjectGet groups parameters of ObjectGetInit operation.
type PrmObjectGet struct {
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
func (x *PrmObjectGet) MarkRaw() {
	x.raw = true
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectGet) MarkLocal() {
	x.local = true
}

// WithinSession specifies session within which object should be read.
func (x *PrmObjectGet) WithinSession(t session.Token) {
	x.session = t
	x.sessionSet = true
}

// WithBearerToken attaches bearer token to be used for the operation.
func (x *PrmObjectGet) WithBearerToken(t token.BearerToken) {
	x.bearer = t
	x.bearerSet = true
}

// FromContainer specifies NeoFS container of the object.
// Required parameter.
func (x *PrmObjectGet) FromContainer(id cid.ID) {
	x.cnr = id
	x.cnrSet = true
}

// ByID specifies identifier of the requested object.
// Required parameter.
func (x *PrmObjectGet) ByID(id oid.ID) {
	x.obj = id
	x.objSet = true
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
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Client default key is used.
func (x *ObjectReader) UseKey(key ecdsa.PrivateKey) {
	x.ctxCall.key = key
}

// ReadHeader reads header of the object. Result means success.
// Failure reason can be received via Close.
func (x *ObjectReader) ReadHeader(dst *object.Object) bool {
	if !x.ctxCall.writeRequest() {
		return false
	} else if !x.ctxCall.readResponse() {
		return false
	}

	var partInit *v2object.GetObjectPartInit

	switch v := x.bodyResp.GetObjectPart().(type) {
	default:
		x.ctxCall.err = fmt.Errorf("unexpected message instead of heading part: %T", v)
		return false
	case *v2object.SplitInfo:
		x.ctxCall.err = object.NewSplitInfoError(object.NewSplitInfoFromV2(v))
		return false
	case *v2object.GetObjectPartInit:
		partInit = v
	}

	var objv2 v2object.Object

	objv2.SetObjectID(partInit.GetObjectID())
	objv2.SetHeader(partInit.GetHeader())
	objv2.SetSignature(partInit.GetSignature())

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

func (x *ObjectReader) Read(p []byte) (int, error) {
	n, ok := x.readChunk(p)
	if !ok {
		res, err := x.close(false)
		if err != nil {
			return n, err
		} else if !x.ctxCall.resolveAPIFailures {
			return n, apistatus.ErrFromStatus(res.Status())
		}
	}

	return n, nil
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching is done using the ObjectWriter.
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
