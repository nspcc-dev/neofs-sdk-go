package user_test

import (
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Encoding mechanisms are used to transfer identifiers on client.
func ExampleID_EncodeToString() {
	var id user.ID
	// ...

	_ = id.EncodeToString()
}

// Encoding mechanisms are used to transfer identifiers on server.
func ExampleID_DecodeString() {
	var id user.ID
	// ...

	var s string
	_ = id.DecodeString(s)
}

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleID_marshalling() {
	// On the client side.

	var id user.ID
	msg := id.ProtoMessage()
	// *send message*

	// On the server side.

	_ = id.FromProtoMessage(msg)
}
