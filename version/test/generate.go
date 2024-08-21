package versiontest

import (
	"math/rand/v2"

	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Version returns random version.Version.
func Version() (v version.Version) {
	v.SetMajor(rand.Uint32())
	v.SetMinor(rand.Uint32())
	return v
}
