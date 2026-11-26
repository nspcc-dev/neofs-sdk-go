package netmap

import (
	"cmp"
	"errors"
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

// getSelection returns nodes grouped by s.attribute.
// Last argument specifies if more buckets can be used to fulfill CBF.
func (c *context) getSelection(policy PlacementPolicy, s Selector) ([]nodes, error) {
	var bucketCount int
	nodesInBucket := int(s.NumberOfNodes())

	buckets := c.getSelectionBase(s)

	rigidCBF := !isFlexibleCBF(c.cbf)

	selectAll := nodesInBucket == 0

	var maxNodesInBucket int
	if selectAll {
		if rigidCBF {
			return nil, errors.New("SELECT * with non-flexible CBF")
		}

		if s.IsSame() {
			bucketCount = 1
			if n := getMinRequiredNodeCountForSelector(policy, s.Name()); n > 0 {
				nodesInBucket = n
			} // else keep 0 so that every bucket fits
		} else { // distinct
			bucketCount = len(buckets)
		}
	} else {
		if s.IsSame() {
			bucketCount = 1
		} else { // distinct
			bucketCount = nodesInBucket
			nodesInBucket = 1
		}

		maxNodesInBucket = nodesInBucket
		if rigidCBF {
			maxNodesInBucket *= int(c.cbf)
		}
	}

	if rigidCBF && len(buckets) < bucketCount {
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

	res := make([]nodes, 0, len(buckets))
	var fallback []nodes
	if !selectAll {
		fallback = make([]nodes, 0, len(buckets))
	}

	for i := range buckets {
		ns := buckets[i].nodes
		if selectAll {
			if s.IsSame() {
				if len(ns) >= nodesInBucket {
					res = append(res, ns)
				}
			} else { // distinct
				res = append(res, ns[:1])
			}
		} else if len(ns) >= maxNodesInBucket {
			res = append(res, ns[:maxNodesInBucket])
		} else if len(ns) >= nodesInBucket {
			fallback = append(fallback, ns)
		}
	}

	if len(res) < bucketCount {
		// Fallback to using minimum allowed backup factor (1).
		res = append(res, fallback...)
		if rigidCBF && len(res) < bucketCount {
			return nil, fmt.Errorf("%w: '%s'", ErrNotEnoughNodes, s.Name())
		}
	}

	if len(c.hrwSeed) != 0 {
		weights := make([]float64, len(res))
		for i := range res {
			weights[i] = res[i].calcBucketWeight()
		}

		hrw.SortWeighted(res, weights, c.hrwSeedHash)
	}

	bucketCount = min(bucketCount, len(res))

	if !selectAll && s.BucketAttribute() == "" {
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
			hrw.SortWeighted(result[i].nodes, result[i].nodes.weights(), c.hrwSeedHash)
		}
	}

	return result
}

func getMinRequiredNodeCountForSelector(policy PlacementPolicy, sel string) int {
	m := -1
	for _, r := range policy.Replicas() {
		match := (sel != "" && r.SelectorName() == sel) || (r.SelectorName() == "" && len(policy.Selectors()) == 1)
		if !match {
			continue
		}
		n := int(r.NumberOfObjects())
		if n > m {
			m = n
		}
	}
	for _, r := range policy.ecRules {
		match := (sel != "" && r.SelectorName() == sel) || (r.SelectorName() == "" && len(policy.Selectors()) == 1)
		if !match {
			continue
		}
		n := int(r.DataPartNum() + r.ParityPartNum())
		if n > m {
			m = n
		}
	}
	return m
}
