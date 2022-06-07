package netmap

import (
	"errors"

	"github.com/nspcc-dev/hrw"
)

// context contains references to named filters and cached numeric values.
type context struct {
	// Netmap is a netmap structure to operate on.
	Netmap *Netmap
	// Filters stores processed filters.
	Filters map[string]*Filter
	// Selectors stores processed selectors.
	Selectors map[string]*Selector
	// Selections stores result of selector processing.
	Selections map[string][]nodes

	// numCache stores parsed numeric values.
	numCache map[string]uint64
	// pivot is a seed for HRW.
	pivot []byte
	// pivotHash is a saved HRW hash of pivot
	pivotHash uint64
	// aggregator is returns aggregator determining bucket weight.
	// By default it returns mean value from IQR interval.
	aggregator func() aggregator
	// weightFunc is a weighting function for determining node priority.
	// By default in combines favours low price and high capacity.
	weightFunc weightFunc
	// container backup factor is a factor for selector counters that expand
	// amount of chosen nodes.
	cbf uint32
}

// Various validation errors.
var (
	ErrMissingField      = errors.New("netmap: nil field")
	ErrInvalidFilterName = errors.New("netmap: filter name is invalid")
	ErrInvalidNumber     = errors.New("netmap: number value expected")
	ErrInvalidFilterOp   = errors.New("netmap: invalid filter operation")
	ErrFilterNotFound    = errors.New("netmap: filter not found")
	ErrNonEmptyFilters   = errors.New("netmap: simple filter must no contain sub-filters")
	ErrNotEnoughNodes    = errors.New("netmap: not enough nodes to SELECT from")
	ErrSelectorNotFound  = errors.New("netmap: selector not found")
	ErrUnnamedTopFilter  = errors.New("netmap: all filters on top level must be named")
)

// newContext creates new context. It contains various caches.
// In future it may create hierarchical netmap structure to work with.
func newContext(nm *Netmap) *context {
	return &context{
		Netmap:     nm,
		Filters:    make(map[string]*Filter),
		Selectors:  make(map[string]*Selector),
		Selections: make(map[string][]nodes),

		numCache:   make(map[string]uint64),
		aggregator: newMeanIQRAgg,
		weightFunc: defaultWeightFunc(nm.nodes),
		cbf:        defaultCBF,
	}
}

func (c *context) setPivot(pivot []byte) {
	if len(pivot) != 0 {
		c.pivot = pivot
		c.pivotHash = hrw.Hash(pivot)
	}
}

func (c *context) setCBF(cbf uint32) {
	if cbf == 0 {
		c.cbf = defaultCBF
	} else {
		c.cbf = cbf
	}
}

// defaultWeightFunc returns default weighting function.
func defaultWeightFunc(ns nodes) weightFunc {
	mean := newMeanAgg()
	min := newMinAgg()

	for i := range ns {
		mean.Add(float64(ns[i].capacity()))
		min.Add(float64(ns[i].price()))
	}

	return newWeightFunc(
		newSigmoidNorm(mean.Compute()),
		newReverseMinNorm(min.Compute()))
}
