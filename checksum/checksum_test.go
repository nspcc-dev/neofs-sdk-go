package checksum_test

import (
	"crypto/sha256"
	"hash"
	"hash/adler32"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

func TestType_String(t *testing.T) {
	for _, tc := range []struct {
		typ checksum.Type
		exp string
	}{
		{0, "CHECKSUM_TYPE_UNSPECIFIED"},
		{checksum.SHA256, "SHA256"},
		{checksum.TillichZemor, "TZ"},
		{3, "3"},
	} {
		require.Equal(t, tc.exp, tc.typ.String())
	}
}

func TestChecksum_String(t *testing.T) {
	data := []byte("Hello, world!")
	for _, tc := range []struct {
		cs  checksum.Checksum
		exp string
	}{
		{checksum.New(0, data), "CHECKSUM_TYPE_UNSPECIFIED:48656c6c6f2c20776f726c6421"},
		{checksum.New(127, data), "127:48656c6c6f2c20776f726c6421"},
		{checksum.NewSHA256(sha256.Sum256(data)), "SHA256:315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3"},
		{checksum.NewTillichZemor(tz.Sum(data)), "TZ:0000014249f10795c0240eddca8a6ebf000001c9c4dc98b017fd92ad62979c8c0000008d94cd98a457b983e937838dcd000000dbc8689e75c7dd8925ad0df727"},
	} {
		require.Equal(t, tc.exp, tc.cs.String())
	}
}

func TestChecksumDecodingFailures(t *testing.T) {
	t.Run("api", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*refs.Checksum)
		}{
			{name: "value/nil", err: "missing value", corrupt: func(cs *refs.Checksum) { cs.Sum = nil }},
			{name: "value/empty", err: "missing value", corrupt: func(cs *refs.Checksum) { cs.Sum = []byte{} }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				var src, dst checksum.Checksum

				m := src.ProtoMessage()
				tc.corrupt(m)
				require.ErrorContains(t, dst.FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestNew(t *testing.T) {
	typ := checksum.Type(rand.Int31())
	val := make([]byte, 128)
	//nolint:staticcheck
	rand.Read(val)
	cs := checksum.New(typ, val)
	require.Equal(t, typ, cs.Type())
	require.Equal(t, val, cs.Value())

	otherTyp := checksum.Type(rand.Int31())
	otherVal := make([]byte, 128)
	//nolint:staticcheck
	rand.Read(otherVal)
	cs = checksum.New(otherTyp, otherVal)
	require.Equal(t, otherTyp, cs.Type())
	require.Equal(t, otherVal, cs.Value())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := checksum.New(typ, val)
			var dst checksum.Checksum

			m := src.ProtoMessage()
			switch actual := m.GetType(); typ {
			default:
				require.EqualValues(t, typ, actual)
			case checksum.TillichZemor:
				require.Equal(t, refs.ChecksumType_TZ, actual)
			case checksum.SHA256:
				require.Equal(t, refs.ChecksumType_SHA256, actual)
			}
			require.Equal(t, val, m.GetSum())
			require.NoError(t, dst.FromProtoMessage(m))
			require.Equal(t, typ, dst.Type())
			require.Equal(t, val, dst.Value())
		})
	})
}

func testTypeConstructor[T [sha256.Size]byte | [tz.Size]byte](
	t *testing.T,
	typ checksum.Type,
	typAPI refs.ChecksumType,
	cons func(T) checksum.Checksum,
) {
	// array generics do not support T[:] op, so randomize through conversion
	b := make([]byte, 128) // more than any popular hash
	//nolint:staticcheck
	rand.Read(b)
	val := T(b)
	cs := cons(val)
	require.Equal(t, typ, cs.Type())
	require.Len(t, cs.Value(), len(val))
	require.Equal(t, val, T(cs.Value()))

	//nolint:staticcheck
	rand.Read(b)
	otherVal := T(b)
	cs = cons(otherVal)
	require.Equal(t, typ, cs.Type())
	require.Len(t, cs.Value(), len(otherVal))
	require.Equal(t, otherVal, T(cs.Value()))

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := cons(val)
			var dst checksum.Checksum

			m := src.ProtoMessage()
			require.Equal(t, typAPI, m.GetType())
			require.Len(t, m.GetSum(), len(val))
			require.Equal(t, val, T(m.GetSum()))
			require.NoError(t, dst.FromProtoMessage(m))
			require.Equal(t, typ, dst.Type())
			require.Len(t, dst.Value(), len(val))
			require.Equal(t, val, T(dst.Value()))
		})
	})
}

func TestNewSHA256(t *testing.T) {
	testTypeConstructor(t, checksum.SHA256, refs.ChecksumType_SHA256, checksum.NewSHA256)
}

func TestNewTZ(t *testing.T) {
	testTypeConstructor(t, checksum.TillichZemor, refs.ChecksumType_TZ, checksum.NewTillichZemor)
}

func TestNewFromHash(t *testing.T) {
	var h hash.Hash = adler32.New() // any non-standard hash just for example
	h.Write([]byte("Hello, world!"))
	hb := []byte{32, 94, 4, 138}

	typ := checksum.Type(rand.Int31())
	cs := checksum.NewFromHash(typ, h)
	require.Equal(t, typ, cs.Type())
	require.Equal(t, hb, cs.Value())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := checksum.NewFromHash(typ, h)
			var dst checksum.Checksum

			m := src.ProtoMessage()
			switch actual := m.GetType(); typ {
			default:
				require.EqualValues(t, typ, actual)
			case checksum.TillichZemor:
				require.Equal(t, refs.ChecksumType_TZ, actual)
			case checksum.SHA256:
				require.Equal(t, refs.ChecksumType_SHA256, actual)
			}
			require.Equal(t, hb, m.GetSum())
			require.NoError(t, dst.FromProtoMessage(m))
			require.Equal(t, typ, dst.Type())
			require.Equal(t, hb, dst.Value())
		})
	})
}

func TestNewFromData(t *testing.T) {
	_, err := checksum.NewFromData(0, nil)
	require.EqualError(t, err, "unsupported checksum type 0")
	var cs checksum.Checksum
	checksum.Calculate(&cs, 0, nil)
	require.Zero(t, cs)
	_, err = checksum.NewFromData(checksum.TillichZemor+1, nil)
	require.EqualError(t, err, "unsupported checksum type 3")
	checksum.Calculate(&cs, checksum.TillichZemor+1, nil)
	require.Zero(t, cs)

	payload := []byte("Hello, world!")
	hSHA256 := [sha256.Size]byte{49, 95, 91, 219, 118, 208, 120, 196, 59, 138, 192, 6, 78, 74, 1, 100, 97, 43, 31, 206, 119, 200, 105, 52, 91, 252, 148, 199, 88, 148, 237, 211}
	hTZ := [tz.Size]byte{0, 0, 1, 66, 73, 241, 7, 149, 192, 36, 14, 221, 202, 138, 110, 191, 0, 0, 1, 201, 196, 220, 152, 176, 23, 253, 146, 173, 98, 151, 156, 140,
		0, 0, 0, 141, 148, 205, 152, 164, 87, 185, 131, 233, 55, 131, 141, 205, 0, 0, 0, 219, 200, 104, 158, 117, 199, 221, 137, 37, 173, 13, 247, 39}

	t.Run("SHA256", func(t *testing.T) {
		c, err := checksum.NewFromData(checksum.SHA256, payload)
		require.NoError(t, err)
		require.Equal(t, hSHA256[:], c.Value())
		require.Equal(t, checksum.SHA256, c.Type())

		c = checksum.Checksum{}
		checksum.Calculate(&c, checksum.SHA256, payload)
		require.Equal(t, hSHA256[:], c.Value())
		require.Equal(t, checksum.SHA256, c.Type())
	})

	t.Run("TillichZemor", func(t *testing.T) {
		c, err := checksum.NewFromData(checksum.TillichZemor, payload)
		require.NoError(t, err)
		require.Equal(t, hTZ[:], c.Value())
		require.Equal(t, checksum.TillichZemor, c.Type())

		c = checksum.Checksum{}
		checksum.Calculate(&c, checksum.TillichZemor, payload)
		require.Equal(t, hTZ[:], c.Value())
		require.Equal(t, checksum.TillichZemor, c.Type())
	})
}

func TestChecksum_SetSHA256(t *testing.T) {
	testTypeConstructor(t, checksum.SHA256, refs.ChecksumType_SHA256, func(b [sha256.Size]byte) (c checksum.Checksum) { c.SetSHA256(b); return })
}

func TestChecksum_SetTillichZemor(t *testing.T) {
	testTypeConstructor(t, checksum.TillichZemor, refs.ChecksumType_TZ, func(b [tz.Size]byte) (c checksum.Checksum) { c.SetTillichZemor(b); return })
}
