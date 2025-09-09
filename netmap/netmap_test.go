package netmap_test

import (
	"slices"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	"github.com/stretchr/testify/require"
)

var (
	// set by init.
	anyValidNodes = make([]netmap.NodeInfo, 2)
)

// set by init.
var validNetmap netmap.NetMap

func init() {
	anyValidNodes[0].SetPublicKey([]byte("public_key_0"))
	anyValidNodes[1].SetPublicKey([]byte("public_key_1"))
	anyValidNodes[0].SetNetworkEndpoints("endpoint_0_0", "endpoint_0_1")
	anyValidNodes[1].SetNetworkEndpoints("endpoint_1_0", "endpoint_1_1")
	anyValidNodes[0].SetOffline()
	anyValidNodes[1].SetMaintenance()
	anyValidNodes[0].SetAttribute("k_0_0", "v_0_0")
	anyValidNodes[0].SetAttribute("k_0_1", "v_0_1")
	anyValidNodes[1].SetAttribute("k_1_0", "v_1_0")
	anyValidNodes[1].SetAttribute("k_1_1", "v_1_1")
	anyValidNodes[0].SetPrice(15371821855482965181)
	anyValidNodes[1].SetPrice(10938191854812987682)
	anyValidNodes[0].SetCapacity(661331659952220723)
	anyValidNodes[1].SetCapacity(14730606796610755576)
	anyValidNodes[0].SetLOCODE("SE STO")
	anyValidNodes[1].SetLOCODE("FI HEL")
	anyValidNodes[0].SetCountryCode("SE")
	anyValidNodes[1].SetCountryCode("FI")
	anyValidNodes[0].SetCountryName("Sweden")
	anyValidNodes[1].SetCountryName("Finland")
	anyValidNodes[0].SetLocationName("Stockholm")
	anyValidNodes[1].SetLocationName("Helsinki")
	anyValidNodes[0].SetSubdivisionCode("AB")
	anyValidNodes[1].SetSubdivisionCode("AI")
	anyValidNodes[0].SetSubdivisionName("Stockholms län")
	anyValidNodes[1].SetSubdivisionName("Helmand")
	anyValidNodes[0].SetContinentName("Europe")
	anyValidNodes[1].SetContinentName("Africa") // wrong, but only for diff
	anyValidNodes[0].SetExternalAddresses("ext_endpoint_0_0", "ext_endpoint_0_1")
	anyValidNodes[1].SetExternalAddresses("ext_endpoint_1_0", "ext_endpoint_1_1")
	anyValidNodes[0].SetVersion("v0.1.2")
	anyValidNodes[1].SetVersion("v3.4.5")
	anyValidNodes[0].SetVerifiedNodesDomain("domain0.neofs")
	anyValidNodes[1].SetVerifiedNodesDomain("domain1.neofs")

	validNetmap.SetEpoch(anyValidCurrentEpoch)
	validNetmap.SetNodes(anyValidNodes)
}

func TestNetMap_FromProtoMessage(t *testing.T) {
	m := &protonetmap.Netmap{
		Epoch: anyValidCurrentEpoch,
		Nodes: []*protonetmap.NodeInfo{
			{
				PublicKey: []byte("public_key_0"),
				Addresses: []string{"endpoint_0_0", "endpoint_0_1"},
				Attributes: []*protonetmap.NodeInfo_Attribute{
					{Key: "k_0_0", Value: "v_0_0"},
					{Key: "k_0_1", Value: "v_0_1"},
				},
				State: protonetmap.NodeInfo_OFFLINE,
			},
			{
				PublicKey: []byte("public_key_1"),
				Addresses: []string{"endpoint_1_0", "endpoint_1_1"},
				Attributes: []*protonetmap.NodeInfo_Attribute{
					{Key: "k_1_0", Value: "v_1_0"},
					{Key: "k_1_1", Value: "v_1_1"},
				},
				State: protonetmap.NodeInfo_MAINTENANCE,
			},
		},
	}

	var val netmap.NetMap
	require.NoError(t, val.FromProtoMessage(m))

	require.EqualValues(t, anyValidCurrentEpoch, val.Epoch())
	ns := val.Nodes()
	require.Len(t, ns, 2)
	require.EqualValues(t, "public_key_0", ns[0].PublicKey())
	require.EqualValues(t, "public_key_1", ns[1].PublicKey())
	require.True(t, ns[0].IsOffline())
	require.True(t, ns[1].IsMaintenance())

	require.EqualValues(t, 2, ns[0].NumberOfNetworkEndpoints())
	require.EqualValues(t, 2, ns[1].NumberOfNetworkEndpoints())
	collectedEndpoints := slices.Collect(ns[0].NetworkEndpoints())
	require.Equal(t, []string{"endpoint_0_0", "endpoint_0_1"}, collectedEndpoints)
	collectedEndpoints = slices.Collect(ns[1].NetworkEndpoints())
	require.Equal(t, []string{"endpoint_1_0", "endpoint_1_1"}, collectedEndpoints)

	require.EqualValues(t, 2, ns[0].NumberOfAttributes())
	require.EqualValues(t, 2, ns[1].NumberOfAttributes())
	require.Equal(t, "v_0_0", ns[0].Attribute("k_0_0"))
	require.Equal(t, "v_0_1", ns[0].Attribute("k_0_1"))
	require.Equal(t, "v_1_0", ns[1].Attribute("k_1_0"))
	require.Equal(t, "v_1_1", ns[1].Attribute("k_1_1"))
	var collectedAttrs []string
	for k, v := range ns[0].Attributes() {
		collectedAttrs = append(collectedAttrs, k, v)
	}
	require.Equal(t, []string{
		"k_0_0", "v_0_0",
		"k_0_1", "v_0_1",
	}, collectedAttrs)
	collectedAttrs = nil
	for k, v := range ns[1].Attributes() {
		collectedAttrs = append(collectedAttrs, k, v)
	}
	require.Equal(t, []string{
		"k_1_0", "v_1_0",
		"k_1_1", "v_1_1",
	}, collectedAttrs)

	// reset optional fields
	m.Epoch = 0
	m.Nodes = nil
	val2 := val
	require.NoError(t, val2.FromProtoMessage(m))
	require.Zero(t, val2.Epoch())
	require.Zero(t, val2.Nodes())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(netMap *protonetmap.Netmap)
		}{
			{name: "nodes/nil", err: "nil node info #1",
				corrupt: func(m *protonetmap.Netmap) { m.Nodes[1] = nil }},
			{name: "nodes/public key/nil", err: "invalid node info: missing public key",
				corrupt: func(m *protonetmap.Netmap) { m.Nodes[1].PublicKey = nil }},
			{name: "nodes/public key/empty", err: "invalid node info: missing public key",
				corrupt: func(m *protonetmap.Netmap) { m.Nodes[1].PublicKey = []byte{} }},
			{name: "nodes/endpoints/empty", err: "invalid node info: missing network endpoints",
				corrupt: func(m *protonetmap.Netmap) { m.Nodes[1].Addresses = nil }},
			{name: "nodes/attributes/no key", err: "invalid node info: empty key of the attribute #1",
				corrupt: func(m *protonetmap.Netmap) { setNodeAttributes(m.Nodes[1], "k1", "v1", "", "v2") }},
			{name: "nodes/attributes/no value", err: `invalid node info: empty "k2" attribute value`,
				corrupt: func(m *protonetmap.Netmap) { setNodeAttributes(m.Nodes[1], "k1", "v1", "k2", "") }},
			{name: "nodes/attributes/duplicated", err: "invalid node info: duplicated attribute k1",
				corrupt: func(m *protonetmap.Netmap) { setNodeAttributes(m.Nodes[1], "k1", "v1", "k2", "v2", "k1", "v3") }},
			{name: "nodes/attributes/capacity", err: "invalid node info: invalid Capacity attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *protonetmap.Netmap) { setNodeAttributes(m.Nodes[1], "Capacity", "foo") }},
			{name: "nodes/attributes/price", err: "invalid node info: invalid Price attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *protonetmap.Netmap) { setNodeAttributes(m.Nodes[1], "Price", "foo") }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(netmap.NetMap).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestNetMap_ProtoMessage(t *testing.T) {
	var val netmap.NetMap

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetEpoch())
	require.Zero(t, m.GetNodes())

	// filled
	m = validNetmap.ProtoMessage()
	require.EqualValues(t, anyValidCurrentEpoch, m.GetEpoch())
	ns := m.GetNodes()
	require.Len(t, ns, 2)
	require.EqualValues(t, "public_key_0", ns[0].GetPublicKey())
	require.EqualValues(t, "public_key_1", ns[1].GetPublicKey())
}

func TestNetMap_SetNodes(t *testing.T) {
	var nm netmap.NetMap
	require.Zero(t, nm.Nodes())

	nm.SetNodes(anyValidNodes)
	require.Equal(t, anyValidNodes, nm.Nodes())
}

func TestNetMap_SetEpoch(t *testing.T) {
	var nm netmap.NetMap
	require.Zero(t, nm.Epoch())

	nm.SetEpoch(anyValidCurrentEpoch)
	require.EqualValues(t, anyValidCurrentEpoch, nm.Epoch())

	nm.SetEpoch(anyValidCurrentEpoch + 1)
	require.EqualValues(t, anyValidCurrentEpoch+1, nm.Epoch())
}

func TestNetMap_ContainerNodes(t *testing.T) {
	t.Run("unlinked selectors", func(t *testing.T) {
		t.Run("more than REP rules", func(t *testing.T) {
			nodes := nNodes(5)
			const ps = "REP 1 CBF 1 SELECT 1 FROM * SELECT 1 FROM *"

			var p netmap.PlacementPolicy
			require.NoError(t, p.DecodeString(ps))

			var nm netmap.NetMap
			nm.SetNodes(nodes)

			res, err := nm.ContainerNodes(p, cidtest.ID())
			require.NoError(t, err)
			require.Len(t, res, 1) // num of REP statements
			require.Len(t, res[0], 2)
			for _, set := range res {
				require.Len(t, set, 2)
				require.Subset(t, nodes, set)
			}
		})
	})
}

func nNodes(n int) []netmap.NodeInfo {
	res := make([]netmap.NodeInfo, n)
	for i := range res {
		res[i] = netmaptest.NodeInfo()
	}
	return res
}
