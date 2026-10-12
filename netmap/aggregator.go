package netmap

import (
	"slices"
)

type (
	// aggregator can calculate some value across all netmap
	// such as median, minimum or maximum.
	aggregator interface {
		Add(float64)
		Compute() float64
	}

	meanIQRAgg struct {
		k   float64
		arr []float64
	}
)

var (
	_ aggregator = (*meanIQRAgg)(nil)
)

// newMeanIQRAgg returns an aggregator which
// computes mean value of values from IQR interval.
func newMeanIQRAgg() aggregator {
	return new(meanIQRAgg)
}

func (a *meanIQRAgg) Add(n float64) {
	a.arr = append(a.arr, n)
}

func (a *meanIQRAgg) Compute() float64 {
	l := len(a.arr)
	if l == 0 {
		return 0
	}

	slices.Sort(a.arr)

	var minV, maxV float64

	const minLn = 4

	if l < minLn {
		minV, maxV = a.arr[0], a.arr[l-1]
	} else {
		start, end := l/minLn, l*3/minLn-1
		iqr := a.k * (a.arr[end] - a.arr[start])
		minV, maxV = a.arr[start]-iqr, a.arr[end]+iqr
	}

	count := 0
	sum := float64(0)

	for _, e := range a.arr {
		if e >= minV && e <= maxV {
			sum += e
			count++
		}
	}

	return sum / float64(count)
}
