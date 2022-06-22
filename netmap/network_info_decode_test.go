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
