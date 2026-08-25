package accounting

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

// Field numbers of [Decimal] message.
const (
	_ = iota
	FieldDecimalValue
	FieldDecimalPrecision
)

// MarshaledSize returns size of the Decimal in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Decimal) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeVarint(FieldDecimalValue, x.Value) +
			protoencoding.SizeVarint(FieldDecimalPrecision, x.Precision)
	}
	return sz
}

// MarshalStable writes the Decimal in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Decimal.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Decimal) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldDecimalValue, x.Value)
		protoencoding.MarshalToVarint(b[off:], FieldDecimalPrecision, x.Precision)
	}
}

// Field numbers of [BalanceRequest_Body] message.
const (
	_ = iota
	FieldBalanceRequestBodyOwner
)

// MarshaledSize returns size of the BalanceRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *BalanceRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldBalanceRequestBodyOwner, x.OwnerId)
	}
	return sz
}

// MarshalStable writes the BalanceRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [BalanceRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *BalanceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldBalanceRequestBodyOwner, x.OwnerId)
	}
}

// Field numbers of [BalanceResponse_Body] message.
const (
	_ = iota
	FieldBalanceResponseBodyBalance
)

// MarshaledSize returns size of the BalanceResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *BalanceResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldBalanceResponseBodyBalance, x.Balance)
	}
	return sz
}

// MarshalStable writes the BalanceResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [BalanceResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *BalanceResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldBalanceResponseBodyBalance, x.Balance)
	}
}
