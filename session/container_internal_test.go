package session

import (
	"bytes"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/stretchr/testify/require"
)

func TestContainer_CopyTo(t *testing.T) {
	var container Container

	containerID := cidtest.ID()

	container.ForVerb(VerbContainerDelete)
	container.ApplyOnlyTo(containerID)

	t.Run("copy", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		emptyWriter := func(body *protosession.SessionToken_Body) {
			body.Context = &protosession.SessionToken_Body_Container{}
		}

		require.Equal(t, container, dst)
		require.True(t, bytes.Equal(container.marshal(emptyWriter), dst.marshal(emptyWriter)))
	})

	t.Run("change", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.Equal(t, container.verb, dst.verb)
		require.NotZero(t, container.cnr)
		require.NotZero(t, dst.cnr)

		container.ForVerb(VerbContainerSetEACL)

		require.NotEqual(t, container.verb, dst.verb)
		require.NotZero(t, container.cnr)
		require.NotZero(t, dst.cnr)
	})

	t.Run("overwrite container id", func(t *testing.T) {
		var local Container
		require.Zero(t, local.cnr)

		var dst Container
		dst.ApplyOnlyTo(containerID)
		require.NotZero(t, dst.cnr)

		local.CopyTo(&dst)
		emptyWriter := func(body *protosession.SessionToken_Body) {
			body.Context = &protosession.SessionToken_Body_Container{}
		}

		require.Equal(t, local, dst)
		require.True(t, bytes.Equal(local.marshal(emptyWriter), dst.marshal(emptyWriter)))

		require.Zero(t, local.cnr)
		require.Zero(t, dst.cnr)

		dst.ApplyOnlyTo(containerID)
		require.NotZero(t, dst.cnr)
		require.Zero(t, local.cnr)
	})
}
