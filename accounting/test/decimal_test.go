package accountingtest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	accountingtest "github.com/nspcc-dev/neofs-sdk-go/accounting/test"
	"github.com/stretchr/testify/require"
)

func TestDecimal(t *testing.T) {
	d := accountingtest.Decimal()
	require.NotEqual(t, d, accountingtest.Decimal())

	m := d.ProtoMessage()
	var d2 accounting.Decimal
	require.NoError(t, d2.FromProtoMessage(m))
	require.Equal(t, d, d2)

	require.NoError(t, new(accounting.Decimal).Unmarshal(d.Marshal()))
}
