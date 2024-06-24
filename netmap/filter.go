package netmap

import (
	"fmt"
	"strconv"
)

// mainFilterName is a name of the filter
// which points to the whole netmap.
const mainFilterName = "*"

// processFilters processes filters and returns error is any of them is invalid.
func (c *context) processFilters(p PlacementPolicy) error {
	for i := range p.filters {
		if err := c.processFilter(p.filters[i], true); err != nil {
			return fmt.Errorf("process filter #%d (%s): %w", i, p.filters[i].Name(), err)
		}
	}

	return nil
}

func (c *context) processFilter(f Filter, top bool) error {
	fName := f.Name()
	if fName == mainFilterName {
		return fmt.Errorf("%w: '%s' is reserved", errInvalidFilterName, mainFilterName)
	}

	if top && fName == "" {
		return errUnnamedTopFilter
	}

	if !top && fName != "" && c.processedFilters[fName] == nil {
		return errFilterNotFound
	}

	inner := f.SubFilters()

	switch op := f.Op(); op {
	case FilterOpAND, FilterOpOR:
		for i := range inner {
			if err := c.processFilter(inner[i], false); err != nil {
				return fmt.Errorf("process inner filter #%d: %w", i, err)
			}
		}
	default:
		if len(inner) != 0 {
			return errNonEmptyFilters
		} else if !top && fName != "" { // named reference
			return nil
		}

		switch op {
		case FilterOpEQ, FilterOpNE:
		case FilterOpGT, FilterOpGE, FilterOpLT, FilterOpLE:
			val := f.Value()
			n, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("%w: '%s'", errInvalidNumber, f.Value())
			}

			c.numCache[val] = n
		default:
			return fmt.Errorf("%w: %s", errInvalidFilterOp, op)
		}
	}

	if top {
		c.processedFilters[fName] = &f
	}

	return nil
}

// match matches f against b. It returns no errors because
// filter should have been parsed during context creation
// and missing node properties are considered as a regular fail.
func (c *context) match(f *Filter, b NodeInfo) bool {
	switch f.Op() {
	case FilterOpAND, FilterOpOR:
		inner := f.SubFilters()
		for i := range inner {
			fSub := &inner[i]
			if name := inner[i].Name(); name != "" {
				fSub = c.processedFilters[name]
			}

			ok := c.match(fSub, b)
			if ok == (f.Op() == FilterOpOR) {
				return ok
			}
		}

		return f.Op() == FilterOpAND
	default:
		return c.matchKeyValue(f, b)
	}
}

func (c *context) matchKeyValue(f *Filter, b NodeInfo) bool {
	switch op := f.Op(); op {
	case FilterOpEQ:
		return b.Attribute(f.Key()) == f.Value()
	case FilterOpNE:
		return b.Attribute(f.Key()) != f.Value()
	default:
		var attr uint64

		switch f.Key() {
		case attrPrice:
			attr = b.Price()
		case attrCapacity:
			attr = b.Capacity()
		default:
			var err error

			attr, err = strconv.ParseUint(b.Attribute(f.Key()), 10, 64)
			if err != nil {
				// Note: because filters are somewhat independent from nodes attributes,
				// We don't report an error here, and fail filter instead.
				return false
			}
		}

		switch op {
		case FilterOpGT:
			return attr > c.numCache[f.Value()]
		case FilterOpGE:
			return attr >= c.numCache[f.Value()]
		case FilterOpLT:
			return attr < c.numCache[f.Value()]
		case FilterOpLE:
			return attr <= c.numCache[f.Value()]
		default:
			// do nothing and return false
		}
	}
	// will not happen if context was created from f (maybe panic?)
	return false
}
