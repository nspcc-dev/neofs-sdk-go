package accounting

import (
	"github.com/nspcc-dev/neofs-api-go/v2/accounting"
)

// Decimal represents decimal number for accounting operations.
//
// Decimal is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/accounting.Decimal
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
//
//	_ = Decimal(accounting.Decimal{}) // not recommended
type Decimal accounting.Decimal

// ReadFromV2 reads Decimal from the accounting.Decimal message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (d *Decimal) ReadFromV2(m accounting.Decimal) error {
	*d = Decimal(m)
	return nil
}

// WriteToV2 writes Decimal to the accounting.Decimal message.
// The message must not be nil.
//
// See also ReadFromV2.
func (d Decimal) WriteToV2(m *accounting.Decimal) {
	*m = (accounting.Decimal)(d)
}

// Value returns value of the decimal number.
//
// Zero Decimal has zero value.
//
// See also SetValue.
func (d Decimal) Value() int64 {
	return (*accounting.Decimal)(&d).GetValue()
}

// SetValue sets value of the decimal number.
//
// See also Value.
func (d *Decimal) SetValue(v int64) {
	(*accounting.Decimal)(d).SetValue(v)
}

// Precision returns precision of the decimal number.
//
// Zero Decimal has zero precision.
//
// See also SetPrecision.
func (d Decimal) Precision() uint32 {
	return (*accounting.Decimal)(&d).GetPrecision()
}

// SetPrecision sets precision of the decimal number.
//
// See also Precision.
func (d *Decimal) SetPrecision(p uint32) {
	(*accounting.Decimal)(d).SetPrecision(p)
}

// Marshal encodes Decimal into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (d Decimal) Marshal() []byte {
	var m accounting.Decimal
	d.WriteToV2(&m)

	return m.StableMarshal(nil)
}

// Unmarshal decodes NeoFS API protocol binary format into the Decimal
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (d *Decimal) Unmarshal(data []byte) error {
	var m accounting.Decimal

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return d.ReadFromV2(m)
}
