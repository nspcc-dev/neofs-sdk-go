package neofscrypto_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

const anyUnsupportedScheme = neofscrypto.ECDSA_WALLETCONNECT + 1

func TestSignatureLifecycle(t *testing.T) {
	data := []byte("Hello, world!")
	signer := neofscryptotest.Signer()
	scheme := signer.Scheme()
	pubKey := signer.Public()
	bPubKey := neofscrypto.PublicKeyBytes(pubKey)

	var clientSig neofscrypto.Signature

	err := clientSig.Calculate(signer, data)
	require.NoError(t, err)

	testSig := func(sig neofscrypto.Signature) {
		require.Equal(t, signer.Scheme(), sig.Scheme())
		require.Equal(t, signer.Public(), sig.PublicKey())
		require.Equal(t, bPubKey, sig.PublicKeyBytes())
		require.NotEmpty(t, sig.Value())
		require.True(t, sig.Verify(data))
	}

	testSig(clientSig)

	var sigV2 refs.Signature
	clientSig.WriteToV2(&sigV2)

	require.Equal(t, refs.SignatureScheme(scheme), sigV2.GetScheme())
	require.Equal(t, bPubKey, sigV2.GetKey())
	require.Equal(t, clientSig.Value(), sigV2.GetSign())

	// sigV2 transmitted to server over the network

	var serverSig neofscrypto.Signature

	err = serverSig.ReadFromV2(sigV2)
	require.NoError(t, err)

	testSig(serverSig)

	// break the message in different ways
	for i, breakSig := range []func(*refs.Signature){
		func(sigV2 *refs.Signature) { sigV2.SetScheme(refs.SignatureScheme(anyUnsupportedScheme)) },
		func(sigV2 *refs.Signature) {
			key := sigV2.GetKey()
			sigV2.SetKey(key[:len(key)-1])
		},
		func(sigV2 *refs.Signature) { sigV2.SetKey(append(sigV2.GetKey(), 1)) },
		func(sigV2 *refs.Signature) { sigV2.SetSign(nil) },
	} {
		sigV2Cp := sigV2
		breakSig(&sigV2Cp)

		err = serverSig.ReadFromV2(sigV2Cp)
		require.Errorf(t, err, "break func #%d", i)
	}
}

func TestNewSignature(t *testing.T) {
	signer := neofscryptotest.Signer()
	scheme := signer.Scheme()
	pubKey := signer.Public()
	val := []byte("Hello, world!") // may be any for this test

	sig := neofscrypto.NewSignature(scheme, pubKey, val)

	checkFields := func(sig neofscrypto.Signature) {
		require.Equal(t, scheme, sig.Scheme())
		require.Equal(t, pubKey, sig.PublicKey())
		require.Equal(t, neofscrypto.PublicKeyBytes(pubKey), sig.PublicKeyBytes())
		require.Equal(t, val, sig.Value())
	}

	checkFields(sig)

	var sigMsg refs.Signature
	sig.WriteToV2(&sigMsg)

	var sig2 neofscrypto.Signature

	err := sig2.ReadFromV2(sigMsg)
	require.NoError(t, err)

	checkFields(sig2)
}
