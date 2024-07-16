package storagegroup_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/storagegroup"
)

func TestStorageGroup_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*storagegroup.StorageGroup{
		{
			ValidationDataSize: prototest.RandUint64(),
			ValidationHash:     prototest.RandChecksum(),
			ExpirationEpoch:    prototest.RandUint64(),
			Members:            prototest.RandObjectIDs(),
		},
	})
}
