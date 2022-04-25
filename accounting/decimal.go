package accounting

import "github.com/nspcc-dev/neofs-api-go/v2/accounting"

// Decimal represents decimal number for accounting operations.
type Decimal accounting.Decimal

// NewDecimal creates, initializes and returns empty Decimal instance.
//
// Defaults:
//  - value: 0
//  - precision: 0
func NewDecimal() *Decimal {
	return NewDecimalFromV2(new(accounting.Decimal))
}

// NewDecimalFromV2 converts v2 Decimal to Decimal.
//
// Nil Decimal converts to nil.
func NewDecimalFromV2(d *accounting.Decimal) *Decimal {
	return (*Decimal)(d)
}

// ToV2 returns the v2 Decimal message.
//
// Nil Decimal converts to nil.
func (d *Decimal) ToV2() *accounting.Decimal {
	return (*accounting.Decimal)(d)
}

// Value returns value of the decimal number.
func (d *Decimal) Value() int64 {
	return (*accounting.Decimal)(d).GetValue()
}

// SetValue sets value of the decimal number.
func (d *Decimal) SetValue(v int64) {
	(*accounting.Decimal)(d).SetValue(v)
}

// Precision returns precision of the decimal number.
func (d *Decimal) Precision() uint32 {
	return (*accounting.Decimal)(d).GetPrecision()
}

// SetPrecision sets precision of the decimal number.
func (d *Decimal) SetPrecision(p uint32) {
	(*accounting.Decimal)(d).SetPrecision(p)
}

// Marshal marshals Decimal into a protobuf binary form.
func (d *Decimal) Marshal() []byte {
	return (*accounting.Decimal)(d).StableMarshal(nil)
}

// Unmarshal unmarshalls protobuf binary representation of Decimal.
func (d *Decimal) Unmarshal(data []byte) error {
	return (*accounting.Decimal)(d).Unmarshal(data)
}

// MarshalJSON encodes Decimal to protobuf JSON format.
func (d *Decimal) MarshalJSON() ([]byte, error) {
	return (*accounting.Decimal)(d).MarshalJSON()
}

// UnmarshalJSON decodes Decimal from protobuf JSON format.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	return (*accounting.Decimal)(d).UnmarshalJSON(data)
}
