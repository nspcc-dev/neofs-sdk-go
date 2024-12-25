package netmap

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/nspcc-dev/hrw/v2"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	"github.com/stretchr/testify/require"
)

type hashableUint uint64

func (h hashableUint) Hash() uint64 {
	return uint64(h)
}

func BenchmarkHRWSort(b *testing.B) {
	const netmapSize = 1000

	vectors := make([]nodes, netmapSize)
	weights := make([]float64, netmapSize)
	for i := range vectors {
		key := make([]byte, 33)
		//nolint:staticcheck
		rand.Read(key)

		var node NodeInfo
		node.SetPrice(1)
		node.SetCapacity(100)
		node.SetPublicKey(key)

		vectors[i] = nodes{node}
		weights[i] = float64(rand.Uint32()%10) / 10.0
	}

	pivot := hashableUint(rand.Uint64())
	b.Run("sort by value, no weight", func(b *testing.B) {
		realNodes := make([]nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, vectors)
			b.StartTimer()

			hrw.Sort(realNodes, pivot)
		}
	})
	b.Run("sort by value", func(b *testing.B) {
		realNodes := make([]nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, vectors)
			b.StartTimer()

			hrw.SortWeighted(realNodes, weights, pivot)
		}
	})
}

func BenchmarkPolicyHRWType(b *testing.B) {
	const netmapSize = 100

	p := newPlacementPolicy(1,
		[]ReplicaDescriptor{
			newReplica(1, "loc1"),
			newReplica(1, "loc2")},
		[]Selector{
			newSelector("loc1", "Location", 1, "loc1", (*Selector).SelectSame),
			newSelector("loc2", "Location", 1, "loc2", (*Selector).SelectSame)},
		[]Filter{
			newFilter("loc1", "Location", "Shanghai", FilterOpEQ),
			newFilter("loc2", "Location", "Shanghai", FilterOpNE),
		})

	nodes := make([]NodeInfo, netmapSize)
	for i := range nodes {
		var loc string
		switch i % 20 {
		case 0:
			loc = "Shanghai"
		default:
			loc = strconv.Itoa(i % 20)
		}

		// Having the same price and capacity ensures equal weights for all nodes.
		// This way placement is more dependent on the initial order.
		nodes[i] = nodeInfoFromAttributes("Location", loc, "Price", "1", "Capacity", "10")
		pub := make([]byte, 33)
		pub[0] = byte(i)
		nodes[i].SetPublicKey(pub)
	}

	var nm NetMap
	nm.SetNodes(nodes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nm.ContainerNodes(p, cid.ID{1})
		if err != nil {
			b.Fatal()
		}
	}
}

func TestPlacementPolicy_DeterministicOrder(t *testing.T) {
	const netmapSize = 100

	p := newPlacementPolicy(1,
		[]ReplicaDescriptor{
			newReplica(1, "loc1"),
			newReplica(1, "loc2")},
		[]Selector{
			newSelector("loc1", "Location", 1, "loc1", (*Selector).SelectSame),
			newSelector("loc2", "Location", 1, "loc2", (*Selector).SelectSame)},
		[]Filter{
			newFilter("loc1", "Location", "Shanghai", FilterOpEQ),
			newFilter("loc2", "Location", "Shanghai", FilterOpNE),
		})

	nodeList := make([]NodeInfo, netmapSize)
	for i := range nodeList {
		var loc string
		switch i % 20 {
		case 0:
			loc = "Shanghai"
		default:
			loc = strconv.Itoa(i % 20)
		}

		// Having the same price and capacity ensures equal weights for all nodes.
		// This way placement is more dependent on the initial order.
		nodeList[i] = nodeInfoFromAttributes("Location", loc, "Price", "1", "Capacity", "10")
		pub := make([]byte, 33)
		pub[0] = byte(i)
		nodeList[i].SetPublicKey(pub)
	}

	var nm NetMap
	nm.SetNodes(nodeList)

	getIndices := func(t *testing.T) (uint64, uint64) {
		v, err := nm.ContainerNodes(p, cid.ID{1})
		require.NoError(t, err)

		nss := make([]nodes, len(v))
		for i := range v {
			nss[i] = v[i]
		}

		ns := flattenNodes(nss)
		require.Equal(t, 2, len(ns))
		return ns[0].Hash(), ns[1].Hash()
	}

	a, b := getIndices(t)
	for range 10 {
		x, y := getIndices(t)
		require.Equal(t, a, x)
		require.Equal(t, b, y)
	}
}

func TestPlacementPolicy_ProcessSelectors(t *testing.T) {
	p := newPlacementPolicy(2, nil,
		[]Selector{
			newSelector("SameRU", "City", 2, "FromRU", (*Selector).SelectSame),
			newSelector("DistinctRU", "City", 2, "FromRU", (*Selector).SelectDistinct),
			newSelector("Good", "Country", 2, "Good", (*Selector).SelectDistinct),
			newSelector("Main", "Country", 3, "*", (*Selector).SelectDistinct),
		},
		[]Filter{
			newFilter("FromRU", "Country", "Russia", FilterOpEQ),
			newFilter("Good", "Rating", "4", FilterOpGE),
		})
	nodes := []NodeInfo{
		nodeInfoFromAttributes("Country", "Russia", "Rating", "1", "City", "SPB"),
		nodeInfoFromAttributes("Country", "Germany", "Rating", "5", "City", "Berlin"),
		nodeInfoFromAttributes("Country", "Russia", "Rating", "6", "City", "Moscow"),
		nodeInfoFromAttributes("Country", "France", "Rating", "4", "City", "Paris"),
		nodeInfoFromAttributes("Country", "France", "Rating", "1", "City", "Lyon"),
		nodeInfoFromAttributes("Country", "Russia", "Rating", "5", "City", "SPB"),
		nodeInfoFromAttributes("Country", "Russia", "Rating", "7", "City", "Moscow"),
		nodeInfoFromAttributes("Country", "Germany", "Rating", "3", "City", "Darmstadt"),
		nodeInfoFromAttributes("Country", "Germany", "Rating", "7", "City", "Frankfurt"),
		nodeInfoFromAttributes("Country", "Russia", "Rating", "9", "City", "SPB"),
		nodeInfoFromAttributes("Country", "Russia", "Rating", "9", "City", "SPB"),
	}

	var nm NetMap
	nm.SetNodes(nodes)
	c := newContext(nm)
	c.setCBF(p.backupFactor)
	require.NoError(t, c.processFilters(p))
	require.NoError(t, c.processSelectors(p))

	for _, s := range p.selectors {
		sel := c.selections[s.Name()]
		s := c.processedSelectors[s.Name()]
		bucketCount, nodesInBucket := calcNodesCount(*s)
		nodesInBucket *= int(c.cbf)
		targ := fmt.Sprintf("selector '%s'", s.Name())
		require.Equal(t, bucketCount, len(sel), targ)
		fName := s.FilterName()
		for _, res := range sel {
			require.Equal(t, nodesInBucket, len(res), targ)
			for j := range res {
				require.True(t, fName == mainFilterName || c.match(c.processedFilters[fName], res[j]), targ)
			}
		}
	}
}

func TestSelector_SetName(t *testing.T) {
	const name = "some name"
	var s Selector

	require.Zero(t, s.Name())

	s.SetName(name)
	require.Equal(t, name, s.Name())
}

func TestSelector_SetNumberOfNodes(t *testing.T) {
	const num = 3
	var s Selector

	require.Zero(t, s.NumberOfNodes())

	s.SetNumberOfNodes(num)

	require.EqualValues(t, num, s.NumberOfNodes())
}

func TestSelectorClauses(t *testing.T) {
	var s Selector

	require.Equal(t, protonetmap.Clause_CLAUSE_UNSPECIFIED, s.clause)
	require.False(t, s.IsSame())
	require.False(t, s.IsDistinct())

	s.SelectDistinct()
	require.Equal(t, protonetmap.Clause_DISTINCT, s.clause)
	require.False(t, s.IsSame())
	require.True(t, s.IsDistinct())

	s.SelectSame()
	require.Equal(t, protonetmap.Clause_SAME, s.clause)
	require.True(t, s.IsSame())
	require.False(t, s.IsDistinct())
}

func TestSelector_SelectByBucketAttribute(t *testing.T) {
	const attr = "some attribute"
	var s Selector

	require.Zero(t, s.BucketAttribute())

	s.SelectByBucketAttribute(attr)
	require.Equal(t, attr, s.BucketAttribute())
}

func TestSelector_SetFilterName(t *testing.T) {
	const fName = "some filter"
	var s Selector

	require.Zero(t, s.FilterName())

	s.SetFilterName(fName)
	require.Equal(t, fName, s.FilterName())
}
