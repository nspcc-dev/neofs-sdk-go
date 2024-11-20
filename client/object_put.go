package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

var (
	// ErrNoSessionExplicitly is a special error to show auto-session is disabled.
	ErrNoSessionExplicitly = errors.New("session was removed explicitly")
)

// used part of [protoobject.ObjectService_PutClient] simplifying test
// implementations.
type putObjectStream interface {
	// Send writes next message with the object part to the stream. No error does
	// not guarantee delivery to the server. Send returns [io.EOF] after the server
	// sent the response and gracefully finished the stream: the result can be
	// accessed via CloseAndRecv. Any other error means stream abort.
	Send(*protoobject.PutRequest) error
	// CloseAndRecv finishes the stream and reads response from the server.
	CloseAndRecv() (*protoobject.PutResponse, error)
}

// shortStatisticCallback is a shorter version of [stat.OperationCallback] which is calling from [client.Client].
// The difference is the client already know some info about itself. Despite it the client doesn't know
// duration and error from writer/reader.
type shortStatisticCallback func(err error)

// PrmObjectPutInit groups parameters of ObjectPutInit operation.
type PrmObjectPutInit struct {
	sessionContainer

	copyNum uint32
}

// SetCopiesNumber sets the minimal number of copies (out of the number specified by container placement policy) for
// the object PUT operation to succeed. This means that object operation will return with successful status even before
// container placement policy is completely satisfied.
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
type ObjectWriter interface {
	io.WriteCloser
	GetResult() ResObjectPut
}

// DefaultObjectWriter implements [ObjectWriter].
//
// Must be initialized using [Client.ObjectPutInit], any other usage is unsafe.
type DefaultObjectWriter struct {
	cancelCtxStream context.CancelFunc

	client           *Client
	stream           putObjectStream
	singleMsgTimeout time.Duration
	streamClosed     bool

	signer neofscrypto.Signer
	res    ResObjectPut
	err    error

	chunkCalled bool

	req       v2object.PutRequest
	partInit  v2object.PutObjectPartInit
	partChunk v2object.PutObjectPartChunk

	statisticCallback shortStatisticCallback

	buf              []byte
	bufCleanCallback func()
}

// WithBearerToken attaches bearer token to be used for the operation.
// Should be called once before any writing steps.
func (x *PrmObjectPutInit) WithBearerToken(t bearer.Token) {
	var v2token acl.BearerToken
	t.WriteToV2(&v2token)
	x.meta.SetBearerToken(&v2token)
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

// writeHeader writes header of the object. Result means success.
// Failure reason can be received via [DefaultObjectWriter.Close].
func (x *DefaultObjectWriter) writeHeader(hdr object.Object) error {
	v2Hdr := hdr.ToV2()

	x.partInit.SetObjectID(v2Hdr.GetObjectID())
	x.partInit.SetHeader(v2Hdr.GetHeader())
	x.partInit.SetSignature(v2Hdr.GetSignature())

	x.req.GetBody().SetObjectPart(&x.partInit)
	x.req.SetVerificationHeader(nil)

	x.err = signServiceMessage(x.signer, &x.req, x.buf)
	if x.err != nil {
		x.err = fmt.Errorf("sign message: %w", x.err)
		return x.err
	}

	x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
		return x.stream.Send(x.req.ToGRPCMessage().(*protoobject.PutRequest))
	})
	return x.err
}

// WritePayloadChunk writes chunk of the object payload. Result means success.
// Failure reason can be received via [DefaultObjectWriter.Close].
func (x *DefaultObjectWriter) Write(chunk []byte) (n int, err error) {
	if !x.chunkCalled {
		x.chunkCalled = true
		x.req.GetBody().SetObjectPart(&x.partChunk)
	}

	var writtenBytes int

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

		x.err = signServiceMessage(x.signer, &x.req, x.buf)
		if x.err != nil {
			x.err = fmt.Errorf("sign message: %w", x.err)
			return writtenBytes, x.err
		}

		x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
			return x.stream.Send(x.req.ToGRPCMessage().(*protoobject.PutRequest))
		})
		if x.err != nil {
			if errors.Is(x.err, io.EOF) {
				var resp *protoobject.PutResponse
				x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
					var err error
					resp, err = x.stream.CloseAndRecv()
					return err
				})
				if x.err != nil {
					return writtenBytes, x.err
				}
				var respV2 v2object.PutResponse
				if x.err = respV2.FromGRPCMessage(resp); x.err != nil {
					return writtenBytes, x.err
				}
				x.err = x.client.processResponse(&respV2)
				x.streamClosed = true
				x.cancelCtxStream()
			}

			return writtenBytes, x.err
		}

		writtenBytes += len(chunk[:ln])
		chunk = chunk[ln:]
	}

	return writtenBytes, nil
}

// Close ends writing the object and returns the result of the operation
// along with the final results. Must be called after using the [DefaultObjectWriter].
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
func (x *DefaultObjectWriter) Close() error {
	if x.statisticCallback != nil {
		defer func() {
			x.statisticCallback(x.err)
		}()
	}

	if x.bufCleanCallback != nil {
		defer x.bufCleanCallback()
	}

	if x.streamClosed {
		return nil
	}

	defer x.cancelCtxStream()

	// Ignore io.EOF error, because it is expected error for client-side
	// stream termination by the server. E.g. when stream contains invalid
	// message. Server returns an error in response message (in status).
	if x.err != nil && !errors.Is(x.err, io.EOF) {
		return x.err
	}

	var resp *protoobject.PutResponse
	if x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
		var err error
		resp, err = x.stream.CloseAndRecv()
		return err
	}); x.err != nil {
		return x.err
	}
	var respV2 v2object.PutResponse
	if x.err = respV2.FromGRPCMessage(resp); x.err != nil {
		return x.err
	}

	if x.err = x.client.processResponse(&respV2); x.err != nil {
		return x.err
	}

	const fieldID = "ID"

	idV2 := respV2.GetBody().GetObjectID()
	if idV2 == nil {
		x.err = newErrMissingResponseField(fieldID)
		return x.err
	}

	x.err = x.res.obj.ReadFromV2(*idV2)
	if x.err != nil {
		x.err = newErrInvalidResponseField(fieldID, x.err)
	}

	return x.err
}

// GetResult returns the put operation result.
func (x *DefaultObjectWriter) GetResult() ResObjectPut {
	return x.res
}

// ObjectPutInit initiates writing an object through a remote server using NeoFS API protocol.
// Header length is limited to [object.MaxHeaderLen].
//
// The call only opens the transmission channel, explicit recording is done using the [ObjectWriter].
// Exactly one return value is non-nil. Resulting writer must be finally closed.
//
// Context is required and must not be nil. It will be used for network communication for the whole object transmission,
// including put init (this method) and subsequent object payload writes via ObjectWriter.
//
// Signer is required and must not be nil. The operation is executed on behalf of
// the account corresponding to the specified Signer, which is taken into account, in particular, for access control.
//
// Returns errors:
//   - [ErrMissingSigner]
func (c *Client) ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm PrmObjectPutInit) (ObjectWriter, error) {
	var err error
	defer func() {
		c.sendStatistic(stat.MethodObjectPut, err)()
	}()
	var w DefaultObjectWriter
	w.statisticCallback = func(err error) {
		c.sendStatistic(stat.MethodObjectPutStream, err)()
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}

	ctx, cancel := context.WithCancel(ctx)
	stream, err := c.object.Put(ctx)
	if err != nil {
		cancel()
		err = fmt.Errorf("open stream: %w", err)
		return nil, err
	}

	buf := c.buffers.Get().(*[]byte)
	w.buf = *buf
	w.bufCleanCallback = func() {
		c.buffers.Put(buf)
	}

	w.signer = signer
	w.cancelCtxStream = cancel
	w.client = c
	w.stream = stream
	w.singleMsgTimeout = c.streamTimeout
	w.partInit.SetCopiesNumber(prm.copyNum)
	w.req.SetBody(new(v2object.PutRequestBody))
	c.prepareRequest(&w.req, &prm.meta)

	if err = w.writeHeader(hdr); err != nil {
		_ = w.Close()
		err = fmt.Errorf("header write: %w", err)
		return nil, err
	}

	return &w, nil
}
