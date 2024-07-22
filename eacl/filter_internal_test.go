package eacl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter_CopyTo(t *testing.T) {
	var filter Filter
	filter.value = staticStringer("value")
	filter.from = 1
	filter.matcher = 1
	filter.key = "1"

	var dst Filter
	t.Run("copy", func(t *testing.T) {
		filter.CopyTo(&dst)

		require.Equal(t, filter, dst)
		require.True(t, bytes.Equal(filter.Marshal(), dst.Marshal()))
	})

	t.Run("change", func(t *testing.T) {
		require.Equal(t, filter.value, dst.value)
		require.Equal(t, filter.from, dst.from)
		require.Equal(t, filter.matcher, dst.matcher)
		require.Equal(t, filter.key, dst.key)

		dst.value = staticStringer("value2")
		dst.from = 2
		dst.matcher = 2
		dst.key = "2"

		require.NotEqual(t, filter.value, dst.value)
		require.NotEqual(t, filter.from, dst.from)
		require.NotEqual(t, filter.matcher, dst.matcher)
		require.NotEqual(t, filter.key, dst.key)
	})
}
