package audit_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/audit"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

func TestDataAuditResult_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &audit.DataAuditResult{
			PassSg: []*refs.ObjectID{nil, {}},
			FailSg: []*refs.ObjectID{nil, {}},
		}

		var dst audit.DataAuditResult
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		ps := dst.GetPassSg()
		require.Len(t, ps, 2)
		require.Equal(t, ps[0], new(refs.ObjectID))
		require.Equal(t, ps[1], new(refs.ObjectID))
		fs := dst.GetFailSg()
		require.Len(t, fs, 2)
		require.Equal(t, fs[0], new(refs.ObjectID))
		require.Equal(t, fs[1], new(refs.ObjectID))
	})

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
