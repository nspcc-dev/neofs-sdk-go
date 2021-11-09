package audittest

import (
	"github.com/nspcc-dev/neofs-sdk-go/audit"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
)

// Generate returns random audit.Result.
func Generate() *audit.Result {
	x := audit.NewResult()

	x.SetVersion(versiontest.Version())
	x.SetContainerID(cidtest.GenerateID())
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
	x.SetPassSG([]*object.ID{objecttest.ID(), objecttest.ID()})
	x.SetFailSG([]*object.ID{objecttest.ID(), objecttest.ID()})

	return x
}
