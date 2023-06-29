package object

import (
	"crypto/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

func TestVerificationFields(t *testing.T) {
	obj := New()

	payload := make([]byte, 10)
	_, _ = rand.Read(payload)

	obj.SetPayload(payload)
	obj.SetPayloadSize(uint64(len(payload)))

	require.NoError(t, obj.SetVerificationFields(test.RandomSigner(t)))

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
