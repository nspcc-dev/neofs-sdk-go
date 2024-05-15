package netmap_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNodeInfo_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("public key", func(t *testing.T) {
			n := netmaptest.NodeInfo()
			var m apinetmap.NodeInfo

			n.WriteToV2(&m)
			m.PublicKey = nil
			require.ErrorContains(t, n.ReadFromV2(&m), "missing public key")
			m.PublicKey = []byte{}
			require.ErrorContains(t, n.ReadFromV2(&m), "missing public key")
		})
		t.Run("network endpoints", func(t *testing.T) {
			n := netmaptest.NodeInfo()
			var m apinetmap.NodeInfo

			n.WriteToV2(&m)
			m.Addresses = nil
			require.ErrorContains(t, n.ReadFromV2(&m), "missing network endpoints")
			m.Addresses = []string{}
			require.ErrorContains(t, n.ReadFromV2(&m), "missing network endpoints")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("network endpoints", func(t *testing.T) {
			n := netmaptest.NodeInfo()
			var m apinetmap.NodeInfo

			n.WriteToV2(&m)
			m.Addresses = []string{"any", "", "any"}
			require.ErrorContains(t, n.ReadFromV2(&m), "empty network endpoint #1")
		})
		t.Run("attributes", func(t *testing.T) {
			t.Run("missing key", func(t *testing.T) {
				n := netmaptest.NodeInfo()
				var m apinetmap.NodeInfo

				n.WriteToV2(&m)
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "key_valid", Value: "any"},
					{Key: "", Value: "any"},
				}
				require.ErrorContains(t, n.ReadFromV2(&m), "invalid attribute #1: missing key")
			})
			t.Run("repeated keys", func(t *testing.T) {
				n := netmaptest.NodeInfo()
				var m apinetmap.NodeInfo

				n.WriteToV2(&m)
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "k1", Value: "any"},
					{Key: "k2", Value: "1"},
					{Key: "k3", Value: "any"},
					{Key: "k2", Value: "2"},
				}
				require.ErrorContains(t, n.ReadFromV2(&m), "multiple attributes with key=k2")
			})
			t.Run("missing value", func(t *testing.T) {
				n := netmaptest.NodeInfo()
				var m apinetmap.NodeInfo

				n.WriteToV2(&m)
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "key1", Value: "any"},
					{Key: "key2", Value: ""},
				}
				require.ErrorContains(t, n.ReadFromV2(&m), "invalid attribute #1 (key2): missing value")
			})
			t.Run("price format", func(t *testing.T) {
				n := netmaptest.NodeInfo()
				var m apinetmap.NodeInfo

				n.WriteToV2(&m)
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "any", Value: "any"},
					{Key: "Price", Value: "not_a_number"},
				}
				require.ErrorContains(t, n.ReadFromV2(&m), "invalid price attribute (#1): invalid integer")
			})
			t.Run("capacity format", func(t *testing.T) {
				n := netmaptest.NodeInfo()
				var m apinetmap.NodeInfo

				n.WriteToV2(&m)
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "any", Value: "any"},
					{Key: "Capacity", Value: "not_a_number"},
				}
				require.ErrorContains(t, n.ReadFromV2(&m), "invalid capacity attribute (#1): invalid integer")
			})
		})
	})
}

func TestNodeInfo_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var n netmap.NodeInfo
		msg := []byte("definitely_not_protobuf")
		err := n.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("network endpoints", func(t *testing.T) {
			var n netmap.NodeInfo
			var m apinetmap.NodeInfo
			m.Addresses = []string{"any", "", "any"}
			b, err := proto.Marshal(&m)
			require.NoError(t, err)
			require.ErrorContains(t, n.Unmarshal(b), "empty network endpoint #1")
		})
		t.Run("attributes", func(t *testing.T) {
			t.Run("missing key", func(t *testing.T) {
				var n netmap.NodeInfo
				var m apinetmap.NodeInfo
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "key_valid", Value: "any"},
					{Key: "", Value: "any"},
				}
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, n.Unmarshal(b), "invalid attribute #1: missing key")
			})
			t.Run("repeated keys", func(t *testing.T) {
				var n netmap.NodeInfo
				var m apinetmap.NodeInfo
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "k1", Value: "any"},
					{Key: "k2", Value: "1"},
					{Key: "k3", Value: "any"},
					{Key: "k2", Value: "2"},
				}
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, n.Unmarshal(b), "multiple attributes with key=k2")
			})
			t.Run("missing value", func(t *testing.T) {
				var n netmap.NodeInfo
				var m apinetmap.NodeInfo
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "key1", Value: "any"},
					{Key: "key2", Value: ""},
				}
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, n.Unmarshal(b), "invalid attribute #1 (key2): missing value")
			})
			t.Run("price", func(t *testing.T) {
				var n netmap.NodeInfo
				var m apinetmap.NodeInfo
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "any", Value: "any"},
					{Key: "Price", Value: "not_a_number"},
				}
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, n.Unmarshal(b), "invalid price attribute (#1): invalid integer")
			})
			t.Run("capacity", func(t *testing.T) {
				var n netmap.NodeInfo
				var m apinetmap.NodeInfo
				m.Attributes = []*apinetmap.NodeInfo_Attribute{
					{Key: "any", Value: "any"},
					{Key: "Capacity", Value: "not_a_number"},
				}
				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, n.Unmarshal(b), "invalid capacity attribute (#1): invalid integer")
			})
		})
	})
}

func TestNodeInfo_UnmarshalJSON(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		var n netmap.NodeInfo
		msg := []byte("definitely_not_protojson")
		err := n.UnmarshalJSON(msg)
		require.ErrorContains(t, err, "decode protojson")
	})
	t.Run("invalid fields", func(t *testing.T) {
		testCases := []struct {
			name string
			err  string
			json string
		}{{name: "empty network endpoint", err: "empty network endpoint #1", json: `
{
  "addresses": ["any", "", "any"]
}`},
			{name: "attributes/missing key", err: "invalid attribute #1: missing key", json: `
{
  "attributes": [
    {"key": "key_valid","value": "any"},
    {"key": "","value": "any"}
  ]
}`},
			{name: "attributes/repeated keys", err: "multiple attributes with key=k2", json: `
{
  "attributes": [
    {"key": "k1","value": "any"},
    {"key": "k2","value": "1"},
    {"key": "k3","value": "any"},
    {"key": "k2","value": "2"}
  ]
}`},
			{name: "attributes/missing value", err: "invalid attribute #1 (key2): missing value", json: `
{
  "attributes": [
    {"key": "key1","value": "any"},
    {"key": "key2","value": ""}
  ]
}`},
			{name: "attributes/price", err: "invalid price attribute (#1): invalid integer", json: `
{
  "attributes": [
    {"key": "any","value": "any"},
    {"key": "Price","value": "not_a_number"}
  ]
}`},
			{name: "attributes/capacity", err: "invalid capacity attribute (#1): invalid integer", json: `
{
  "attributes": [
    {"key": "any","value": "any"},
    {"key": "Capacity","value": "not_a_number"}
  ]
}`},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				var n netmap.NodeInfo
				require.ErrorContains(t, n.UnmarshalJSON([]byte(testCase.json)), testCase.err)
			})
		}
	})
}

func TestNodeInfo_SortAttributes(t *testing.T) {
	var n netmap.NodeInfo
	const a1, a2, a3 = "a1", "a2", "a3"
	require.Less(t, a1, a2)
	require.Less(t, a2, a3)

	// set unordered
	n.SetAttribute(a3, a3)
	n.SetAttribute(a1, a1)
	n.SetAttribute(a2, a2)

	n.SortAttributes()

	b := n.Marshal()
	var m apinetmap.NodeInfo
	require.NoError(t, proto.Unmarshal(b, &m))
	n.WriteToV2(&m)
	require.Equal(t, []*apinetmap.NodeInfo_Attribute{
		{Key: a1, Value: a1},
		{Key: a2, Value: a2},
		{Key: a3, Value: a3},
	}, m.Attributes)
}

func collectNodeAttributes(n netmap.NodeInfo) [][2]string {
	var res [][2]string
	n.IterateAttributes(func(key, value string) {
		res = append(res, [2]string{key, value})
	})
	return res
}

func TestNodeInfo_SetAttribute(t *testing.T) {
	var n netmap.NodeInfo
	require.Panics(t, func() { n.SetAttribute("", "") })
	require.Panics(t, func() { n.SetAttribute("", "val") })
	require.Panics(t, func() { n.SetAttribute("key", "") })

	const key1, val1 = "some_key1", "some_value1"
	const key2, val2 = "some_key2", "some_value2"

	require.Zero(t, n.Attribute(key1))
	require.Zero(t, n.Attribute(key2))
	require.Zero(t, n.NumberOfAttributes())
	require.Zero(t, collectNodeAttributes(n))

	n.SetAttribute(key1, val1)
	n.SetAttribute(key2, val2)
	require.Equal(t, val1, n.Attribute(key1))
	require.Equal(t, val2, n.Attribute(key2))
	require.EqualValues(t, 2, n.NumberOfAttributes())
	attrs := collectNodeAttributes(n)
	require.Len(t, attrs, 2)
	require.Contains(t, attrs, [2]string{key1, val1})
	require.Contains(t, attrs, [2]string{key2, val2})

	n.SetAttribute(key1, val2)
	n.SetAttribute(key2, val1)
	require.Equal(t, val2, n.Attribute(key1))
	require.Equal(t, val1, n.Attribute(key2))
	attrs = collectNodeAttributes(n)
	require.Len(t, attrs, 2)
	require.Contains(t, attrs, [2]string{key1, val2})
	require.Contains(t, attrs, [2]string{key2, val1})

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetAttribute(key1+key2, val1+val2)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.Attribute(key1))
			require.Zero(t, dst.Attribute(key2))
			require.Zero(t, dst.NumberOfAttributes())
			require.Zero(t, collectNodeAttributes(dst))

			src.SetAttribute(key1, val1)
			src.SetAttribute(key2, val2)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, val1, dst.Attribute(key1))
			require.Equal(t, val2, dst.Attribute(key2))
			require.EqualValues(t, 2, dst.NumberOfAttributes())
			attrs := collectNodeAttributes(dst)
			require.Len(t, attrs, 2)
			require.Contains(t, attrs, [2]string{key1, val1})
			require.Contains(t, attrs, [2]string{key2, val2})
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NodeInfo
			var msg apinetmap.NodeInfo

			// set required data just to satisfy decoder
			src.SetPublicKey([]byte("any"))
			src.SetNetworkEndpoints([]string{"any"})

			dst.SetAttribute(key1, val1)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.Attribute(key1))
			require.Zero(t, dst.Attribute(key2))
			require.Zero(t, dst.NumberOfAttributes())
			require.Zero(t, collectNodeAttributes(dst))

			src.SetAttribute(key1, val1)
			src.SetAttribute(key2, val2)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.NodeInfo_Attribute{
				{Key: key1, Value: val1},
				{Key: key2, Value: val2},
			}, msg.Attributes)

			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, val1, dst.Attribute(key1))
			require.Equal(t, val2, dst.Attribute(key2))
			require.EqualValues(t, 2, dst.NumberOfAttributes())
			attrs := collectNodeAttributes(dst)
			require.Len(t, attrs, 2)
			require.Contains(t, attrs, [2]string{key1, val1})
			require.Contains(t, attrs, [2]string{key2, val2})
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetAttribute(key1, val1)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.Attribute(key1))
			require.Zero(t, dst.Attribute(key2))
			require.Zero(t, dst.NumberOfAttributes())
			require.Zero(t, collectNodeAttributes(dst))

			src.SetAttribute(key1, val1)
			src.SetAttribute(key2, val2)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, val1, dst.Attribute(key1))
			require.Equal(t, val2, dst.Attribute(key2))
			require.EqualValues(t, 2, dst.NumberOfAttributes())
			attrs := collectNodeAttributes(dst)
			require.Len(t, attrs, 2)
			require.Contains(t, attrs, [2]string{key1, val1})
			require.Contains(t, attrs, [2]string{key2, val2})
		})
	})
}

func testNodeInfoState(t *testing.T, get func(netmap.NodeInfo) bool, set func(*netmap.NodeInfo), apiVal apinetmap.NodeInfo_State) {
	var n netmap.NodeInfo
	require.False(t, get(n))
	set(&n)
	require.True(t, get(n))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			set(&dst)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, get(dst))

			set(&src)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, get(dst))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NodeInfo
			var msg apinetmap.NodeInfo

			// set required data just to satisfy decoder
			src.SetPublicKey([]byte("any"))
			src.SetNetworkEndpoints([]string{"any"})

			set(&dst)

			src.WriteToV2(&msg)
			require.Zero(t, msg.State)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.False(t, get(dst))

			set(&src)

			src.WriteToV2(&msg)
			require.Equal(t, apiVal, msg.State)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.True(t, get(dst))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			set(&dst)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, get(dst))

			set(&src)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, get(dst))
		})
	})
}

func TestNodeInfoState(t *testing.T) {
	t.Run("online", func(t *testing.T) {
		testNodeInfoState(t, netmap.NodeInfo.IsOnline, (*netmap.NodeInfo).SetOnline, apinetmap.NodeInfo_ONLINE)
	})
	t.Run("offline", func(t *testing.T) {
		testNodeInfoState(t, netmap.NodeInfo.IsOffline, (*netmap.NodeInfo).SetOffline, apinetmap.NodeInfo_OFFLINE)
	})
	t.Run("maintenance", func(t *testing.T) {
		testNodeInfoState(t, netmap.NodeInfo.IsMaintenance, (*netmap.NodeInfo).SetMaintenance, apinetmap.NodeInfo_MAINTENANCE)
	})
}

func TestNodeInfo_SetPublicKey(t *testing.T) {
	var n netmap.NodeInfo

	require.Zero(t, n.PublicKey())

	key := []byte("any_public_key")
	n.SetPublicKey(key)
	require.Equal(t, key, n.PublicKey())

	keyOther := append(key, "_other"...)
	n.SetPublicKey(keyOther)
	require.Equal(t, keyOther, n.PublicKey())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetPublicKey(keyOther)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.PublicKey())

			src.SetPublicKey(key)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, key, dst.PublicKey())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NodeInfo
			var msg apinetmap.NodeInfo

			// set required data just to satisfy decoder
			src.SetNetworkEndpoints([]string{"any"})

			src.SetPublicKey(key)

			src.WriteToV2(&msg)
			require.Equal(t, key, msg.PublicKey)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, key, dst.PublicKey())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetPublicKey(keyOther)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.PublicKey())

			src.SetPublicKey(key)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, key, dst.PublicKey())
		})
	})
}

func TestNodeInfo_SetNetworkEndpoints(t *testing.T) {
	var n netmap.NodeInfo

	require.Zero(t, n.NetworkEndpoints())

	endpoints := []string{"endpoint1", "endpoint2"}
	n.SetNetworkEndpoints(endpoints)
	require.Equal(t, endpoints, n.NetworkEndpoints())

	endpointsOther := []string{"endpoint3", "endpoint4", "endpoint5"}
	n.SetNetworkEndpoints(endpointsOther)
	require.Equal(t, endpointsOther, n.NetworkEndpoints())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetNetworkEndpoints(endpointsOther)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.NetworkEndpoints())

			src.SetNetworkEndpoints(endpoints)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, endpoints, dst.NetworkEndpoints())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NodeInfo
			var msg apinetmap.NodeInfo

			// set required data just to satisfy decoder
			src.SetPublicKey([]byte("any"))

			src.SetNetworkEndpoints(endpoints)

			src.WriteToV2(&msg)
			require.Equal(t, endpoints, msg.Addresses)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, endpoints, dst.NetworkEndpoints())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetNetworkEndpoints(endpointsOther)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.PublicKey())

			src.SetNetworkEndpoints(endpoints)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, endpoints, dst.NetworkEndpoints())
		})
	})
}

func TestNodeInfo_SetExternalAddresses(t *testing.T) {
	var n netmap.NodeInfo

	require.Zero(t, n.ExternalAddresses())
	require.Panics(t, func() { n.SetExternalAddresses(nil) })
	require.Panics(t, func() { n.SetExternalAddresses([]string{}) })

	addrs := []string{"addr1", "addr2", "addr3"}
	n.SetExternalAddresses(addrs)
	require.Equal(t, addrs, n.ExternalAddresses())

	addrsOther := []string{"addr4", "addr5"}
	n.SetExternalAddresses(addrsOther)
	require.Equal(t, addrsOther, n.ExternalAddresses())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetExternalAddresses(addrsOther)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.ExternalAddresses())

			src.SetExternalAddresses(addrs)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, addrs, dst.ExternalAddresses())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NodeInfo
			var msg apinetmap.NodeInfo

			// set required data just to satisfy decoder
			src.SetPublicKey([]byte("any"))
			src.SetNetworkEndpoints([]string{"any"})

			dst.SetExternalAddresses(addrsOther)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst.ExternalAddresses())

			src.SetExternalAddresses(addrs)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.NodeInfo_Attribute{
				{Key: "ExternalAddr", Value: "addr1,addr2,addr3"},
			}, msg.Attributes)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, addrs, dst.ExternalAddresses())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			dst.SetExternalAddresses(addrsOther)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.ExternalAddresses())

			src.SetExternalAddresses(addrs)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, addrs, dst.ExternalAddresses())
		})
	})
}

func testNodeAttribute[Type uint64 | string](t *testing.T, get func(netmap.NodeInfo) Type, set func(*netmap.NodeInfo, Type), apiAttr string,
	rand func() (_ Type, api string)) {
	var n netmap.NodeInfo

	require.Zero(t, get(n))

	val, apiVal := rand()
	set(&n, val)
	require.EqualValues(t, val, get(n))

	valOther, _ := rand()
	set(&n, valOther)
	require.EqualValues(t, valOther, get(n))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			set(&dst, val)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, get(dst))

			set(&src, val)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NodeInfo
			var msg apinetmap.NodeInfo

			// set required data just to satisfy decoder
			src.SetPublicKey([]byte("any"))
			src.SetNetworkEndpoints([]string{"any"})

			set(&dst, val)

			src.WriteToV2(&msg)
			require.Zero(t, msg.Attributes)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, get(dst))

			set(&src, val)

			src.WriteToV2(&msg)
			require.Equal(t, []*apinetmap.NodeInfo_Attribute{
				{Key: apiAttr, Value: apiVal},
			}, msg.Attributes)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst netmap.NodeInfo

			set(&dst, val)

			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.Version())

			set(&src, val)

			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
	})
}

func testNodeAttributeNum(t *testing.T, get func(netmap.NodeInfo) uint64, set func(*netmap.NodeInfo, uint64), apiAttr string) {
	testNodeAttribute(t, get, set, apiAttr, func() (uint64, string) {
		n := rand.Uint64()
		return n, strconv.FormatUint(n, 10)
	})
}

func testNodeAttributeString(t *testing.T, get func(netmap.NodeInfo) string, set func(*netmap.NodeInfo, string), apiAttr string) {
	testNodeAttribute(t, get, set, apiAttr, func() (string, string) {
		s := fmt.Sprintf("str_%d", rand.Uint64())
		return s, s
	})
}

func TestNodeInfo_SetCapacity(t *testing.T) {
	testNodeAttributeNum(t, netmap.NodeInfo.Capacity, (*netmap.NodeInfo).SetCapacity, "Capacity")
}

func TestNodeInfo_SetPrice(t *testing.T) {
	testNodeAttributeNum(t, netmap.NodeInfo.Price, (*netmap.NodeInfo).SetPrice, "Price")
}

func TestNodeInfo_SetVersion(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.Version, (*netmap.NodeInfo).SetVersion, "Version")
}

func TestNodeInfo_SetLOCODE(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.LOCODE, (*netmap.NodeInfo).SetLOCODE, "UN-LOCODE")
}

func TestNodeInfo_SetCountryCode(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.CountryCode, (*netmap.NodeInfo).SetCountryCode, "CountryCode")
}

func TestNodeInfo_SetCountryName(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.CountryName, (*netmap.NodeInfo).SetCountryName, "Country")
}

func TestNodeInfo_SetLocationName(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.LocationName, (*netmap.NodeInfo).SetLocationName, "Location")
}

func TestNodeInfo_SetSubdivisionCode(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.SubdivisionCode, (*netmap.NodeInfo).SetSubdivisionCode, "SubDivCode")
}

func TestNodeInfo_SetSubdivisionName(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.SubdivisionName, (*netmap.NodeInfo).SetSubdivisionName, "SubDiv")
}

func TestNodeInfo_SetContinentName(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.ContinentName, (*netmap.NodeInfo).SetContinentName, "Continent")
}

func TestNodeInfo_SetVerifiedNodesDomain(t *testing.T) {
	testNodeAttributeString(t, netmap.NodeInfo.VerifiedNodesDomain, (*netmap.NodeInfo).SetVerifiedNodesDomain, "VerifiedNodesDomain")
}
