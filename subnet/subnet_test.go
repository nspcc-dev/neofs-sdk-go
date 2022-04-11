package subnet_test

import (
	"testing"

	subnetv2 "github.com/nspcc-dev/neofs-api-go/v2/subnet"
	subnettest "github.com/nspcc-dev/neofs-api-go/v2/subnet/test"
	. "github.com/nspcc-dev/neofs-sdk-go/subnet"
	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestInfoZero(t *testing.T) {
	var info Info

	var id subnetid.ID
	info.ReadID(&id)

	require.True(t, subnetid.IsZero(id))
}

func TestInfo_SetID(t *testing.T) {
	var (
		id   subnetid.ID
		info Info
	)

	id.SetNumber(222)

	info.SetID(id)

	require.True(t, IDEquals(info, id))
}

func TestInfo_SetOwner(t *testing.T) {
	var (
		id   user.ID
		info Info
	)

	id = *usertest.ID()

	require.False(t, IsOwner(info, id))

	info.SetOwner(id)

	require.True(t, IsOwner(info, id))
}

func TestInfo_WriteToV2(t *testing.T) {
	var (
		infoTo, infoFrom Info

		infoV2From, infoV2To subnetv2.Info
	)

	infoV2From = *subnettest.GenerateSubnetInfo(false)

	infoFrom.FromV2(infoV2From)

	infoFrom.WriteToV2(&infoV2To)

	infoTo.FromV2(infoV2To)

	require.Equal(t, infoV2From, infoV2To)
	require.Equal(t, infoFrom, infoTo)
}
