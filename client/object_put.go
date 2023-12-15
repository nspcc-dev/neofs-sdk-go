package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/common"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protowire"
)

// maxChunkLen restricts maximum byte length of the chunk
// transmitted in a single stream message. It depends on
// server settings and other message fields, but for now
// we simply assume that 3MB is large enough to reduce the
// number of messages, and not to exceed the limit
// (4MB by default for gRPC servers).
const maxChunkLen = 3 << 20

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

type PutFullObjectToNodeOptions struct {
	sessionToken v2session.Token
	bearerToken  acl.BearerToken
}

// WithinSession specifies session within which the operation should be
// executed. The token must be signed.
func (x *PutFullObjectToNodeOptions) WithinSession(t session.Object) {
	t.WriteToV2(&x.sessionToken)
}

// WithBearerToken attaches bearer token to be used for the operation.
func (x *PutFullObjectToNodeOptions) WithBearerToken(t bearer.Token) {
	t.WriteToV2(&x.bearerToken)
}

func (c *Client) PutFullObjectToNode(ctx context.Context, obj object.Object, signer neofscrypto.Signer, opts PutFullObjectToNodeOptions) error {
	if signer == nil {
		// note that we don't stat this error
		return ErrMissingSigner
	}

	var err error
	defer func() {
		c.sendStatistic(stat.MethodObjectPut, err)()
	}()

	const svcName = "neo.fs.v2.object.ObjectService"
	const opName = "Put"
	stream, err := c.c.Init(common.CallMethodInfoClientStream(svcName, opName),
		client.WithContext(ctx), client.AllowBinarySendingOnly())
	if err != nil {
		return fmt.Errorf("init service=%s/op=%s RPC: %w", svcName, opName, err)
	}

	err = streamFullObject(ctx, obj, signer, opts, func(msg []byte) error {
		return stream.WriteMessage(client.BinaryMessage(msg))
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("send request: %w", err)
	}

	if err == nil {
		err = stream.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("finish object stream: %w", err)
		}
	}

	var resp v2object.PutResponse
	err = stream.ReadMessage(&resp)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}

		return fmt.Errorf("recv response: %w", err)
	}

	return c.processResponse(&resp)
}

func NewSharedPutFullObjectContext(parent context.Context) context.Context {
	return &sharedPutFullObjectContext{
		Context: parent,
	}
}

type sharedPutFullObjectContext struct {
	context.Context

	mtx sync.Mutex

	err error

	markup putObjectStreamMsgMarkup
	msgs   [][]byte
}

func streamFullObject(ctx context.Context, obj object.Object, signer neofscrypto.Signer, opts PutFullObjectToNodeOptions, writeMsg func([]byte) error) error {
	payload := obj.Payload()

	msgNum := 1 + len(payload)/maxChunkLen // 1 for header
	if len(payload)%maxChunkLen != 0 {
		msgNum += 1
	}

	var markup putObjectStreamMsgMarkup
	var headingMsg []byte

	sharedCtx, isShared := ctx.(*sharedPutFullObjectContext)
	if isShared {
		sharedCtx.mtx.Lock()

		if sharedCtx.err != nil {
			err := sharedCtx.err
			sharedCtx.mtx.Unlock()
			return err
		}

		if sharedCtx.msgs == nil {
			sharedCtx.msgs = make([][]byte, msgNum)
			sharedCtx.msgs[0], sharedCtx.markup, sharedCtx.err = prepPutObjectHeadingMsgWithMarkup(obj, signer, opts)
			if sharedCtx.err != nil {
				err := sharedCtx.err
				sharedCtx.mtx.Unlock()
				return err
			}
		}

		headingMsg, markup = sharedCtx.msgs[0], sharedCtx.markup

		sharedCtx.mtx.Unlock()
	} else {
		var err error
		headingMsg, markup, err = prepPutObjectHeadingMsgWithMarkup(obj, signer, opts)
		if err != nil {
			return err
		}
	}

	err := writeMsg(headingMsg)
	if err != nil {
		return fmt.Errorf("write heading message: %w", err)
	}

	var msg []byte
	chunkSize := maxChunkLen
	msgPrefix := headingMsg[:markup.bodyFieldOff]

	for i := 1; i < msgNum; i++ {
		if len(payload) < chunkSize {
			chunkSize = len(payload)
		}

		if isShared {
			sharedCtx.mtx.Lock()

			if sharedCtx.err != nil {
				err := sharedCtx.err
				sharedCtx.mtx.Unlock()
				return err
			}

			if sharedCtx.msgs[i] == nil {
				sharedCtx.msgs[i], sharedCtx.err = prepPutObjectChunkMsg(msgPrefix, markup, signer, payload[:chunkSize])
				if sharedCtx.err != nil {
					err := sharedCtx.err
					sharedCtx.mtx.Unlock()
					return err
				}
			}

			msg = sharedCtx.msgs[i]

			sharedCtx.mtx.Unlock()
		} else {
			msg, err = prepPutObjectChunkMsg(msgPrefix, markup, signer, payload[:chunkSize])
			if err != nil {
				return err
			}
		}

		err = writeMsg(msg)
		if err != nil {
			return fmt.Errorf("write chunk message: %w", err)
		}

		payload = payload[chunkSize:]
	}

	return nil
}

type putObjectStreamMsgMarkup struct {
	fixedSigSize int
	bodySigOff   int

	bodyFieldOff int
}

func prepPutObjectHeadingMsgWithMarkup(obj object.Object, signer neofscrypto.Signer, opts PutFullObjectToNodeOptions) (msg []byte, markup putObjectStreamMsgMarkup, err error) {
	emptyDataSig, err := signer.Sign(nil)
	if err != nil {
		return msg, markup, fmt.Errorf("calculate empty data signature: %w", err)
	}

	const ttl = 1
	markup.fixedSigSize = len(emptyDataSig)
	vNeoFS := version.Current()
	vNeoFSMjr := uint64(vNeoFS.Major())
	vNeoFSMnr := uint64(vNeoFS.Minor())

	metaHdrVersionFieldSize := protowire.SizeTag(fieldNumVersionMajor) + protowire.SizeVarint(uint64(vNeoFSMjr)) +
		protowire.SizeTag(fieldNumVersionMinor) + protowire.SizeVarint(vNeoFSMnr)

	metaHdrFieldSize := protowire.SizeTag(fieldNumRequestMetaVersion) + protowire.SizeBytes(metaHdrVersionFieldSize) +
		protowire.SizeTag(fieldNumRequestMetaTTL) + protowire.SizeVarint(ttl)

	sessionTokenSize := opts.sessionToken.StableSize()
	if sessionTokenSize > 0 {
		metaHdrFieldSize += protowire.SizeTag(fieldNumRequestMetaSession) + protowire.SizeBytes(sessionTokenSize)
	}

	bearerTokenSize := opts.bearerToken.StableSize()
	if bearerTokenSize > 0 {
		metaHdrFieldSize += protowire.SizeTag(fieldNumRequestMetaBearer) + protowire.SizeBytes(bearerTokenSize)
	}

	const fieldNumRequestMetaHdr = 2
	msgSize := protowire.SizeTag(fieldNumRequestMetaHdr) + protowire.SizeBytes(metaHdrFieldSize)

	pubKey := signer.Public()
	sigScheme := uint64(signer.Scheme())
	bPubKey := neofscrypto.PublicKeyBytes(pubKey) // can be improved through Encode but requires fixed size
	sigMsgSize := protowire.SizeTag(fieldNumSigPubKey) + protowire.SizeBytes(len(bPubKey)) +
		protowire.SizeTag(fieldNumSigVal) + +protowire.SizeBytes(markup.fixedSigSize) +
		protowire.SizeTag(fieldNumSigScheme) + protowire.SizeVarint(sigScheme)
	verifyHdrSize := protowire.SizeTag(fieldNumVerifyHdrBodySig) + protowire.SizeBytes(sigMsgSize) +
		protowire.SizeTag(fieldNumVerifyHdrMetaSig) + protowire.SizeBytes(sigMsgSize) +
		protowire.SizeTag(fieldNumVerifyHdrOriginSig) + protowire.SizeBytes(sigMsgSize)

	const fieldNumRequestVerifyHdr = 3
	const fieldNumRequestBody = 1
	msgSize += protowire.SizeTag(fieldNumRequestVerifyHdr) + protowire.SizeBytes(verifyHdrSize) +
		protowire.SizeTag(fieldNumRequestBody)

	var objHdr *v2object.Object
	if len(obj.Payload()) > 0 {
		objHdr = obj.CutPayload().ToV2()
	} else {
		objHdr = obj.ToV2()
	}

	const fieldNumObjHdr = 1
	objHdrSize := objHdr.StableSize()
	objHdrFieldSize := protowire.SizeTag(fieldNumObjHdr) + protowire.SizeBytes(objHdrSize)

	// TODO: support custom allocs
	msg = make([]byte, 0, msgSize+protowire.SizeBytes(objHdrFieldSize))

	msg = protowire.AppendTag(msg, fieldNumRequestMetaHdr, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(metaHdrFieldSize))
	metaHdrOff := len(msg)
	msg = protowire.AppendTag(msg, fieldNumRequestMetaVersion, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(metaHdrVersionFieldSize))
	msg = protowire.AppendTag(msg, fieldNumVersionMajor, protowire.VarintType)
	msg = protowire.AppendVarint(msg, vNeoFSMjr)
	msg = protowire.AppendTag(msg, fieldNumVersionMinor, protowire.VarintType)
	msg = protowire.AppendVarint(msg, vNeoFSMnr)
	msg = protowire.AppendTag(msg, fieldNumRequestMetaTTL, protowire.VarintType)
	msg = protowire.AppendVarint(msg, ttl)
	if sessionTokenSize > 0 {
		msg = protowire.AppendTag(msg, fieldNumRequestMetaSession, protowire.BytesType)
		msg = protowire.AppendVarint(msg, uint64(sessionTokenSize))
		msg = msg[:len(msg)+sessionTokenSize]
		opts.sessionToken.StableMarshal(msg[len(msg)-sessionTokenSize:])
	}
	if bearerTokenSize > 0 {
		msg = protowire.AppendTag(msg, fieldNumRequestMetaBearer, protowire.BytesType)
		msg = protowire.AppendVarint(msg, uint64(bearerTokenSize))
		msg = msg[:len(msg)+bearerTokenSize]
		opts.bearerToken.StableMarshal(msg[len(msg)-bearerTokenSize:])
	}

	metaHdrSig, err := signer.Sign(msg[metaHdrOff:])
	if err != nil {
		return msg, markup, fmt.Errorf("sign meta header: %w", err)
	} else if len(metaHdrSig) != markup.fixedSigSize {
		return msg, markup, fmt.Errorf("various sizes of signatures detected: %d and %d", markup.fixedSigSize, len(metaHdrSig))
	}

	msg = protowire.AppendTag(msg, fieldNumRequestVerifyHdr, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(verifyHdrSize))
	msg = protowire.AppendTag(msg, fieldNumVerifyHdrBodySig, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(sigMsgSize))
	msg = protowire.AppendTag(msg, fieldNumSigPubKey, protowire.BytesType)
	msg = protowire.AppendBytes(msg, bPubKey)
	msg = protowire.AppendTag(msg, fieldNumSigVal, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(markup.fixedSigSize))
	markup.bodySigOff = len(msg)
	msg = msg[:markup.bodySigOff+markup.fixedSigSize]
	msg = protowire.AppendTag(msg, fieldNumSigScheme, protowire.VarintType)
	msg = protowire.AppendVarint(msg, sigScheme)
	msg = protowire.AppendTag(msg, fieldNumVerifyHdrMetaSig, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(sigMsgSize))
	msg = protowire.AppendTag(msg, fieldNumSigPubKey, protowire.BytesType)
	msg = protowire.AppendBytes(msg, bPubKey)
	msg = protowire.AppendTag(msg, fieldNumSigVal, protowire.BytesType)
	msg = protowire.AppendBytes(msg, metaHdrSig)
	msg = protowire.AppendTag(msg, fieldNumSigScheme, protowire.VarintType)
	msg = protowire.AppendVarint(msg, sigScheme)
	msg = protowire.AppendTag(msg, fieldNumVerifyHdrOriginSig, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(sigMsgSize))
	msg = protowire.AppendTag(msg, fieldNumSigPubKey, protowire.BytesType)
	msg = protowire.AppendBytes(msg, bPubKey)
	msg = protowire.AppendTag(msg, fieldNumSigVal, protowire.BytesType)
	msg = protowire.AppendBytes(msg, emptyDataSig)
	msg = protowire.AppendTag(msg, fieldNumSigScheme, protowire.VarintType)
	msg = protowire.AppendVarint(msg, sigScheme)
	msg = protowire.AppendTag(msg, fieldNumRequestBody, protowire.BytesType)
	markup.bodyFieldOff = len(msg)
	msg = protowire.AppendVarint(msg, uint64(objHdrFieldSize))
	bodyOff := len(msg)
	msg = protowire.AppendTag(msg, fieldNumObjHdr, protowire.BytesType)
	msg = protowire.AppendVarint(msg, uint64(objHdrSize))
	msg = msg[:len(msg)+objHdrSize]
	objHdr.StableMarshal(msg[len(msg)-objHdrSize:])

	bodySig, err := signer.Sign(msg[bodyOff:])
	if err != nil {
		return msg, markup, fmt.Errorf("sign request body: %w", err)
	} else if len(bodySig) != markup.fixedSigSize {
		return msg, markup, fmt.Errorf("various sizes of signatures detected: %d and %d", markup.fixedSigSize, len(bodySig))
	}

	copy(msg[markup.bodySigOff:], bodySig)

	return
}

func prepPutObjectChunkMsg(prefix []byte, markup putObjectStreamMsgMarkup, signer neofscrypto.Signer, chunk []byte) ([]byte, error) {
	const fieldNumObjPayloadChunk = 2
	bodySize := protowire.SizeTag(fieldNumObjPayloadChunk) + protowire.SizeBytes(len(chunk))

	// TODO: support custom allocs
	msg := make([]byte, 0, len(prefix)+protowire.SizeBytes(bodySize))
	msg = append(msg, prefix...)
	msg = protowire.AppendVarint(msg, uint64(bodySize))
	bodyOff := len(msg)
	msg = protowire.AppendTag(msg, fieldNumObjPayloadChunk, protowire.BytesType)
	msg = protowire.AppendBytes(msg, chunk)

	bodySig, err := signer.Sign(msg[bodyOff:])
	if err != nil {
		return nil, fmt.Errorf("sign request body: %w", err)
	} else if len(bodySig) != markup.fixedSigSize {
		return nil, fmt.Errorf("various sizes of signatures detected: %d and %d", markup.fixedSigSize, len(bodySig))
	}

	copy(msg[markup.bodySigOff:], bodySig)

	return msg, nil
}
