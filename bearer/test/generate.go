package bearertest

import (
	"fmt"
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Token returns random bearer.Token. To get unsigned token, use [UnsignedToken].
func Token() bearer.Token {
	usr, _ := usertest.TwoUsers()
	tok := UnsignedToken()
	if err := tok.Sign(usr); err != nil {
		panic(fmt.Errorf("unexpected sign failure: %w", err))
	}
	return tok
}

// UnsignedToken returns random unsigned bearer.Token. To get signed token, use
// [Token].
func UnsignedToken() bearer.Token {
	var tok bearer.Token
	tok.SetExp(uint64(rand.Int()))
	tok.SetNbf(uint64(rand.Int()))
	tok.SetIat(uint64(rand.Int()))
	tok.ForUser(usertest.ID())
	tok.SetEACLTable(eacltest.Table())
	return tok
}
