package lock

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldLockMembers
)

// MarshaledSize returns size of the Lock in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Lock) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.Members {
			sz += proto.SizeEmbedded(fieldLockMembers, x.Members[i])
		}
	}
	return sz
}

// MarshalStable writes the Lock in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Lock.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Lock) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Members {
			off += proto.MarshalToEmbedded(b[off:], fieldLockMembers, x.Members[i])
		}
	}
}
