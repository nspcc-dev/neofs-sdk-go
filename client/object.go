package client

import (
	"bytes"
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
	"github.com/nspcc-dev/neofs-api-go/v2/signature"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	signer "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

// Object contains methods for working with objects.
type Object interface {
	// PutObject puts new object to NeoFS.
	PutObject(context.Context, *PutObjectParams, ...CallOption) (*ObjectPutRes, error)

	// DeleteObject deletes object to NeoFS.
	DeleteObject(context.Context, *DeleteObjectParams, ...CallOption) (*ObjectDeleteRes, error)

	// GetObject returns object stored in NeoFS.
	GetObject(context.Context, *GetObjectParams, ...CallOption) (*ObjectGetRes, error)

	// HeadObject returns object header.
	HeadObject(context.Context, *ObjectHeaderParams, ...CallOption) (*ObjectHeadRes, error)

	// ObjectPayloadRangeData returns range of object payload.
	ObjectPayloadRangeData(context.Context, *RangeDataParams, ...CallOption) (*ObjectRangeRes, error)

	// HashObjectPayloadRanges returns hashes of the object payload ranges from NeoFS.
	HashObjectPayloadRanges(context.Context, *RangeChecksumParams, ...CallOption) (*ObjectRangeHashRes, error)

	// SearchObjects searches for objects in NeoFS using provided parameters.
	SearchObjects(context.Context, *SearchObjectParams, ...CallOption) (*ObjectSearchRes, error)
}

type PutObjectParams struct {
	obj *object.Object

	r io.Reader
}

// ObjectAddressWriter is an interface of the
// component that writes the object address.
type ObjectAddressWriter interface {
	SetAddress(*object.Address)
}

type DeleteObjectParams struct {
	addr *object.Address

	tombTgt ObjectAddressWriter
}

type GetObjectParams struct {
	addr *object.Address

	raw bool

	w io.Writer

	readerHandler ReaderHandler
}

type ObjectHeaderParams struct {
	addr *object.Address

	raw bool

	short bool
}

type RangeDataParams struct {
	addr *object.Address

	raw bool

	r *object.Range

	w io.Writer
}

type RangeChecksumParams struct {
	tz bool

	addr *object.Address

	rs []*object.Range

	salt []byte
}

type SearchObjectParams struct {
	cid *cid.ID

	filters object.SearchFilters
}

type putObjectV2Reader struct {
	r io.Reader
}

type putObjectV2Writer struct {
	key *ecdsa.PrivateKey

	chunkPart *v2object.PutObjectPartChunk

	req *v2object.PutRequest

	stream *rpcapi.PutRequestWriter
}

type checksumType int

const (
	_ checksumType = iota
	checksumSHA256
	checksumTZ
)

const chunkSize = 3 * (1 << 20)

const TZSize = 64

const searchQueryVersion uint32 = 1

func rangesToV2(rs []*object.Range) []*v2object.Range {
	r2 := make([]*v2object.Range, 0, len(rs))

	for i := range rs {
		r2 = append(r2, rs[i].ToV2())
	}

	return r2
}

func (t checksumType) toV2() v2refs.ChecksumType {
	switch t {
	case checksumSHA256:
		return v2refs.SHA256
	case checksumTZ:
		return v2refs.TillichZemor
	default:
		panic(fmt.Sprintf("invalid checksum type %d", t))
	}
}

func (w *putObjectV2Reader) Read(p []byte) (int, error) {
	return w.r.Read(p)
}

func (w *putObjectV2Writer) Write(p []byte) (int, error) {
	w.chunkPart.SetChunk(p)

	w.req.SetVerificationHeader(nil)

	if err := signature.SignServiceMessage(w.key, w.req); err != nil {
		return 0, fmt.Errorf("could not sign chunk request message: %w", err)
	}

	if err := w.stream.Write(w.req); err != nil {
		return 0, fmt.Errorf("could not send chunk request message: %w", err)
	}

	return len(p), nil
}

func (p *PutObjectParams) WithObject(v *object.Object) *PutObjectParams {
	if p != nil {
		p.obj = v
	}

	return p
}

func (p *PutObjectParams) Object() *object.Object {
	if p != nil {
		return p.obj
	}

	return nil
}

func (p *PutObjectParams) WithPayloadReader(v io.Reader) *PutObjectParams {
	if p != nil {
		p.r = v
	}

	return p
}

func (p *PutObjectParams) PayloadReader() io.Reader {
	if p != nil {
		return p.r
	}

	return nil
}

type ObjectPutRes struct {
	statusRes

	id *object.ID
}

func (x *ObjectPutRes) setID(id *object.ID) {
	x.id = id
}

func (x ObjectPutRes) ID() *object.ID {
	return x.id
}

func (c *clientImpl) PutObject(ctx context.Context, p *PutObjectParams, opts ...CallOption) (*ObjectPutRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.PutRequest)

	// initialize request body
	body := new(v2object.PutRequestBody)
	req.SetBody(body)

	v2Addr := new(v2refs.Address)
	v2Addr.SetObjectID(p.obj.ID().ToV2())
	v2Addr.SetContainerID(p.obj.ContainerID().ToV2())

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: v2Addr,
		verb: v2session.ObjectVerbPut,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// initialize init part
	initPart := new(v2object.PutObjectPartInit)
	body.SetObjectPart(initPart)

	obj := p.obj.ToV2()

	// set init part fields
	initPart.SetObjectID(obj.GetObjectID())
	initPart.SetSignature(obj.GetSignature())
	initPart.SetHeader(obj.GetHeader())

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// open stream
	resp := new(v2object.PutResponse)

	stream, err := rpcapi.PutObject(c.Raw(), resp, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("stream opening failed: %w", err)
	}

	// send init part
	err = stream.Write(req)
	if err != nil {
		return nil, fmt.Errorf("sending the initial message to stream failed: %w", err)
	}

	// create payload bytes reader
	var rPayload io.Reader = bytes.NewReader(obj.GetPayload())
	if p.r != nil {
		rPayload = io.MultiReader(rPayload, p.r)
	}

	// create v2 payload stream writer
	chunkPart := new(v2object.PutObjectPartChunk)
	body.SetObjectPart(chunkPart)

	w := &putObjectV2Writer{
		key:       callOpts.key,
		chunkPart: chunkPart,
		req:       req,
		stream:    stream,
	}

	r := &putObjectV2Reader{r: rPayload}

	// copy payload from reader to stream writer
	_, err = io.CopyBuffer(w, r, make([]byte, chunkSize))
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("payload streaming failed: %w", err)
	}

	// close object stream and receive response from remote node
	err = stream.Close()
	if err != nil {
		return nil, fmt.Errorf("closing the stream failed: %w", err)
	}

	var (
		res     = new(ObjectPutRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	// convert object identifier
	id := object.NewIDFromV2(resp.GetBody().GetObjectID())

	res.setID(id)

	return res, nil
}

func (p *DeleteObjectParams) WithAddress(v *object.Address) *DeleteObjectParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *DeleteObjectParams) Address() *object.Address {
	if p != nil {
		return p.addr
	}

	return nil
}

// WithTombstoneAddressTarget sets target component to write tombstone address.
func (p *DeleteObjectParams) WithTombstoneAddressTarget(v ObjectAddressWriter) *DeleteObjectParams {
	if p != nil {
		p.tombTgt = v
	}

	return p
}

// TombstoneAddressTarget returns target component to write tombstone address.
func (p *DeleteObjectParams) TombstoneAddressTarget() ObjectAddressWriter {
	if p != nil {
		return p.tombTgt
	}

	return nil
}

type ObjectDeleteRes struct {
	statusRes

	tombAddr *object.Address
}

func (x ObjectDeleteRes) TombstoneAddress() *object.Address {
	return x.tombAddr
}

func (x *ObjectDeleteRes) setTombstoneAddress(addr *object.Address) {
	x.tombAddr = addr
}

// DeleteObject removes object by address.
//
// If target of tombstone address is not set, the address is ignored.
func (c *clientImpl) DeleteObject(ctx context.Context, p *DeleteObjectParams, opts ...CallOption) (*ObjectDeleteRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.DeleteRequest)

	// initialize request body
	body := new(v2object.DeleteRequestBody)
	req.SetBody(body)

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: p.addr.ToV2(),
		verb: v2session.ObjectVerbDelete,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// fill body fields
	body.SetAddress(p.addr.ToV2())

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// send request
	resp, err := rpcapi.DeleteObject(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("sending the request failed: %w", err)
	}

	var (
		res     = new(ObjectDeleteRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	addrv2 := resp.GetBody().GetTombstone()

	res.setTombstoneAddress(object.NewAddressFromV2(addrv2))

	return res, nil
}

func (p *GetObjectParams) WithAddress(v *object.Address) *GetObjectParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *GetObjectParams) Address() *object.Address {
	if p != nil {
		return p.addr
	}

	return nil
}

func (p *GetObjectParams) WithPayloadWriter(w io.Writer) *GetObjectParams {
	if p != nil {
		p.w = w
	}

	return p
}

func (p *GetObjectParams) PayloadWriter() io.Writer {
	if p != nil {
		return p.w
	}

	return nil
}

func (p *GetObjectParams) WithRawFlag(v bool) *GetObjectParams {
	if p != nil {
		p.raw = v
	}

	return p
}

func (p *GetObjectParams) RawFlag() bool {
	if p != nil {
		return p.raw
	}

	return false
}

// ReaderHandler is a function over io.Reader.
type ReaderHandler func(io.Reader)

// WithPayloadReaderHandler sets handler of the payload reader.
//
// If provided, payload reader is composed after receiving the header.
// In this case payload writer set via WithPayloadWriter is ignored.
//
// Handler should not be nil.
func (p *GetObjectParams) WithPayloadReaderHandler(f ReaderHandler) *GetObjectParams {
	if p != nil {
		p.readerHandler = f
	}

	return p
}

// wrapper over the Object Get stream that provides io.Reader.
type objectPayloadReader struct {
	stream interface {
		Read(*v2object.GetResponse) error
	}

	resp v2object.GetResponse

	tail []byte
}

func (x *objectPayloadReader) Read(p []byte) (read int, err error) {
	// read remaining tail
	read = copy(p, x.tail)

	x.tail = x.tail[read:]

	if len(p)-read == 0 {
		return
	}

	// receive message from server stream
	err = x.stream.Read(&x.resp)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.EOF
			return
		}

		err = fmt.Errorf("reading the response failed: %w", err)
		return
	}

	// get chunk part message
	part := x.resp.GetBody().GetObjectPart()

	chunkPart, ok := part.(*v2object.GetObjectPartChunk)
	if !ok {
		err = errWrongMessageSeq
		return
	}

	// verify response structure
	if err = signature.VerifyServiceMessage(&x.resp); err != nil {
		err = fmt.Errorf("response verification failed: %w", err)
		return
	}

	// read new chunk
	chunk := chunkPart.GetChunk()

	tailOffset := copy(p[read:], chunk)

	read += tailOffset

	// save the tail
	x.tail = append(x.tail, chunk[tailOffset:]...)

	return
}

var errWrongMessageSeq = errors.New("incorrect message sequence")

type ObjectGetRes struct {
	statusRes
	objectRes
}

type objectRes struct {
	obj *object.Object
}

func (x *objectRes) setObject(obj *object.Object) {
	x.obj = obj
}

func (x objectRes) Object() *object.Object {
	return x.obj
}

func writeUnexpectedMessageTypeErr(res resCommon, val interface{}) {
	var st apistatus.ServerInternal // specific API status should be used

	apistatus.WriteInternalServerErr(&st, fmt.Errorf("unexpected message type %T", val))

	res.setStatus(st)
}

func (c *clientImpl) GetObject(ctx context.Context, p *GetObjectParams, opts ...CallOption) (*ObjectGetRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.GetRequest)

	// initialize request body
	body := new(v2object.GetRequestBody)
	req.SetBody(body)

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: p.addr.ToV2(),
		verb: v2session.ObjectVerbGet,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// fill body fields
	body.SetAddress(p.addr.ToV2())
	body.SetRaw(p.raw)

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// open stream
	stream, err := rpcapi.GetObject(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("stream opening failed: %w", err)
	}

	var (
		headWas bool
		payload []byte
		obj     = new(v2object.Object)
		resp    = new(v2object.GetResponse)

		messageWas bool

		res     = new(ObjectGetRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

loop:
	for {
		// receive message from server stream
		err := stream.Read(resp)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !messageWas {
					return nil, errWrongMessageSeq
				}

				break
			}

			return nil, fmt.Errorf("reading the response failed: %w", err)
		}

		messageWas = true

		// process response in general
		if c.processResponseV2(&procRes, procPrm) {
			if procRes.cliErr != nil {
				return nil, procRes.cliErr
			}

			return res, nil
		}

		switch v := resp.GetBody().GetObjectPart().(type) {
		default:
			return nil, errWrongMessageSeq
		case *v2object.GetObjectPartInit:
			if headWas {
				return nil, errWrongMessageSeq
			}

			headWas = true

			obj.SetObjectID(v.GetObjectID())
			obj.SetSignature(v.GetSignature())

			hdr := v.GetHeader()
			obj.SetHeader(hdr)

			if p.readerHandler != nil {
				p.readerHandler(&objectPayloadReader{
					stream: stream,
				})

				break loop
			}

			if p.w == nil {
				payload = make([]byte, 0, hdr.GetPayloadLength())
			}
		case *v2object.GetObjectPartChunk:
			if !headWas {
				return nil, errWrongMessageSeq
			}

			if p.w != nil {
				if _, err := p.w.Write(v.GetChunk()); err != nil {
					return nil, fmt.Errorf("could not write payload chunk: %w", err)
				}
			} else {
				payload = append(payload, v.GetChunk()...)
			}
		case *v2object.SplitInfo:
			if headWas {
				return nil, errWrongMessageSeq
			}

			si := object.NewSplitInfoFromV2(v)
			return nil, object.NewSplitInfoError(si)
		}
	}

	obj.SetPayload(payload)

	// convert the object
	res.setObject(object.NewFromV2(obj))

	return res, nil
}

func (p *ObjectHeaderParams) WithAddress(v *object.Address) *ObjectHeaderParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *ObjectHeaderParams) Address() *object.Address {
	if p != nil {
		return p.addr
	}

	return nil
}

func (p *ObjectHeaderParams) WithAllFields() *ObjectHeaderParams {
	if p != nil {
		p.short = false
	}

	return p
}

// AllFields return true if parameter set to return all header fields, returns
// false if parameter set to return only main fields of header.
func (p *ObjectHeaderParams) AllFields() bool {
	if p != nil {
		return !p.short
	}

	return false
}

func (p *ObjectHeaderParams) WithMainFields() *ObjectHeaderParams {
	if p != nil {
		p.short = true
	}

	return p
}

func (p *ObjectHeaderParams) WithRawFlag(v bool) *ObjectHeaderParams {
	if p != nil {
		p.raw = v
	}

	return p
}

func (p *ObjectHeaderParams) RawFlag() bool {
	if p != nil {
		return p.raw
	}

	return false
}

type ObjectHeadRes struct {
	statusRes
	objectRes
}

func (c *clientImpl) HeadObject(ctx context.Context, p *ObjectHeaderParams, opts ...CallOption) (*ObjectHeadRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.HeadRequest)

	// initialize request body
	body := new(v2object.HeadRequestBody)
	req.SetBody(body)

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: p.addr.ToV2(),
		verb: v2session.ObjectVerbHead,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// fill body fields
	body.SetAddress(p.addr.ToV2())
	body.SetMainOnly(p.short)
	body.SetRaw(p.raw)

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// send Head request
	resp, err := rpcapi.HeadObject(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("sending the request failed: %w", err)
	}

	var (
		res     = new(ObjectHeadRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	var (
		hdr   *v2object.Header
		idSig *v2refs.Signature
	)

	switch v := resp.GetBody().GetHeaderPart().(type) {
	case nil:
		writeUnexpectedMessageTypeErr(res, v)
		return res, nil
	case *v2object.ShortHeader:
		if !p.short {
			writeUnexpectedMessageTypeErr(res, v)
			return res, nil
		}

		h := v

		hdr = new(v2object.Header)
		hdr.SetPayloadLength(h.GetPayloadLength())
		hdr.SetVersion(h.GetVersion())
		hdr.SetOwnerID(h.GetOwnerID())
		hdr.SetObjectType(h.GetObjectType())
		hdr.SetCreationEpoch(h.GetCreationEpoch())
		hdr.SetPayloadHash(h.GetPayloadHash())
		hdr.SetHomomorphicHash(h.GetHomomorphicHash())
	case *v2object.HeaderWithSignature:
		if p.short {
			writeUnexpectedMessageTypeErr(res, v)
			return res, nil
		}

		hdr = v.GetHeader()
		idSig = v.GetSignature()
	case *v2object.SplitInfo:
		si := object.NewSplitInfoFromV2(v)

		return nil, object.NewSplitInfoError(si)
	}

	obj := new(v2object.Object)
	obj.SetHeader(hdr)
	obj.SetSignature(idSig)

	raw := object.NewRawFromV2(obj)
	raw.SetID(p.addr.ObjectID())

	res.setObject(raw.Object())

	return res, nil
}

func (p *RangeDataParams) WithAddress(v *object.Address) *RangeDataParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *RangeDataParams) Address() *object.Address {
	if p != nil {
		return p.addr
	}

	return nil
}

func (p *RangeDataParams) WithRaw(v bool) *RangeDataParams {
	if p != nil {
		p.raw = v
	}

	return p
}

func (p *RangeDataParams) Raw() bool {
	if p != nil {
		return p.raw
	}

	return false
}

func (p *RangeDataParams) WithRange(v *object.Range) *RangeDataParams {
	if p != nil {
		p.r = v
	}

	return p
}

func (p *RangeDataParams) Range() *object.Range {
	if p != nil {
		return p.r
	}

	return nil
}

func (p *RangeDataParams) WithDataWriter(v io.Writer) *RangeDataParams {
	if p != nil {
		p.w = v
	}

	return p
}

func (p *RangeDataParams) DataWriter() io.Writer {
	if p != nil {
		return p.w
	}

	return nil
}

type ObjectRangeRes struct {
	statusRes

	data []byte
}

func (x *ObjectRangeRes) setData(data []byte) {
	x.data = data
}

func (x ObjectRangeRes) Data() []byte {
	return x.data
}

func (c *clientImpl) ObjectPayloadRangeData(ctx context.Context, p *RangeDataParams, opts ...CallOption) (*ObjectRangeRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.GetRangeRequest)

	// initialize request body
	body := new(v2object.GetRangeRequestBody)
	req.SetBody(body)

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: p.addr.ToV2(),
		verb: v2session.ObjectVerbRange,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// fill body fields
	body.SetAddress(p.addr.ToV2())
	body.SetRange(p.r.ToV2())
	body.SetRaw(p.raw)

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// open stream
	stream, err := rpcapi.GetObjectRange(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not create Get payload range stream: %w", err)
	}

	var payload []byte
	if p.w != nil {
		payload = make([]byte, 0, p.r.GetLength())
	}

	var (
		resp = new(v2object.GetRangeResponse)

		chunkWas, messageWas bool

		res     = new(ObjectRangeRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

	for {
		// receive message from server stream
		err := stream.Read(resp)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !messageWas {
					return nil, errWrongMessageSeq
				}

				break
			}

			return nil, fmt.Errorf("reading the response failed: %w", err)
		}

		messageWas = true

		// process response in general
		if c.processResponseV2(&procRes, procPrm) {
			if procRes.cliErr != nil {
				return nil, procRes.cliErr
			}

			return res, nil
		}

		switch v := resp.GetBody().GetRangePart().(type) {
		case nil:
			writeUnexpectedMessageTypeErr(res, v)
			return res, nil
		case *v2object.GetRangePartChunk:
			chunkWas = true

			if p.w != nil {
				if _, err = p.w.Write(v.GetChunk()); err != nil {
					return nil, fmt.Errorf("could not write payload chunk: %w", err)
				}
			} else {
				payload = append(payload, v.GetChunk()...)
			}
		case *v2object.SplitInfo:
			if chunkWas {
				return nil, errWrongMessageSeq
			}

			si := object.NewSplitInfoFromV2(v)

			return nil, object.NewSplitInfoError(si)
		}
	}

	res.setData(payload)

	return res, nil
}

func (p *RangeChecksumParams) WithAddress(v *object.Address) *RangeChecksumParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *RangeChecksumParams) Address() *object.Address {
	if p != nil {
		return p.addr
	}

	return nil
}

func (p *RangeChecksumParams) WithRangeList(rs ...*object.Range) *RangeChecksumParams {
	if p != nil {
		p.rs = rs
	}

	return p
}

func (p *RangeChecksumParams) RangeList() []*object.Range {
	if p != nil {
		return p.rs
	}

	return nil
}

func (p *RangeChecksumParams) WithSalt(v []byte) *RangeChecksumParams {
	if p != nil {
		p.salt = v
	}

	return p
}

func (p *RangeChecksumParams) Salt() []byte {
	if p != nil {
		return p.salt
	}

	return nil
}

func (p *RangeChecksumParams) TZ() *RangeChecksumParams {
	p.tz = true
	return p
}

type ObjectRangeHashRes struct {
	statusRes

	hashes [][]byte
}

func (x *ObjectRangeHashRes) setHashes(v [][]byte) {
	x.hashes = v
}

func (x ObjectRangeHashRes) Hashes() [][]byte {
	return x.hashes
}

func (c *clientImpl) HashObjectPayloadRanges(ctx context.Context, p *RangeChecksumParams, opts ...CallOption) (*ObjectRangeHashRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.GetRangeHashRequest)

	// initialize request body
	body := new(v2object.GetRangeHashRequestBody)
	req.SetBody(body)

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: p.addr.ToV2(),
		verb: v2session.ObjectVerbRangeHash,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// fill body fields
	body.SetAddress(p.addr.ToV2())
	body.SetSalt(p.salt)

	typ := checksumSHA256
	if p.tz {
		typ = checksumTZ
	}

	typV2 := typ.toV2()
	body.SetType(typV2)

	rsV2 := rangesToV2(p.rs)
	body.SetRanges(rsV2)

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// send request
	resp, err := rpcapi.HashObjectRange(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("sending the request failed: %w", err)
	}

	var (
		res     = new(ObjectRangeHashRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	res.setHashes(resp.GetBody().GetHashList())

	return res, nil
}

func (p *SearchObjectParams) WithContainerID(v *cid.ID) *SearchObjectParams {
	if p != nil {
		p.cid = v
	}

	return p
}

func (p *SearchObjectParams) ContainerID() *cid.ID {
	if p != nil {
		return p.cid
	}

	return nil
}

func (p *SearchObjectParams) WithSearchFilters(v object.SearchFilters) *SearchObjectParams {
	if p != nil {
		p.filters = v
	}

	return p
}

func (p *SearchObjectParams) SearchFilters() object.SearchFilters {
	if p != nil {
		return p.filters
	}

	return nil
}

type ObjectSearchRes struct {
	statusRes

	ids []*object.ID
}

func (x *ObjectSearchRes) setIDList(v []*object.ID) {
	x.ids = v
}

func (x ObjectSearchRes) IDList() []*object.ID {
	return x.ids
}

func (c *clientImpl) SearchObjects(ctx context.Context, p *SearchObjectParams, opts ...CallOption) (*ObjectSearchRes, error) {
	callOpts := c.defaultCallOptions()

	for i := range opts {
		if opts[i] != nil {
			opts[i](callOpts)
		}
	}

	// create request
	req := new(v2object.SearchRequest)

	// initialize request body
	body := new(v2object.SearchRequestBody)
	req.SetBody(body)

	v2Addr := new(v2refs.Address)
	v2Addr.SetContainerID(p.cid.ToV2())

	// set meta header
	meta := v2MetaHeaderFromOpts(callOpts)

	if err := c.attachV2SessionToken(callOpts, meta, v2SessionReqInfo{
		addr: v2Addr,
		verb: v2session.ObjectVerbSearch,
	}); err != nil {
		return nil, fmt.Errorf("could not attach session token: %w", err)
	}

	req.SetMetaHeader(meta)

	// fill body fields
	body.SetContainerID(v2Addr.GetContainerID())
	body.SetVersion(searchQueryVersion)
	body.SetFilters(p.filters.ToV2())

	// sign the request
	if err := signature.SignServiceMessage(callOpts.key, req); err != nil {
		return nil, fmt.Errorf("signing the request failed: %w", err)
	}

	// create search stream
	stream, err := rpcapi.SearchObjects(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("stream opening failed: %w", err)
	}

	var (
		searchResult []*object.ID
		resp         = new(v2object.SearchResponse)

		messageWas bool

		res     = new(ObjectSearchRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOpts
	procPrm.resp = resp

	procRes.statusRes = res

	for {
		// receive message from server stream
		err := stream.Read(resp)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !messageWas {
					return nil, errWrongMessageSeq
				}

				break
			}

			return nil, fmt.Errorf("reading the response failed: %w", err)
		}

		messageWas = true

		// process response in general
		if c.processResponseV2(&procRes, procPrm) {
			if procRes.cliErr != nil {
				return nil, procRes.cliErr
			}

			return res, nil
		}

		chunk := resp.GetBody().GetIDList()
		for i := range chunk {
			searchResult = append(searchResult, object.NewIDFromV2(chunk[i]))
		}
	}

	res.setIDList(searchResult)

	return res, nil
}

func (c *clientImpl) attachV2SessionToken(opts *callOptions, hdr *v2session.RequestMetaHeader, info v2SessionReqInfo) error {
	if opts.session == nil {
		return nil
	}

	// Do not resign already prepared session token
	if opts.session.Signature() != nil {
		hdr.SetSessionToken(opts.session.ToV2())
		return nil
	}

	opCtx := new(v2session.ObjectSessionContext)
	opCtx.SetAddress(info.addr)
	opCtx.SetVerb(info.verb)

	lt := new(v2session.TokenLifetime)
	lt.SetIat(info.iat)
	lt.SetNbf(info.nbf)
	lt.SetExp(info.exp)

	body := new(v2session.SessionTokenBody)
	body.SetID(opts.session.ID())
	body.SetOwnerID(opts.session.OwnerID().ToV2())
	body.SetSessionKey(opts.session.SessionKey())
	body.SetContext(opCtx)
	body.SetLifetime(lt)

	token := new(v2session.SessionToken)
	token.SetBody(body)

	signWrapper := signature.StableMarshalerWrapper{SM: token.GetBody()}

	err := signer.SignDataWithHandler(opts.key, signWrapper, func(key []byte, sig []byte) {
		sessionTokenSignature := new(v2refs.Signature)
		sessionTokenSignature.SetKey(key)
		sessionTokenSignature.SetSign(sig)
		token.SetSignature(sessionTokenSignature)
	})
	if err != nil {
		return err
	}

	hdr.SetSessionToken(token)

	return nil
}
