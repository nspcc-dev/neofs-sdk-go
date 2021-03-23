package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
)

// Attribute represents v2-compatible object attribute.
//
// Attribute is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/object.Attribute
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Attribute(object.Attribute{}) // not recommended
type Attribute object.Attribute

// ReadFromV2 reads Attribute from the object.Attribute message.
//
// See also WriteToV2.
func (a *Attribute) ReadFromV2(m object.Attribute) {
	*a = Attribute(m)
}

// WriteToV2 writes Attribute to the object.Attribute message.
// The message must not be nil.
//
// See also ReadFromV2.
func (a Attribute) WriteToV2(m *object.Attribute) {
	*m = (object.Attribute)(a)
}

// Key returns key to the object attribute.
//
// Zero Attribute has empty key.
//
// See also SetKey.
func (a Attribute) Key() string {
	v2 := (object.Attribute)(a)
	return v2.GetKey()
}

// SetKey sets key to the object attribute.
//
// See also Key.
func (a *Attribute) SetKey(v string) {
	(*object.Attribute)(a).SetKey(v)
}

// Value returns value of the object attribute.
//
// Zero Attribute has empty value.
//
// See also SetValue.
func (a Attribute) Value() string {
	v2 := (object.Attribute)(a)
	return v2.GetValue()
}

// SetValue sets value of the object attribute.
//
// See also Value.
func (a *Attribute) SetValue(v string) {
	(*object.Attribute)(a).SetValue(v)
}

// Attributes groups object attributes.
type Attributes []Attribute

// ReadFromV2 reads Attributes from the []object.Attribute.
//
// See also WriteToV2.
func (aa *Attributes) ReadFromV2(m []object.Attribute) {
	attrs := make(Attributes, len(m))
	var attr Attribute

	for i := range m {
		attr.ReadFromV2(m[i])
		attrs[i] = attr
	}

	*aa = attrs
}

// WriteToV2 writes Attributes to the []object.Attribute.
// The message must not be nil.
//
// See also ReadFromV2.
func (aa Attributes) WriteToV2(m *[]object.Attribute) {
	attrs := make([]object.Attribute, len(aa))
	var attrV2 object.Attribute

	for i := range aa {
		aa[i].WriteToV2(&attrV2)
		attrs[i] = attrV2
	}

	*m = attrs
}

// Len returns the number of attributes.
//
// Zero Attributes has 0 length.
func (aa Attributes) Len() int {
	return len(aa)
}

// Append appends attributes.
func (aa *Attributes) Append(newA ...Attribute) {
	newLen := len(newA) + len(*aa)

	if newLen > cap(*aa) {
		newSlice := make([]Attribute, 0, newLen)
		newSlice = append(newSlice, append(*aa, newA...)...)

		*aa = newSlice

		return
	}

	*aa = append(*aa, newA...)
}

// Iterate iterates attributes and calls passes function on
// them. Stops when either all attributes have been handled,
// or the passed functions returned `true`.
func (aa Attributes) Iterate(f func(attribute Attribute) bool) {
	for i := range aa {
		if f(aa[i]) {
			return
		}
	}
}

// Marshal marshals Attribute into a protobuf binary form.
func (a Attribute) Marshal() ([]byte, error) {
	v2 := (object.Attribute)(a)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Attribute.
func (a *Attribute) Unmarshal(data []byte) error {
	return (*object.Attribute)(a).Unmarshal(data)
}

// MarshalJSON encodes Attribute to protobuf JSON format.
func (a Attribute) MarshalJSON() ([]byte, error) {
	v2 := (object.Attribute)(a)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes Attribute from protobuf JSON format.
func (a *Attribute) UnmarshalJSON(data []byte) error {
	return (*object.Attribute)(a).UnmarshalJSON(data)
}
