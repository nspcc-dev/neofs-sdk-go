package client

import (
	"context"
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
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	signer "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

// ObjectAddressWriter is an interface of the
// component that writes the object address.
type ObjectAddressWriter interface {
	SetAddress(*address.Address)
}

type DeleteObjectParams struct {
	addr *address.Address

	tombTgt ObjectAddressWriter
}

type ObjectHeaderParams struct {
	addr *address.Address

	raw bool

	short bool
}

type RangeDataParams struct {
	addr *address.Address

	raw bool

	r *object.Range

	w io.Writer
}

type RangeChecksumParams struct {
	tz bool

	addr *address.Address

	rs []*object.Range

	salt []byte
}

type SearchObjectParams struct {
	cid *cid.ID

	filters object.SearchFilters
}

type checksumType int

const (
	_ checksumType = iota
	checksumSHA256
	checksumTZ
)

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

func (p *DeleteObjectParams) WithAddress(v *address.Address) *DeleteObjectParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *DeleteObjectParams) Address() *address.Address {
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

	tombAddr *address.Address
}

func (x ObjectDeleteRes) TombstoneAddress() *address.Address {
	return x.tombAddr
}

func (x *ObjectDeleteRes) setTombstoneAddress(addr *address.Address) {
	x.tombAddr = addr
}

// DeleteObject removes object by address.
//
// If target of tombstone address is not set, the address is ignored.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) DeleteObject(ctx context.Context, p *DeleteObjectParams, opts ...CallOption) (*ObjectDeleteRes, error) {
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

	res.setTombstoneAddress(address.NewAddressFromV2(addrv2))

	return res, nil
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

func (p *ObjectHeaderParams) WithAddress(v *address.Address) *ObjectHeaderParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *ObjectHeaderParams) Address() *address.Address {
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

// HeadObject receives object's header through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) HeadObject(ctx context.Context, p *ObjectHeaderParams, opts ...CallOption) (*ObjectHeadRes, error) {
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

func (p *RangeDataParams) WithAddress(v *address.Address) *RangeDataParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *RangeDataParams) Address() *address.Address {
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

// ObjectPayloadRangeData receives object's range payload data through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) ObjectPayloadRangeData(ctx context.Context, p *RangeDataParams, opts ...CallOption) (*ObjectRangeRes, error) {
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

func (p *RangeChecksumParams) WithAddress(v *address.Address) *RangeChecksumParams {
	if p != nil {
		p.addr = v
	}

	return p
}

func (p *RangeChecksumParams) Address() *address.Address {
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

// HashObjectPayloadRanges receives range hash of the object
// payload data through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) HashObjectPayloadRanges(ctx context.Context, p *RangeChecksumParams, opts ...CallOption) (*ObjectRangeHashRes, error) {
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

	ids []*oid.ID
}

func (x *ObjectSearchRes) setIDList(v []*oid.ID) {
	x.ids = v
}

func (x ObjectSearchRes) IDList() []*oid.ID {
	return x.ids
}

// SearchObjects searches for the objects through NeoFS API call.
//
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
func (c *Client) SearchObjects(ctx context.Context, p *SearchObjectParams, opts ...CallOption) (*ObjectSearchRes, error) {
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
		searchResult []*oid.ID
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
			searchResult = append(searchResult, oid.NewIDFromV2(chunk[i]))
		}
	}

	res.setIDList(searchResult)

	return res, nil
}

func (c *Client) attachV2SessionToken(opts *callOptions, hdr *v2session.RequestMetaHeader, info v2SessionReqInfo) error {
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

	body := new(v2session.SessionTokenBody)
	body.SetID(opts.session.ID())
	body.SetOwnerID(opts.session.OwnerID().ToV2())
	body.SetSessionKey(opts.session.SessionKey())
	body.SetContext(opCtx)

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
