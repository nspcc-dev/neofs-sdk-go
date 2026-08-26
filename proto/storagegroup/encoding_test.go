package storagegroup_test

import (
	"testing"

	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/proto/storagegroup"
	"github.com/stretchr/testify/require"
)

func TestStorageGroup_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &storagegroup.StorageGroup{
			Members: []*refs.ObjectID{nil, {}},
		}

		var dst storagegroup.StorageGroup
		require.NoError(t, protoencoding.UnmarshalMessage(protoencoding.MarshalMessage(src), &dst))

		ms := dst.GetMembers()
		require.Len(t, ms, 2)
		require.Equal(t, ms[0], new(refs.ObjectID))
		require.Equal(t, ms[1], new(refs.ObjectID))
	})

	prototest.TestMarshalStable(t, []*storagegroup.StorageGroup{
		{
			ValidationDataSize: prototest.RandUint64(),
			ValidationHash:     prototest.RandChecksum(),
			ExpirationEpoch:    prototest.RandUint64(),
			Members:            prototest.RandObjectIDs(),
		},
	})
}
