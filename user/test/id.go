package usertest

import (
	"math/rand"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ID returns random user.ID.
func ID() user.ID {
	var h util.Uint160
	//nolint:staticcheck
	rand.Read(h[:])
	var res user.ID
	res.SetScriptHash(h)
	return res
}

// OtherID returns random user.ID other than any given one.
func OtherID(vs ...user.ID) user.ID {
loop:
	for {
		v := ID()
		for i := range vs {
			if v.Equals(vs[i]) {
				continue loop
			}
		}
		return v
	}
}

// IDs returns n random user.ID instances.
func IDs(n int) []user.ID {
	res := make([]user.ID, n)
	for i := range res {
		res[i] = ID()
	}
	return res
}
