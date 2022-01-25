package address

import (
	"github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	"github.com/nspcc-dev/neofs-sdk-go/object/id/test"
)

// Address returns random object.Address.
func Address() *address.Address {
	x := address.NewAddress()

	x.SetContainerID(cidtest.ID())
	x.SetObjectID(test.ID())

	return x
}
