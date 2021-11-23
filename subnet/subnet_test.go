package subnet_test

import (
	"testing"

	subnetv2 "github.com/nspcc-dev/neofs-api-go/v2/subnet"
	subnettest "github.com/nspcc-dev/neofs-api-go/v2/subnet/test"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	"github.com/nspcc-dev/neofs-sdk-go/subnet"
	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
	"github.com/stretchr/testify/require"
)

func TestInfoZero(t *testing.T) {
	var info subnet.Info

	var id subnetid.ID
	info.ReadID(&id)

	require.True(t, subnetid.IsZero(id))

	require.False(t, info.HasOwner())
}

func TestInfo_SetID(t *testing.T) {
	var (
		idFrom, idTo subnetid.ID

		info subnet.Info
	)

	idFrom.SetNumber(222)

	info.SetID(idFrom)

	info.ReadID(&idTo)

	require.True(t, idTo.Equals(&idFrom))
}

func TestInfo_SetOwner(t *testing.T) {
	var (
		idFrom, idTo owner.ID

		info subnet.Info
	)

	idFrom = *ownertest.GenerateID()

	require.False(t, info.HasOwner())

	info.SetOwner(idFrom)

	require.True(t, info.HasOwner())

	info.ReadOwner(&idTo)

	require.True(t, idTo.Equal(&idFrom))
}

func TestInfo_WriteToV2(t *testing.T) {
	var (
		infoTo, infoFrom subnet.Info

		infoV2From, infoV2To subnetv2.Info
	)

	infoV2From = *subnettest.GenerateSubnetInfo(false)

	infoFrom.FromV2(infoV2From)

	infoFrom.WriteToV2(&infoV2To)

	infoTo.FromV2(infoV2To)

	require.Equal(t, infoV2From, infoV2To)
	require.Equal(t, infoFrom, infoTo)
}
