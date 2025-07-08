package object

import (
	"fmt"

	"github.com/google/uuid"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	prototombstone "github.com/nspcc-dev/neofs-sdk-go/proto/tombstone"
)

// Tombstone represents object tombstone structure.
//
// DEPRECATED: use [Object.AssociateDeleted] instead, deleting exactly one object
// per Tombstone.
type Tombstone struct {
	exp     uint64
	splitID []byte
	members []oid.ID
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

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// t from it.
//
// See also [Tombstone.ProtoMessage].
func (t *Tombstone) FromProtoMessage(m *prototombstone.Tombstone) error {
	if m.Members != nil {
		t.members = make([]oid.ID, len(m.Members))
		for i := range m.Members {
			if m.Members[i] == nil {
				return fmt.Errorf("nil member #%d", i)
			}
			if err := t.members[i].FromProtoMessage(m.Members[i]); err != nil {
				return fmt.Errorf("invalid member #%d: %w", i, err)
			}
		}
	} else {
		t.members = nil
	}
	if t.splitID = m.SplitId; len(m.SplitId) > 0 {
		var uid uuid.UUID
		if err := uid.UnmarshalBinary(m.SplitId); err != nil {
			return fmt.Errorf("invalid split ID: %w", err)
		} else if v := uid.Version(); v != 4 {
			return fmt.Errorf("invalid split ID: wrong UUID version %d, expected 4", v)
		}
	}
	t.exp = m.ExpirationEpoch //nolint:staticcheck // must be supported still
	return nil
}

// ProtoMessage converts t into message to transmit using the NeoFS API
// protocol.
//
// See also [Tombstone.FromProtoMessage].
func (t Tombstone) ProtoMessage() *prototombstone.Tombstone {
	m := &prototombstone.Tombstone{
		ExpirationEpoch: t.exp,
		SplitId:         t.splitID,
	}
	if t.members != nil {
		m.Members = make([]*refs.ObjectID, len(t.members))
		for i := range t.members {
			m.Members[i] = t.members[i].ProtoMessage()
		}
	}
	return m
}

// ExpirationEpoch returns the last NeoFS epoch number of the tombstone lifetime.
//
// See also [Tombstone.SetExpirationEpoch].
func (t Tombstone) ExpirationEpoch() uint64 {
	return t.exp
}

// SetExpirationEpoch sets the last NeoFS epoch number of the tombstone lifetime.
//
// See also [Tombstone.ExpirationEpoch].
func (t *Tombstone) SetExpirationEpoch(v uint64) {
	t.exp = v
}

// SplitID returns identifier of object split hierarchy.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Tombstone.SetSplitID].
func (t Tombstone) SplitID() *SplitID {
	return NewSplitIDFromV2(t.splitID)
}

// SetSplitID sets identifier of object split hierarchy.
//
// See also [Tombstone.SplitID].
func (t *Tombstone) SetSplitID(v *SplitID) {
	t.splitID = v.ToV2()
}

// Members returns list of objects to be deleted.
//
// See also [Tombstone.SetMembers].
func (t Tombstone) Members() []oid.ID {
	return t.members
}

// SetMembers sets list of objects to be deleted.
//
// See also [Tombstone.Members].
func (t *Tombstone) SetMembers(v []oid.ID) {
	t.members = v
}

// Marshal marshals [Tombstone] into a protobuf binary form.
//
// See also [Tombstone.Unmarshal].
func (t Tombstone) Marshal() []byte {
	return neofsproto.Marshal(t)
}

// Unmarshal unmarshals protobuf binary representation of [Tombstone].
//
// See also [Tombstone.Marshal].
func (t *Tombstone) Unmarshal(data []byte) error {
	return neofsproto.Unmarshal(data, t)
}

// MarshalJSON encodes [Tombstone] to protobuf JSON format.
//
// See also [Tombstone.UnmarshalJSON].
func (t Tombstone) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(t)
}

// UnmarshalJSON decodes [Tombstone] from protobuf JSON format.
//
// See also [Tombstone.MarshalJSON].
func (t *Tombstone) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSON(data, t)
}
