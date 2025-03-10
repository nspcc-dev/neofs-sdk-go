package netmap

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/nspcc-dev/hrw/v2"
)

// processSelectors processes selectors and returns error is any of them is invalid.
func (c *context) processSelectors(p PlacementPolicy) error {
	for i := range p.selectors {
		fName := p.selectors[i].FilterName()
		if fName != mainFilterName {
			_, ok := c.processedFilters[p.selectors[i].FilterName()]
			if !ok {
				return fmt.Errorf("%w: SELECT FROM '%s'", errFilterNotFound, fName)
			}
		}

		sName := p.selectors[i].Name()

		c.processedSelectors[sName] = &p.selectors[i]

		result, err := c.getSelection(p, p.selectors[i])
		if err != nil {
			return err
		}

		c.selections[sName] = result
	}

	return nil
}

// calcNodesCount returns number of buckets and minimum number of nodes in every bucket
// for the given selector.
func calcNodesCount(s Selector) (int, int) {
	n := int(s.NumberOfNodes())
	if s.IsSame() {
		return 1, n
	}
	return n, 1
}

// calcBucketWeight computes weight for a node bucket.
func calcBucketWeight(ns nodes, a aggregator, wf weightFunc) float64 {
	for i := range ns {
		a.Add(wf(ns[i]))
	}

	return a.Compute()
}

// getSelection returns nodes grouped by s.attribute.
// Last argument specifies if more buckets can be used to fulfill CBF.
func (c *context) getSelection(_ PlacementPolicy, s Selector) ([]nodes, error) {
	bucketCount, nodesInBucket := calcNodesCount(s)
	buckets := c.getSelectionBase(s)

	if len(buckets) < bucketCount {
		return nil, fmt.Errorf("%w: '%s'", ErrNotEnoughNodes, s.Name())
	}

	// We need deterministic output in case there is no pivot.
	// If pivot is set, buckets are sorted by HRW.
	// However, because initial order influences HRW order for buckets with equal weights,
	// we also need to have deterministic input to HRW sorting routine.
	if len(c.hrwSeed) == 0 {
		if s.BucketAttribute() == "" {
			slices.SortFunc(buckets, func(a, b nodeAttrPair) int {
				return cmp.Compare(a.nodes[0].Hash(), b.nodes[0].Hash())
			})
		} else {
			slices.SortFunc(buckets, func(a, b nodeAttrPair) int {
				return cmp.Compare(a.attr, b.attr)
			})
		}
	}

	maxNodesInBucket := nodesInBucket * int(c.cbf)
	res := make([]nodes, 0, len(buckets))
	fallback := make([]nodes, 0, len(buckets))

	for i := range buckets {
		ns := buckets[i].nodes
		if len(ns) >= maxNodesInBucket {
			res = append(res, ns[:maxNodesInBucket])
		} else if len(ns) >= nodesInBucket {
			fallback = append(fallback, ns)
		}
	}

	if len(res) < bucketCount {
		// Fallback to using minimum allowed backup factor (1).
		res = append(res, fallback...)
		if len(res) < bucketCount {
			return nil, fmt.Errorf("%w: '%s'", ErrNotEnoughNodes, s.Name())
		}
	}

	if len(c.hrwSeed) != 0 {
		weights := make([]float64, len(res))
		for i := range res {
			weights[i] = calcBucketWeight(res[i], newMeanIQRAgg(), c.weightFunc)
		}

		hrw.SortWeighted(res, weights, c.hrwSeedHash)
	}

	if s.BucketAttribute() == "" {
		res, fallback = res[:bucketCount], res[bucketCount:]
		for i := range fallback {
			index := i % bucketCount
			if len(res[index]) >= maxNodesInBucket {
				break
			}
			res[index] = append(res[index], fallback[i]...)
		}
	}

	return res[:bucketCount], nil
}

type nodeAttrPair struct {
	attr  string
	nodes nodes
}

// getSelectionBase returns nodes grouped by selector attribute.
// It it guaranteed that each pair will contain at least one node.
func (c *context) getSelectionBase(s Selector) []nodeAttrPair {
	fName := s.FilterName()
	f := c.processedFilters[fName]
	isMain := fName == mainFilterName
	result := []nodeAttrPair{}
	nodeMap := map[string][]NodeInfo{}
	attr := s.BucketAttribute()

	for i := range c.netMap.nodes {
		if isMain || c.match(f, c.netMap.nodes[i]) {
			if attr == "" {
				// Default attribute is transparent identifier which is different for every node.
				result = append(result, nodeAttrPair{attr: "", nodes: nodes{c.netMap.nodes[i]}})
			} else {
				v := c.netMap.nodes[i].Attribute(attr)
				nodeMap[v] = append(nodeMap[v], c.netMap.nodes[i])
			}
		}
	}

	if attr != "" {
		for k, ns := range nodeMap {
			result = append(result, nodeAttrPair{attr: k, nodes: ns})
		}
	}

	if len(c.hrwSeed) != 0 {
		for i := range result {
			hrw.SortWeighted(result[i].nodes, result[i].nodes.weights(c.weightFunc), c.hrwSeedHash)
		}
	}

	return result
}
