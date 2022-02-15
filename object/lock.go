package object

import (
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Lock represents record with locked objects. It is compatible with
// NeoFS API V2 protocol.
type Lock v2object.Lock

// NumberOfMembers returns number of members in lock list.
func (x Lock) NumberOfMembers() int {
	return (*v2object.Lock)(&x).NumberOfMembers()
}

// ReadMembers reads list of locked members.
//
// Buffer length must not be less than NumberOfMembers.
func (x Lock) ReadMembers(buf []oid.ID) {
	var i int

	(*v2object.Lock)(&x).IterateMembers(func(id refs.ObjectID) {
		buf[i] = *oid.NewIDFromV2(&id) // need smth better
		i++
	})
}

// WriteMembers writes list of locked members.
func (x *Lock) WriteMembers(ids []oid.ID) {
	var members []refs.ObjectID

	if ids != nil {
		members = make([]refs.ObjectID, len(ids))

		for i := range ids {
			members[i] = *ids[i].ToV2() // need smth better
		}
	}

	(*v2object.Lock)(x).SetMembers(members)
}

// Marshal encodes the Lock into a NeoFS protocol binary format.
func (x Lock) Marshal() []byte {
	data, err := (*v2object.Lock)(&x).StableMarshal(nil)
	if err != nil {
		panic(err)
	}

	return data
}

// Unmarshal decodes the Lock from its NeoFS protocol binary representation.
func (x *Lock) Unmarshal(data []byte) error {
	return (*v2object.Lock)(x).Unmarshal(data)
}
