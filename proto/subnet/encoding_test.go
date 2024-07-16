package subnet_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/subnet"
)

func TestSubnetInfo_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*subnet.SubnetInfo{
		{
			Id:    prototest.RandSubnetID(),
			Owner: prototest.RandOwnerID(),
		},
	})
}
