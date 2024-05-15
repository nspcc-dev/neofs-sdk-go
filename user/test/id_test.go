package usertest_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	id := usertest.ID()
	require.NotEqual(t, id, usertest.ID())

	var m refs.OwnerID
	id.WriteToV2(&m)
	var id2 user.ID
	require.NoError(t, id2.ReadFromV2(&m))
}

func TestChangeID(t *testing.T) {
	id := usertest.ID()
	require.NotEqual(t, id, usertest.ChangeID(id))
}

func TestNIDs(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, cidtest.NIDs(n), n)
}

func TestTwoUsers(t *testing.T) {
	usr1, usr2 := usertest.TwoUsers()
	require.NotEqual(t, usr1.UserID(), usr2.UserID())
	require.NotEqual(t, usr1.PrivateKey, usr2.PrivateKey)
	require.NotEqual(t, usr1.PublicKeyBytes, usr2.PublicKeyBytes)
	require.NotEqual(t, usr1.Signer, usr2.Signer)
	require.NotEqual(t, usr1.SignerRFC6979, usr2.SignerRFC6979)
	require.NotEqual(t, usr1.SignerWalletConnect, usr2.SignerWalletConnect)

	var pubKey1, pubKey2 ecdsa.PublicKey
	pubKey1.Curve = elliptic.P256()
	pubKey1.X, pubKey1.Y = elliptic.UnmarshalCompressed(pubKey1.Curve, usr1.PublicKeyBytes)
	require.NotNil(t, pubKey1.X)
	require.Equal(t, usr1.PrivateKey.PublicKey, pubKey1)
	pubKey2.Curve = elliptic.P256()
	pubKey2.X, pubKey2.Y = elliptic.UnmarshalCompressed(pubKey2.Curve, usr2.PublicKeyBytes)
	require.NotNil(t, pubKey2.X)
	require.Equal(t, usr2.PrivateKey.PublicKey, pubKey2)
	require.Equal(t, usr1.UserID(), user.ResolveFromECDSAPublicKey(pubKey1))
	require.Equal(t, usr1.UserID(), usr1.UserID())
	require.Equal(t, usr1.UserID(), usr1.SignerRFC6979.UserID())
	require.Equal(t, usr1.UserID(), usr1.SignerWalletConnect.UserID())
	require.Equal(t, usr2.UserID(), user.ResolveFromECDSAPublicKey(pubKey2))
	require.Equal(t, usr2.UserID(), usr2.SignerRFC6979.UserID())
	require.Equal(t, usr2.UserID(), usr2.SignerWalletConnect.UserID())
}
