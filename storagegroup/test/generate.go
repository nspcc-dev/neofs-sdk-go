package storagegrouptest

import (
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// StorageGroup returns random storagegroup.StorageGroup.
func StorageGroup() *storagegroup.StorageGroup {
	x := storagegroup.New()

	x.SetExpirationEpoch(66)
	x.SetValidationDataSize(322)
	x.SetValidationDataHash(checksumtest.Checksum())
	x.SetMembers([]*oid.ID{test.ID(), test.ID()})

	return x
}
