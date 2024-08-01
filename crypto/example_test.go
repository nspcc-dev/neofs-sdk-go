package neofscrypto_test

import (
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
