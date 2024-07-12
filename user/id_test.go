package user_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
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
	require.Equal(t, s, user.ID(b).EncodeToString())
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
				_, err2 := user.DecodeString(tc.str)
				if tc.contains {
					require.ErrorContains(t, err, tc.err, tc)
					require.ErrorContains(t, err2, tc.err, tc)
				} else {
					require.EqualError(t, err, tc.err, tc)
					require.EqualError(t, err2, tc.err, tc)
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

func TestNewFromScriptHash(t *testing.T) {
	scriptHash := util.Uint160{11, 169, 152, 95, 44, 8, 18, 164, 109, 197, 177, 25, 236, 41, 179, 46, 235, 84, 113, 97}
	id := user.NewFromScriptHash(scriptHash)
	require.EqualValues(t, 0x35, id[0])
	require.Equal(t, scriptHash[:], id[1:21])
	require.Equal(t, []byte{78, 31, 235, 139}, id[21:])
}

func TestNewFromECDSAPublicKey(t *testing.T) {
	x := []byte{41, 129, 156, 121, 170, 4, 67, 132, 75, 159, 26, 118, 120, 134, 213, 180, 46, 250, 210, 31, 218, 99, 126, 71, 153, 132, 123, 219, 142, 18, 121, 135}
	y := []byte{128, 105, 166, 6, 88, 228, 216, 235, 151, 24, 251, 57, 219, 196, 207, 189, 209, 250, 68, 113, 26, 197, 77, 31, 193, 247, 157, 253, 162, 127, 59, 43}
	expected := [user.IDSize]byte{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
	pub := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(x),
		Y:     new(big.Int).SetBytes(y),
	}
	id := user.NewFromECDSAPublicKey(pub)
	require.EqualValues(t, expected, id)
}

func TestDecodeString(t *testing.T) {
	const s = "NfPCfAR4inFKGZqrjnpMX6yQ5hZCtVpXUx"
	b := [user.IDSize]byte{53, 213, 144, 221, 254, 189, 129, 167, 41, 216, 106, 91, 19, 100, 248, 81, 99, 172, 115, 203, 120, 154, 192, 43, 69}
	id, err := user.DecodeString(s)
	require.NoError(t, err)
	require.EqualValues(t, b, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			str      string
			contains bool
			err      string
		}{
			{name: "base58", str: "NfPCfAR4inFKGZqrjnpMX6yQ5hZCtVpXU_", contains: true, err: "decode base58"},
			{name: "undersize", str: "5qatkFMhsW6ff9SG6ZYFy1qr6nFN2NBvM", err: "invalid length 24, expected 25"},
			{name: "oversize", str: "2er5Dy738K5sXP1QgDNrP3GwRJRKqgizR4Nc", err: "invalid length 26, expected 25"},
			{name: "prefix", str: "TtsRYkmYiw6aVsQCw4HR6wKheRqSJLFsLQ", err: "invalid prefix byte 0x42, expected 0x35"},
			{name: "checksum", str: "NXsxvBYg6bbDFe2WmitnJ9eZPxDCKmwWVJ", err: "checksum mismatch"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, err := user.DecodeString(tc.str)
				if tc.contains {
					require.ErrorContains(t, err, tc.err, tc)
				} else {
					require.EqualError(t, err, tc.err, tc)
				}
			})
		}
	})
}
