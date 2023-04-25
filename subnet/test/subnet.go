package subnettest

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/subnet"
	subnetidtest "github.com/nspcc-dev/neofs-sdk-go/subnet/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Info generates and returns random subnet.Info.
func Info(t *testing.T) (x subnet.Info) {
	x.SetID(subnetidtest.ID())
	x.SetOwner(*usertest.ID(t))
	return
}
