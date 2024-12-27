package object_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	prototombstone "github.com/nspcc-dev/neofs-sdk-go/proto/tombstone"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

func TestTombstone_FromProtoMessage(t *testing.T) {
	ms := make([]*refs.ObjectID, len(anyValidIDs))
	for i := range anyValidIDs {
		ms[i] = protoIDFromBytes(anyValidIDs[i][:])
	}

	m := &prototombstone.Tombstone{
		ExpirationEpoch: anyValidExpirationEpoch,
		SplitId:         anyValidSplitIDBytes,
		Members:         ms,
	}

	var ts object.Tombstone
	require.NoError(t, ts.FromProtoMessage(m))
	require.EqualValues(t, anyValidExpirationEpoch, ts.ExpirationEpoch())
	require.Equal(t, anyValidSplitID, ts.SplitID())
	require.Equal(t, anyValidIDs, ts.Members())

	// reset optional fields
	m.ExpirationEpoch = 0 //nolint:staticcheck // must be tested still
	m.SplitId = nil
	m.Members = nil
	ts2 := ts
	require.NoError(t, ts2.FromProtoMessage(m))
	require.Zero(t, ts2)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*prototombstone.Tombstone)
		}{
			{name: "members/nil value", err: "invalid member #1: invalid length 0",
				corrupt: func(m *prototombstone.Tombstone) {
					m.Members[1].Value = nil
				}},
			{name: "members/empty value", err: "invalid member #1: invalid length 0",
				corrupt: func(m *prototombstone.Tombstone) {
					m.Members[1].Value = []byte{}
				}},
			{name: "members/undersize", err: "invalid member #1: invalid length 31",
				corrupt: func(m *prototombstone.Tombstone) {
					m.Members[1].Value = make([]byte, 31)
				}},
			{name: "members/oversize", err: "invalid member #1: invalid length 33",
				corrupt: func(m *prototombstone.Tombstone) {
					m.Members[1].Value = make([]byte, 33)
				}},
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				corrupt: func(m *prototombstone.Tombstone) { m.SplitId = anyValidSplitIDBytes[:15] }},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				corrupt: func(m *prototombstone.Tombstone) { m.SplitId = append(anyValidSplitIDBytes[:], 1) }},
			{name: "split ID/wrong version", err: "invalid split ID: wrong UUID version 3, expected 4",
				corrupt: func(m *prototombstone.Tombstone) {
					m.SplitId = bytes.Clone(anyValidSplitIDBytes[:])
					m.SplitId[6] = 3 << 4
				}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m := proto.Clone(ts.ProtoMessage()).(*prototombstone.Tombstone)
				tc.corrupt(m)
				require.EqualError(t, new(object.Tombstone).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestTombstone_ProtoMessage(t *testing.T) {
	var ts object.Tombstone

	// zero
	m := ts.ProtoMessage()
	require.Zero(t, m.GetExpirationEpoch()) //nolint:staticcheck // must be tested still
	require.Zero(t, m.GetSplitId())
	require.Zero(t, m.GetMembers())

	// filled
	m = validTombstone.ProtoMessage()
	require.EqualValues(t, anyValidExpirationEpoch, m.GetExpirationEpoch()) //nolint:staticcheck // must be tested still
	require.EqualValues(t, anyValidSplitIDBytes, m.GetSplitId())
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
			{name: "split ID/wrong version", err: "invalid split ID: wrong UUID version 3, expected 4",
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
			{name: "split ID/wrong version", err: "invalid split ID: wrong UUID version 3, expected 4",
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

func TestNewTombstone(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		ts := object.NewTombstone()

		// check initial values
		require.Nil(t, ts.SplitID())
		require.Nil(t, ts.Members())
		require.Zero(t, ts.ExpirationEpoch())

		// convert to v2 message
		m := ts.ProtoMessage()

		require.Nil(t, m.GetSplitId())
		require.Nil(t, m.GetMembers())
		require.Zero(t, m.GetExpirationEpoch()) //nolint:staticcheck // must be tested still
	})
}
