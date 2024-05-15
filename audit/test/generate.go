package audittest

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/nspcc-dev/neofs-sdk-go/audit"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
)

// Result returns random audit.Result.
func Result() audit.Result {
	var x audit.Result

	x.ForContainer(cidtest.ID())
	auditorKey := make([]byte, 33)
	rand.Read(auditorKey)
	x.SetAuditorKey(auditorKey)
	x.SetCompleted(rand.Int()%2 == 0)
	x.ForEpoch(rand.Uint64())
	x.SetHits(rand.Uint32())
	x.SetMisses(rand.Uint32())
	x.SetFailures(rand.Uint32())
	x.SetRequestsPoR(rand.Uint32())
	x.SetRequestsPoR(rand.Uint32())
	failedNodes := make([][]byte, rand.Int()%4)
	for i := range failedNodes {
		failedNodes[i] = []byte("failed_node_" + strconv.Itoa(i+1))
	}
	x.SetFailedStorageNodes(failedNodes)
	passedNodes := make([][]byte, rand.Int()%4)
	for i := range passedNodes {
		passedNodes[i] = []byte("passed_node_" + strconv.Itoa(i+1))
	}
	x.SetPassedStorageNodes(passedNodes)
	x.SetPassedStorageGroups(oidtest.NIDs(rand.Int() % 4))
	x.SetFailedStorageGroups(oidtest.NIDs(rand.Int() % 4))

	if err := x.Unmarshal(x.Marshal()); err != nil { // to set all defaults
		panic(fmt.Errorf("unexpected encode-decode failure: %w", err))
	}

	return x
}
