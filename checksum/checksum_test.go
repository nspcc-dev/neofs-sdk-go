package checksum_test

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"hash/adler32"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

func TestChecksum_CopyTo(t *testing.T) {
	src := checksumtest.Checksum()
	var dst checksum.Checksum
	src.CopyTo(&dst)

	require.Equal(t, src.Value(), dst.Value())
	require.Equal(t, src.Type(), dst.Type())

	originVal := src.Value()
	originValCp := bytes.Clone(originVal)
	originVal[0]++
	require.Equal(t, originVal, src.Value())
	require.Equal(t, originValCp, dst.Value())
}

func TestChecksumDecoding(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		for _, testCase := range []struct {
			name, err string
			corrupt   func(*refs.Checksum)
		}{
			{name: "value/nil", err: "missing value", corrupt: func(cs *refs.Checksum) {
				cs.Sum = nil
			}},
			{name: "value/empty", err: "missing value", corrupt: func(cs *refs.Checksum) {
				cs.Sum = []byte{}
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				var src, dst checksum.Checksum
				var m refs.Checksum

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)
			})
		}
	})
}

func TestNewSHA256(t *testing.T) {
	var val [sha256.Size]byte
	rand.Read(val[:])
	cs := checksum.NewSHA256(val)
	require.Equal(t, checksum.SHA256, cs.Type())
	require.Equal(t, val[:], cs.Value())

	otherVal := val
	otherVal[0]++
	cs = checksum.NewSHA256(otherVal)
	require.Equal(t, checksum.SHA256, cs.Type())
	require.Equal(t, otherVal[:], cs.Value())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := checksum.NewSHA256(val)
			var dst checksum.Checksum
			var m refs.Checksum

			src.WriteToV2(&m)
			require.Equal(t, refs.ChecksumType_SHA256, m.Type)
			require.Equal(t, val[:], m.Sum)
			require.NoError(t, dst.ReadFromV2(&m))
			require.Equal(t, checksum.SHA256, dst.Type())
			require.Equal(t, val[:], dst.Value())
		})
	})
}

func TestNewTZ(t *testing.T) {
	var val [tz.Size]byte
	rand.Read(val[:])
	cs := checksum.NewTZ(val)
	require.Equal(t, checksum.TZ, cs.Type())
	require.Equal(t, val[:], cs.Value())

	otherVal := val
	otherVal[0]++
	cs = checksum.NewTZ(otherVal)
	require.Equal(t, checksum.TZ, cs.Type())
	require.Equal(t, otherVal[:], cs.Value())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := checksum.NewTZ(val)
			var dst checksum.Checksum
			var m refs.Checksum

			src.WriteToV2(&m)
			require.Equal(t, refs.ChecksumType_TZ, m.Type)
			require.Equal(t, val[:], m.Sum)
			require.NoError(t, dst.ReadFromV2(&m))
			require.Equal(t, checksum.TZ, dst.Type())
			require.Equal(t, val[:], dst.Value())
		})
	})
}

func TestNewFromHash(t *testing.T) {
	var h hash.Hash = adler32.New() // any hash just for example
	h.Write([]byte("Hello, world!"))
	hb := []byte{32, 94, 4, 138}

	typ := checksum.Type(rand.Uint32() % 256)
	cs := checksum.NewFromHash(typ, h)
	require.Equal(t, typ, cs.Type())
	require.Equal(t, hb, cs.Value())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			src := checksum.NewFromHash(typ, h)
			var dst checksum.Checksum
			var m refs.Checksum

			src.WriteToV2(&m)
			switch typ {
			default:
				require.EqualValues(t, typ, m.Type)
			case checksum.TZ:
				require.Equal(t, refs.ChecksumType_TZ, m.Type)
			case checksum.SHA256:
				require.Equal(t, refs.ChecksumType_SHA256, m.Type)
			}
			require.Equal(t, hb, m.Sum)
			require.NoError(t, dst.ReadFromV2(&m))
			require.Equal(t, typ, dst.Type())
			require.Equal(t, hb, dst.Value())
		})
	})
}

func TestCalculate(t *testing.T) {
	payload := []byte("Hello, world!")
	hSHA256 := [32]byte{49, 95, 91, 219, 118, 208, 120, 196, 59, 138, 192, 6, 78, 74, 1, 100, 97, 43, 31, 206, 119, 200, 105, 52, 91, 252, 148, 199, 88, 148, 237, 211}
	hTZ := [64]byte{0, 0, 1, 66, 73, 241, 7, 149, 192, 36, 14, 221, 202, 138, 110, 191, 0, 0, 1, 201, 196, 220, 152, 176, 23, 253, 146, 173, 98, 151, 156, 140,
		0, 0, 0, 141, 148, 205, 152, 164, 87, 185, 131, 233, 55, 131, 141, 205, 0, 0, 0, 219, 200, 104, 158, 117, 199, 221, 137, 37, 173, 13, 247, 39}

	require.Panics(t, func() { checksum.Calculate(0, []byte("any")) })
	require.Panics(t, func() { checksum.Calculate(checksum.TZ+1, []byte("any")) })

	t.Run("SHA256", func(t *testing.T) {
		c := checksum.Calculate(checksum.SHA256, payload)
		require.Equal(t, checksum.SHA256, c.Type())
		require.Equal(t, hSHA256[:], c.Value())
	})

	t.Run("TZ", func(t *testing.T) {
		c := checksum.Calculate(checksum.TZ, payload)
		require.Equal(t, checksum.TZ, c.Type())
		require.Equal(t, hTZ[:], c.Value())
	})
}
