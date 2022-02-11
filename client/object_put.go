package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)

// PrmObjectPutInit groups parameters of ObjectPutInit operation.
//
// At the moment the operation is not parameterized, however,
// the structure is still declared for backward compatibility.
type PrmObjectPutInit struct{}

// ResObjectPut groups the final result values of ObjectPutInit operation.
type ResObjectPut struct {
	statusRes

	resp v2object.PutResponse
}

// ReadStoredObjectID reads identifier of the saved object.
// Returns false if ID is missing (not read).
func (x *ResObjectPut) ReadStoredObjectID(id *oid.ID) bool {
	idv2 := x.resp.GetBody().GetObjectID()
	if idv2 == nil {
		return false
	}

	*id = *oid.NewIDFromV2(idv2) // need smth better

	return true
}

// ObjectWriter is designed to write one object to NeoFS system.
//
// Must be initialized using Client.ObjectPutInit, any other
// usage is unsafe.
type ObjectWriter struct {
	cancelCtxStream context.CancelFunc

	ctxCall contextCall

	// initially bound tp contextCall
	metaHdr v2session.RequestMetaHeader

	// initially bound to contextCall
	partInit v2object.PutObjectPartInit

	chunkCalled bool

	partChunk v2object.PutObjectPartChunk
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Client default key is used.
func (x *ObjectWriter) UseKey(key ecdsa.PrivateKey) {
	x.ctxCall.key = key
}

// WithBearerToken attaches bearer token to be used for the operation.
// Should be called once before any writing steps.
func (x *ObjectWriter) WithBearerToken(t token.BearerToken) {
	x.metaHdr.SetBearerToken(t.ToV2())
}

// WithinSession specifies session within which object should be stored.
// Should be called once before any writing steps.
func (x *ObjectWriter) WithinSession(t session.Token) {
	x.metaHdr.SetSessionToken(t.ToV2())
}

// MarkLocal tells the server to execute the operation locally.
func (x *ObjectWriter) MarkLocal() {
	x.metaHdr.SetTTL(1)
}

// WriteHeader writes header of the object. Result means success.
// Failure reason can be received via Close.
func (x *ObjectWriter) WriteHeader(hdr object.Object) bool {
	v2Hdr := hdr.ToV2()

	x.partInit.SetObjectID(v2Hdr.GetObjectID())
	x.partInit.SetHeader(v2Hdr.GetHeader())
	x.partInit.SetSignature(v2Hdr.GetSignature())

	return x.ctxCall.writeRequest()
}

// WritePayloadChunk writes chunk of the object payload. Result means success.
// Failure reason can be received via Close.
func (x *ObjectWriter) WritePayloadChunk(chunk []byte) bool {
	if !x.chunkCalled {
		x.chunkCalled = true
		x.ctxCall.req.(*v2object.PutRequest).GetBody().SetObjectPart(&x.partChunk)
	}

	for ln := len(chunk); ln > 0; ln = len(chunk) {
		if ln > 512 {
			ln = 512
		}

		x.partChunk.SetChunk(chunk[:ln])

		if !x.ctxCall.writeRequest() {
			return false
		}

		chunk = chunk[ln:]
	}

	return true
}

// Close ends writing the object and returns the result of the operation
// along with the final results. Must be called after using the ObjectWriter.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return statuses:
//   global (see Client docs).
func (x *ObjectWriter) Close() (*ResObjectPut, error) {
	defer x.cancelCtxStream()

	if x.ctxCall.err != nil {
		return nil, x.ctxCall.err
	}

	if !x.ctxCall.close() {
		return nil, x.ctxCall.err
	}

	if !x.ctxCall.processResponse() {
		return nil, x.ctxCall.err
	}

	return x.ctxCall.statusRes.(*ResObjectPut), nil
}

// ObjectPutInit initiates writing an object through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit recording is done using the ObjectWriter.
// Exactly one return value is non-nil. Resulting writer must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
func (c *Client) ObjectPutInit(ctx context.Context, _ PrmObjectPutInit) (*ObjectWriter, error) {
	// check parameters
	if ctx == nil {
		panic(panicMsgMissingContext)
	}

	// open stream
	var (
		res ResObjectPut
		w   ObjectWriter
	)

	ctx, w.cancelCtxStream = context.WithCancel(ctx)

	stream, err := rpcapi.PutObject(c.Raw(), &res.resp, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	// form request body
	var body v2object.PutRequestBody

	// form request
	var req v2object.PutRequest

	req.SetBody(&body)

	req.SetMetaHeader(&w.metaHdr)
	body.SetObjectPart(&w.partInit)

	// init call context
	c.initCallContext(&w.ctxCall)
	w.ctxCall.req = &req
	w.ctxCall.statusRes = &res
	w.ctxCall.resp = &res.resp
	w.ctxCall.wReq = func() error {
		return stream.Write(&req)
	}
	w.ctxCall.closer = stream.Close

	return &w, nil
}
