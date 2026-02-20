package tombstone

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

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
		return proto.SizeVarint(FieldTombstoneExpirationEpoch, x.ExpirationEpoch) +
			proto.SizeBytes(FieldTombstoneSplitID, x.SplitId) +
			proto.SizeRepeatedMessages(FieldTombstoneMembers, x.Members)
	}
	return 0
}

// MarshalStable writes the Tombstone in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Tombstone.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Tombstone) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldTombstoneExpirationEpoch, x.ExpirationEpoch)
		off += proto.MarshalToBytes(b[off:], FieldTombstoneSplitID, x.SplitId)
		proto.MarshalToRepeatedMessages(b[off:], FieldTombstoneMembers, x.Members)
	}
}
