package container_test

import (
	"testing"

	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	"github.com/stretchr/testify/require"
)

func TestSizeEstimation_ReadFromV2(t *testing.T) {
	t.Run("missing fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			s := containertest.SizeEstimation()
			var m apicontainer.AnnounceUsedSpaceRequest_Body_Announcement

			s.WriteToV2(&m)
			m.ContainerId = nil
			require.ErrorContains(t, s.ReadFromV2(&m), "missing container")
		})
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("container", func(t *testing.T) {
			t.Run("container", func(t *testing.T) {
				s := containertest.SizeEstimation()
				var m apicontainer.AnnounceUsedSpaceRequest_Body_Announcement

				s.WriteToV2(&m)
				m.ContainerId.Value = nil
				require.ErrorContains(t, s.ReadFromV2(&m), "invalid container: missing value field")
				m.ContainerId.Value = []byte{}
				require.ErrorContains(t, s.ReadFromV2(&m), "invalid container: missing value field")
				m.ContainerId.Value = make([]byte, 31)
				require.ErrorContains(t, s.ReadFromV2(&m), "invalid container: invalid value length 31")
				m.ContainerId.Value = make([]byte, 33)
				require.ErrorContains(t, s.ReadFromV2(&m), "invalid container: invalid value length 33")
			})
		})
	})
}

func testSizeEstimationNumField(t *testing.T, get func(container.SizeEstimation) uint64, set func(*container.SizeEstimation, uint64),
	getAPI func(*apicontainer.AnnounceUsedSpaceRequest_Body_Announcement) uint64) {
	var s container.SizeEstimation

	require.Zero(t, get(s))

	const val = 13
	set(&s, val)
	require.EqualValues(t, val, get(s))

	const valOther = 42
	set(&s, valOther)
	require.EqualValues(t, valOther, get(s))

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst container.SizeEstimation
			var msg apicontainer.AnnounceUsedSpaceRequest_Body_Announcement

			// set required data just to satisfy decoder
			src.SetContainer(cidtest.ID())

			set(&dst, val)

			src.WriteToV2(&msg)
			require.Zero(t, getAPI(&msg))
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, get(dst))

			set(&src, val)

			src.WriteToV2(&msg)
			require.EqualValues(t, val, getAPI(&msg))
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
	})
}

func TestSizeEstimation_SetEpoch(t *testing.T) {
	testSizeEstimationNumField(t, container.SizeEstimation.Epoch, (*container.SizeEstimation).SetEpoch,
		(*apicontainer.AnnounceUsedSpaceRequest_Body_Announcement).GetEpoch)
}

func TestSizeEstimation_SetValue(t *testing.T) {
	testSizeEstimationNumField(t, container.SizeEstimation.Value, (*container.SizeEstimation).SetValue,
		(*apicontainer.AnnounceUsedSpaceRequest_Body_Announcement).GetUsedSpace)
}

func TestSizeEstimation_SetContainer(t *testing.T) {
	var s container.SizeEstimation

	require.Zero(t, s.Container())

	cnr := cidtest.ID()

	s.SetContainer(cnr)
	require.Equal(t, cnr, s.Container())

	cnrOther := cidtest.ChangeID(cnr)
	s.SetContainer(cnrOther)
	require.Equal(t, cnrOther, s.Container())

	t.Run("encoding", func(t *testing.T) {
		t.Run("api", func(t *testing.T) {
			var src, dst container.SizeEstimation
			var msg apicontainer.AnnounceUsedSpaceRequest_Body_Announcement

			dst.SetContainer(cnr)

			src.WriteToV2(&msg)
			require.Equal(t, make([]byte, 32), msg.ContainerId.Value)
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Zero(t, dst.Container())

			dst.SetContainer(cnrOther)
			src.SetContainer(cnr)
			src.WriteToV2(&msg)
			require.Equal(t, cnr[:], msg.ContainerId.Value)
			err = dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.Equal(t, cnr, dst.Container())
		})
	})
}
