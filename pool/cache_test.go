package pool

import (
	"testing"
	"time"

	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/session/v2"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_GetUnmodifiedToken(t *testing.T) {
	const key = "Foo"
	target := sessiontest.Token()
	target.SetExp(uint64(time.Now().Unix() + 1000))

	check := func(t *testing.T, tok session.Token, extra string) {
		require.False(t, tok.VerifySignature(), extra)
	}

	cache, err := newCache(defaultSessionCacheSize)
	require.NoError(t, err)

	cache.Put(key, target)
	value, ok := cache.Get(key)
	require.True(t, ok)
	check(t, value, "before sign")

	err = value.Sign(usertest.User())
	require.NoError(t, err)

	value, ok = cache.Get(key)
	require.True(t, ok)
	check(t, value, "after sign")
}
