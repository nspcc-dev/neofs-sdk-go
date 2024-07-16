package storagegrouptest

import (
	"math/rand"

	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// StorageGroup returns random storagegroup.StorageGroup.
func StorageGroup() storagegroup.StorageGroup {
	x := storagegroup.New(322, checksumtest.Checksum(), oidtest.IDs(2))
	x.SetExpirationEpoch(rand.Uint64())

	return x
}
