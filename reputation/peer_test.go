package reputation_test

import (
	"testing"

	apireputation "github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/stretchr/testify/require"
)

func TestPeerIDComparable(t *testing.T) {
	id1 := reputationtest.PeerID()
	require.True(t, id1 == id1)
	id2 := reputationtest.ChangePeerID(id1)
	require.NotEqual(t, id1, id2)
	require.False(t, id1 == id2)
}

func TestPeerID_String(t *testing.T) {
	id1 := reputationtest.PeerID()
	id2 := reputationtest.ChangePeerID(id1)
	require.NotEmpty(t, id1.String())
	require.Equal(t, id1.String(), id1.String())
	require.NotEqual(t, id1.String(), id2.String())
}

func TestPeerID_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("value", func(t *testing.T) {
			id := reputationtest.PeerID()
			var m apireputation.PeerID

			id.WriteToV2(&m)
			m.PublicKey = nil
			require.ErrorContains(t, id.ReadFromV2(&m), "missing value field")
			m.PublicKey = []byte{}
			require.ErrorContains(t, id.ReadFromV2(&m), "missing value field")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("value", func(t *testing.T) {
			id := reputationtest.PeerID()
			var m apireputation.PeerID

			id.WriteToV2(&m)
			m.PublicKey = make([]byte, 32)
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid value length 32")
			m.PublicKey = make([]byte, 34)
			require.ErrorContains(t, id.ReadFromV2(&m), "invalid value length 34")
		})
	})
}

func TestPeerID_DecodeString(t *testing.T) {
	var id reputation.PeerID

	const zeroIDString = "111111111111111111111111111111111"
	require.Equal(t, zeroIDString, id.EncodeToString())

	var bin = [33]byte{106, 6, 81, 91, 166, 102, 170, 186, 188, 108, 51, 93, 37, 154, 31, 156, 67, 97, 148, 186, 222, 175, 255, 251, 153, 158, 211, 222, 251, 168, 26, 141, 16}
	const str = "YVmEgnwZTZsnXFnRKsrk88LAfj4YKm1B83LzR6GCcnCvj"
	require.NoError(t, id.DecodeString(str))
	require.Equal(t, str, id.EncodeToString())
	require.EqualValues(t, bin, id)

	var binOther = [33]byte{14, 5, 25, 39, 25, 170, 76, 164, 133, 133, 150, 101, 89, 226, 39, 70, 35, 200, 81, 200, 121, 104, 205, 74, 36, 179, 14, 151, 244, 135, 93, 244, 229}
	const strOther = "5AZLUQPv8CuUHSqqcFqWf3UdWLP46zxc5Z4riS3gxmkPn"
	require.NoError(t, id.DecodeString(strOther))
	require.Equal(t, strOther, id.EncodeToString())
	require.EqualValues(t, binOther, id)

	t.Run("invalid", func(t *testing.T) {
		var id reputation.PeerID
		for _, testCase := range []struct{ input, err string }{
			{input: "not_a_base58_string", err: "decode base58"},
			{input: "", err: "invalid value length 0"},
			{input: "zJd3YyeBk9o6q281fE81LxiYucTbdToLK3RzdMR8quc", err: "invalid value length 32"},
		} {
			require.ErrorContains(t, id.DecodeString(testCase.input), testCase.err, testCase)
		}
	})
	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst reputation.PeerID
			var msg apireputation.PeerID

			require.NoError(t, dst.DecodeString(str))

			dst[0]++
			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 33), msg.PublicKey)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, dst)

			require.NoError(t, src.DecodeString(str))

			src.WriteToV2(&msg)
			require.Equal(t, bin[:], msg.PublicKey)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, bin, dst)
			require.Equal(t, str, dst.EncodeToString())
		})
	})
}
