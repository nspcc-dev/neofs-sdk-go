package subnet

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldSubnetInfoID
	fieldSubnetInfoOwner
)

// MarshaledSize returns size of the SubnetInfo in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *SubnetInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldSubnetInfoID, x.Id)
		sz += proto.SizeEmbedded(fieldSubnetInfoOwner, x.Owner)
	}
	return sz
}

// MarshalStable writes the SubnetInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SubnetInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SubnetInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSubnetInfoID, x.Id)
		proto.MarshalToEmbedded(b[off:], fieldSubnetInfoOwner, x.Owner)
	}
}
