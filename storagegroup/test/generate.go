package storagegrouptest

import (
	"math/rand/v2"

	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// StorageGroup returns random storagegroup.StorageGroup.
func StorageGroup() storagegroup.StorageGroup {
	n := 1 + rand.N(10)
	x := storagegroup.New(rand.Uint64(), checksumtest.Checksum(), oidtest.IDs(n))
	x.SetExpirationEpoch(rand.Uint64())

	return x
}
