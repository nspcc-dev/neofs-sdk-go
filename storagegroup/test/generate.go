package storagegrouptest

import (
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// StorageGroup returns random storagegroup.StorageGroup.
func StorageGroup() storagegroup.StorageGroup {
	var x storagegroup.StorageGroup

	x.SetExpirationEpoch(66)
	x.SetValidationDataSize(322)
	x.SetValidationDataHash(checksumtest.Checksum())
	x.SetMembers([]oid.ID{oidtest.ID(), oidtest.ID()})

	return x
}
