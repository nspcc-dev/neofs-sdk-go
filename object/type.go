package object

import (
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

// EncodeToString returns string representation of [Type].
//
// String mapping:
//   - [TypeTombstone]: TOMBSTONE;
//   - [TypeStorageGroup]: STORAGE_GROUP;
//   - [TypeLock]: LOCK;
//   - [TypeRegular], default: REGULAR.
//   - [TypeLink], default: LINK.
func (t Type) EncodeToString() string {
	return object.Type(t).String()
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as [Type.EncodeToString]. String MUST NOT
// be used to encode ID into NeoFS protocol string.
func (t Type) String() string {
	return t.EncodeToString()
}

// DecodeString parses [Type] from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
func (t *Type) DecodeString(s string) bool {
	var g object.Type

	ok := g.FromString(s)

	if ok {
		*t = Type(g)
	}

	return ok
}
