package clientutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

// global default signer.
var signerDefault ecdsa.PrivateKey

func init() {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Errorf("generate default private key: %v", err))
	}

	signerDefault = *k
}
