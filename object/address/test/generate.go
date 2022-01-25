package address

import (
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
)

// Address returns random object.Address.
func Address() *address.Address {
	x := address.NewAddress()

	x.SetContainerID(cidtest.ID())
	x.SetObjectID(oidtest.ID())

	return x
}
