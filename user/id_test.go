package user_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

var (
	validX = []byte{41, 129, 156, 121, 170, 4, 67, 132, 75, 159, 26, 118, 120, 134, 213, 180,
		46, 250, 210, 31, 218, 99, 126, 71, 153, 132, 123, 219, 142, 18, 121, 135}
	validY = []byte{128, 105, 166, 6, 88, 228, 216, 235, 151, 24, 251, 57, 219, 196, 207, 189,
		209, 250, 68, 113, 26, 197, 77, 31, 193, 247, 157, 253, 162, 127, 59, 43}
	// corresponds to (validX, validY) public key.
	validBytes = [user.IDSize]byte{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57,
		160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
)

// corresponds to validBytes.
var validString = "NQZkR7mG74rJsGAHnpkiFeU9c4f5VLN54f"

func TestID_WalletBytes(t *testing.T) {
	id := usertest.ID()
	require.Equal(t, id[:], id.WalletBytes())
}

func TestID_ReadFromV2(t *testing.T) {
	var m refs.OwnerID
	m.SetValue(validBytes[:])
	var id user.ID
	require.NoError(t, id.ReadFromV2(m))
	require.EqualValues(t, validBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			err  string
			val  []byte
		}{
			{name: "nil value", err: "invalid length 0, expected 25", val: nil},
			{name: "empty value", err: "invalid length 0, expected 25", val: []byte{}},
			{name: "undersized value", err: "invalid length 24, expected 25", val: validBytes[:24]},
			{name: "oversized value", err: "invalid length 26, expected 25", val: append(validBytes[:], 1)},
		} {
			t.Run(tc.name, func(t *testing.T) {
				var m refs.OwnerID
				m.SetValue(tc.val)
				require.EqualError(t, new(user.ID).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestID_WriteToV2(t *testing.T) {
	id := usertest.ID()
	var m refs.OwnerID
	id.WriteToV2(&m)
	require.Equal(t, id[:], m.GetValue())
}

func TestID_EncodeToString(t *testing.T) {
	require.Equal(t, validString, user.ID(validBytes).EncodeToString())
}

func TestID_DecodeString(t *testing.T) {
	var id user.ID
	require.NoError(t, id.DecodeString(validString))
	require.EqualValues(t, validBytes, id)

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
	var id2 user.ID
	id2.SetScriptHash(scriptHash)
	require.Equal(t, id, id2)
}

func TestNewFromECDSAPublicKey(t *testing.T) {
	pub := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(validX),
		Y:     new(big.Int).SetBytes(validY),
	}
	id := user.NewFromECDSAPublicKey(pub)
	require.EqualValues(t, validBytes, id)
}

func TestID_String(t *testing.T) {
	id := usertest.ID()
	require.NotEmpty(t, id.String())
	require.Equal(t, id.String(), id.String())
	require.NotEqual(t, id.String(), usertest.OtherID(id).String())
}
