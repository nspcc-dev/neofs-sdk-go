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
}
