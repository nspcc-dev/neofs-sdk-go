package storagegroup_test

import (
	apiGoStoragegroup "github.com/nspcc-dev/neofs-api-go/v2/storagegroup"
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

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleStorageGroup_marshalling() {
	// import apiGoStoragegroup "github.com/nspcc-dev/neofs-api-go/v2/storagegroup"

	// On the client side.

	var sg storagegroup.StorageGroup
	var msg apiGoStoragegroup.StorageGroup
	sg.WriteToV2(&msg)
	// *send message*

	// On the server side.

	_ = sg.ReadFromV2(msg)
}
