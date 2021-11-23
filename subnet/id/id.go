package subnetid

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// ID represents NeoFS subnet identifier.
//
// The type is compatible with the corresponding message from NeoFS API V2 protocol.
//
// Zero value and nil pointer is equivalent to zero subnet ID.
type ID refs.SubnetID

// FromV2 initializes ID from refs.SubnetID message structure. Must not be called on nil.
//
// Note: nil refs.SubnetID corresponds to zero ID value or nil pointer to it.
func (x *ID) FromV2(msg refs.SubnetID) {
	*x = ID(msg)
}

// WriteToV2 writes ID to refs.SubnetID message structure. The message must not be nil.
//
// Note: nil ID corresponds to zero refs.SubnetID value or nil pointer to it.
func (x ID) WriteToV2(msg *refs.SubnetID) {
	*msg = refs.SubnetID(x)
}

// Equals returns true iff both instances identify the same subnet.
//
// Method is NPE-safe: nil pointer equals to pointer to zero value.
func (x *ID) Equals(x2 *ID) bool {
	return (*refs.SubnetID)(x).GetValue() == (*refs.SubnetID)(x2).GetValue()
}

// MarshalText encodes ID into text format according to particular NeoFS API protocol.
// Supported versions:
//  * V2 (see refs.SubnetID type).
//
// Implements encoding.TextMarshaler.
func (x *ID) MarshalText() ([]byte, error) {
	return (*refs.SubnetID)(x).MarshalText()
}

// UnmarshalText decodes ID from the text according to particular NeoFS API protocol.
// Must not be called on nil. Supported versions:
//  * V2 (see refs.SubnetID type).
//
// Implements encoding.TextUnmarshaler.
func (x *ID) UnmarshalText(text []byte) error {
	return (*refs.SubnetID)(x).UnmarshalText(text)
}

// String returns string representation of ID using MarshalText.
// Returns string with message on error.
//
// Implements fmt.Stringer.
func (x *ID) String() string {
	text, err := x.MarshalText()
	if err != nil {
		return fmt.Sprintf("<invalid> %v", err)
	}

	return string(text)
}

// Marshal encodes ID into a binary format of NeoFS API V2 protocol (Protocol Buffers with direct field order).
func (x *ID) Marshal() ([]byte, error) {
	return (*refs.SubnetID)(x).StableMarshal(nil)
}

// Unmarshal decodes ID from NeoFS API V2 binary format (see Marshal). Must not be called on nil.
//
// Note: empty data corresponds to zero ID value or nil pointer to it.
func (x *ID) Unmarshal(data []byte) error {
	return (*refs.SubnetID)(x).Unmarshal(data)
}

// SetNumber sets ID value in uint32 format. Must not be called on nil.
// By default, number is 0 which refers to zero subnet.
func (x *ID) SetNumber(num uint32) {
	(*refs.SubnetID)(x).SetValue(num)
}

// IsZero returns true iff the ID refers to zero subnet.
func IsZero(id ID) bool {
	return id.Equals(nil)
}

// MakeZero makes ID to refer to zero subnet. Arg must not be nil (it is already zero).
//
// Makes no sense to call on zero value (e.g. declared using var).
func MakeZero(id *ID) {
	id.SetNumber(0)
}
