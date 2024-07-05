package accounting_test

import (
	"math/rand"
	"testing"

	apiaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/stretchr/testify/require"
)

func testDecimalField[T uint32 | int64](
	t *testing.T,
	get func(accounting.Decimal) T,
	set func(*accounting.Decimal, T),
	getAPI func(*apiaccounting.Decimal) T,
) {
	var d accounting.Decimal
	require.Zero(t, get(d))

	val := T(rand.Uint64())
	set(&d, val)
	require.EqualValues(t, val, get(d))
	valOther := val + 1
	set(&d, valOther)
	require.EqualValues(t, valOther, get(d))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst accounting.Decimal

			set(&dst, val)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Zero(t, get(dst))

			set(&src, val)
			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.EqualValues(t, val, get(dst))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst accounting.Decimal
			var msg apiaccounting.Decimal

			set(&dst, val)
			src.WriteToV2(&msg)
			require.Zero(t, getAPI(&msg))
			require.NoError(t, dst.ReadFromV2(msg))
			require.Zero(t, get(dst))

			set(&src, val)
			src.WriteToV2(&msg)
			require.EqualValues(t, val, getAPI(&msg))
			require.NoError(t, dst.ReadFromV2(msg))
			require.EqualValues(t, val, get(dst))
		})
	})
}

func TestDecimal_SetValue(t *testing.T) {
	testDecimalField(t, accounting.Decimal.Value, (*accounting.Decimal).SetValue, (*apiaccounting.Decimal).GetValue)
}

func TestDecimal_SetPrecision(t *testing.T) {
	testDecimalField(t, accounting.Decimal.Precision, (*accounting.Decimal).SetPrecision, (*apiaccounting.Decimal).GetPrecision)
}
