package subnet

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

// Field numbers of [SubnetInfo] message.
const (
	_ = iota
	FieldSubnetInfoID
	FieldSubnetInfoOwner
)

// MarshaledSize returns size of the SubnetInfo in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *SubnetInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldSubnetInfoID, x.Id)
		sz += proto.SizeEmbedded(FieldSubnetInfoOwner, x.Owner)
	}
	return sz
}

// MarshalStable writes the SubnetInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SubnetInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SubnetInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSubnetInfoID, x.Id)
		proto.MarshalToEmbedded(b[off:], FieldSubnetInfoOwner, x.Owner)
	}
}
