package container_test

import (
	"testing"

	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
)

func TestContainer_NetworkConfig(t *testing.T) {
	c := containertest.Container(t)
	nc := netmaptest.NetworkInfo()

	t.Run("default", func(t *testing.T) {
		require.False(t, c.IsHomomorphicHashingDisabled())

		res := c.AssertNetworkConfig(nc)

		require.True(t, res)
	})

	nc.DisableHomomorphicHashing()

	t.Run("apply", func(t *testing.T) {
		require.False(t, c.IsHomomorphicHashingDisabled())

		c.ApplyNetworkConfig(nc)

		require.True(t, c.IsHomomorphicHashingDisabled())
	})
}
