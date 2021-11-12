package ownertest

import (
	"crypto/sha256"
	"math/rand"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// ID returns owner.ID calculated
// from a random owner.NEO3Wallet.
func ID() *owner.ID {
	u := make([]byte, owner.NEO3WalletSize)
	u[0] = 0x35
	rand.Read(u[1:21])
	h1 := sha256.Sum256(u[:21])
	h2 := sha256.Sum256(h1[:])
	copy(u[21:], h2[:4])
	return IDFromBytes(u)
}

// IDFromBytes returns owner.ID generated
// from a passed byte slice.
func IDFromBytes(val []byte) *owner.ID {
	idV2 := new(refs.OwnerID)
	idV2.SetValue(val)

	return owner.NewIDFromV2(idV2)
}
