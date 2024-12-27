package neofscryptotest_test

import (
	"testing"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	s := neofscryptotest.Signature()
	require.NotEqual(t, s, neofscryptotest.Signature())

	m := s.ProtoMessage()
	var s2 neofscrypto.Signature
	require.NoError(t, s2.FromProtoMessage(m))
	require.Equal(t, s, s2)
}

func TestECDSAPrivateKey(t *testing.T) {
	require.NotEqual(t, neofscryptotest.ECDSAPrivateKey(), neofscryptotest.ECDSAPrivateKey())
}

func TestFailSigner(t *testing.T) {
	s := neofscryptotest.FailSigner(neofscryptotest.Signer())
	_, err := s.Sign(nil)
	require.EqualError(t, err, "[test] failed to sign")
}

func TestSigner(t *testing.T) {
	s := neofscryptotest.Signer()
	require.NotEqual(t, s, neofscryptotest.Signer())

	require.EqualValues(t, s.Signer, s.ECDSAPrivateKey)
	require.Equal(t, neofscrypto.ECDSA_SHA512, s.Scheme())
	require.EqualValues(t, s.RFC6979, s.ECDSAPrivateKey)
	require.Equal(t, neofscrypto.ECDSA_DETERMINISTIC_SHA256, s.RFC6979.Scheme())
	require.EqualValues(t, s.WalletConnect, s.ECDSAPrivateKey)
	require.Equal(t, neofscrypto.ECDSA_WALLETCONNECT, s.WalletConnect.Scheme())

	var pub neofsecdsa.PublicKey
	require.NoError(t, pub.Decode(s.PublicKeyBytes))
	require.EqualValues(t, s.ECDSAPrivateKey.PublicKey, pub)
}
