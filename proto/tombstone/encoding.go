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
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTombstoneExp, x.ExpirationEpoch)
		sz += proto.SizeBytes(fieldTombstoneSplitID, x.SplitId)
		for i := range x.Members {
			sz += proto.SizeEmbedded(fieldTombstoneMembers, x.Members[i])
		}
	}
	return sz
}

// MarshalStable writes the Tombstone in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Tombstone.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Tombstone) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTombstoneExp, x.ExpirationEpoch)
		off += proto.MarshalToBytes(b[off:], fieldTombstoneSplitID, x.SplitId)
		for i := range x.Members {
			off += proto.MarshalToEmbedded(b[off:], fieldTombstoneMembers, x.Members[i])
		}
	}
}
