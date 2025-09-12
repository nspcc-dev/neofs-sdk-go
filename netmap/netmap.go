package netmap

import (
	"fmt"
	"slices"

	"github.com/nspcc-dev/hrw/v2"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
)

// NetMap represents NeoFS network map. It includes information about all
// storage nodes registered in NeoFS the network.
//
// NetMap is mutually compatible with [protonetmap.Netmap] message. See
// [NetMap.FromProtoMessage] / [NetMap.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type NetMap struct {
	epoch uint64

	nodes []NodeInfo
}

// FromProtoMessage validates msg according to the NeoFS API protocol and
// restores m from it.
//
// See also [NetMap.ProtoMessage].
func (m *NetMap) FromProtoMessage(msg *protonetmap.Netmap) error {
	var err error

	if msg.Nodes == nil {
		m.nodes = nil
	} else {
		m.nodes = make([]NodeInfo, len(msg.Nodes))

		for i := range msg.Nodes {
			if msg.Nodes[i] == nil {
				return fmt.Errorf("nil node info #%d", i)
			}
			err = m.nodes[i].FromProtoMessage(msg.Nodes[i])
			if err != nil {
				return fmt.Errorf("invalid node info: %w", err)
			}
		}
	}

	m.epoch = msg.Epoch

	return nil
}

// ProtoMessage converts m into message to transmit using the NeoFS API
// protocol.
//
// See also [NetMap.FromProtoMessage].
func (m NetMap) ProtoMessage() *protonetmap.Netmap {
	msg := &protonetmap.Netmap{
		Epoch: m.epoch,
	}
	if m.nodes != nil {
		msg.Nodes = make([]*protonetmap.NodeInfo, len(m.nodes))

		for i := range m.nodes {
			msg.Nodes[i] = m.nodes[i].ProtoMessage()
		}
	}

	return msg
}

// SetNodes sets information list about all storage nodes from the NeoFS network.
//
// Argument MUST NOT be mutated, make a copy first.
//
// See also Nodes.
func (m *NetMap) SetNodes(nodes []NodeInfo) {
	m.nodes = nodes
}

// Nodes returns nodes set using SetNodes.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (m NetMap) Nodes() []NodeInfo {
	return m.nodes
}

// SetEpoch specifies revision number of the NetMap.
//
// See also Epoch.
func (m *NetMap) SetEpoch(epoch uint64) {
	m.epoch = epoch
}

// Epoch returns epoch set using SetEpoch.
//
// Zero NetMap has zero revision.
func (m NetMap) Epoch() uint64 {
	return m.epoch
}

// nodes is a slice of NodeInfo instances needed for HRW sorting.
type nodes []NodeInfo

// assert nodes type provides hrw.Hasher required for HRW sorting.
var _ hrw.Hashable = nodes{}

// Hash is a function from hrw.Hasher interface. It is implemented
// to support weighted hrw sorting of buckets. Each bucket is already sorted by hrw,
// thus giving us needed "randomness".
func (n nodes) Hash() uint64 {
	if len(n) > 0 {
		return n[0].Hash()
	}

	return 0
}

// weights returns slice of nodes weights W.
func (n nodes) weights(wf weightFunc) []float64 {
	w := make([]float64, 0, len(n))
	for i := range n {
		w = append(w, wf(n[i]))
	}

	return w
}

func flattenNodes(ns []nodes) nodes {
	var sz, i int

	for i = range ns {
		sz += len(ns[i])
	}

	result := make(nodes, 0, sz)

	for i := range ns {
		result = append(result, ns[i]...)
	}

	return result
}

// PlacementVectors sorts container nodes returned by ContainerNodes method
// and returns placement vectors for the entity identified by the given object id.
// For example, in order to build node list to store the object, binary-encoded
// object identifier can be used as pivot. Result is deterministic for
// the fixed NetMap and parameters.
func (m NetMap) PlacementVectors(vectors [][]NodeInfo, objectID oid.ID) ([][]NodeInfo, error) {
	pivot := slices.Clone(objectID[:])

	h := hrw.WrapBytes(pivot)
	wf := defaultWeightFunc(m.nodes)
	result := make([][]NodeInfo, len(vectors))

	for i := range vectors {
		result[i] = slices.Clone(vectors[i])
		hrw.SortWeighted(result[i], nodes(result[i]).weights(wf), h)
	}

	return result, nil
}

// ContainerNodes returns two-dimensional list of nodes as a result of applying
// given PlacementPolicy to the NetMap. Each line of the list corresponds to a
// replica descriptor. Line order corresponds to order of ReplicaDescriptor list
// in the policy. Nodes are pre-filtered according to the Filter list from
// the policy, and then selected by Selector list. Result is not deterministic and
// node order in each vector may vary for call.
//
// Result can be used in PlacementVectors.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// Returns ErrNotEnoughNodes if placement policy cannot be satisfied in the
// current network map because of a lack of container nodes.
func (m NetMap) ContainerNodes(p PlacementPolicy, containerID cid.ID) ([][]NodeInfo, error) {
	c := newContext(m)
	c.setCBF(p.backupFactor)

	pivot := slices.Clone(containerID[:])
	c.setPivot(pivot)

	if err := c.processFilters(p); err != nil {
		return nil, err
	}

	if err := c.processSelectors(p); err != nil {
		return nil, err
	}

	result := make([][]NodeInfo, len(p.replicas))

	for i := range p.replicas {
		sName := p.replicas[i].SelectorName()
		if sName == "" {
			if len(p.selectors) == 0 {
				var s Selector
				s.SetNumberOfNodes(p.replicas[i].NumberOfObjects())
				s.SetFilterName(mainFilterName)

				nodes, err := c.getSelection(p, s)
				if err != nil {
					return nil, err
				}

				result[i] = flattenNodes(nodes)
			}

			for _, s := range p.selectors {
				result[i] = append(result[i], flattenNodes(c.selections[s.Name()])...)
			}

			continue
		}

		nodes, ok := c.selections[sName]
		if !ok {
			return nil, fmt.Errorf("selector not found: REPLICA '%s'", sName)
		}

		result[i] = append(result[i], flattenNodes(nodes)...)
	}

	return result, nil
}
