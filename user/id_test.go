package user_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestIDComparable(t *testing.T) {
	id1 := usertest.ID()
	require.True(t, id1 == id1)
	id2 := usertest.ChangeID(id1)
	require.NotEqual(t, id1, id2)
	require.False(t, id1 == id2)
}

func TestID_String(t *testing.T) {
	id1 := usertest.ID()
	id2 := usertest.ChangeID(id1)
	require.NotEmpty(t, id1.String())
	require.Equal(t, id1.String(), id1.String())
	require.NotEqual(t, id1.String(), id2.String())
}

func TestNewID(t *testing.T) {
	var h util.Uint160
	rand.Read(h[:])

	id := user.NewID(h)
	require.EqualValues(t, 0x35, id[0])
	require.Equal(t, h[:], id[1:21])
	require.Equal(t, hash.Checksum(append([]byte{0x35}, h[:]...)), id[21:])

	var msg refs.OwnerID
	id.WriteToV2(&msg)
	b := []byte{0x35}
	b = append(b, h[:]...)
	b = append(b, hash.Checksum(b)...)
	require.Equal(t, b, msg.Value)
	var id2 user.ID
	require.NoError(t, id2.ReadFromV2(&msg))
	require.Equal(t, id2, id)

	var id3 user.ID
	require.NoError(t, id3.DecodeString(id.EncodeToString()))
	require.Equal(t, id3, id)
}

func TestID_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("value", func(t *testing.T) {
			id := usertest.ID()
			var m refs.OwnerID

			id.WriteToV2(&m)
			m.Value = nil
			require.ErrorContains(t, id.ReadFromV2(&m), "missing value field")
			m.Value = []byte{}
			require.ErrorContains(t, id.ReadFromV2(&m), "missing value field")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("value", func(t *testing.T) {
			id := usertest.ID()
			var m refs.OwnerID

			id.WriteToV2(&m)
			m.Value = make([]byte, 24)
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid value length 24")
			m.Value = make([]byte, 26)
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid value length 26")
			m.Value = make([]byte, 25)
			m.Value[0] = 0x34
			copy(m.Value[21:], hash.Checksum(m.Value[:21]))
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid prefix byte 0x34, expected 0x35")
			m.Value[0] = 0x35 // checksum become broken
			require.ErrorContains(t, id.ReadFromV2(&m), "value checksum mismatch")
		})
	})
}

func TestID_DecodeString(t *testing.T) {
	var id user.ID

	const zeroIDString = "1111111111111111111111111"
	require.Equal(t, zeroIDString, id.EncodeToString())
	id = usertest.ChangeID(id)
	require.Error(t, id.DecodeString(zeroIDString))

	var bin = [25]byte{53, 72, 207, 149, 237, 209, 4, 50, 202, 244, 5, 17, 110, 81, 232, 216, 209, 218, 182, 113, 105, 9, 34, 73, 84}
	const str = "NSYxWsYXboEjX31dMqVSC9aUTJnCaSP4v7"
	require.NoError(t, id.DecodeString(str))
	require.Equal(t, str, id.EncodeToString())
	require.EqualValues(t, bin, id)

	var binOther = [25]byte{53, 3, 151, 68, 134, 6, 234, 16, 104, 195, 133, 153, 6, 87, 28, 4, 41, 45, 67, 87, 121, 135, 107, 199, 252}
	const strOther = "NLExRbNWNQpc6pEyBJXrFaPqcXMFFHChQo"
	require.NoError(t, id.DecodeString(strOther))
	require.Equal(t, strOther, id.EncodeToString())
	require.EqualValues(t, binOther, id)

	t.Run("invalid", func(t *testing.T) {
		var id user.ID
		for _, testCase := range []struct{ input, err string }{
			{input: "not_a_base58_string", err: "decode base58"},
			{input: "", err: "invalid value length 0"},
			{input: "MzB7cw27FZpBdcLiexQN6DSgriAa9WERdM", err: "invalid prefix byte 0x34, expected 0x35"},
			{input: "NgkmJY4DsqYsomn5y1TKz4GoBHmW55ZrwC", err: "value checksum mismatch"},
		} {
			require.ErrorContains(t, id.DecodeString(testCase.input), testCase.err, testCase)
		}
	})
	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst user.ID
			var msg refs.OwnerID

			require.NoError(t, dst.DecodeString(str))

			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 25), msg.Value)
			require.Error(t, dst.ReadFromV2(&msg))

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
	var id user.ID
	require.True(t, id.IsZero())
	for i := range id {
		id2 := id
		id2[i]++
		require.False(t, id2.IsZero())
	}
}
