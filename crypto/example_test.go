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

	// ...
}

// Signature can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.refs package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleSignature_WriteToV2() {
	// import "github.com/nspcc-dev/neofs-api-go/v2/refs"

	var sig neofscrypto.Signature
	var msg refs.Signature

	sig.WriteToV2(&msg)

	// send msg
}

// Signature can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.refs package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleSignature_ReadFromV2() {
	// import "github.com/nspcc-dev/neofs-api-go/v2/refs"

	// recv msg

	var msg refs.Signature
	var sig neofscrypto.Signature

	_ = sig.ReadFromV2(msg)

	// process sig
}
