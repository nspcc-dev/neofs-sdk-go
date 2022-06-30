package netmap

import (
	"math/big"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/encoding/bigint"
	"github.com/stretchr/testify/require"
)

func TestDecodeUint64(t *testing.T) {
	testCases := []uint64{
		0,
		12,
		129,
		0x1234,
		0x12345678,
		0x1234567891011,
	}

	for _, expected := range testCases {
		val := bigint.ToBytes(big.NewInt(int64(expected)))

		actual, err := decodeConfigValueUint64(val)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
}

func TestDecodeBool(t *testing.T) {
	testCases := []struct {
		expected bool
		raw      []byte
	}{
		{
			false,
			[]byte{0},
		},
		{
			false,
			[]byte{0, 0, 0, 0},
		},
		{
			true,
			[]byte{1},
		},
		{
			true,
			[]byte{1, 1, 1, 1, 1},
		},
		{
			true,
			[]byte{0, 0, 0, 0, 1}, // neo-go casts any value that does not consist of zeroes as `true`
		},
	}

	for _, test := range testCases {
		actual, err := decodeConfigValueBool(test.raw)
		require.NoError(t, err)

		require.Equal(t, test.expected, actual)
	}
}
