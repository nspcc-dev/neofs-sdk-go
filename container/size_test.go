package container_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	protocontainer "github.com/nspcc-dev/neofs-sdk-go/proto/container"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
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

func TestSizeEstimation_FromProtoMessage(t *testing.T) {
	m := &protocontainer.AnnounceUsedSpaceRequest_Body_Announcement{
		Epoch:       anyValidEpoch,
		ContainerId: &refs.ContainerID{Value: anyValidID[:]},
		UsedSpace:   anyValidVolume,
	}

	var val container.SizeEstimation
	require.NoError(t, val.FromProtoMessage(m))
	require.EqualValues(t, anyValidEpoch, val.Epoch())
	require.Equal(t, anyValidID, val.Container())
	require.EqualValues(t, anyValidVolume, val.Value())

	// reset optional fields
	m.Epoch = 0
	m.UsedSpace = 0
	val2 := val
	require.NoError(t, val2.FromProtoMessage(m))
	require.Zero(t, val2.Epoch())
	require.Zero(t, val2.Value())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(announcement *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement)
		}{
			{name: "container/missing", err: "missing container",
				corrupt: func(m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) { m.ContainerId = nil }},
			{name: "container/nil", err: "invalid container: invalid length 0",
				corrupt: func(m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) { m.ContainerId.Value = nil }},
			{name: "container/empty", err: "invalid container: invalid length 0",
				corrupt: func(m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) { m.ContainerId.Value = []byte{} }},
			{name: "container/undersize", err: "invalid container: invalid length 31",
				corrupt: func(m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) {
					m.ContainerId.Value = anyValidID[:31]
				}},
			{name: "container/oversize", err: "invalid container: invalid length 33",
				corrupt: func(m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) {
					m.ContainerId.Value = append(anyValidID[:], 1)
				}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				val2 := val
				m := val2.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(container.SizeEstimation).FromProtoMessage(m), tc.err)
			})
		}
		t.Run("container/zero", func(t *testing.T) {
			m := &protocontainer.AnnounceUsedSpaceRequest_Body_Announcement{
				ContainerId: &refs.ContainerID{Value: make([]byte, cid.Size)},
			}
			require.ErrorIs(t, val2.FromProtoMessage(m), cid.ErrZero)
		})
	})
}

func TestSizeEstimation_ProtoMessage(t *testing.T) {
	var val container.SizeEstimation

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetEpoch())
	require.Zero(t, m.GetContainerId())
	require.Zero(t, m.GetUsedSpace())

	// filled
	m = validSizeEstimation.ProtoMessage()
	require.EqualValues(t, anyValidEpoch, m.GetEpoch())
	require.Equal(t, anyValidID[:], m.GetContainerId().GetValue())
	require.EqualValues(t, anyValidVolume, m.GetUsedSpace())
}
