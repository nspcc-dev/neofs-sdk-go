package neofscrypto_test

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

func ExampleSignature_Calculate() {
	var signer neofscrypto.Signer
	var data []byte

	// instantiate Signer
	// select data to be signed

	var sig neofscrypto.Signature
	_ = sig.Calculate(signer, data)

	// attach signature to the request
}

// PublicKey allows to verify signatures.
func ExampleSignature_Verify() {
	var sig neofscrypto.Signature

	var data []byte
	sig.Verify(data)
}

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleSignature_marshalling() {
	// import "github.com/nspcc-dev/neofs-api-go/v2/refs"

	// On the client side.

	var sig neofscrypto.Signature
	var msg refs.Signature
	sig.WriteToV2(&msg)
	// *send message*

	// On the server side.

	_ = sig.ReadFromV2(msg)
}
