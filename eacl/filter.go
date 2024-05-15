package eacl

import (
	"errors"

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
)

// Filter describes a binary property of an access-controlled NeoFS resource
// according to meta information about it. The meta information is represented
// by a set of key-value attributes of various types.
type Filter struct {
	attrType AttributeType
	matcher  Match
	key      string
	value    string
}

// Various keys to object filters.
const (
	objectFilterPrefix                     = "$Object:"
	FilterObjectVersion                    = objectFilterPrefix + "version"
	FilterObjectID                         = objectFilterPrefix + "objectID"
	FilterObjectContainerID                = objectFilterPrefix + "containerID"
	FilterObjectOwnerID                    = objectFilterPrefix + "ownerID"
	FilterObjectCreationEpoch              = objectFilterPrefix + "creationEpoch"
	FilterObjectPayloadSize                = objectFilterPrefix + "payloadLength"
	FilterObjectPayloadChecksum            = objectFilterPrefix + "payloadHash"
	FilterObjectType                       = objectFilterPrefix + "objectType"
	FilterObjectPayloadHomomorphicChecksum = objectFilterPrefix + "homomorphicHash"
)

// CopyTo writes deep copy of the [Filter] to dst.
func (f Filter) CopyTo(dst *Filter) {
	*dst = f
}

func isEmptyFilter(f Filter) bool {
	return f.attrType == 0 && f.matcher == 0 && f.key == "" && f.value == ""
}

func filterToAPI(f Filter) *apiacl.EACLRecord_Filter {
	if isEmptyFilter(f) {
		return nil
	}
	return &apiacl.EACLRecord_Filter{
		HeaderType: apiacl.HeaderType(f.attrType),
		MatchType:  apiacl.MatchType(f.matcher),
		Key:        f.key,
		Value:      f.value,
	}
}

func (f *Filter) readFromV2(m *apiacl.EACLRecord_Filter, checkFieldPresence bool) error {
	if checkFieldPresence && m.Key == "" {
		return errors.New("missing key")
	}

	f.attrType = AttributeType(m.HeaderType)
	f.matcher = Match(m.MatchType)
	f.key = m.Key
	f.value = m.Value

	return nil
}

// SetValue sets value of the access-controlled resource's attribute to match.
//
// See also [Filter.Value].
func (f *Filter) SetValue(value string) {
	f.value = value
}

// Value returns value of the access-controlled resource's attribute to match.
//
// See also [Filter.SetValue].
func (f Filter) Value() string {
	return f.value
}

// Matcher returns operator to match the attribute.
func (f Filter) Matcher() Match {
	return f.matcher
}

// SetMatcher sets operator to match the attribute.
//
// See also [Filter.Matcher].
func (f *Filter) SetMatcher(m Match) {
	f.matcher = m
}

// Key returns key to the access-controlled resource's attribute to match.
//
// See also [Filter.SetKey].
func (f Filter) Key() string {
	return f.key
}

// SetKey sets key to the access-controlled resource's attribute to match.
//
// See also [Filter.Key].
func (f *Filter) SetKey(key string) {
	f.key = key
}

// AttributeType returns type of access-controlled resource's attribute to
// match.
//
// See also [Filter.AttributeType].
func (f Filter) AttributeType() AttributeType {
	return f.attrType
}

// SetAttributeType sets type of access-controlled resource's attribute to match.
//
// See also [Filter.AttributeType].
func (f *Filter) SetAttributeType(v AttributeType) {
	f.attrType = v
}
