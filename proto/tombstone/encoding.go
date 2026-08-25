package tombstone

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

// Field numbers of [Tombstone] message.
const (
	_ = iota
	FieldTombstoneExpirationEpoch
	FieldTombstoneSplitID
	FieldTombstoneMembers
)

// MarshaledSize returns size of the Tombstone in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Tombstone) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeVarint(FieldTombstoneExpirationEpoch, x.ExpirationEpoch) +
			protoencoding.SizeBytes(FieldTombstoneSplitID, x.SplitId) +
			protoencoding.SizeRepeatedMessages(FieldTombstoneMembers, x.Members)
	}
	return 0
}

// MarshalStable writes the Tombstone in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Tombstone.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Tombstone) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldTombstoneExpirationEpoch, x.ExpirationEpoch)
		off += protoencoding.MarshalToBytes(b[off:], FieldTombstoneSplitID, x.SplitId)
		protoencoding.MarshalToRepeatedMessages(b[off:], FieldTombstoneMembers, x.Members)
	}
}
