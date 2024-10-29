package object_test

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/tombstone"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

var validTombstone object.Tombstone // set by init.

func init() {
	validTombstone.SetExpirationEpoch(anyValidExpirationEpoch)
	validTombstone.SetSplitID(anyValidSplitID)
	validTombstone.SetMembers(anyValidIDs[:3])
}

// corresponds to validTombstone.
var validBinTombstone = []byte{
	8, 140, 160, 255, 199, 209, 164, 215, 128, 84, 18, 16, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147,
	41, 26, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117,
	173, 70, 246, 8, 139, 247, 174, 53, 60, 26, 34, 10, 32, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214,
	76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 26, 34, 10, 32, 206, 228, 247, 217, 41,
	247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202,
	197, 242,
}

// corresponds to validTombstone.
const validJSONTombstone = `
{
 "expirationEpoch": "6053221788077248524",
 "splitID": "4IQDUCAsRbi5IOLJzsSTKQ==",
 "members": [
  {
   "value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="
  },
  {
   "value": "5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+s="
  },
  {
   "value": "zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="
  }
 ]
}
`

func TestTombstone_ReadFromV2(t *testing.T) {
	ms := make([]refs.ObjectID, len(anyValidIDs))
	for i := range anyValidIDs {
		ms[i].SetValue(anyValidIDs[i][:])
	}

	var m tombstone.Tombstone
	m.SetExpirationEpoch(anyValidExpirationEpoch)
	m.SetSplitID(anyValidSplitIDBytes)
	m.SetMembers(ms)

	var ts object.Tombstone
	require.NoError(t, ts.ReadFromV2(m))
	require.EqualValues(t, anyValidExpirationEpoch, ts.ExpirationEpoch())
	require.Equal(t, anyValidSplitID, ts.SplitID())
	require.Equal(t, anyValidIDs, ts.Members())

	// reset optional fields
	m.SetExpirationEpoch(0)
	m.SetSplitID(nil)
	m.SetMembers(nil)
	ts2 := ts
	require.NoError(t, ts2.ReadFromV2(m))
	require.Zero(t, ts2)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*tombstone.Tombstone)
		}{
			{name: "members/nil value", err: "invalid member #1: invalid length 0",
				corrupt: func(m *tombstone.Tombstone) {
					ms := slices.Clone(ms)
					ms[1] = *protoIDFromBytes(nil)
					m.SetMembers(ms)
				}},
			{name: "members/empty value", err: "invalid member #1: invalid length 0",
				corrupt: func(m *tombstone.Tombstone) {
					ms := slices.Clone(ms)
					ms[1] = *protoIDFromBytes([]byte{})
					m.SetMembers(ms)
				}},
			{name: "members/undersize", err: "invalid member #1: invalid length 31",
				corrupt: func(m *tombstone.Tombstone) {
					ms := slices.Clone(ms)
					ms[1] = *protoIDFromBytes(make([]byte, 31))
					m.SetMembers(ms)
				}},
			{name: "members/oversize", err: "invalid member #1: invalid length 33",
				corrupt: func(m *tombstone.Tombstone) {
					ms := slices.Clone(ms)
					ms[1] = *protoIDFromBytes(make([]byte, 33))
					m.SetMembers(ms)
				}},
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				corrupt: func(m *tombstone.Tombstone) { m.SetSplitID(anyValidSplitIDBytes[:15]) }},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				corrupt: func(m *tombstone.Tombstone) { m.SetSplitID(append(anyValidSplitIDBytes[:], 1)) }},
			{name: "split ID/wrong version", err: "invalid split UUID version 3",
				corrupt: func(m *tombstone.Tombstone) {
					b := bytes.Clone(anyValidSplitIDBytes[:])
					b[6] = 3 << 4
					m.SetSplitID(b)
				}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m := ts.ToV2()
				tc.corrupt(m)
				require.EqualError(t, new(object.Tombstone).ReadFromV2(*m), tc.err)
			})
		}
	})
}

func TestTombstone_WriteToV2(t *testing.T) {
	var ts object.Tombstone

	// zero
	m := ts.ToV2()
	require.Zero(t, m.GetExpirationEpoch())
	require.Zero(t, m.GetSplitID())
	require.Zero(t, m.GetMembers())

	// filled
	m = validTombstone.ToV2()
	require.EqualValues(t, anyValidExpirationEpoch, m.GetExpirationEpoch())
	require.EqualValues(t, anyValidSplitIDBytes, m.GetSplitID())
	ms := m.GetMembers()
	require.Len(t, ms, 3)
	for i := range ms {
		require.Equal(t, anyValidIDs[i][:], ms[i].GetValue())
	}
}

func TestTombstone_Marshal(t *testing.T) {
	require.Equal(t, validBinTombstone, validTombstone.Marshal())
}

func TestContainer_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(object.Tombstone).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "members/empty value", err: "invalid member #1: invalid length 0",
				b: []byte{26, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77,
					44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 26, 0}},
			{name: "members/undersize", err: "invalid member #1: invalid length 31",
				b: []byte{26, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77,
					44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 26, 33, 10, 31, 229, 77, 63, 235, 2, 9, 165, 123,
					116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179}},
			{name: "members/oversize", err: "invalid member #1: invalid length 33",
				b: []byte{26, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77,
					44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 26, 35, 10, 33, 229, 77, 63, 235, 2, 9, 165, 123,
					116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 1}},
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				b: []byte{18, 15, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147}},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				b: []byte{18, 17, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41, 1}},
			{name: "split ID/wrong version", err: "invalid split UUID version 3",
				b: []byte{18, 16, 224, 132, 3, 80, 32, 44, 48, 184, 185, 32, 226, 201, 206, 196, 147, 41}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.Tombstone).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var ts object.Tombstone
	// zero
	require.NoError(t, ts.Unmarshal(nil))
	require.Zero(t, ts)

	// filled
	require.NoError(t, ts.Unmarshal(validBinTombstone))
	require.Equal(t, validTombstone, ts)
}

func TestTombstone_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validTombstone, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONTombstone, string(b))
}

func TestTombstone_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(object.Tombstone).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range []struct{ name, err, j string }{
			{name: "members/empty value", err: "invalid member #1: invalid length 0",
				j: `{"members":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},{"value":""}]}`},
			{name: "members/undersize", err: "invalid member #1: invalid length 31",
				j: `{"members":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, {"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/sw=="}]}`},
			{name: "members/oversize", err: "invalid member #1: invalid length 33",
				j: `{"members":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, {"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+sB"}]}`},
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				j: `{"splitID":"4IQDUCAsRbi5IOLJzsST"}`},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				j: `{"splitID":"4IQDUCAsRbi5IOLJzsSTKQE="}`},
			{name: "split ID/wrong version", err: "invalid split UUID version 3",
				j: `{"splitID":"4IQDUCAsMLi5IOLJzsSTKQ=="}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.Tombstone).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var ts object.Tombstone
	// zero
	require.NoError(t, ts.UnmarshalJSON([]byte("{}")))
	require.Zero(t, ts)

	// filled
	require.NoError(t, ts.UnmarshalJSON([]byte(validJSONTombstone)))
	require.Equal(t, validTombstone, ts)
}

func TestTombstone_SetExpirationEpoch(t *testing.T) {
	var ts object.Tombstone
	require.Zero(t, ts.ExpirationEpoch())

	ts.SetExpirationEpoch(anyValidExpirationEpoch)
	require.EqualValues(t, anyValidExpirationEpoch, ts.ExpirationEpoch())

	ts.SetExpirationEpoch(anyValidExpirationEpoch + 1)
	require.EqualValues(t, anyValidExpirationEpoch+1, ts.ExpirationEpoch())
}

func TestTombstone_SetSplitID(t *testing.T) {
	var ts object.Tombstone
	require.Zero(t, ts.SplitID())

	ts.SetSplitID(anyValidSplitID)
	require.Equal(t, anyValidSplitID, ts.SplitID())

	b := bytes.Clone(anyValidSplitIDBytes)
	b[0]++
	otherSplitID := object.NewSplitIDFromV2(b)
	ts.SetSplitID(otherSplitID)
	require.Equal(t, otherSplitID, ts.SplitID())
}

func TestTombstone_SetMembers(t *testing.T) {
	var ts object.Tombstone
	require.Zero(t, ts.Members())

	ts.SetMembers(anyValidIDs)
	require.Equal(t, anyValidIDs, ts.Members())

	otherIDs := oidtest.IDs(3)
	ts.SetMembers(otherIDs)
	require.Equal(t, otherIDs, ts.Members())
}

func TestNewTombstoneFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *tombstone.Tombstone

		require.Nil(t, object.NewTombstoneFromV2(x))
	})
}

func TestNewTombstone(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		ts := object.NewTombstone()

		// check initial values
		require.Nil(t, ts.SplitID())
		require.Nil(t, ts.Members())
		require.Zero(t, ts.ExpirationEpoch())

		// convert to v2 message
		tsV2 := ts.ToV2()

		require.Nil(t, tsV2.GetSplitID())
		require.Nil(t, tsV2.GetMembers())
		require.Zero(t, tsV2.GetExpirationEpoch())
	})
}
