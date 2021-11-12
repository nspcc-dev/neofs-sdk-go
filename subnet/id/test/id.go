package subnetidtest

import (
	"math/rand"

	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
)

// ID generates and returns random subnetid.ID using math/rand.Uint32.
func ID() *subnetid.ID {
	var id subnetid.ID

	id.SetNumber(rand.Uint32())

	return &id
}
