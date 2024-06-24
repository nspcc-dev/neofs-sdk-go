package container_test

import (
	"encoding/json"
	"fmt"
	"time"

	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// To create new container in the NeoFS network Container instance should be initialized.
func ExampleNew() {
	var account user.ID
	var policy netmap.PlacementPolicy
	if err := policy.DecodeString("REP 3"); err != nil {
		fmt.Printf("failed to init policy: %v\n", err)
		return
	}

	cnr := container.New(account, acl.PublicRWExtended, policy)

	cnr.SetBasicACL(acl.PublicRWExtended)

	// optional
	cnr.SetName("awesome container name")
	cnr.SetCreationTime(time.Now())
	cnr.SetAttribute("attr_1", "val_1")
	cnr.SetAttribute("attr_2", "val_2")

	var domain container.Domain
	domain.SetName("my-cnr1")
	cnr.SetDomain(domain)

	j, err := json.MarshalIndent(cnr, "", " ")
	if err != nil {
		fmt.Printf("failed to encode container: %v\n", err)
		return
	}

	fmt.Println(string(j))
}

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleContainer_marshalling() {
	// On the client side.

	var cnr container.Container
	var msg apicontainer.Container
	cnr.WriteToV2(&msg)
	// *send message*

	// On the server side.

	_ = cnr.ReadFromV2(&msg)
}
