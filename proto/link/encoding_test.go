package link_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/link"
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
	prototest.TestMarshalStable(t, []*link.Link{
		{Children: randObjectMeasurements()},
	})
}
