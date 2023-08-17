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

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.storagegroup package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleStorageGroup_WriteToV2() {
	// import apiGoStoragegroup "github.com/nspcc-dev/neofs-api-go/v2/storagegroup"

	var sg storagegroup.StorageGroup
	var msg apiGoStoragegroup.StorageGroup
	sg.WriteToV2(&msg)

	// send msg
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.storagegroup package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleStorageGroup_ReadFromV2() {
	// import apiGoStoragegroup "github.com/nspcc-dev/neofs-api-go/v2/storagegroup"

	// recv msg

	var sg storagegroup.StorageGroup
	var msg apiGoStoragegroup.StorageGroup
	_ = sg.ReadFromV2(msg)

	// process sg
}
