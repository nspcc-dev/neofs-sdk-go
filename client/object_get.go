package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

var errInvalidSplitInfo = errors.New("invalid split info")

// shared parameters of GET/HEAD/RANGE.
type prmObjectRead struct {
	prmCommonMeta
	sessionContainer
	bearerToken *bearer.Token
	local       bool

	raw bool
}

// MarkRaw marks an intent to read physically stored object.
func (x *prmObjectRead) MarkRaw() {
	x.raw = true
}

// MarkLocal tells the server to execute the operation locally.
func (x *prmObjectRead) MarkLocal() {
	x.local = true
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *prmObjectRead) WithBearerToken(t bearer.Token) {
	x.bearerToken = &t
}

// PrmObjectGet groups optional parameters of ObjectGetInit operation.
type PrmObjectGet struct {
	prmObjectRead
}

// used part of [protoobject.ObjectService_GetClient] simplifying test
// implementations.
type getObjectResponseStream interface {
	// Recv reads next message with the object part from the stream. Recv returns
	// [io.EOF] after the server sent the last message and gracefully finished the
	// stream. Any other error means stream abort.
	Recv() (*protoobject.GetResponse, error)
}

// PayloadReader is a data stream of the particular NeoFS object. Implements
// [io.ReadCloser].
//
// Must be initialized using Client.ObjectGetInit, any other
// usage is unsafe.
type PayloadReader struct {
	cancelCtxStream context.CancelFunc

	stream           getObjectResponseStream
	singleMsgTimeout time.Duration

	err error

	tailPayload []byte

	remainingPayloadLen int

	statisticCallback shortStatisticCallback
	startTime         time.Time // if statisticCallback is set only
}

// readHeader reads header of the object. Result means success.
// Failure reason can be received via Close.
func (x *PayloadReader) readHeader(dst *object.Object) bool {
	var resp *protoobject.GetResponse
	x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
		var err error
		resp, err = x.stream.Recv()
		return err
	})
	if x.err != nil {
		return false
	}

	if x.err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); x.err != nil {
		return false
	}

	var partInit *protoobject.GetResponse_Body_Init

	switch v := resp.GetBody().GetObjectPart().(type) {
	default:
		x.err = fmt.Errorf("unexpected message instead of heading part: %T", v)
		return false
	case *protoobject.GetResponse_Body_SplitInfo:
		if v == nil || v.SplitInfo == nil {
			x.err = fmt.Errorf("%w: nil split info field", errInvalidSplitInfo)
			return false
		}
		var si object.SplitInfo
		if x.err = si.FromProtoMessage(v.SplitInfo); x.err != nil {
			x.err = fmt.Errorf("%w: %w", errInvalidSplitInfo, x.err)
			return false
		}
		x.err = object.NewSplitInfoError(&si)
		return false
	case *protoobject.GetResponse_Body_Init_:
		if v == nil || v.Init == nil {
			x.err = errors.New("nil header oneof field")
			return false
		}
		partInit = v.Init
	}

	if partInit.ObjectId == nil {
		x.err = newErrMissingResponseField("object ID")
		return false
	}
	if partInit.Signature == nil {
		x.err = newErrMissingResponseField("signature")
		return false
	}
	if partInit.Header == nil {
		x.err = newErrMissingResponseField("header")
		return false
	}

	x.remainingPayloadLen = int(partInit.Header.GetPayloadLength())

	x.err = dst.FromProtoMessage(&protoobject.Object{
		ObjectId:  partInit.ObjectId,
		Signature: partInit.Signature,
		Header:    partInit.Header,
	})
	return x.err == nil
}

func (x *PayloadReader) readChunk(buf []byte) (int, bool) {
	var read int

	// read remaining tail
	read = copy(buf, x.tailPayload)

	x.tailPayload = x.tailPayload[read:]

	if len(buf) == read {
		return read, true
	}

	var chunk []byte
	var lastRead int

	for {
		var resp *protoobject.GetResponse
		x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
			var err error
			resp, err = x.stream.Recv()
			return err
		})
		if x.err != nil {
			return read, false
		}

		if x.err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); x.err != nil {
			return read, false
		}

		part := resp.GetBody().GetObjectPart()
		partChunk, ok := part.(*protoobject.GetResponse_Body_Chunk)
		if !ok {
			x.err = fmt.Errorf("unexpected message instead of chunk part: %T", part)
			return read, false
		}
		if partChunk == nil {
			x.err = errors.New("nil chunk oneof field")
			return read, false
		}

		// read new chunk
		chunk = partChunk.Chunk
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

func (x *PayloadReader) close(ignoreEOF bool) error {
	defer x.cancelCtxStream()

	if errors.Is(x.err, io.EOF) {
		if ignoreEOF {
			return nil
		}
		if x.remainingPayloadLen > 0 {
			return io.ErrUnexpectedEOF
		}
	}
	return x.err
}

// Close ends reading the object payload. Must be called after using the
// PayloadReader.
func (x *PayloadReader) Close() error {
	var err error
	if x.statisticCallback != nil {
		defer func() {
			x.statisticCallback(time.Since(x.startTime), err)
		}()
	}
	err = x.close(true)
	return err
}

// Read implements io.Reader of the object payload.
func (x *PayloadReader) Read(p []byte) (int, error) {
	n, ok := x.readChunk(p)

	x.remainingPayloadLen -= n

	if !ok {
		if err := x.close(false); err != nil {
			return n, err
		}

		return n, x.err
	}

	if x.remainingPayloadLen < 0 {
		return n, errors.New("payload size overflow")
	}

	return n, nil
}

// ObjectGetInit initiates reading an object through a remote server using NeoFS API protocol.
// Returns header of the requested object and stream of its payload separately.
//
// Exactly one return value is non-nil. Resulting PayloadReader must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
// If signer implements [neofscrypto.SignerV2], signing is done using it. In
// this case, [neofscrypto.Signer] methods are not called.
// [neofscrypto.OverlapSigner] may be used to pass [neofscrypto.SignerV2] when
// [neofscrypto.Signer] is unimplemented.
//
// Return errors:
//   - global (see Client docs)
//   - [ErrMissingSigner]
//   - *[object.SplitInfoError] (returned on virtual objects with PrmObjectGet.MakeRaw)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectAlreadyRemoved]
//   - [apistatus.ErrSessionTokenExpired]
func (c *Client) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectGet) (object.Object, *PayloadReader, error) {
	var (
		hdr object.Object
		err error
	)

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectGet, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return hdr, nil, ErrMissingSigner
	}

	req := &protoobject.GetRequest{
		Body: &protoobject.GetRequest_Body{
			Address: oid.NewAddress(containerID, objectID).ProtoMessage(),
			Raw:     prm.raw,
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)
	if prm.local {
		req.MetaHeader.Ttl = localRequestTTL
	} else {
		req.MetaHeader.Ttl = defaultRequestTTL
	}
	if prm.session != nil {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
	}
	if prm.bearerToken != nil {
		req.MetaHeader.BearerToken = prm.bearerToken.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return hdr, nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := c.object.Get(ctx, req)
	if err != nil {
		cancel()
		err = fmt.Errorf("open stream: %w", err)
		return hdr, nil, err
	}

	var r PayloadReader
	r.cancelCtxStream = cancel
	r.stream = stream
	r.singleMsgTimeout = c.streamTimeout
	if c.prm.statisticCallback != nil {
		r.startTime = time.Now()
		r.statisticCallback = func(dur time.Duration, err error) {
			c.sendStatistic(stat.MethodObjectGetStream, dur, err)
		}
	}

	if !r.readHeader(&hdr) {
		err = fmt.Errorf("read header: %w", r.Close())
		return hdr, nil, err
	}

	return hdr, &r, nil
}

// PrmObjectHead groups optional parameters of ObjectHead operation.
type PrmObjectHead struct {
	prmObjectRead
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
// If signer implements [neofscrypto.SignerV2], signing is done using it. In
// this case, [neofscrypto.Signer] methods are not called.
// [neofscrypto.OverlapSigner] may be used to pass [neofscrypto.SignerV2] when
// [neofscrypto.Signer] is unimplemented.
//
// Return errors:
//   - global (see Client docs)
//   - [ErrMissingSigner]
//   - *[object.SplitInfoError] (returned on virtual objects with PrmObjectHead.MakeRaw)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectAlreadyRemoved]
//   - [apistatus.ErrSessionTokenExpired]
func (c *Client) ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectHead) (*object.Object, error) {
	var err error

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectHead, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}

	req := &protoobject.HeadRequest{
		Body: &protoobject.HeadRequest_Body{
			Address: oid.NewAddress(containerID, objectID).ProtoMessage(),
			Raw:     prm.raw,
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)
	if prm.local {
		req.MetaHeader.Ttl = localRequestTTL
	} else {
		req.MetaHeader.Ttl = defaultRequestTTL
	}
	if prm.session != nil {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
	}
	if prm.bearerToken != nil {
		req.MetaHeader.BearerToken = prm.bearerToken.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.HeadRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
	}

	resp, err := c.object.Head(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return nil, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return nil, err
	}

	switch v := resp.GetBody().GetHead().(type) {
	default:
		err = fmt.Errorf("unexpected header type %T", v)
		return nil, err
	case *protoobject.HeadResponse_Body_SplitInfo:
		if v == nil || v.SplitInfo == nil {
			err = fmt.Errorf("%w: nil split info field", errInvalidSplitInfo)
			return nil, err
		}
		var si object.SplitInfo
		if err = si.FromProtoMessage(v.SplitInfo); err != nil {
			err = fmt.Errorf("%w: %w", errInvalidSplitInfo, err)
			return nil, err
		}
		err = object.NewSplitInfoError(&si)
		return nil, err
	case *protoobject.HeadResponse_Body_Header:
		if v == nil {
			return nil, errors.New("empty header")
		}
		if v.Header.Signature == nil {
			err = newErrMissingResponseField("signature")
			return nil, err
		}
		if v.Header.Header == nil {
			err = newErrMissingResponseField("header")
			return nil, err
		}

		var obj object.Object
		if err = obj.FromProtoMessage(&protoobject.Object{
			ObjectId:  objectID.ProtoMessage(),
			Signature: v.Header.Signature,
			Header:    v.Header.Header,
		}); err != nil {
			return nil, fmt.Errorf("invalid header response: %w", err)
		}
		return &obj, nil
	}
}

// PrmObjectRange groups optional parameters of ObjectRange operation.
type PrmObjectRange struct {
	prmObjectRead
}

// used part of [protoobject.ObjectService_GetRangeClient] simplifying test
// implementations.
type getObjectPayloadRangeResponseStream interface {
	// Recv reads next message with the object payload part from the stream. Recv
	// returns [io.EOF] after the server sent the last message and gracefully
	// finished the stream. Any other error means stream abort.
	Recv() (*protoobject.GetRangeResponse, error)
}

// ObjectRangeReader is designed to read payload range of one object
// from NeoFS system. Implements [io.ReadCloser].
//
// Must be initialized using Client.ObjectRangeInit, any other
// usage is unsafe.
type ObjectRangeReader struct {
	cancelCtxStream context.CancelFunc

	err error

	stream           getObjectPayloadRangeResponseStream
	singleMsgTimeout time.Duration

	tailPayload []byte

	requestedLen, receivedLen uint64

	statisticCallback shortStatisticCallback
	startTime         time.Time // if statisticCallback is set only
}

func (x *ObjectRangeReader) readChunk(buf []byte) (int, bool) {
	var read int

	// read remaining tail
	read = copy(buf, x.tailPayload)

	x.tailPayload = x.tailPayload[read:]

	if len(buf) == read {
		return read, true
	}

	var chunk []byte
	var lastRead int

	for {
		var resp *protoobject.GetRangeResponse
		x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
			var err error
			resp, err = x.stream.Recv()
			return err
		})
		if x.err != nil {
			return read, false
		}

		if x.err = neofscrypto.VerifyResponseWithBuffer[*protoobject.GetRangeResponse_Body](resp, nil); x.err != nil {
			x.err = fmt.Errorf("%w: %w", errResponseSignatures, x.err)
			return read, false
		}

		if x.err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); x.err != nil {
			return read, false
		}

		// get chunk message
		switch v := resp.GetBody().GetRangePart().(type) {
		default:
			x.err = fmt.Errorf("unexpected message received: %T", v)
			return read, false
		case *protoobject.GetRangeResponse_Body_SplitInfo:
			if v == nil || v.SplitInfo == nil {
				x.err = fmt.Errorf("%w: nil split info field", errInvalidSplitInfo)
				return read, false
			}
			var si object.SplitInfo
			if x.err = si.FromProtoMessage(v.SplitInfo); x.err != nil {
				x.err = fmt.Errorf("%w: %w", errInvalidSplitInfo, x.err)
				return read, false
			}
			x.err = object.NewSplitInfoError(&si)
			return read, false
		case *protoobject.GetRangeResponse_Body_Chunk:
			if v == nil {
				x.err = errors.New("nil header oneof field")
				return read, false
			}
			chunk = v.Chunk
		}

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

func (x *ObjectRangeReader) close(ignoreEOF bool) error {
	defer x.cancelCtxStream()

	if errors.Is(x.err, io.EOF) {
		if ignoreEOF {
			return nil
		}
		if x.receivedLen < x.requestedLen {
			return io.ErrUnexpectedEOF
		}
	}
	return x.err
}

// Close ends reading the payload range and returns the result of the operation
// along with the final results. Must be called after using the ObjectRangeReader.
//
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return errors:
//   - global (see Client docs)
//   - *[object.SplitInfoError] (returned on virtual objects with PrmObjectRange.MakeRaw)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectAlreadyRemoved]
//   - [apistatus.ErrObjectOutOfRange]
//   - [apistatus.ErrSessionTokenExpired]
func (x *ObjectRangeReader) Close() error {
	var err error
	if x.statisticCallback != nil {
		defer func() {
			x.statisticCallback(time.Since(x.startTime), err)
		}()
	}
	err = x.close(true)
	return err
}

// Read implements io.Reader of the object payload.
func (x *ObjectRangeReader) Read(p []byte) (int, error) {
	n, ok := x.readChunk(p)

	x.receivedLen += uint64(n)

	if !ok {
		err := x.close(false)
		if err != nil {
			return n, err
		}

		return n, x.err
	}

	if x.requestedLen > 0 && x.receivedLen > x.requestedLen { // zero means full payload, we don't know its size
		return n, errors.New("payload range size overflow")
	}

	return n, nil
}

// ObjectRangeInit initiates reading an object's payload range through a remote
// server using NeoFS API protocol.
//
// To get full payload, set both offset and length to zero. Otherwise, length
// must not be zero.
//
// The call only opens the transmission channel, explicit fetching is done using the ObjectRangeReader.
// Exactly one return value is non-nil. Resulting reader must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
// If signer implements [neofscrypto.SignerV2], signing is done using it. In
// this case, [neofscrypto.Signer] methods are not called.
// [neofscrypto.OverlapSigner] may be used to pass [neofscrypto.SignerV2] when
// [neofscrypto.Signer] is unimplemented.
//
// Return errors:
//   - [ErrZeroRangeLength]
//   - [ErrMissingSigner]
func (c *Client) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer user.Signer, prm PrmObjectRange) (*ObjectRangeReader, error) {
	var err error

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectRange, time.Since(startTime), err)
		}()
	}

	if length == 0 && offset != 0 {
		err = ErrZeroRangeLength
		return nil, err
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}

	req := &protoobject.GetRangeRequest{
		Body: &protoobject.GetRangeRequest_Body{
			Address: oid.NewAddress(containerID, objectID).ProtoMessage(),
			Range:   &protoobject.Range{Offset: offset, Length: length},
			Raw:     prm.raw,
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)
	if prm.local {
		req.MetaHeader.Ttl = localRequestTTL
	} else {
		req.MetaHeader.Ttl = defaultRequestTTL
	}
	if prm.session != nil {
		req.MetaHeader.SessionToken = prm.session.ProtoMessage()
	}
	if prm.bearerToken != nil {
		req.MetaHeader.BearerToken = prm.bearerToken.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.GetRangeRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := c.object.GetRange(ctx, req)
	if err != nil {
		cancel()
		err = fmt.Errorf("open stream: %w", err)
		return nil, err
	}

	var r ObjectRangeReader
	r.requestedLen = length
	r.cancelCtxStream = cancel
	r.stream = stream
	r.singleMsgTimeout = c.streamTimeout
	if c.prm.statisticCallback != nil {
		r.startTime = time.Now()
		r.statisticCallback = func(dur time.Duration, err error) {
			c.sendStatistic(stat.MethodObjectRangeStream, dur, err)
		}
	}

	return &r, nil
}
