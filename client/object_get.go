package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// shared parameters of GET/HEAD/RANGE.
type prmObjectRead struct {
	meta v2session.RequestMetaHeader

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

// WithinSession specifies session within which object should be read.
//
// Creator of the session acquires the authorship of the request.
// This may affect the execution of an operation (e.g. access control).
//
// Must be signed.
func (x *prmObjectRead) WithinSession(t session.Object) {
	var tokv2 v2session.Token
	t.WriteToV2(&tokv2)
	x.meta.SetSessionToken(&tokv2)
}

// IsSessionSet checks is session within which object should be stored is set.
func (x prmObjectRead) IsSessionSet() bool {
	return x.meta.GetSessionToken() != nil
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

	signer neofscrypto.Signer
}

// ObjectReader is designed to read one object from NeoFS system.
//
// Must be initialized using Client.ObjectGetInit, any other
// usage is unsafe.
type ObjectReader struct {
	cancelCtxStream context.CancelFunc

	client *Client
	stream interface {
		Read(resp *v2object.GetResponse) error
	}

	err error

	tailPayload []byte

	remainingPayloadLen int
}

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Client default signer is used.
func (x *PrmObjectGet) UseSigner(signer neofscrypto.Signer) {
	x.signer = signer
}

// Signer returns associated with request signer.
func (x *PrmObjectGet) Signer() neofscrypto.Signer {
	return x.signer
}

// readHeader reads header of the object. Result means success.
// Failure reason can be received via Close.
func (x *ObjectReader) readHeader(dst *object.Object) bool {
	var resp v2object.GetResponse
	x.err = x.stream.Read(&resp)
	if x.err != nil {
		return false
	}

	x.err = x.client.processResponse(&resp)
	if x.err != nil {
		return false
	}

	var partInit *v2object.GetObjectPartInit

	switch v := resp.GetBody().GetObjectPart().(type) {
	default:
		x.err = fmt.Errorf("unexpected message instead of heading part: %T", v)
		return false
	case *v2object.SplitInfo:
		x.err = object.NewSplitInfoError(object.NewSplitInfoFromV2(v))
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

	var chunk []byte
	var lastRead int

	for {
		var resp v2object.GetResponse
		x.err = x.stream.Read(&resp)
		if x.err != nil {
			return read, false
		}

		x.err = x.client.processResponse(&resp)
		if x.err != nil {
			return read, false
		}

		part := resp.GetBody().GetObjectPart()
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

// ReadChunk reads another chunk of the object payload. Works similar to
// io.Reader.Read but returns success flag instead of error.
//
// Failure reason can be received via Close.
func (x *ObjectReader) ReadChunk(buf []byte) (int, bool) {
	return x.readChunk(buf)
}

func (x *ObjectReader) close(ignoreEOF bool) error {
	defer x.cancelCtxStream()

	if x.err != nil {
		if !errors.Is(x.err, io.EOF) {
			return x.err
		} else if !ignoreEOF {
			if x.remainingPayloadLen > 0 {
				return io.ErrUnexpectedEOF
			}

			return io.EOF
		}
	}

	return nil
}

// Close ends reading the object and returns the result of the operation
// along with the final results. Must be called after using the ObjectReader.
//
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return errors:
//   - global (see Client docs)
//   - *[object.SplitInfoError] (returned on virtual objects with PrmObjectGet.MakeRaw)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrObjectAlreadyRemoved]
//   - [apistatus.ErrSessionTokenExpired]
func (x *ObjectReader) Close() error {
	return x.close(true)
}

// Read implements io.Reader of the object payload.
func (x *ObjectReader) Read(p []byte) (int, error) {
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
//
// The call only opens the transmission channel, explicit fetching is done using the ObjectReader.
// Exactly one return value is non-nil. Resulting reader must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectGet) (object.Object, *ObjectReader, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		body  v2object.GetRequestBody
		hdr   object.Object
	)

	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return hdr, nil, err
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

	err = signServiceMessage(signer, &req)
	if err != nil {
		return hdr, nil, fmt.Errorf("sign request: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := rpcapi.GetObject(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		cancel()
		return hdr, nil, fmt.Errorf("open stream: %w", err)
	}

	var r ObjectReader
	r.cancelCtxStream = cancel
	r.stream = stream
	r.client = c

	if !r.readHeader(&hdr) {
		return hdr, nil, fmt.Errorf("header: %w", r.Close())
	}

	return hdr, &r, nil
}

// PrmObjectHead groups optional parameters of ObjectHead operation.
type PrmObjectHead struct {
	prmObjectRead

	signer neofscrypto.Signer
}

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Client default signer is used.
func (x *PrmObjectHead) UseSigner(signer neofscrypto.Signer) {
	x.signer = signer
}

// Signer returns associated with request signer.
func (x *PrmObjectHead) Signer() neofscrypto.Signer {
	return x.signer
}

// ResObjectHead groups resulting values of ObjectHead operation.
type ResObjectHead struct {
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

	obj := object.NewFromV2(&objv2)
	obj.SetID(x.idObj)

	*dst = *obj

	return true
}

// ObjectHead reads object header through a remote server using NeoFS API protocol.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
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
func (c *Client) ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectHead) (*ResObjectHead, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		body  v2object.HeadRequestBody
	)

	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return nil, err
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

	// sign the request
	err = signServiceMessage(signer, &req)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}

	resp, err := rpcapi.HeadObject(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	var res ResObjectHead
	if err = c.processResponse(resp); err != nil {
		return nil, err
	}

	_ = res.idObj.ReadFromV2(*addr.GetObjectID())

	switch v := resp.GetBody().GetHeaderPart().(type) {
	default:
		return nil, fmt.Errorf("unexpected header type %T", v)
	case *v2object.SplitInfo:
		return nil, object.NewSplitInfoError(object.NewSplitInfoFromV2(v))
	case *v2object.HeaderWithSignature:
		res.hdr = v
	}

	return &res, nil
}

// PrmObjectRange groups optional parameters of ObjectRange operation.
type PrmObjectRange struct {
	prmObjectRead

	signer neofscrypto.Signer
}

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Client default signer is used.
func (x *PrmObjectRange) UseSigner(signer neofscrypto.Signer) {
	x.signer = signer
}

// ObjectRangeReader is designed to read payload range of one object
// from NeoFS system.
//
// Must be initialized using Client.ObjectRangeInit, any other
// usage is unsafe.
type ObjectRangeReader struct {
	cancelCtxStream context.CancelFunc

	client *Client

	err error

	stream interface {
		Read(resp *v2object.GetRangeResponse) error
	}

	tailPayload []byte

	remainingPayloadLen int
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
		var resp v2object.GetRangeResponse
		x.err = x.stream.Read(&resp)
		if x.err != nil {
			return read, false
		}

		x.err = x.client.processResponse(&resp)
		if x.err != nil {
			return read, false
		}

		// get chunk message
		switch v := resp.GetBody().GetRangePart().(type) {
		default:
			x.err = fmt.Errorf("unexpected message received: %T", v)
			return read, false
		case *v2object.SplitInfo:
			x.err = object.NewSplitInfoError(object.NewSplitInfoFromV2(v))
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

// ReadChunk reads another chunk of the object payload range.
// Works similar to io.Reader.Read but returns success flag instead of error.
//
// Failure reason can be received via Close.
func (x *ObjectRangeReader) ReadChunk(buf []byte) (int, bool) {
	return x.readChunk(buf)
}

func (x *ObjectRangeReader) close(ignoreEOF bool) error {
	defer x.cancelCtxStream()

	if x.err != nil {
		if !errors.Is(x.err, io.EOF) {
			return x.err
		} else if !ignoreEOF {
			if x.remainingPayloadLen > 0 {
				return io.ErrUnexpectedEOF
			}

			return io.EOF
		}
	}

	return nil
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
	return x.close(true)
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
// Return errors:
//   - [ErrZeroRangeLength]
//   - [ErrMissingSigner]
func (c *Client) ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, prm PrmObjectRange) (*ObjectRangeReader, error) {
	var (
		addr  v2refs.Address
		cidV2 v2refs.ContainerID
		oidV2 v2refs.ObjectID
		rngV2 v2object.Range
		body  v2object.GetRangeRequestBody
	)

	if length == 0 {
		return nil, ErrZeroRangeLength
	}

	signer, err := c.getSigner(prm.signer)
	if err != nil {
		return nil, err
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

	err = signServiceMessage(signer, &req)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	stream, err := rpcapi.GetObjectRange(&c.c, &req, client.WithContext(ctx))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("open stream: %w", err)
	}

	var r ObjectRangeReader
	r.remainingPayloadLen = int(length)
	r.cancelCtxStream = cancel
	r.stream = stream
	r.client = c

	return &r, nil
}
