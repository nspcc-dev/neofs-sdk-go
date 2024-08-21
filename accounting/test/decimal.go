package accountingtest

import (
	"math/rand/v2"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
)

// Decimal returns random accounting.Decimal.
func Decimal() accounting.Decimal {
	var d accounting.Decimal
	d.SetValue(rand.Int64())
	d.SetPrecision(rand.Uint32())

	return d
}
