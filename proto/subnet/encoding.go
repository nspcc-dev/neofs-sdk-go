package subnet

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

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
		sz = protoencoding.SizeEmbedded(FieldSubnetInfoID, x.Id)
		sz += protoencoding.SizeEmbedded(FieldSubnetInfoOwner, x.Owner)
	}
	return sz
}

// MarshalStable writes the SubnetInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SubnetInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SubnetInfo) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSubnetInfoID, x.Id)
		protoencoding.MarshalToEmbedded(b[off:], FieldSubnetInfoOwner, x.Owner)
	}
}
