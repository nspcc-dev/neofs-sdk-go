package link

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

// Field numbers of [Link_MeasuredObject] message.
const (
	_ = iota
	FieldLinkMeasuredObjectID
	FieldLinkMeasuredObjectSize
)

// MarshaledSize returns size of the Link_MeasuredObject in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *Link_MeasuredObject) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldLinkMeasuredObjectID, x.Id) +
			protoencoding.SizeVarint(FieldLinkMeasuredObjectSize, x.Size)
	}
	return sz
}

// MarshalStable writes the Link_MeasuredObject in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [Link_MeasuredObject.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *Link_MeasuredObject) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldLinkMeasuredObjectID, x.Id)
		protoencoding.MarshalToVarint(b[off:], FieldLinkMeasuredObjectSize, x.Size)
	}
}

// Field numbers of [Link] message.
const (
	_ = iota
	FieldLinkChildren
)

// MarshaledSize returns size of the Link in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Link) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeRepeatedMessages(FieldLinkChildren, x.Children)
	}
	return 0
}

// MarshalStable writes the Link in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Link.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Link) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToRepeatedMessages(b, FieldLinkChildren, x.Children)
	}
}
