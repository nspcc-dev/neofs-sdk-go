package accounting_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	apiaccounting "github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	"github.com/stretchr/testify/require"
)

func TestDecimal_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var d accounting.Decimal
		msg := []byte("definitely_not_protobuf")
		err := d.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
}

func testDecimalField[Type uint32 | int64](t *testing.T, get func(accounting.Decimal) Type, set func(*accounting.Decimal, Type),
	getAPI func(info *apiaccounting.Decimal) Type) {
	var d accounting.Decimal

	require.Zero(t, get(d))

	const val = 13
	set(&d, val)
	require.EqualValues(t, val, get(d))

	const valOther = 42
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
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, get(dst))

			set(&src, val)

			src.WriteToV2(&msg)
			require.EqualValues(t, val, getAPI(&msg))
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
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
