/*
Package neofscryptotest provides helper functions to test code importing [neofscrypto].
*/
package neofscryptotest

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/stretchr/testify/require"
)

// SignedComponent contains data to be signed and its signature to be verified.
type SignedComponent interface {
	SignedData() []byte
	Sign(neofscrypto.Signer) error
	VerifySignature() bool
}

// TestSignedData tests signing and verification of
// [SignedComponent.SignedData].
func TestSignedData(tb testing.TB, signer neofscrypto.Signer, cmp SignedComponent) {
	data := cmp.SignedData()

	sig, err := signer.Sign(data)
	require.NoError(tb, err)

	static := neofscrypto.NewStaticSigner(signer.Scheme(), sig, signer.Public())

	err = cmp.Sign(static)
	require.NoError(tb, err)

	require.True(tb, cmp.VerifySignature())
}

type failedSigner struct {
	neofscrypto.Signer
}

func (x failedSigner) Sign([]byte) ([]byte, error) { return nil, errors.New("[test] failed to sign") }

// FailSigner wraps s to always return error from Sign method.
func FailSigner(s neofscrypto.Signer) neofscrypto.Signer {
	return failedSigner{s}
}

// Signature returns random neofscrypto.Signature.
func Signature() neofscrypto.Signature {
	sig := make([]byte, 1+rand.Intn(128))
	//nolint:staticcheck // cryptorandom is not required for testing
	rand.Read(sig)
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
