package netmap

import (
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
)

// Replica represents v2-compatible object replica descriptor.
type Replica netmap.Replica

// NewReplica creates and returns new Replica instance.
//
// Defaults:
//  - count: 0;
//  - selector: "".
func NewReplica() *Replica {
	return NewReplicaFromV2(new(netmap.Replica))
}

// NewReplicaFromV2 converts v2 Replica to Replica.
//
// Nil netmap.Replica converts to nil.
func NewReplicaFromV2(f *netmap.Replica) *Replica {
	return (*Replica)(f)
}

// ToV2 converts Replica to v2 Replica.
//
// Nil Replica converts to nil.
func (r *Replica) ToV2() *netmap.Replica {
	return (*netmap.Replica)(r)
}

// Count returns number of object replicas.
func (r *Replica) Count() uint32 {
	return (*netmap.Replica)(r).GetCount()
}

// SetCount sets number of object replicas.
func (r *Replica) SetCount(c uint32) {
	(*netmap.Replica)(r).SetCount(c)
}

// Selector returns name of selector bucket to put replicas.
func (r *Replica) Selector() string {
	return (*netmap.Replica)(r).GetSelector()
}

// SetSelector sets name of selector bucket to put replicas.
func (r *Replica) SetSelector(s string) {
	(*netmap.Replica)(r).SetSelector(s)
}

// Marshal marshals Replica into a protobuf binary form.
func (r *Replica) Marshal() ([]byte, error) {
	return (*netmap.Replica)(r).StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of Replica.
func (r *Replica) Unmarshal(data []byte) error {
	return (*netmap.Replica)(r).Unmarshal(data)
}

// MarshalJSON encodes Replica to protobuf JSON format.
func (r *Replica) MarshalJSON() ([]byte, error) {
	return (*netmap.Replica)(r).MarshalJSON()
}

// UnmarshalJSON decodes Replica from protobuf JSON format.
func (r *Replica) UnmarshalJSON(data []byte) error {
	return (*netmap.Replica)(r).UnmarshalJSON(data)
}
