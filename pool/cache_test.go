package pool

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_GetUnmodifiedToken(t *testing.T) {
	const key = "Foo"
	target := sessiontest.Token()

	pk, err := keys.NewPrivateKey()
	require.NoError(t, err)

	check := func(t *testing.T, tok *session.Token, extra string) {
		require.False(t, tok.VerifySignature(), extra)
		require.Nil(t, tok.Context(), extra)
	}

	cache, err := newCache()
	require.NoError(t, err)

	cache.Put(key, target)
	value := cache.Get(key)
	check(t, value, "before sign")

	err = value.Sign(&pk.PrivateKey)
	require.NoError(t, err)

	octx := sessiontest.ObjectContext()
	value.SetContext(octx)

	value = cache.Get(key)
	check(t, value, "after sign")
}
