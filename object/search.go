package object

import (
	"fmt"
	"strconv"

	objectv2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// SearchMatchType indicates match operation for selectable object properties.
//
// SearchMatchType is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/object.MatchType
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = SearchMatchType(object.MatchType{}) // not recommended
type SearchMatchType uint32

const (
	// MatchUnknown is a default SearchMatchType.
	MatchUnknown SearchMatchType = iota
	// MatchStringEqual is a SearchMatchType for exact
	// string matching.
	MatchStringEqual
	// MatchStringNotEqual is a SearchMatchType for
	// string mismatching.
	MatchStringNotEqual
	// MatchNotPresent is a SearchMatchType for
	// missing object headers.
	MatchNotPresent
	// MatchCommonPrefix is a SearchMatchType for
	// matching object header prefixes.
	MatchCommonPrefix
)

// ReadFromV2 reads SearchMatchType from the object.MatchType message.
//
// See also WriteToV2.
func (s *SearchMatchType) ReadFromV2(m objectv2.MatchType) {
	switch m {
	case objectv2.MatchStringEqual:
		*s = MatchStringEqual
	case objectv2.MatchStringNotEqual:
		*s = MatchStringNotEqual
	case objectv2.MatchNotPresent:
		*s = MatchNotPresent
	case objectv2.MatchCommonPrefix:
		*s = MatchCommonPrefix
	default:
		*s = MatchUnknown
	}
}

// WriteToV2 writes SearchMatchType to the object.MatchType message.
// The message must not be nil.
//
// See also ReadFromV2.
func (s SearchMatchType) WriteToV2(m *objectv2.MatchType) {
	switch s {
	case MatchStringEqual:
		*m = objectv2.MatchStringEqual
	case MatchStringNotEqual:
		*m = objectv2.MatchStringNotEqual
	case MatchNotPresent:
		*m = objectv2.MatchNotPresent
	case MatchCommonPrefix:
		*m = objectv2.MatchCommonPrefix
	default:
		*m = objectv2.MatchUnknown
	}
}

// String implements fmt.Stringer interface method.
func (s SearchMatchType) String() string {
	var v2 objectv2.MatchType
	s.WriteToV2(&v2)

	return v2.String()
}

// Parse is a reverse action to String().
func (s *SearchMatchType) Parse(str string) bool {
	var g objectv2.MatchType

	ok := g.FromString(str)

	if ok {
		*s = (SearchMatchType)(g)
	}

	return ok
}

// SearchFilter groups information about
// certain search filter:
// 	* header key;
// 	* search filter type;
// 	* operand of the filter.
type SearchFilter struct {
	header filterKey
	value  fmt.Stringer
	op     SearchMatchType
}

// Header returns header that filter is
// applied to.
//
// Zero SearchFilter has empty header.
func (f SearchFilter) Header() string {
	return f.header.String()
}

// Value returns filter's operand.
//
// Calling that method on zero SearchFilter
// leads to panic.
func (f SearchFilter) Value() string {
	return f.value.String()
}

// Operation returns filter's match type.
//
// Zero SearchFilter has MatchUnknown search type.
func (f SearchFilter) Operation() SearchMatchType {
	return f.op
}

type staticStringer string

type filterKey struct {
	typ filterKeyType

	str string
}

// enumeration of reserved filter keys.
type filterKeyType int

// SearchFilters groups search filters.
type SearchFilters []SearchFilter

// ReadFromV2 reads SearchFilters from the []object.SearchFilter messages.
//
// See also WriteToV2.
func (f *SearchFilters) ReadFromV2(v2 []objectv2.SearchFilter) {
	*f = make(SearchFilters, 0, len(v2))

	var smt SearchMatchType

	for i := range v2 {
		smt.ReadFromV2(v2[i].GetMatchType())
		f.AddFilter(
			v2[i].GetKey(),
			v2[i].GetValue(),
			smt,
		)
	}
}

// WriteToV2 writes SearchFilters to the []object.SearchFilter messages.
// The message must not be nil.
//
// See also ReadFromV2.
func (f SearchFilters) WriteToV2(v2 *[]objectv2.SearchFilter) {
	*v2 = make([]objectv2.SearchFilter, len(f))
	var mtv2 objectv2.MatchType

	for i := range f {
		f[i].op.WriteToV2(&mtv2)

		(*v2)[i].SetKey(f[i].header.String())
		(*v2)[i].SetValue(f[i].value.String())
		(*v2)[i].SetMatchType(mtv2)
	}
}

const (
	_ filterKeyType = iota
	fKeyVersion
	fKeyObjectID
	fKeyContainerID
	fKeyOwnerID
	fKeyCreationEpoch
	fKeyPayloadLength
	fKeyPayloadHash
	fKeyType
	fKeyHomomorphicHash
	fKeyParent
	fKeySplitID
	fKeyPropRoot
	fKeyPropPhy
)

// String implements fmt.Stringer interface method.// String implements fmt.Stringer interface method.
func (k filterKey) String() string {
	switch k.typ {
	default:
		return k.str
	case fKeyVersion:
		return objectv2.FilterHeaderVersion
	case fKeyObjectID:
		return objectv2.FilterHeaderObjectID
	case fKeyContainerID:
		return objectv2.FilterHeaderContainerID
	case fKeyOwnerID:
		return objectv2.FilterHeaderOwnerID
	case fKeyCreationEpoch:
		return objectv2.FilterHeaderCreationEpoch
	case fKeyPayloadLength:
		return objectv2.FilterHeaderPayloadLength
	case fKeyPayloadHash:
		return objectv2.FilterHeaderPayloadHash
	case fKeyType:
		return objectv2.FilterHeaderObjectType
	case fKeyHomomorphicHash:
		return objectv2.FilterHeaderHomomorphicHash
	case fKeyParent:
		return objectv2.FilterHeaderParent
	case fKeySplitID:
		return objectv2.FilterHeaderSplitID
	case fKeyPropRoot:
		return objectv2.FilterPropertyRoot
	case fKeyPropPhy:
		return objectv2.FilterPropertyPhy
	}
}

// String implements fmt.Stringer interface method.
func (s staticStringer) String() string {
	return string(s)
}

// AddFilter appends SearchFilters with the provided search operation.
func (f *SearchFilters) AddFilter(header, value string, op SearchMatchType) {
	f.addFilter(op, 0, header, staticStringer(value))
}

// AddObjectVersionFilter appends SearchFilters with object version
// search filter.
func (f *SearchFilters) AddObjectVersionFilter(op SearchMatchType, v *version.Version) {
	f.addReservedFilter(op, fKeyVersion, v)
}

// AddObjectContainerIDFilter appends SearchFilters with container ID
// search filter.
func (f *SearchFilters) AddObjectContainerIDFilter(m SearchMatchType, id *cid.ID) {
	f.addReservedFilter(m, fKeyContainerID, id)
}

// AddObjectOwnerIDFilter appends SearchFilters with owner ID
// search filter.
func (f *SearchFilters) AddObjectOwnerIDFilter(m SearchMatchType, id *owner.ID) {
	f.addReservedFilter(m, fKeyOwnerID, id)
}

// AddNotificationEpochFilter appends SearchFilters with a filter
// that matches object notification epoch tick with provided value.
func (f *SearchFilters) AddNotificationEpochFilter(epoch uint64) {
	f.addFilter(MatchStringEqual, 0, objectv2.SysAttributeTickEpoch, staticStringer(strconv.FormatUint(epoch, 10)))
}

// AddRootFilter appends SearchFilters with ROOT flag.
func (f *SearchFilters) AddRootFilter() {
	f.addRootFilter()
}

// AddPhyFilter appends SearchFilters with PHY flag.
func (f *SearchFilters) AddPhyFilter() {
	f.addPhyFilter()
}

// AddParentIDFilter adds filter by parent identifier.
func (f *SearchFilters) AddParentIDFilter(m SearchMatchType, id *oid.ID) {
	f.addReservedFilter(m, fKeyParent, id)
}

// AddObjectIDFilter adds filter by object identifier.
func (f *SearchFilters) AddObjectIDFilter(m SearchMatchType, id *oid.ID) {
	f.addReservedFilter(m, fKeyObjectID, id)
}

func (f *SearchFilters) AddSplitIDFilter(m SearchMatchType, id *SplitID) {
	f.addReservedFilter(m, fKeySplitID, id)
}

// AddTypeFilter adds filter by object type.
func (f *SearchFilters) AddTypeFilter(m SearchMatchType, typ Type) {
	f.addReservedFilter(m, fKeyType, typ)
}

func (f *SearchFilters) addReservedFilter(op SearchMatchType, keyTyp filterKeyType, val fmt.Stringer) {
	f.addFilter(op, keyTyp, "", val)
}

// addFlagFilters adds filters that works like flags: they don't need to have
// specific match type or value. They processed by NeoFS nodes by the fact
// of presence in search query. E.g.: PHY, ROOT.
func (f *SearchFilters) addFlagFilter(keyTyp filterKeyType) {
	f.addFilter(MatchUnknown, keyTyp, "", staticStringer(""))
}

func (f *SearchFilters) addPhyFilter() {
	f.addFlagFilter(fKeyPropPhy)
}

func (f *SearchFilters) addRootFilter() {
	f.addFlagFilter(fKeyPropRoot)
}

func (f *SearchFilters) addFilter(op SearchMatchType, keyTyp filterKeyType, key string, val fmt.Stringer) {
	if *f == nil {
		*f = make(SearchFilters, 0, 1)
	}

	*f = append(*f, SearchFilter{
		header: filterKey{
			typ: keyTyp,
			str: key,
		},
		value: val,
		op:    op,
	})
}
