package oid_test

import (
	"crypto/rand"
	"crypto/sha256"
	"strconv"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

const emptyID = "11111111111111111111111111111111"

func randSHA256Checksum(t *testing.T) (cs [sha256.Size]byte) {
	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	return
}

func TestIDV2(t *testing.T) {
	var id oid.ID

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

	var id1 oid.ID
	id1.SetSHA256(cs)

	var id2 oid.ID
	id2.SetSHA256(cs)

	var id3 oid.ID
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
				var oid oid.ID

				require.NoError(t, oid.DecodeString(str))

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
				var oid oid.ID

				require.Error(t, oid.DecodeString(str))
			})
		}
	})
}

func TestID_String(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		var id oid.ID
		require.Equal(t, emptyID, id.EncodeToString())
	})

	t.Run("should be equal", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				cs := randSHA256Checksum(t)
				str := base58.Encode(cs[:])
				var oid oid.ID

				require.NoError(t, oid.DecodeString(str))
				require.Equal(t, str, oid.EncodeToString())
			})
		}
	})
}

func TestObjectIDEncoding(t *testing.T) {
	id := oidtest.ID()

	t.Run("binary", func(t *testing.T) {
		data, err := id.Marshal()
		require.NoError(t, err)

		var id2 oid.ID
		require.NoError(t, id2.Unmarshal(data))

		require.Equal(t, id, id2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := id.MarshalJSON()
		require.NoError(t, err)

		var id2 oid.ID
		require.NoError(t, id2.UnmarshalJSON(data))

		require.Equal(t, id, id2)
	})
}

func TestNewIDFromV2(t *testing.T) {
	t.Run("from zero", func(t *testing.T) {
		var (
			x  oid.ID
			v2 refs.ObjectID
		)

		require.Error(t, x.ReadFromV2(v2))
	})
}

func TestID_ToV2(t *testing.T) {
	t.Run("zero to v2", func(t *testing.T) {
		var (
			x  oid.ID
			v2 refs.ObjectID
		)

		x.WriteToV2(&v2)

		require.Equal(t, oid.Size, len(v2.GetValue()))
		require.Equal(t, emptyID, base58.Encode(v2.GetValue()))
	})
}

func TestID_Encode(t *testing.T) {
	var id oid.ID

	t.Run("panic", func(t *testing.T) {
		dst := make([]byte, oid.Size-1)

		require.Panics(t, func() {
			id.Encode(dst)
		})
	})

	t.Run("correct", func(t *testing.T) {
		dst := make([]byte, oid.Size)

		require.NotPanics(t, func() {
			id.Encode(dst)
		})
		require.Equal(t, emptyID, id.EncodeToString())
	})
}

func TestID_Unmarshal(t *testing.T) {
	t.Run("invalid value length", func(t *testing.T) {
		var id oid.ID
		var m refs.ObjectID
		for _, tc := range []struct {
			err string
			val []byte
		}{
			{"invalid length 0", nil},
			{"invalid length 0", []byte{}},
			{"invalid length 31", make([]byte, 31)},
			{"invalid length 33", make([]byte, 33)},
		} {
			m.SetValue(tc.val)
			b := m.StableMarshal(nil)
			require.EqualError(t, id.Unmarshal(b), tc.err)
		}
	})
}

func TestIDDecodingFailures(t *testing.T) {
	var id oid.ID
	var m refs.ObjectID
	for _, tc := range []struct {
		err     string
		corrupt func(*refs.ObjectID)
	}{
		{"invalid length 0", func(m *refs.ObjectID) { m.SetValue(nil) }},
		{"invalid length 0", func(m *refs.ObjectID) { m.SetValue([]byte{}) }},
		{"invalid length 31", func(m *refs.ObjectID) { m.SetValue(make([]byte, 31)) }},
		{"invalid length 33", func(m *refs.ObjectID) { m.SetValue(make([]byte, 33)) }},
	} {
		id.WriteToV2(&m)
		tc.corrupt(&m)
		t.Run(tc.err+"/api", func(t *testing.T) {
			require.EqualError(t, id.ReadFromV2(m), tc.err)
		})
		t.Run(tc.err+"/binary", func(t *testing.T) {
			b := m.StableMarshal(nil)
			require.EqualError(t, id.Unmarshal(b), tc.err)
		})
		t.Run(tc.err+"/json", func(t *testing.T) {
			b, err := m.MarshalJSON()
			require.NoError(t, err, tc.err)
			require.EqualError(t, id.UnmarshalJSON(b), tc.err)
		})
	}
}

func TestIDComparable(t *testing.T) {
	x := oidtest.ID()
	y := x
	require.True(t, x == y)
	require.False(t, x != y)
	y = oidtest.OtherID(x)
	require.False(t, x == y)
	require.True(t, x != y)
}
