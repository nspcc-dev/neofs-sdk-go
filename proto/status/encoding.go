package status

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldDetailID
	fieldDetailValue
)

// MarshaledSize returns size of the Status_Detail in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *Status_Detail) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldDetailID, x.Id) +
			proto.SizeBytes(fieldDetailValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Status_Detail in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Status_Detail.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Status_Detail) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldDetailID, x.Id)
		proto.MarshalToBytes(b[off:], fieldDetailValue, x.Value)
	}
}

const (
	_ = iota
	fieldStatusCode
	fieldStatusMessage
	fieldStatusDetails
)

// MarshaledSize returns size of the Status in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Status) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(fieldStatusCode, x.Code) +
			proto.SizeBytes(fieldStatusMessage, x.Message) +
			proto.SizeRepeatedMessages(fieldStatusDetails, x.Details)
	}
	return 0
}

// MarshalStable writes the Status in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Status.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Status) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldStatusCode, x.Code)
		off += proto.MarshalToBytes(b[off:], fieldStatusMessage, x.Message)
		proto.MarshalToRepeatedMessages(b[off:], fieldStatusDetails, x.Details)
	}
}
