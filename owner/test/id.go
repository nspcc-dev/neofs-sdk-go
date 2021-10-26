package ownertest

import (
	"math/rand"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// GenerateID returns owner.ID calculated
// from a random owner.NEO3Wallet.
func GenerateID() *owner.ID {
	data := make([]byte, owner.NEO3WalletSize)

	rand.Read(data)

	return GenerateIDFromBytes(data)
}

// GenerateIDFromBytes returns owner.ID generated
// from a passed byte slice.
func GenerateIDFromBytes(val []byte) *owner.ID {
	idV2 := new(refs.OwnerID)
	idV2.SetValue(val)

	return owner.NewIDFromV2(idV2)
}
