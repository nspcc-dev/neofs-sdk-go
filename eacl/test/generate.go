package eacltest

import (
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
)

// Target returns random eacl.Target.
func Target() *eacl.Target {
	x := eacl.NewTarget()

	x.SetRole(eacl.RoleSystem)
	x.SetBinaryKeys([][]byte{
		{1, 2, 3},
		{4, 5, 6},
	})

	return x
}

// Record returns random eacl.Record.
func Record() *eacl.Record {
	x := eacl.NewRecord()

	x.SetAction(eacl.ActionAllow)
	x.SetOperation(eacl.OperationRangeHash)
	x.SetTargets(Target(), Target())
	x.AddObjectContainerIDFilter(eacl.MatchStringEqual, cidtest.GenerateID())
	x.AddObjectOwnerIDFilter(eacl.MatchStringNotEqual, ownertest.GenerateID())

	return x
}

func Table() *eacl.Table {
	x := eacl.NewTable()

	x.SetCID(cidtest.GenerateID())
	x.AddRecord(Record())
	x.AddRecord(Record())
	x.SetVersion(*versiontest.Version())

	return x
}
