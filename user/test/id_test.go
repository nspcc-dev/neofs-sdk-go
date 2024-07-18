package usertest_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
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
	require.NoError(t, id2.ReadFromV2(m))
}

func TestNIDs(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, cidtest.IDs(n), n)
}

func TestOtherID(t *testing.T) {
	ids := usertest.IDs(100)
	require.NotContains(t, ids, usertest.OtherID(ids...))
}

func TestFailSigner(t *testing.T) {
	s := usertest.FailSigner(usertest.User())
	_, err := s.Sign(nil)
	require.EqualError(t, err, "[test] failed to sign")
}

func TestUser(t *testing.T) {
	s := usertest.User()
	require.NotEqual(t, s, usertest.User())

	require.Equal(t, s.ID, s.UserID())
	require.Equal(t, s.ID, user.NewFromECDSAPublicKey(s.ECDSAPrivateKey.PublicKey))

	var pub neofsecdsa.PublicKey
	require.NoError(t, pub.Decode(s.PublicKeyBytes))
	require.EqualValues(t, s.ECDSAPrivateKey.PublicKey, pub)

	data := []byte("Hello, world!")

	sig, err := s.Sign(data)
	require.NoError(t, err)
	require.True(t, pub.Verify(data, sig))
	require.Equal(t, neofscrypto.ECDSA_SHA512, s.Scheme())

	sig, err = s.RFC6979.Sign(data)
	require.NoError(t, err)
	require.True(t, (*neofsecdsa.PublicKeyRFC6979)(&pub).Verify(data, sig))
	require.Equal(t, neofscrypto.ECDSA_DETERMINISTIC_SHA256, s.RFC6979.Scheme())

	sig, err = s.WalletConnect.Sign(data)
	require.NoError(t, err)
	require.True(t, (*neofsecdsa.PublicKeyWalletConnect)(&pub).Verify(data, sig))
	require.Equal(t, neofscrypto.ECDSA_WALLETCONNECT, s.WalletConnect.Scheme())
}
