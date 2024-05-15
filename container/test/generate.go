package containertest

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Container returns random container.Container.
func Container() container.Container {
	x := container.New(usertest.ID(), BasicACL(), netmaptest.PlacementPolicy())
	x.SetName("name_" + strconv.Itoa(rand.Int()))
	x.SetCreationTime(time.Now())
	x.SetHomomorphicHashingDisabled(rand.Int()%2 == 0)
	var d container.Domain
	d.SetName("domain_" + strconv.Itoa(rand.Int()))
	d.SetZone("zone_" + strconv.Itoa(rand.Int()))
	x.SetDomain(d)

	nAttr := rand.Int() % 4
	for i := 0; i < nAttr; i++ {
		si := strconv.Itoa(rand.Int())
		x.SetAttribute("key_"+si, "val_"+si)
	}

	return x
}

// SizeEstimation returns random container.SizeEstimation.
func SizeEstimation() container.SizeEstimation {
	var x container.SizeEstimation
	x.SetContainer(cidtest.ID())
	x.SetEpoch(rand.Uint64())
	x.SetValue(rand.Uint64())

	return x
}

// BasicACL returns random acl.Basic.
func BasicACL() acl.Basic {
	var x acl.Basic
	x.FromBits(rand.Uint32())
	return x
}
