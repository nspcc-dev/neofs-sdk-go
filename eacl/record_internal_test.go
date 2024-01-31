package eacl

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util/slice"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

func TestNewRecord(t *testing.T) {
	const anyValidAction = ActionDeny
	const anyValidOp = acl.OpObjectHash
	anyValidTarget := NewTargetWithRole(RoleContainerOwner)
	anyValidFilters := []Filter{
		NewFilter(HeaderFromObject, FilterObjectType, MatchStringEqual, "UNKNOWN"),
		NewFilter(HeaderFromRequest, "ContentType", MatchStringNotEqual, "img/png"),
	}

	for i, action := range []Action{
		0,
		lastAction,
		lastAction + 1,
	} {
		t.Run(fmt.Sprintf("invalid action=%d", action), func(t *testing.T) {
			require.Panics(t, func() { NewRecord(action, anyValidOp, anyValidTarget, anyValidFilters...) }, i)
		})
	}

	for i, op := range []acl.Op{
		0,
		lastOp,
		lastOp + 1,
	} {
		t.Run(fmt.Sprintf("invalid op=%d", op), func(t *testing.T) {
			require.Panics(t, func() { NewRecord(anyValidAction, op, anyValidTarget, anyValidFilters...) }, i)
		})
	}
}

func TestRecord_IsForRole(t *testing.T) {
	const anyValidAction = ActionDeny
	const anyValidOp = acl.OpObjectHash
	anyValidTarget := NewTargetWithRole(RoleContainerOwner)
	anyValidFilters := []Filter{
		NewFilter(HeaderFromObject, FilterObjectType, MatchStringEqual, "UNKNOWN"),
		NewFilter(HeaderFromRequest, "ContentType", MatchStringNotEqual, "img/png"),
	}
	anyRecord := NewRecord(anyValidAction, anyValidOp, anyValidTarget, anyValidFilters...)

	for _, role := range []Role{
		0,
		lastRole,
		lastRole + 1,
	} {
		t.Run(fmt.Sprintf("invalid role=%d", role), func(t *testing.T) {
			require.Panics(t, func() { anyRecord.IsForRole(role) })
		})
	}
}

func TestRecordDeepCopy(t *testing.T) {
	const srcAction = ActionAllow
	const srcOp = acl.OpObjectDelete

	srcTarget := NewTarget([]Role{
		RoleContainerOwner,
		RoleOthers,
	}, []neofscrypto.PublicKey{
		test.RandomPublicKey(),
		test.RandomPublicKey(),
	})

	src := NewRecord(srcAction, srcOp, srcTarget,
		NewFilter(HeaderFromObject, FilterObjectType, MatchStringEqual, "UNKNOWN"),
		NewFilter(HeaderFromRequest, "ContentType", MatchStringNotEqual, "img/png"),
	)

	srcTarget = Target{
		roles: make([]Role, len(src.target.roles)),
		keys:  make([][]byte, len(src.target.keys)),
	}

	copy(srcTarget.roles, src.target.roles)
	for i := range src.target.keys {
		srcTarget.keys[i] = slice.Copy(src.target.keys[i])
	}

	srcFilters := make([]Filter, len(src.filters))
	copy(srcFilters, src.filters)

	var dst Record
	src.copyTo(&dst)

	require.Equal(t, src, dst)

	changeAllRecordFields(&dst)

	require.Equal(t, srcAction, src.action)
	require.Equal(t, srcOp, src.operation)
	require.Equal(t, srcTarget, src.target)
	require.Equal(t, srcFilters, src.filters)

	t.Run("full-to-full", func(t *testing.T) {
		target := NewTarget([]Role{RoleContainerOwner}, []neofscrypto.PublicKey{test.RandomPublicKey()})
		dst = NewRecord(ActionDeny, acl.OpObjectPut, target,
			NewFilter(HeaderFromService, "key", MatchStringEqual, "val"),
		)

		src.copyTo(&dst)
		require.Equal(t, src, dst)
	})

	t.Run("zero-to-full", func(t *testing.T) {
		var zero Record
		zero.copyTo(&src)
		require.Zero(t, src)
	})
}
