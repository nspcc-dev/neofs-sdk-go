/*
Package accounting provides primitives to perform accounting operations in NeoFS.

Decimal type provides functionality to process user balances. For example, when
working with Fixed8 balance precision:
	var dec accounting.Decimal
	dec.SetValue(val)
	dec.SetPrecision(8)

Instances can be also used to process NeoFS API V2 protocol messages
(see neo.fs.v2.accounting package in https://github.com/nspcc-dev/neofs-api).

On client side:
	import "github.com/nspcc-dev/neofs-api-go/v2/accounting"

	var msg accounting.Decimal
	dec.WriteToMessageV2(&msg)

	// send msg

On server side:
	// recv msg

	var dec accounting.Decimal
	dec.ReadFromMessageV2(msg)

	// process dec

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

*/
package accounting
