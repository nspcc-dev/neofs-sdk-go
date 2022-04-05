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
	rand.Read(data)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)

	var s neofscrypto.Signature
	var m refs.Signature

	for _, f := range []func() neofscrypto.Signer{
		func() neofscrypto.Signer {
			var key neofsecdsa.Signer
			key.SetKey(k.PrivateKey)

			return &key
		},
		func() neofscrypto.Signer {
			var key neofsecdsa.Signer
			key.SetKey(k.PrivateKey)
			key.MakeDeterministic()

			return &key
		},
	} {
		signer := f()

		err := s.Calculate(signer, data)
		require.NoError(t, err)

		s.WriteToV2(&m)

		s.ReadFromV2(m)

		valid := s.Verify(data)
		require.True(t, valid)
	}
}
