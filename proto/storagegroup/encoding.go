package storagegroup

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

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
		return protoencoding.SizeVarint(FieldStorageGroupValidationDataSize, x.ValidationDataSize) +
			protoencoding.SizeEmbedded(FieldStorageGroupValidationDataHash, x.ValidationHash) +
			protoencoding.SizeVarint(FieldStorageGroupExpirationEpoch, x.ExpirationEpoch) +
			protoencoding.SizeRepeatedMessages(FieldStorageGroupMembers, x.Members)
	}
	return 0
}

// MarshalStable writes the StorageGroup in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [StorageGroup.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *StorageGroup) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldStorageGroupValidationDataSize, x.ValidationDataSize)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldStorageGroupValidationDataHash, x.ValidationHash)
		off += protoencoding.MarshalToVarint(b[off:], FieldStorageGroupExpirationEpoch, x.ExpirationEpoch)
		protoencoding.MarshalToRepeatedMessages(b[off:], FieldStorageGroupMembers, x.Members)
	}
}
