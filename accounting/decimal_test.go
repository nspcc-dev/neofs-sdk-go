package accounting_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	test "github.com/nspcc-dev/neofs-sdk-go/accounting/test"
	"github.com/stretchr/testify/require"
)

func TestDecimal(t *testing.T) {
	const v, p = 4, 2

	d := accounting.NewDecimal()
	d.SetValue(v)
	d.SetPrecision(p)

	require.EqualValues(t, v, d.Value())
	require.EqualValues(t, p, d.Precision())
}

func TestDecimalEncoding(t *testing.T) {
	d := test.GenerateDecimal()

	t.Run("binary", func(t *testing.T) {
		data, err := d.Marshal()
		require.NoError(t, err)

		d2 := accounting.NewDecimal()
		require.NoError(t, d2.Unmarshal(data))

		require.Equal(t, d, d2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := d.MarshalJSON()
		require.NoError(t, err)

		d2 := accounting.NewDecimal()
		require.NoError(t, d2.UnmarshalJSON(data))

		require.Equal(t, d, d2)
	})
}
