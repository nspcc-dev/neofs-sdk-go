package session

import (
	"bytes"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestObject_CopyTo(t *testing.T) {
	var container Object

	containerID := cidtest.ID()

	container.ForVerb(VerbObjectDelete)
	container.BindContainer(containerID)
	container.LimitByObjects(oidtest.IDs(2)...)

	t.Run("copy", func(t *testing.T) {
		var dst Object
		container.CopyTo(&dst)

		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}

		require.Equal(t, container, dst)
		require.True(t, bytes.Equal(container.marshal(emptyWriter), dst.marshal(emptyWriter)))
	})

	t.Run("change simple fields", func(t *testing.T) {
		var dst Object
		container.CopyTo(&dst)

		require.Equal(t, container.verb, dst.verb)
		require.NotZero(t, container.cnr)
		require.NotZero(t, dst.cnr)

		container.ForVerb(VerbObjectHead)

		require.NotEqual(t, container.verb, dst.verb)
		require.NotZero(t, container.cnr)
		require.NotZero(t, dst.cnr)
	})

	t.Run("change ids", func(t *testing.T) {
		var dst Object
		container.CopyTo(&dst)

		for i := range container.objs {
			require.True(t, container.objs[i] == dst.objs[i])

			// change object id in the new object
			for j := range dst.objs[i] {
				dst.objs[i][j] = byte(j)
			}
		}

		for i := range container.objs {
			require.False(t, container.objs[i] == dst.objs[i])
		}
	})

	t.Run("overwrite container id", func(t *testing.T) {
		var local Object
		require.Zero(t, local.cnr)

		var dst Object
		dst.BindContainer(containerID)
		require.NotZero(t, dst.cnr)

		local.CopyTo(&dst)
		emptyWriter := func() session.TokenContext {
			return &session.ContainerSessionContext{}
		}

		require.Equal(t, local, dst)
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.Zero(t, local.cnr)
		require.Zero(t, dst.cnr)

		dst.BindContainer(containerID)
		require.NotZero(t, dst.cnr)
		require.Zero(t, local.cnr)
	})
}
