package container

import "fmt"

// sets n-th bit in num (starting at 0).
func setBit(num *uint32, n uint8) {
	*num |= 1 << n
}

// resets n-th bit in num (starting at 0).
func resetBit(num *uint32, n uint8) {
	var mask uint32
	setBit(&mask, n)

	*num &= ^mask
}

// checks if n-th bit in num is set (starting at 0).
func isBitSet(num uint32, n uint8) bool {
	mask := uint32(1 << n)
	return mask != 0 && num&mask == mask
}

// maps ACLOp to op-section index in BasicACL. Filled on init.
var mOrder map[ACLOp]uint8

// sets n-th bit in num for the given op. Panics if op is unsupported.
func setOpBit(num *uint32, op ACLOp, opBitPos uint8) {
	n, ok := mOrder[op]
	if !ok {
		panic(fmt.Sprintf("op is unsupported %v", op))
	}

	setBit(num, n*bitsPerOp+opBitPos)
}

// checks if n-th bit in num for the given op is set. Panics if op is unsupported.
func isOpBitSet(num uint32, op ACLOp, n uint8) bool {
	off, ok := mOrder[op]
	if !ok {
		panic(fmt.Sprintf("op is unsupported %v", op))
	}

	return isBitSet(num, bitsPerOp*off+n)
}
