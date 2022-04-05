package oid

import (
	"crypto/rand"
	"crypto/sha256"
	"strconv"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/stretchr/testify/require"
)

func randID(t *testing.T) ID {
	var id ID
	id.SetSHA256(randSHA256Checksum(t))

	return id
}

func randSHA256Checksum(t *testing.T) (cs [sha256.Size]byte) {
	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	return
}

func TestIDV2(t *testing.T) {
	var id ID

	checksum := [sha256.Size]byte{}

	_, err := rand.Read(checksum[:])
	require.NoError(t, err)

	id.SetSHA256(checksum)

	var idV2 refs.ObjectID
	id.WriteToV2(&idV2)

	require.Equal(t, checksum[:], idV2.GetValue())
}

func TestID_Equal(t *testing.T) {
	cs := randSHA256Checksum(t)

	var id1 ID
	id1.SetSHA256(cs)

	var id2 ID
	id2.SetSHA256(cs)

	var id3 ID
	id3.SetSHA256(randSHA256Checksum(t))

	require.True(t, id1.Equals(id2))
	require.False(t, id1.Equals(id3))
}

func TestID_Parse(t *testing.T) {
	t.Run("should parse successful", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				cs := randSHA256Checksum(t)
				str := base58.Encode(cs[:])
				var oid ID

				require.NoError(t, oid.Parse(str))

				var oidV2 refs.ObjectID
				oid.WriteToV2(&oidV2)

				require.Equal(t, cs[:], oidV2.GetValue())
			})
		}
	})

	t.Run("should failure on parse", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			j := i
			t.Run(strconv.Itoa(j), func(t *testing.T) {
				cs := []byte{1, 2, 3, 4, 5, byte(j)}
				str := base58.Encode(cs)
				var oid ID

				require.Error(t, oid.Parse(str))
			})
		}
	})
}

func TestID_String(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var id ID
		require.Empty(t, id.String())
	})

	t.Run("should be equal", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				cs := randSHA256Checksum(t)
				str := base58.Encode(cs[:])
				var oid ID

				require.NoError(t, oid.Parse(str))
				require.Equal(t, str, oid.String())
			})
		}
	})
}

func TestObjectIDEncoding(t *testing.T) {
	id := randID(t)

	t.Run("binary", func(t *testing.T) {
		data, err := id.Marshal()
		require.NoError(t, err)

		var id2 ID
		require.NoError(t, id2.Unmarshal(data))

		require.Equal(t, id, id2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := id.MarshalJSON()
		require.NoError(t, err)

		var a2 ID
		require.NoError(t, a2.UnmarshalJSON(data))

		require.Equal(t, id, a2)
	})
}

func TestNewIDFromV2(t *testing.T) {
	t.Run("from zero", func(t *testing.T) {
		var (
			x  ID
			v2 refs.ObjectID
		)

		x.ReadFromV2(v2)

		require.Empty(t, x.String())
	})
}

func TestID_ToV2(t *testing.T) {
	t.Run("zero to v2", func(t *testing.T) {
		var (
			x  ID
			v2 refs.ObjectID
		)

		x.WriteToV2(&v2)

		require.Nil(t, v2.GetValue())
	})
}
