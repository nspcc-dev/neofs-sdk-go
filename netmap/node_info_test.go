package netmap_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	"github.com/stretchr/testify/require"
)

const (
	anyValidNodePrice           = uint64(10993309018040354285)
	anyValidNodeCapacity        = uint64(9010937245406684209)
	anyValidLOCODE              = "SE STO"
	anyValidCountryCode         = "SE"
	anyValidCountryName         = "Sweden"
	anyValidLocationName        = "Stockholm"
	anyValidSubdivCode          = "AB"
	anyValidSubdivName          = "Stockholms län"
	anyValidContinentName       = "Europe"
	anyValidNodeVersion         = "v1.2.3"
	anyValidVerifiedNodesDomain = "example.some-org.neofs"
)

var (
	anyValidPublicKey = []byte{3, 57, 228, 197, 42, 18, 179, 5, 89, 50, 71, 221, 118, 152, 192, 108, 201, 220, 179, 171, 53, 215,
		121, 249, 1, 162, 172, 246, 94, 39, 117, 42, 73}
	anyValidNetworkEndpoints         = []string{"endpoint1", "endpoint2", "endpoint3"}
	anyValidExternalNetworkEndpoints = []string{"ext_endpoint1", "ext_endpoint2", "", "ext_endpoint3"}
)

// set by init.
var validNodeInfo netmap.NodeInfo

func init() {
	validNodeInfo.SetPublicKey(anyValidPublicKey)
	validNodeInfo.SetNetworkEndpoints(anyValidNetworkEndpoints...)
	validNodeInfo.SetMaintenance()
	validNodeInfo.SetAttribute("k1", "v1")
	validNodeInfo.SetAttribute("k2", "v2")
	validNodeInfo.SetPrice(anyValidNodePrice)
	validNodeInfo.SetCapacity(anyValidNodeCapacity)
	validNodeInfo.SetLOCODE(anyValidLOCODE)
	validNodeInfo.SetCountryCode(anyValidCountryCode)
	validNodeInfo.SetCountryName(anyValidCountryName)
	validNodeInfo.SetLocationName(anyValidLocationName)
	validNodeInfo.SetSubdivisionCode(anyValidSubdivCode)
	validNodeInfo.SetSubdivisionName(anyValidSubdivName)
	validNodeInfo.SetContinentName(anyValidContinentName)
	validNodeInfo.SetExternalAddresses(anyValidExternalNetworkEndpoints...)
	validNodeInfo.SetVersion(anyValidNodeVersion)
	validNodeInfo.SetVerifiedNodesDomain(anyValidVerifiedNodesDomain)
}

// corresponds to validNodeInfo.
var validBinNodeInfo = []byte{
	10, 33, 3, 57, 228, 197, 42, 18, 179, 5, 89, 50, 71, 221, 118, 152, 192, 108, 201, 220, 179, 171, 53, 215, 121, 249, 1, 162,
	172, 246, 94, 39, 117, 42, 73, 18, 9, 101, 110, 100, 112, 111, 105, 110, 116, 49, 18, 9, 101, 110, 100, 112, 111, 105, 110, 116, 50,
	18, 9, 101, 110, 100, 112, 111, 105, 110, 116, 51, 26, 8, 10, 2, 107, 49, 18, 2, 118, 49, 26, 8, 10, 2, 107, 50, 18, 2, 118, 50,
	26, 29, 10, 5, 80, 114, 105, 99, 101, 18, 20, 49, 48, 57, 57, 51, 51, 48, 57, 48, 49, 56, 48, 52, 48, 51, 53, 52, 50, 56, 53,
	26, 31, 10, 8, 67, 97, 112, 97, 99, 105, 116, 121, 18, 19, 57, 48, 49, 48, 57, 51, 55, 50, 52, 53, 52, 48, 54, 54, 56, 52, 50,
	48, 57, 26, 19, 10, 9, 85, 78, 45, 76, 79, 67, 79, 68, 69, 18, 6, 83, 69, 32, 83, 84, 79, 26, 17, 10, 11, 67, 111, 117, 110, 116,
	114, 121, 67, 111, 100, 101, 18, 2, 83, 69, 26, 17, 10, 7, 67, 111, 117, 110, 116, 114, 121, 18, 6, 83, 119, 101, 100, 101, 110, 26,
	21, 10, 8, 76, 111, 99, 97, 116, 105, 111, 110, 18, 9, 83, 116, 111, 99, 107, 104, 111, 108, 109, 26, 16, 10, 10, 83, 117, 98, 68, 105,
	118, 67, 111, 100, 101, 18, 2, 65, 66, 26, 25, 10, 6, 83, 117, 98, 68, 105, 118, 18, 15, 83, 116, 111, 99, 107, 104, 111, 108, 109,
	115, 32, 108, 195, 164, 110, 26, 19, 10, 9, 67, 111, 110, 116, 105, 110, 101, 110, 116, 18, 6, 69, 117, 114, 111, 112, 101, 26, 58, 10,
	12, 69, 120, 116, 101, 114, 110, 97, 108, 65, 100, 100, 114, 18, 42, 101, 120, 116, 95, 101, 110, 100, 112, 111, 105, 110, 116, 49,
	44, 101, 120, 116, 95, 101, 110, 100, 112, 111, 105, 110, 116, 50, 44, 44, 101, 120, 116, 95, 101, 110, 100, 112, 111, 105, 110, 116,
	51, 26, 17, 10, 7, 86, 101, 114, 115, 105, 111, 110, 18, 6, 118, 49, 46, 50, 46, 51, 26, 45, 10, 19, 86, 101, 114, 105, 102, 105, 101,
	100, 78, 111, 100, 101, 115, 68, 111, 109, 97, 105, 110, 18, 22, 101, 120, 97, 109, 112, 108, 101, 46, 115, 111, 109, 101, 45, 111,
	114, 103, 46, 110, 101, 111, 102, 115, 32, 3,
}

var validJSONNodeInfo = `
{
 "publicKey": "AznkxSoSswVZMkfddpjAbMncs6s113n5AaKs9l4ndSpJ",
 "addresses": [
  "endpoint1",
  "endpoint2",
  "endpoint3"
 ],
 "attributes": [
  {
   "key": "k1",
   "value": "v1",
   "parents": []
  },
  {
   "key": "k2",
   "value": "v2",
   "parents": []
  },
  {
   "key": "Price",
   "value": "10993309018040354285",
   "parents": []
  },
  {
   "key": "Capacity",
   "value": "9010937245406684209",
   "parents": []
  },
  {
   "key": "UN-LOCODE",
   "value": "SE STO",
   "parents": []
  },
  {
   "key": "CountryCode",
   "value": "SE",
   "parents": []
  },
  {
   "key": "Country",
   "value": "Sweden",
   "parents": []
  },
  {
   "key": "Location",
   "value": "Stockholm",
   "parents": []
  },
  {
   "key": "SubDivCode",
   "value": "AB",
   "parents": []
  },
  {
   "key": "SubDiv",
   "value": "Stockholms län",
   "parents": []
  },
  {
   "key": "Continent",
   "value": "Europe",
   "parents": []
  },
  {
   "key": "ExternalAddr",
   "value": "ext_endpoint1,ext_endpoint2,,ext_endpoint3",
   "parents": []
  },
  {
   "key": "Version",
   "value": "v1.2.3",
   "parents": []
  },
  {
   "key": "VerifiedNodesDomain",
   "value": "example.some-org.neofs",
   "parents": []
  }
 ],
 "state": "MAINTENANCE"
}
`

func TestNodeInfo_SetPublicKey(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.PublicKey())

	pub := []byte("any_bytes")
	n.SetPublicKey(pub)
	require.Equal(t, pub, n.PublicKey())

	otherPub := append(pub, "_other"...)
	n.SetPublicKey(otherPub)
	require.Equal(t, otherPub, n.PublicKey())
}

func TestStringifyPublicKey(t *testing.T) {
	var n netmap.NodeInfo
	require.Empty(t, netmap.StringifyPublicKey(n))

	n.SetPublicKey(anyValidPublicKey)
	require.Equal(t, "0339e4c52a12b305593247dd7698c06cc9dcb3ab35d779f901a2acf65e27752a49", netmap.StringifyPublicKey(n))
}

func TestNodeInfo_SetNetworkEndpoints(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.NumberOfNetworkEndpoints())
	n.IterateNetworkEndpoints(func(string) bool {
		t.Fatal("handler must not be called")
		return false
	})

	require.EqualValues(t, 3, validNodeInfo.NumberOfNetworkEndpoints())
	var collected []string
	validNodeInfo.IterateNetworkEndpoints(func(el string) bool {
		collected = append(collected, el)
		return false
	})
	require.Equal(t, anyValidNetworkEndpoints, collected)
}

func TestNodeInfo_IterateNetworkEndpoints(t *testing.T) {
	var collected []string
	validNodeInfo.IterateNetworkEndpoints(func(el string) bool {
		collected = append(collected, el)
		return len(collected) == 2
	})
	require.Equal(t, anyValidNetworkEndpoints[:2], collected)
}

func TestIterateNetworkEndpoints(t *testing.T) {
	var collected []string
	netmap.IterateNetworkEndpoints(validNodeInfo, func(el string) {
		collected = append(collected, el)
	})
	require.Equal(t, anyValidNetworkEndpoints, collected)
}

func TestNodeInfo_Hash(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.Hash())

	require.EqualValues(t, uint64(11151666526377957836), validNodeInfo.Hash())
}

func TestNodeInfo_SetPrice(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.Price())

	n.SetPrice(anyValidNodePrice)
	require.EqualValues(t, anyValidNodePrice, n.Price())

	n.SetPrice(anyValidNodePrice + 1)
	require.EqualValues(t, anyValidNodePrice+1, n.Price())
}

func TestNodeInfo_SetCapacity(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.Attribute("Capacity"))

	n.SetCapacity(anyValidNodeCapacity)
	require.EqualValues(t, "9010937245406684209", n.Attribute("Capacity"))
}

func TestNodeInfo_SetLOCODE(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.LOCODE())

	n.SetLOCODE(anyValidLOCODE)
	require.Equal(t, anyValidLOCODE, n.LOCODE())
}

func TestNodeInfo_SetCountryCode(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.CountryCode())

	n.SetCountryCode(anyValidCountryCode)
	require.Equal(t, anyValidCountryCode, n.CountryCode())
}

func TestNodeInfo_SetCountryName(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.CountryName())

	n.SetCountryName(anyValidCountryName)
	require.Equal(t, anyValidCountryName, n.CountryName())
}

func TestNodeInfo_SetLocationName(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.LocationName())

	n.SetLocationName(anyValidLocationName)
	require.Equal(t, anyValidLocationName, n.LocationName())
}

func TestNodeInfo_SetSubdivisionCode(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.SubdivisionCode())

	n.SetSubdivisionCode(anyValidSubdivCode)
	require.Equal(t, anyValidSubdivCode, n.SubdivisionCode())
}

func TestNodeInfo_SetSubdivisionName(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.SubdivisionName())

	n.SetSubdivisionName(anyValidSubdivName)
	require.Equal(t, anyValidSubdivName, n.SubdivisionName())
}

func TestNodeInfo_SetContinentName(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.ContinentName())

	n.SetContinentName(anyValidContinentName)
	require.Equal(t, anyValidContinentName, n.ContinentName())
}

func TestNodeInfo_SetAttributes(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.NumberOfAttributes())
	n.IterateAttributes(func(string, string) {
		t.Fatal("handler must not be called")
	})

	const k1, v1 = "k1", "v1"
	const k2, v2 = "k2", "v2"
	const empty = ""

	attr1 := [2]string{empty, v1}
	attr2 := [2]string{k2, empty}

	require.Panics(t, func() {
		n.SetAttributes([][2]string{attr1})
	})

	require.Panics(t, func() {
		n.SetAttributes([][2]string{attr2})
	})

	attr1 = [2]string{k1, v1}
	n.SetAttributes([][2]string{attr1})
	require.Equal(t, attr1[1], n.Attribute(k1))
	require.Equal(t, empty, n.Attribute(k2))
	require.Equal(t, 1, n.NumberOfAttributes())

	attr2 = [2]string{k2, v2}
	n.SetAttributes([][2]string{attr2})
	require.Equal(t, attr2[1], n.Attribute(k2))
	require.Equal(t, empty, n.Attribute(k1))
	require.Equal(t, 1, n.NumberOfAttributes())

	n.SetAttributes([][2]string{attr1, attr2})
	require.Equal(t, attr1[1], n.Attribute(k1))
	require.Equal(t, attr2[1], n.Attribute(k2))
	require.Equal(t, 2, n.NumberOfAttributes())

	n.SetAttributes([][2]string{})
	require.Zero(t, n.NumberOfAttributes())
}

func TestNodeInfo_GetAttributes(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.NumberOfAttributes())
	n.IterateAttributes(func(string, string) {
		t.Fatal("handler must not be called")
	})

	const k1, v1 = "k1", "v1"
	const k2, v2 = "k2", "v2"

	attr1 := [2]string{k1, v1}
	attr2 := [2]string{k2, v2}

	n.SetAttributes([][2]string{attr1})
	attrs := n.GetAttributes()
	require.Len(t, attrs, 1)
	require.Equal(t, attr1, attrs[0])

	n.SetAttributes([][2]string{attr1, attr2})
	attrs = n.GetAttributes()
	require.Len(t, attrs, 2)
	require.Equal(t, attr1, attrs[0])
	require.Equal(t, attr2, attrs[1])
}

func TestNodeInfo_SetAttribute(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.NumberOfAttributes())
	n.IterateAttributes(func(string, string) {
		t.Fatal("handler must not be called")
	})
	require.Panics(t, func() { n.SetAttribute("", "v") })
	require.Panics(t, func() { n.SetAttribute("k", "") })

	const k1, v1 = "k1", "v1"
	const k2, v2 = "k2", "v2"

	n.SetAttribute(k1, v1)
	n.SetAttribute(k2, v2)
	require.Equal(t, v1, n.Attribute(k1))
	require.Equal(t, v2, n.Attribute(k2))
	require.EqualValues(t, 2, n.NumberOfAttributes())

	var collected []string
	n.IterateAttributes(func(k, v string) {
		collected = append(collected, k, v)
	})
	require.Equal(t, []string{
		k1, v1,
		k2, v2,
	}, collected)

	n.SetAttribute(k1, v1+"_other")
	require.Equal(t, v1+"_other", n.Attribute(k1))
}

func TestNodeInfo_SortAttributes(t *testing.T) {
	var n netmap.NodeInfo

	const k1, v1 = "k3", "v3"
	const k2, v2 = "k2", "v2"
	const k3, v3 = "k1", "v1"
	const k4, v4 = "k4", "v4"

	n.SetAttribute(k1, v1)
	n.SetAttribute(k2, v2)
	n.SetAttribute(k3, v3)
	n.SetAttribute(k4, v4)

	var collected []string
	n.IterateAttributes(func(k, v string) {
		collected = append(collected, k, v)
	})
	require.Equal(t, []string{
		k1, v1,
		k2, v2,
		k3, v3,
		k4, v4,
	}, collected)

	n.SortAttributes()
	collected = nil
	n.IterateAttributes(func(k, v string) {
		collected = append(collected, k, v)
	})
	require.Equal(t, []string{
		k3, v3,
		k2, v2,
		k1, v1,
		k4, v4,
	}, collected)
}

func testNodeStatusChange(t *testing.T, get func(netmap.NodeInfo) bool, set func(*netmap.NodeInfo), others []func(*netmap.NodeInfo)) {
	var n netmap.NodeInfo
	require.False(t, get(n))

	for _, change := range others {
		set(&n)
		require.True(t, get(n))
		change(&n)
		require.False(t, get(n))
	}
}

func TestNodeInfo_SetOffline(t *testing.T) {
	testNodeStatusChange(t, netmap.NodeInfo.IsOffline, (*netmap.NodeInfo).SetOffline, []func(*netmap.NodeInfo){
		(*netmap.NodeInfo).SetOnline,
		(*netmap.NodeInfo).SetMaintenance,
	})
}

func TestNodeInfo_SetOnline(t *testing.T) {
	testNodeStatusChange(t, netmap.NodeInfo.IsOnline, (*netmap.NodeInfo).SetOnline, []func(*netmap.NodeInfo){
		(*netmap.NodeInfo).SetOffline,
		(*netmap.NodeInfo).SetMaintenance,
	})
}

func TestNodeInfo_SetMaintenance(t *testing.T) {
	testNodeStatusChange(t, netmap.NodeInfo.IsMaintenance, (*netmap.NodeInfo).SetMaintenance, []func(*netmap.NodeInfo){
		(*netmap.NodeInfo).SetOffline,
		(*netmap.NodeInfo).SetOnline,
	})
}

func TestNodeInfo_SetVersion(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.Attribute("Version"))

	n.SetVersion(anyValidNodeVersion)
	require.Equal(t, anyValidNodeVersion, n.Attribute("Version"))
}

func TestNodeInfo_SetExternalAddresses(t *testing.T) {
	var n netmap.NodeInfo
	require.Zero(t, n.ExternalAddresses())

	n.SetExternalAddresses(anyValidExternalNetworkEndpoints...)
	require.Equal(t, anyValidExternalNetworkEndpoints, n.ExternalAddresses())
}

func TestNodeInfo_SetVerifiedNodesDomain(t *testing.T) {
	var n netmap.NodeInfo

	require.Zero(t, n.VerifiedNodesDomain())

	n.SetVerifiedNodesDomain(anyValidVerifiedNodesDomain)
	require.Equal(t, anyValidVerifiedNodesDomain, n.VerifiedNodesDomain())
}

func setNodeAttributes(ni *protonetmap.NodeInfo, els ...string) {
	if len(els)%2 != 0 {
		panic("must be even")
	}
	ni.Attributes = make([]*protonetmap.NodeInfo_Attribute, len(els)/2)
	for i := range len(els) / 2 {
		ni.Attributes[i] = &protonetmap.NodeInfo_Attribute{Key: els[2*i], Value: els[2*i+1]}
	}
}

func TestNodeInfo_FromProtoMessage(t *testing.T) {
	m := &protonetmap.NodeInfo{
		PublicKey: anyValidPublicKey,
		Addresses: anyValidNetworkEndpoints,
		Attributes: []*protonetmap.NodeInfo_Attribute{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: "Capacity", Value: "9010937245406684209"},
			{Key: "Price", Value: "10993309018040354285"},
			{Key: "UN-LOCODE", Value: anyValidLOCODE},
			{Key: "CountryCode", Value: anyValidCountryCode},
			{Key: "Country", Value: anyValidCountryName},
			{Key: "Location", Value: anyValidLocationName},
			{Key: "SubDivCode", Value: anyValidSubdivCode},
			{Key: "SubDiv", Value: anyValidSubdivName},
			{Key: "SubDivName", Value: anyValidSubdivName},
			{Key: "Continent", Value: anyValidContinentName},
			{Key: "ExternalAddr", Value: strings.Join(anyValidExternalNetworkEndpoints, ",")},
			{Key: "Version", Value: anyValidNodeVersion},
			{Key: "VerifiedNodesDomain", Value: anyValidVerifiedNodesDomain},
		},
	}

	var val netmap.NodeInfo
	require.NoError(t, val.FromProtoMessage(m))
	require.Equal(t, anyValidPublicKey, val.PublicKey())
	var i int
	val.IterateNetworkEndpoints(func(el string) bool {
		require.Equal(t, anyValidNetworkEndpoints[i], el)
		i++
		return false
	})
	require.Len(t, anyValidNetworkEndpoints, i)
	for _, checkState := range []func(netmap.NodeInfo) bool{
		netmap.NodeInfo.IsOnline,
		netmap.NodeInfo.IsOffline,
		netmap.NodeInfo.IsMaintenance,
	} {
		require.False(t, checkState(val))
	}
	require.EqualValues(t, 15, val.NumberOfAttributes())
	require.Equal(t, "v1", val.Attribute("k1"))
	require.Equal(t, "v2", val.Attribute("k2"))
	require.Equal(t, "9010937245406684209", val.Attribute("Capacity"))
	require.Equal(t, anyValidNodePrice, val.Price())
	require.Equal(t, anyValidLOCODE, val.LOCODE())
	require.Equal(t, anyValidCountryCode, val.CountryCode())
	require.Equal(t, anyValidCountryName, val.CountryName())
	require.Equal(t, anyValidLocationName, val.LocationName())
	require.Equal(t, anyValidSubdivCode, val.SubdivisionCode())
	require.Equal(t, anyValidSubdivName, val.SubdivisionName())
	require.Equal(t, anyValidContinentName, val.ContinentName())
	require.Equal(t, anyValidExternalNetworkEndpoints, val.ExternalAddresses())
	require.Equal(t, anyValidNodeVersion, val.Attribute("Version"))
	require.Equal(t, anyValidVerifiedNodesDomain, val.VerifiedNodesDomain())

	for _, tc := range []struct {
		st    protonetmap.NodeInfo_State
		check func(netmap.NodeInfo) bool
	}{
		{st: protonetmap.NodeInfo_ONLINE, check: netmap.NodeInfo.IsOnline},
		{st: protonetmap.NodeInfo_OFFLINE, check: netmap.NodeInfo.IsOffline},
		{st: protonetmap.NodeInfo_MAINTENANCE, check: netmap.NodeInfo.IsMaintenance},
	} {
		m.State = tc.st
		require.NoError(t, val.FromProtoMessage(m), tc.st)
		require.True(t, tc.check(val))
	}

	// reset optional fields
	m.Attributes = nil
	m.State = 0
	val2 := val
	require.NoError(t, val2.FromProtoMessage(m))
	require.Zero(t, val2.NumberOfAttributes())
	val2.IterateAttributes(func(string, string) {
		t.Fatal("handler must not be called")
	})
	for _, checkState := range []func(netmap.NodeInfo) bool{
		netmap.NodeInfo.IsOnline,
		netmap.NodeInfo.IsOffline,
		netmap.NodeInfo.IsMaintenance,
	} {
		require.False(t, checkState(val2))
	}

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(info *protonetmap.NodeInfo)
		}{
			{name: "public key/nil", err: "missing public key",
				corrupt: func(m *protonetmap.NodeInfo) { m.PublicKey = nil }},
			{name: "public key/empty", err: "missing public key",
				corrupt: func(m *protonetmap.NodeInfo) { m.PublicKey = []byte{} }},
			{name: "endpoints/nil", err: "missing network endpoints",
				corrupt: func(m *protonetmap.NodeInfo) { m.Addresses = nil }},
			{name: "endpoints/empty", err: "missing network endpoints",
				corrupt: func(m *protonetmap.NodeInfo) { m.Addresses = []string{} }},
			{name: "attributes/nil", err: "nil attribute #1",
				corrupt: func(m *protonetmap.NodeInfo) { m.Attributes[1] = nil }},
			{name: "attributes/no key", err: "empty key of the attribute #1",
				corrupt: func(m *protonetmap.NodeInfo) { setNodeAttributes(m, "k1", "v1", "", "v2") }},
			{name: "attributes/no value", err: `empty "k2" attribute value`,
				corrupt: func(m *protonetmap.NodeInfo) { setNodeAttributes(m, "k1", "v1", "k2", "") }},
			{name: "attributes/duplicated", err: "duplicated attribute k1",
				corrupt: func(m *protonetmap.NodeInfo) { setNodeAttributes(m, "k1", "v1", "k2", "v2", "k1", "v3") }},
			{name: "attributes/capacity", err: "invalid Capacity attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *protonetmap.NodeInfo) { setNodeAttributes(m, "Capacity", "foo") }},
			{name: "attributes/price", err: "invalid Price attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *protonetmap.NodeInfo) { setNodeAttributes(m, "Price", "foo") }},
			{name: "state/negative", err: "negative state -1",
				corrupt: func(m *protonetmap.NodeInfo) { m.State = -1 }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(netmap.NodeInfo).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestNodeInfo_ProtoMessage(t *testing.T) {
	var val netmap.NodeInfo

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetPublicKey())
	require.Zero(t, m.GetAddresses())
	require.Zero(t, m.GetAttributes())
	require.Zero(t, m.GetState())

	// filled
	m = validNodeInfo.ProtoMessage()
	require.Equal(t, anyValidPublicKey, m.GetPublicKey())
	require.Equal(t, anyValidNetworkEndpoints, m.GetAddresses())

	mas := m.GetAttributes()
	require.Len(t, mas, 14)
	for i, pair := range [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"Price", "10993309018040354285"},
		{"Capacity", "9010937245406684209"},
		{"UN-LOCODE", anyValidLOCODE},
		{"CountryCode", anyValidCountryCode},
		{"Country", anyValidCountryName},
		{"Location", anyValidLocationName},
		{"SubDivCode", anyValidSubdivCode},
		{"SubDiv", anyValidSubdivName},
		{"Continent", anyValidContinentName},
		{"ExternalAddr", strings.Join(anyValidExternalNetworkEndpoints, ",")},
		{"Version", anyValidNodeVersion},
		{"VerifiedNodesDomain", anyValidVerifiedNodesDomain},
	} {
		require.EqualValues(t, pair[0], mas[i].GetKey())
		require.EqualValues(t, pair[1], mas[i].GetValue())
		require.Zero(t, mas[i].GetParents())
	}

	for _, tc := range []struct {
		setState func(*netmap.NodeInfo)
		exp      protonetmap.NodeInfo_State
	}{
		{setState: (*netmap.NodeInfo).SetOnline, exp: protonetmap.NodeInfo_ONLINE},
		{setState: (*netmap.NodeInfo).SetOffline, exp: protonetmap.NodeInfo_OFFLINE},
		{setState: (*netmap.NodeInfo).SetMaintenance, exp: protonetmap.NodeInfo_MAINTENANCE},
	} {
		val2 := validNodeInfo
		tc.setState(&val2)
		m := val2.ProtoMessage()
		require.Equal(t, tc.exp, m.GetState(), tc.exp)
	}
}

func TestNodeInfo_Marshal(t *testing.T) {
	require.Equal(t, validBinNodeInfo, validNodeInfo.Marshal())
}

func TestNodeInfo_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(netmap.NodeInfo).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "attributes/no key", err: "empty key of the attribute #1",
				b: []byte{26, 8, 10, 2, 107, 49, 18, 2, 118, 49, 26, 4, 18, 2, 118, 50}},
			{name: "attributes/no value", err: `empty "k2" attribute value`,
				b: []byte{26, 8, 10, 2, 107, 49, 18, 2, 118, 49, 26, 4, 10, 2, 107, 50}},
			{name: "attributes/duplicated", err: "duplicated attribute k1",
				b: []byte{26, 8, 10, 2, 107, 49, 18, 2, 118, 49, 26, 8, 10, 2, 107, 50, 18, 2, 118, 50, 26, 8, 10, 2, 107, 49,
					18, 2, 118, 51}},
			{name: "attributes/capacity", err: "invalid Capacity attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				b: []byte{26, 15, 10, 8, 67, 97, 112, 97, 99, 105, 116, 121, 18, 3, 102, 111, 111}},
			{name: "attributes/price", err: "invalid Price attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				b: []byte{26, 12, 10, 5, 80, 114, 105, 99, 101, 18, 3, 102, 111, 111}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(netmap.NodeInfo).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var val netmap.NodeInfo
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Zero(t, val.PublicKey())
	require.Zero(t, val.NumberOfNetworkEndpoints())
	val.IterateNetworkEndpoints(func(string) bool { t.Fatal("handler must not be called"); return false })
	require.Zero(t, val.NumberOfAttributes())
	val.IterateAttributes(func(string, string) { t.Fatal("handler must not be called") })
	for _, checkState := range []func(netmap.NodeInfo) bool{
		netmap.NodeInfo.IsOnline,
		netmap.NodeInfo.IsOffline,
		netmap.NodeInfo.IsMaintenance,
	} {
		require.False(t, checkState(val))
	}

	// filled
	require.NoError(t, val.Unmarshal(validBinNodeInfo))
	require.Equal(t, validNodeInfo, val)
}

func TestNodeInfo_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validNodeInfo, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONNodeInfo, string(b))
}

func TestNodeInfo_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(netmap.NodeInfo).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range []struct{ name, err, j string }{
			{name: "attributes/no key", err: "empty key of the attribute #1",
				j: `{"attributes":[{"key":"k1", "value":"v1"}, {"key":"", "value":"v2"}]}`},
			{name: "attributes/no value", err: `empty "k2" attribute value`,
				j: `{"attributes":[{"key":"k1","value":"v1"},{"key":"k2"}]}`},
			{name: "attributes/duplicated", err: "duplicated attribute k1",
				j: `{"attributes":[{"key":"k1","value":"v1"},{"key":"k2","value":"v2"},{"key":"k1","value":"v3"}]}`},
			{name: "attributes/capacity", err: "invalid Capacity attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				j: `{"attributes":[{"key":"Capacity","value":"foo"}]}`},
			{name: "attributes/price", err: "invalid Price attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
				j: `{"attributes":[{"key":"Price","value":"foo"}]}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(netmap.NodeInfo).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var val netmap.NodeInfo
	// zero
	require.NoError(t, val.UnmarshalJSON([]byte("{}")))
	require.Zero(t, val.PublicKey())
	require.Zero(t, val.NumberOfNetworkEndpoints())
	val.IterateNetworkEndpoints(func(string) bool { t.Fatal("handler must not be called"); return false })
	require.Zero(t, val.NumberOfAttributes())
	val.IterateAttributes(func(string, string) { t.Fatal("handler must not be called") })
	for _, checkState := range []func(netmap.NodeInfo) bool{
		netmap.NodeInfo.IsOnline,
		netmap.NodeInfo.IsOffline,
		netmap.NodeInfo.IsMaintenance,
	} {
		require.False(t, checkState(val))
	}

	// filled
	require.NoError(t, val.UnmarshalJSON([]byte(validJSONNodeInfo)))
	require.Equal(t, validNodeInfo, val)
}
