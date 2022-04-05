package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/tombstone"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Tombstone represents v2-compatible tombstone structure.
//
// Tombstone is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/tombstone.Tombstone
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Tombstone(tombstone.Tombstone{}) // not recommended
type Tombstone tombstone.Tombstone

// ReadFromV2 reads Tombstone from the tombstone.Tombstone message.
//
// See also WriteToV2.
func (t *Tombstone) ReadFromV2(m tombstone.Tombstone) {
	*t = Tombstone(m)
}

// WriteToV2 writes Tombstone to the tombstone.Tombstone message.
// The message must not be nil.
//
// See also ReadFromV2.
func (t Tombstone) WriteToV2(m *tombstone.Tombstone) {
	*m = (tombstone.Tombstone)(t)
}

// ExpirationEpoch returns number of tombstone expiration epoch.
//
// Zero Tombstone has 0 expiration epoch.
//
// See also SetExpirationEpoch.
func (t Tombstone) ExpirationEpoch() uint64 {
	v2 := (tombstone.Tombstone)(t)
	return v2.GetExpirationEpoch()
}

// SetExpirationEpoch sets number of tombstone expiration epoch.
//
// See also ExpirationEpoch.
func (t *Tombstone) SetExpirationEpoch(v uint64) {
	(*tombstone.Tombstone)(t).SetExpirationEpoch(v)
}

// SplitID returns identifier of object split hierarchy.
//
// Zero Tombstone has zero SplitID.
//
// See also SetExpirationEpoch.
func (t Tombstone) SplitID() SplitID {
	v2 := (tombstone.Tombstone)(t)
	return NewSplitIDFromBytes(v2.GetSplitID())
}

// SetSplitID sets identifier of object split hierarchy.
//
// See also SplitID.
func (t *Tombstone) SetSplitID(v SplitID) {
	(*tombstone.Tombstone)(t).SetSplitID(v.ToBytes())
}

// Members returns list of objects to be deleted.
//
// Zero Tombstone has nil members.
//
// See also SetMembers.
func (t Tombstone) Members() []oid.ID {
	v2 := (tombstone.Tombstone)(t)
	msV2 := v2.GetMembers()

	if msV2 == nil {
		return nil
	}

	var (
		ms = make([]oid.ID, len(msV2))
		id oid.ID
	)

	for i := range msV2 {
		id.ReadFromV2(msV2[i])
		ms[i] = id
	}

	return ms
}

// SetMembers sets list of objects to be deleted.
//
// See also Members.
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

// Marshal marshals Tombstone into a protobuf binary form.
func (t Tombstone) Marshal() ([]byte, error) {
	v2 := (tombstone.Tombstone)(t)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Tombstone.
func (t *Tombstone) Unmarshal(data []byte) error {
	return (*tombstone.Tombstone)(t).Unmarshal(data)
}

// MarshalJSON encodes Tombstone to protobuf JSON format.
func (t Tombstone) MarshalJSON() ([]byte, error) {
	v2 := (tombstone.Tombstone)(t)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes Tombstone from protobuf JSON format.
func (t *Tombstone) UnmarshalJSON(data []byte) error {
	return (*tombstone.Tombstone)(t).UnmarshalJSON(data)
}
