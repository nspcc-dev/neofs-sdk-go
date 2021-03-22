/*
Package checksum provides primitives to work with checksums of
NeoFS objects and storage groups.

Checksum type provides functionality to specify sum type and value.
For example, retrieving payload sum of the object:
	// recv obj from any source

	checksum := obj.PayloadChecksum()
	raw := checksum.Sum() // byte slice of the sum
	sumType := checksum.Type() // type of the sum
	hexString := checksum.String() // string representation

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

*/
package checksum
