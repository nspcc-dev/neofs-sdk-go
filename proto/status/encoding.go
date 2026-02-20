package status

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Field numbers of [Status_Detail] message.
const (
	_ = iota
	FieldStatusDetailID
	FieldStatusDetailValue
)

// MarshaledSize returns size of the Status_Detail in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *Status_Detail) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldStatusDetailID, x.Id) +
			proto.SizeBytes(FieldStatusDetailValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Status_Detail in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Status_Detail.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Status_Detail) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldStatusDetailID, x.Id)
		proto.MarshalToBytes(b[off:], FieldStatusDetailValue, x.Value)
	}
}

// Field numbers of [Status_Detail] message.
const (
	_ = iota
	FieldStatusCode
	FieldStatusMessage
	FieldStatusDetails
)

// MarshaledSize returns size of the Status in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Status) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldStatusCode, x.Code) +
			proto.SizeBytes(FieldStatusMessage, x.Message) +
			proto.SizeRepeatedMessages(FieldStatusDetails, x.Details)
	}
	return 0
}

// MarshalStable writes the Status in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Status.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Status) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldStatusCode, x.Code)
		off += proto.MarshalToBytes(b[off:], FieldStatusMessage, x.Message)
		proto.MarshalToRepeatedMessages(b[off:], FieldStatusDetails, x.Details)
	}
}
