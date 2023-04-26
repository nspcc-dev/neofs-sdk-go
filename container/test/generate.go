package containertest

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Container returns random container.Container.
func Container(t *testing.T) (x container.Container) {
	owner := usertest.ID(t)

	x.Init()
	x.SetAttribute("some attribute", "value")
	x.SetOwner(*owner)
	x.SetBasicACL(BasicACL())
	x.SetPlacementPolicy(netmaptest.PlacementPolicy())

	return x
}

// SizeEstimation returns random container.SizeEstimation.
func SizeEstimation() (x container.SizeEstimation) {
	x.SetContainer(cidtest.ID())
	x.SetEpoch(rand.Uint64())
	x.SetValue(rand.Uint64())

	return x
}

// BasicACL returns random acl.Basic.
func BasicACL() (x acl.Basic) {
	x.FromBits(rand.Uint32())
	return
}
