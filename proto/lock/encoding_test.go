package lock_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/lock"
)

func TestLock_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*lock.Lock{
		{Members: prototest.RandObjectIDs()},
	})
}
