package storagegrouptest

import (
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// Generate returns random storagegroup.StorageGroup.
func Generate() *storagegroup.StorageGroup {
	x := storagegroup.New()

	x.SetExpirationEpoch(66)
	x.SetValidationDataSize(322)
	x.SetValidationDataHash(checksumtest.Checksum())
	x.SetMembers([]*object.ID{objecttest.ID(), objecttest.ID()})

	return x
}
