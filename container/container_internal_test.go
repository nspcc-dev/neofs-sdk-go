package container

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestContainer_CopyTo(t *testing.T) {
	owner := usertest.ID()

	var container Container
	container.Init()

	attrOne := "a0"
	attrValue := "a1"

	container.SetOwner(owner)
	container.SetBasicACL(acl.PublicRWExtended)
	container.SetAttribute(attrOne, attrValue)

	var pp netmap.PlacementPolicy
	pp.SetContainerBackupFactor(123)

	var rd netmap.ReplicaDescriptor
	rd.SetSelectorName("selector")
	rd.SetNumberOfObjects(100)
	pp.SetReplicas([]netmap.ReplicaDescriptor{rd})

	var f netmap.Filter
	f.SetName("filter")
	pp.SetFilters([]netmap.Filter{f})

	var s netmap.Selector
	s.SetName("selector")
	pp.SetSelectors([]netmap.Selector{s})

	container.SetPlacementPolicy(pp)

	t.Run("copy", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.Equal(t, container, dst)
		require.True(t, bytes.Equal(container.Marshal(), dst.Marshal()))
	})

	t.Run("change acl", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.Equal(t, container.BasicACL(), dst.BasicACL())
		dst.SetBasicACL(acl.Private)
		require.NotEqual(t, container.BasicACL(), dst.BasicACL())
	})

	t.Run("change owner", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.True(t, container.Owner() == dst.Owner())

		newOwner := usertest.ID()
		dst.v2.GetOwnerID().SetValue(newOwner[:])

		require.False(t, container.Owner() == dst.Owner())
	})

	t.Run("replace owner", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.True(t, container.Owner() == dst.Owner())

		newOwner := usertest.ID()
		dst.SetOwner(newOwner)

		require.False(t, container.Owner() == dst.Owner())
	})

	t.Run("change nonce", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.True(t, bytes.Equal(container.v2.GetNonce(), dst.v2.GetNonce()))
		dst.v2.SetNonce([]byte{1, 2, 3})
		require.False(t, bytes.Equal(container.v2.GetNonce(), dst.v2.GetNonce()))
	})

	t.Run("overwrite nonce", func(t *testing.T) {
		var local Container
		require.Empty(t, local.v2.GetNonce())

		var dst Container
		dst.v2.SetNonce([]byte{1, 2, 3})
		require.NotEmpty(t, dst.v2.GetNonce())

		local.CopyTo(&dst)
		require.True(t, bytes.Equal(local.Marshal(), dst.Marshal()))

		require.Empty(t, local.v2.GetNonce())
		require.Empty(t, dst.v2.GetNonce())

		require.True(t, bytes.Equal(local.v2.GetNonce(), dst.v2.GetNonce()))
		dst.v2.SetNonce([]byte{1, 2, 3})
		require.False(t, bytes.Equal(local.v2.GetNonce(), dst.v2.GetNonce()))
	})

	t.Run("change version", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		oldVer := container.v2.GetVersion()
		require.NotNil(t, oldVer)

		newVer := dst.v2.GetVersion()
		require.NotNil(t, newVer)

		require.Equal(t, oldVer.GetMajor(), newVer.GetMajor())
		require.Equal(t, oldVer.GetMinor(), newVer.GetMinor())

		newVer.SetMajor(rand.Uint32())
		newVer.SetMinor(rand.Uint32())

		require.NotEqual(t, oldVer.GetMajor(), newVer.GetMajor())
		require.NotEqual(t, oldVer.GetMinor(), newVer.GetMinor())
	})

	t.Run("change attributes", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.Equal(t, container.Attribute(attrOne), dst.Attribute(attrOne))
		dst.SetAttribute(attrOne, "value")
		require.NotEqual(t, container.Attribute(attrOne), dst.Attribute(attrOne))
	})
}
