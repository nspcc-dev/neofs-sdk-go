package pool

import "math/rand"

// sampler implements weighted random number generation using Vose's Alias
// Method (https://www.keithschwarz.com/darts-dice-coins/).
type sampler struct {
	randomGenerator *rand.Rand
	probabilities   []float64
	alias           []int
}

// newSampler creates new sampler with a given set of probabilities using
// given source of randomness. Created sampler will produce numbers from
// 0 to len(probabilities).
func newSampler(probabilities []float64, source rand.Source) *sampler {
	sampler := &sampler{}
	var (
		small workList
		large workList
	)
	n := len(probabilities)
	sampler.randomGenerator = rand.New(source)
	sampler.probabilities = make([]float64, n)
	sampler.alias = make([]int, n)
	// Compute scaled probabilities.
	p := make([]float64, n)
	for i := 0; i < n; i++ {
		p[i] = probabilities[i] * float64(n)
	}
	for i, pi := range p {
		if pi < 1 {
			small.add(i)
		} else {
			large.add(i)
		}
	}
	for len(small) > 0 && len(large) > 0 {
		l, g := small.remove(), large.remove()
		sampler.probabilities[l] = p[l]
		sampler.alias[l] = g
		p[g] = p[g] + p[l] - 1
		if p[g] < 1 {
			small.add(g)
		} else {
			large.add(g)
		}
	}
	for len(large) > 0 {
		g := large.remove()
		sampler.probabilities[g] = 1
	}
	for len(small) > 0 {
		l := small.remove()
		sampler.probabilities[l] = 1
	}
	return sampler
}

// Next returns the next (not so) random number from sampler.
func (g *sampler) Next() int {
	n := len(g.alias)
	i := g.randomGenerator.Intn(n)
	if g.randomGenerator.Float64() < g.probabilities[i] {
		return i
	}
	return g.alias[i]
}

type workList []int

func (wl *workList) add(e int) {
	*wl = append(*wl, e)
}

func (wl *workList) remove() int {
	l := len(*wl) - 1
	n := (*wl)[l]
	*wl = (*wl)[:l]
	return n
}
