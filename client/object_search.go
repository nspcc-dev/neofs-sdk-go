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
	igrpc "github.com/nspcc-dev/neofs-sdk-go/internal/grpc"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	grpcprotobuf "github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"google.golang.org/grpc/mem"
)

const (
	defaultSearchObjectsQueryVersion = 1

	// MaxSearchObjectsCount is the maximal allowed number of objects requested
	// in a single [Client.SearchObjects] call.
	MaxSearchObjectsCount       = 1000
	maxSearchObjectsFilterCount = 8
	maxSearchObjectsAttrCount   = 8
)

// SearchResultItem groups data of an object matching particular search query.
type SearchResultItem struct {
	ID         oid.ID
	Attributes []string
}

// SearchObjectsOptions groups optional parameters of [Client.SearchObjects].
type SearchObjectsOptions struct {
	prmCommonMeta
	sessionToken   *session.Object
	sessionTokenV2 *sessionv2.Token
	bearerToken    *bearer.Token
	noForwarding   bool

	count uint32
}

// DisableForwarding disables request forwarding by the server and limits
// execution to its local storage. Mostly used for system purposes.
func (x *SearchObjectsOptions) DisableForwarding() { x.noForwarding = true }

// WithSessionToken specifies session token to attach to the request. The token
// must be issued for the request signer and target the requested container and
// operation.
func (x *SearchObjectsOptions) WithSessionToken(st session.Object) { x.sessionToken = &st }

// WithSessionTokenV2 specifies session token V2 to attach to the request. The token
// must be issued for the request signer and target the requested container and
// operation. V2 tokens support multiple subjects, delegation chains, and unified contexts.
func (x *SearchObjectsOptions) WithSessionTokenV2(st sessionv2.Token) { x.sessionTokenV2 = &st }

// WithBearerToken specifies bearer token to attach to the request. The token
// must be issued by the container owner for the request signer.
func (x *SearchObjectsOptions) WithBearerToken(bt bearer.Token) { x.bearerToken = &bt }

// SetCount limits the search result to a given number. Must be in [1, [client.MaxSearchObjectsCount]]
// range. Defaults to [client.MaxSearchObjectsCount].
func (x *SearchObjectsOptions) SetCount(count uint32) { x.count = count }

// Count returns limit for the search result.
func (x SearchObjectsOptions) Count() uint32 { return x.count }

// SearchObjects selects objects from a given container by applying specified
// filters, collects values of requested attributes, and returns the sorted
// result.
// SearchObjects also returns an opaque continuation cursor: when passed to a
// repeat call, it specifies where to continue the operation from. To start a
// new search, pass an empty cursor.
//
// The result is sorted lexicographically by the first attribute, then by
// object ID. When the first filter is an integer, numeric comparison is used.
// System attributes can be included using special aliases like
// [object.FilterPayloadSize].
//
// The maximum number of filters is 8. The maximum number of attributes is 8.
// If attributes are specified, the first filter must correspond to the first
// attribute. Neither filters nor attributes may contain
// [object.FilterContainerID] or [object.FilterID]. Filters using
// [object.FilterRoot] and [object.FilterPhysical] must have zero value and matcher.
//
// Note that if requested attribute is missing in the matching object, the
// corresponding element in its [SearchResultItem.Attributes] is empty.
func (c *Client) SearchObjects(ctx context.Context, cnr cid.ID, filters object.SearchFilters, attrs []string, cursor string,
	signer neofscrypto.Signer, opts SearchObjectsOptions) ([]SearchResultItem, string, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectSearchV2, time.Since(startTime), err)
		}()
	}

	switch {
	case signer == nil:
		return nil, "", ErrMissingSigner
	case opts.sessionToken != nil && opts.sessionTokenV2 != nil:
		err = errSessionTokenBothVersionsSet
		return nil, "", err
	case cnr.IsZero():
		err = cid.ErrZero
		return nil, "", err
	case opts.count > MaxSearchObjectsCount:
		err = fmt.Errorf("count is out of [1, %d] range", MaxSearchObjectsCount)
		return nil, "", err
	case len(filters) > maxSearchObjectsFilterCount:
		err = fmt.Errorf("more than %d filters", maxSearchObjectsFilterCount)
		return nil, "", err
	case len(attrs) > 0:
		if len(attrs) > maxSearchObjectsAttrCount {
			err = fmt.Errorf("more than %d attributes", maxSearchObjectsAttrCount)
			return nil, "", err
		}
		for i := range attrs {
			switch attrs[i] {
			case "":
				err = fmt.Errorf("empty attribute #%d", i)
				return nil, "", err
			case object.FilterContainerID, object.FilterID:
				err = fmt.Errorf("prohibited attribute %s", attrs[i])
				return nil, "", err
			}
			for j := i + 1; j < len(attrs); j++ {
				if attrs[i] == attrs[j] {
					err = fmt.Errorf("duplicated attribute %q", attrs[i])
					return nil, "", err
				}
			}
		}
		if len(filters) == 0 || filters[0].Header() != attrs[0] {
			err = fmt.Errorf("1st attribute %q is requested but not filtered 1st", attrs[0])
			return nil, "", err
		}
	}
	for i := range filters {
		if err = verifySearchFilter(filters[i]); err != nil {
			err = fmt.Errorf("invalid filter #%d: %w", i, err)
			return nil, "", err
		}
	}

	if opts.count == 0 {
		opts.count = MaxSearchObjectsCount
	}

	// pre-calculate body and meta header message lengths
	var sessionV1TokenMsg *protosession.SessionToken
	var sessionV1TokenLen int
	if opts.sessionToken != nil {
		sessionV1TokenMsg = opts.sessionToken.ProtoMessage()
		sessionV1TokenLen = sessionV1TokenMsg.MarshaledSize()
	}

	var bearerTokenMsg *protoacl.BearerToken
	var bearerTokenLen int
	if opts.bearerToken != nil {
		bearerTokenMsg = opts.bearerToken.ProtoMessage()
		bearerTokenLen = bearerTokenMsg.MarshaledSize()
	}

	var sessionV2TokenMsg *protosession.SessionTokenV2
	var sessionV2TokenLen int
	if opts.sessionTokenV2 != nil {
		sessionV2TokenMsg = opts.sessionTokenV2.ProtoMessage()
		sessionV2TokenLen = sessionV2TokenMsg.MarshaledSize()
	}

	bodyLen := calculateSearchObjectsRequestBodyLength(filters, len(cursor), opts.count, attrs)

	versionLen := c.apiVersion.MarshaledSize()
	xHeadersLen := calculateRequestXHeadersLength(opts.xHeaders)

	metaHdrLen := calculateRequestMetaHeaderFieldLengths(versionLen, opts.noForwarding, xHeadersLen, sessionV1TokenLen, bearerTokenLen, sessionV2TokenLen)

	bodyWithMetaHdrLen := neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestBody, bodyLen)
	bodyWithMetaHdrLen += neofsproto.SizeEmbeddedLENField(grpcprotobuf.FieldRequestMetaHeader, metaHdrLen)

	// acquire buffer for body + meta header
	bufItem := c.buffers.Get().(*[]byte)
	var reqMemBuf *igrpc.MemBuffer
	var buf []byte
	if len(*bufItem) >= bodyWithMetaHdrLen {
		// TODO: this is an extra alloc which can be avoided with pool of fix-size buffers. TBD within https://github.com/nspcc-dev/neofs-sdk-go/issues/666
		reqMemBuf = igrpc.NewMemBuffer(bufItem, c.buffers)
		buf = *bufItem
	} else {
		c.buffers.Put(bufItem)
		buf = make([]byte, bodyWithMetaHdrLen)
	}

	// encode body
	off := writeSearchObjectsRequestBody(buf, bodyLen, cnr, filters, cursor, opts.count, attrs)

	// memorize body for signing
	signedBody := buf[off-bodyLen : off]

	// encode meta header
	off += writeRequestMetaHeader(buf[off:], metaHdrLen, versionLen, c.apiVersion, opts.noForwarding, opts.xHeaders, sessionV1TokenLen, sessionV1TokenMsg, bearerTokenLen, bearerTokenMsg, sessionV2TokenLen, sessionV2TokenMsg)

	var ttl uint32 = defaultRequestTTL
	if opts.noForwarding {
		ttl = localRequestTTL
	}

	var reqBuffers mem.BufferSlice
	if c.shouldSignRequest(ttl) {
		reqBuffers, err = appendVerificationHeader(signer, reqMemBuf, buf, bodyWithMetaHdrLen, signedBody, buf[off-metaHdrLen:off], c.apiVersion)
		if err != nil {
			if reqMemBuf != nil {
				reqMemBuf.Free()
			}
			return nil, "", err
		}
	} else {
		reqSliceBuf := mem.SliceBuffer(buf[:off])
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqSliceBuf
			reqBuffers = mem.BufferSlice{reqMemBuf}
		} else {
			reqBuffers = mem.BufferSlice{reqSliceBuf}
		}
	}

	var resp protoobject.SearchV2Response
	err = callUnary(ctx, c.conn, protoobject.ObjectService_SearchV2_FullMethodName, reqBuffers, &resp)
	if err != nil {
		err = rpcErr(err)
		return nil, "", err
	}

	var statusError error
	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		if !errors.Is(err, apistatus.ErrIncomplete) {
			return nil, "", err
		}
		statusError = err
	}

	if resp.Body == nil {
		return nil, "", statusError
	}

	n := uint32(len(resp.Body.Result))
	const cursorField = "cursor"
	if n == 0 {
		if resp.Body.Cursor != "" {
			err = newErrInvalidResponseField(cursorField, errors.New("set while result is empty"))
			return nil, "", err
		}
		return nil, "", statusError
	}
	if cursor != "" && resp.Body.Cursor == cursor {
		err = newErrInvalidResponseField(cursorField, errors.New("repeats the initial one"))
		return nil, "", err
	}
	const resultField = "result"
	if n > opts.count {
		err = newErrInvalidResponseField(resultField, fmt.Errorf("more items than requested: %d", n))
		return nil, "", err
	}

	res := make([]SearchResultItem, n)
	localFilteredAttributeless := opts.noForwarding && len(attrs) == 0 && len(filters) > 0
	for i, r := range resp.Body.Result {
		switch {
		case r == nil:
			err = newErrInvalidResponseField(resultField, fmt.Errorf("nil element #%d", i))
			return nil, "", err
		case r.Id == nil:
			err = newErrInvalidResponseField(resultField, fmt.Errorf("invalid element #%d: missing ID", i))
			return nil, "", err
		case (!localFilteredAttributeless && len(r.Attributes) != len(attrs)) || (localFilteredAttributeless && len(r.Attributes) > 1):
			err = newErrInvalidResponseField(resultField, fmt.Errorf("invalid element #%d: wrong attribute count %d", i, len(r.Attributes)))
			return nil, "", err
		}
		if err = res[i].ID.FromProtoMessage(r.Id); err != nil {
			err = newErrInvalidResponseField(resultField, fmt.Errorf("invalid element #%d: invalid ID: %w", i, err))
			return nil, "", err
		}
		res[i].Attributes = r.Attributes
	}

	return res, resp.Body.Cursor, statusError
}

func verifySearchFilter(f object.SearchFilter) error {
	switch attr := f.Header(); attr {
	case "":
		return errors.New("missing attribute")
	case object.FilterContainerID, object.FilterID:
		return fmt.Errorf("prohibited attribute %s", attr)
	case object.FilterRoot, object.FilterPhysical:
		if m := f.Operation(); m != 0 {
			return fmt.Errorf("non-zero matcher %s for attribute %s", m, attr)
		}
		if val := f.Value(); val != "" {
			return fmt.Errorf("value for attribute %s is prohibited", attr)
		}
	}
	return nil
}

// PrmObjectSearch groups optional parameters of ObjectSearch operation.
type PrmObjectSearch struct {
	sessionContainer
	prmCommonMeta
	bearerToken *bearer.Token
	local       bool

	filters object.SearchFilters
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectSearch) MarkLocal() {
	x.local = true
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectSearch) WithBearerToken(t bearer.Token) {
	x.bearerToken = &t
}

// SetFilters sets filters by which to select objects. All container objects
// match unset/empty filters.
func (x *PrmObjectSearch) SetFilters(filters object.SearchFilters) {
	x.filters = filters
}

// used part of [protoobject.ObjectService_SearchClient] simplifying test
// implementations.
type searchObjectsResponseStream interface {
	// Recv reads next message with found object IDs from the stream. Recv returns
	// [io.EOF] after the server sent the last message and gracefully finished the
	// stream. Any other error means stream abort.
	Recv() (*protoobject.SearchResponse, error)
}

// ObjectListReader is designed to read list of object identifiers from NeoFS system.
//
// Must be initialized using Client.ObjectSearch, any other usage is unsafe.
type ObjectListReader struct {
	cancelCtxStream  context.CancelFunc
	err              error
	stream           searchObjectsResponseStream
	singleMsgTimeout time.Duration
	tail             []*refs.ObjectID

	statisticCallback shortStatisticCallback
	startTime         time.Time // if statisticCallback is set only
}

// Read reads another list of the object identifiers. Works similar to
// io.Reader.Read but copies oid.ID.
//
// Failure reason can be received via Close.
//
// Panics if buf has zero length.
func (x *ObjectListReader) Read(buf []oid.ID) (int, error) {
	if len(buf) == 0 {
		panic("empty buffer in ObjectListReader.ReadList")
	}

	read := copyIDBuffers(buf, x.tail)
	x.tail = x.tail[read:]

	if len(buf) == read {
		return read, nil
	}

	for {
		var resp *protoobject.SearchResponse
		x.err = dowithTimeout(x.singleMsgTimeout, x.cancelCtxStream, func() error {
			var err error
			resp, err = x.stream.Recv()
			return err
		})
		if x.err != nil {
			return read, x.err
		}

		if x.err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); x.err != nil {
			return read, x.err
		}

		// read new chunk of objects
		ids := resp.GetBody().GetIdList()
		if len(ids) == 0 {
			// just skip empty lists since they are not prohibited by protocol
			continue
		}

		ln := copyIDBuffers(buf[read:], ids)
		read += ln

		if read == len(buf) {
			// save the tail
			x.tail = append(x.tail, ids[ln:]...)

			return read, nil
		}
	}
}

func copyIDBuffers(dst []oid.ID, src []*refs.ObjectID) int {
	var i int
	for ; i < len(dst) && i < len(src); i++ {
		copy(dst[i][:], src[i].GetValue())
	}
	return i
}

// Iterate iterates over the list of found object identifiers.
// f can return true to stop iteration earlier.
//
// Returns an error if object can't be read.
func (x *ObjectListReader) Iterate(f func(oid.ID) bool) error {
	buf := make([]oid.ID, 1)

	for {
		_, err := x.Read(buf)
		if err != nil {
			return x.Close()
		}
		if f(buf[0]) {
			return nil
		}
	}
}

// Close ends reading list of the matched objects and returns the result of the operation
// along with the final results. Must be called after using the ObjectListReader.
//
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return errors:
//   - global (see Client docs)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrIncomplete]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrSessionTokenExpired]
func (x *ObjectListReader) Close() error {
	var err error
	if x.statisticCallback != nil {
		defer func() {
			x.statisticCallback(time.Since(x.startTime), err)
		}()
	}

	defer x.cancelCtxStream()

	if x.err != nil && !errors.Is(x.err, io.EOF) {
		err = x.err
		return err
	}

	return nil
}

// ObjectSearchInit initiates object selection through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching of matched objects
// is done using the ObjectListReader. Exactly one return value is non-nil.
// Resulting reader must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ObjectSearchInit(ctx context.Context, containerID cid.ID, signer user.Signer, prm PrmObjectSearch) (*ObjectListReader, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodObjectSearch, time.Since(startTime), err)
		}()
	}

	if signer == nil {
		return nil, ErrMissingSigner
	}
	if prm.session != nil && prm.sessionV2 != nil {
		return nil, errSessionTokenBothVersionsSet
	}

	req := &protoobject.SearchRequest{
		Body: &protoobject.SearchRequest_Body{
			ContainerId: containerID.ProtoMessage(),
			Version:     defaultSearchObjectsQueryVersion,
			Filters:     prm.filters.ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: c.apiVersion,
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
	if prm.sessionV2 != nil {
		req.MetaHeader.SessionTokenV2 = prm.sessionV2.ProtoMessage()
	}
	if prm.bearerToken != nil {
		req.MetaHeader.BearerToken = prm.bearerToken.ProtoMessage()
	}

	if c.shouldSignRequest(req.MetaHeader.Ttl) {
		buf := c.buffers.Get().(*[]byte)
		defer c.buffers.Put(buf)

		req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.SearchRequest_Body](signer, req, *buf)
		if err != nil {
			err = fmt.Errorf("%w: %w", errSignRequest, err)
			return nil, err
		}
	}

	var r ObjectListReader
	ctx, r.cancelCtxStream = context.WithCancel(ctx)

	r.stream, err = c.object.Search(ctx, req)
	if err != nil {
		err = fmt.Errorf("open stream: %w", err)
		return nil, err
	}
	r.singleMsgTimeout = c.streamTimeout
	if c.prm.statisticCallback != nil {
		r.startTime = time.Now()
		r.statisticCallback = func(dur time.Duration, err error) {
			c.sendStatistic(stat.MethodObjectSearchStream, dur, err)
		}
	}

	return &r, nil
}
