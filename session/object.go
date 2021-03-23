package session

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
)

// ObjectContext represents NeoFS API v2-compatible
// context of the object session.
//
// It is a wrapper over session.ObjectSessionContext
// which allows abstracting from details of the message
// structure.
type ObjectContext session.ObjectSessionContext

// NewObjectContext creates and returns blank ObjectContext.
//
// Defaults:
//  - not bound to any operation;
//  - nil object address.
func NewObjectContext() *ObjectContext {
	v2 := new(session.ObjectSessionContext)

	return NewObjectContextFromV2(v2)
}

// NewObjectContextFromV2 wraps session.ObjectSessionContext
// into ObjectContext.
func NewObjectContextFromV2(v *session.ObjectSessionContext) *ObjectContext {
	return (*ObjectContext)(v)
}

// ToV2 converts ObjectContext to session.ObjectSessionContext
// message structure.
func (x *ObjectContext) ToV2() *session.ObjectSessionContext {
	return (*session.ObjectSessionContext)(x)
}

// ApplyTo specifies which object the ObjectContext applies to.
func (x *ObjectContext) ApplyTo(a *address.Address) {
	v2 := (*session.ObjectSessionContext)(x)

	if a == nil {
		v2.SetAddress(nil)
		return
	}

	var aV2 refs.Address
	a.WriteToV2(&aV2)

	v2.SetAddress(&aV2)
}

// Address returns identifier of the object
// to which the ObjectContext applies.
func (x *ObjectContext) Address() *address.Address {
	v2 := (*session.ObjectSessionContext)(x)
	var a address.Address

	v2Addr := v2.GetAddress()
	if v2Addr == nil {
		return nil
	}

	a.ReadFromV2(*v2Addr)

	return &a
}

func (x *ObjectContext) forVerb(v session.ObjectSessionVerb) {
	(*session.ObjectSessionContext)(x).
		SetVerb(v)
}

func (x *ObjectContext) isForVerb(v session.ObjectSessionVerb) bool {
	return (*session.ObjectSessionContext)(x).
		GetVerb() == v
}

// ForPut binds the ObjectContext to
// PUT operation.
func (x *ObjectContext) ForPut() {
	x.forVerb(session.ObjectVerbPut)
}

// IsForPut checks if ObjectContext is bound to
// PUT operation.
func (x *ObjectContext) IsForPut() bool {
	return x.isForVerb(session.ObjectVerbPut)
}

// ForDelete binds the ObjectContext to
// DELETE operation.
func (x *ObjectContext) ForDelete() {
	x.forVerb(session.ObjectVerbDelete)
}

// IsForDelete checks if ObjectContext is bound to
// DELETE operation.
func (x *ObjectContext) IsForDelete() bool {
	return x.isForVerb(session.ObjectVerbDelete)
}

// ForGet binds the ObjectContext to
// GET operation.
func (x *ObjectContext) ForGet() {
	x.forVerb(session.ObjectVerbGet)
}

// IsForGet checks if ObjectContext is bound to
// GET operation.
func (x *ObjectContext) IsForGet() bool {
	return x.isForVerb(session.ObjectVerbGet)
}

// ForHead binds the ObjectContext to
// HEAD operation.
func (x *ObjectContext) ForHead() {
	x.forVerb(session.ObjectVerbHead)
}

// IsForHead checks if ObjectContext is bound to
// HEAD operation.
func (x *ObjectContext) IsForHead() bool {
	return x.isForVerb(session.ObjectVerbHead)
}

// ForSearch binds the ObjectContext to
// SEARCH operation.
func (x *ObjectContext) ForSearch() {
	x.forVerb(session.ObjectVerbSearch)
}

// IsForSearch checks if ObjectContext is bound to
// SEARCH operation.
func (x *ObjectContext) IsForSearch() bool {
	return x.isForVerb(session.ObjectVerbSearch)
}

// ForRange binds the ObjectContext to
// RANGE operation.
func (x *ObjectContext) ForRange() {
	x.forVerb(session.ObjectVerbRange)
}

// IsForRange checks if ObjectContext is bound to
// RANGE operation.
func (x *ObjectContext) IsForRange() bool {
	return x.isForVerb(session.ObjectVerbRange)
}

// ForRangeHash binds the ObjectContext to
// RANGEHASH operation.
func (x *ObjectContext) ForRangeHash() {
	x.forVerb(session.ObjectVerbRangeHash)
}

// IsForRangeHash checks if ObjectContext is bound to
// RANGEHASH operation.
func (x *ObjectContext) IsForRangeHash() bool {
	return x.isForVerb(session.ObjectVerbRangeHash)
}

// Marshal marshals ObjectContext into a protobuf binary form.
func (x *ObjectContext) Marshal() ([]byte, error) {
	return x.ToV2().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of ObjectContext.
func (x *ObjectContext) Unmarshal(data []byte) error {
	return x.ToV2().Unmarshal(data)
}

// MarshalJSON encodes ObjectContext to protobuf JSON format.
func (x *ObjectContext) MarshalJSON() ([]byte, error) {
	return x.ToV2().MarshalJSON()
}

// UnmarshalJSON decodes ObjectContext from protobuf JSON format.
func (x *ObjectContext) UnmarshalJSON(data []byte) error {
	return x.ToV2().UnmarshalJSON(data)
}
