/*
Package neofscryptotest provides helper functions to test code importing [neofscrypto].
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
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
)

type failedSigner struct {
	neofscrypto.Signer
}

func (x failedSigner) Sign([]byte) ([]byte, error) { return nil, errors.New("[test] failed to sign") }

// FailSigner wraps s to always return error from Sign method.
func FailSigner(s neofscrypto.Signer) neofscrypto.Signer {
	return failedSigner{s}
}

type failedAfterSigner struct {
	successSignature int
	neofscrypto.Signer
}

// FailAfterSigner wraps s and can successfully Sign success number of calls,
// then it will always fail.
func FailAfterSigner(s neofscrypto.Signer, success int) neofscrypto.Signer {
	return &failedAfterSigner{successSignature: success, Signer: s}
}

func (x *failedAfterSigner) Sign(data []byte) ([]byte, error) {
	x.successSignature--
	if x.successSignature < 0 {
		return nil, errors.New("[test] failed to sign")
	}
	return x.Signer.Sign(data)
}

// Signature returns random neofscrypto.Signature.
func Signature() neofscrypto.Signature {
	sig := testutil.RandByteSlice(1 + rand.Intn(128))
	return neofscrypto.NewSignature(neofscrypto.Scheme(rand.Int31()%3), Signer().Public(), sig)
}

// ECDSAPrivateKey returns random ECDSA private key.
func ECDSAPrivateKey() ecdsa.PrivateKey {
	p, err := keys.NewPrivateKey()
	if err != nil {
		panic(fmt.Errorf("unexpected private key randomizaiton failure: %w", err))
	}
	return p.PrivateKey
}

// VariableSigner unites various elements of NeoFS cryptography frequently used
// in testing.
type VariableSigner struct {
	ECDSAPrivateKey ecdsa.PrivateKey
	// Components calculated for ECDSAPrivateKey.
	PublicKeyBytes []byte
	neofscrypto.Signer
	RFC6979       neofscrypto.Signer
	WalletConnect neofscrypto.Signer
}

// Signer returns random VariableSigner.
func Signer() VariableSigner {
	pk := ECDSAPrivateKey()
	return VariableSigner{
		ECDSAPrivateKey: pk,
		Signer:          neofsecdsa.Signer(pk),
		RFC6979:         neofsecdsa.SignerRFC6979(pk),
		WalletConnect:   neofsecdsa.SignerWalletConnect(pk),
		PublicKeyBytes:  neofscrypto.PublicKeyBytes((*neofsecdsa.PublicKey)(&pk.PublicKey)),
	}
}
