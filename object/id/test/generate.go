package oidtest

import (
	"crypto/sha256"
	"math/rand"

	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// ID returns random object.ID.
func ID() oid.ID {
	checksum := [sha256.Size]byte{}

	rand.Read(checksum[:])

	return idWithChecksum(checksum)
}

// idWithChecksum returns object.ID initialized
// with specified checksum.
func idWithChecksum(cs [sha256.Size]byte) oid.ID {
	var id oid.ID
	id.SetSHA256(cs)

	return id
}
