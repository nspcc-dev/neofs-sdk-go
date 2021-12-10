package ownertest

import (
	"math/rand"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// ID returns owner.ID calculated
// from a random owner.NEO3Wallet.
func ID() *owner.ID {
	u := util.Uint160{}
	rand.Read(u[:])

	addr := address.Uint160ToString(u)
	data, err := base58.Decode(addr)
	if err != nil {
		panic(err)
	}
	return IDFromBytes(data)
}

// IDFromBytes returns owner.ID generated
// from a passed byte slice.
func IDFromBytes(val []byte) *owner.ID {
	idV2 := new(refs.OwnerID)
	idV2.SetValue(val)

	return owner.NewIDFromV2(idV2)
}
