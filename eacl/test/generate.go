package eacltest

import (
	"math/rand"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
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
	x := eacl.NewRecord()

	x.SetAction(eacl.ActionAllow)
	x.SetOperation(eacl.OperationRangeHash)
	x.SetTargets(Target(), Target())
	x.AddObjectContainerIDFilter(eacl.MatchStringEqual, cidtest.ID())
	usr := usertest.ID()
	x.AddObjectOwnerIDFilter(eacl.MatchStringNotEqual, &usr)

	return *x
}

func Table() eacl.Table {
	x := eacl.NewTable()

	x.SetCID(cidtest.ID())
	r1 := Record()
	x.AddRecord(&r1)
	r2 := Record()
	x.AddRecord(&r2)
	x.SetVersion(versiontest.Version())

	return *x
}
