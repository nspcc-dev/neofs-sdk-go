package storagegroup_test

import (
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
)

// StorageGroup type groups verification values for Data Audit sessions.
func ExampleStorageGroup_validation() {
	// receive sg info

	var sg storagegroup.StorageGroup

	sg.ExpirationEpoch()    // expiration of the storage group
	sg.Members()            // objects in the group
	sg.ValidationDataHash() // hash for objects validation
	sg.ValidationDataSize() // total objects' payload size
}
