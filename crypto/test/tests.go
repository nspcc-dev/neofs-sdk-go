/*
Package tests provides special help functions for testing NeoFS API and its environment.

All functions accepting `t *testing.T` that emphasize there are only for tests purposes.
*/
package test

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/stretchr/testify/require"
)

// RandomSigner return neofscrypto.Signer ONLY for TESTs purposes.
// It may be used like helper to get new neofscrypto.Signer if you need it in yours tests.
func RandomSigner(tb testing.TB) neofscrypto.Signer {
	p, err := keys.NewPrivateKey()
	require.NoError(tb, err)

	return neofsecdsa.Signer(p.PrivateKey)
}

// RandomSignerRFC6979 return neofscrypto.Signer ONLY for TESTs purposes.
// It may be used like helper to get new neofscrypto.Signer if you need it in yours tests.
func RandomSignerRFC6979(tb testing.TB) neofscrypto.Signer {
	p, err := keys.NewPrivateKey()
	require.NoError(tb, err)

	return neofsecdsa.SignerRFC6979(p.PrivateKey)
}
