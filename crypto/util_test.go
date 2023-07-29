package neofscrypto_test

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyBytes(t *testing.T) {
	k, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pubKey := neofsecdsa.PublicKey(k.PrivateKey.PublicKey)

	bPubKey := neofscrypto.PublicKeyBytes(&pubKey)

	var restoredPubKey neofsecdsa.PublicKey

	err = restoredPubKey.Decode(bPubKey)
	require.NoError(t, err)

	require.Equal(t, pubKey, restoredPubKey)
}
