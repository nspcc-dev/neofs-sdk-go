package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/tombstone"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Tombstone represents v2-compatible tombstone structure.
type Tombstone tombstone.Tombstone

// NewTombstoneFromV2 wraps v2 [tombstone.Tombstone] message to [Tombstone].
//
// Nil [tombstone.Tombstone] converts to nil.
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
	return NewTombstoneFromV2(new(tombstone.Tombstone))
}

// ToV2 converts [Tombstone] to v2 [tombstone.Tombstone] message.
//
// Nil [Tombstone] converts to nil.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (t *Tombstone) ToV2() *tombstone.Tombstone {
	return (*tombstone.Tombstone)(t)
}

// ExpirationEpoch returns the last NeoFS epoch number of the tombstone lifetime.
//
// See also [Tombstone.SetExpirationEpoch].
func (t *Tombstone) ExpirationEpoch() uint64 {
	return (*tombstone.Tombstone)(t).GetExpirationEpoch()
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
func (t *Tombstone) SplitID() *SplitID {
	return NewSplitIDFromV2(
		(*tombstone.Tombstone)(t).GetSplitID())
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
func (t *Tombstone) Members() []oid.ID {
	v2 := (*tombstone.Tombstone)(t)
	msV2 := v2.GetMembers()

	if msV2 == nil {
		return nil
	}

	var (
		ms = make([]oid.ID, len(msV2))
		id oid.ID
	)

	for i := range msV2 {
		_ = id.ReadFromV2(msV2[i])
		ms[i] = id
	}

	return ms
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
func (t *Tombstone) Marshal() []byte {
	return (*tombstone.Tombstone)(t).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of [Tombstone].
//
// See also [Tombstone.Marshal].
func (t *Tombstone) Unmarshal(data []byte) error {
	return (*tombstone.Tombstone)(t).Unmarshal(data)
}

// MarshalJSON encodes [Tombstone] to protobuf JSON format.
//
// See also [Tombstone.UnmarshalJSON].
func (t *Tombstone) MarshalJSON() ([]byte, error) {
	return (*tombstone.Tombstone)(t).MarshalJSON()
}

// UnmarshalJSON decodes [Tombstone] from protobuf JSON format.
//
// See also [Tombstone.MarshalJSON].
func (t *Tombstone) UnmarshalJSON(data []byte) error {
	return (*tombstone.Tombstone)(t).UnmarshalJSON(data)
}
