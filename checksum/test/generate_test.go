package checksumtest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	"github.com/stretchr/testify/require"
)

func TestChecksum(t *testing.T) {
	cs := checksumtest.Checksum()
	require.NotEqual(t, cs, checksumtest.Checksum())

	m := cs.ProtoMessage()
	var cs2 checksum.Checksum
	require.NoError(t, cs2.FromProtoMessage(m))
	require.Equal(t, cs, cs2)
}
