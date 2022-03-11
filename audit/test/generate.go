package audittest

import (
	"github.com/nspcc-dev/neofs-sdk-go/audit"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
)

// Result returns random audit.Result.
func Result() *audit.Result {
	x := audit.NewResult()

	x.SetVersion(versiontest.Version())
	x.SetContainerID(cidtest.ID())
	x.SetPublicKey([]byte("key"))
	x.SetComplete(true)
	x.SetAuditEpoch(44)
	x.SetHit(55)
	x.SetMiss(66)
	x.SetFail(77)
	x.SetRetries(88)
	x.SetRequests(99)
	x.SetFailNodes([][]byte{
		[]byte("node1"),
		[]byte("node2"),
	})
	x.SetPassNodes([][]byte{
		[]byte("node3"),
		[]byte("node4"),
	})
	x.SetPassSG([]oid.ID{*oidtest.ID(), *oidtest.ID()})
	x.SetFailSG([]oid.ID{*oidtest.ID(), *oidtest.ID()})

	return x
}
