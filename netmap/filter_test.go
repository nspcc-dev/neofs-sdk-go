package netmap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext_ProcessFilters(t *testing.T) {
	fs := []Filter{
		newFilter("StorageSSD", "Storage", "SSD", FilterOpEQ),
		newFilter("GoodRating", "Rating", "4", FilterOpGE),
		newFilter("Main", "", "", FilterOpAND,
			newFilter("StorageSSD", "", "", 0),
			newFilter("", "IntField", "123", FilterOpLT),
			newFilter("GoodRating", "", "", 0)),
	}

	c := newContext(NetMap{})
	p := newPlacementPolicy(1, nil, nil, fs)
	require.NoError(t, c.processFilters(p))
	require.Equal(t, 3, len(c.processedFilters))
	for _, f := range fs {
		require.Equal(t, f, *c.processedFilters[f.Name()])
	}

	require.Equal(t, uint64(4), c.numCache[fs[1].Value()])
	require.Equal(t, uint64(123), c.numCache[fs[2].SubFilters()[1].Value()])
}

func TestContext_ProcessFiltersInvalid(t *testing.T) {
	errTestCases := []struct {
		name   string
		filter Filter
		err    error
	}{
		{
			"UnnamedTop",
			newFilter("", "Storage", "SSD", FilterOpEQ),
			errUnnamedTopFilter,
		},
		{
			"InvalidReference",
			newFilter("Main", "", "", FilterOpAND,
				newFilter("StorageSSD", "", "", 0)),
			errFilterNotFound,
		},
		{
			"NonEmptyKeyed",
			newFilter("Main", "Storage", "SSD", FilterOpEQ,
				newFilter("StorageSSD", "", "", 0)),
			errNonEmptyFilters,
		},
		{
			"InvalidNumber",
			newFilter("Main", "Rating", "three", FilterOpGE),
			errInvalidNumber,
		},
		{
			"InvalidOp",
			newFilter("Main", "Rating", "3", 0),
			errInvalidFilterOp,
		},
		{
			"InvalidName",
			newFilter("*", "Rating", "3", FilterOpGE),
			errInvalidFilterName,
		},
	}
	for _, tc := range errTestCases {
		t.Run(tc.name, func(t *testing.T) {
			c := newContext(NetMap{})
			p := newPlacementPolicy(1, nil, nil, []Filter{tc.filter})
			err := c.processFilters(p)
			require.True(t, errors.Is(err, tc.err), "got: %v", err)
		})
	}
}

func TestFilter_MatchSimple_InvalidOp(t *testing.T) {
	var b NodeInfo
	b.SetAttribute("Rating", "4")
	b.SetAttribute("Country", "Germany")

	f := newFilter("Main", "Rating", "5", FilterOpEQ)
	c := newContext(NetMap{})
	p := newPlacementPolicy(1, nil, nil, []Filter{f})
	require.NoError(t, c.processFilters(p))

	// just for the coverage
	f.op = 0
	require.False(t, c.match(&f, b))
}
