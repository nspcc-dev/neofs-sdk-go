package oid_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

var validIDBytes = [oid.Size]byte{231, 189, 121, 7, 173, 134, 254, 165, 63, 186, 60, 89, 33, 95, 46, 103,
	217, 57, 164, 87, 82, 204, 251, 226, 1, 100, 32, 72, 251, 0, 7, 172}

var validIDProtoBytes = append([]byte{10, 32}, validIDBytes[:]...)

// corresponds to validIDBytes.
const validIDString = "GbckSBPEdM2P41Gkb9cVapFYb5HmRPDTZZp9JExGnsCF"

// corresponds to validIDBytes.
var validIDJSON = `{"value":"5715B62G/qU/ujxZIV8uZ9k5pFdSzPviAWQgSPsAB6w="}`

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

func toProtoBytes(b []byte) []byte { return protowire.AppendBytes([]byte{10}, b) }

func toProtoJSON(b []byte) []byte {
	b, err := neofsproto.MarshalMessageJSON(&refs.ObjectID{Value: b})
	if err != nil {
		panic(fmt.Sprintf("unexpected MarshalJSON error: %v", err))
	}
	return b
}

func TestID_FromProtoMessage(t *testing.T) {
	m := &refs.ObjectID{Value: validIDBytes[:]}
	var id oid.ID
	require.NoError(t, id.FromProtoMessage(m))
	require.EqualValues(t, validIDBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range append(invalidValueTestcases, invalidValueTestCase{
			name: "zero value", err: "zero object ID", val: make([]byte, cid.Size),
		}) {
			t.Run(tc.name, func(t *testing.T) {
				m := &refs.ObjectID{Value: tc.val}
				require.EqualError(t, new(oid.ID).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestID_Decode(t *testing.T) {
	var id oid.ID
	require.NoError(t, id.Decode(validIDBytes[:]))
	require.EqualValues(t, validIDBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidValueTestcases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := oid.DecodeBytes(tc.val)
				require.EqualError(t, err, tc.err)
				require.EqualError(t, new(oid.ID).Decode(tc.val), tc.err)
			})
		}
	})
}

func TestID_ProtoMessage(t *testing.T) {
	id := oidtest.ID()
	m := id.ProtoMessage()
	require.Equal(t, id[:], m.GetValue())
}

func TestID_EncodeToString(t *testing.T) {
	require.Equal(t, validIDString, oid.ID(validIDBytes).EncodeToString())
}

func TestID_DecodeString(t *testing.T) {
	var id oid.ID
	require.NoError(t, id.DecodeString(validIDString))
	require.EqualValues(t, validIDBytes, id)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			str      string
			contains bool
			err      string
		}{
			{name: "base58", str: "64sexUVdHg9dnqxHTB3uGsHfhPTnuegrupF4bJGXSbZ_", contains: true, err: "decode base58"},
			{name: "undersize", str: "tjkeRJWKpNa7Xjh3sZymJWZG5UGTGoZtEimGUjQesV", err: "invalid length 31"},
			{name: "oversize", str: "2Jwe3Yy4t5hgV2VeXxPkxYV5GNwPJEBjD3oZFDUiHneQ7J", err: "invalid length 33"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				err := new(oid.ID).DecodeString(tc.str)
				_, err2 := oid.DecodeString(tc.str)
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

func TestID_Marshal(t *testing.T) {
	require.Equal(t, validIDProtoBytes, oid.ID(validIDBytes).Marshal())
}

func TestID_Unmarshal(t *testing.T) {
	t.Run("invalid protobuf", func(t *testing.T) {
		err := new(oid.ID).Unmarshal([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "cannot parse invalid wire-format data")
	})

	var id oid.ID
	require.NoError(t, id.Unmarshal(validIDProtoBytes))
	require.EqualValues(t, validIDBytes, id)

	t.Run("protocol violation", func(t *testing.T) {
		for _, tc := range invalidValueTestcases {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(oid.ID).Unmarshal(toProtoBytes(tc.val)), tc.err)
			})
		}
	})
}

func TestID_MarshalJSON(t *testing.T) {
	id := oid.ID(validIDBytes)
	b, err := id.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, validIDJSON, string(b))
}

func TestID_UnmarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := new(oid.ID).UnmarshalJSON([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var id oid.ID
	require.NoError(t, id.UnmarshalJSON([]byte(validIDJSON)))
	require.EqualValues(t, validIDBytes, id)

	t.Run("protocol violation", func(t *testing.T) {
		for _, tc := range invalidValueTestcases {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(oid.ID).UnmarshalJSON(toProtoJSON(tc.val)), tc.err)
			})
		}
	})
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

func TestNewFromObjectHeaderBinary(t *testing.T) {
	// use any binary just for the test
	hdr := testutil.RandByteSlice(512)
	id := oid.NewFromObjectHeaderBinary(hdr)
	require.EqualValues(t, sha256.Sum256(hdr), id)
	require.Equal(t, id, oid.NewFromObjectHeaderBinary(hdr))
	for i := range hdr {
		hdrCp := bytes.Clone(hdr)
		hdrCp[i]++
		require.NotEqual(t, id, oid.NewFromObjectHeaderBinary(hdrCp))
	}
}

func TestID_String(t *testing.T) {
	id := oidtest.ID()
	require.NotEmpty(t, id.String())
	require.Equal(t, id.String(), id.String())
	require.NotEqual(t, id.String(), oidtest.OtherID(id).String())
}

func TestID_CalculateIDSignature(t *testing.T) {
	usr := usertest.User()
	id := oidtest.ID()

	_, err := id.CalculateIDSignature(neofscryptotest.FailSigner(usr))
	require.Error(t, err)

	for _, s := range []neofscrypto.Signer{
		usr,
		usr.RFC6979,
		usr.WalletConnect,
	} {
		sig, err := id.CalculateIDSignature(s)
		require.NoError(t, err)
		require.Equal(t, s.Scheme(), sig.Scheme())
		require.Equal(t, s.Public(), sig.PublicKey())
		require.Equal(t, usr.PublicKeyBytes, sig.PublicKeyBytes())
		require.True(t, sig.Verify(id.Marshal()))
		require.True(t, s.Public().Verify(id.Marshal(), sig.Value()))
	}
}

func TestID_IsZero(t *testing.T) {
	var id oid.ID
	require.True(t, id.IsZero())
	for i := range oid.Size {
		var id2 oid.ID
		id2[i]++
		require.False(t, id2.IsZero())
	}
}

func TestID_Compare(t *testing.T) {
	var id1, id2 oid.ID

	require.Equal(t, 0, id1.Compare(id2))
	id1[0] = 1
	require.Equal(t, 1, id1.Compare(id2))
	id2[0] = 2
	require.Equal(t, -1, id1.Compare(id2))
}
