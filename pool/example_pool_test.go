package pool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
)

func ExampleNew_easiestWay() {
	// Signer generation, like example.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := neofsecdsa.SignerRFC6979(*pk)

	pool, _ := New(NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}), signer, DefaultOptions())
	_ = pool

	// ...
}

func ExampleNew_adjustingParameters() {
	// Signer generation, like example.
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	signer := neofsecdsa.SignerRFC6979(*pk)

	opts := DefaultOptions()
	opts.SetErrorThreshold(10)

	pool, _ := New(NewFlatNodeParams([]string{"grpc://localhost:8080", "grpcs://localhost:8081"}), signer, opts)
	_ = pool

	// ...
}