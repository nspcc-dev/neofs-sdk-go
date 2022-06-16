package container

func init() {
	// left-to-right order of the object operations
	orderedOps := [...]ACLOp{
		ACLOpObjectGet,
		ACLOpObjectHead,
		ACLOpObjectPut,
		ACLOpObjectDelete,
		ACLOpObjectSearch,
		ACLOpObjectRange,
		ACLOpObjectHash,
	}

	mOrder = make(map[ACLOp]uint8, len(orderedOps))

	for i := range orderedOps {
		mOrder[orderedOps[i]] = uint8(i)
	}

	// numbers are taken from NeoFS Specification
	BasicACLPrivate.fromUint32(0x1C8C8CCC)
	BasicACLPrivateExtended.fromUint32(0x0C8C8CCC)
	BasicACLPublicRO.fromUint32(0x1FBF8CFF)
	BasicACLPublicROExtended.fromUint32(0x0FBF8CFF)
	BasicACLPublicRW.fromUint32(0x1FBFBFFF)
	BasicACLPublicRWExtended.fromUint32(0x0FBFBFFF)
	BasicACLPublicAppend.fromUint32(0x1FBF9FFF)
	BasicACLPublicAppendExtended.fromUint32(0x0FBF9FFF)
}
