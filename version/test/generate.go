package test

import (
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Version returns random version.Version.
func Version() *version.Version {
	x := version.New()

	x.SetMajor(rand.Uint32())
	x.SetMinor(rand.Uint32())

	return x
}
