package tombstone

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldTombstoneExp
	fieldTombstoneSplitID
	fieldTombstoneMembers
)

// MarshaledSize returns size of the Tombstone in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Tombstone) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(fieldTombstoneExp, x.ExpirationEpoch) +
			proto.SizeBytes(fieldTombstoneSplitID, x.SplitId) +
			proto.SizeRepeatedMessages(fieldTombstoneMembers, x.Members)
	}
	return 0
}

// MarshalStable writes the Tombstone in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Tombstone.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Tombstone) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTombstoneExp, x.ExpirationEpoch)
		off += proto.MarshalToBytes(b[off:], fieldTombstoneSplitID, x.SplitId)
		proto.MarshalToRepeatedMessages(b[off:], fieldTombstoneMembers, x.Members)
	}
}
