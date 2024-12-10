package object

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/tombstone"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Tombstone represents v2-compatible tombstone structure.
type Tombstone tombstone.Tombstone

// NewTombstoneFromV2 wraps v2 [tombstone.Tombstone] message to [Tombstone].
//
// Nil [tombstone.Tombstone] converts to nil.
// Deprecated: BUG: members' ID length is not checked. Use
// [Tombstone.ReadFromV2] instead.
func NewTombstoneFromV2(tV2 *tombstone.Tombstone) *Tombstone {
	return (*Tombstone)(tV2)
}

// NewTombstone creates and initializes blank [Tombstone].
//
// Defaults:
//   - exp: 0;
//   - splitID: nil;
//   - members: nil.
func NewTombstone() *Tombstone {
	return new(Tombstone)
}

// ReadFromV2 reads Tombstone from the [tombstone.Tombstone] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
func (t *Tombstone) ReadFromV2(m tombstone.Tombstone) error {
	var id oid.ID
	ms := m.GetMembers()
	for i := range ms {
		if err := id.ReadFromV2(ms[i]); err != nil {
			return fmt.Errorf("invalid member #%d: %w", i, err)
		}
	}
	if b := m.GetSplitID(); len(b) > 0 {
		var uid uuid.UUID
		if err := uid.UnmarshalBinary(b); err != nil {
			return fmt.Errorf("invalid split ID: %w", err)
		} else if v := uid.Version(); v != 4 {
			return fmt.Errorf("invalid split ID: wrong UUID version %d, expected 4", v)
		}
	}
	*t = Tombstone(m)
	return nil
}

// ToV2 converts [Tombstone] to v2 [tombstone.Tombstone] message.
//
// Nil [Tombstone] converts to nil.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (t Tombstone) ToV2() *tombstone.Tombstone {
	return (*tombstone.Tombstone)(&t)
}

// ExpirationEpoch returns the last NeoFS epoch number of the tombstone lifetime.
//
// See also [Tombstone.SetExpirationEpoch].
func (t Tombstone) ExpirationEpoch() uint64 {
	return (*tombstone.Tombstone)(&t).GetExpirationEpoch()
}

// SetExpirationEpoch sets the last NeoFS epoch number of the tombstone lifetime.
//
// See also [Tombstone.ExpirationEpoch].
func (t *Tombstone) SetExpirationEpoch(v uint64) {
	(*tombstone.Tombstone)(t).SetExpirationEpoch(v)
}

// SplitID returns identifier of object split hierarchy.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Tombstone.SetSplitID].
func (t Tombstone) SplitID() *SplitID {
	return NewSplitIDFromV2(
		(*tombstone.Tombstone)(&t).GetSplitID())
}

// SetSplitID sets identifier of object split hierarchy.
//
// See also [Tombstone.SplitID].
func (t *Tombstone) SetSplitID(v *SplitID) {
	(*tombstone.Tombstone)(t).SetSplitID(v.ToV2())
}

// Members returns list of objects to be deleted.
//
// See also [Tombstone.SetMembers].
func (t Tombstone) Members() []oid.ID {
	msV2 := (*tombstone.Tombstone)(&t).GetMembers()

	if msV2 == nil {
		return nil
	}

	res := make([]oid.ID, len(msV2))
	for i := range msV2 {
		if err := res[i].ReadFromV2(msV2[i]); err != nil {
			panic(fmt.Errorf("invalid member #%d: %w", i, err))
		}
	}

	return res
}

// SetMembers sets list of objects to be deleted.
//
// See also [Tombstone.Members].
func (t *Tombstone) SetMembers(v []oid.ID) {
	var ms []refs.ObjectID

	if v != nil {
		ms = (*tombstone.Tombstone)(t).
			GetMembers()

		if ln := len(v); cap(ms) >= ln {
			ms = ms[:0]
		} else {
			ms = make([]refs.ObjectID, 0, ln)
		}

		var idV2 refs.ObjectID

		for i := range v {
			v[i].WriteToV2(&idV2)
			ms = append(ms, idV2)
		}
	}

	(*tombstone.Tombstone)(t).SetMembers(ms)
}

// Marshal marshals [Tombstone] into a protobuf binary form.
//
// See also [Tombstone.Unmarshal].
func (t Tombstone) Marshal() []byte {
	return (*tombstone.Tombstone)(&t).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of [Tombstone].
//
// See also [Tombstone.Marshal].
func (t *Tombstone) Unmarshal(data []byte) error {
	var m tombstone.Tombstone
	if err := m.Unmarshal(data); err != nil {
		return err
	}
	return t.ReadFromV2(m)
}

// MarshalJSON encodes [Tombstone] to protobuf JSON format.
//
// See also [Tombstone.UnmarshalJSON].
func (t Tombstone) MarshalJSON() ([]byte, error) {
	return (*tombstone.Tombstone)(&t).MarshalJSON()
}

// UnmarshalJSON decodes [Tombstone] from protobuf JSON format.
//
// See also [Tombstone.MarshalJSON].
func (t *Tombstone) UnmarshalJSON(data []byte) error {
	var m tombstone.Tombstone
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}
	return t.ReadFromV2(m)
}
