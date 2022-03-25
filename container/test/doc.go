/*
Package containertest provides functions for convenient testing of container package API.

Note that importing the package into source files is highly discouraged.

Random instance generation functions can be useful when testing expects any value, e.g.:
	import containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"

	a := containertest.Attribute()
	// test the value

	c := containertest.Container()
	// test the value

	ua := containertest.UsedSpaceAnnouncement()
	// test the value
*/
package containertest
