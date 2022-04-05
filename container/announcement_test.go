package container_test

import (
	"crypto/sha256"
	"testing"

	containerv2 "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	"github.com/stretchr/testify/require"
)

func TestAnnouncement(t *testing.T) {
	const epoch, usedSpace uint64 = 10, 100

	cidValue := [sha256.Size]byte{1, 2, 3}
	id := cidtest.IDWithChecksum(cidValue)

	var a container.UsedSpaceAnnouncement
	a.SetEpoch(epoch)
	a.SetContainerID(&id)
	a.SetUsedSpace(usedSpace)

	require.Equal(t, epoch, a.Epoch())
	require.Equal(t, usedSpace, a.UsedSpace())
	require.Equal(t, &id, a.ContainerID())

	t.Run("test v2", func(t *testing.T) {
		const newEpoch, newUsedSpace uint64 = 20, 200

		newCidValue := [32]byte{4, 5, 6}
		newCID := new(refs.ContainerID)
		newCID.SetValue(newCidValue[:])

		var v2 containerv2.UsedSpaceAnnouncement
		a.WriteToV2(&v2)

		require.Equal(t, usedSpace, v2.GetUsedSpace())
		require.Equal(t, epoch, v2.GetEpoch())
		require.Equal(t, cidValue[:], v2.GetContainerID().GetValue())

		v2.SetEpoch(newEpoch)
		v2.SetUsedSpace(newUsedSpace)
		v2.SetContainerID(newCID)

		var newA container.UsedSpaceAnnouncement
		newA.ReadFromV2(v2)

		var cID cid.ID
		cID.ReadFromV2(*newCID)

		require.Equal(t, newEpoch, newA.Epoch())
		require.Equal(t, newUsedSpace, newA.UsedSpace())
		require.Equal(t, &cID, newA.ContainerID())
	})
}

func TestUsedSpaceEncoding(t *testing.T) {
	a := containertest.UsedSpaceAnnouncement()

	t.Run("binary", func(t *testing.T) {
		data, err := a.Marshal()
		require.NoError(t, err)

		var a2 container.UsedSpaceAnnouncement
		require.NoError(t, a2.Unmarshal(data))

		require.Equal(t, a, a2)
	})
}

func TestUsedSpaceAnnouncement_ToV2(t *testing.T) {
	t.Run("to zero V2", func(t *testing.T) {
		var (
			x  container.UsedSpaceAnnouncement
			v2 containerv2.UsedSpaceAnnouncement
		)

		x.WriteToV2(&v2)

		require.Nil(t, v2.GetContainerID())
		require.Zero(t, v2.GetUsedSpace())
		require.Zero(t, v2.GetEpoch())
	})

	t.Run("default values", func(t *testing.T) {
		var announcement container.UsedSpaceAnnouncement

		// check initial values
		require.Zero(t, announcement.Epoch())
		require.Zero(t, announcement.UsedSpace())
		require.Nil(t, announcement.ContainerID())
	})
}

func TestNewAnnouncementFromV2(t *testing.T) {
	t.Run("from zero V2", func(t *testing.T) {
		var (
			x  container.UsedSpaceAnnouncement
			v2 containerv2.UsedSpaceAnnouncement
		)

		x.ReadFromV2(v2)

		require.Nil(t, x.ContainerID())
		require.Zero(t, x.UsedSpace())
		require.Zero(t, x.Epoch())
	})
}
