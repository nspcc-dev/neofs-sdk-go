package object

import (
	"crypto/rand"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestVerificationFields(t *testing.T) {
	obj := New()

	payload := make([]byte, 10)
	_, _ = rand.Read(payload)

	obj.SetPayload(payload)
	obj.SetPayloadSize(uint64(len(payload)))

	require.NoError(t, obj.SetVerificationFields(neofscryptotest.Signer()))

	require.NoError(t, obj.CheckVerificationFields())

	items := []struct {
		corrupt func()
		restore func()
	}{
		{
			corrupt: func() {
				payload[0]++
			},
			restore: func() {
				payload[0]--
			},
		},
		{
			corrupt: func() {
				obj.SetPayloadSize(obj.PayloadSize() + 1)
			},
			restore: func() {
				obj.SetPayloadSize(obj.PayloadSize() - 1)
			},
		},
		{
			corrupt: func() {
				obj.ToV2().GetObjectID().GetValue()[0]++
			},
			restore: func() {
				obj.ToV2().GetObjectID().GetValue()[0]--
			},
		},
	}

	for _, item := range items {
		item.corrupt()

		require.Error(t, obj.CheckVerificationFields())

		item.restore()

		require.NoError(t, obj.CheckVerificationFields())
	}
}

func TestObject_SignedData(t *testing.T) {
	signer := neofscryptotest.Signer()
	uid := usertest.ID()

	rf := RequiredFields{
		Container: cidtest.ID(),
		Owner:     uid,
	}
	var val Object

	val.InitCreation(rf)

	require.NoError(t, val.SetVerificationFields(signer))

	neofscryptotest.TestSignedData(t, signer, &val)
}
