package object

import (
	"strconv"
)

// Type is an enumerator for possible object types.
type Type int32

const (
	TypeRegular Type = iota
	TypeTombstone
	TypeStorageGroup
	TypeLock
	TypeLink
)

const (
	typeStringRegular      = "REGULAR"
	typeStringTombstone    = "TOMBSTONE"
	typeStringStorageGroup = "STORAGE_GROUP"
	typeStringLock         = "LOCK"
	typeStringLink         = "LINK"
)

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
func (t Type) EncodeToString() string { return t.String() }

// String implements [fmt.Stringer] with the following string mapping:
//   - [TypeRegular]: REGULAR
//   - [TypeTombstone]: TOMBSTONE
//   - [TypeStorageGroup]: STORAGE_GROUP
//   - [TypeLock]: LOCK
//   - [TypeLink]: LINK
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format.
//
// String is reverse to [Type.DecodeString].
func (t Type) String() string {
	switch t {
	default:
		return strconv.FormatInt(int64(t), 10)
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

// DecodeString parses Type from a string representation. It is a reverse action
// to [Type.String].
//
// Returns true if s was parsed successfully.
func (t *Type) DecodeString(s string) bool {
	switch s {
	default:
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}
		*t = Type(n)
	case typeStringRegular:
		*t = TypeRegular
	case typeStringTombstone:
		*t = TypeTombstone
	case typeStringStorageGroup:
		*t = TypeStorageGroup
	case typeStringLock:
		*t = TypeLock
	case typeStringLink:
		*t = TypeLink
	}
	return true
}
