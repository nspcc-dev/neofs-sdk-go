package containertest

import (
	"math/rand/v2"

	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Container returns random container.Container.
func Container() (x container.Container) {
	owner := usertest.ID()

	x.Init()
	x.SetAttribute("some attribute", "value")
	x.SetOwner(owner)
	x.SetBasicACL(BasicACL())
	x.SetPlacementPolicy(netmaptest.PlacementPolicy())

	return x
}

// BasicACL returns random acl.Basic.
func BasicACL() (x acl.Basic) {
	x.FromBits(rand.Uint32())
	return
}
