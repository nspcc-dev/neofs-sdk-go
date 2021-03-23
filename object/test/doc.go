/*
Package objecttest provides functions for convenient testing of object package API.

Note that importing the package into source files is highly discouraged.

Random instance generation functions can be useful when testing expects any value, e.g.:
	import oidtest "github.com/nspcc-dev/neofs-sdk-go/object/test"

	r := objecttest.Range()
	// test the value

	a := objecttest.Attribute()
	// test the value

	s := objecttest.SplitID()
	// test the value

	o := objecttest.Object()
	// test the value

	// etc

*/
package objecttest
