package netmap

import (
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
)

// PlacementPolicy represents v2-compatible placement policy.
type PlacementPolicy netmap.PlacementPolicy

// NewPlacementPolicy creates and returns new PlacementPolicy instance.
//
// Defaults:
//  - backupFactor: 0;
//  - replicas nil;
//  - selectors nil;
//  - filters nil.
func NewPlacementPolicy() *PlacementPolicy {
	return NewPlacementPolicyFromV2(new(netmap.PlacementPolicy))
}

// NewPlacementPolicyFromV2 converts v2 PlacementPolicy to PlacementPolicy.
//
// Nil netmap.PlacementPolicy converts to nil.
func NewPlacementPolicyFromV2(f *netmap.PlacementPolicy) *PlacementPolicy {
	return (*PlacementPolicy)(f)
}

// ToV2 converts PlacementPolicy to v2 PlacementPolicy.
//
// Nil PlacementPolicy converts to nil.
func (p *PlacementPolicy) ToV2() *netmap.PlacementPolicy {
	return (*netmap.PlacementPolicy)(p)
}

// SubnetID returns subnet to select nodes from.
func (p *PlacementPolicy) SubnetID() *subnetid.ID {
	idv2 := (*netmap.PlacementPolicy)(p).GetSubnetID()
	if idv2 == nil {
		return nil
	}

	var id subnetid.ID

	err := id.ReadFromV2(*idv2)
	if err != nil {
		panic(err) // will disappear after netmap package refactor
	}

	return &id
}

// SetSubnetID sets subnet to select nodes from.
func (p *PlacementPolicy) SetSubnetID(subnet *subnetid.ID) {
	var idv2 *refs.SubnetID

	if subnet != nil {
		idv2 = new(refs.SubnetID)
		subnet.WriteToV2(idv2)
	}

	(*netmap.PlacementPolicy)(p).SetSubnetID(idv2)
}

// Replicas returns list of object replica descriptors.
func (p *PlacementPolicy) Replicas() []Replica {
	rs := (*netmap.PlacementPolicy)(p).
		GetReplicas()

	if rs == nil {
		return nil
	}

	res := make([]Replica, len(rs))

	for i := range rs {
		res[i] = *NewReplicaFromV2(&rs[i])
	}

	return res
}

// SetReplicas sets list of object replica descriptors.
func (p *PlacementPolicy) SetReplicas(rs ...Replica) {
	var rsV2 []netmap.Replica

	if rs != nil {
		rsV2 = make([]netmap.Replica, len(rs))

		for i := range rs {
			rsV2[i] = *rs[i].ToV2()
		}
	}

	(*netmap.PlacementPolicy)(p).SetReplicas(rsV2)
}

// ContainerBackupFactor returns container backup factor.
func (p *PlacementPolicy) ContainerBackupFactor() uint32 {
	return (*netmap.PlacementPolicy)(p).
		GetContainerBackupFactor()
}

// SetContainerBackupFactor sets container backup factor.
func (p *PlacementPolicy) SetContainerBackupFactor(f uint32) {
	(*netmap.PlacementPolicy)(p).
		SetContainerBackupFactor(f)
}

// Selector returns set of selectors to form the container's nodes subset.
func (p *PlacementPolicy) Selectors() []Selector {
	rs := (*netmap.PlacementPolicy)(p).
		GetSelectors()

	if rs == nil {
		return nil
	}

	res := make([]Selector, len(rs))

	for i := range rs {
		res[i] = *NewSelectorFromV2(&rs[i])
	}

	return res
}

// SetSelectors sets set of selectors to form the container's nodes subset.
func (p *PlacementPolicy) SetSelectors(ss ...Selector) {
	var ssV2 []netmap.Selector

	if ss != nil {
		ssV2 = make([]netmap.Selector, len(ss))

		for i := range ss {
			ssV2[i] = *ss[i].ToV2()
		}
	}

	(*netmap.PlacementPolicy)(p).SetSelectors(ssV2)
}

// Filters returns list of named filters to reference in selectors.
func (p *PlacementPolicy) Filters() []Filter {
	return filtersFromV2(
		(*netmap.PlacementPolicy)(p).
			GetFilters(),
	)
}

// SetFilters sets list of named filters to reference in selectors.
func (p *PlacementPolicy) SetFilters(fs ...Filter) {
	(*netmap.PlacementPolicy)(p).
		SetFilters(filtersToV2(fs))
}

// Marshal marshals PlacementPolicy into a protobuf binary form.
func (p *PlacementPolicy) Marshal() ([]byte, error) {
	return (*netmap.PlacementPolicy)(p).StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of PlacementPolicy.
func (p *PlacementPolicy) Unmarshal(data []byte) error {
	return (*netmap.PlacementPolicy)(p).Unmarshal(data)
}

// MarshalJSON encodes PlacementPolicy to protobuf JSON format.
func (p *PlacementPolicy) MarshalJSON() ([]byte, error) {
	return (*netmap.PlacementPolicy)(p).MarshalJSON()
}

// UnmarshalJSON decodes PlacementPolicy from protobuf JSON format.
func (p *PlacementPolicy) UnmarshalJSON(data []byte) error {
	return (*netmap.PlacementPolicy)(p).UnmarshalJSON(data)
}
