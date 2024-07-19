package eacltest

import (
	"math/rand"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Target returns random eacl.Target.
func Target() eacl.Target {
	if rand.Int()%2 == 0 {
		return eacl.NewTargetByRole(eacl.Role(rand.Uint32()))
	}
	return eacl.NewTargetByAccounts(usertest.IDs(1 + rand.Intn(10)))
}

// Record returns random eacl.Record.
func Record() eacl.Record {
	return eacl.ConstructRecord(eacl.ActionAllow, eacl.OperationRangeHash, []eacl.Target{Target(), Target()},
		eacl.NewFilterObjectsFromContainer(cidtest.ID()),
		eacl.NewFilterObjectOwnerEquals(usertest.ID()),
	)
}

func Table() eacl.Table {
	return eacl.NewTableForContainer(cidtest.ID(), []eacl.Record{Record(), Record()})
}
