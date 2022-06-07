package containertest

import (
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
)

// Attribute returns random container.Attribute.
func Attribute() *container.Attribute {
	x := container.NewAttribute()

	x.SetKey("key")
	x.SetValue("value")

	return x
}

// Attributes returns random container.Attributes.
func Attributes() container.Attributes {
	return container.Attributes{*Attribute(), *Attribute()}
}

// Container returns random container.Container.
func Container() *container.Container {
	x := container.New()
	ver := versiontest.Version()

	x.SetVersion(&ver)
	x.SetAttributes(Attributes())
	x.SetOwnerID(usertest.ID())
	x.SetBasicACL(123)
	p := netmaptest.PlacementPolicy()
	x.SetPlacementPolicy(&p)

	return x
}

// UsedSpaceAnnouncement returns random container.UsedSpaceAnnouncement.
func UsedSpaceAnnouncement() *container.UsedSpaceAnnouncement {
	x := container.NewAnnouncement()

	x.SetContainerID(cidtest.ID())
	x.SetEpoch(55)
	x.SetUsedSpace(999)

	return x
}
