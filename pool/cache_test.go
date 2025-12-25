package pool

import (
	"testing"
	"time"

	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_GetUnmodifiedToken(t *testing.T) {
	const key = "Foo"
	target1 := sessiontest.Token()
	target1.SetExp(time.Now().Add(1000 * time.Second))
	target2 := sessiontest.Object()

	check := func(t *testing.T, tok interface{ VerifySignature() bool }, extra string) {
		require.False(t, tok.VerifySignature(), extra)
	}

	cache, err := newCache(defaultSessionCacheSize)
	require.NoError(t, err)

	cache.Put(key, target2)
	value, ok := cache.Get(key)
	require.True(t, ok)
	check(t, value, "before sign")

	valueV2, ok := cache.GetV2(key)
	require.False(t, ok)
	require.Empty(t, valueV2)

	require.NoError(t, value.Sign(usertest.User()))

	value, ok = cache.Get(key)
	require.True(t, ok)
	check(t, value, "after sign")

	cache.PutV2(key, target1)
	valueV2, ok = cache.GetV2(key)
	require.True(t, ok)
	check(t, valueV2, "before sign v2")

	value, ok = cache.Get(key)
	require.False(t, ok)
	require.Empty(t, value)

	require.NoError(t, valueV2.Sign(usertest.User()))

	valueV2, ok = cache.GetV2(key)
	require.True(t, ok)
	check(t, valueV2, "after sign v2")
}
