package accounting_test

import (
	apiGoAccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
)

func Example() {
	var val int64

	var dec accounting.Decimal
	dec.SetValue(val)
	dec.SetPrecision(8)

	// Instances can be also used to process NeoFS API V2 protocol messages. See neofs-api-go package.

	// On the client side.

	// import apiGoAccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	var msg apiGoAccounting.Decimal
	dec.WriteToV2(&msg)
	// *send message*

	// On the server side.
	_ = dec.ReadFromV2(msg)
}
