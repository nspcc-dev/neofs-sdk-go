package acl

func init() {
	// left-to-right order of the object operations
	orderedOps := [...]Op{
		OpObjectGet,
		OpObjectHead,
		OpObjectPut,
		OpObjectDelete,
		OpObjectSearch,
		OpObjectRange,
		OpObjectHash,
	}

	mOrder = make(map[Op]uint8, len(orderedOps))

	for i := range orderedOps {
		mOrder[orderedOps[i]] = uint8(i)
	}

	// numbers are taken from NeoFS Specification
	Private.FromBits(0x1C8C8CCC)
	PrivateExtended.FromBits(0x0C8C8CCC)
	PublicRO.FromBits(0x1FBF8CFF)
	PublicROExtended.FromBits(0x0FBF8CFF)
	PublicRW.FromBits(0x1FBFBFFF)
	PublicRWExtended.FromBits(0x0FBFBFFF)
	PublicAppend.FromBits(0x1FBF9FFF)
	PublicAppendExtended.FromBits(0x0FBF9FFF)
}
