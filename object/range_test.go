package object_test

import (
	"testing"

	objectV2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	objecttestv2 "github.com/nspcc-dev/neofs-api-go/v2/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func TestRange_SetOffset(t *testing.T) {
	var r object.Range

	off := uint64(13)
	r.SetOffset(off)

	require.Equal(t, off, r.Offset())
}

func TestRange_SetLength(t *testing.T) {
	var r object.Range

	ln := uint64(7)
	r.SetLength(ln)

	require.Equal(t, ln, r.Length())
}

func TestNewRangeFromV2(t *testing.T) {
	t.Run("from v2", func(t *testing.T) {
		var (
			x  object.Range
			v2 = objecttestv2.GenerateRange(false)
		)

		l := v2.GetLength()
		o := v2.GetOffset()

		x.ReadFromV2(*v2)

		require.Equal(t, l, x.Length())
		require.Equal(t, o, x.Offset())
	})
}

func TestRange_ToV2(t *testing.T) {
	t.Run("to v2", func(t *testing.T) {
		var (
			x  = objecttest.Range()
			v2 objectV2.Range
		)

		x.WriteToV2(&v2)

		require.Equal(t, x.Offset(), v2.GetOffset())
		require.Equal(t, x.Length(), v2.GetLength())
	})
}

func TestNewRange(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var r object.Range

		// check initial values
		require.Zero(t, r.Length())
		require.Zero(t, r.Offset())

		// convert to v2 message
		var rV2 objectV2.Range

		r.WriteToV2(&rV2)

		require.Zero(t, rV2.GetLength())
		require.Zero(t, rV2.GetOffset())
	})
}
