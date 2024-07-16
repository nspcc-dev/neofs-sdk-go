package accounting

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldDecimalValue
	fieldDecimalPrecision
)

// MarshaledSize returns size of the Decimal in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Decimal) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldDecimalValue, x.Value) +
			proto.SizeVarint(fieldDecimalPrecision, x.Precision)
	}
	return sz
}

// MarshalStable writes the Decimal in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Decimal.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Decimal) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldDecimalValue, x.Value)
		proto.MarshalToVarint(b[off:], fieldDecimalPrecision, x.Precision)
	}
}

const (
	_ = iota
	fieldBalanceReqOwner
)

// MarshaledSize returns size of the BalanceRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *BalanceRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldBalanceReqOwner, x.OwnerId)
	}
	return sz
}

// MarshalStable writes the BalanceRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [BalanceRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *BalanceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldBalanceReqOwner, x.OwnerId)
	}
}

const (
	_ = iota
	fieldBalanceRespBalance
)

// MarshaledSize returns size of the BalanceResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *BalanceResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldBalanceRespBalance, x.Balance)
	}
	return sz
}

// MarshalStable writes the BalanceResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [BalanceResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *BalanceResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldBalanceRespBalance, x.Balance)
	}
}
