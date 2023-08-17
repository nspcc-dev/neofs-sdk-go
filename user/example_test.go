package user_test

import (
	"github.com/nspcc-dev/neo-go/pkg/util"
	apiGoRefs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// NeoFS user identification is compatible with Neo accounts.
func ExampleID_SetScriptHash() {
	// import "github.com/nspcc-dev/neo-go/pkg/util"
	var id user.ID

	var scriptHash util.Uint160 // user account in NeoFS
	id.SetScriptHash(scriptHash)
}

// ID is compatible with the NeoFS Smart Contract API.
func ExampleID_WalletBytes() {
	var id user.ID
	// ...

	wallet := id.WalletBytes()
	_ = wallet

	// use wallet in call
}

// Encoding mechanisms are used to transfer identifiers on transmitter.
func ExampleID_EncodeToString() {
	var id user.ID
	// ...

	_ = id.EncodeToString() //
}

// Encoding mechanisms are used to transfer identifiers on receiver.
func ExampleID_DecodeString() {
	var id user.ID
	// ...

	var s string
	_ = id.DecodeString(s) // on receiver
}

// Instances can be also used to process NeoFS API protocol messages
// (see neo.fs.v2.refs package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleID_WriteToV2() {
	// import apiGoRefs "github.com/nspcc-dev/neofs-api-go/v2/refs"

	var id user.ID
	var msg apiGoRefs.OwnerID
	id.WriteToV2(&msg)

	// send msg
}

// Instances can be also used to process NeoFS API protocol messages
// (see neo.fs.v2.refs package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleID_ReadFromV2() {
	// import apiGoRefs "github.com/nspcc-dev/neofs-api-go/v2/refs"

	var id user.ID
	var msg apiGoRefs.OwnerID
	_ = id.ReadFromV2(msg)
	// send msg
}
