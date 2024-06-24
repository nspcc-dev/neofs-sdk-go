package container_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

func TestContainer_ApplyNetworkConfig(t *testing.T) {
	var c container.Container
	var n netmap.NetworkInfo

	require.True(t, c.AssertNetworkConfig(n))

	n.SetHomomorphicHashingDisabled(true)
	require.False(t, c.AssertNetworkConfig(n))

	for _, testCase := range []struct {
		cnr, net bool
	}{
		{cnr: false, net: false},
		{cnr: false, net: true},
		{cnr: true, net: false},
		{cnr: true, net: true},
	} {
		c.SetHomomorphicHashingDisabled(testCase.cnr)
		n.SetHomomorphicHashingDisabled(testCase.net)
		require.Equal(t, testCase.cnr == testCase.net, c.AssertNetworkConfig(n), testCase)

		c.ApplyNetworkConfig(n)
		require.True(t, c.AssertNetworkConfig(n))
	}
}
