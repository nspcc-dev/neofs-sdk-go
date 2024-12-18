package accounting_test

import (
	"math/rand/v2"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	protoaccounting "github.com/nspcc-dev/neofs-sdk-go/proto/accounting"
	"github.com/stretchr/testify/require"
)

const (
	anyValidValue     = 323508430
	anyValidPrecision = 895494320
)

// set by init.
var validDecimal accounting.Decimal

func init() {
	validDecimal.SetValue(anyValidValue)
	validDecimal.SetPrecision(anyValidPrecision)
}

var validBinDecimal = []byte{8, 206, 177, 161, 154, 1, 16, 176, 209, 128, 171, 3}

func testDecimalField[T uint32 | int64](
	t *testing.T,
	get func(accounting.Decimal) T,
	set func(*accounting.Decimal, T),
) {
	var d accounting.Decimal
	require.Zero(t, get(d))

	val := T(rand.Uint64())
	set(&d, val)
	require.EqualValues(t, val, get(d))
	valOther := val + 1
	set(&d, valOther)
	require.EqualValues(t, valOther, get(d))
}

func TestDecimal_SetValue(t *testing.T) {
	testDecimalField(t, accounting.Decimal.Value, (*accounting.Decimal).SetValue)
}

func TestDecimal_SetPrecision(t *testing.T) {
	testDecimalField(t, accounting.Decimal.Precision, (*accounting.Decimal).SetPrecision)
}

func TestDecimal_FromProtoMessage(t *testing.T) {
	var m protoaccounting.Decimal
	m.Value = anyValidValue
	m.Precision = anyValidPrecision

	var val accounting.Decimal
	require.NoError(t, val.FromProtoMessage(&m))
	require.EqualValues(t, anyValidValue, val.Value())
	require.EqualValues(t, anyValidPrecision, val.Precision())
}

func TestDecimal_ProtoMessage(t *testing.T) {
	var val accounting.Decimal

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetValue())
	require.Zero(t, m.GetPrecision())

	// filled
	val.SetValue(anyValidValue)
	val.SetPrecision(anyValidPrecision)

	m = val.ProtoMessage()
	require.EqualValues(t, anyValidValue, m.GetValue())
	require.EqualValues(t, anyValidPrecision, m.GetPrecision())
}

func TestToken_Marshal(t *testing.T) {
	require.Equal(t, validBinDecimal, validDecimal.Marshal())
}

func TestToken_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(accounting.Decimal).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
	})

	var val accounting.Decimal
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Zero(t, val.Value())
	require.Zero(t, val.Precision())

	// filled
	err := val.Unmarshal(validBinDecimal)
	require.NoError(t, err)
	require.Equal(t, validDecimal, val)
}
