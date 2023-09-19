/*
Package tests provides special help functions for testing NeoFS API and its environment.

All functions accepting `t *testing.T` that emphasize there are only for tests purposes.
*/
package test

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

// RandomSigner return neofscrypto.Signer ONLY for TESTs purposes.
// It may be used like helper to get new neofscrypto.Signer if you need it in yours tests.
func RandomSigner(tb testing.TB) neofscrypto.Signer {
	p, err := keys.NewPrivateKey()
	require.NoError(tb, err)

	return neofsecdsa.Signer(p.PrivateKey)
}

// RandomSignerRFC6979 return [user.Signer] ONLY for TESTs purposes.
// It may be used like helper to get new [user.Signer] if you need it in yours tests.
func RandomSignerRFC6979(tb testing.TB) user.Signer {
	p, err := keys.NewPrivateKey()
	require.NoError(tb, err)

	return user.NewAutoIDSignerRFC6979(p.PrivateKey)
}

// SignedComponent describes component which can signed and the signature may be verified.
type SignedComponent interface {
	SignedData() []byte
	Sign(neofscrypto.Signer) error
	VerifySignature() bool
}

// SignedComponentUserSigner is the same as [SignedComponent] but uses [user.Signer] instead of [neofscrypto.Signer].
// It helps to cover all cases.
type SignedComponentUserSigner interface {
	SignedData() []byte
	Sign(user.Signer) error
	VerifySignature() bool
}

// SignedDataComponent tests [SignedComponent] for valid data generation by SignedData function.
func SignedDataComponent(tb testing.TB, signer neofscrypto.Signer, cmp SignedComponent) {
	data := cmp.SignedData()

	sig, err := signer.Sign(data)
	require.NoError(tb, err)

	static := neofscrypto.NewStaticSigner(signer.Scheme(), sig, signer.Public())

	err = cmp.Sign(static)
	require.NoError(tb, err)

	require.True(tb, cmp.VerifySignature())
}

// SignedDataComponentUser tests [SignedComponentUserSigner] for valid data generation by SignedData function.
func SignedDataComponentUser(tb testing.TB, signer user.Signer, cmp SignedComponentUserSigner) {
	data := cmp.SignedData()

	sig, err := signer.Sign(data)
	require.NoError(tb, err)

	static := neofscrypto.NewStaticSigner(signer.Scheme(), sig, signer.Public())

	err = cmp.Sign(user.NewSigner(static, signer.UserID()))
	require.NoError(tb, err)

	require.True(tb, cmp.VerifySignature())
}

// RandomPublicKey returns random [neofscrypto.PublicKey].
func RandomPublicKey() neofscrypto.PublicKey {
	p, err := keys.NewPrivateKey()
	if err != nil {
		panic(fmt.Errorf("randomize private key: %w", err))
	}

	return neofsecdsa.Signer(p.PrivateKey).Public()
}
