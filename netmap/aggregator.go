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

	// normalizer normalizes weight.
	normalizer interface {
		Normalize(w float64) float64
	}

	meanAgg struct {
		mean  float64
		count int
	}

	minAgg struct {
		min *float64
	}

	meanIQRAgg struct {
		k   float64
		arr []float64
	}

	reverseMinNorm struct {
		min float64
	}

	sigmoidNorm struct {
		scale float64
	}

	// weightFunc calculates n's weight.
	weightFunc = func(NodeInfo) float64
)

var (
	_ aggregator = (*meanAgg)(nil)
	_ aggregator = (*minAgg)(nil)
	_ aggregator = (*meanIQRAgg)(nil)

	_ normalizer = (*reverseMinNorm)(nil)
	_ normalizer = (*sigmoidNorm)(nil)
)

// newWeightFunc returns weightFunc which multiplies normalized
// capacity and price.
func newWeightFunc(capNorm, priceNorm normalizer) weightFunc {
	return func(n NodeInfo) float64 {
		return capNorm.Normalize(float64(n.capacity())) * priceNorm.Normalize(float64(n.Price()))
	}
}

// newMeanAgg returns an aggregator which
// computes mean value by recalculating it on
// every addition.
func newMeanAgg() aggregator {
	return new(meanAgg)
}

// newMinAgg returns an aggregator which
// computes min value.
func newMinAgg() aggregator {
	return new(minAgg)
}

// newMeanIQRAgg returns an aggregator which
// computes mean value of values from IQR interval.
func newMeanIQRAgg() aggregator {
	return new(meanIQRAgg)
}

// newReverseMinNorm returns a normalizer which
// normalize values in range of 0.0 to 1.0 to a minimum value.
func newReverseMinNorm(min float64) normalizer {
	return &reverseMinNorm{min: min}
}

// newSigmoidNorm returns a normalizer which
// normalize values in range of 0.0 to 1.0 to a scaled sigmoid.
func newSigmoidNorm(scale float64) normalizer {
	return &sigmoidNorm{scale: scale}
}

func (a *meanAgg) Add(n float64) {
	c := a.count + 1
	a.mean = a.mean*(float64(a.count)/float64(c)) + n/float64(c)
	a.count++
}

func (a *meanAgg) Compute() float64 {
	return a.mean
}

func (a *minAgg) Add(n float64) {
	if a.min == nil || n < *a.min {
		a.min = &n
	}
}

func (a *minAgg) Compute() float64 {
	if a.min == nil {
		return 0
	}

	return *a.min
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

func (r *reverseMinNorm) Normalize(w float64) float64 {
	if w == 0 {
		return 0
	}

	return r.min / w
}

func (r *sigmoidNorm) Normalize(w float64) float64 {
	if r.scale == 0 {
		return 0
	}

	x := w / r.scale

	return x / (1 + x)
}
