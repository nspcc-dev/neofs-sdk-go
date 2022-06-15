package subnettest

import (
	"github.com/nspcc-dev/neofs-sdk-go/subnet"
	subnetidtest "github.com/nspcc-dev/neofs-sdk-go/subnet/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Info generates and returns random subnet.Info.
func Info() (x subnet.Info) {
	x.SetID(subnetidtest.ID())
	x.SetOwner(*usertest.ID())
	return
}
