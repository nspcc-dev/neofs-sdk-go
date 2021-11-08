package tokentest

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	"github.com/nspcc-dev/neofs-sdk-go/token"
)

// Generate returns random token.BearerToken.
//
// Resulting token is unsigned.
func Generate() *token.BearerToken {
	x := token.NewBearerToken()

	x.SetLifetime(3, 2, 1)
	x.SetOwner(ownertest.GenerateID())
	x.SetEACLTable(eacltest.Table())

	return x
}

// GenerateSigned returns signed random token.BearerToken.
//
// Panics if token could not be signed (actually unexpected).
func GenerateSigned() *token.BearerToken {
	tok := Generate()

	p, err := keys.NewPrivateKey()
	if err != nil {
		panic(err)
	}

	err = tok.SignToken(&p.PrivateKey)
	if err != nil {
		panic(err)
	}

	return tok
}
