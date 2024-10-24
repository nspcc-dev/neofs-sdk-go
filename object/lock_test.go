package object_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

var validLock object.Lock // set by init.

func init() {
	validLock.WriteMembers(anyValidIDs)
}

var validBinLock = []byte{
	10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70,
	246, 8, 139, 247, 174, 53, 60, 10, 34, 10, 32, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225,
	21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 10, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79,
	226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242,
}

func TestLock_WriteMembers(t *testing.T) {
	var l object.Lock
	require.Zero(t, l.NumberOfMembers())
	buf := oidtest.IDs(1)
	el0Cp := buf[0]
	l.ReadMembers(buf)
	require.Equal(t, el0Cp, buf[0])

	l.WriteMembers(anyValidIDs)
	require.EqualValues(t, len(anyValidIDs), l.NumberOfMembers())
	buf = oidtest.IDs(len(anyValidIDs))
	l.ReadMembers(buf)
	require.Equal(t, anyValidIDs, buf)
}

func TestLock_Marshal(t *testing.T) {
	require.Equal(t, validBinLock, validLock.Marshal())
}

func TestLock_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(object.Lock).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "members/empty value", err: "invalid member #1: invalid length 0",
				b: []byte{10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77,
					44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 10, 0, 10, 34, 10, 32, 206, 228, 247, 217, 41, 247,
					159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229,
					161, 202, 197, 242}},
			{name: "members/undersize", err: "invalid member #1: invalid length 31",
				b: []byte{10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77,
					44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 10, 33, 10, 31, 229, 77, 63, 235, 2, 9, 165, 123, 116,
					123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 10, 34,
					10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236,
					112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "members/oversize", err: "invalid member #1: invalid length 33",
				b: []byte{10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77,
					44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 10, 35, 10, 33, 229, 77, 63, 235, 2, 9, 165, 123, 116,
					123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 1,
					10, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35,
					220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.Lock).Unmarshal(tc.b), tc.err)
			})
		}
	})

	l := validLock
	// zero
	require.NoError(t, l.Unmarshal(nil))
	require.Zero(t, l)

	// filled
	require.NoError(t, l.Unmarshal(validBinLock))
	require.Equal(t, validLock, l)
}
