package sessiontest

import (
	"math/rand"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

// ContainerContext returns session.ContainerContext
// which applies to random operation on a random container.
func ContainerContext() *session.ContainerContext {
	c := session.NewContainerContext()

	setters := []func(){
		c.ForPut,
		c.ForDelete,
		c.ForSetEACL,
	}

	setters[rand.Uint32()%uint32(len(setters))]()

	c.ApplyTo(cidtest.ID())

	return c
}
