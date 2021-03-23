/*
Package addresstest provides functions for convenient testing of address package API.

Note that importing the package into source files is highly discouraged.

Random instance generation functions can be useful when testing expects any value, e.g.:
	import accountingtest "github.com/nspcc-dev/neofs-sdk-go/object/address/test"

	a := addresstest.Address()
	// test the value

*/
package addresstest
