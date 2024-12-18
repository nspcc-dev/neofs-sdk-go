package netmap

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlacementPolicy_CopyTo(t *testing.T) {
	var pp PlacementPolicy
	pp.SetContainerBackupFactor(123)

	var rd ReplicaDescriptor
	rd.SetSelectorName("selector")
	rd.SetNumberOfObjects(100)
	pp.SetReplicas([]ReplicaDescriptor{rd})

	var f Filter
	f.SetName("filter")
	pp.SetFilters([]Filter{f})

	var s Selector
	s.SetName("selector")
	pp.SetSelectors([]Selector{s})

	t.Run("copy", func(t *testing.T) {
		var dst PlacementPolicy
		pp.CopyTo(&dst)

		require.Equal(t, pp, dst)
		require.True(t, bytes.Equal(pp.Marshal(), dst.Marshal()))
	})

	t.Run("change filter", func(t *testing.T) {
		var dst PlacementPolicy
		pp.CopyTo(&dst)

		var f2 Filter
		f2.SetName("filter2")

		require.Equal(t, pp.filters[0].Name(), dst.filters[0].Name())
		dst.filters[0].SetName("f2")
		require.NotEqual(t, pp.filters[0].Name(), dst.filters[0].Name())

		dst.filters[0] = f2
		require.NotEqual(t, pp.filters[0].Name(), dst.filters[0].Name())
	})

	t.Run("internal filters", func(t *testing.T) {
		var includedFilter Filter
		includedFilter.SetName("includedFilter")

		var topFilter Filter
		topFilter.SetName("topFilter")
		topFilter.setInnerFilters(FilterOpEQ, []Filter{includedFilter})

		var policy PlacementPolicy
		policy.SetFilters([]Filter{topFilter})

		var dst PlacementPolicy
		policy.CopyTo(&dst)
		require.True(t, bytes.Equal(policy.Marshal(), dst.Marshal()))

		t.Run("change extra filter", func(t *testing.T) {
			require.Equal(t, topFilter.Name(), dst.filters[0].Name())
			require.Equal(t, topFilter.SubFilters()[0].Name(), dst.filters[0].SubFilters()[0].Name())

			dst.filters[0].SubFilters()[0].SetName("someInternalFilterName")

			require.Equal(t, topFilter.Name(), dst.filters[0].Name())
			require.NotEqual(t, topFilter.SubFilters()[0].Name(), dst.filters[0].SubFilters()[0].Name())
		})
	})

	t.Run("empty filters", func(t *testing.T) {
		var ppFilters PlacementPolicy
		ppFilters.SetContainerBackupFactor(123)

		var dst PlacementPolicy
		ppFilters.CopyTo(&dst)

		require.True(t, bytes.Equal(ppFilters.Marshal(), dst.Marshal()))
	})

	t.Run("change selector", func(t *testing.T) {
		var dst PlacementPolicy
		pp.CopyTo(&dst)

		require.Equal(t, pp.selectors[0].Name(), dst.selectors[0].Name())
		dst.selectors[0].SetName("s2")
		require.NotEqual(t, pp.selectors[0].Name(), dst.selectors[0].Name())

		var s2 Selector
		s2.SetName("selector2")

		dst.selectors[0] = s2
		require.NotEqual(t, pp.selectors[0].Name(), dst.selectors[0].Name())
	})

	t.Run("change replica", func(t *testing.T) {
		var dst PlacementPolicy
		pp.CopyTo(&dst)

		require.Equal(t, pp.replicas[0].SelectorName(), dst.replicas[0].SelectorName())
		dst.replicas[0].SetSelectorName("s2")
		require.NotEqual(t, pp.replicas[0].SelectorName(), dst.replicas[0].SelectorName())
	})
}
