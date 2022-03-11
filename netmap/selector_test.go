package netmap

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	"github.com/nspcc-dev/hrw"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	testv2 "github.com/nspcc-dev/neofs-api-go/v2/netmap/test"
	"github.com/stretchr/testify/require"
)

func BenchmarkHRWSort(b *testing.B) {
	const netmapSize = 1000

	nodes := make([]Nodes, netmapSize)
	weights := make([]float64, netmapSize)
	for i := range nodes {
		nodes[i] = Nodes{{
			ID:       rand.Uint64(),
			Index:    i,
			Capacity: 100,
			Price:    1,
			AttrMap:  nil,
		}}
		weights[i] = float64(rand.Uint32()%10) / 10.0
	}

	pivot := rand.Uint64()
	b.Run("sort by index, no weight", func(b *testing.B) {
		realNodes := make([]Nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, nodes)
			b.StartTimer()

			hrw.SortSliceByIndex(realNodes, pivot)
		}
	})
	b.Run("sort by value, no weight", func(b *testing.B) {
		realNodes := make([]Nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, nodes)
			b.StartTimer()

			hrw.SortSliceByValue(realNodes, pivot)
		}
	})
	b.Run("only sort by index", func(b *testing.B) {
		realNodes := make([]Nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, nodes)
			b.StartTimer()

			hrw.SortSliceByWeightIndex(realNodes, weights, pivot)
		}
	})
	b.Run("sort by value", func(b *testing.B) {
		realNodes := make([]Nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, nodes)
			b.StartTimer()

			hrw.SortSliceByWeightValue(realNodes, weights, pivot)
		}
	})
	b.Run("sort by ID, then by index (deterministic)", func(b *testing.B) {
		realNodes := make([]Nodes, netmapSize)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			copy(realNodes, nodes)
			b.StartTimer()

			sort.Slice(nodes, func(i, j int) bool {
				return nodes[i][0].ID < nodes[j][0].ID
			})
			hrw.SortSliceByWeightIndex(realNodes, weights, pivot)
		}
	})
}

func BenchmarkPolicyHRWType(b *testing.B) {
	const netmapSize = 100

	p := newPlacementPolicy(1,
		[]Replica{
			newReplica(1, "loc1"),
			newReplica(1, "loc2")},
		[]Selector{
			newSelector("loc1", "Location", ClauseSame, 1, "loc1"),
			newSelector("loc2", "Location", ClauseSame, 1, "loc2")},
		[]Filter{
			newFilter("loc1", "Location", "Shanghai", OpEQ),
			newFilter("loc2", "Location", "Shanghai", OpNE),
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

	nm, err := NewNetmap(NodesFromInfo(nodes))
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nm.GetContainerNodes(p, []byte{1})
		if err != nil {
			b.Fatal()
		}
	}
}

func TestPlacementPolicy_DeterministicOrder(t *testing.T) {
	const netmapSize = 100

	p := newPlacementPolicy(1,
		[]Replica{
			newReplica(1, "loc1"),
			newReplica(1, "loc2")},
		[]Selector{
			newSelector("loc1", "Location", ClauseSame, 1, "loc1"),
			newSelector("loc2", "Location", ClauseSame, 1, "loc2")},
		[]Filter{
			newFilter("loc1", "Location", "Shanghai", OpEQ),
			newFilter("loc2", "Location", "Shanghai", OpNE),
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

	nm, err := NewNetmap(NodesFromInfo(nodes))
	require.NoError(t, err)
	getIndices := func(t *testing.T) (int, int) {
		v, err := nm.GetContainerNodes(p, []byte{1})
		require.NoError(t, err)
		ns := v.Flatten()
		require.Equal(t, 2, len(ns))
		return ns[0].Index, ns[1].Index
	}

	a, b := getIndices(t)
	for i := 0; i < 10; i++ {
		x, y := getIndices(t)
		require.Equal(t, a, x)
		require.Equal(t, b, y)
	}
}

func TestPlacementPolicy_ProcessSelectors(t *testing.T) {
	p := newPlacementPolicy(2, nil,
		[]Selector{
			newSelector("SameRU", "City", ClauseSame, 2, "FromRU"),
			newSelector("DistinctRU", "City", ClauseDistinct, 2, "FromRU"),
			newSelector("Good", "Country", ClauseDistinct, 2, "Good"),
			newSelector("Main", "Country", ClauseDistinct, 3, "*"),
		},
		[]Filter{
			newFilter("FromRU", "Country", "Russia", OpEQ),
			newFilter("Good", "Rating", "4", OpGE),
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

	nm, err := NewNetmap(NodesFromInfo(nodes))
	require.NoError(t, err)
	c := NewContext(nm)
	c.setCBF(p.ContainerBackupFactor())
	require.NoError(t, c.processFilters(p))
	require.NoError(t, c.processSelectors(p))

	for _, s := range p.Selectors() {
		sel := c.Selections[s.Name()]
		s := c.Selectors[s.Name()]
		bucketCount, nodesInBucket := GetNodesCount(p, s)
		nodesInBucket *= int(c.cbf)
		targ := fmt.Sprintf("selector '%s'", s.Name())
		require.Equal(t, bucketCount, len(sel), targ)
		for _, res := range sel {
			require.Equal(t, nodesInBucket, len(res), targ)
			for j := range res {
				require.True(t, c.applyFilter(s.Filter(), &res[j]), targ)
			}
		}
	}
}

func testSelector() *Selector {
	s := new(Selector)
	s.SetName("name")
	s.SetCount(3)
	s.SetFilter("filter")
	s.SetAttribute("attribute")
	s.SetClause(ClauseDistinct)

	return s
}

func TestSelector_Name(t *testing.T) {
	s := NewSelector()
	name := "some name"

	s.SetName(name)

	require.Equal(t, name, s.Name())
}

func TestSelector_Count(t *testing.T) {
	s := NewSelector()
	c := uint32(3)

	s.SetCount(c)

	require.Equal(t, c, s.Count())
}

func TestSelector_Clause(t *testing.T) {
	s := NewSelector()
	c := ClauseSame

	s.SetClause(c)

	require.Equal(t, c, s.Clause())
}

func TestSelector_Attribute(t *testing.T) {
	s := NewSelector()
	a := "some attribute"

	s.SetAttribute(a)

	require.Equal(t, a, s.Attribute())
}

func TestSelector_Filter(t *testing.T) {
	s := NewSelector()
	f := "some filter"

	s.SetFilter(f)

	require.Equal(t, f, s.Filter())
}

func TestSelectorEncoding(t *testing.T) {
	s := newSelector("name", "atte", ClauseSame, 1, "filter")

	t.Run("binary", func(t *testing.T) {
		data, err := s.Marshal()
		require.NoError(t, err)

		s2 := *NewSelector()
		require.NoError(t, s2.Unmarshal(data))

		require.Equal(t, s, s2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := s.MarshalJSON()
		require.NoError(t, err)

		s2 := *NewSelector()
		require.NoError(t, s2.UnmarshalJSON(data))

		require.Equal(t, s, s2)
	})
}

func TestSelector_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *Selector

		require.Nil(t, x.ToV2())
	})
}

func TestNewSelectorFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *netmap.Selector

		require.Nil(t, NewSelectorFromV2(x))
	})

	t.Run("from non-nil", func(t *testing.T) {
		sV2 := testv2.GenerateSelector(false)

		s := NewSelectorFromV2(sV2)

		require.Equal(t, sV2, s.ToV2())
	})
}

func TestNewSelector(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		s := NewSelector()

		// check initial values
		require.Zero(t, s.Count())
		require.Equal(t, ClauseUnspecified, s.Clause())
		require.Empty(t, s.Attribute())
		require.Empty(t, s.Name())
		require.Empty(t, s.Filter())

		// convert to v2 message
		sV2 := s.ToV2()

		require.Zero(t, sV2.GetCount())
		require.Equal(t, netmap.UnspecifiedClause, sV2.GetClause())
		require.Empty(t, sV2.GetAttribute())
		require.Empty(t, sV2.GetName())
		require.Empty(t, sV2.GetFilter())
	})
}
