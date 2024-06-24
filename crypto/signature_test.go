package neofscrypto_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const anyUnsupportedScheme = neofscrypto.ECDSA_WALLETCONNECT + 1

func TestSignatureLifecycle(t *testing.T) {
	data := []byte("Hello, world!")
	signer := neofscryptotest.RandomSigner()
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

	var apiSig refs.Signature
	clientSig.WriteToV2(&apiSig)

	require.Equal(t, refs.SignatureScheme(scheme), apiSig.GetScheme())
	require.Equal(t, bPubKey, apiSig.GetKey())
	require.Equal(t, clientSig.Value(), apiSig.GetSign())

	// apiSig transmitted to server over the network

	var serverSig neofscrypto.Signature

	err = serverSig.ReadFromV2(&apiSig)
	require.NoError(t, err)

	testSig(serverSig)

	// break the message in different ways
	for i, breakSig := range []func(*refs.Signature){
		func(apiSig *refs.Signature) { apiSig.Scheme = refs.SignatureScheme(anyUnsupportedScheme) },
		func(apiSig *refs.Signature) {
			key := apiSig.GetKey()
			apiSig.Key = key[:len(key)-1]
		},
		func(apiSig *refs.Signature) { apiSig.Key = append(apiSig.GetKey(), 1) },
		func(apiSig *refs.Signature) { apiSig.Sign = nil },
	} {
		sigV2Cp := proto.Clone(&apiSig).(*refs.Signature)
		breakSig(sigV2Cp)

		err = serverSig.ReadFromV2(sigV2Cp)
		require.Errorf(t, err, "break func #%d", i)
	}
}

func TestNewSignature(t *testing.T) {
	signer := neofscryptotest.RandomSigner()
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

	err := sig2.ReadFromV2(&sigMsg)
	require.NoError(t, err)

	checkFields(sig2)
}
