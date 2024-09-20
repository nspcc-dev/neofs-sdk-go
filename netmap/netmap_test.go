package netmap_test

import (
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
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

func TestNetMap_ReadFromV2(t *testing.T) {
	mns := make([]apinetmap.NodeInfo, 2)
	mns[0].SetPublicKey([]byte("public_key_0"))
	mns[1].SetPublicKey([]byte("public_key_1"))
	mns[0].SetAddresses("endpoint_0_0", "endpoint_0_1")
	mns[1].SetAddresses("endpoint_1_0", "endpoint_1_1")
	mns[0].SetState(apinetmap.Offline)
	mns[1].SetState(apinetmap.Maintenance)

	addAttr := func(m *apinetmap.NodeInfo, k, v string) {
		var a apinetmap.Attribute
		a.SetKey(k)
		a.SetValue(v)
		m.SetAttributes(append(m.GetAttributes(), a))
	}
	addAttr(&mns[0], "k_0_0", "v_0_0")
	addAttr(&mns[0], "k_0_1", "v_0_1")
	addAttr(&mns[1], "k_1_0", "v_1_0")
	addAttr(&mns[1], "k_1_1", "v_1_1")

	var m apinetmap.NetMap
	m.SetEpoch(anyValidCurrentEpoch)
	m.SetNodes(mns)

	var val netmap.NetMap
	require.NoError(t, val.ReadFromV2(m))

	require.EqualValues(t, anyValidCurrentEpoch, val.Epoch())
	ns := val.Nodes()
	require.Len(t, ns, 2)
	require.EqualValues(t, "public_key_0", ns[0].PublicKey())
	require.EqualValues(t, "public_key_1", ns[1].PublicKey())
	require.True(t, ns[0].IsOffline())
	require.True(t, ns[1].IsMaintenance())

	require.EqualValues(t, 2, ns[0].NumberOfNetworkEndpoints())
	require.EqualValues(t, 2, ns[1].NumberOfNetworkEndpoints())
	var collectedEndpoints []string
	ns[0].IterateNetworkEndpoints(func(el string) bool {
		collectedEndpoints = append(collectedEndpoints, el)
		return false
	})
	require.Equal(t, []string{"endpoint_0_0", "endpoint_0_1"}, collectedEndpoints)
	collectedEndpoints = nil
	ns[1].IterateNetworkEndpoints(func(el string) bool { collectedEndpoints = append(collectedEndpoints, el); return false })
	require.Equal(t, []string{"endpoint_1_0", "endpoint_1_1"}, collectedEndpoints)

	require.EqualValues(t, 2, ns[0].NumberOfAttributes())
	require.EqualValues(t, 2, ns[1].NumberOfAttributes())
	require.Equal(t, "v_0_0", ns[0].Attribute("k_0_0"))
	require.Equal(t, "v_0_1", ns[0].Attribute("k_0_1"))
	require.Equal(t, "v_1_0", ns[1].Attribute("k_1_0"))
	require.Equal(t, "v_1_1", ns[1].Attribute("k_1_1"))
	var collectedAttrs []string
	ns[0].IterateAttributes(func(k, v string) { collectedAttrs = append(collectedAttrs, k, v) })
	require.Equal(t, []string{
		"k_0_0", "v_0_0",
		"k_0_1", "v_0_1",
	}, collectedAttrs)
	collectedAttrs = nil
	ns[1].IterateAttributes(func(k, v string) { collectedAttrs = append(collectedAttrs, k, v) })
	require.Equal(t, []string{
		"k_1_0", "v_1_0",
		"k_1_1", "v_1_1",
	}, collectedAttrs)

	// reset optional fields
	m.SetEpoch(0)
	m.SetNodes(nil)
	val2 := val
	require.NoError(t, val2.ReadFromV2(m))
	require.Zero(t, val2.Epoch())
	require.Zero(t, val2.Nodes())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(netMap *apinetmap.NetMap)
		}{
			{name: "nodes/public key/nil", err: "invalid node info: missing public key",
				corrupt: func(m *apinetmap.NetMap) { m.Nodes()[1].SetPublicKey(nil) }},
			{name: "nodes/public key/empty", err: "invalid node info: missing public key",
				corrupt: func(m *apinetmap.NetMap) { m.Nodes()[1].SetPublicKey([]byte{}) }},
			{name: "nodes/endpoints/empty", err: "invalid node info: missing network endpoints",
				corrupt: func(m *apinetmap.NetMap) { m.Nodes()[1].SetAddresses() }},
			{name: "nodes/attributes/no key", err: "invalid node info: empty key of the attribute #1",
				corrupt: func(m *apinetmap.NetMap) { setNodeAttributes(&m.Nodes()[1], "k1", "v1", "", "v2") }},
			{name: "nodes/attributes/no value", err: "invalid node info: empty value of the attribute k2",
				corrupt: func(m *apinetmap.NetMap) { setNodeAttributes(&m.Nodes()[1], "k1", "v1", "k2", "") }},
			{name: "nodes/attributes/duplicated", err: "invalid node info: duplicated attribute k1",
				corrupt: func(m *apinetmap.NetMap) { setNodeAttributes(&m.Nodes()[1], "k1", "v1", "k2", "v2", "k1", "v3") }},
			{name: "nodes/attributes/capacity", err: "invalid node info: invalid Capacity attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *apinetmap.NetMap) { setNodeAttributes(&m.Nodes()[1], "Capacity", "foo") }},
			{name: "nodes/attributes/price", err: "invalid node info: invalid Price attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *apinetmap.NetMap) { setNodeAttributes(&m.Nodes()[1], "Price", "foo") }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				var m apinetmap.NetMap
				st.WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(netmap.NetMap).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestNetMap_WriteToV2(t *testing.T) {
	var val netmap.NetMap
	var m apinetmap.NetMap

	// zero
	val.WriteToV2(&m)
	require.Zero(t, m.Epoch())
	require.Zero(t, m.Nodes())

	// filled
	validNetmap.WriteToV2(&m)
	require.EqualValues(t, anyValidCurrentEpoch, m.Epoch())
	ns := m.Nodes()
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
