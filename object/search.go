package object

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
)

// SearchMatchType indicates match operation on specified header.
type SearchMatchType uint32

// MatchUnknown is an SearchMatchType value used to mark operator as undefined.
// Deprecated: use MatchUnspecified instead.
const MatchUnknown = MatchUnspecified

const (
	MatchUnspecified SearchMatchType = iota
	MatchStringEqual
	MatchStringNotEqual
	MatchNotPresent
	MatchCommonPrefix
	MatchNumGT
	MatchNumGE
	MatchNumLT
	MatchNumLE
)

// ToV2 converts [SearchMatchType] to v2 [v2object.MatchType] enum value.
// Deprecated: cast instead.
func (m SearchMatchType) ToV2() v2object.MatchType {
	return v2object.MatchType(m)
}

// SearchMatchFromV2 converts v2 [v2object.MatchType] to [SearchMatchType] enum value.
// Deprecated: cast instead.
func SearchMatchFromV2(t v2object.MatchType) SearchMatchType {
	return SearchMatchType(t)
}

// EncodeToString returns string representation of [SearchMatchType].
//
// String mapping:
//   - [MatchStringEqual]: STRING_EQUAL;
//   - [MatchStringNotEqual]: STRING_NOT_EQUAL;
//   - [MatchNotPresent]: NOT_PRESENT;
//   - [MatchCommonPrefix]: COMMON_PREFIX;
//   - [MatchNumGT], default: NUM_GT;
//   - [MatchNumGE], default: NUM_GE;
//   - [MatchNumLT], default: NUM_LT;
//   - [MatchNumLE], default: NUM_LE;
//   - [MatchUnknown]: MATCH_TYPE_UNSPECIFIED.
//
// All other values are base-10 integers.
func (m SearchMatchType) EncodeToString() string {
	return v2object.MatchType(m).String()
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as [EncodeToString]. String MUST NOT
// be used to encode ID into NeoFS protocol string.
func (m SearchMatchType) String() string {
	return m.EncodeToString()
}

// DecodeString parses [SearchMatchType] from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
func (m *SearchMatchType) DecodeString(s string) bool {
	var g v2object.MatchType

	ok := g.FromString(s)

	if ok {
		*m = SearchMatchType(g)
	}

	return ok
}

type stringEncoder interface {
	EncodeToString() string
}

// SearchFilter describes a single filter record.
type SearchFilter struct {
	header string
	value  stringEncoder
	op     SearchMatchType
}

type staticStringer string

// SearchFilters is type to describe a group of filters.
type SearchFilters []SearchFilter

// Various header filters.
const (
	FilterVersion                = v2object.FilterHeaderVersion
	FilterID                     = v2object.FilterHeaderObjectID
	FilterContainerID            = v2object.FilterHeaderContainerID
	FilterOwnerID                = v2object.FilterHeaderOwnerID
	FilterPayloadChecksum        = v2object.FilterHeaderPayloadHash
	FilterType                   = v2object.FilterHeaderObjectType
	FilterPayloadHomomorphicHash = v2object.FilterHeaderHomomorphicHash
	FilterParentID               = v2object.FilterHeaderParent
	FilterSplitID                = v2object.FilterHeaderSplitID
	FilterFirstSplitObject       = v2object.ReservedFilterPrefix + "split.first"
	FilterCreationEpoch          = v2object.FilterHeaderCreationEpoch
	FilterPayloadSize            = v2object.FilterHeaderPayloadLength
)

// Various filters to match certain object properties.
const (
	// FilterRoot filters objects that are root: objects of TypeRegular type
	// with user data that are not system-specific. In addition to such objects, the
	// system may contain service objects that do not fall under this property
	// (like split leaves, tombstones, storage groups, etc.).
	FilterRoot = v2object.FilterPropertyRoot
	// FilterPhysical filters indivisible objects that are intended to be stored
	// on the physical devices of the system. In addition to such objects, the
	// system may contain so-called "virtual" objects that exist in the system in
	// disassembled form (like "huge" user object sliced into smaller ones).
	FilterPhysical = v2object.FilterPropertyPhy
)

func (s staticStringer) EncodeToString() string {
	return string(s)
}

// Header returns filter header value.
func (f SearchFilter) Header() string {
	return f.header
}

// Value returns filter value.
func (f SearchFilter) Value() string {
	return f.value.EncodeToString()
}

// Operation returns filter operation value.
func (f SearchFilter) Operation() SearchMatchType {
	return f.op
}

// IsNonAttribute checks if SearchFilter is non-attribute: such filter is
// related to the particular property of the object instead of its attribute.
func (f SearchFilter) IsNonAttribute() bool {
	return strings.HasPrefix(f.header, v2object.ReservedFilterPrefix)
}

// NewSearchFilters constructs empty filter group.
func NewSearchFilters() SearchFilters {
	return SearchFilters{}
}

// NewSearchFiltersFromV2 converts slice of [v2object.SearchFilter] to [SearchFilters].
func NewSearchFiltersFromV2(v2 []v2object.SearchFilter) SearchFilters {
	filters := make(SearchFilters, 0, len(v2))

	for i := range v2 {
		filters.AddFilter(
			v2[i].GetKey(),
			v2[i].GetValue(),
			SearchMatchType(v2[i].GetMatchType()),
		)
	}

	return filters
}

func (f *SearchFilters) addFilter(op SearchMatchType, key string, val stringEncoder) {
	if *f == nil {
		*f = make(SearchFilters, 0, 1)
	}

	*f = append(*f, SearchFilter{
		header: key,
		value:  val,
		op:     op,
	})
}

// AddFilter adds a filter to group by simple plain parameters.
//
// If op is numeric (like [MatchNumGT]), value must be a base-10 integer.
func (f *SearchFilters) AddFilter(key, value string, op SearchMatchType) {
	f.addFilter(op, key, staticStringer(value))
}

// addFlagFilters adds filters that works like flags: they don't need to have
// specific match type or value. They processed by NeoFS nodes by the fact
// of presence in search query. E.g.: FilterRoot, FilterPhysical.
func (f *SearchFilters) addFlagFilter(key string) {
	f.addFilter(MatchUnspecified, key, staticStringer(""))
}

// AddObjectVersionFilter adds a filter by version.
//
// The op must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddObjectVersionFilter(op SearchMatchType, v version.Version) {
	f.addFilter(op, FilterVersion, staticStringer(version.EncodeToString(v)))
}

// AddObjectContainerIDFilter adds a filter by container id.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddObjectContainerIDFilter(m SearchMatchType, id cid.ID) {
	f.addFilter(m, FilterContainerID, id)
}

// AddObjectOwnerIDFilter adds a filter by object owner id.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddObjectOwnerIDFilter(m SearchMatchType, id user.ID) {
	f.addFilter(m, FilterOwnerID, id)
}

// ToV2 converts [SearchFilters] to [v2object.SearchFilter] slice.
func (f SearchFilters) ToV2() []v2object.SearchFilter {
	result := make([]v2object.SearchFilter, len(f))

	for i := range f {
		result[i].SetKey(f[i].header)
		result[i].SetValue(f[i].value.EncodeToString())
		result[i].SetMatchType(v2object.MatchType(f[i].op))
	}

	return result
}

func (f *SearchFilters) addRootFilter() {
	f.addFlagFilter(FilterRoot)
}

// AddRootFilter adds filter by objects that have been created by a user explicitly.
func (f *SearchFilters) AddRootFilter() {
	f.addRootFilter()
}

func (f *SearchFilters) addPhyFilter() {
	f.addFlagFilter(FilterPhysical)
}

// AddPhyFilter adds filter by objects that are physically stored in the system.
func (f *SearchFilters) AddPhyFilter() {
	f.addPhyFilter()
}

// AddParentIDFilter adds filter by parent identifier.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddParentIDFilter(m SearchMatchType, id oid.ID) {
	f.addFilter(m, FilterParentID, id)
}

// AddObjectIDFilter adds filter by object identifier.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddObjectIDFilter(m SearchMatchType, id oid.ID) {
	f.addFilter(m, FilterID, id)
}

// AddSplitIDFilter adds filter by split ID.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddSplitIDFilter(m SearchMatchType, id SplitID) {
	f.addFilter(m, FilterSplitID, staticStringer(id.String()))
}

// AddFirstSplitObjectFilter adds filter by first object ID.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddFirstSplitObjectFilter(m SearchMatchType, id oid.ID) {
	f.addFilter(m, FilterFirstSplitObject, staticStringer(id.String()))
}

// AddTypeFilter adds filter by object type.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddTypeFilter(m SearchMatchType, typ Type) {
	f.addFilter(m, FilterType, staticStringer(typ.EncodeToString()))
}

// MarshalJSON encodes [SearchFilters] to protobuf JSON format.
//
// See also [SearchFilters.UnmarshalJSON].
func (f SearchFilters) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.ToV2())
}

// UnmarshalJSON decodes [SearchFilters] from protobuf JSON format.
//
// See also [SearchFilters.MarshalJSON].
func (f *SearchFilters) UnmarshalJSON(data []byte) error {
	var fsV2 []v2object.SearchFilter

	if err := json.Unmarshal(data, &fsV2); err != nil {
		return err
	}

	*f = NewSearchFiltersFromV2(fsV2)

	return nil
}

// AddPayloadHashFilter adds filter by payload hash.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddPayloadHashFilter(m SearchMatchType, sum [sha256.Size]byte) {
	f.addFilter(m, FilterPayloadChecksum, staticStringer(hex.EncodeToString(sum[:])))
}

// AddHomomorphicHashFilter adds filter by homomorphic hash.
//
// The m must not be numeric (like [MatchNumGT]).
func (f *SearchFilters) AddHomomorphicHashFilter(m SearchMatchType, sum [tz.Size]byte) {
	f.addFilter(m, FilterPayloadHomomorphicHash, staticStringer(hex.EncodeToString(sum[:])))
}

// AddCreationEpochFilter adds filter by creation epoch.
func (f *SearchFilters) AddCreationEpochFilter(m SearchMatchType, epoch uint64) {
	f.addFilter(m, FilterCreationEpoch, staticStringer(strconv.FormatUint(epoch, 10)))
}

// AddPayloadSizeFilter adds filter by payload size.
func (f *SearchFilters) AddPayloadSizeFilter(m SearchMatchType, size uint64) {
	f.addFilter(m, FilterPayloadSize, staticStringer(strconv.FormatUint(size, 10)))
}
