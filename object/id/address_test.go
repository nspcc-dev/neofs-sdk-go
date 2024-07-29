package oid_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

var validContainerBytes = [cid.Size]byte{198, 173, 193, 21, 128, 250, 249, 109, 30, 208, 188, 153, 209, 195, 177, 12, 188,
	163, 191, 204, 15, 238, 54, 211, 39, 92, 211, 250, 136, 69, 113, 146}

// corresponds to validContainerBytes.
const validContainerString = "ENZPp185BvQRhnyeXJxteFC4rkC6a25Kb133KbtcVw1B"

const validAddressString = validContainerString + "/" + validIDString

// corresponds to validIDBytes and validContainerBytes.
const validAddressJSON = `{"containerID":{"value":"xq3BFYD6+W0e0LyZ0cOxDLyjv8wP7jbTJ1zT+ohFcZI="},"objectID":{"value":"5715B62G/qU/ujxZIV8uZ9k5pFdSzPviAWQgSPsAB6w="}}`

var invalidAddressTestcases = []struct {
	name    string
	err     string
	corrupt func(*refs.Address)
}{
	{name: "empty", err: "missing container ID", corrupt: func(a *refs.Address) {
		a.SetContainerID(nil)
		a.SetObjectID(nil)
	}},
	{name: "container/missing", err: "missing container ID", corrupt: func(a *refs.Address) { a.SetContainerID(nil) }},
	{name: "container/nil value", err: "invalid container ID: invalid length 0", corrupt: func(a *refs.Address) {
		a.SetContainerID(new(refs.ContainerID))
	}},
	{name: "container/empty value", err: "invalid container ID: invalid length 0", corrupt: func(a *refs.Address) {
		var m refs.ContainerID
		m.SetValue([]byte{})
		a.SetContainerID(&m)
	}},
	{name: "container/undersize", err: "invalid container ID: invalid length 31", corrupt: func(a *refs.Address) {
		var m refs.ContainerID
		m.SetValue(make([]byte, 31))
		a.SetContainerID(&m)
	}},
	{name: "container/oversize", err: "invalid container ID: invalid length 33", corrupt: func(a *refs.Address) {
		var m refs.ContainerID
		m.SetValue(make([]byte, 33))
		a.SetContainerID(&m)
	}},
	{name: "object/missing", err: "missing object ID", corrupt: func(a *refs.Address) { a.SetObjectID(nil) }},
	{name: "object/nil value", err: "invalid object ID: invalid length 0", corrupt: func(a *refs.Address) {
		a.SetObjectID(new(refs.ObjectID))
	}},
	{name: "object/empty value", err: "invalid object ID: invalid length 0", corrupt: func(a *refs.Address) {
		var m refs.ObjectID
		m.SetValue([]byte{})
		a.SetObjectID(&m)
	}},
	{name: "object/undersize", err: "invalid object ID: invalid length 31", corrupt: func(a *refs.Address) {
		var m refs.ObjectID
		m.SetValue(make([]byte, 31))
		a.SetObjectID(&m)
	}},
	{name: "object/oversize", err: "invalid object ID: invalid length 33", corrupt: func(a *refs.Address) {
		var m refs.ObjectID
		m.SetValue(make([]byte, 33))
		a.SetObjectID(&m)
	}},
}

func TestNewAddress(t *testing.T) {
	cnr := cidtest.ID()
	obj := oidtest.ID()
	a := oid.NewAddress(cnr, obj)
	require.Equal(t, cnr, a.Container())
	require.Equal(t, obj, a.Object())
}

func testAddressIDField[T ~[32]byte](
	t *testing.T,
	randFunc func(...T) T,
	get func(oid.Address) T,
	set func(*oid.Address, T),
	getAPI func(*refs.Address) []byte,
) {
	var a oid.Address
	require.Zero(t, get(a))

	val := randFunc()
	set(&a, val)
	require.Equal(t, val, get(a))
	valOther := randFunc(val)
	set(&a, valOther)
	require.Equal(t, valOther, get(a))

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := oidtest.Address()
			var dst oid.Address
			var msg refs.Address

			set(&src, val)
			src.WriteToV2(&msg)
			require.EqualValues(t, val[:], getAPI(&msg))
			require.NoError(t, dst.ReadFromV2(msg))
			require.Equal(t, val, get(dst))
		})
		t.Run("json", func(t *testing.T) {
			src := oidtest.Address()
			var dst oid.Address

			set(&src, val)
			b, err := src.MarshalJSON()
			require.NoError(t, err)
			require.NoError(t, dst.UnmarshalJSON(b))
			require.Equal(t, val, get(dst))
		})
	})
}

func TestAddress_SetContainer(t *testing.T) {
	testAddressIDField(t, cidtest.OtherID, oid.Address.Container, (*oid.Address).SetContainer, func(m *refs.Address) []byte {
		return m.GetContainerID().GetValue()
	})
}

func TestAddress_SetObject(t *testing.T) {
	testAddressIDField(t, oidtest.OtherID, oid.Address.Object, (*oid.Address).SetObject, func(m *refs.Address) []byte {
		return m.GetObjectID().GetValue()
	})
}

func TestAddress_ReadFromV2(t *testing.T) {
	var mc refs.ContainerID
	mc.SetValue(validContainerBytes[:])
	var mo refs.ObjectID
	mo.SetValue(validIDBytes[:])
	var m refs.Address
	m.SetContainerID(&mc)
	m.SetObjectID(&mo)
	var a oid.Address
	require.NoError(t, a.ReadFromV2(m))
	require.EqualValues(t, validContainerBytes, a.Container())
	require.EqualValues(t, validIDBytes, a.Object())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidAddressTestcases {
			t.Run(tc.name, func(t *testing.T) {
				var m refs.Address
				oidtest.Address().WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(oid.Address).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestAddress_EncodeToString(t *testing.T) {
	a := oid.NewAddress(validContainerBytes, validIDBytes)
	require.Equal(t, validAddressString, a.EncodeToString())
}

func TestAddress_DecodeString(t *testing.T) {
	var a oid.Address
	require.NoError(t, a.DecodeString(validAddressString))
	require.EqualValues(t, validContainerBytes, a.Container())
	require.EqualValues(t, validIDBytes, a.Object())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			str      string
			contains bool
			err      string
		}{
			{name: "empty", str: "", err: "missing delimiter"},
			{name: "delimiter", str: "/", contains: true, err: "decode container string: decode base58"},
			{name: "delimiter prefix", str: "/" + validAddressString, contains: true, err: "decode container string: decode base58"},
			{name: "delimiter suffix", str: validAddressString + "/", contains: true, err: "decode object string: decode base58"},
			{name: "missing delimiter", str: strings.ReplaceAll(validAddressString, "/", ""), err: "missing delimiter"},
			{name: "container/missing", str: "/" + validIDString, contains: true,
				err: "decode container string: decode base58"},
			{name: "container/invalid base58", str: "ENZPp185BvQRhnyeXJxteFC4rkC6a25Kb133KbtcVw1_/" + validIDString, contains: true,
				err: "decode container string: decode base58"},
			{name: "container/undersize", str: "3RArVxBNPE4rZ9f5oHwxTFi7LTSY1fQ3BzNJZat2ZoV/" + validIDString,
				err: "decode container string: invalid length 31"},
			{name: "container/oversize", str: "CJ1rzsceKvtmKtZcuEJssiLVqDgBc6rPp5dwxhNxChbap/" + validIDString,
				err: "decode container string: invalid length 33"},
			{name: "object/missing", str: validContainerString + "/", contains: true,
				err: "decode object string: decode base58"},
			{name: "object/invalid base58", str: validContainerString + "/64sexUVdHg9dnqxHTB3uGsHfhPTnuegrupF4bJGXSbZ_", contains: true,
				err: "decode object string: decode base58"},
			{name: "object/undersize", str: validContainerString + "/tjkeRJWKpNa7Xjh3sZymJWZG5UGTGoZtEimGUjQesV",
				err: "decode object string: invalid length 31"},
			{name: "object/oversize", str: validContainerString + "/2Jwe3Yy4t5hgV2VeXxPkxYV5GNwPJEBjD3oZFDUiHneQ7J",
				err: "decode object string: invalid length 33"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				err := new(oid.Address).DecodeString(tc.str)
				_, err2 := oid.DecodeAddressString(tc.str)
				if tc.contains {
					require.ErrorContains(t, err, tc.err)
					require.ErrorContains(t, err2, tc.err)
				} else {
					require.EqualError(t, err, tc.err)
					require.EqualError(t, err2, tc.err)
				}
			})
		}
	})
}

func TestAddressComparable(t *testing.T) {
	x := oidtest.Address()
	y := x
	require.True(t, x == y)
	require.False(t, x != y)
	y = oidtest.OtherAddress(x)
	require.False(t, x == y)
	require.True(t, x != y)
}

func TestAddress_MarshalJSON(t *testing.T) {
	a := oid.NewAddress(validContainerBytes, validIDBytes)
	b, err := a.MarshalJSON()
	require.NoError(t, err)
	// sometimes b != validAddressJSON https://github.com/golang/protobuf/issues/1121
	var a2 oid.Address
	require.NoError(t, json.Unmarshal(b, &a2))
	require.Equal(t, a, a2)
}

func TestAddress_UnmarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := new(oid.Address).UnmarshalJSON([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var a oid.Address
	require.NoError(t, a.UnmarshalJSON([]byte(validAddressJSON)))
	require.EqualValues(t, validContainerBytes, a.Container())
	require.EqualValues(t, validIDBytes, a.Object())

	t.Run("protocol violation", func(t *testing.T) {
		for _, tc := range invalidAddressTestcases {
			t.Run(tc.name, func(t *testing.T) {
				var m refs.Address
				oidtest.Address().WriteToV2(&m)
				tc.corrupt(&m)
				b, err := m.MarshalJSON()
				require.NoError(t, err)
				require.EqualError(t, new(oid.Address).UnmarshalJSON(b), tc.err)
			})
		}
	})
}

func TestAddress_String(t *testing.T) {
	a := oidtest.Address()
	require.NotEmpty(t, a.String())
	require.Equal(t, a.String(), a.String())
	require.NotEqual(t, a.String(), oidtest.OtherAddress(a).String())
}
