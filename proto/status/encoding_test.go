package status_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

func TestStatus_Detail_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*status.Status_Detail{
		prototest.RandStatusDetail(),
	})
}

func TestStatus_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*status.Status{
		prototest.RandStatus(),
	})
}
