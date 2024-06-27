package bearertest

import (
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Token returns random bearer.Token.
//
// Resulting token is unsigned.
func Token() (tok bearer.Token) {
	tok.SetExp(3)
	tok.SetNbf(2)
	tok.SetIat(1)
	tok.ForUser(usertest.ID())
	tok.SetEACLTable(eacltest.Table())

	return tok
}
