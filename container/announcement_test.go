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

	a := container.NewAnnouncement()
	a.SetEpoch(epoch)
	a.SetContainerID(id)
	a.SetUsedSpace(usedSpace)

	require.Equal(t, epoch, a.Epoch())
	require.Equal(t, usedSpace, a.UsedSpace())
	require.Equal(t, id, a.ContainerID())

	t.Run("test v2", func(t *testing.T) {
		const newEpoch, newUsedSpace uint64 = 20, 200

		newCidValue := [32]byte{4, 5, 6}
		newCID := new(refs.ContainerID)
		newCID.SetValue(newCidValue[:])

		v2 := a.ToV2()
		require.Equal(t, usedSpace, v2.GetUsedSpace())
		require.Equal(t, epoch, v2.GetEpoch())
		require.Equal(t, cidValue[:], v2.GetContainerID().GetValue())

		v2.SetEpoch(newEpoch)
		v2.SetUsedSpace(newUsedSpace)
		v2.SetContainerID(newCID)

		newA := container.NewAnnouncementFromV2(v2)

		require.Equal(t, newEpoch, newA.Epoch())
		require.Equal(t, newUsedSpace, newA.UsedSpace())
		require.Equal(t, cid.NewFromV2(newCID), newA.ContainerID())
	})
}

func TestUsedSpaceEncoding(t *testing.T) {
	a := containertest.UsedSpaceAnnouncement()

	t.Run("binary", func(t *testing.T) {
		data := a.Marshal()

		a2 := container.NewAnnouncement()
		require.NoError(t, a2.Unmarshal(data))

		require.Equal(t, a, a2)
	})
}

func TestUsedSpaceAnnouncement_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *container.UsedSpaceAnnouncement

		require.Nil(t, x.ToV2())
	})

	t.Run("default values", func(t *testing.T) {
		announcement := container.NewAnnouncement()

		// check initial values
		require.Zero(t, announcement.Epoch())
		require.Zero(t, announcement.UsedSpace())
		require.Nil(t, announcement.ContainerID())

		// convert to v2 message
		announcementV2 := announcement.ToV2()

		require.Zero(t, announcementV2.GetEpoch())
		require.Zero(t, announcementV2.GetUsedSpace())
		require.Nil(t, announcementV2.GetContainerID())
	})
}

func TestNewAnnouncementFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *containerv2.UsedSpaceAnnouncement

		require.Nil(t, container.NewAnnouncementFromV2(x))
	})
}
