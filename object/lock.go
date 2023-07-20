package object

import (
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Lock represents record with locked objects. It is compatible with
// NeoFS API V2 protocol.
//
// Lock instance can be written to the [Object], see WriteLock/ReadLock.
type Lock v2object.Lock

// WriteLock writes [Lock] to the [Object], and sets its type to [TypeLock].
//
// See also ReadLock.
func (o *Object) WriteLock(l Lock) {
	o.SetType(TypeLock)
	o.SetPayload(l.Marshal())
}

// ReadLock reads [Lock] from the [Object]. The lock must not be nil.
// Returns an error describing incorrect format. Makes sense only
// if object has [TypeLock] type.
//
// See also [Object.WriteLock].
func (o *Object) ReadLock(l *Lock) error {
	return l.Unmarshal(o.Payload())
}

// NumberOfMembers returns number of members in lock list.
func (x Lock) NumberOfMembers() int {
	return (*v2object.Lock)(&x).NumberOfMembers()
}

// ReadMembers reads list of locked members.
//
// Buffer length must not be less than [Lock.NumberOfMembers].
func (x Lock) ReadMembers(buf []oid.ID) {
	var i int

	(*v2object.Lock)(&x).IterateMembers(func(idV2 refs.ObjectID) {
		_ = buf[i].ReadFromV2(idV2)
		i++
	})
}

// WriteMembers writes list of locked members.
//
// See also [Lock.ReadMembers].
func (x *Lock) WriteMembers(ids []oid.ID) {
	var members []refs.ObjectID

	if ids != nil {
		members = make([]refs.ObjectID, len(ids))

		for i := range ids {
			ids[i].WriteToV2(&members[i])
		}
	}

	(*v2object.Lock)(x).SetMembers(members)
}

// Marshal encodes the [Lock] into a NeoFS protocol binary format.
//
// See also [Lock.Unmarshal].
func (x Lock) Marshal() []byte {
	return (*v2object.Lock)(&x).StableMarshal(nil)
}

// Unmarshal decodes the [Lock] from its NeoFS protocol binary representation.
//
// See also [Lock.Marshal].
func (x *Lock) Unmarshal(data []byte) error {
	return (*v2object.Lock)(x).Unmarshal(data)
}
