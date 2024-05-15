package netmap_test

import (
	"fmt"
	"strconv"
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
)

func TestNetMap_ReadFromV2(t *testing.T) {
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("nodes", func(t *testing.T) {
			testCases := []struct {
				name    string
				err     string
				corrupt func(*apinetmap.NodeInfo)
			}{
				{name: "nil public key", err: "missing public key", corrupt: func(n *apinetmap.NodeInfo) {
					n.PublicKey = nil
				}},
				{name: "empty public key", err: "missing public key", corrupt: func(n *apinetmap.NodeInfo) {
					n.PublicKey = []byte{}
				}},
				{name: "nil network endpoints", err: "missing network endpoints", corrupt: func(n *apinetmap.NodeInfo) {
					n.Addresses = nil
				}},
				{name: "empty network endpoints", err: "missing network endpoints", corrupt: func(n *apinetmap.NodeInfo) {
					n.Addresses = []string{}
				}},
				{name: "attributes/missing key", err: "invalid attribute #1: missing key", corrupt: func(n *apinetmap.NodeInfo) {
					n.Attributes = []*apinetmap.NodeInfo_Attribute{
						{Key: "key_valid", Value: "any"},
						{Key: "", Value: "any"},
					}
				}},
				{name: "attributes/repeated keys", err: "multiple attributes with key=k2", corrupt: func(n *apinetmap.NodeInfo) {
					n.Attributes = []*apinetmap.NodeInfo_Attribute{
						{Key: "k1", Value: "any"},
						{Key: "k2", Value: "1"},
						{Key: "k3", Value: "any"},
						{Key: "k2", Value: "2"},
					}
				}},
				{name: "attributes/missing value", err: "invalid attribute #1 (key2): missing value", corrupt: func(n *apinetmap.NodeInfo) {
					n.Attributes = []*apinetmap.NodeInfo_Attribute{
						{Key: "key1", Value: "any"},
						{Key: "key2", Value: ""},
					}
				}},
				{name: "attributes/price format", err: "invalid price attribute (#1): invalid integer", corrupt: func(n *apinetmap.NodeInfo) {
					n.Attributes = []*apinetmap.NodeInfo_Attribute{
						{Key: "any", Value: "any"},
						{Key: "Price", Value: "not_a_number"},
					}
				}},
				{name: "attributes/capacity format", err: "invalid capacity attribute (#1): invalid integer", corrupt: func(n *apinetmap.NodeInfo) {
					n.Attributes = []*apinetmap.NodeInfo_Attribute{
						{Key: "any", Value: "any"},
						{Key: "Capacity", Value: "not_a_number"},
					}
				}},
			}

			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					n := netmaptest.Netmap()
					n.SetNodes(netmaptest.NNodes(3))
					var m apinetmap.Netmap

					n.WriteToV2(&m)
					testCase.corrupt(m.Nodes[1])
					require.ErrorContains(t, n.ReadFromV2(&m), fmt.Sprintf("invalid node info #1: %s", testCase.err))
				})
			}
		})
	})
}

func TestNetMap_SetNodes(t *testing.T) {
	var nm netmap.NetMap

	require.Zero(t, nm.Nodes())

	nodes := netmaptest.NNodes(3)
	nm.SetNodes(nodes)
	require.Equal(t, nodes, nm.Nodes())

	nodesOther := netmaptest.NNodes(2)
	nm.SetNodes(nodesOther)
	require.Equal(t, nodesOther, nm.Nodes())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NetMap
			var msg apinetmap.Netmap

			dst.SetNodes(nodes)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Nodes)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Nodes())

			nodes := make([]netmap.NodeInfo, 3)
			for i := range nodes {
				si := strconv.Itoa(i + 1)
				nodes[i].SetPublicKey([]byte("pubkey_" + si))
				nodes[i].SetNetworkEndpoints([]string{"addr_" + si + "_1", "addr_" + si + "_2"})
				nodes[i].SetAttribute("attr_"+si+"_1", "val_"+si+"_1")
				nodes[i].SetAttribute("attr_"+si+"_2", "val_"+si+"_2")
			}
			nodes[0].SetOnline()
			nodes[1].SetOffline()
			nodes[2].SetMaintenance()

			src.SetNodes(nodes)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.NodeInfo{
				{PublicKey: []byte("pubkey_1"), Addresses: []string{"addr_1_1", "addr_1_2"},
					Attributes: []*apinetmap.NodeInfo_Attribute{
						{Key: "attr_1_1", Value: "val_1_1"},
						{Key: "attr_1_2", Value: "val_1_2"},
					},
					State: apinetmap.NodeInfo_ONLINE},
				{PublicKey: []byte("pubkey_2"), Addresses: []string{"addr_2_1", "addr_2_2"},
					Attributes: []*apinetmap.NodeInfo_Attribute{
						{Key: "attr_2_1", Value: "val_2_1"},
						{Key: "attr_2_2", Value: "val_2_2"},
					},
					State: apinetmap.NodeInfo_OFFLINE},
				{PublicKey: []byte("pubkey_3"), Addresses: []string{"addr_3_1", "addr_3_2"},
					Attributes: []*apinetmap.NodeInfo_Attribute{
						{Key: "attr_3_1", Value: "val_3_1"},
						{Key: "attr_3_2", Value: "val_3_2"},
					},
					State: apinetmap.NodeInfo_MAINTENANCE,
				},
			}, msg.Nodes)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, nodes, dst.Nodes())
		})
	})
}

func TestNetMap_SetEpoch(t *testing.T) {
	var nm netmap.NetMap

	require.Zero(t, nm.Epoch())

	const epoch = 13
	nm.SetEpoch(epoch)
	require.EqualValues(t, epoch, nm.Epoch())

	const epochOther = 42
	nm.SetEpoch(epochOther)
	require.EqualValues(t, epochOther, nm.Epoch())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NetMap
			var msg apinetmap.Netmap

			dst.SetEpoch(epoch)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Epoch)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Epoch())

			src.SetEpoch(epoch)

			src.WriteToV2(&msg)
			require.EqualValues(t, epoch, msg.Epoch)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, epoch, dst.Epoch())
		})
	})
}
