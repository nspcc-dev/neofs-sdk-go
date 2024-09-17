package neofscrypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
)

func ExampleSignature_Calculate() {
	// instantiate signer
	k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	// get data to be signed
	data := []byte("Hello, world!")
	var sig neofscrypto.Signature

	for _, signer := range []neofscrypto.Signer{
		neofsecdsa.Signer(*k),
		neofsecdsa.SignerRFC6979(*k),
		neofsecdsa.SignerWalletConnect(*k),
	} {
		// calculate signature
		_ = sig.Calculate(signer, data)
		fmt.Printf("scheme: %d; public key: %x; signature: %x\n",
			sig.Scheme(), sig.PublicKeyBytes(), sig.Value())
		// attach the signature to the request or data structure
	}
}

func ExampleSignature_Verify() {
	// get data to be verified
	data := []byte("Hello, world!")

	for _, tc := range []struct {
		scheme neofscrypto.Scheme
		pub    string
		val    string
	}{
		{
			scheme: neofscrypto.ECDSA_SHA512,
			pub:    "034f10e101921787b4bd69098350cdfea4c5493f6b1d0f8829276b2a460988bd6f",
			val:    "0401cf064093860d4996a9e22df3638c6bcebbd326e82f01cf9c37942a652a79a4fea18ec3777fda6c999162a9414180ce8e11aad97ede7a5da930faebf195aaa8",
		},
		{
			scheme: neofscrypto.ECDSA_DETERMINISTIC_SHA256,
			pub:    "034f10e101921787b4bd69098350cdfea4c5493f6b1d0f8829276b2a460988bd6f",
			val:    "274e6a5ebae221f2c57ea9b9429c6410f31c17580a1d579088ee1839539a044f9cb89886552ed8b21b3c618ee223a061e9957518c327c81397fed074bd31d796",
		},
		{
			scheme: neofscrypto.ECDSA_WALLETCONNECT,
			pub:    "034f10e101921787b4bd69098350cdfea4c5493f6b1d0f8829276b2a460988bd6f",
			val:    "b41afc64b7a77b15ce269f9a52e39131dbe06e827cd469010ee575b7d8e4df6a627c7c79673a292f32066640b9c9b487a5aee74c45dcef8e7a0da4a2a20e9bae018647a3a43e253c26749e964593e542",
		},
	} {
		// get signature
		pub, _ := hex.DecodeString(tc.pub)
		val, _ := hex.DecodeString(tc.val)
		sig := neofscrypto.NewSignatureFromRawKey(tc.scheme, pub, val)
		// verify the signature
		valid := sig.Verify(data)
		fmt.Printf("valid: %t, scheme: %d; public key: %s; signature: %s\n",
			valid, tc.scheme, tc.pub, tc.val)
	}
}
