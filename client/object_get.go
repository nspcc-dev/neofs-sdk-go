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
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

var errInvalidSplitInfo = errors.New("invalid split info")

// shared parameters of GET/HEAD/RANGE.
type prmObjectRead struct {
	sessionContainer

	raw bool
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *prmObjectRead) WithXHeaders(hs ...string) {
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

	client           *Client
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
	var respV2 v2object.GetResponse
	if x.err = respV2.FromGRPCMessage(resp); x.err != nil {
		return false
	}

	x.err = x.client.processResponse(&respV2)
	if x.err != nil {
		return false
	}

	var partInit *v2object.GetObjectPartInit

	switch v := respV2.GetBody().GetObjectPart().(type) {
	default:
		x.err = fmt.Errorf("unexpected message instead of heading part: %T", v)
		return false
	case *v2object.SplitInfo:
		if v == nil {
			x.err = fmt.Errorf("%w: nil split info field", errInvalidSplitInfo)
			return false
		}
		var si object.SplitInfo
		if x.err = si.ReadFromV2(*v); x.err != nil {
			x.err = fmt.Errorf("%w: %w", errInvalidSplitInfo, x.err)
			return false
		}
		x.err = object.NewSplitInfoError(&si)
		return false
	case *v2object.GetObjectPartInit:
		if v == nil {
			x.err = newErrMissingResponseField("init")
			return false
		}
		partInit = v
	}

	id := partInit.GetObjectID()
	if id == nil {
		x.err = newErrMissingResponseField("object ID")
		return false
	}
	sig := partInit.GetSignature()
	if sig == nil {
		x.err = newErrMissingResponseField("signature")
		return false
	}
	hdr := partInit.GetHeader()
	if hdr == nil {
		x.err = newErrMissingResponseField("header")
		return false
	}

	var objv2 v2object.Object

	objv2.SetObjectID(id)
	objv2.SetHeader(hdr)
	objv2.SetSignature(sig)

	x.remainingPayloadLen = int(hdr.GetPayloadLength())

	x.err = dst.ReadFromV2(objv2)
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
		var respV2 v2object.GetResponse
		if x.err = respV2.FromGRPCMessage(resp); x.err != nil {
			return read, false
		}

		x.err = x.client.processResponse(&respV2)
		if x.err != nil {
			return read, false
		}

		part := respV2.GetBody().GetObjectPart()
		partChunk, ok := part.(*v2object.GetObjectPartChunk)
		if !ok {
			x.err = fmt.Errorf("unexpected message instead of chunk part: %T", part)
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
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		body  v2object.GetRequestBody
		hdr   object.Object
		err   error
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

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	body.SetRaw(prm.raw)
	body.SetAddress(&addr)

	// form request
	var req v2object.GetRequest

	req.SetBody(&body)
	c.prepareRequest(&req, &prm.meta)
	buf := c.buffers.Get().(*[]byte)
	err = signServiceMessage(signer, &req, *buf)
	c.buffers.Put(buf)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return hdr, nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := c.object.Get(ctx, req.ToGRPCMessage().(*protoobject.GetRequest))
	if err != nil {
		cancel()
		err = fmt.Errorf("open stream: %w", err)
		return hdr, nil, err
	}

	var r PayloadReader
	r.cancelCtxStream = cancel
	r.stream = stream
	r.singleMsgTimeout = c.streamTimeout
	r.client = c
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
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		body  v2object.HeadRequestBody
		err   error
	)

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectHead, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	body.SetRaw(prm.raw)
	body.SetAddress(&addr)

	var req v2object.HeadRequest
	req.SetBody(&body)
	c.prepareRequest(&req, &prm.meta)

	buf := c.buffers.Get().(*[]byte)
	err = signServiceMessage(signer, &req, *buf)
	c.buffers.Put(buf)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return nil, err
	}

	resp, err := c.object.Head(ctx, req.ToGRPCMessage().(*protoobject.HeadRequest))
	if err != nil {
		err = rpcErr(err)
		return nil, err
	}
	var respV2 v2object.HeadResponse
	if err = respV2.FromGRPCMessage(resp); err != nil {
		return nil, err
	}

	if err = c.processResponse(&respV2); err != nil {
		return nil, err
	}

	switch v := respV2.GetBody().GetHeaderPart().(type) {
	default:
		err = fmt.Errorf("unexpected header type %T", v)
		return nil, err
	case *v2object.SplitInfo:
		if v == nil {
			err = fmt.Errorf("%w: nil split info field", errInvalidSplitInfo)
			return nil, err
		}
		var si object.SplitInfo
		if err = si.ReadFromV2(*v); err != nil {
			err = fmt.Errorf("%w: %w", errInvalidSplitInfo, err)
			return nil, err
		}
		err = object.NewSplitInfoError(&si)
		return nil, err
	case *v2object.HeaderWithSignature:
		if v == nil {
			return nil, errors.New("empty header")
		}
		sig := v.GetSignature()
		if sig == nil {
			err = newErrMissingResponseField("signature")
			return nil, err
		}
		hdr := v.GetHeader()
		if hdr == nil {
			err = newErrMissingResponseField("header")
			return nil, err
		}

		var objv2 v2object.Object
		objv2.SetHeader(hdr)
		objv2.SetSignature(sig)

		var obj object.Object
		if err = obj.ReadFromV2(objv2); err != nil {
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

	client *Client

	err error

	stream           getObjectPayloadRangeResponseStream
	singleMsgTimeout time.Duration

	tailPayload []byte

	remainingPayloadLen int

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

	var partChunk *v2object.GetRangePartChunk
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
		var respV2 v2object.GetRangeResponse
		if x.err = respV2.FromGRPCMessage(resp); x.err != nil {
			return read, false
		}

		x.err = x.client.processResponse(&respV2)
		if x.err != nil {
			return read, false
		}

		// get chunk message
		switch v := respV2.GetBody().GetRangePart().(type) {
		default:
			x.err = fmt.Errorf("unexpected message received: %T", v)
			return read, false
		case *v2object.SplitInfo:
			if v == nil {
				x.err = fmt.Errorf("%w: nil split info field", errInvalidSplitInfo)
				return read, false
			}
			var si object.SplitInfo
			if x.err = si.ReadFromV2(*v); x.err != nil {
				x.err = fmt.Errorf("%w: %w", errInvalidSplitInfo, x.err)
				return read, false
			}
			x.err = object.NewSplitInfoError(&si)
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

func (x *ObjectRangeReader) close(ignoreEOF bool) error {
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

	x.remainingPayloadLen -= n

	if !ok {
		err := x.close(false)
		if err != nil {
			return n, err
		}

		return n, x.err
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
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
//
// Return errors:
//   - [ErrZeroRangeLength]
//   - [ErrMissingSigner]
func (c *Client) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer user.Signer, prm PrmObjectRange) (*ObjectRangeReader, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		rngV2 v2object.Range
		body  v2object.GetRangeRequestBody
		err   error
	)

	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectRange, time.Since(startTime), err)
		}()
	}

	if length == 0 {
		err = ErrZeroRangeLength
		return nil, err
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}

	containerID.WriteToV2(&cidV2)
	addr.SetContainerID(&cidV2)

	objectID.WriteToV2(&oidV2)
	addr.SetObjectID(&oidV2)

	rngV2.SetOffset(offset)
	rngV2.SetLength(length)

	// form request body
	body.SetRaw(prm.raw)
	body.SetAddress(&addr)
	body.SetRange(&rngV2)

	// form request
	var req v2object.GetRangeRequest

	req.SetBody(&body)
	c.prepareRequest(&req, &prm.meta)

	buf := c.buffers.Get().(*[]byte)
	err = signServiceMessage(signer, &req, *buf)
	c.buffers.Put(buf)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := c.object.GetRange(ctx, req.ToGRPCMessage().(*protoobject.GetRangeRequest))
	if err != nil {
		cancel()
		err = fmt.Errorf("open stream: %w", err)
		return nil, err
	}

	var r ObjectRangeReader
	r.remainingPayloadLen = int(length)
	r.cancelCtxStream = cancel
	r.stream = stream
	r.singleMsgTimeout = c.streamTimeout
	r.client = c
	if c.prm.statisticCallback != nil {
		r.startTime = time.Now()
		r.statisticCallback = func(dur time.Duration, err error) {
			c.sendStatistic(stat.MethodObjectRangeStream, dur, err)
		}
	}

	return &r, nil
}
