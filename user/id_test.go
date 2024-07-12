package user_test

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestID_WalletBytes(t *testing.T) {
	var scriptHash util.Uint160
	//nolint:staticcheck
	rand.Read(scriptHash[:])

	var id user.ID
	id.SetScriptHash(scriptHash)

	w := id.WalletBytes()

	var m refs.OwnerID
	m.SetValue(w)

	err := id.ReadFromV2(m)
	require.NoError(t, err)
}

func TestID_SetScriptHash(t *testing.T) {
	var scriptHash util.Uint160
	//nolint:staticcheck
	rand.Read(scriptHash[:])

	var id user.ID
	id.SetScriptHash(scriptHash)

	var m refs.OwnerID
	id.WriteToV2(&m)

	var id2 user.ID

	err := id2.ReadFromV2(m)
	require.NoError(t, err)

	require.True(t, id2 == id)
}

func TestV2_ID(t *testing.T) {
	id := usertest.ID()
	var m refs.OwnerID
	var id2 user.ID

	t.Run("OK", func(t *testing.T) {
		id.WriteToV2(&m)

		err := id2.ReadFromV2(m)
		require.NoError(t, err)
		require.True(t, id2 == id)
	})

	val := m.GetValue()

	t.Run("invalid size", func(t *testing.T) {
		m.SetValue(val[:24])

		err := id2.ReadFromV2(m)
		require.Error(t, err)
	})

	t.Run("invalid prefix", func(t *testing.T) {
		val := bytes.Clone(val)
		val[0]++

		m.SetValue(val)

		err := id2.ReadFromV2(m)
		require.Error(t, err)
	})

	t.Run("invalid checksum", func(t *testing.T) {
		val := bytes.Clone(val)
		val[21]++

		m.SetValue(val)

		err := id2.ReadFromV2(m)
		require.Error(t, err)
	})
}

func TestID_EncodeToString(t *testing.T) {
	const s = "NXWcEedga62wcBmfb9dPwar3vbbZJrMtT1"
	b := [user.IDSize]byte{53, 127, 54, 116, 58, 70, 206, 247, 185, 103, 214, 89, 184, 42, 40, 234, 173, 68, 209, 25, 168, 14, 134, 47, 224}
	var id user.ID

	var m refs.OwnerID
	m.SetValue(b[:])
	require.NoError(t, id.ReadFromV2(m))

	require.Equal(t, s, id.EncodeToString())
}

func TestID_DecodeString(t *testing.T) {
	const s = "NXemr2z9qwACUgnCarY7iYF7uLEePVKBtA"
	b := [user.IDSize]byte{53, 128, 193, 205, 198, 129, 137, 93, 70, 128, 33, 126, 194, 179, 75, 94, 138, 202, 178, 249, 166, 25, 204, 50, 15}
	var id user.ID
	require.NoError(t, id.DecodeString(s))
	require.Equal(t, b[:], id.WalletBytes())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			str      string
			contains bool
			err      string
		}{
			{name: "base58", str: "NXemr2z9qwACUgnCarY7iYF7uLEePVKBt_", contains: true, err: "decode base58"},
			{name: "undersize", str: "zqerPAhxBxcTC7sWuY6D7x6PNSYKifPx", err: "invalid length 24, expected 25"},
			{name: "oversize", str: "5sgJAVddETxrF3w7Eh2KRtrHEUMB8wXMZtkf", err: "invalid length 26, expected 25"},
			{name: "prefix", str: "ToY9yFMFk3usAd7zc1gkb5PGYMeDekjkXH", err: "invalid prefix byte 0x42, expected 0x35"},
			{name: "checksum", str: "NRMge8AquoXsj534yXqs3miJwNn87JsXJP", err: "checksum mismatch"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				err := new(user.ID).DecodeString(tc.str)
				if tc.contains {
					require.ErrorContains(t, err, tc.err, tc)
				} else {
					require.EqualError(t, err, tc.err, tc)
				}
			})
		}
	})
}

func TestID_Equal(t *testing.T) {
	id1 := usertest.ID()
	id2 := usertest.ID()
	id3 := id1

	require.True(t, id1.Equals(id1)) // self-equality
	require.True(t, id1.Equals(id3))
	require.True(t, id3.Equals(id1)) // commutativity
	require.False(t, id1.Equals(id2))
}

func TestIDComparable(t *testing.T) {
	x := usertest.ID()
	y := x
	require.True(t, x == y)
	require.False(t, x != y)
	y = usertest.OtherID(x)
	require.False(t, x == y)
	require.True(t, x != y)
}
