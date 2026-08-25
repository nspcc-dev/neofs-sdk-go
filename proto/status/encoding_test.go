package status_test

import (
	"testing"

	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/status"
	"github.com/stretchr/testify/require"
)

func TestStatus_Detail_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*status.Status_Detail{
		prototest.RandStatusDetail(),
	})
}

func TestStatus_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &status.Status{
			Details: []*status.Status_Detail{nil, {}},
		}

		var dst status.Status
		require.NoError(t, protoencoding.UnmarshalMessage(protoencoding.MarshalMessage(src), &dst))

		ds := dst.GetDetails()
		require.Len(t, ds, 2)
		require.Equal(t, ds[0], new(status.Status_Detail))
		require.Equal(t, ds[1], new(status.Status_Detail))
	})

	prototest.TestMarshalStable(t, []*status.Status{
		prototest.RandStatus(),
	})
}
