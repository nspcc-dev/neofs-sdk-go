package subnetidtest

import (
	"math/rand"

	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
)

// GenerateID generates and returns random subnetid.ID using math/rand.Uint32.
func GenerateID() *subnetid.ID {
	var id subnetid.ID

	id.SetNumber(rand.Uint32())

	return &id
}
