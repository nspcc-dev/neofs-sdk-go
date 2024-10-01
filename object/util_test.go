package object_test

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

const (
	anyValidExpirationEpoch = uint64(6053221788077248524)
)

var (
	anyValidSplitIDBytes = []byte{224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41}
	anyValidSplitID      = object.NewSplitIDFromV2(anyValidSplitIDBytes)
	anyValidIDs          = []oid.ID{
		{178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8,
			139, 247, 174, 53, 60},
		{229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243,
			57, 253, 127, 179, 235},
		{206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235,
			126, 173, 229, 161, 202, 197, 242},
	}
)

func protoIDFromBytes(b []byte) *refs.ObjectID {
	var id refs.ObjectID
	id.SetValue(b)
	return &id
}
