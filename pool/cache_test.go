package pool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSessionCache_GetAccessTime(t *testing.T) {
	const key = "Foo"

	cache, err := newCache()
	require.NoError(t, err)

	cache.Put(key, nil)

	t1, ok := cache.GetAccessTime(key)
	require.True(t, ok)

	time.Sleep(10 * time.Millisecond)

	t2, ok := cache.GetAccessTime(key)
	require.True(t, ok)
	require.Equal(t, t1, t2)

	_ = cache.Get(key)
	t3, ok := cache.GetAccessTime(key)
	require.True(t, ok)
	require.NotEqual(t, t1, t3)
}
