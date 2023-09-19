package eacl_test

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

func TestNewRecord(t *testing.T) {
	const anyValidAction = eacl.ActionDeny
	const anyValidOp = acl.OpObjectHash
	allValidRoles := []eacl.Role{
		eacl.RoleContainerOwner,
		eacl.RoleOthers,
	}
	anyValidKeys := []neofscrypto.PublicKey{
		test.RandomPublicKey(),
		test.RandomPublicKey(),
	}
	validTarget := eacl.NewTarget(allValidRoles, anyValidKeys)
	validFilters := []eacl.Filter{
		eacl.NewFilter(eacl.HeaderFromObject, eacl.FilterObjectType, eacl.MatchStringEqual, "UNKNOWN"),
		eacl.NewFilter(eacl.HeaderFromRequest, "ContentType", eacl.MatchStringNotEqual, "img/png"),
	}

	t.Run("uninitialized target", func(t *testing.T) {
		require.Panics(t, func() {
			var zeroTarget eacl.Target
			eacl.NewRecord(anyValidAction, anyValidOp, zeroTarget, validFilters...)
		})
	})

	for i, action := range []eacl.Action{
		eacl.ActionAllow,
		eacl.ActionDeny,
	} {
		t.Run(fmt.Sprintf("support action=%v", action), func(t *testing.T) {
			r := eacl.NewRecord(action, anyValidOp, validTarget, validFilters...)
			require.Equal(t, action, r.Action(), i)
		})
	}

	for i, op := range []acl.Op{
		acl.OpObjectGet,
		acl.OpObjectHead,
		acl.OpObjectPut,
		acl.OpObjectDelete,
		acl.OpObjectSearch,
		acl.OpObjectRange,
		acl.OpObjectHash,
	} {
		t.Run(fmt.Sprintf("support op=%v", op), func(t *testing.T) {
			r := eacl.NewRecord(anyValidAction, op, validTarget, validFilters...)
			require.Equal(t, op, r.Op())
			require.True(t, r.IsForOp(op), i)
		})
	}

	r := eacl.NewRecord(anyValidAction, anyValidOp, validTarget, validFilters...)
	require.Equal(t, anyValidAction, r.Action())
	require.Equal(t, validFilters, r.Filters())
	require.Equal(t, anyValidOp, r.Op())
	require.True(t, r.IsForOp(anyValidOp))

	for i := range allValidRoles {
		require.True(t, r.IsForRole(allValidRoles[i]))
	}

	bKeys := r.TargetBinaryKeys()
	require.Len(t, bKeys, len(anyValidKeys))

	for i := range anyValidKeys {
		require.True(t, r.IsForKey(anyValidKeys[i]))
		require.Contains(t, bKeys, neofscrypto.PublicKeyBytes(anyValidKeys[i]))
	}
}

func TestRecordTargetRole(t *testing.T) {
	const anyValidAction = eacl.ActionDeny
	const anyValidOp = acl.OpObjectHash
	anyValidFilters := []eacl.Filter{
		eacl.NewFilter(eacl.HeaderFromObject, eacl.FilterObjectType, eacl.MatchStringEqual, "UNKNOWN"),
		eacl.NewFilter(eacl.HeaderFromRequest, "ContentType", eacl.MatchStringNotEqual, "img/png"),
	}

	supportedRoles := []eacl.Role{
		eacl.RoleContainerOwner,
		eacl.RoleSystem,
		eacl.RoleOthers,
	}

	const targetRole = eacl.RoleContainerOwner
	r := eacl.NewRecord(anyValidAction, anyValidOp, eacl.NewTargetWithRole(targetRole), anyValidFilters...)

	for _, role := range supportedRoles {
		t.Run(fmt.Sprintf("support role=%s", role), func(t *testing.T) {
			require.NotPanics(t, func() { r.IsForRole(role) })
			if role == targetRole {
				require.True(t, r.IsForRole(role))
			} else {
				require.False(t, r.IsForRole(role))
			}
		})
	}

	t.Run("keys only", func(t *testing.T) {
		r = eacl.NewRecord(anyValidAction, anyValidOp, eacl.NewTargetWithKey(test.RandomPublicKey()), anyValidFilters...)
		for i := range supportedRoles {
			require.False(t, r.IsForRole(supportedRoles[i]), i)
		}
	})
}

func TestRecordTargetKeys(t *testing.T) {
	const anyValidAction = eacl.ActionDeny
	const anyValidOp = acl.OpObjectHash
	anyValidFilters := []eacl.Filter{
		eacl.NewFilter(eacl.HeaderFromObject, eacl.FilterObjectType, eacl.MatchStringEqual, "UNKNOWN"),
		eacl.NewFilter(eacl.HeaderFromRequest, "ContentType", eacl.MatchStringNotEqual, "img/png"),
	}

	targetKeys := []neofscrypto.PublicKey{
		test.RandomPublicKey(),
		test.RandomPublicKey(),
	}
	otherKey := test.RandomPublicKey()

	r := eacl.NewRecord(anyValidAction, anyValidOp, eacl.NewTargetWithKeys(targetKeys), anyValidFilters...)

	bKeys := r.TargetBinaryKeys()
	require.Len(t, bKeys, len(targetKeys))

	for i := range targetKeys {
		require.True(t, r.IsForKey(targetKeys[i]))
		require.Contains(t, bKeys, neofscrypto.PublicKeyBytes(targetKeys[i]))
	}

	require.False(t, r.IsForKey(otherKey))
	require.NotContains(t, bKeys, neofscrypto.PublicKeyBytes(otherKey))
}
