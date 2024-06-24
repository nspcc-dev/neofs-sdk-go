package storagegrouptest

import (
	"math/rand"

	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// StorageGroup returns random storagegroup.StorageGroup.
func StorageGroup() storagegroup.StorageGroup {
	var x storagegroup.StorageGroup
	x.SetValidationDataSize(rand.Uint64())
	x.SetValidationDataHash(checksumtest.Checksum())
	x.SetMembers(oidtest.NIDs(1 + rand.Int()%3))

	return x
}
