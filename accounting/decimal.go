package accounting

import (
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protoaccounting "github.com/nspcc-dev/neofs-sdk-go/proto/accounting"
)

// Decimal represents decimal number for accounting operations.
//
// Decimal is mutually compatible with [protoaccounting.Decimal] message. See
// [Decimal.FromProtoMessage] / [Decimal.FromProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type Decimal struct {
	val  int64
	prec uint32
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// d from it.
//
// See also [Decimal.ProtoMessage].
func (d *Decimal) FromProtoMessage(m *protoaccounting.Decimal) error {
	d.val = m.Value
	d.prec = m.Precision
	return nil
}

// ProtoMessage converts d into message to transmit using the NeoFS API
// protocol.
//
// See also [Decimal.FromProtoMessage].
func (d Decimal) ProtoMessage() *protoaccounting.Decimal {
	return &protoaccounting.Decimal{
		Value:     d.val,
		Precision: d.prec,
	}
}

// Value returns value of the decimal number.
//
// Zero Decimal has zero value.
//
// See also SetValue.
func (d Decimal) Value() int64 {
	return d.val
}

// SetValue sets value of the decimal number.
//
// See also Value.
func (d *Decimal) SetValue(v int64) {
	d.val = v
}

// Precision returns precision of the decimal number.
//
// Zero Decimal has zero precision.
//
// See also SetPrecision.
func (d Decimal) Precision() uint32 {
	return d.prec
}

// SetPrecision sets precision of the decimal number.
//
// See also Precision.
func (d *Decimal) SetPrecision(p uint32) {
	d.prec = p
}

// Marshal encodes Decimal into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (d Decimal) Marshal() []byte {
	return neofsproto.Marshal(d)
}

// Unmarshal decodes NeoFS API protocol binary format into the Decimal
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (d *Decimal) Unmarshal(data []byte) error {
	return neofsproto.Unmarshal(data, d)
}
