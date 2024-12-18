package link_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/link"
	"github.com/stretchr/testify/require"
)

// returns random link.Link_MeasuredObject with all non-zero fields.
func randObjectMeasurement() *link.Link_MeasuredObject {
	return &link.Link_MeasuredObject{
		Id:   prototest.RandObjectID(),
		Size: prototest.RandUint32(),
	}
}

func randObjectMeasurements() []*link.Link_MeasuredObject {
	return prototest.RandRepeated(randObjectMeasurement)
}

func TestLink_MeasuredObject_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*link.Link_MeasuredObject{
		randObjectMeasurement(),
	})
}

func TestLink_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &link.Link{
			Children: []*link.Link_MeasuredObject{nil, {}},
		}

		var dst link.Link
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		cs := dst.GetChildren()
		require.Len(t, cs, 2)
		require.Equal(t, cs[0], new(link.Link_MeasuredObject))
		require.Equal(t, cs[1], new(link.Link_MeasuredObject))
	})

	prototest.TestMarshalStable(t, []*link.Link{
		{Children: randObjectMeasurements()},
	})
}
