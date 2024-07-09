package audit_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/proto/audit"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
)

func TestDataAuditResult_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*audit.DataAuditResult{
		{
			Version:     prototest.RandVersion(),
			AuditEpoch:  prototest.RandUint64(),
			ContainerId: prototest.RandContainerID(),
			PublicKey:   prototest.RandBytes(),
			Complete:    true,
			Requests:    prototest.RandUint32(),
			Retries:     prototest.RandUint32(),
			PassSg:      prototest.RandObjectIDs(),
			FailSg:      prototest.RandObjectIDs(),
			Hit:         prototest.RandUint32(),
			Miss:        prototest.RandUint32(),
			Fail:        prototest.RandUint32(),
			PassNodes:   prototest.RandRepeatedBytes(),
			FailNodes:   prototest.RandRepeatedBytes(),
		},
	})
}
