package tombstone_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/proto/tombstone"
	"github.com/stretchr/testify/require"
)

func TestTombstone_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &tombstone.Tombstone{
			Members: []*refs.ObjectID{nil, {}},
		}

		var dst tombstone.Tombstone
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		ds := dst.GetMembers()
		require.Len(t, ds, 2)
		require.Equal(t, ds[0], new(refs.ObjectID))
		require.Equal(t, ds[1], new(refs.ObjectID))
	})

	prototest.TestMarshalStable(t, []*tombstone.Tombstone{
		{
			ExpirationEpoch: prototest.RandUint64(),
			SplitId:         prototest.RandBytes(),
			Members:         prototest.RandObjectIDs(),
		},
	})
}
