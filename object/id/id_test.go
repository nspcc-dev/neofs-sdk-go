package oid_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestIDComparable(t *testing.T) {
	id1 := oidtest.ID()
	require.True(t, id1 == id1)
	id2 := oidtest.ChangeID(id1)
	require.NotEqual(t, id1, id2)
	require.False(t, id1 == id2)
}

func TestID_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("value", func(t *testing.T) {
			id := oidtest.ID()
			var m refs.ObjectID

			id.WriteToV2(&m)
			m.Value = nil
			require.ErrorContains(t, id.ReadFromV2(&m), "missing value field")
			m.Value = []byte{}
			require.ErrorContains(t, id.ReadFromV2(&m), "missing value field")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("value", func(t *testing.T) {
			id := oidtest.ID()
			var m refs.ObjectID

			id.WriteToV2(&m)
			m.Value = make([]byte, 31)
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid value length 31")
			m.Value = make([]byte, 33)
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid value length 33")
		})
	})
}

func TestID_DecodeString(t *testing.T) {
	var id oid.ID

	const zeroIDString = "11111111111111111111111111111111"
	require.Equal(t, zeroIDString, id.EncodeToString())
	id = oidtest.ChangeID(id)
	require.NoError(t, id.DecodeString(zeroIDString))
	require.Equal(t, zeroIDString, id.EncodeToString())
	require.Zero(t, id)

	var bin = [32]byte{231, 129, 236, 104, 74, 71, 155, 100, 72, 209, 186, 80, 2, 184, 9, 161, 10, 76, 18, 203, 126, 94, 101, 42, 157, 211, 66, 99, 247, 143, 226, 23}
	const str = "Gai5pjZVewwmscQ5UczQbj2W8Wkh9d1BGUoRNzjR6QCN"
	require.NoError(t, id.DecodeString(str))
	require.Equal(t, str, id.EncodeToString())
	require.EqualValues(t, bin, id)

	var binOther = [32]byte{216, 146, 23, 99, 156, 90, 232, 244, 202, 213, 0, 92, 22, 194, 164, 150, 233, 163, 175, 199, 187, 45, 65, 7, 190, 124, 77, 99, 8, 172, 36, 112}
	const strOther = "FaQGU3PHuHjhHbce1u8AuHuabx4Ra9CxREsMcZffXwM1"
	require.NoError(t, id.DecodeString(strOther))
	require.Equal(t, strOther, id.EncodeToString())
	require.EqualValues(t, binOther, id)

	t.Run("invalid", func(t *testing.T) {
		var id oid.ID
		for _, testCase := range []struct{ input, err string }{
			{input: "not_a_base58_string", err: "decode base58"},
			{input: "", err: "invalid value length 0"},
			{input: "qxAE9SLuDq7dARPAFaWG6vbuGoocwoTn19LK5YVqnS", err: "invalid value length 31"},
			{input: "HJJEkEKthnvMw7NsZNgzBEQ4tf9AffmaBYWxfBULvvbPW", err: "invalid value length 33"},
		} {
			require.ErrorContains(t, id.DecodeString(testCase.input), testCase.err, testCase)
		}
	})
	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src oid.ID
			var msg refs.ObjectID

			msg.Value = []byte("any")

			require.NoError(t, proto.Unmarshal(src.Marshal(), &msg))
			require.Equal(t, make([]byte, 32), msg.Value)

			require.NoError(t, src.DecodeString(str))

			require.NoError(t, proto.Unmarshal(src.Marshal(), &msg))
			require.Equal(t, bin[:], msg.Value)
		})
		t.Run("api", func(t *testing.T) {
			var src, dst oid.ID
			var msg refs.ObjectID

			require.NoError(t, dst.DecodeString(str))

			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 32), msg.Value)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst)
			require.Equal(t, zeroIDString, dst.EncodeToString())

			require.NoError(t, src.DecodeString(str))

			src.WriteToV2(&msg)
			require.Equal(t, bin[:], msg.Value)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, bin, dst)
			require.Equal(t, str, dst.EncodeToString())
		})
	})
}

func TestID_IsZero(t *testing.T) {
	var id oid.ID
	require.True(t, id.IsZero())
	for i := range id {
		id2 := id
		id2[i]++
		require.False(t, id2.IsZero())
	}
}
