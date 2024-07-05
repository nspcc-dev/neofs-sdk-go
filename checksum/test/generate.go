package checksumtest

import (
	"crypto/sha256"
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
)

// Checksum returns random checksum.Checksum.
func Checksum() checksum.Checksum {
	var cs [sha256.Size]byte
	//nolint:staticcheck
	rand.Read(cs[:])
	return checksum.NewSHA256(cs)
}
