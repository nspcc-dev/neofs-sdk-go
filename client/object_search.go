package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

const (
	defaultSearchObjectsQueryVersion = 1

	maxSearchObjectsCount       = 1000
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
	sessionToken *session.Object
	bearerToken  *bearer.Token
	noForwarding bool

	count uint32
}

// DisableForwarding disables request forwarding by the server and limits
// execution to its local storage. Mostly used for system purposes.
func (x *SearchObjectsOptions) DisableForwarding() { x.noForwarding = true }

// WithSessionToken specifies session token to attach to the request. The token
// must be issued for the request signer and target the requested container and
// operation.
func (x *SearchObjectsOptions) WithSessionToken(st session.Object) { x.sessionToken = &st }

// WithBearerToken specifies bearer token to attach to the request. The token
// must be issued by the container owner for the request signer.
func (x *SearchObjectsOptions) WithBearerToken(bt bearer.Token) { x.bearerToken = &bt }

// SetCount limits the search result to a given number. Must be in [1, 1000]
// range. Defaults to 1000.
func (x *SearchObjectsOptions) SetCount(count uint32) { x.count = count }

// SearchObjects selects objects from given container by applying specified
// filters, collects values of requested attributes and returns the result
// sorted. Elements are compared by attributes' values lexicographically in
// priority from first to last, closing with the default sorting by IDs. System
// attributes can be included using special aliases like
// [object.FilterPayloadSize]. SearchObjects also returns opaque continuation
// cursor: when passed to a repeat call, it specifies where to continue the
// operation from. To start the search anew, pass an empty cursor.
//
// Max number of filters is 8. Max number of attributes is 8. If attributes are
// specified, filters must include the 1st of them.
//
// Note that if requested attribute is missing in the matching object,
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
	case cnr.IsZero():
		err = cid.ErrZero
		return nil, "", err
	case opts.count > maxSearchObjectsCount:
		err = fmt.Errorf("count is out of [1, %d] range", maxSearchObjectsCount)
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
			if attrs[i] == "" {
				err = fmt.Errorf("empty attribute #%d", i)
				return nil, "", err
			}
			for j := i + 1; j < len(attrs); j++ {
				if attrs[i] == attrs[j] {
					err = fmt.Errorf("duplicated attribute %q", attrs[i])
					return nil, "", err
				}
			}
		}
		if !slices.ContainsFunc(filters, func(f object.SearchFilter) bool { return f.Header() == attrs[0] }) {
			err = fmt.Errorf("attribute %q is requested but not filtered", attrs[0])
			return nil, "", err
		}
	}

	if opts.count == 0 {
		opts.count = maxSearchObjectsCount
	}

	req := &protoobject.SearchV2Request{
		Body: &protoobject.SearchV2Request_Body{
			ContainerId: cnr.ProtoMessage(),
			Version:     defaultSearchObjectsQueryVersion,
			Filters:     filters.ProtoMessage(),
			Cursor:      cursor,
			Count:       opts.count,
			Attributes:  attrs,
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
		},
	}
	writeXHeadersToMeta(opts.xHeaders, req.MetaHeader)
	if opts.noForwarding {
		req.MetaHeader.Ttl = localRequestTTL
	} else {
		req.MetaHeader.Ttl = defaultRequestTTL
	}
	if opts.sessionToken != nil {
		req.MetaHeader.SessionToken = opts.sessionToken.ProtoMessage()
	}
	if opts.bearerToken != nil {
		req.MetaHeader.BearerToken = opts.bearerToken.ProtoMessage()
	}

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.SearchV2Request_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, "", err
	}

	resp, err := c.object.SearchV2(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return nil, "", err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return nil, "", err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protoobject.SearchV2Response_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return nil, "", err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return nil, "", err
	}

	if resp.Body == nil {
		return nil, "", nil
	}

	n := uint32(len(resp.Body.Result))
	const cursorField = "cursor"
	if n == 0 {
		if resp.Body.Cursor != "" {
			err = newErrInvalidResponseField(cursorField, errors.New("set while result is empty"))
			return nil, "", err
		}
		return nil, "", nil
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
	for i, r := range resp.Body.Result {
		switch {
		case r == nil:
			err = newErrInvalidResponseField(resultField, fmt.Errorf("nil element #%d", i))
			return nil, "", err
		case r.Id == nil:
			err = newErrInvalidResponseField(resultField, fmt.Errorf("invalid element #%d: missing ID", i))
			return nil, "", err
		case len(r.Attributes) != len(attrs):
			err = newErrInvalidResponseField(resultField, fmt.Errorf("invalid element #%d: wrong attribute count %d", i, len(r.Attributes)))
			return nil, "", err
		}
		if err = res[i].ID.FromProtoMessage(r.Id); err != nil {
			err = newErrInvalidResponseField(resultField, fmt.Errorf("invalid element #%d: invalid ID: %w", i, err))
			return nil, "", err
		}
		res[i].Attributes = r.Attributes
	}

	return res, resp.Body.Cursor, nil
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

		if x.err = neofscrypto.VerifyResponseWithBuffer[*protoobject.SearchResponse_Body](resp, nil); x.err != nil {
			x.err = fmt.Errorf("%w: %w", errResponseSignatures, x.err)
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

	req := &protoobject.SearchRequest{
		Body: &protoobject.SearchRequest_Body{
			ContainerId: containerID.ProtoMessage(),
			Version:     defaultSearchObjectsQueryVersion,
			Filters:     prm.filters.ProtoMessage(),
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

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.SearchRequest_Body](signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
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
