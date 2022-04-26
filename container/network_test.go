package container_test

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
)

func TestContainer_NetworkConfig(t *testing.T) {
	c := containertest.Container()
	nc := netmaptest.NetworkConfig()

	t.Run("default", func(t *testing.T) {
		require.True(t, c.HomomorphicHashingState())

		res, err := c.AssertNetworkConfig(*nc)
		require.NoError(t, err)

		require.True(t, res)
	})

	np := netmap.NewNetworkParameter()
	np.SetKey([]byte(container.HomomorphicHashingDisabledKey))
	np.SetValue(stackitem.NewBool(true).Bytes())

	nc.SetParameters(*np)

	t.Run("apply", func(t *testing.T) {
		require.True(t, c.HomomorphicHashingState())

		err := c.ApplyNetworkConfig(*nc)
		require.NoError(t, err)

		require.False(t, c.HomomorphicHashingState())
	})

	t.Run("apply_incorrect_net_cfg", func(t *testing.T) {
		np.SetValue(make([]byte, stackitem.MaxBigIntegerSizeBits))
		nc.SetParameters(*np)

		err := c.ApplyNetworkConfig(*nc)
		require.Error(t, err)
	})
}
