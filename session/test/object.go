package sessiontest

import (
	"math/rand"

	addresstest "github.com/nspcc-dev/neofs-sdk-go/object/address/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// ObjectContext returns session.ObjectContext
// which applies to random operation on a random object.
func ObjectContext() *session.ObjectContext {
	c := session.NewObjectContext()

	setters := []func(){
		c.ForPut,
		c.ForDelete,
		c.ForHead,
		c.ForRange,
		c.ForRangeHash,
		c.ForSearch,
		c.ForGet,
	}

	setters[rand.Uint32()%uint32(len(setters))]()

	addr := addresstest.Address()
	c.ApplyTo(&addr)

	return c
}
