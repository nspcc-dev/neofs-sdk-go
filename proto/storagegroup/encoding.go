package storagegroup

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldStorageGroupSize
	fieldStorageGroupHash
	fieldStorageGroupExp
	fieldStorageGroupMembers
)

// MarshaledSize returns size of the StorageGroup in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *StorageGroup) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldStorageGroupSize, x.ValidationDataSize)
		sz += proto.SizeEmbedded(fieldStorageGroupHash, x.ValidationHash)
		sz += proto.SizeVarint(fieldStorageGroupExp, x.ExpirationEpoch)
		for i := range x.Members {
			sz += proto.SizeEmbedded(fieldStorageGroupMembers, x.Members[i])
		}
	}
	return sz
}

// MarshalStable writes the StorageGroup in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [StorageGroup.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *StorageGroup) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldStorageGroupSize, x.ValidationDataSize)
		off += proto.MarshalToEmbedded(b[off:], fieldStorageGroupHash, x.ValidationHash)
		off += proto.MarshalToVarint(b[off:], fieldStorageGroupExp, x.ExpirationEpoch)
		for i := range x.Members {
			off += proto.MarshalToEmbedded(b[off:], fieldStorageGroupMembers, x.Members[i])
		}
	}
}
