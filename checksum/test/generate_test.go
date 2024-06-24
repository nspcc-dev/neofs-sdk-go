package checksumtest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	"github.com/stretchr/testify/require"
)

func TestChecksum(t *testing.T) {
	cs := checksumtest.Checksum()
	require.NotEqual(t, cs, checksumtest.Checksum())

	var m refs.Checksum
	cs.WriteToV2(&m)
	var cs2 checksum.Checksum
	require.NoError(t, cs2.ReadFromV2(&m))
	require.Equal(t, cs, cs2)
}
