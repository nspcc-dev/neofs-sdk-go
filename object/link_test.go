package object_test

import (
	"slices"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

var (
	anyValidMeasuredObjects []object.MeasuredObject // set by init.
)

var validLink object.Link // set by init.

func init() {
	anyValidMeasuredObjects = make([]object.MeasuredObject, 3)
	anyValidMeasuredObjects[0].SetObjectSize(2473604180)
	anyValidMeasuredObjects[0].SetObjectID(oid.ID{22, 197, 235, 181, 5, 250, 3, 235, 44, 72, 182, 111, 153, 102, 51, 102, 159,
		140, 45, 99, 49, 239, 153, 40, 115, 14, 122, 54, 234, 22, 40, 17})
	anyValidMeasuredObjects[1].SetObjectSize(2257112744)
	anyValidMeasuredObjects[1].SetObjectID(oid.ID{251, 27, 231, 203, 246, 154, 82, 100, 130, 18, 38, 45, 77, 24, 73, 14, 101,
		240, 94, 211, 254, 55, 32, 76, 29, 234, 229, 252, 181, 122, 51, 174})
	anyValidMeasuredObjects[2].SetObjectSize(3532081186)
	anyValidMeasuredObjects[2].SetObjectID(oid.ID{227, 93, 189, 146, 92, 112, 76, 244, 236, 3, 94, 40, 212, 131, 229, 216,
		130, 248, 195, 85, 23, 68, 65, 234, 106, 157, 101, 185, 227, 85, 182, 106})

	validLink.SetObjects(anyValidMeasuredObjects)
}

// corresponds to validLink.
var validBinLink = []byte{
	10, 42, 10, 34, 10, 32, 22, 197, 235, 181, 5, 250, 3, 235, 44, 72, 182, 111, 153, 102, 51, 102, 159, 140, 45, 99, 49, 239,
	153, 40, 115, 14, 122, 54, 234, 22, 40, 17, 16, 212, 232, 192, 155, 9, 10, 42, 10, 34, 10, 32, 251, 27, 231, 203, 246,
	154, 82, 100, 130, 18, 38, 45, 77, 24, 73, 14, 101, 240, 94, 211, 254, 55, 32, 76, 29, 234, 229, 252, 181, 122, 51, 174,
	16, 168, 157, 163, 180, 8, 10, 42, 10, 34, 10, 32, 227, 93, 189, 146, 92, 112, 76, 244, 236, 3, 94, 40, 212, 131, 229, 216,
	130, 248, 195, 85, 23, 68, 65, 234, 106, 157, 101, 185, 227, 85, 182, 106, 16, 162, 144, 157, 148, 13,
}

func TestLink_SetObjects(t *testing.T) {
	var l object.Link
	require.Empty(t, l.Objects())

	l.SetObjects(anyValidMeasuredObjects)
	require.Equal(t, anyValidMeasuredObjects, l.Objects())

	otherObjs := slices.Clone(anyValidMeasuredObjects)
	for i := range otherObjs {
		otherObjs[i].SetObjectSize(otherObjs[i].ObjectSize() + 1)
	}
	l.SetObjects(otherObjs)
	require.Equal(t, otherObjs, l.Objects())
}

func TestLink_Marshal(t *testing.T) {
	require.Equal(t, validBinLink, validLink.Marshal())
}

func TestLink_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(object.Link).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "members/empty value", err: "invalid member #1: invalid length 0",
				b: []byte{10, 36, 10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190,
					224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 10, 2, 10, 0, 10, 36, 10, 34, 10, 32, 206,
					228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101,
					24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "members/undersize", err: "invalid member #1: invalid length 31",
				b: []byte{10, 36, 10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190,
					224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 10, 35, 10, 33, 10, 31, 229, 77, 63, 235,
					2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57,
					253, 127, 179, 10, 36, 10, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16,
					102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "members/oversize", err: "invalid member #1: invalid length 33",
				b: []byte{10, 36, 10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190,
					224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 10, 37, 10, 35, 10, 33, 229, 77, 63, 235,
					2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57,
					253, 127, 179, 235, 1, 10, 36, 10, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153,
					133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.Link).Unmarshal(tc.b), tc.err)
			})
		}
	})

	l := validLink
	// zero
	require.NoError(t, l.Unmarshal(nil))
	require.Zero(t, l)

	// filled
	require.NoError(t, l.Unmarshal(validBinLink))
	require.Equal(t, validLink, l)
}
