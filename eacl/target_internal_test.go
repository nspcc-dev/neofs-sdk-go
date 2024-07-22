package eacl

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTarget_CopyTo(t *testing.T) {
	var target Target
	target.SetRole(1)
	target.SetBinaryKeys([][]byte{
		{1, 2, 3},
	})

	t.Run("copy", func(t *testing.T) {
		var dst Target
		target.CopyTo(&dst)

		require.Equal(t, target, dst)
		require.True(t, bytes.Equal(target.Marshal(), dst.Marshal()))
	})

	t.Run("change", func(t *testing.T) {
		var dst Target
		target.CopyTo(&dst)

		require.Equal(t, target.role, dst.role)
		dst.SetRole(2)
		require.NotEqual(t, target.role, dst.role)

		require.True(t, bytes.Equal(target.keys[0], dst.keys[0]))
		// change some key data
		dst.keys[0][0] = 5
		require.False(t, bytes.Equal(target.keys[0], dst.keys[0]))
	})
}
