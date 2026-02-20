package storagegroup

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

// Field numbers of [StorageGroup] message.
const (
	_ = iota
	FieldStorageGroupValidationDataSize
	FieldStorageGroupValidationDataHash
	FieldStorageGroupExpirationEpoch
	FieldStorageGroupMembers
)

// MarshaledSize returns size of the StorageGroup in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *StorageGroup) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldStorageGroupValidationDataSize, x.ValidationDataSize) +
			proto.SizeEmbedded(FieldStorageGroupValidationDataHash, x.ValidationHash) +
			proto.SizeVarint(FieldStorageGroupExpirationEpoch, x.ExpirationEpoch) +
			proto.SizeRepeatedMessages(FieldStorageGroupMembers, x.Members)
	}
	return 0
}

// MarshalStable writes the StorageGroup in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [StorageGroup.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *StorageGroup) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldStorageGroupValidationDataSize, x.ValidationDataSize)
		off += proto.MarshalToEmbedded(b[off:], FieldStorageGroupValidationDataHash, x.ValidationHash)
		off += proto.MarshalToVarint(b[off:], FieldStorageGroupExpirationEpoch, x.ExpirationEpoch)
		proto.MarshalToRepeatedMessages(b[off:], FieldStorageGroupMembers, x.Members)
	}
}
