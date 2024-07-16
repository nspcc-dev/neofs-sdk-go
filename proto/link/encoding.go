package link

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldLinkMeasuredObjectID
	fieldLinkMeasuredObjectSize
)

// MarshaledSize returns size of the Link_MeasuredObject in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *Link_MeasuredObject) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldLinkMeasuredObjectID, x.Id) +
			proto.SizeVarint(fieldLinkMeasuredObjectSize, x.Size)
	}
	return sz
}

// MarshalStable writes the Link_MeasuredObject in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [Link_MeasuredObject.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *Link_MeasuredObject) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldLinkMeasuredObjectID, x.Id)
		proto.MarshalToVarint(b[off:], fieldLinkMeasuredObjectSize, x.Size)
	}
}

const (
	_ = iota
	fieldLinkChildren
)

// MarshaledSize returns size of the Link in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Link) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.Children {
			sz += proto.SizeEmbedded(fieldLinkChildren, x.Children[i])
		}
	}
	return sz
}

// MarshalStable writes the Link in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Link.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Link) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Children {
			off += proto.MarshalToEmbedded(b[off:], fieldLinkChildren, x.Children[i])
		}
	}
}
