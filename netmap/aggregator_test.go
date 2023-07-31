package netmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMeanAgg(t *testing.T) {
	tt := []struct {
		vals []float64
		res  float64
	}{
		{
			vals: []float64{0, 1, 3, 4, 5},
			res:  2.6,
		},
		{
			vals: []float64{0, 0, 0, 0},
			res:  0,
		},
		{
			vals: []float64{1, 1, 1, 1},
			res:  1,
		},
	}

	for _, test := range tt {
		a := newMeanAgg()
		for _, val := range test.vals {
			a.Add(val)
		}

		require.Equal(t, test.res, a.Compute())
	}
}
