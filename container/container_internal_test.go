package container

import (
	"bytes"
	"math/rand/v2"
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
		copy(dst.owner[:], newOwner[:])

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

		require.True(t, container.nonce == dst.nonce)
		copy(dst.nonce[:], []byte{1, 2, 3})
		require.False(t, container.nonce == dst.nonce)
	})

	t.Run("overwrite nonce", func(t *testing.T) {
		var local Container
		require.Zero(t, local.nonce)

		var dst Container
		copy(dst.nonce[:], []byte{1, 2, 3})
		require.NotZero(t, dst.nonce)

		local.CopyTo(&dst)
		require.True(t, bytes.Equal(local.Marshal(), dst.Marshal()))

		require.Zero(t, local.nonce)
		require.Zero(t, dst.nonce)

		require.True(t, local.nonce == dst.nonce)
		copy(dst.nonce[:], []byte{1, 2, 3})
		require.False(t, local.nonce == dst.nonce)
	})

	t.Run("change version", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.Equal(t, container.Version(), dst.Version())

		cp := container.Version()
		dst.version.SetMajor(rand.Uint32())
		dst.version.SetMinor(rand.Uint32())

		require.NotEqual(t, cp, dst.Version())
		require.Equal(t, cp, container.Version())
	})

	t.Run("change attributes", func(t *testing.T) {
		var dst Container
		container.CopyTo(&dst)

		require.Equal(t, container.Attribute(attrOne), dst.Attribute(attrOne))
		dst.SetAttribute(attrOne, "value")
		require.NotEqual(t, container.Attribute(attrOne), dst.Attribute(attrOne))
	})
}
