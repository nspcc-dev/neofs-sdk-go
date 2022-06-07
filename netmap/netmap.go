package netmap

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/nspcc-dev/hrw"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
)

const defaultCBF = 3

var _, _ hrw.Hasher = NodeInfo{}, nodes{}

// Hash implements hrw.Hasher interface.
//
// Hash is needed to support weighted HRW therefore sort function sorts nodes
// based on their public key.
func (i NodeInfo) Hash() uint64 {
	return hrw.Hash(i.m.GetPublicKey())
}

func (i NodeInfo) less(i2 NodeInfo) bool {
	return bytes.Compare(i.PublicKey(), i2.PublicKey()) < 0
}

// attribute returns value of the node attribute by the given key. Returns empty
// string if attribute is missing.
//
// Method is needed to internal placement needs.
func (i NodeInfo) attribute(key string) string {
	as := i.m.GetAttributes()
	for j := range as {
		if as[j].GetKey() == key {
			return as[j].GetValue()
		}
	}

	return ""
}

func (i *NodeInfo) syncAttributes() {
	as := i.m.GetAttributes()
	for j := range as {
		switch as[j].GetKey() {
		case AttrPrice:
			i.priceAttr, _ = strconv.ParseUint(as[j].GetValue(), 10, 64)
		case AttrCapacity:
			i.capAttr, _ = strconv.ParseUint(as[j].GetValue(), 10, 64)
		}
	}
}

func (i *NodeInfo) setPrice(price uint64) {
	i.priceAttr = price

	as := i.m.GetAttributes()
	for j := range as {
		if as[j].GetKey() == AttrPrice {
			as[j].SetValue(strconv.FormatUint(i.capAttr, 10))
			return
		}
	}

	as = append(as, netmap.Attribute{})
	as[len(as)-1].SetKey(AttrPrice)
	as[len(as)-1].SetValue(strconv.FormatUint(i.capAttr, 10))

	i.m.SetAttributes(as)
}

func (i *NodeInfo) price() uint64 {
	return i.priceAttr
}

func (i *NodeInfo) setCapacity(capacity uint64) {
	i.capAttr = capacity

	as := i.m.GetAttributes()
	for j := range as {
		if as[j].GetKey() == AttrCapacity {
			as[j].SetValue(strconv.FormatUint(i.capAttr, 10))
			return
		}
	}

	as = append(as, netmap.Attribute{})
	as[len(as)-1].SetKey(AttrCapacity)
	as[len(as)-1].SetValue(strconv.FormatUint(i.capAttr, 10))

	i.m.SetAttributes(as)
}

func (i NodeInfo) capacity() uint64 {
	return i.capAttr
}

// Netmap represents netmap which contains preprocessed nodes.
type Netmap struct {
	nodes []NodeInfo
}

func (m *Netmap) SetNodes(nodes []NodeInfo) {
	m.nodes = nodes
}

type nodes []NodeInfo

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

// GetPlacementVectors returns placement vectors for an object given containerNodes cnt.
func (m *Netmap) GetPlacementVectors(vectors [][]NodeInfo, pivot []byte) ([][]NodeInfo, error) {
	h := hrw.Hash(pivot)
	wf := GetDefaultWeightFunc(m.nodes)
	result := make([][]NodeInfo, len(vectors))

	for i := range vectors {
		result[i] = make([]NodeInfo, len(vectors[i]))
		copy(result[i], vectors[i])
		hrw.SortSliceByWeightValue(result[i], nodes(result[i]).weights(wf), h)
	}

	return result, nil
}

// GetContainerNodes returns nodes corresponding to each replica.
// Order of returned nodes corresponds to order of replicas in p.
// pivot is a seed for HRW sorting.
func (m *Netmap) GetContainerNodes(p *PlacementPolicy, pivot []byte) ([][]NodeInfo, error) {
	c := newContext(m)
	c.setPivot(pivot)
	c.setCBF(p.ContainerBackupFactor())

	if err := c.processFilters(p); err != nil {
		return nil, err
	}

	if err := c.processSelectors(p); err != nil {
		return nil, err
	}

	result := make([][]NodeInfo, len(p.Replicas()))

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

	return result, nil
}
