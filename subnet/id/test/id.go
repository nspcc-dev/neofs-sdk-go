package subnetidtest

import (
	"math/rand"

	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
)

// ID generates and returns random subnetid.ID.
func ID() (x subnetid.ID) {
	x.SetNumeric(rand.Uint32())
	return
}
