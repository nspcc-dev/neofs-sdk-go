package accountingtest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	accountingtest "github.com/nspcc-dev/neofs-sdk-go/accounting/test"
	apiaccounting "github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	"github.com/stretchr/testify/require"
)

func TestDecimal(t *testing.T) {
	d := accountingtest.Decimal()
	require.NotEqual(t, d, accountingtest.Decimal())

	var d2 accounting.Decimal
	require.NoError(t, d2.Unmarshal(d.Marshal()))
	require.Equal(t, d, d2)

	var m apiaccounting.Decimal
	d.WriteToV2(&m)
	var d3 accounting.Decimal
	require.NoError(t, d3.ReadFromV2(&m))
	require.Equal(t, d, d3)
}
