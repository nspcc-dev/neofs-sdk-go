package pool

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_GetUnmodifiedToken(t *testing.T) {
	const key = "Foo"
	target := *sessiontest.Object()

	pk, err := keys.NewPrivateKey()
	require.NoError(t, err)

	check := func(t *testing.T, tok session.Object, extra string) {
		require.False(t, tok.VerifySignature(), extra)
	}

	cache, err := newCache()
	require.NoError(t, err)

	cache.Put(key, target)
	value, ok := cache.Get(key)
	require.True(t, ok)
	check(t, value, "before sign")

	err = value.Sign(neofsecdsa.Signer(pk.PrivateKey))
	require.NoError(t, err)

	value, ok = cache.Get(key)
	require.True(t, ok)
	check(t, value, "after sign")
}
