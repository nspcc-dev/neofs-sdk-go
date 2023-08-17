package version_test

import (
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// In most of the cases it will be enough to use the latest supported NeoFS API version in SDK.
func ExampleCurrent() {
	_ = version.Current()
}

// It is possible to specify arbitrary version by setting major and minor numbers.
func ExampleVersion_SetMajor() {
	var ver version.Version
	ver.SetMajor(2)
	ver.SetMinor(5)
}
