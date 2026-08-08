package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	sync "sync"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	igrpc "github.com/nspcc-dev/neofs-sdk-go/internal/grpc"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	grpcprotobuf "github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/mem"
	"google.golang.org/protobuf/encoding/protowire"
)

var (
	// ErrNoSessionExplicitly is a special error to show auto-session is disabled.
	ErrNoSessionExplicitly = errors.New("session was removed explicitly")
)

// maxChunkLen restricts maximum byte length of the chunk
// transmitted in a single stream message. It depends on
// server settings and other message fields, but for now
// we simply assume that 3MB is large enough to reduce the
// number of messages, and not to exceed the limit
// (4MB by default for gRPC servers).
const maxChunkLen = 3 << 20

// used part of [protoobject.ObjectService_PutClient] simplifying test
// implementations.
type putObjectStream interface {
	// SendMsg is a part of [grpc.ClientStream] interface.
	SendMsg(any) error
	// CloseAndRecv finishes the stream and reads response from the server.
	CloseAndRecv() (*protoobject.PutResponse, error)
}

// shortStatisticCallback is a shorter version of [stat.OperationCallback] which is calling from [client.Client].
// The difference is the client already know some info about itself. Despite it the client doesn't know
// duration and error from writer/reader.
type shortStatisticCallback func(dur time.Duration, err error)

// PrmObjectPutInit groups parameters of ObjectPutInit operation.
type PrmObjectPutInit struct {
	prmCommonMeta
	sessionContainer
	bearerToken *bearer.Token
	local       bool
}

// SetCopiesNumber sets the minimal number of copies (out of the number specified by container placement policy) for
// the object PUT operation to succeed. This means that object operation will return with successful status even before
// container placement policy is completely satisfied.
//
// Deprecated: Specify max replicas in container's initial placement policy
// instead. This parameter no longer has an effect.
func (x *PrmObjectPutInit) SetCopiesNumber(uint32) {}

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
	io.ReaderFrom
	GetResult() ResObjectPut
}

// DefaultObjectWriter implements [ObjectWriter].
//
// Must be initialized using [Client.ObjectPutInit], any other usage is unsafe.
type DefaultObjectWriter struct {
	cancelCtxStream context.CancelFunc

	stream           putObjectStream
	singleMsgTimeout time.Duration
	streamClosed     bool

	signer            neofscrypto.Signer
	shouldSignRequest bool
	res               ResObjectPut
	err               error

	apiVersion *refs.Version
	opts       PrmObjectPutInit

	statisticCallback shortStatisticCallback
	startTime         time.Time // if statisticCallback is set only

	payloadSizeFromHeader uint64

	bufferPool *sync.Pool

	versionLen        int
	xHeadersLen       int
	sessionV1TokenLen int
	sessionV1TokenMsg *protosession.SessionToken
	bearerTokenLen    int
	bearerTokenMsg    *protoacl.BearerToken
	sessionV2TokenLen int
	sessionV2TokenMsg *protosession.SessionTokenV2
	metaHdrLen        int
	// === group of vars defined if shouldSignRequest only === //
	signerPublicKeyBinary             []byte // corresponds to signer
	metaHeaderSignature               neofscrypto.Signature
	originVerificationHeaderSignature neofscrypto.Signature
	needSignOrigin                    bool
	// === //
}

// WithBearerToken attaches bearer token to be used for the operation.
// Should be called once before any writing steps.
func (x *PrmObjectPutInit) WithBearerToken(t bearer.Token) {
	x.bearerToken = &t
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectPutInit) MarkLocal() {
	x.local = true
}

// sendRequest transmits signed req to the server recording the result in x.err.
func (x *DefaultObjectWriter) sendRequest(req any, reqMemBuf *igrpc.MemBuffer) error {
	var sent bool
	x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
		sent = true
		return x.stream.SendMsg(req)
	})
	if !sent && reqMemBuf != nil {
		reqMemBuf.Free()
	}
	if x.err != nil && errors.Is(x.err, io.EOF) {
		var resp *protoobject.PutResponse
		x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
			var err error
			resp, err = x.stream.CloseAndRecv()
			return err
		})
		if x.err != nil {
			return x.err
		}

		x.err = apistatus.ToError(resp.GetMetaHeader().GetStatus())

		x.streamClosed = true
		x.cancelCtxStream()
	}
	return x.err
}

// writeHeader writes header of the object. Result means success.
// Failure reason can be received via [DefaultObjectWriter.Close].
func (x *DefaultObjectWriter) writeHeader(hdr object.Object) error {
	mh := hdr.ProtoMessage()
	mh.Payload = nil

	if x.opts.session != nil {
		x.sessionV1TokenMsg = x.opts.session.ProtoMessage()
		x.sessionV1TokenLen = x.sessionV1TokenMsg.MarshaledSize()
	}

	if x.opts.bearerToken != nil {
		x.bearerTokenMsg = x.opts.bearerToken.ProtoMessage()
		x.bearerTokenLen = x.bearerTokenMsg.MarshaledSize()
	}

	if x.opts.sessionV2 != nil {
		x.sessionV2TokenMsg = x.opts.sessionV2.ProtoMessage()
		x.sessionV2TokenLen = x.sessionV2TokenMsg.MarshaledSize()
	}

	initFldLen := mh.MarshaledSize()
	bodyLen := calculatePutObjectHeadingRequestBodyLength(initFldLen)

	x.versionLen = x.apiVersion.MarshaledSize()
	x.xHeadersLen = calculateRequestXHeadersLength(x.opts.xHeaders)

	x.metaHdrLen = calculateRequestMetaHeaderFieldLengths(x.versionLen, x.opts.local, x.xHeadersLen, x.sessionV1TokenLen, x.bearerTokenLen, x.sessionV2TokenLen)

	bodyWithMetaHdrLen := neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestBody, bodyLen)
	bodyWithMetaHdrLen += neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestMetaHeader, x.metaHdrLen)

	// acquire buffer for body + meta header
	bufItem := x.bufferPool.Get().(*[]byte)
	var reqMemBuf *igrpc.MemBuffer
	var buf []byte
	if len(*bufItem) >= bodyWithMetaHdrLen {
		// TODO: this is an extra alloc which can be avoided with pool of fix-size buffers. TBD within https://github.com/nspcc-dev/neofs-sdk-go/issues/666
		reqMemBuf = igrpc.NewMemBuffer(bufItem, x.bufferPool)
		buf = *bufItem
	} else {
		x.bufferPool.Put(bufItem)
		buf = make([]byte, bodyWithMetaHdrLen)
	}

	// encode body
	off := writePutObjectHeadingRequestBody(buf, bodyLen, initFldLen, mh)

	// memorize body for signing
	signedBody := buf[off-bodyLen : off]

	// encode meta header
	off += writeRequestMetaHeader(buf[off:], x.metaHdrLen, x.versionLen, x.apiVersion, x.opts.local, x.opts.xHeaders, x.sessionV1TokenLen, x.sessionV1TokenMsg, x.bearerTokenLen, x.bearerTokenMsg, x.sessionV2TokenLen, x.sessionV2TokenMsg)

	var reqBuffers mem.BufferSlice
	if x.shouldSignRequest {
		// append verification header
		x.needSignOrigin = needsOriginSig(x.apiVersion)
		x.signerPublicKeyBinary = neofscrypto.PublicKeyBytes(x.signer.Public())

		bodySigVal, metaHdrSigVal, originVerifHdrSigVal, err := signRequestParts(x.signer, signedBody, buf[off-x.metaHdrLen:off], x.needSignOrigin)
		if err != nil {
			if reqMemBuf != nil {
				reqMemBuf.Free()
			}
			x.err = err
			return err
		}

		scheme := x.signer.Scheme()
		bodySig := neofscrypto.NewSignatureFromRawKey(scheme, x.signerPublicKeyBinary, bodySigVal)
		x.metaHeaderSignature = neofscrypto.NewSignatureFromRawKey(scheme, x.signerPublicKeyBinary, metaHdrSigVal)
		if x.needSignOrigin {
			x.originVerificationHeaderSignature = neofscrypto.NewSignatureFromRawKey(scheme, x.signerPublicKeyBinary, originVerifHdrSigVal)
		}

		reqBuffers = appendVerificationHeaderSignatures(reqMemBuf, buf, bodyWithMetaHdrLen, bodySig, x.metaHeaderSignature, x.originVerificationHeaderSignature)
	} else {
		reqSliceBuf := mem.SliceBuffer(buf[:off])
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqSliceBuf
			reqBuffers = mem.BufferSlice{reqMemBuf}
		} else {
			reqBuffers = mem.BufferSlice{reqSliceBuf}
		}
	}

	var sent bool
	x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
		sent = true
		return x.stream.SendMsg(reqBuffers)
	})
	if !sent && reqMemBuf != nil {
		reqMemBuf.Free()
	}

	return x.err
}

// WritePayloadChunk writes chunk of the object payload. Result means success.
// Failure reason can be received via [DefaultObjectWriter.Close].
func (x *DefaultObjectWriter) Write(chunk []byte) (n int, err error) {
	var writtenBytes int

	for ln := len(chunk); ln > 0; ln = len(chunk) {
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
		bodyLen := calculatePutObjectPayloadChunkRequestBodyLength(ln)

		bodyWithMetaHdrLen := neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestBody, bodyLen)
		bodyWithMetaHdrLen += neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestMetaHeader, x.metaHdrLen)

		// acquire buffer for body + meta header
		bufItem := x.bufferPool.Get().(*[]byte)
		var reqMemBuf *igrpc.MemBuffer
		var buf []byte
		if len(*bufItem) >= bodyWithMetaHdrLen {
			// TODO: this is an extra alloc which can be avoided with pool of fix-size buffers. TBD within https://github.com/nspcc-dev/neofs-sdk-go/issues/666
			reqMemBuf = igrpc.NewMemBuffer(bufItem, x.bufferPool)
			buf = *bufItem
		} else {
			x.bufferPool.Put(bufItem)
			buf = make([]byte, bodyWithMetaHdrLen)
		}

		// encode body
		off := writePutObjectPayloadChunkRequestBody(buf, bodyLen, chunk[:ln])

		bodyTo := off

		// encode meta header
		off += writeRequestMetaHeader(buf[off:], x.metaHdrLen, x.versionLen, x.apiVersion, x.opts.local, x.opts.xHeaders, x.sessionV1TokenLen, x.sessionV1TokenMsg, x.bearerTokenLen, x.bearerTokenMsg, x.sessionV2TokenLen, x.sessionV2TokenMsg)

		var reqBuffers mem.BufferSlice
		if x.shouldSignRequest {
			// append verification header
			bodySigVal, err := x.signer.Sign(buf[bodyTo-bodyLen : bodyTo])
			if err != nil {
				if reqMemBuf != nil {
					reqMemBuf.Free()
				}
				x.err = fmt.Errorf("sign body: %w", err)
				return writtenBytes, x.err
			}

			bodySig := neofscrypto.NewSignatureFromRawKey(x.signer.Scheme(), x.signerPublicKeyBinary, bodySigVal)

			reqBuffers = appendVerificationHeaderSignatures(reqMemBuf, buf, off, bodySig, x.metaHeaderSignature, x.originVerificationHeaderSignature)
		} else {
			reqSliceBuf := mem.SliceBuffer(buf[:off])
			if reqMemBuf != nil {
				reqMemBuf.SliceBuffer = reqSliceBuf
				reqBuffers = mem.BufferSlice{reqMemBuf}
			} else {
				reqBuffers = mem.BufferSlice{reqSliceBuf}
			}
		}

		if err = x.sendRequest(reqBuffers, reqMemBuf); err != nil {
			return writtenBytes, err
		}
		if x.streamClosed {
			// server gracefully finished the stream ahead of the full payload
			return writtenBytes, nil
		}

		writtenBytes += len(chunk[:ln])
		chunk = chunk[ln:]
	}

	return writtenBytes, nil
}

// ReadFrom reads r until EOF and writes object payload.
// Failure reason can be received via [DefaultObjectWriter.Close].
// ReadFrom implements [io.ReaderFrom].
func (x *DefaultObjectWriter) ReadFrom(r io.Reader) (int64, error) {
	var maxReadChunkLen = maxChunkLen
	// For the case where the object size is less than maxChunkLen, according to the object header.
	if x.payloadSizeFromHeader > 0 && x.payloadSizeFromHeader < uint64(maxReadChunkLen) {
		maxReadChunkLen = int(x.payloadSizeFromHeader)
	}

	maxChunkVarlen := protowire.SizeVarint(uint64(maxReadChunkLen))
	maxBodyVarlen := protowire.SizeVarint(uint64(1 + maxChunkVarlen + maxReadChunkLen)) // 1 for grpcprotobuf.TagBytes1
	chunkOff := 1 + maxBodyVarlen + 1 + maxChunkVarlen                                  // first 1 for grpcprotobuf.TagBytes1, second for grpcprotobuf.TagBytes2

	maxBodyWithMetaHdrLen := chunkOff + maxReadChunkLen
	maxBodyWithMetaHdrLen += neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestMetaHeader, x.metaHdrLen)

	var writtenBytes int64

	for {
		// acquire buffer for body + meta header
		bufItem := x.bufferPool.Get().(*[]byte)
		var reqMemBuf *igrpc.MemBuffer
		var buf []byte
		if len(*bufItem) >= maxBodyWithMetaHdrLen {
			// TODO: this is an extra alloc which can be avoided with pool of fix-size buffers. TBD within https://github.com/nspcc-dev/neofs-sdk-go/issues/666
			reqMemBuf = igrpc.NewMemBuffer(bufItem, x.bufferPool)
			buf = *bufItem
		} else {
			x.bufferPool.Put(bufItem)
			buf = make([]byte, maxBodyWithMetaHdrLen)
		}

		actualRead, err := readFull(r, buf[chunkOff:][:maxReadChunkLen])
		if actualRead > 0 {
			chunkVarlen := protowire.SizeVarint(uint64(actualRead))
			chunkFldOff := chunkOff - chunkVarlen - 1

			buf[chunkFldOff] = grpcprotobuf.TagBytes2 // chunk
			binary.PutUvarint(buf[chunkFldOff+1:], uint64(actualRead))

			bodyLen := 1 + chunkVarlen + actualRead
			bodyVarlen := protowire.SizeVarint(uint64(bodyLen))
			bodyOff := chunkFldOff + -bodyVarlen - 1

			buf[bodyOff] = grpcprotobuf.TagBytes1 // body
			binary.PutUvarint(buf[bodyOff+1:], uint64(bodyLen))

			off := chunkOff + actualRead

			// encode meta header
			off += writeRequestMetaHeader(buf[off:], x.metaHdrLen, x.versionLen, x.apiVersion, x.opts.local, x.opts.xHeaders, x.sessionV1TokenLen, x.sessionV1TokenMsg, x.bearerTokenLen, x.bearerTokenMsg, x.sessionV2TokenLen, x.sessionV2TokenMsg)

			var reqBuffers mem.BufferSlice
			if x.shouldSignRequest {
				// append verification header
				bodySigVal, err := x.signer.Sign(buf[chunkFldOff : chunkOff+actualRead])
				if err != nil {
					if reqMemBuf != nil {
						reqMemBuf.Free()
					}
					return writtenBytes, fmt.Errorf("sign body: %w", err)
				}

				bodySig := neofscrypto.NewSignatureFromRawKey(x.signer.Scheme(), x.signerPublicKeyBinary, bodySigVal)

				reqBuffers = appendVerificationHeaderSignatures(reqMemBuf, buf[bodyOff:], off-bodyOff, bodySig, x.metaHeaderSignature, x.originVerificationHeaderSignature)
			} else {
				reqSliceBuf := mem.SliceBuffer(buf[bodyOff:off])
				if reqMemBuf != nil {
					reqMemBuf.SliceBuffer = reqSliceBuf
					reqBuffers = mem.BufferSlice{reqMemBuf}
				} else {
					reqBuffers = mem.BufferSlice{reqSliceBuf}
				}
			}

			if writeErr := x.sendRequest(reqBuffers, reqMemBuf); writeErr != nil {
				return writtenBytes, fmt.Errorf("aaaa: %w", writeErr)
			}

			if x.streamClosed {
				return writtenBytes, io.ErrShortWrite
			}

			writtenBytes += int64(actualRead)
		} else if reqMemBuf != nil {
			reqMemBuf.Free()
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return writtenBytes, nil
			}

			return writtenBytes, err
		}
	}
}

// readFull reads from r until b is full or r fails.
func readFull(r io.Reader, b []byte) (int, error) {
	var n int

	for n < len(b) {
		read, err := r.Read(b[n:])
		n += read

		if err != nil {
			return n, err
		}
	}

	return n, nil
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
//   - [apistatus.ErrIncomplete]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectLocked]
//   - [apistatus.ErrLockNonRegularObject]
//   - [apistatus.ErrQuotaExceeded]
//   - [apistatus.ErrSessionTokenNotFound]
//   - [apistatus.ErrSessionTokenExpired]
func (x *DefaultObjectWriter) Close() error {
	if x.statisticCallback != nil {
		defer func() {
			x.statisticCallback(time.Since(x.startTime), x.err)
		}()
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

	var statusError error
	if x.err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); x.err != nil {
		if !errors.Is(x.err, apistatus.ErrIncomplete) {
			return x.err
		}
		statusError = x.err
	}

	const fieldID = "ID"

	idV2 := resp.GetBody().GetObjectId()
	if idV2 == nil {
		x.err = newErrMissingResponseField(fieldID)
		return x.err
	}

	x.err = x.res.obj.FromProtoMessage(idV2)
	if x.err != nil {
		x.err = newErrInvalidResponseField(fieldID, x.err)
	}
	if x.err == nil {
		x.err = statusError
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
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectPut, time.Since(startTime), err)
		}()
	}
	var w DefaultObjectWriter
	if c.prm.statisticCallback != nil {
		w.startTime = time.Now()
		w.statisticCallback = func(dur time.Duration, err error) {
			c.sendStatistic(stat.MethodObjectPutStream, dur, err)
		}
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}
	if prm.session != nil && prm.sessionV2 != nil {
		return nil, errSessionTokenBothVersionsSet
	}

	ctx, cancel := context.WithCancel(ctx)
	stream, err := openStream(ctx, c.conn, clientStreamDesc, protoobject.ObjectService_Put_FullMethodName)
	if err != nil {
		cancel()
		err = fmt.Errorf("open stream: %w", err)
		return nil, err
	}

	w.bufferPool = c.buffers
	w.apiVersion = c.apiVersion
	w.signer = signer
	w.shouldSignRequest = !prm.local || !c.skipSignatureForLocalRequests
	w.cancelCtxStream = cancel
	w.stream = &grpc.GenericClientStream[protoobject.PutRequest, protoobject.PutResponse]{ClientStream: stream}
	w.singleMsgTimeout = c.streamTimeout
	w.opts = prm
	w.payloadSizeFromHeader = hdr.PayloadSize()
	if err = w.writeHeader(hdr); err != nil {
		_ = w.Close()
		err = fmt.Errorf("header write: %w", err)
		return nil, err
	}

	return &w, nil
}
