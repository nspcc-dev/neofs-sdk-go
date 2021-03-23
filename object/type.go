package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
)

// Type is an object type identifier.
//
// Type is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/object.Type
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Type(object.Type{}) // not recommended
type Type object.Type

const (
	// TypeRegular is a default object type
	// that is a container for raw object payload
	// in NeoFS.
	TypeRegular Type = iota
	// TypeTombstone is a tombstone object type.
	TypeTombstone
	// TypeStorageGroup is a storage group object type.
	TypeStorageGroup
	// TypeLock is a lock object type.
	TypeLock
)

// ReadFromV2 reads Type from the object.Type message.
//
// See also WriteToV2.
func (t *Type) ReadFromV2(m object.Type) {
	*t = Type(m)
}

// WriteToV2 writes Type to the object.Type message.
// The message must not be nil.
//
// See also ReadFromV2.
func (t Type) WriteToV2(m *object.Type) {
	*m = object.Type(t)
}

// String implements fmt.Stringer interface method.
func (t Type) String() string {
	var v2 object.Type
	t.WriteToV2(&v2)

	return v2.String()
}

// Parse is a reverse action to String().
func (t *Type) Parse(s string) bool {
	var g object.Type

	ok := g.FromString(s)

	if ok {
		var x Type
		x.ReadFromV2(g)

		*t = x
	}

	return ok
}
