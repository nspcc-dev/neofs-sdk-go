package neofscrypto_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	data := make([]byte, 512)
	//nolint:staticcheck
	rand.Read(data)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)

	var m refs.Signature

	for _, f := range []func() neofscrypto.Signer{
		func() neofscrypto.Signer {
			return neofsecdsa.Signer(k.PrivateKey)
		},
		func() neofscrypto.Signer {
			return neofsecdsa.SignerRFC6979(k.PrivateKey)
		},
		func() neofscrypto.Signer {
			return neofsecdsa.SignerWalletConnect(k.PrivateKey)
		},
	} {
		signer := f()

		s, err := neofscrypto.CalculateDataSignature(signer, data)
		require.NoError(t, err)

		s.WriteToV2(&m)

		require.NoError(t, s.ReadFromV2(m))

		valid := neofscrypto.IsValidDataSignature(s, data)
		require.True(t, valid, "type %T", signer)
	}
}
