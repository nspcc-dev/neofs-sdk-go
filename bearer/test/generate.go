package bearertest

import (
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
)

// Token returns random bearer.Token.
//
// Resulting token is unsigned.
func Token() (t bearer.Token) {
	t.SetExpiration(3)
	t.SetNotBefore(2)
	t.SetIssuedAt(1)
	t.SetOwnerID(*ownertest.ID())
	t.SetEACLTable(*eacltest.Table())

	return t
}
