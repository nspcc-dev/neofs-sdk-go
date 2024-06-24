package neofscryptotest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	v := neofscryptotest.Signature()
	require.NotEqual(t, v, neofscryptotest.Signature())

	var m refs.Signature
	v.WriteToV2(&m)
	var v2 neofscrypto.Signature
	require.NoError(t, v2.ReadFromV2(&m))
	require.Equal(t, v, v2)
}
