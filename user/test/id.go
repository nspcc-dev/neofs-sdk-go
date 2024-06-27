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
