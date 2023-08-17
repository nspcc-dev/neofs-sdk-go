package accounting_test

import (
	apiGoAccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
)

// Decimal type provides functionality to process user balances. For example, when
// working with Fixed8 balance precision.
func ExampleDecimal_SetValue() {
	var val int64

	var dec accounting.Decimal
	dec.SetValue(val)
	dec.SetPrecision(8)
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.accounting package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleDecimal_WriteToV2() {
	// import apiGoAccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"

	var dec accounting.Decimal
	var msg apiGoAccounting.Decimal
	dec.WriteToV2(&msg)

	// send msg
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.accounting package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleDecimal_ReadFromV2() {
	// import apiGoAccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"

	var dec accounting.Decimal
	var msg apiGoAccounting.Decimal

	_ = dec.ReadFromV2(msg)

	// send msg
}
