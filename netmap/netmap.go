package netmap

import (
	"crypto/sha256"
	"fmt"

	"github.com/nspcc-dev/hrw"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

const defaultCBF = 3

// Netmap represents netmap which contains preprocessed nodes.
type Netmap struct {
	Nodes Nodes
}

// NewNetmap constructs netmap from the list of raw nodes.
func NewNetmap(nodes Nodes) (*Netmap, error) {
	return &Netmap{
		Nodes: nodes,
	}, nil
}

func flattenNodes(ns []Nodes) Nodes {
	result := make(Nodes, 0, len(ns))
	for i := range ns {
		result = append(result, ns[i]...)
	}

	return result
}

// GetPlacementVectors returns placement vectors for an object given containerNodes cnt.
func (m *Netmap) GetPlacementVectors(cnt ContainerNodes, pivot []byte) ([]Nodes, error) {
	h := hrw.Hash(pivot)
	wf := GetDefaultWeightFunc(m.Nodes)
	result := make([]Nodes, len(cnt.Replicas()))

	for i, rep := range cnt.Replicas() {
		result[i] = make(Nodes, len(rep))
		copy(result[i], rep)
		hrw.SortSliceByWeightValue(result[i], result[i].Weights(wf), h)
	}

	return result, nil
}

// GetContainerNodes returns nodes corresponding to each replica.
// Order of returned nodes corresponds to order of replicas in p.
// Container ID is used as a seed for HRW sorting.
func (m *Netmap) GetContainerNodes(p *PlacementPolicy, cnrID cid.ID) (ContainerNodes, error) {
	pivot := make([]byte, sha256.Size)
	cnrID.Encode(pivot)

	c := NewContext(m)
	c.setPivot(pivot)
	c.setCBF(p.ContainerBackupFactor())

	if err := c.processFilters(p); err != nil {
		return nil, err
	}

	if err := c.processSelectors(p); err != nil {
		return nil, err
	}

	result := make([]Nodes, len(p.Replicas()))

	for i, r := range p.Replicas() {
		if r.Selector() == "" {
			if len(p.Selectors()) == 0 {
				s := new(Selector)
				s.SetCount(r.Count())
				s.SetFilter(MainFilterName)

				nodes, err := c.getSelection(p, s)
				if err != nil {
					return nil, err
				}

				result[i] = flattenNodes(nodes)
			}

			for _, s := range p.Selectors() {
				result[i] = append(result[i], flattenNodes(c.Selections[s.Name()])...)
			}

			continue
		}

		nodes, ok := c.Selections[r.Selector()]
		if !ok {
			return nil, fmt.Errorf("%w: REPLICA '%s'", ErrSelectorNotFound, r.Selector())
		}

		result[i] = append(result[i], flattenNodes(nodes)...)
	}

	return containerNodes(result), nil
}
