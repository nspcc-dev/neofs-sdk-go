package container

import (
	"github.com/nspcc-dev/neofs-api-go/v2/container"
)

// Attribute represents container attribute in NeoFS.
//
// Attribute is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/container.Attribute
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Attribute(container.Attribute{}) // not recommended
type Attribute container.Attribute

// Attributes groups container attributes.
type Attributes []Attribute

// ReadFromV2 reads Attribute from the container.Attribute message.
//
// See also WriteToV2.
func (a *Attribute) ReadFromV2(m container.Attribute) {
	*a = Attribute(m)
}

// WriteToV2 writes Attribute to the container.Attribute message.
// The message must not be nil.
//
// See also ReadFromV2.
func (a Attribute) WriteToV2(m *container.Attribute) {
	*m = (container.Attribute)(a)
}

// Key returns attribute's key.
//
// Zero Attribute has empty key.
//
// See also SetKey.
func (a Attribute) Key() string {
	v2 := (container.Attribute)(a)
	return v2.GetKey()
}

// SetKey sets attribute's key.
//
// See also Key.
func (a *Attribute) SetKey(v string) {
	(*container.Attribute)(a).SetKey(v)
}

// Value returns attribute's value.
//
// Zero Attribute has empty value.
//
// See also SetValue.
func (a Attribute) Value() string {
	v2 := (container.Attribute)(a)
	return v2.GetValue()
}

// SetValue sets attribute's value.
//
// See also Value.
func (a *Attribute) SetValue(v string) {
	(*container.Attribute)(a).SetValue(v)
}

// ReadFromV2 reads Attributes from the []container.Attribute.
//
// See also WriteToV2.
func (aa *Attributes) ReadFromV2(m []container.Attribute) {
	attrs := make(Attributes, len(m))
	var attr Attribute

	for i := range m {
		attr.ReadFromV2(m[i])
		attrs[i] = attr
	}

	*aa = attrs
}

// WriteToV2 writes Attributes to the []container.Attribute.
// The message must not be nil.
//
// See also ReadFromV2.
func (aa Attributes) WriteToV2(m *[]container.Attribute) {
	attrs := make([]container.Attribute, len(aa))
	var attrV2 container.Attribute

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

// sets value of the attribute by key.
func setAttribute(c *Container, key, value string) {
	attrs := c.Attributes()
	found := false

	for i := range attrs {
		if attrs[i].Key() == key {
			attrs[i].SetValue(value)
			found = true
			break
		}
	}

	if !found {
		index := len(attrs)
		attrs = append(attrs, Attribute{})
		attrs[index].SetKey(key)
		attrs[index].SetValue(value)
	}

	c.SetAttributes(attrs)
}

// iterates over container attributes. Stops at f's true return.
//
// Handler must not be nil.
func iterateAttributes(c *Container, f func(*Attribute) bool) {
	attrs := c.Attributes()
	for i := range attrs {
		if f(&attrs[i]) {
			return
		}
	}
}

// SetNativeNameWithZone sets container native name and its zone.
//
// Use SetNativeName to set default zone.
//
// See also GetNativeNameWithZone.
func SetNativeNameWithZone(c *Container, name, zone string) {
	setAttribute(c, container.SysAttributeName, name)
	setAttribute(c, container.SysAttributeZone, zone)
}

// SetNativeName sets container native name with default zone (container).
func SetNativeName(c *Container, name string) {
	SetNativeNameWithZone(c, name, container.SysAttributeZoneDefault)
}

// GetNativeNameWithZone returns container native name and its zone.
//
// See also SetNativeNameWithZone.
func GetNativeNameWithZone(c *Container) (name string, zone string) {
	iterateAttributes(c, func(a *Attribute) bool {
		if key := a.Key(); key == container.SysAttributeName {
			name = a.Value()
		} else if key == container.SysAttributeZone {
			zone = a.Value()
		}

		return name != "" && zone != ""
	})

	return
}
