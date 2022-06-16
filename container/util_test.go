package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBits(t *testing.T) {
	num := uint32(0b10110)

	require.False(t, isBitSet(num, 0))
	require.True(t, isBitSet(num, 1))
	require.True(t, isBitSet(num, 2))
	require.False(t, isBitSet(num, 3))
	require.True(t, isBitSet(num, 4))
	require.False(t, isBitSet(num, 5))

	setBit(&num, 3)
	require.EqualValues(t, 0b11110, num)

	setBit(&num, 6)
	require.EqualValues(t, 0b1011110, num)

	resetBit(&num, 1)
	require.EqualValues(t, 0b1011100, num)
}

func TestOpBits(t *testing.T) {
	num := uint32(0b_1001_0101_1100_0011_0110_0111_1000_1111)

	require.Panics(t, func() { isOpBitSet(num, aclOpZero, 0) })
	require.Panics(t, func() { isOpBitSet(num, aclOpLast, 0) })

	cpNum := num

	require.Panics(t, func() { setOpBit(&num, aclOpZero, 0) })
	require.EqualValues(t, cpNum, num)
	require.Panics(t, func() { setOpBit(&num, aclOpLast, 0) })
	require.EqualValues(t, cpNum, num)

	for _, tc := range []struct {
		op   ACLOp
		set  [4]bool   // is bit set (left-to-right)
		bits [4]uint32 // result of setting i-th bit (left-to-right) to zero num
	}{
		{
			op:  ACLOpObjectHash,
			set: [4]bool{false, true, false, true},
			bits: [4]uint32{
				0b_0000_1000_0000_0000_0000_0000_0000_0000,
				0b_0000_0100_0000_0000_0000_0000_0000_0000,
				0b_0000_0010_0000_0000_0000_0000_0000_0000,
				0b_0000_0001_0000_0000_0000_0000_0000_0000,
			},
		},
		{
			op:  ACLOpObjectRange,
			set: [4]bool{true, true, false, false},
			bits: [4]uint32{
				0b_0000_0000_1000_0000_0000_0000_0000_0000,
				0b_0000_0000_0100_0000_0000_0000_0000_0000,
				0b_0000_0000_0010_0000_0000_0000_0000_0000,
				0b_0000_0000_0001_0000_0000_0000_0000_0000,
			},
		},
		{
			op:  ACLOpObjectSearch,
			set: [4]bool{false, false, true, true},
			bits: [4]uint32{
				0b_0000_0000_0000_1000_0000_0000_0000_0000,
				0b_0000_0000_0000_0100_0000_0000_0000_0000,
				0b_0000_0000_0000_0010_0000_0000_0000_0000,
				0b_0000_0000_0000_0001_0000_0000_0000_0000,
			},
		},
		{
			op:  ACLOpObjectDelete,
			set: [4]bool{false, true, true, false},
			bits: [4]uint32{
				0b_0000_0000_0000_0000_1000_0000_0000_0000,
				0b_0000_0000_0000_0000_0100_0000_0000_0000,
				0b_0000_0000_0000_0000_0010_0000_0000_0000,
				0b_0000_0000_0000_0000_0001_0000_0000_0000,
			},
		},
		{
			op:  ACLOpObjectPut,
			set: [4]bool{false, true, true, true},
			bits: [4]uint32{
				0b_0000_0000_0000_0000_0000_1000_0000_0000,
				0b_0000_0000_0000_0000_0000_0100_0000_0000,
				0b_0000_0000_0000_0000_0000_0010_0000_0000,
				0b_0000_0000_0000_0000_0000_0001_0000_0000,
			},
		},
		{
			op:  ACLOpObjectHead,
			set: [4]bool{true, false, false, false},
			bits: [4]uint32{
				0b_0000_0000_0000_0000_0000_0000_1000_0000,
				0b_0000_0000_0000_0000_0000_0000_0100_0000,
				0b_0000_0000_0000_0000_0000_0000_0010_0000,
				0b_0000_0000_0000_0000_0000_0000_0001_0000,
			},
		},
		{
			op:  ACLOpObjectGet,
			set: [4]bool{true, true, true, true},
			bits: [4]uint32{
				0b_0000_0000_0000_0000_0000_0000_0000_1000,
				0b_0000_0000_0000_0000_0000_0000_0000_0100,
				0b_0000_0000_0000_0000_0000_0000_0000_0010,
				0b_0000_0000_0000_0000_0000_0000_0000_0001,
			},
		},
	} {
		for i := range tc.set {
			require.EqualValues(t, tc.set[i], isOpBitSet(num, tc.op, uint8(len(tc.set)-1-i)),
				fmt.Sprintf("op %s, left bit #%d", tc.op, i),
			)
		}

		for i := range tc.bits {
			num := uint32(0)

			setOpBit(&num, tc.op, uint8(len(tc.bits)-1-i))

			require.EqualValues(t, tc.bits[i], num)
		}
	}
}
