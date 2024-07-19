package eacl

import (
	"strconv"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
)

// Filter describes a binary property of an access-controlled NeoFS resource
// according to meta information about it. The meta information is represented
// by a set of key-value attributes of various types.
//
// Filter should be created using one of the constructors.
type Filter struct {
	from    FilterHeaderType
	matcher Match
	key     string
	value   stringEncoder
}

// ConstructFilter constructs new Filter instance.
func ConstructFilter(h FilterHeaderType, k string, m Match, v string) Filter {
	return Filter{from: h, matcher: m, key: k, value: staticStringer(v)}
}

// NewObjectPropertyFilter constructs new Filter for the object property.
func NewObjectPropertyFilter(k string, m Match, v string) Filter {
	return ConstructFilter(HeaderFromObject, k, m, v)
}

// NewRequestHeaderFilter constructs new Filter for the request X-header.
func NewRequestHeaderFilter(k string, m Match, v string) Filter {
	return ConstructFilter(HeaderFromRequest, k, m, v)
}

// NewCustomServiceFilter constructs new Filter for the custom app-level
// property.
func NewCustomServiceFilter(k string, m Match, v string) Filter {
	return ConstructFilter(HeaderFromService, k, m, v)
}

type staticStringer string

type u64Stringer uint64

// Various keys to object filters.
const (
	FilterObjectVersion                    = "$Object:version"
	FilterObjectID                         = "$Object:objectID"
	FilterObjectContainerID                = "$Object:containerID"
	FilterObjectOwnerID                    = "$Object:ownerID"
	FilterObjectCreationEpoch              = "$Object:creationEpoch"
	FilterObjectPayloadSize                = "$Object:payloadLength"
	FilterObjectPayloadChecksum            = "$Object:payloadHash"
	FilterObjectType                       = "$Object:objectType"
	FilterObjectPayloadHomomorphicChecksum = "$Object:homomorphicHash"
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

// Value returns value of the access-controlled resource's attribute to match.
func (f Filter) Value() string {
	return f.value.EncodeToString()
}

// Matcher returns operator to match the attribute.
func (f Filter) Matcher() Match {
	return f.matcher
}

// Key returns key to the access-controlled resource's attribute to match.
func (f Filter) Key() string {
	return f.key
}

// From returns type of access-controlled resource's attribute to match.
func (f Filter) From() FilterHeaderType {
	return f.from
}

// ToV2 converts Filter to v2 acl.EACLRecord.Filter message.
//
// Nil Filter converts to nil.
// Deprecated: do not use it.
func (f *Filter) ToV2() *v2acl.HeaderFilter {
	if f != nil {
		return f.toProtoMessage()
	}
	return nil
}

func (f Filter) toProtoMessage() *v2acl.HeaderFilter {
	filter := new(v2acl.HeaderFilter)
	filter.SetValue(f.value.EncodeToString())
	filter.SetKey(f.key)
	filter.SetMatchType(v2acl.MatchType(f.matcher))
	filter.SetHeaderType(v2acl.HeaderType(f.from))
	return filter
}

func (f *Filter) fromProtoMessage(m *v2acl.HeaderFilter) error {
	f.from = FilterHeaderType(m.GetHeaderType())
	f.matcher = Match(m.GetMatchType())
	f.key = m.GetKey()
	f.value = staticStringer(m.GetValue())
	return nil
}

// NewFilter creates, initializes and returns blank Filter instance.
//
// Defaults:
//   - header type: HeaderTypeUnspecified;
//   - matcher: MatchUnspecified;
//   - key: "";
//   - value: "".
//
// Deprecated: use [ConstructFilter] instead.
func NewFilter() *Filter {
	f := ConstructFilter(0, "", 0, "")
	return &f
}

// NewFilterFromV2 converts v2 acl.EACLRecord.Filter message to Filter.
// Deprecated: do not use it.
func NewFilterFromV2(filter *v2acl.HeaderFilter) *Filter {
	f := new(Filter)

	if filter == nil {
		return f
	}

	_ = f.fromProtoMessage(filter)
	return f
}

// Marshal marshals Filter into a protobuf binary form.
func (f *Filter) Marshal() []byte {
	return f.toProtoMessage().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Filter.
func (f *Filter) Unmarshal(data []byte) error {
	m := new(v2acl.HeaderFilter)
	if err := m.Unmarshal(data); err != nil {
		return err
	}
	return f.fromProtoMessage(m)
}

// MarshalJSON encodes Filter to protobuf JSON format.
func (f *Filter) MarshalJSON() ([]byte, error) {
	return f.toProtoMessage().MarshalJSON()
}

// UnmarshalJSON decodes Filter from protobuf JSON format.
func (f *Filter) UnmarshalJSON(data []byte) error {
	m := new(v2acl.HeaderFilter)
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}
	return f.fromProtoMessage(m)
}
