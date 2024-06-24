/*
Package neofscryptotest provides special help functions for testing NeoFS API and its environment.
*/
package neofscryptotest

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
)

// ECDSAPrivateKey returns random ECDSA private key.
func ECDSAPrivateKey() ecdsa.PrivateKey {
	p, err := keys.NewPrivateKey()
	if err != nil {
		panic(fmt.Errorf("unexpected private key randomizaiton failure: %w", err))
	}
	return p.PrivateKey
}

// TODO: drop Random, other functions don't write it
// RandomSigner returns random signer.
func RandomSigner() neofscrypto.Signer {
	return neofsecdsa.Signer(ECDSAPrivateKey())
}

// RandomSignerRFC6979 returns random signer with
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme.
func RandomSignerRFC6979() neofscrypto.Signer {
	return neofsecdsa.SignerRFC6979(ECDSAPrivateKey())
}

type failedSigner struct {
	neofscrypto.Signer
}

func (x failedSigner) Sign([]byte) ([]byte, error) { return nil, errors.New("failed to sign") }

// FailSigner wraps s to always return error from Sign method.
func FailSigner(s neofscrypto.Signer) neofscrypto.Signer {
	return failedSigner{s}
}

// Signature returns random neofscrypto.Signature.
func Signature() neofscrypto.Signature {
	sig := make([]byte, 64)
	rand.Read(sig)
	return neofscrypto.NewSignature(neofscrypto.Scheme(rand.Uint32()%3), RandomSigner().Public(), sig)
}
