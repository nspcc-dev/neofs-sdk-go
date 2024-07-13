package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
)

// Various system attributes.
const (
	// AttributeExpirationEpoch is a key to an object attribute that determines
	// after what epoch the object becomes expired. Objects that do not have this
	// attribute never expire.
	//
	// Reaction of NeoFS system components to the objects' 'expired' property may
	// vary. For example, in the basic scenario, expired objects are auto-deleted
	// from the storage. Detailed behavior can be found in the NeoFS Specification.
	//
	// Note that the value determines exactly the last epoch of the object's
	// relevance: for example, with the value N, the object is relevant in epoch N
	// and expired in any epoch starting from N+1.
	AttributeExpirationEpoch = object.SysAttributeExpEpoch
)

// Attribute represents v2-compatible object attribute.
type Attribute object.Attribute

// NewAttributeFromV2 wraps v2 [object.Attribute] message to [Attribute].
//
// Nil [object.Attribute] converts to nil.
func NewAttributeFromV2(aV2 *object.Attribute) *Attribute {
	return (*Attribute)(aV2)
}

// NewAttribute creates and initializes new [Attribute].
func NewAttribute(key, value string) *Attribute {
	attr := new(object.Attribute)
	attr.SetKey(key)
	attr.SetValue(value)

	return NewAttributeFromV2(attr)
}

// Key returns key to the object attribute.
func (a *Attribute) Key() string {
	return (*object.Attribute)(a).GetKey()
}

// SetKey sets key to the object attribute.
func (a *Attribute) SetKey(v string) {
	(*object.Attribute)(a).SetKey(v)
}

// Value return value of the object attribute.
func (a *Attribute) Value() string {
	return (*object.Attribute)(a).GetValue()
}

// SetValue sets value of the object attribute.
func (a *Attribute) SetValue(v string) {
	(*object.Attribute)(a).SetValue(v)
}

// ToV2 converts [Attribute] to v2 [object.Attribute] message.
//
// Nil [Attribute] converts to nil.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (a *Attribute) ToV2() *object.Attribute {
	return (*object.Attribute)(a)
}

// Marshal marshals [Attribute] into a protobuf binary form.
//
// See also [Attribute.Unmarshal].
func (a *Attribute) Marshal() []byte {
	return (*object.Attribute)(a).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of [Attribute].
//
// See also [Attribute.Marshal].
func (a *Attribute) Unmarshal(data []byte) error {
	return (*object.Attribute)(a).Unmarshal(data)
}

// MarshalJSON encodes [Attribute] to protobuf JSON format.
//
// See also [Attribute.UnmarshalJSON].
func (a *Attribute) MarshalJSON() ([]byte, error) {
	return (*object.Attribute)(a).MarshalJSON()
}

// UnmarshalJSON decodes [Attribute] from protobuf JSON format.
//
// See also [Attribute.MarshalJSON].
func (a *Attribute) UnmarshalJSON(data []byte) error {
	return (*object.Attribute)(a).UnmarshalJSON(data)
}
