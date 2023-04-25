package usertest

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ID returns random user.ID.
func ID() *user.ID {
	key, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	var x user.ID
	if err = user.IDFromSigner(&x, neofsecdsa.Signer(key.PrivateKey)); err != nil {
		return nil
	}

	return &x
}
