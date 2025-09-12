package eacl

import (
	"fmt"
	"strconv"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
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

func uint64Value(e uint64) string { return strconv.FormatUint(e, 10) }

// ConstructFilter constructs new Filter instance.
func ConstructFilter(h FilterHeaderType, k string, m Match, v string) Filter {
	return Filter{from: h, matcher: m, key: k, value: staticStringer(v)}
}

// NewObjectPropertyFilter constructs new Filter for the object property.
func NewObjectPropertyFilter(k string, m Match, v string) Filter {
	return ConstructFilter(HeaderFromObject, k, m, v)
}

// NewFilterObjectWithID constructs Filter that limits the access rule to the
// referenced object only.
func NewFilterObjectWithID(obj oid.ID) Filter {
	return NewObjectPropertyFilter(FilterObjectID, MatchStringEqual, obj.EncodeToString())
}

// NewFilterObjectsFromContainer constructs Filter that limits the access rule to
// objects from the referenced container only.
func NewFilterObjectsFromContainer(cnr cid.ID) Filter {
	return NewObjectPropertyFilter(FilterObjectContainerID, MatchStringEqual, cnr.EncodeToString())
}

// NewFilterObjectOwnerEquals constructs Filter that limits the access rule to
// objects owner by the given user only.
func NewFilterObjectOwnerEquals(usr user.ID) Filter {
	return NewObjectPropertyFilter(FilterObjectOwnerID, MatchStringEqual, usr.EncodeToString())
}

// NewFilterObjectCreationEpochIs constructs Filter that limits the access rule to
// objects with matching creation epoch only.
func NewFilterObjectCreationEpochIs(m Match, e uint64) Filter {
	return NewObjectPropertyFilter(FilterObjectCreationEpoch, m, uint64Value(e))
}

// NewFilterObjectPayloadSizeIs constructs Filter that limits the access rule to
// objects with matching payload size only.
func NewFilterObjectPayloadSizeIs(m Match, e uint64) Filter {
	return NewObjectPropertyFilter(FilterObjectPayloadSize, m, uint64Value(e))
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

func (f Filter) protoMessage() *protoacl.EACLRecord_Filter {
	return &protoacl.EACLRecord_Filter{
		HeaderType: protoacl.HeaderType(f.from),
		MatchType:  protoacl.MatchType(f.matcher),
		Key:        f.key,
		Value:      f.value.EncodeToString(),
	}
}

func (f *Filter) fromProtoMessage(m *protoacl.EACLRecord_Filter) error {
	if m.HeaderType < 0 {
		return fmt.Errorf("negative header type %d", m.HeaderType)
	}
	if m.MatchType < 0 {
		return fmt.Errorf("negative match type %d", m.MatchType)
	}
	f.from = FilterHeaderType(m.HeaderType)
	f.matcher = Match(m.MatchType)
	f.key = m.Key
	f.value = staticStringer(m.Value)
	return nil
}

// Marshal marshals Filter into a protobuf binary form.
func (f Filter) Marshal() []byte {
	return neofsproto.MarshalMessage(f.protoMessage())
}

// Unmarshal unmarshals protobuf binary representation of Filter.
func (f *Filter) Unmarshal(data []byte) error {
	m := new(protoacl.EACLRecord_Filter)
	if err := neofsproto.UnmarshalMessage(data, m); err != nil {
		return err
	}
	return f.fromProtoMessage(m)
}

// MarshalJSON encodes Filter to protobuf JSON format.
func (f Filter) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalMessageJSON(f.protoMessage())
}

// UnmarshalJSON decodes Filter from protobuf JSON format.
func (f *Filter) UnmarshalJSON(data []byte) error {
	m := new(protoacl.EACLRecord_Filter)
	if err := neofsproto.UnmarshalMessageJSON(data, m); err != nil {
		return err
	}
	return f.fromProtoMessage(m)
}
