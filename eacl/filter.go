package eacl

import (
	"strconv"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
)

// Filter defines check conditions if request header is matched or not. Matched
// header means that request should be processed according to ContainerEACL action.
//
// Filter is compatible with v2 acl.EACLRecord.Filter message.
type Filter struct {
	from    FilterHeaderType
	matcher Match
	key     string
	value   stringEncoder
}

type staticStringer string

type u64Stringer uint64

// Various keys to object filters.
const (
	FilterObjectVersion                    = v2acl.FilterObjectVersion
	FilterObjectID                         = v2acl.FilterObjectID
	FilterObjectContainerID                = v2acl.FilterObjectContainerID
	FilterObjectOwnerID                    = v2acl.FilterObjectOwnerID
	FilterObjectCreationEpoch              = v2acl.FilterObjectCreationEpoch
	FilterObjectPayloadSize                = v2acl.FilterObjectPayloadLength
	FilterObjectPayloadChecksum            = v2acl.FilterObjectPayloadHash
	FilterObjectType                       = v2acl.FilterObjectType
	FilterObjectPayloadHomomorphicChecksum = v2acl.FilterObjectHomomorphicHash
)

func (s staticStringer) EncodeToString() string {
	return string(s)
}

func (u u64Stringer) EncodeToString() string {
	return strconv.FormatUint(uint64(u), 10)
}

// CopyTo writes deep copy of the [Filter] to dst.
func (f Filter) CopyTo(dst *Filter) {
	dst.from = f.from
	dst.matcher = f.matcher
	dst.key = f.key
	dst.value = f.value
}

// Value returns filtered string value.
func (f Filter) Value() string {
	return f.value.EncodeToString()
}

// Matcher returns filter Match type.
func (f Filter) Matcher() Match {
	return f.matcher
}

// Key returns key to the filtered header.
func (f Filter) Key() string {
	return f.key
}

// From returns FilterHeaderType that defined which header will be filtered.
func (f Filter) From() FilterHeaderType {
	return f.from
}

// ToV2 converts Filter to v2 acl.EACLRecord.Filter message.
//
// Nil Filter converts to nil.
func (f *Filter) ToV2() *v2acl.HeaderFilter {
	if f == nil {
		return nil
	}

	filter := new(v2acl.HeaderFilter)
	filter.SetValue(f.value.EncodeToString())
	filter.SetKey(f.key)
	filter.SetMatchType(f.matcher.ToV2())
	filter.SetHeaderType(f.from.ToV2())

	return filter
}

// NewFilter creates, initializes and returns blank Filter instance.
//
// Defaults:
//   - header type: HeaderTypeUnspecified;
//   - matcher: MatchUnspecified;
//   - key: "";
//   - value: "".
func NewFilter() *Filter {
	return NewFilterFromV2(new(v2acl.HeaderFilter))
}

// NewFilterFromV2 converts v2 acl.EACLRecord.Filter message to Filter.
func NewFilterFromV2(filter *v2acl.HeaderFilter) *Filter {
	f := new(Filter)

	if filter == nil {
		return f
	}

	f.from = FilterHeaderTypeFromV2(filter.GetHeaderType())
	f.matcher = MatchFromV2(filter.GetMatchType())
	f.key = filter.GetKey()
	f.value = staticStringer(filter.GetValue())

	return f
}

// Marshal marshals Filter into a protobuf binary form.
func (f *Filter) Marshal() []byte {
	return f.ToV2().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Filter.
func (f *Filter) Unmarshal(data []byte) error {
	fV2 := new(v2acl.HeaderFilter)
	if err := fV2.Unmarshal(data); err != nil {
		return err
	}

	*f = *NewFilterFromV2(fV2)

	return nil
}

// MarshalJSON encodes Filter to protobuf JSON format.
func (f *Filter) MarshalJSON() ([]byte, error) {
	return f.ToV2().MarshalJSON()
}

// UnmarshalJSON decodes Filter from protobuf JSON format.
func (f *Filter) UnmarshalJSON(data []byte) error {
	fV2 := new(v2acl.HeaderFilter)
	if err := fV2.UnmarshalJSON(data); err != nil {
		return err
	}

	*f = *NewFilterFromV2(fV2)

	return nil
}

// equalFilters compares Filter with each other.
func equalFilters(f1, f2 Filter) bool {
	return f1.From() == f2.From() &&
		f1.Matcher() == f2.Matcher() &&
		f1.Key() == f2.Key() &&
		f1.Value() == f2.Value()
}
