package containertest_test

import (
	"testing"

	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	"github.com/stretchr/testify/require"
)

func TestContainer(t *testing.T) {
	v := containertest.Container()
	require.NotEqual(t, v, containertest.Container())

	var v2 container.Container
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apicontainer.Container
	v.WriteToV2(&m)
	var v3 container.Container
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v4 container.Container
	require.NoError(t, v4.UnmarshalJSON(j))
	require.Equal(t, v, v4)
}

func TestBasicACL(t *testing.T) {
	require.NotEqual(t, containertest.BasicACL(), containertest.BasicACL())
}

func TestSizeEstimation(t *testing.T) {
	v := containertest.SizeEstimation()
	require.NotEqual(t, v, containertest.SizeEstimation())

	var m apicontainer.AnnounceUsedSpaceRequest_Body_Announcement
	v.WriteToV2(&m)
	var v2 container.SizeEstimation
	require.NoError(t, v2.ReadFromV2(&m))
	require.Equal(t, v, v2)
}
