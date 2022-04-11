package bearertest

import (
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Token returns random bearer.Token.
//
// Resulting token is unsigned.
func Token() (t bearer.Token) {
	t.SetExpiration(3)
	t.SetNotBefore(2)
	t.SetIssuedAt(1)
	t.SetOwnerID(*usertest.ID())
	t.SetEACLTable(*eacltest.Table())

	return t
}
