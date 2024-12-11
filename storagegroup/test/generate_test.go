package storagegrouptest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
	storagegrouptest "github.com/nspcc-dev/neofs-sdk-go/storagegroup/test"
	"github.com/stretchr/testify/require"
)

func TestStorageGroup(t *testing.T) {
	sg := storagegrouptest.StorageGroup()
	require.NotEqual(t, sg, storagegrouptest.StorageGroup())

	var sg2 storagegroup.StorageGroup
	require.NoError(t, sg2.Unmarshal(sg.Marshal()))
	require.Equal(t, sg, sg2)

	b, err := sg.MarshalJSON()
	require.NoError(t, err)
	require.NoError(t, sg2.UnmarshalJSON(b))
	require.Equal(t, sg, sg2)

	m := sg.ProtoMessage()
	require.NoError(t, sg2.FromProtoMessage(m))
	require.Equal(t, sg, sg2)
}
