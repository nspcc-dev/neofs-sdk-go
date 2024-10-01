package container_test

import (
	"testing"

	v2container "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/stretchr/testify/require"
)

const (
	anyValidEpoch  = uint64(16115098591605387680)
	anyValidVolume = uint64(18109042538078259560)
)

var (
	anyValidID = cid.ID{5, 144, 209, 65, 128, 32, 195, 190, 53, 29, 166, 168, 26, 57, 63, 235, 163, 61, 76, 6, 21, 12,
		102, 38, 68, 146, 142, 239, 28, 161, 20, 207}
)

var validSizeEstimation container.SizeEstimation // set by init.

func init() {
	validSizeEstimation.SetEpoch(anyValidEpoch)
	validSizeEstimation.SetContainer(anyValidID)
	validSizeEstimation.SetValue(anyValidVolume)
}

func TestSizeEstimation_Epoch(t *testing.T) {
	var val container.SizeEstimation
	require.Zero(t, val.Epoch())

	val.SetEpoch(anyValidEpoch)
	require.EqualValues(t, anyValidEpoch, val.Epoch())

	val.SetEpoch(anyValidEpoch + 1)
	require.EqualValues(t, anyValidEpoch+1, val.Epoch())
}

func TestSizeEstimation_Container(t *testing.T) {
	var val container.SizeEstimation
	require.Zero(t, val.Container())

	val.SetContainer(anyValidID)
	require.True(t, val.Container() == anyValidID)

	otherContainer := cidtest.OtherID(anyValidID)
	val.SetContainer(otherContainer)
	require.True(t, val.Container() == otherContainer)
}

func TestSizeEstimation_Value(t *testing.T) {
	var val container.SizeEstimation
	require.Zero(t, val.Value())

	val.SetValue(anyValidVolume)
	require.EqualValues(t, anyValidVolume, val.Value())

	val.SetValue(anyValidVolume + 1)
	require.EqualValues(t, anyValidVolume+1, val.Value())
}

func protoIDFromBytes(b []byte) *refs.ContainerID {
	var m refs.ContainerID
	m.SetValue(b)
	return &m
}

func TestSizeEstimation_ReadFromV2(t *testing.T) {
	var m v2container.UsedSpaceAnnouncement
	m.SetEpoch(anyValidEpoch)
	m.SetContainerID(protoIDFromBytes(anyValidID[:]))
	m.SetUsedSpace(anyValidVolume)

	var val container.SizeEstimation
	require.NoError(t, val.ReadFromV2(m))
	require.EqualValues(t, anyValidEpoch, val.Epoch())
	require.Equal(t, anyValidID, val.Container())
	require.EqualValues(t, anyValidVolume, val.Value())

	// reset optional fields
	m.SetEpoch(0)
	m.SetUsedSpace(0)
	val2 := val
	require.NoError(t, val2.ReadFromV2(m))
	require.Zero(t, val2.Epoch())
	require.Zero(t, val2.Value())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(announcement *v2container.UsedSpaceAnnouncement)
		}{
			{name: "container/missing", err: "missing container",
				corrupt: func(m *v2container.UsedSpaceAnnouncement) { m.SetContainerID(nil) }},
			{name: "container/nil", err: "invalid container: invalid length 0",
				corrupt: func(m *v2container.UsedSpaceAnnouncement) { m.SetContainerID(protoIDFromBytes(nil)) }},
			{name: "container/empty", err: "invalid container: invalid length 0",
				corrupt: func(m *v2container.UsedSpaceAnnouncement) { m.SetContainerID(protoIDFromBytes([]byte{})) }},
			{name: "container/undersize", err: "invalid container: invalid length 31",
				corrupt: func(m *v2container.UsedSpaceAnnouncement) { m.SetContainerID(protoIDFromBytes(anyValidID[:31])) }},
			{name: "container/oversize", err: "invalid container: invalid length 33",
				corrupt: func(m *v2container.UsedSpaceAnnouncement) {
					m.SetContainerID(protoIDFromBytes(append(anyValidID[:], 1)))
				}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				val2 := val
				var m v2container.UsedSpaceAnnouncement
				val2.WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(container.SizeEstimation).ReadFromV2(m), tc.err)
			})
		}
		t.Run("container/zero", func(t *testing.T) {
			var m v2container.UsedSpaceAnnouncement
			m.SetContainerID(protoIDFromBytes(make([]byte, cid.Size)))
			require.ErrorIs(t, val2.ReadFromV2(m), cid.ErrZero)
		})
	})
}

func TestSizeEstimation_WriteToV2(t *testing.T) {
	var val container.SizeEstimation
	var m v2container.UsedSpaceAnnouncement

	// zero
	val.WriteToV2(&m)
	require.Zero(t, val.Epoch())
	require.Zero(t, val.Container())
	require.Zero(t, val.Value())

	// filled
	validSizeEstimation.WriteToV2(&m)
	require.EqualValues(t, anyValidEpoch, m.GetEpoch())
	require.Equal(t, anyValidID[:], m.GetContainerID().GetValue())
	require.EqualValues(t, anyValidVolume, m.GetUsedSpace())
}
