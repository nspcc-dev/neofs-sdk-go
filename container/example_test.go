package container_test

import (
	"time"

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

	var rd netmap.ReplicaDescriptor
	rd.SetNumberOfObjects(1)

	// placement policy and replicas definition is required
	var pp netmap.PlacementPolicy
	pp.SetContainerBackupFactor(1)
	pp.SetReplicas([]netmap.ReplicaDescriptor{rd})

	cnr.SetPlacementPolicy(pp)
}

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleContainer_marshalling() {
	// On the client side.

	var cnr container.Container
	msg := cnr.ProtoMessage()
	// *send message*

	// On the server side.

	_ = cnr.FromProtoMessage(msg)
}
