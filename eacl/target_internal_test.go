package eacl

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util/slice"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

func TestNewTarget(t *testing.T) {
	checkPanicWithRole := func(role Role) {
		t.Run(fmt.Sprintf("invalid role=%d", role), func(t *testing.T) {
			require.Panics(t, func() { NewTarget([]Role{role}, nil) })
		})
	}
	checkPanicWithRole(0)
	checkPanicWithRole(lastRole)
	checkPanicWithRole(lastRole + 1)
}

func TestNewTargetWithRole(t *testing.T) {
	checkPanicWithRole := func(role Role) {
		t.Run(fmt.Sprintf("invalid role=%d", role), func(t *testing.T) {
			require.Panics(t, func() { NewTargetWithRole(role) })
		})
	}
	checkPanicWithRole(0)
	checkPanicWithRole(lastRole)
	checkPanicWithRole(lastRole + 1)
}

func TestTargetDeepCopy(t *testing.T) {
	src := NewTarget([]Role{
		RoleContainerOwner,
		RoleOthers,
	}, []neofscrypto.PublicKey{
		test.RandomPublicKey(),
		test.RandomPublicKey(),
	})

	srcKeys := make([][]byte, len(src.keys))
	for i := range src.keys {
		srcKeys[i] = slice.Copy(src.keys[i])
	}

	srcRoles := make([]Role, len(src.roles))
	copy(srcRoles, src.roles)

	var dst Target
	src.copyTo(&dst)

	require.Equal(t, src, dst)

	changeAllTargetFields(&dst)

	require.Equal(t, srcRoles, src.roles)
	require.Equal(t, srcKeys, src.keys)

	t.Run("full-to-full", func(t *testing.T) {
		dst = NewTarget([]Role{RoleContainerOwner}, []neofscrypto.PublicKey{test.RandomPublicKey()})
		src.copyTo(&dst)
		require.Equal(t, src, dst)
	})

	t.Run("zero-to-full", func(t *testing.T) {
		var zero Target
		zero.copyTo(&src)
		require.Zero(t, src)
	})
}
