/*
Package storagegrouptest provides functions for convenient testing of storagegroup package API.

Note that importing the package into source files is highly discouraged.

Random instance generation functions can be useful when testing expects any value, e.g.:
	import storagegrouptest "github.com/nspcc-dev/neofs-sdk-go/storagegroup/test"

	dec := storagegrouptest.StorageGroup()
	// test the value

*/
package storagegrouptest
