package status

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldDetailID
	fieldDetailValue
)

func (x *Status_Detail) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldDetailID, x.Id) +
			proto.SizeBytes(fieldDetailValue, x.Value)
	}
	return sz
}

func (x *Status_Detail) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldDetailID, x.Id)
		proto.MarshalBytes(b[off:], fieldDetailValue, x.Value)
	}
}

const (
	_ = iota
	fieldStatusCode
	fieldStatusMessage
	fieldStatusDetails
)

func (x *Status) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldStatusCode, x.Code) +
			proto.SizeBytes(fieldStatusMessage, x.Message)
		for i := range x.Details {
			sz += proto.SizeNested(fieldStatusDetails, x.Details[i])
		}
	}
	return sz
}

func (x *Status) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldStatusCode, x.Code)
		off += proto.MarshalBytes(b[off:], fieldStatusMessage, x.Message)
		for i := range x.Details {
			off += proto.MarshalNested(b[off:], fieldStatusDetails, x.Details[i])
		}
	}
}
