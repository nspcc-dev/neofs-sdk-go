package eacltest

import (
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
)

// Target returns random eacl.Target.
func Target() eacl.Target {
	x := eacl.NewTarget()

	x.SetRole(eacl.RoleSystem)
	x.SetBinaryKeys([][]byte{
		{1, 2, 3},
		{4, 5, 6},
	})

	return *x
}

// Record returns random eacl.Record.
func Record(tb testing.TB) eacl.Record {
	x := eacl.NewRecord()

	x.SetAction(eacl.ActionAllow)
	x.SetOperation(eacl.OperationRangeHash)
	x.SetTargets(Target(), Target())
	x.AddObjectContainerIDFilter(eacl.MatchStringEqual, cidtest.ID())
	usr := usertest.ID(tb)
	x.AddObjectOwnerIDFilter(eacl.MatchStringNotEqual, &usr)

	return *x
}

func Table(tb testing.TB) eacl.Table {
	x := eacl.NewTable()

	x.SetCID(cidtest.ID())
	r1 := Record(tb)
	x.AddRecord(&r1)
	r2 := Record(tb)
	x.AddRecord(&r2)
	x.SetVersion(versiontest.Version())

	return *x
}
