/*
Package checksum provides primitives to work with checksums.

Checksum is a basic type of data checksums.
For example, calculating checksums:
	// retrieving any payload for hashing

	var sha256Sum Checksum
	Calculate(&sha256Sum, SHA256, payload) // sha256Sum contains SHA256 hash of the payload

	var tzSum Checksum
	Calculate(&tzSum, TZ, payload) // tzSum contains TZ hash of the payload

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

*/
package checksum
