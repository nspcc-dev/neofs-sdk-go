package container_test

import (
	"time"

	apiGoContainer "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// To create new container in the NeoFS network Container instance should be initialized.
func ExampleContainer_Init() {
	// import "github.com/nspcc-dev/neofs-sdk-go/container/acl"
	// import "github.com/nspcc-dev/neofs-sdk-go/user"
	// import "github.com/nspcc-dev/neofs-sdk-go/netmap"

	var account user.ID

	var cnr container.Container
	cnr.Init()

	// required fields
	cnr.SetOwner(account)
	cnr.SetBasicACL(acl.PublicRWExtended)

	// optional
	cnr.SetName("awesome container name")
	cnr.SetCreationTime(time.Now())
	// ...

	// placement policy and replicas definition is required
	var pp netmap.PlacementPolicy
	pp.SetContainerBackupFactor(1)

	var rd netmap.ReplicaDescriptor
	rd.SetNumberOfObjects(1)
	pp.AddReplicas(rd)

	cnr.SetPlacementPolicy(pp)
}

// After the container is persisted in the NeoFS network, applications can process
// it using the instance of Container types.
func ExampleContainer_Unmarshal() {
	// recv binary container

	var bin []byte
	var cnr container.Container

	_ = cnr.Unmarshal(bin)

	// process the container data
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.container package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleContainer_WriteToV2() {
	// import apiGoContainer "github.com/nspcc-dev/neofs-api-go/v2/container"

	var cnr container.Container
	var msg apiGoContainer.Container

	cnr.WriteToV2(&msg)
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.container package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleContainer_ReadFromV2() {
	// import apiGoContainer "github.com/nspcc-dev/neofs-api-go/v2/container"

	var cnr container.Container
	var msg apiGoContainer.Container

	_ = cnr.ReadFromV2(msg)

	// ...
}
