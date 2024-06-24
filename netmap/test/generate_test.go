package netmaptest_test

import (
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
)

func TestPlacementPolicy(t *testing.T) {
	v := netmaptest.PlacementPolicy()
	require.NotEqual(t, v, netmaptest.PlacementPolicy())

	var v2 netmap.PlacementPolicy
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apinetmap.PlacementPolicy
	v.WriteToV2(&m)
	var v3 netmap.PlacementPolicy
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v4 netmap.PlacementPolicy
	require.NoError(t, v4.UnmarshalJSON(j))
	require.Equal(t, v, v4)
}

func TestNetworkInfo(t *testing.T) {
	v := netmaptest.NetworkInfo()
	require.NotEqual(t, v, netmaptest.NetworkInfo())

	var v2 netmap.NetworkInfo
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apinetmap.NetworkInfo
	v.WriteToV2(&m)
	var v3 netmap.NetworkInfo
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)
}
