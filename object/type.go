package object

import (
	"strconv"

	"github.com/nspcc-dev/neofs-api-go/v2/object"
)

// Type is an enumerator for possible object types.
type Type object.Type

const (
	TypeRegular Type = iota
	TypeTombstone
	TypeStorageGroup
	TypeLock
	TypeLink
)

// ToV2 converts [Type] to v2 [object.Type].
// Deprecated: cast instead.
func (t Type) ToV2() object.Type {
	return object.Type(t)
}

// TypeFromV2 converts v2 [object.Type] to [Type].
// Deprecated: cast instead.
func TypeFromV2(t object.Type) Type {
	return Type(t)
}

const (
	typeStringRegular      = "REGULAR"
	typeStringTombstone    = "TOMBSTONE"
	typeStringStorageGroup = "STORAGE_GROUP"
	typeStringLock         = "LOCK"
	typeStringLink         = "LINK"
)

// TypeToString maps Type values to strings:
//   - [TypeRegular]: REGULAR
//   - [TypeTombstone]: TOMBSTONE
//   - [TypeStorageGroup]: STORAGE_GROUP
//   - [TypeLock]: LOCK
//   - [TypeLink]: LINK
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format. Use [Type.String] to
// get any human-readable text for printing.
func TypeToString(t Type) string {
	switch t {
	default:
		return strconv.FormatUint(uint64(t), 10)
	case TypeRegular:
		return typeStringRegular
	case TypeTombstone:
		return typeStringTombstone
	case TypeStorageGroup:
		return typeStringStorageGroup
	case TypeLock:
		return typeStringLock
	case TypeLink:
		return typeStringLink
	}
}

// TypeFromString maps strings to Type values in reverse to [TypeToString].
// Returns false if s is incorrect.
func TypeFromString(s string) (Type, bool) {
	switch s {
	default:
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			return Type(n), true
		}
		return 0, false
	case typeStringRegular:
		return TypeRegular, true
	case typeStringTombstone:
		return TypeTombstone, true
	case typeStringStorageGroup:
		return TypeStorageGroup, true
	case typeStringLock:
		return TypeLock, true
	case typeStringLink:
		return TypeLink, true
	}
}

// EncodeToString returns string representation of [Type].
//
// String mapping:
//   - [TypeTombstone]: TOMBSTONE;
//   - [TypeStorageGroup]: STORAGE_GROUP;
//   - [TypeLock]: LOCK;
//   - [TypeRegular], default: REGULAR.
//   - [TypeLink], default: LINK.
//
// Deprecated: use [TypeToString] instead.
func (t Type) EncodeToString() string { return TypeToString(t) }

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. Use [TypeToString] and [TypeFromString] for consistent mapping.
func (t Type) String() string {
	return TypeToString(t)
}

// DecodeString parses [Type] from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
// Deprecated: use [TypeFromString] instead.
func (t *Type) DecodeString(s string) bool {
	if v, ok := TypeFromString(s); ok {
		*t = v
		return true
	}
	return false
}
