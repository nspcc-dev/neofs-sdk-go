package cidtest

import (
	"crypto/sha256"
	"math/rand"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// GenerateID returns random cid.ID.
func GenerateID() *cid.ID {
	checksum := [sha256.Size]byte{}

	rand.Read(checksum[:])

	return GenerateIDWithChecksum(checksum)
}

// GenerateIDWithChecksum returns cid.ID initialized
// with specified checksum.
func GenerateIDWithChecksum(cs [sha256.Size]byte) *cid.ID {
	id := cid.New()
	id.SetSHA256(cs)

	return id
}
