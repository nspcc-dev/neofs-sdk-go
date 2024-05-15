package object

import (
	"fmt"
	"strconv"
)

// Type defines the payload format of the object.
type Type uint16

// Supported Type values.
const (
	TypeRegular      Type = iota // uninterpretable plain data
	TypeTombstone                // [Tombstone] carrier
	TypeStorageGroup             // storage group carrier
	TypeLock                     // [Lock] carrier
	TypeLink                     // [SplitChain] carrier
)

// EncodeToString encodes Type into NeoFS API V2 protocol string.
//
// See also [Type.DecodeString].
func (t Type) EncodeToString() string {
	switch t {
	default:
		return strconv.FormatUint(uint64(t), 10)
	case TypeRegular:
		return "REGULAR"
	case TypeTombstone:
		return "TOMBSTONE"
	case TypeStorageGroup:
		return "STORAGE_GROUP"
	case TypeLock:
		return "LOCK"
	case TypeLink:
		return "LINK"
	}
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as [Type.EncodeToString]. String
// MUST NOT be used to encode Type into NeoFS protocol string.
func (t Type) String() string {
	switch t {
	default:
		return fmt.Sprintf("UNKNOWN#%d", t)
	case TypeRegular:
		return "REGULAR"
	case TypeTombstone:
		return "TOMBSTONE"
	case TypeStorageGroup:
		return "STORAGE_GROUP"
	case TypeLock:
		return "LOCK"
	case TypeLink:
		return "LINK"
	}
}

// DecodeString decodes string into Type according to NeoFS API protocol.
// Returns an error if s is malformed.
//
// See also [Type.EncodeToString].
func (t *Type) DecodeString(s string) error {
	switch s {
	default:
		n, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return fmt.Errorf("decode numeric value: %w", err)
		}
		*t = Type(n)
	case "REGULAR":
		*t = TypeRegular
	case "TOMBSTONE":
		*t = TypeTombstone
	case "STORAGE_GROUP":
		*t = TypeStorageGroup
	case "LOCK":
		*t = TypeLock
	case "LINK":
		*t = TypeLink
	}
	return nil
}
