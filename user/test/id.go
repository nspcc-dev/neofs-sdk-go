package usertest

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ID returns random user.ID.
func ID(tb testing.TB) *user.ID {
	var x user.ID
	s := test.RandomSignerRFC6979(tb)
	x = s.UserID()

	return &x
}
