package object

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// FilterOp defines the matching property.
type FilterOp uint32

// Supported FilterOp values.
const (
	_                    FilterOp = iota
	FilterOpEQ                    // String 'equal'
	FilterOpNE                    // String 'not equal'
	FilterOpNotPresent            // Missing property
	FilterOpCommonPrefix          // Prefix matches in strings
	FilterOpGT                    // Numeric 'greater than'
	FilterOpGE                    // Numeric 'greater or equal than'
	FilterOpLT                    // Numeric 'less than'
	FilterOpLE                    // Numeric 'less or equal than'
)

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (x FilterOp) String() string {
	switch x {
	default:
		return fmt.Sprintf("UNKNOWN#%d", x)
	case FilterOpEQ:
		return "STRING_EQUAL"
	case FilterOpNE:
		return "STRING_NOT_EQUAL"
	case FilterOpNotPresent:
		return "NOT_PRESENT"
	case FilterOpCommonPrefix:
		return "COMMON_PREFIX"
	case FilterOpGT:
		return "NUMERIC_GT"
	case FilterOpGE:
		return "NUMERIC_GE"
	case FilterOpLT:
		return "NUMERIC_LT"
	case FilterOpLE:
		return "NUMERIC_LE"
	}
}

// SearchFilter describes object property filter.
type SearchFilter struct {
	key   string
	value string
	op    FilterOp
}

const systemFilterPrefix = "$Object:"

// Various filters by object header.
const (
	FilterID                     = systemFilterPrefix + "objectID"
	FilterOwnerID                = systemFilterPrefix + "ownerID"
	FilterPayloadChecksum        = systemFilterPrefix + "payloadHash"
	FilterType                   = systemFilterPrefix + "objectType"
	FilterPayloadHomomorphicHash = systemFilterPrefix + "homomorphicHash"
	FilterParentID               = systemFilterPrefix + "split.parent"
	FilterSplitID                = systemFilterPrefix + "split.splitID"
	FilterFirstSplitObject       = systemFilterPrefix + "split.first"
	FilterCreationEpoch          = systemFilterPrefix + "creationEpoch"
	FilterPayloadSize            = systemFilterPrefix + "payloadLength"
)

// Various filters to match certain object properties.
const (
	// FilterRoot filters objects that are root: objects of TypeRegular type
	// with user data that are not system-specific. In addition to such objects, the
	// system may contain service objects that do not fall under this property
	// (like split leaves, tombstones, storage groups, etc.).
	FilterRoot = systemFilterPrefix + "ROOT"
	// FilterPhysical filters indivisible objects that are intended to be stored
	// on the physical devices of the system. In addition to such objects, the
	// system may contain so-called "virtual" objects that exist in the system in
	// disassembled form (like "huge" user object sliced into smaller ones).
	FilterPhysical = systemFilterPrefix + "PHY"
)

// ReadFromV2 reads SearchFilter from the apiobject.SearchRequest_Body_Filter
// message. Returns an error if the message is malformed according to the NeoFS
// API V2 protocol. The message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [SearchFilter.WriteToV2].
func (f *SearchFilter) ReadFromV2(m *apiobject.SearchRequest_Body_Filter) error {
	if m.MatchType < 0 {
		return errors.New("negative op")
	} else if m.Key == "" {
		return errors.New("missing key")
	}
	return nil
}

// WriteToV2 writes SearchFilter to the apiobject.SearchRequest_Body_Filter
// message of the NeoFS API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [SearchFilter.ReadFromV2].
func (f SearchFilter) WriteToV2(m *apiobject.SearchRequest_Body_Filter) {
	m.MatchType = apiobject.MatchType(f.op)
	m.Key = f.key
	m.Value = f.value
}

// Key returns key to the object property.
func (f SearchFilter) Key() string {
	return f.key
}

// Value returns filtered property value.
func (f SearchFilter) Value() string {
	return f.value
}

// Operation returns operator to match the property.
func (f SearchFilter) Operation() FilterOp {
	return f.op
}

// IsNonAttribute checks if SearchFilter is non-attribute: such filter is
// related to the particular property of the object instead of its attribute.
func (f SearchFilter) IsNonAttribute() bool {
	return strings.HasPrefix(f.key, systemFilterPrefix)
}

// NewSearchFilter constructs new object search filter instance. Additional
// helper constructors are also available to ease encoding.
func NewSearchFilter(key string, op FilterOp, value string) SearchFilter {
	return SearchFilter{
		key:   key,
		value: value,
		op:    op,
	}
}

// FilterRootObjects returns search filter selecting only root user objects (see
// [FilterRoot] for details).
func FilterRootObjects() SearchFilter {
	return NewSearchFilter(FilterRoot, 0, "")
}

// FilterPhysicalObjects returns search filter selecting physically stored
// objects only (see [FilterPhysical] for details).
func FilterPhysicalObjects() SearchFilter {
	return NewSearchFilter(FilterPhysical, 0, "")
}

// FilterOwnerIs returns search filter selecting objects owned by given user
// only. Relates to [Header.OwnerID].
func FilterOwnerIs(usr user.ID) SearchFilter {
	return NewSearchFilter(FilterOwnerID, FilterOpEQ, usr.EncodeToString())
}

// FilterParentIs returns search filter selecting only child objects for the
// given one. Relates to [Header.ParentID].
func FilterParentIs(id oid.ID) SearchFilter {
	return NewSearchFilter(FilterParentID, FilterOpEQ, id.EncodeToString())
}

// FilterFirstSplitObjectIs returns search filter selecting split-chain elements
// with the specified first one. Relates to [Header.FirstSplitObject] and
// [SplitInfo.FirstPart].
func FilterFirstSplitObjectIs(id oid.ID) SearchFilter {
	return NewSearchFilter(FilterFirstSplitObject, FilterOpEQ, id.EncodeToString())
}

// FilterTypeIs returns search filter selecting objects of certain type. Relates
// to [Header.Type].
func FilterTypeIs(typ Type) SearchFilter {
	return NewSearchFilter(FilterType, FilterOpEQ, typ.EncodeToString())
}

// FilterByCreationEpoch returns search filter selecting objects by creation
// time in NeoFS epochs. Relates to [Header.CreationEpoch]. Use
// [FilterByCreationTime] to specify Unix time format.
func FilterByCreationEpoch(op FilterOp, val uint64) SearchFilter {
	return NewSearchFilter(FilterCreationEpoch, op, strconv.FormatUint(val, 10))
}

// FilterByPayloadSize returns search filter selecting objects by payload size.
// Relates to [Header.PayloadSize].
func FilterByPayloadSize(op FilterOp, val uint64) SearchFilter {
	return NewSearchFilter(FilterPayloadSize, op, strconv.FormatUint(val, 10))
}

// FilterByName returns filter selecting objects by their human-readable names
// set as 'Name' attribute (see [SetName]).
func FilterByName(op FilterOp, name string) SearchFilter {
	return NewSearchFilter(attributeName, op, name)
}

// FilterByFileName returns filter selecting objects by file names associated
// with them through 'FileName' attribute (see [SetFileName]).
func FilterByFileName(op FilterOp, name string) SearchFilter {
	return NewSearchFilter(attributeFileName, op, name)
}

// FilterByFilePath returns filter selecting objects by filesystem paths
// associated with them through 'FilePath' attribute (see [SetFilePath]).
func FilterByFilePath(op FilterOp, name string) SearchFilter {
	return NewSearchFilter(attributeFilePath, op, name)
}

// FilterByCreationTime returns filter selecting objects by their creation time
// in Unix Timestamp format set as 'Timestamp' attribute (see
// [SetCreationTime]). Use [FilterByCreationEpoch] to specify NeoFS time format.
func FilterByCreationTime(op FilterOp, t time.Time) SearchFilter {
	return NewSearchFilter(attributeTimestamp, op, strconv.FormatInt(t.Unix(), 10))
}

// FilterByContentType returns filter selecting objects by content type of their
// payload set as 'Content-Type' attribute (see [SetContentType]).
func FilterByContentType(op FilterOp, name string) SearchFilter {
	return NewSearchFilter(attributeContentType, op, name)
}
