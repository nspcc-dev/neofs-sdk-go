package acl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic_DisableExtension(t *testing.T) {
	var val, val2 Basic

	require.True(t, val.Extendable())
	val2.FromBits(val.Bits())
	require.True(t, val2.Extendable())

	val.DisableExtension()

	require.False(t, val.Extendable())
	val2.FromBits(val.Bits())
	require.False(t, val2.Extendable())
}

func TestBasic_MakeSticky(t *testing.T) {
	var val, val2 Basic

	require.False(t, val.Sticky())
	val2.FromBits(val.Bits())
	require.False(t, val2.Sticky())

	val.MakeSticky()

	require.True(t, val.Sticky())
	val2.FromBits(val.Bits())
	require.True(t, val2.Sticky())
}

func TestBasic_AllowBearerRules(t *testing.T) {
	var val Basic

	require.Panics(t, func() { val.AllowBearerRules(opZero) })
	require.Panics(t, func() { val.AllowBearerRules(opLast) })

	require.Panics(t, func() { val.AllowedBearerRules(opZero) })
	require.Panics(t, func() { val.AllowedBearerRules(opLast) })

	for op := opZero + 1; op < opLast; op++ {
		val := val

		require.False(t, val.AllowedBearerRules(op))

		val.AllowBearerRules(op)

		for j := opZero + 1; j < opLast; j++ {
			if j == op {
				require.True(t, val.AllowedBearerRules(j), op)
			} else {
				require.False(t, val.AllowedBearerRules(j), op)
			}
		}
	}
}

func TestBasic_AllowOp(t *testing.T) {
	var val, val2 Basic

	require.Panics(t, func() { val.IsOpAllowed(opZero, roleZero+1) })
	require.Panics(t, func() { val.IsOpAllowed(opLast, roleZero+1) })
	require.Panics(t, func() { val.IsOpAllowed(opZero+1, roleZero) })
	require.Panics(t, func() { val.IsOpAllowed(opZero+1, roleLast) })

	for op := opZero + 1; op < opLast; op++ {
		require.Panics(t, func() { val.AllowOp(op, RoleInnerRing) })

		if isReplicationOp(op) {
			require.Panics(t, func() { val.AllowOp(op, RoleContainer) })
			require.True(t, val.IsOpAllowed(op, RoleContainer))
		}
	}

	require.True(t, val.IsOpAllowed(OpObjectGet, RoleInnerRing))
	require.True(t, val.IsOpAllowed(OpObjectHead, RoleInnerRing))
	require.True(t, val.IsOpAllowed(OpObjectSearch, RoleInnerRing))
	require.True(t, val.IsOpAllowed(OpObjectHash, RoleInnerRing))

	const op = opZero + 1
	const role = RoleOthers

	require.False(t, val.IsOpAllowed(op, role))
	val2.FromBits(val.Bits())
	require.False(t, val2.IsOpAllowed(op, role))

	val.AllowOp(op, role)

	require.True(t, val.IsOpAllowed(op, role))
	val2.FromBits(val.Bits())
	require.True(t, val2.IsOpAllowed(op, role))
}

type opsExpected struct {
	owner, container, innerRing, others, bearer bool
}

func testOp(t *testing.T, v Basic, op Op, exp opsExpected) {
	require.Equal(t, exp.owner, v.IsOpAllowed(op, RoleOwner), op)
	require.Equal(t, exp.container, v.IsOpAllowed(op, RoleContainer), op)
	require.Equal(t, exp.innerRing, v.IsOpAllowed(op, RoleInnerRing), op)
	require.Equal(t, exp.others, v.IsOpAllowed(op, RoleOthers), op)
	require.Equal(t, exp.bearer, v.AllowedBearerRules(op), op)
}

type expected struct {
	extendable, sticky bool

	mOps map[Op]opsExpected
}

func testBasicPredefined(t *testing.T, val Basic, name string, exp expected) {
	require.Equal(t, exp.sticky, val.Sticky())
	require.Equal(t, exp.extendable, val.Extendable())

	for op, exp := range exp.mOps {
		testOp(t, val, op, exp)
	}

	s := val.EncodeToString()

	var val2 Basic

	require.NoError(t, val2.DecodeString(s))
	require.Equal(t, val, val2)

	require.NoError(t, val2.DecodeString(name))
	require.Equal(t, val, val2)
}

func TestBasicPredefined(t *testing.T) {
	t.Run("private", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[Op]opsExpected{
				OpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
				OpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				OpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
				OpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				OpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				OpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
				OpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    false,
					bearer:    false,
				},
			},
		}

		testBasicPredefined(t, Private, NamePrivate, exp)
		exp.extendable = true
		testBasicPredefined(t, PrivateExtended, NamePrivateExtended, exp)
	})

	t.Run("public-read", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[Op]opsExpected{
				OpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				OpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				OpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    false,
					bearer:    false,
				},
				OpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
			},
		}

		testBasicPredefined(t, PublicRO, NamePublicRO, exp)
		exp.extendable = true
		testBasicPredefined(t, PublicROExtended, NamePublicROExtended, exp)
	})

	t.Run("public-read-write", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[Op]opsExpected{
				OpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				OpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				OpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				OpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
			},
		}

		testBasicPredefined(t, PublicRW, NamePublicRW, exp)
		exp.extendable = true
		testBasicPredefined(t, PublicRWExtended, NamePublicRWExtended, exp)
	})

	t.Run("public-append", func(t *testing.T) {
		exp := expected{
			extendable: false,
			sticky:     false,
			mOps: map[Op]opsExpected{
				OpObjectHash: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectRange: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				OpObjectSearch: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectDelete: {
					owner:     true,
					container: false,
					innerRing: false,
					others:    false,
					bearer:    true,
				},
				OpObjectPut: {
					owner:     true,
					container: true,
					innerRing: false,
					others:    true,
					bearer:    true,
				},
				OpObjectHead: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
				OpObjectGet: {
					owner:     true,
					container: true,
					innerRing: true,
					others:    true,
					bearer:    true,
				},
			},
		}

		testBasicPredefined(t, PublicAppend, NamePublicAppend, exp)
		exp.extendable = true
		testBasicPredefined(t, PublicAppendExtended, NamePublicAppendExtended, exp)
	})
}
