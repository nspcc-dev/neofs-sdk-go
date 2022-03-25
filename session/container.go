package session

import (
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// ContainerContext represents NeoFS API v2-compatible
// context of the container session.
//
// It is a wrapper over session.ContainerSessionContext
// which allows to abstract from details of the message
// structure.
type ContainerContext session.ContainerSessionContext

// NewContainerContext creates and returns blank ContainerSessionContext.
//
// Defaults:
//  - not bound to any operation;
//  - applied to all containers.
func NewContainerContext() *ContainerContext {
	v2 := new(session.ContainerSessionContext)
	v2.SetWildcard(true)

	return NewContainerContextFromV2(v2)
}

// NewContainerContextFromV2 wraps session.ContainerSessionContext
// into ContainerContext.
func NewContainerContextFromV2(v *session.ContainerSessionContext) *ContainerContext {
	return (*ContainerContext)(v)
}

// ToV2 converts ContainerContext to session.ContainerSessionContext
// message structure.
func (x *ContainerContext) ToV2() *session.ContainerSessionContext {
	return (*session.ContainerSessionContext)(x)
}

// ApplyTo specifies which container the ContainerContext applies to.
//
// If id is nil, ContainerContext is applied to all containers.
func (x *ContainerContext) ApplyTo(id *cid.ID) {
	v2 := (*session.ContainerSessionContext)(x)
	var cidV2 *refs.ContainerID

	if id != nil {
		var c refs.ContainerID
		id.WriteToV2(&c)

		cidV2 = &c
	}

	v2.SetWildcard(id == nil)
	v2.SetContainerID(cidV2)
}

// ApplyToAllContainers is a helper function that conveniently
// applies ContainerContext to all containers.
func ApplyToAllContainers(c *ContainerContext) {
	c.ApplyTo(nil)
}

// Container returns identifier of the container
// to which the ContainerContext applies.
//
// Returns nil if ContainerContext is applied to
// all containers.
func (x *ContainerContext) Container() *cid.ID {
	v2 := (*session.ContainerSessionContext)(x)

	if v2.Wildcard() {
		return nil
	}

	cidV2 := v2.ContainerID()
	if cidV2 == nil {
		return nil
	}

	var cID cid.ID
	cID.ReadFromV2(*cidV2)

	return &cID
}

func (x *ContainerContext) forVerb(v session.ContainerSessionVerb) {
	(*session.ContainerSessionContext)(x).
		SetVerb(v)
}

func (x *ContainerContext) isForVerb(v session.ContainerSessionVerb) bool {
	return (*session.ContainerSessionContext)(x).
		Verb() == v
}

// ForPut binds the ContainerContext to
// PUT operation.
func (x *ContainerContext) ForPut() {
	x.forVerb(session.ContainerVerbPut)
}

// IsForPut checks if ContainerContext is bound to
// PUT operation.
func (x *ContainerContext) IsForPut() bool {
	return x.isForVerb(session.ContainerVerbPut)
}

// ForDelete binds the ContainerContext to
// DELETE operation.
func (x *ContainerContext) ForDelete() {
	x.forVerb(session.ContainerVerbDelete)
}

// IsForDelete checks if ContainerContext is bound to
// DELETE operation.
func (x *ContainerContext) IsForDelete() bool {
	return x.isForVerb(session.ContainerVerbDelete)
}

// ForSetEACL binds the ContainerContext to
// SETEACL operation.
func (x *ContainerContext) ForSetEACL() {
	x.forVerb(session.ContainerVerbSetEACL)
}

// IsForSetEACL checks if ContainerContext is bound to
// SETEACL operation.
func (x *ContainerContext) IsForSetEACL() bool {
	return x.isForVerb(session.ContainerVerbSetEACL)
}

// Marshal marshals ContainerContext into a protobuf binary form.
func (x *ContainerContext) Marshal() ([]byte, error) {
	return x.ToV2().StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of ContainerContext.
func (x *ContainerContext) Unmarshal(data []byte) error {
	return x.ToV2().Unmarshal(data)
}

// MarshalJSON encodes ContainerContext to protobuf JSON format.
func (x *ContainerContext) MarshalJSON() ([]byte, error) {
	return x.ToV2().MarshalJSON()
}

// UnmarshalJSON decodes ContainerContext from protobuf JSON format.
func (x *ContainerContext) UnmarshalJSON(data []byte) error {
	return x.ToV2().UnmarshalJSON(data)
}
