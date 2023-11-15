package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protowire"
)

const (
	fieldNumObjectPutBodyInitPart = 1
	fieldNumObjectPutBodyChunk    = 2
)

var (
	// ErrNoSessionExplicitly is a special error to show auto-session is disabled.
	ErrNoSessionExplicitly = errors.New("session was removed explicitly")
)

var (
	// special variable for test purposes only, to overwrite real RPC calls.
	rpcAPIPutObject = func(cli *client.Client, r *v2object.PutResponse, o ...client.CallOption) (objectWriter, error) {
		return rpcapi.PutObject(cli, r, o...)
	}
)

type objectWriter interface {
	Write(*v2object.PutRequest) error
	Close() error
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

	client       *Client
	stream       objectWriter
	streamClosed bool

	signer neofscrypto.Signer
	res    ResObjectPut
	err    error

	chunkCalled bool

	respV2    v2object.PutResponse
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

	x.err = x.stream.Write(&x.req)
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

		x.err = x.stream.Write(&x.req)
		if x.err != nil {
			if errors.Is(x.err, io.EOF) {
				_ = x.stream.Close()
				x.err = x.client.processResponse(&x.respV2)
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

	if x.err = x.stream.Close(); x.err != nil {
		return x.err
	}

	if x.err = x.client.processResponse(&x.respV2); x.err != nil {
		return x.err
	}

	const fieldID = "ID"

	idV2 := x.respV2.GetBody().GetObjectID()
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
	stream, err := rpcAPIPutObject(&c.c, &w.respV2, client.WithContext(ctx))
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

// CopyBinaryObjectOptions groups [Client.CopyBinaryObject] options tuning the
// default behavior.
type CopyBinaryObjectOptions struct {
	useSingleMsgBuffer bool
}

// UseSingleMessageBuffer enables usage of the single buffer to encode binary
// messages transmitted over the wire. This reduces memory allocations, but may
// also affect message delivery in other ways, therefore, enable the option only
// if its details are familiar.
func (x *CopyBinaryObjectOptions) UseSingleMessageBuffer() {
	x.useSingleMsgBuffer = true
}

// CopyBinaryObject copies binary-encoded NeoFS object from the given
// [io.ReadSeeker] into local storage of the remote node. Provided
// [neofscrypto.Signer] is used to sign produced requests. See
// CopyBinaryObjectOptions to tune the behavior.
//
// Object MUST be encoded in compliance with Protocol Buffers V3 format. The
// binary object MUST contain only fields defined in the current version of the
// protocol. Any of the fields MAY be missing. Object ID, signature and header
// fields MUST follow in exactly this order. Object payload MAY be located
// anywhere (usually after headers): for example, it MAY split other fields
// without disturbing their order.
//
// NOTE: CopyBinaryObject does not seek src to start after the call. If it is
// needed, do not forget to call Seek.
func (c *Client) CopyBinaryObject(ctx context.Context, src io.ReadSeeker, signer neofscrypto.Signer, opts CopyBinaryObjectOptions) error {
	var retErr error
	defer func() {
		c.sendStatistic(stat.MethodObjectPut, retErr)()
	}()

	if signer == nil {
		// note that we don't stat this error
		return ErrMissingSigner
	}

	var resp v2object.PutResponse
	var callOpts []client.CallOption

	if opts.useSingleMsgBuffer {
		callOpts = []client.CallOption{client.SyncWrite()}
	}

	stream, err := rpcapi.PutObjectBinary(ctx, &c.c, &resp, callOpts...)
	if err != nil {
		return fmt.Errorf("init ObjectService.Put RPC client stream: %w", err)
	}

	err = streamBinaryObject(stream, src, signer, opts.useSingleMsgBuffer)
	if err != nil {
		return err
	}

	err = stream.Close()
	if err != nil {
		return fmt.Errorf("close ObjectService.Put RPC client stream: %w", err)
	}

	err = c.processResponse(&resp)
	if err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	return nil
}

// streamBinaryObject performs [Client.CopyBinaryObjectToNode] main job, but
// abstracts object stream so that it can be customized.
//
// TODO: support shuffled object ID, signature and header too (now supports only
// payload field movement)
//
// TODO: test performance when src implements io.ReaderAt. For os.File, we'll
// possibly reduce number of syscalls
func streamBinaryObject(dst binaryMessageWriter, src io.ReadSeeker, signer neofscrypto.Signer, singleBuf bool) error {
	binObjInfo, err := scanBinaryObject(src)
	if err != nil {
		return fmt.Errorf("scan binary object: %w", err)
	}

	// now we prepare buffer(s) for encoding and transmission of all stream
	// messages. Brief desc of messages' binary format is described below. For
	// understanding see https://protobuf.dev/programming-guides/encoding/.
	//
	// |XXX| BODY_SIGNATURE|XXX|BODY_TAG|...BODY_SIZE...|BODY_FIELD_TAG|...BODY_FIELD_SIZE...|...BODY_FIELD...|
	// where:
	//  * XXX are static parts remaining unchanged during the whole stream (e.g. meta header)
	//  * BODY_TAG is fixed and shown as a start point of dynamic payload
	//  * BODY_FIELD is either initial part or object payload chunk
	//  * BODY_FIELD_TAG carries type and field number of BODY_FIELD
	//  * BODY_FIELD_SIZE and, respectively, BODY_SIZE depend on BODY_FIELD
	//  * BODY_SIGNATURE has static offset/size while value is re-calculated each time
	//    for the data starting from BODY_FIELD_TAG
	const ttl = 1
	staticFields, err := initObjectPutRequestStaticFields(signer, ttl, version.Current())
	if err != nil {
		return err
	}

	headerSize := binObjInfo.fullSize
	if binObjInfo.payloadFieldOffset >= 0 {
		headerSize = binObjInfo.fullSize - (binObjInfo.payloadOffset - binObjInfo.payloadFieldOffset + binObjInfo.payloadSize)
	}
	// object ID + object signature + object header are encoded the same as in
	// object (same field numbers and wire types)
	initPartFieldSize := protowire.SizeTag(fieldNumObjectPutBodyInitPart) + protowire.SizeBytes(headerSize)

	// determine what is bigger: initial message with object header message or max
	// chunk message. Remember that any message is limited by 4MB. But unlike
	// payload, initial message could not be chunked, so, if the message is bigger
	// than the limit, we try to send it entirely anyway. If it works, bigger chunks
	// can be sent then which is good: in some cases fewer messages with chunks may
	// be needed.

	// since we don't know length of body field in advance, we use some sane limit
	// (4GB is definitely enough)
	const embeddedMessageSizeTagLimit = binary.MaxVarintLen32
	bufSize := staticFields.fullSize + 2*embeddedMessageSizeTagLimit // one for body field size and one for the whole body size

	chunkPartFieldTagSize := protowire.SizeTag(fieldNumObjectPutBodyChunk)
	fullPayloadChunkFieldSize := chunkPartFieldTagSize + protowire.SizeBytes(binObjInfo.payloadSize)
	var maxChunkSize int

	if initPartFieldSize >= fullPayloadChunkFieldSize || bufSize+initPartFieldSize >= defaultBufferSize {
		maxChunkSize = initPartFieldSize - chunkPartFieldTagSize - 2*embeddedMessageSizeTagLimit
		bufSize += initPartFieldSize
	} else if bufSize+fullPayloadChunkFieldSize < defaultBufferSize {
		maxChunkSize = binObjInfo.payloadSize
		bufSize += fullPayloadChunkFieldSize
	} else {
		maxChunkSize = defaultBufferSize - bufSize - chunkPartFieldTagSize - 2*embeddedMessageSizeTagLimit
		bufSize = defaultBufferSize
	}

	type message struct {
		buffer             []byte
		bodyFieldTagOffset int
		// function to be called when request body is completely read for this message
		signBody func(body []byte) error
	}

	// buffer has exact length to encode message into
	var prepareInitMsg func() (msg message, err error)
	// buffer has static fields and enough capacity for the chunk
	var prepareChunkMsg func(chunkSize int) (msg message, err error)

	if singleBuf {
		var buf []byte
		var signBody func([]byte) error

		prepareInitMsg = func() (message, error) {
			// TODO: support custom allocation (e.g. use sync.Pool)
			b := make([]byte, 0, bufSize)

			var err error
			signBody, err = staticFields.appendTo(b)
			if err != nil {
				return message{}, err
			}

			b = b[:staticFields.fullSize]
			buf = b

			b = protowire.AppendVarint(b, uint64(initPartFieldSize))

			return message{
				buffer:             b[:len(b)+initPartFieldSize],
				bodyFieldTagOffset: len(b),
				signBody:           signBody,
			}, nil
		}

		prepareChunkMsg = func(chunkSize int) (message, error) {
			if buf == nil {
				panic("the buffer should have already been allocated")
			}

			fieldSize := protowire.SizeTag(fieldNumObjectPutBodyChunk) + protowire.SizeBytes(chunkSize)

			b := buf
			b = protowire.AppendVarint(b, uint64(fieldSize))

			return message{
				buffer:             b[:len(b)+fieldSize],
				bodyFieldTagOffset: len(b),
				signBody:           signBody,
			}, nil
		}
	} else {
		prepareInitMsg = func() (message, error) {
			// TODO: support custom allocation (e.g. use sync.Pool)
			buf := make([]byte, 0, staticFields.fullSize+protowire.SizeVarint(uint64(initPartFieldSize))+initPartFieldSize)

			signBody, err := staticFields.appendTo(buf)
			if err != nil {
				return message{}, err
			}

			buf = buf[:staticFields.fullSize]
			buf = protowire.AppendVarint(buf, uint64(initPartFieldSize))

			return message{
				buffer:             buf[:cap(buf)],
				bodyFieldTagOffset: len(buf),
				signBody:           signBody,
			}, nil
		}

		prepareChunkMsg = func(chunkSize int) (message, error) {
			fieldSize := protowire.SizeTag(fieldNumObjectPutBodyChunk) + protowire.SizeBytes(chunkSize)
			fullSize := staticFields.fullSize + protowire.SizeVarint(uint64(fieldSize)) + fieldSize

			// TODO: support custom allocation (e.g. use sync.Pool)
			buf := make([]byte, 0, fullSize)

			signBody, err := staticFields.appendTo(buf)
			if err != nil {
				return message{}, err
			}

			buf = buf[:staticFields.fullSize]
			buf = protowire.AppendVarint(buf, uint64(fieldSize))

			return message{
				buffer:             buf[:cap(buf)],
				bodyFieldTagOffset: len(buf),
				signBody:           signBody,
			}, nil
		}
	}

	sendMessage := func(msg message) error {
		err = msg.signBody(msg.buffer[msg.bodyFieldTagOffset:])
		if err != nil {
			return err
		}

		err = dst.Write(msg.buffer)
		if err != nil {
			return fmt.Errorf("send ready message to the stream: %w", err)
		}

		return nil
	}

	// we are ready to complete the messages and stream them

	msg, err := prepareInitMsg()
	if err != nil {
		return fmt.Errorf("prepare initial stream message: %w", err)
	}

	bodyFieldOff := msg.bodyFieldTagOffset + writeProtoTag(msg.buffer[msg.bodyFieldTagOffset:], fieldNumObjectPutBodyInitPart, protowire.BytesType)
	bodyFieldOff += binary.PutUvarint(msg.buffer[bodyFieldOff:], uint64(headerSize))

	if headerSize > 0 {
		if binObjInfo.payloadSize == 0 || binObjInfo.payloadFieldOffset == 0 || binObjInfo.payloadFieldOffset == headerSize {
			if binObjInfo.payloadFieldOffset == 0 {
				_, err = src.Seek(int64(binObjInfo.payloadOffset+binObjInfo.payloadSize), io.SeekStart)
				if err != nil {
					return fmt.Errorf("seek object header: %w", err)
				}
			}

			err = readExact(src, msg.buffer[bodyFieldOff:bodyFieldOff+headerSize])
			if err != nil {
				return fmt.Errorf("read continuous object header: %w", err)
			}
		} else {
			// the headers have a payload break, so they must be assembled from two parts
			err = readExact(src, msg.buffer[bodyFieldOff:bodyFieldOff+binObjInfo.payloadFieldOffset])
			if err != nil {
				return fmt.Errorf("read 1st part of the object header: %w", err)
			}

			_, err = src.Seek(int64(binObjInfo.payloadOffset+binObjInfo.payloadSize), io.SeekStart)
			if err != nil {
				return fmt.Errorf("seek 2nd part of the object header: %w", err)
			}

			err = readExact(src, msg.buffer[bodyFieldOff+binObjInfo.payloadFieldOffset:])
			if err != nil {
				return fmt.Errorf("read 2nd part of the object header: %w", err)
			}
		}
	}

	err = sendMessage(msg)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("send initial stream message: %w", err)
	}

	if binObjInfo.payloadSize > 0 {
		_, err = src.Seek(int64(binObjInfo.payloadOffset), io.SeekStart)
		if err != nil {
			return fmt.Errorf("seek object payload start: %w", err)
		}

		remainingPayloadSize := binObjInfo.payloadSize
		chunkSize := maxChunkSize

		for {
			if chunkSize > remainingPayloadSize {
				chunkSize = remainingPayloadSize
			}

			msg, err = prepareChunkMsg(chunkSize)
			if err != nil {
				return fmt.Errorf("prepare payload chunk message: %w", err)
			}

			bodyFieldOff = msg.bodyFieldTagOffset + writeProtoTag(msg.buffer[msg.bodyFieldTagOffset:], fieldNumObjectPutBodyChunk, protowire.BytesType)
			bodyFieldOff += binary.PutUvarint(msg.buffer[bodyFieldOff:], uint64(chunkSize))

			err = readExact(src, msg.buffer[bodyFieldOff:])
			if err != nil {
				return fmt.Errorf("read next payload chunk: %w", err)
			}

			err = sendMessage(msg)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}

				return fmt.Errorf("send stream message with next payload chunk: %w", err)
			}

			if chunkSize == remainingPayloadSize {
				break
			}

			remainingPayloadSize -= chunkSize
		}
	}

	return nil
}

type scanBinaryObjectRes struct {
	// FIXME: make all sizes uint64, int may be limited by 4GB
	fullSize int

	payloadFieldOffset int
	payloadOffset      int // payloadFieldOffset + field tag
	payloadSize        int
}

// scans given [io.ReadSeeker] providing access to the NeoFS object encoded into
// Protocol Buffers V3 and collects information about message structure. If no
// error, argument carriage is set to the start.
func scanBinaryObject(src io.ReadSeeker) (scanBinaryObjectRes, error) {
	var res scanBinaryObjectRes
	var err error
	var u64 uint64
	var n, tagLn, ln int
	var fieldNum protowire.Number
	var fieldType protowire.Type
	// create buffer for service bytes (field tag + length)
	// TODO: think is it worth to customize this allocation (e.g. use sync.Pool)
	buf := make([]byte, 2*binary.MaxVarintLen64)

	res.payloadFieldOffset = -1
	res.payloadOffset = -1

	for {
		n, err = src.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return res, fmt.Errorf("read data from source: %w", err)
		}

		if n > 0 {
			fieldNum, fieldType, tagLn = protowire.ConsumeTag(buf[:n])
			if n < 0 {
				return res, protowire.ParseError(n)
			}

			if fieldType != protowire.BytesType {
				// incl. embedded messages according to https://protobuf.dev/programming-guides/encoding/#structure
				return res, fmt.Errorf("invalid/unsupported object field type %d", fieldType)
			}

			const payloadFieldNum = 4
			switch fieldNum {
			default:
				return res, fmt.Errorf("invalid/unsupported object field number %d", fieldNum)
			case 1, 2, 3, payloadFieldNum:
				// all fields have same wire type (see condition above), so encoded identically
				u64, ln = protowire.ConsumeVarint(buf[tagLn:n])
				if ln < 0 {
					return res, protowire.ParseError(ln)
				}

				if fieldNum == payloadFieldNum {
					if res.payloadFieldOffset >= 0 {
						return res, errors.New("repeated payload field")
					}

					res.payloadFieldOffset = res.fullSize
					res.payloadOffset = res.payloadFieldOffset + tagLn + ln
					res.payloadSize = int(u64)
				}

				// move carriage to the next field
				if diff := int64(n - tagLn - ln - int(u64)); diff != 0 {
					// otherwise, source data carriage is exactly on the next field start or EOF
					//
					// can be optimized for diff > 0: because we have already read the next bytes
					// (buffer tail), Seek may be skipped. This will complicate function code, so
					// for now just Seek for simplicity. In practice, diff will be positive for most
					// cases
					_, err = src.Seek(-diff, io.SeekCurrent)
					if err != nil {
						return res, fmt.Errorf("seek next field: %w", err)
					}
				}

				res.fullSize += tagLn + ln + int(u64)
			}
		}

		if errors.Is(err, io.EOF) {
			_, err = src.Seek(0, io.SeekStart)
			if err != nil {
				return res, fmt.Errorf("seek back to start: %w", err)
			}

			return res, nil
		}
	}
}

// describes common markup of all messages transmitted within some
// ObjectService.Put stream. The markup helps simplifies message encoding due to
// static parts remaining unchanged (completely or in terms of buffer ranges)
// within the stream lifetime.
type objectPutRequestStaticFields struct {
	signer       neofscrypto.Signer
	sigScheme    uint64
	bPubKey      []byte
	fixedSigSize uint64
	sigFieldSize uint64

	fullSize int

	metaHeader struct {
		fullSize uint64

		ttl     uint64
		version struct {
			fullSize uint64

			major uint64
			minor uint64
		}
	}

	verificationHeader struct {
		fullSize uint64

		originSig []byte
		metaSig   []byte // lazily calculated once
	}
}

// initializes objectPutRequestStaticFields with the provided parameters that do
// not change within the stream lifetime. The results must not be modified.
//
// TTL must be positive. [neofscrypto.Signer] must produce fixed-size signatures
// only.
func initObjectPutRequestStaticFields(signer neofscrypto.Signer, ttl uint64, vNeoFS version.Version) (objectPutRequestStaticFields, error) {
	var res objectPutRequestStaticFields
	var err error

	res.verificationHeader.originSig, err = signer.Sign(nil)
	if err != nil {
		return res, fmt.Errorf("sign empty origin verification header: %w", err)
	}

	res.bPubKey = neofscrypto.PublicKeyBytes(signer.Public())
	res.sigScheme = uint64(signer.Scheme())
	fixedSigSize := len(res.verificationHeader.originSig)
	res.fixedSigSize = uint64(fixedSigSize)
	res.signer = signer

	res.metaHeader.ttl = ttl
	res.metaHeader.version.major = uint64(vNeoFS.Major())
	res.metaHeader.version.minor = uint64(vNeoFS.Minor())
	metaHeaderVersionFieldFullSize :=
		// major
		protowire.SizeTag(fieldNumVersionMajor) + protowire.SizeVarint(res.metaHeader.version.major) +
			// minor
			protowire.SizeTag(fieldNumVersionMinor) + protowire.SizeVarint(res.metaHeader.version.minor)
	res.metaHeader.version.fullSize = uint64(metaHeaderVersionFieldFullSize)
	metaHeaderFullSize :=
		// version
		protowire.SizeTag(fieldNumRequestMetaVersion) + protowire.SizeBytes(metaHeaderVersionFieldFullSize) +
			// TTL
			protowire.SizeTag(fieldNumRequestMetaTTL) + protowire.SizeVarint(ttl)
	res.metaHeader.fullSize = uint64(metaHeaderFullSize)

	sigMsgSize :=
		// public key
		protowire.SizeTag(fieldNumSigPubKey) + protowire.SizeBytes(len(res.bPubKey)) +
			// value
			protowire.SizeTag(fieldNumSigVal) + +protowire.SizeBytes(fixedSigSize) +
			// scheme
			protowire.SizeTag(fieldNumSigScheme) + protowire.SizeVarint(res.sigScheme)

	verifyHeaderFullSize :=
		// body signature
		protowire.SizeTag(fieldNumVerifyHdrBodySig) + protowire.SizeBytes(sigMsgSize) +
			// meta signature
			protowire.SizeTag(fieldNumVerifyHdrMetaSig) + protowire.SizeBytes(sigMsgSize) +
			// origin signature
			protowire.SizeTag(fieldNumVerifyHdrOriginSig) + protowire.SizeBytes(sigMsgSize)
	res.verificationHeader.fullSize = uint64(verifyHeaderFullSize)
	res.sigFieldSize = uint64(sigMsgSize)

	res.fullSize = // meta header
		protowire.SizeTag(fieldNumRequestMeta) + protowire.SizeBytes(metaHeaderFullSize) +
			// verification header
			protowire.SizeTag(fieldNumRequestVerify) + protowire.SizeBytes(verifyHeaderFullSize) +
			// body (field tag only)
			protowire.SizeTag(fieldNumRequestBody)

	return res, nil
}

// appends the objectPutRequestStaticFields to the given buffer that must have
// at least objectPutRequestStaticFields.fullSize free capacity. The method
// returns function to pass request body message into: signBody calculates and
// writes body signature into the buffer lazily.
func (x *objectPutRequestStaticFields) appendTo(buf []byte) (signBody func(body []byte) error, err error) {
	// meta header
	buf = protowire.AppendTag(buf, fieldNumRequestMeta, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.metaHeader.fullSize)
	metaHeaderOff := len(buf)
	// version
	buf = protowire.AppendTag(buf, fieldNumRequestMetaVersion, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.metaHeader.version.fullSize)
	buf = protowire.AppendTag(buf, fieldNumVersionMajor, protowire.VarintType)
	buf = protowire.AppendVarint(buf, x.metaHeader.version.major)
	buf = protowire.AppendTag(buf, fieldNumVersionMinor, protowire.VarintType)
	buf = protowire.AppendVarint(buf, x.metaHeader.version.minor)
	// TTL
	buf = protowire.AppendTag(buf, fieldNumRequestMetaTTL, protowire.VarintType)
	buf = protowire.AppendVarint(buf, x.metaHeader.ttl)

	if x.verificationHeader.metaSig == nil {
		x.verificationHeader.metaSig, err = x.signer.Sign(buf[metaHeaderOff:])
		if err != nil {
			return signBody, fmt.Errorf("sign meta header: %w", err)
		}

		if uint64(len(x.verificationHeader.metaSig)) != x.fixedSigSize {
			return signBody, fmt.Errorf("unexpected meta header signature size: %d instead of %d",
				len(x.verificationHeader.metaSig), x.fixedSigSize)
		}
	}

	// verification header
	buf = protowire.AppendTag(buf, fieldNumRequestVerify, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.verificationHeader.fullSize)
	// body signature
	buf = protowire.AppendTag(buf, fieldNumVerifyHdrBodySig, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.sigFieldSize)
	// public key
	buf = protowire.AppendTag(buf, fieldNumSigPubKey, protowire.BytesType)
	buf = protowire.AppendBytes(buf, x.bPubKey)
	// value
	buf = protowire.AppendTag(buf, fieldNumSigVal, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.fixedSigSize)
	bodySigOff := uint64(len(buf))
	buf = buf[:bodySigOff+x.fixedSigSize]
	signBody = func(body []byte) error {
		sig, err := x.signer.Sign(body)
		if err != nil {
			return fmt.Errorf("sign request body: %w", err)
		}

		if uint64(len(sig)) != x.fixedSigSize {
			return fmt.Errorf("unexpected request body signature size: %d instead of %d",
				len(sig), x.fixedSigSize)
		}

		copy(buf[bodySigOff:], sig)

		return nil
	}
	// scheme
	buf = protowire.AppendTag(buf, fieldNumSigScheme, protowire.VarintType)
	buf = protowire.AppendVarint(buf, x.sigScheme)
	// meta signature
	buf = protowire.AppendTag(buf, fieldNumVerifyHdrMetaSig, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.sigFieldSize)
	// public key
	buf = protowire.AppendTag(buf, fieldNumSigPubKey, protowire.BytesType)
	buf = protowire.AppendBytes(buf, x.bPubKey)
	// value
	buf = protowire.AppendTag(buf, fieldNumSigVal, protowire.BytesType)
	buf = protowire.AppendBytes(buf, x.verificationHeader.metaSig)
	// scheme
	buf = protowire.AppendTag(buf, fieldNumSigScheme, protowire.VarintType)
	buf = protowire.AppendVarint(buf, x.sigScheme)
	// origin signature
	buf = protowire.AppendTag(buf, fieldNumVerifyHdrOriginSig, protowire.BytesType)
	buf = protowire.AppendVarint(buf, x.sigFieldSize)
	// public key
	buf = protowire.AppendTag(buf, fieldNumSigPubKey, protowire.BytesType)
	buf = protowire.AppendBytes(buf, x.bPubKey)
	// value
	buf = protowire.AppendTag(buf, fieldNumSigVal, protowire.BytesType)
	buf = protowire.AppendBytes(buf, x.verificationHeader.originSig)
	// scheme
	buf = protowire.AppendTag(buf, fieldNumSigScheme, protowire.VarintType)
	buf = protowire.AppendVarint(buf, x.sigScheme)

	// body (only tag is static)
	buf = protowire.AppendTag(buf, fieldNumRequestBody, protowire.BytesType)

	return
}

// writes tag and size of the ObjectService.Put RPC request field carrying the
// object payload chunk into the buffer and returns number of bytes written. The
// function also returns offset to the field following the tag.
func writeObjectPutBodyChunkFieldTags(buf []byte, chunkSize int) (n, off int) {
	off = binary.PutUvarint(buf, uint64(protowire.SizeTag(fieldNumObjectPutBodyChunk)+protowire.SizeBytes(chunkSize)))
	n = off + writeProtoTag(buf[off:], fieldNumObjectPutBodyChunk, protowire.BytesType)
	return n + binary.PutUvarint(buf[n:], uint64(chunkSize)), off
}
