package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/slicer"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmObjectPutInit groups parameters of ObjectPutInit operation.
type PrmObjectPutInit struct {
	copyNum uint32
	signer  neofscrypto.Signer
	meta    v2session.RequestMetaHeader
}

// SetCopiesNumber sets number of object copies that is enough to consider put successful.
func (x *PrmObjectPutInit) SetCopiesNumber(copiesNumber uint32) {
	x.copyNum = copiesNumber
}

// ResObjectPut groups the final result values of ObjectPutInit operation.
type ResObjectPut struct {
	obj oid.ID
}

// StoredObjectID returns identifier of the saved object.
func (x ResObjectPut) StoredObjectID() oid.ID {
	return x.obj
}

// ObjectWriter is designed to write one object to NeoFS system.
//
// Must be initialized using Client.ObjectPutInit, any other
// usage is unsafe.
type ObjectWriter struct {
	cancelCtxStream context.CancelFunc

	client *Client
	stream interface {
		Write(*v2object.PutRequest) error
		Close() error
	}

	signer neofscrypto.Signer
	res    ResObjectPut
	err    error

	chunkCalled bool

	respV2    v2object.PutResponse
	req       v2object.PutRequest
	partInit  v2object.PutObjectPartInit
	partChunk v2object.PutObjectPartChunk
}

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Client default signer is used.
func (x *PrmObjectPutInit) UseSigner(signer neofscrypto.Signer) {
	x.signer = signer
}

// WithBearerToken attaches bearer token to be used for the operation.
// Should be called once before any writing steps.
func (x *PrmObjectPutInit) WithBearerToken(t bearer.Token) {
	var v2token acl.BearerToken
	t.WriteToV2(&v2token)
	x.meta.SetBearerToken(&v2token)
}

// WithinSession specifies session within which object should be stored.
// Should be called once before any writing steps.
func (x *PrmObjectPutInit) WithinSession(t session.Object) {
	var tv2 v2session.Token
	t.WriteToV2(&tv2)

	x.meta.SetSessionToken(&tv2)
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectPutInit) MarkLocal() {
	x.meta.SetTTL(1)
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *PrmObjectPutInit) WithXHeaders(hs ...string) {
	writeXHeadersToMeta(hs, &x.meta)
}

// WriteHeader writes header of the object. Result means success.
// Failure reason can be received via Close.
func (x *ObjectWriter) WriteHeader(hdr object.Object) bool {
	v2Hdr := hdr.ToV2()

	x.partInit.SetObjectID(v2Hdr.GetObjectID())
	x.partInit.SetHeader(v2Hdr.GetHeader())
	x.partInit.SetSignature(v2Hdr.GetSignature())

	x.req.GetBody().SetObjectPart(&x.partInit)
	x.req.SetVerificationHeader(nil)

	x.err = signServiceMessage(x.signer, &x.req)
	if x.err != nil {
		x.err = fmt.Errorf("sign message: %w", x.err)
		return false
	}

	x.err = x.stream.Write(&x.req)
	return x.err == nil
}

// WritePayloadChunk writes chunk of the object payload. Result means success.
// Failure reason can be received via Close.
func (x *ObjectWriter) WritePayloadChunk(chunk []byte) bool {
	if !x.chunkCalled {
		x.chunkCalled = true
		x.req.GetBody().SetObjectPart(&x.partChunk)
	}

	for ln := len(chunk); ln > 0; ln = len(chunk) {
		// maxChunkLen restricts maximum byte length of the chunk
		// transmitted in a single stream message. It depends on
		// server settings and other message fields, but for now
		// we simply assume that 3MB is large enough to reduce the
		// number of messages, and not to exceed the limit
		// (4MB by default for gRPC servers).
		const maxChunkLen = 3 << 20
		if ln > maxChunkLen {
			ln = maxChunkLen
		}

		// we deal with size limit overflow above, but there is another case:
		// what if method is called with "small" chunk many times? We write
		// a message to the stream on each call. Alternatively, we could use buffering.
		// In most cases, the chunk length does not vary between calls. Given this
		// assumption, as well as the length of the payload from the header, it is
		// possible to buffer the data of intermediate chunks, and send a message when
		// the allocated buffer is filled, or when the last chunk is received.
		// It is mentally assumed that allocating and filling the buffer is better than
		// synchronous sending, but this needs to be tested.
		x.partChunk.SetChunk(chunk[:ln])
		x.req.SetVerificationHeader(nil)

		x.err = signServiceMessage(x.signer, &x.req)
		if x.err != nil {
			x.err = fmt.Errorf("sign message: %w", x.err)
			return false
		}

		x.err = x.stream.Write(&x.req)
		if x.err != nil {
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
// Return errors:
//   - global (see Client docs)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectLocked]
//   - [apistatus.ErrLockNonRegularObject]
//   - [apistatus.ErrSessionTokenNotFound]
//   - [apistatus.ErrSessionTokenExpired]
func (x *ObjectWriter) Close() (*ResObjectPut, error) {
	defer x.cancelCtxStream()

	// Ignore io.EOF error, because it is expected error for client-side
	// stream termination by the server. E.g. when stream contains invalid
	// message. Server returns an error in response message (in status).
	if x.err != nil && !errors.Is(x.err, io.EOF) {
		return nil, x.err
	}

	if x.err = x.stream.Close(); x.err != nil {
		return nil, x.err
	}

	if x.err = x.client.processResponse(&x.respV2); x.err != nil {
		return nil, x.err
	}

	const fieldID = "ID"

	idV2 := x.respV2.GetBody().GetObjectID()
	if idV2 == nil {
		return nil, newErrMissingResponseField(fieldID)
	}

	x.err = x.res.obj.ReadFromV2(*idV2)
	if x.err != nil {
		x.err = newErrInvalidResponseField(fieldID, x.err)
	}

	return &x.res, nil
}

// ObjectPutInit initiates writing an object through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit recording is done using the ObjectWriter.
// Exactly one return value is non-nil. Resulting writer must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
//
// Returns errors:
//   - [ErrMissingSigner]
func (c *Client) ObjectPutInit(ctx context.Context, prm PrmObjectPutInit) (*ObjectWriter, error) {
	var w ObjectWriter

	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	stream, err := rpcapi.PutObject(&c.c, &w.respV2, client.WithContext(ctx))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("open stream: %w", err)
	}

	w.signer = signer
	w.cancelCtxStream = cancel
	w.client = c
	w.stream = stream
	w.partInit.SetCopiesNumber(prm.copyNum)
	w.req.SetBody(new(v2object.PutRequestBody))
	c.prepareRequest(&w.req, &prm.meta)

	return &w, nil
}

type objectWriter struct {
	context context.Context
	client  *Client
}

func (x *objectWriter) InitDataStream(header object.Object) (io.Writer, error) {
	var prm PrmObjectPutInit

	stream, err := x.client.ObjectPutInit(x.context, prm)
	if err != nil {
		return nil, fmt.Errorf("init object stream: %w", err)
	}

	if stream.WriteHeader(header) {
		return &payloadWriter{
			stream: stream,
		}, nil
	}

	_, err = stream.Close()
	if err != nil {
		return nil, err
	}

	return nil, errors.New("unexpected error")
}

type payloadWriter struct {
	stream *ObjectWriter
}

func (x *payloadWriter) Write(p []byte) (int, error) {
	if !x.stream.WritePayloadChunk(p) {
		return 0, x.Close()
	}

	return len(p), nil
}

func (x *payloadWriter) Close() error {
	_, err := x.stream.Close()
	if err != nil {
		return err
	}

	return nil
}

// CreateObject creates new NeoFS object with given payload data and stores it
// in specified container of the NeoFS network using provided Client connection.
// The object is created on behalf of provided neofscrypto.Signer, and owned by
// the specified user.ID.
//
// In terms of NeoFS, parameterized neofscrypto.Signer represents object owner,
// object signer and request sender. Container SHOULD be public-write or sender
// SHOULD have corresponding rights.
//
// Client connection MUST be opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// Notice: This API is EXPERIMENTAL and is planned to be replaced/changed in the
// future. Be ready to refactor your code regarding imports and call mechanics,
// in essence the operation will not change.
func CreateObject(ctx context.Context, cli *Client, signer neofscrypto.Signer, cnr cid.ID, owner user.ID, data io.Reader, attributes ...string) (oid.ID, error) {
	s, err := NewDataSlicer(ctx, cli, signer, cnr, owner)
	if err != nil {
		return oid.ID{}, err
	}

	return s.Slice(data, attributes...)
}

// NewDataSlicer creates slicer.Slicer that saves data in the NeoFS network
// through provided Client. The data is packaged into NeoFS objects stored in
// the specified container. Provided signer is being used to sign the resulting
// objects as a system requirement. Produced objects are owned by the
// parameterized NeoFS user.
//
// Notice: This API is EXPERIMENTAL and is planned to be replaced/changed in the
// future. Be ready to refactor your code regarding imports and call mechanics,
// in essence the operation will not change.
func NewDataSlicer(ctx context.Context, cli *Client, signer neofscrypto.Signer, cnr cid.ID, owner user.ID) (*slicer.Slicer, error) {
	netInfo, err := cli.NetworkInfo(ctx, PrmNetworkInfo{})
	if err != nil {
		return nil, fmt.Errorf("read current network info: %w", err)
	}

	var opts slicer.Options
	opts.SetObjectPayloadLimit(netInfo.MaxObjectSize())
	opts.SetCurrentNeoFSEpoch(netInfo.CurrentEpoch())
	if !netInfo.HomomorphicHashingDisabled() {
		opts.CalculateHomomorphicChecksum()
	}

	return slicer.New(signer, cnr, owner, &objectWriter{
		context: ctx,
		client:  cli,
	}, opts), nil
}
