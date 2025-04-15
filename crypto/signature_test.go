package neofscrypto_test

import (
	"testing"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const anyUnsupportedScheme = -1

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

	m := clientSig.ProtoMessage()

	require.Equal(t, refs.SignatureScheme(scheme), m.GetScheme())
	require.Equal(t, bPubKey, m.GetKey())
	require.Equal(t, clientSig.Value(), m.GetSign())

	// m transmitted to server over the network

	var serverSig neofscrypto.Signature

	err = serverSig.FromProtoMessage(m)
	require.NoError(t, err)

	testSig(serverSig)

	// break the message in different ways
	for i, breakSig := range []func(*refs.Signature){
		func(sigV2 *refs.Signature) { sigV2.Scheme = refs.SignatureScheme(anyUnsupportedScheme) },
	} {
		m := proto.Clone(m).(*refs.Signature)
		breakSig(m)

		err = serverSig.FromProtoMessage(m)
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

	m := sig.ProtoMessage()

	var sig2 neofscrypto.Signature

	err := sig2.FromProtoMessage(m)
	require.NoError(t, err)

	checkFields(sig2)
}

func TestNewN3Signature(t *testing.T) {
	invocScript := []byte("foo")
	verifScript := []byte("bar")

	sig := neofscrypto.NewN3Signature(invocScript, verifScript)

	require.Equal(t, neofscrypto.N3, sig.Scheme())
	require.Equal(t, verifScript, sig.PublicKeyBytes())
	require.Equal(t, invocScript, sig.Value())
}
