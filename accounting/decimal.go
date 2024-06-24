package accounting

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	"google.golang.org/protobuf/proto"
)

// Decimal represents decimal number for accounting operations.
//
// Decimal is mutually compatible with [accounting.Decimal] message. See
// [Decimal.ReadFromV2] / [Decimal.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type Decimal struct {
	val  int64
	prec uint32
}

// ReadFromV2 reads Decimal from the [accounting.Decimal] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Decimal.WriteToV2].
func (d *Decimal) ReadFromV2(m *accounting.Decimal) error {
	d.val = m.Value
	d.prec = m.Precision
	return nil
}

// WriteToV2 writes Decimal to the [accounting.Decimal] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Decimal.ReadFromV2].
func (d Decimal) WriteToV2(m *accounting.Decimal) {
	m.Value = d.val
	m.Precision = d.prec
}

// Value returns value of the decimal number.
//
// Zero Decimal has zero value.
//
// See also [Decimal.SetValue].
func (d Decimal) Value() int64 {
	return d.val
}

// SetValue sets value of the decimal number.
//
// See also [Decimal.Value].
func (d *Decimal) SetValue(v int64) {
	d.val = v
}

// Precision returns precision of the decimal number.
//
// Zero Decimal has zero precision.
//
// See also [Decimal.SetPrecision].
func (d Decimal) Precision() uint32 {
	return d.prec
}

// SetPrecision sets precision of the decimal number.
//
// See also [Decimal.Precision].
func (d *Decimal) SetPrecision(p uint32) {
	d.prec = p
}

// TODO: why needed? if so, can be non-deterministic?

// Marshal encodes Decimal into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also [Decimal.Unmarshal].
func (d Decimal) Marshal() []byte {
	var m accounting.Decimal
	d.WriteToV2(&m)

	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes NeoFS API protocol binary format into the Decimal
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also [Decimal.Marshal].
func (d *Decimal) Unmarshal(data []byte) error {
	var m accounting.Decimal
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf")
	}

	return d.ReadFromV2(&m)
}
