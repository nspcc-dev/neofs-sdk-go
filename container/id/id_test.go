package cid_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

var validBytes = [cid.Size]byte{231, 189, 121, 7, 173, 134, 254, 165, 63, 186, 60, 89, 33, 95, 46, 103,
	217, 57, 164, 87, 82, 204, 251, 226, 1, 100, 32, 72, 251, 0, 7, 172}

// corresponds to validBytes.
const validString = "GbckSBPEdM2P41Gkb9cVapFYb5HmRPDTZZp9JExGnsCF"

type invalidValueTestCase struct {
	name string
	err  string
	val  []byte
}

var invalidValueTestcases = []invalidValueTestCase{
	{name: "nil value", err: "invalid length 0", val: nil},
	{name: "empty value", err: "invalid length 0", val: []byte{}},
	{name: "undersized value", err: "invalid length 31", val: make([]byte, 31)},
	{name: "oversized value", err: "invalid length 33", val: make([]byte, 33)},
}

func TestID_FromProtoMessage(t *testing.T) {
	m := &refs.ContainerID{Value: validBytes[:]}
	var id cid.ID
	require.NoError(t, id.FromProtoMessage(m))
	require.EqualValues(t, validBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range append(invalidValueTestcases, invalidValueTestCase{
			name: "zero value", err: "zero container ID", val: make([]byte, cid.Size),
		}) {
			t.Run(tc.name, func(t *testing.T) {
				m := &refs.ContainerID{Value: tc.val}
				require.EqualError(t, new(cid.ID).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestID_Decode(t *testing.T) {
	var id cid.ID
	require.NoError(t, id.Decode(validBytes[:]))
	require.EqualValues(t, validBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidValueTestcases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := cid.DecodeBytes(tc.val)
				require.EqualError(t, err, tc.err)
				require.EqualError(t, new(cid.ID).Decode(tc.val), tc.err)
			})
		}
	})
}

func TestID_EncodeToString(t *testing.T) {
	require.Equal(t, validString, cid.ID(validBytes).EncodeToString())
}

func TestID_DecodeString(t *testing.T) {
	var id cid.ID
	require.NoError(t, id.DecodeString(validString))
	require.EqualValues(t, validBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			str      string
			contains bool
			err      string
		}{
			{name: "base58", str: "Dsa5sCDorTWsL7dYa1pnVPEuHg9bDN39XxdxXxghrwS_", contains: true, err: "decode base58"},
			{name: "undersize", str: "3RArVxBNPE4rZ9f5oHwxTFi7LTSY1fQ3BzNJZat2ZoV", err: "invalid length 31"},
			{name: "oversize", str: "CJ1rzsceKvtmKtZcuEJssiLVqDgBc6rPp5dwxhNxChbap", err: "invalid length 33"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				err := new(cid.ID).DecodeString(tc.str)
				_, err2 := cid.DecodeString(tc.str)
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

func TestID_ProtoMessage(t *testing.T) {
	id := cidtest.ID()
	m := id.ProtoMessage()
	require.Equal(t, id[:], m.GetValue())
}

func TestIDComparable(t *testing.T) {
	x := cidtest.ID()
	y := x
	require.True(t, x == y)
	require.False(t, x != y)
	y = cidtest.OtherID(x)
	require.False(t, x == y)
	require.True(t, x != y)
}

func TestNewFromContainerBinary(t *testing.T) {
	// use any binary just for the test
	cnr := testutil.RandByteSlice(512)
	id := cid.NewFromMarshalledContainer(cnr)
	require.EqualValues(t, sha256.Sum256(cnr), id)
	require.Equal(t, id, cid.NewFromMarshalledContainer(cnr))
	for i := range cnr {
		cnrCp := bytes.Clone(cnr)
		cnrCp[i]++
		require.NotEqual(t, id, cid.NewFromMarshalledContainer(cnrCp))
	}
}

func TestID_String(t *testing.T) {
	id := cidtest.ID()
	require.NotEmpty(t, id.String())
	require.Equal(t, id.String(), id.String())
	require.NotEqual(t, id.String(), cidtest.OtherID(id).String())
}

func TestID_IsZero(t *testing.T) {
	var id cid.ID
	require.True(t, id.IsZero())
	for i := range cid.Size {
		var id2 cid.ID
		id2[i]++
		require.False(t, id2.IsZero())
	}
}
