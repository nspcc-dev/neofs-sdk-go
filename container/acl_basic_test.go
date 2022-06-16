package container

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicACL_DisableExtension(t *testing.T) {
	var val, val2 BasicACL

	require.True(t, val.Extendable())
	val2.fromUint32(val.toUint32())
	require.True(t, val2.Extendable())

	val.DisableExtension()

	require.False(t, val.Extendable())
	val2.fromUint32(val.toUint32())
	require.False(t, val2.Extendable())
}

func TestBasicACL_MakeSticky(t *testing.T) {
	var val, val2 BasicACL

	require.False(t, val.Sticky())
	val2.fromUint32(val.toUint32())
	require.False(t, val2.Sticky())

	val.MakeSticky()

	require.True(t, val.Sticky())
	val2.fromUint32(val.toUint32())
	require.True(t, val2.Sticky())
}

func TestBasicACL_AllowBearerRules(t *testing.T) {
	var val BasicACL

	require.Panics(t, func() { val.AllowBearerRules(aclOpZero) })
	require.Panics(t, func() { val.AllowBearerRules(aclOpLast) })

	require.Panics(t, func() { val.AllowedBearerRules(aclOpZero) })
	require.Panics(t, func() { val.AllowedBearerRules(aclOpLast) })

	for op := aclOpZero + 1; op < aclOpLast; op++ {
		val := val

		require.False(t, val.AllowedBearerRules(op))

		val.AllowBearerRules(op)

		for j := aclOpZero + 1; j < aclOpLast; j++ {
			if j == op {
				require.True(t, val.AllowedBearerRules(j), op)
			} else {
				require.False(t, val.AllowedBearerRules(j), op)
			}
		}
	}
}

func TestBasicACL_AllowOp(t *testing.T) {
	var val, val2 BasicACL

	require.Panics(t, func() { val.IsOpAllowed(aclOpZero, aclRoleZero+1) })
	require.Panics(t, func() { val.IsOpAllowed(aclOpLast, aclRoleZero+1) })
	require.Panics(t, func() { val.IsOpAllowed(aclOpZero+1, aclRoleZero) })
	require.Panics(t, func() { val.IsOpAllowed(aclOpZero+1, aclRoleLast) })

	for op := aclOpZero + 1; op < aclOpLast; op++ {
		require.Panics(t, func() { val.AllowOp(op, ACLRoleInnerRing) })

		if isReplicationOp(op) {
			require.Panics(t, func() { val.AllowOp(op, ACLRoleContainer) })
			require.True(t, val.IsOpAllowed(op, ACLRoleContainer))
		}
	}

	require.True(t, val.IsOpAllowed(ACLOpObjectGet, ACLRoleInnerRing))
	require.True(t, val.IsOpAllowed(ACLOpObjectHead, ACLRoleInnerRing))
	require.True(t, val.IsOpAllowed(ACLOpObjectSearch, ACLRoleInnerRing))
	require.True(t, val.IsOpAllowed(ACLOpObjectHash, ACLRoleInnerRing))

	const op = aclOpZero + 1
	const role = ACLRoleOthers

	require.False(t, val.IsOpAllowed(op, role))
	val2.fromUint32(val.toUint32())
	require.False(t, val2.IsOpAllowed(op, role))

	val.AllowOp(op, role)

	require.True(t, val.IsOpAllowed(op, role))
	val2.fromUint32(val.toUint32())
	require.True(t, val2.IsOpAllowed(op, role))
}

type opsExpected struct {
	owner, container, innerRing, others, bearer bool
}

func testOp(t *testing.T, v BasicACL, op ACLOp, exp opsExpected) {
	require.Equal(t, exp.owner, v.IsOpAllowed(op, ACLRoleOwner), op)
	require.Equal(t, exp.container, v.IsOpAllowed(op, ACLRoleContainer), op)
	require.Equal(t, exp.innerRing, v.IsOpAllowed(op, ACLRoleInnerRing), op)
	require.Equal(t, exp.others, v.IsOpAllowed(op, ACLRoleOthers), op)
	require.Equal(t, exp.bearer, v.AllowedBearerRules(op), op)
}

type expected struct {
	extendable, sticky bool

	mOps map[ACLOp]opsExpected
}

func testBasicACLPredefined(t *testing.T, val BasicACL, name string, exp expected) {
	require.Equal(t, exp.sticky, val.Sticky())
	require.Equal(t, exp.extendable, val.Extendable())

	for op, exp := range exp.mOps {
		testOp(t, val, op, exp)
	}

	s := val.EncodeToString()

	var val2 BasicACL

	require.NoError(t, val2.DecodeString(s))
	require.Equal(t, val, val2)

	require.NoError(t, val2.DecodeString(name))
	require.Equal(t, val, val2)
}

func TestBasicACLPredefined(t *testing.T) {
	t.Run("private", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[ACLOp]opsExpected{
				ACLOpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
			},
		}

		testBasicACLPredefined(t, BasicACLPrivate, BasicACLNamePrivate, exp)
		exp.extendable = true
		testBasicACLPredefined(t, BasicACLPrivateExtended, BasicACLNamePrivateExtended, exp)
	})

	t.Run("public-read", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[ACLOp]opsExpected{
				ACLOpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				ACLOpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
			},
		}

		testBasicACLPredefined(t, BasicACLPublicRO, BasicACLNamePublicRO, exp)
		exp.extendable = true
		testBasicACLPredefined(t, BasicACLPublicROExtended, BasicACLNamePublicROExtended, exp)
	})

	t.Run("public-read-write", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[ACLOp]opsExpected{
				ACLOpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
			},
		}

		testBasicACLPredefined(t, BasicACLPublicRW, BasicACLNamePublicRW, exp)
		exp.extendable = true
		testBasicACLPredefined(t, BasicACLPublicRWExtended, BasicACLNamePublicRWExtended, exp)
	})

	t.Run("public-append", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[ACLOp]opsExpected{
				ACLOpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    true,
				},
				ACLOpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				ACLOpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
			},
		}

		testBasicACLPredefined(t, BasicACLPublicAppend, BasicACLNamePublicAppend, exp)
		exp.extendable = true
		testBasicACLPredefined(t, BasicACLPublicAppendExtended, BasicACLNamePublicAppendExtended, exp)
	})
}
