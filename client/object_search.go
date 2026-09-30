package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	grpcprotobuf "github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"google.golang.org/grpc/mem"
)

const (
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
	sessionToken      *session.Object
	sessionTokenV2    *sessionv2.Token
	bearerToken       *bearer.Token
	noForwarding      bool
	containerRevision *uint64

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

// AttachContainerRevision allows attaching container revision to the request.
// If server's revision differs, [apistatus.ErrContainerRevisionMismatch] error
// will be returned. If revision has been attached, but server does not support
// container revisions, an error will be returned.
func (x *SearchObjectsOptions) AttachContainerRevision(revision uint64) {
	x.containerRevision = &revision
}

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
// Call supports container revision state check, see [SearchObjectsOptions.AttachContainerRevision].
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

	var cnrRev uint64
	if opts.containerRevision != nil {
		if !containerRevisionsSupported(c.apiVersion) {
			return nil, "", unsupportedCnrRevErr(c.apiVersion)
		}
		cnrRev = *opts.containerRevision
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

	filterLenFn := func(i int) int {
		return protoobject.CalculateSearchFilterLength(filters[i].Operation(), filters[i].Header(), filters[i].Value())
	}

	bodyLen := protoobject.CalculateSearchV2RequestBodyLength(1, cursor, opts.count, attrs, len(filters), filterLenFn, cnrRev)

	ttl := localFlagToTTL(opts.noForwarding)
	xHdrLenFn := xHeadersLengthFunc(opts.xHeaders)
	xHdrNum := len(opts.xHeaders) / 2

	metaHdrLen := protosession.CalculateRequestMetaHeaderLength(c.apiVersion.Major, c.apiVersion.Minor, ttl, xHdrNum, xHdrLenFn, sessionV1TokenLen, bearerTokenLen, 0, sessionV2TokenLen)

	bodyWithMetaHdrLen := protoencoding.CalculateRequestBodyWithMetaHeaderLength(bodyLen, metaHdrLen)

	// acquire buffer for body + meta header
	var reqMemBuf *grpcprotobuf.MemBuffer
	var buf []byte
	if bodyWithMetaHdrLen <= defaultRequestBufferLength {
		reqMemBuf = defaultRequestBufferPool.Get()
		buf = reqMemBuf.SliceBuffer
	} else {
		buf = make([]byte, bodyWithMetaHdrLen)
	}

	// encode body
	writeFilterFn := func(buf []byte, i int) int {
		return protoobject.WriteSearchFilter(buf, filters[i].Operation(), filters[i].Header(), filters[i].Value())
	}
	off := protoobject.WriteSearchV2RequestBodyToRequest(buf, cnr, 1, cursor, opts.count, attrs, len(filters), filterLenFn, writeFilterFn, cnrRev)

	// memorize body for signing
	signedBody := buf[off-bodyLen : off]

	// encode meta header
	writeXHeaderFn := writeXHeaderFunc(opts.xHeaders)
	writeSessionV1TokenFn := protoencoding.WriteStablyMarshalledMessageFunc(sessionV1TokenMsg)
	writeBearerTokenFn := protoencoding.WriteStablyMarshalledMessageFunc(bearerTokenMsg)
	writeSessionV2TokenFn := protoencoding.WriteStablyMarshalledMessageFunc(sessionV2TokenMsg)

	off += protosession.WriteRequestMetaHeaderToRequest(buf[off:], c.apiVersion.Major, c.apiVersion.Minor, ttl, xHdrNum, xHdrLenFn, writeXHeaderFn, sessionV1TokenLen, writeSessionV1TokenFn, bearerTokenLen, writeBearerTokenFn, 0, sessionV2TokenLen, writeSessionV2TokenFn)

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
		if reqMemBuf != nil {
			reqMemBuf.SetBounds(0, off)
			reqBuffers = mem.BufferSlice{reqMemBuf}
		} else {
			reqBuffers = mem.BufferSlice{mem.SliceBuffer(buf[:off])}
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
