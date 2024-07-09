package tombstone_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/tombstone"
)

func TestTombstone_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*tombstone.Tombstone{
		{
			ExpirationEpoch: prototest.RandUint64(),
			SplitId:         prototest.RandBytes(),
			Members:         prototest.RandObjectIDs(),
		},
	})
}
