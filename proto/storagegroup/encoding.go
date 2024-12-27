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
	if x != nil {
		return proto.SizeVarint(fieldStorageGroupSize, x.ValidationDataSize) +
			proto.SizeEmbedded(fieldStorageGroupHash, x.ValidationHash) +
			proto.SizeVarint(fieldStorageGroupExp, x.ExpirationEpoch) +
			proto.SizeRepeatedMessages(fieldStorageGroupMembers, x.Members)
	}
	return 0
}

// MarshalStable writes the StorageGroup in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [StorageGroup.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *StorageGroup) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldStorageGroupSize, x.ValidationDataSize)
		off += proto.MarshalToEmbedded(b[off:], fieldStorageGroupHash, x.ValidationHash)
		off += proto.MarshalToVarint(b[off:], fieldStorageGroupExp, x.ExpirationEpoch)
		proto.MarshalToRepeatedMessages(b[off:], fieldStorageGroupMembers, x.Members)
	}
}
