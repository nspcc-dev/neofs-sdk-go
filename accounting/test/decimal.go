package accountingtest

import (
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
)

// GenerateDecimal returns random accounting.Decimal.
func GenerateDecimal() *accounting.Decimal {
	d := accounting.NewDecimal()
	d.SetValue(rand.Int63())
	d.SetPrecision(rand.Uint32())

	return d
}
