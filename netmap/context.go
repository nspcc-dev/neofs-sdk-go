package netmap

import (
	"errors"

	"github.com/nspcc-dev/hrw/v2"
)

// context of a placement build process.
type context struct {
	// network map to operate on
	netMap NetMap

	// cache of processed filters
	processedFilters map[string]*Filter

	// cache of processed selectors
	processedSelectors map[string]*Selector

	// stores results of selector processing
	selections map[string][]nodes

	// cache of parsed numeric values
	numCache map[string]uint64

	hrwSeed []byte

	// hrw.Hash of hrwSeed
	hrwSeedHash hrw.Hashable

	// weightFunc is a weighting function for determining node priority
	// which combines low price and high performance
	weightFunc weightFunc

	// container backup factor
	cbf uint32
}

// Various validation errors.
var (
	// ErrNotEnoughNodes is returned when a placement policy cannot be satisfied
	// due to low numbers of nodes in selected network map.
	ErrNotEnoughNodes = errors.New("not enough nodes to SELECT from")

	errInvalidFilterName = errors.New("filter name is invalid")
	errInvalidNumber     = errors.New("invalid number")
	errInvalidFilterOp   = errors.New("invalid filter operation")
	errFilterNotFound    = errors.New("filter not found")
	errNonEmptyFilters   = errors.New("simple filter contains sub-filters")
	errUnnamedTopFilter  = errors.New("unnamed top-level filter")
)

// newContext returns initialized context.
func newContext(nm NetMap) *context {
	return &context{
		netMap:             nm,
		processedFilters:   make(map[string]*Filter),
		processedSelectors: make(map[string]*Selector),
		selections:         make(map[string][]nodes),

		numCache:   make(map[string]uint64),
		weightFunc: defaultWeightFunc(nm.nodes),
	}
}

func (c *context) setPivot(pivot []byte) {
	if len(pivot) != 0 {
		c.hrwSeed = pivot
		c.hrwSeedHash = hrw.WrapBytes(pivot)
	}
}

func (c *context) setCBF(cbf uint32) {
	if cbf == 0 {
		c.cbf = defaultContainerBackupFactor
	} else {
		c.cbf = cbf
	}
}

func defaultWeightFunc(ns nodes) weightFunc {
	meanA := newMeanAgg()
	minA := newMinAgg()

	for i := range ns {
		meanA.Add(float64(ns[i].capacity()))
		minA.Add(float64(ns[i].Price()))
	}

	return newWeightFunc(
		newSigmoidNorm(meanA.Compute()),
		newReverseMinNorm(minA.Compute()))
}
