package object

import (
	"crypto/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/stretchr/testify/require"
)

func TestVerificationFields(t *testing.T) {
	obj := New()

	payload := make([]byte, 10)
	_, _ = rand.Read(payload)

	obj.SetPayload(payload)
	obj.SetPayloadSize(uint64(len(payload)))

	p, err := keys.NewPrivateKey()
	require.NoError(t, err)
	require.NoError(t, SetVerificationFields(p.PrivateKey, obj))

	require.NoError(t, CheckVerificationFields(obj))

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

		require.Error(t, CheckVerificationFields(obj))

		item.restore()

		require.NoError(t, CheckVerificationFields(obj))
	}
}
