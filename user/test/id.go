package usertest

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ID returns random user.ID.
func ID() *user.ID {
	key, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	var x user.ID
	user.IDFromKey(&x, key.PrivateKey.PublicKey)

	return &x
}
