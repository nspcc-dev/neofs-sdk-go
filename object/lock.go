package object

import (
	"fmt"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protolock "github.com/nspcc-dev/neofs-sdk-go/proto/lock"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
)

// Lock represents record with locked objects. It is compatible with
// NeoFS API V2 protocol.
//
// Lock instance can be written to the [Object], see WriteLock/ReadLock.
type Lock struct {
	members []oid.ID
}

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
func (o Object) ReadLock(l *Lock) error {
	return l.Unmarshal(o.Payload())
}

// NumberOfMembers returns number of members in lock list.
func (x Lock) NumberOfMembers() int {
	return len(x.members)
}

// ReadMembers reads list of locked members.
//
// Buffer length must not be less than [Lock.NumberOfMembers].
func (x Lock) ReadMembers(buf []oid.ID) {
	copy(buf, x.members)
}

// WriteMembers writes list of locked members.
//
// See also [Lock.ReadMembers].
func (x *Lock) WriteMembers(ids []oid.ID) {
	x.members = ids
}

// Marshal encodes the [Lock] into a NeoFS protocol binary format.
//
// See also [Lock.Unmarshal].
func (x Lock) Marshal() []byte {
	if len(x.members) == 0 {
		return nil
	}
	m := &protolock.Lock{
		Members: make([]*refs.ObjectID, len(x.members)),
	}
	for i := range x.members {
		m.Members[i] = x.members[i].ProtoMessage()
	}
	return neofsproto.MarshalMessage(m)
}

// Unmarshal decodes the [Lock] from its NeoFS protocol binary representation.
//
// See also [Lock.Marshal].
func (x *Lock) Unmarshal(data []byte) error {
	m := new(protolock.Lock)
	err := neofsproto.UnmarshalMessage(data, m)
	if err != nil {
		return err
	}

	if m.Members == nil {
		x.members = nil
		return nil
	}

	x.members = make([]oid.ID, len(m.Members))
	for i := range m.Members {
		if m.Members[i] == nil {
			return fmt.Errorf("nil member #%d", i)
		}
		if err = x.members[i].FromProtoMessage(m.Members[i]); err != nil {
			return fmt.Errorf("invalid member #%d: %w", i, err)
		}
	}

	return nil
}
