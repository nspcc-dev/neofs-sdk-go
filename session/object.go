package session

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// Object represents token of the NeoFS Object session. A session is opened
// between any two sides of the system, and implements a mechanism for transferring
// the power of attorney of actions to another network member. The session has a
// limited validity period, and applies to a strictly defined set of operations.
// See methods for details.
//
// Object is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/session.Token
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
type Object struct {
	commonData

	verb ObjectVerb

	cnrSet bool
	cnr    cid.ID

	objSet bool
	obj    oid.ID
}

func (x *Object) readContext(c session.TokenContext, checkFieldPresence bool) error {
	cObj, ok := c.(*session.ObjectSessionContext)
	if !ok || cObj == nil {
		return fmt.Errorf("invalid context %T", c)
	}

	addr := cObj.GetAddress()
	if checkFieldPresence && addr == nil {
		return errors.New("missing object address")
	}

	var err error

	cnr := addr.GetContainerID()
	if x.cnrSet = cnr != nil; x.cnrSet {
		err := x.cnr.ReadFromV2(*cnr)
		if err != nil {
			return fmt.Errorf("invalid container ID: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing container in object address")
	}

	obj := addr.GetObjectID()
	if x.objSet = obj != nil; x.objSet {
		err = x.obj.ReadFromV2(*obj)
		if err != nil {
			return fmt.Errorf("invalid object ID: %w", err)
		}
	}

	x.verb = ObjectVerb(cObj.GetVerb())

	return nil
}

func (x *Object) readFromV2(m session.Token, checkFieldPresence bool) error {
	return x.commonData.readFromV2(m, checkFieldPresence, x.readContext)
}

// ReadFromV2 reads Object from the session.Token message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *Object) ReadFromV2(m session.Token) error {
	return x.readFromV2(m, true)
}

func (x Object) writeContext() session.TokenContext {
	var c session.ObjectSessionContext
	c.SetVerb(session.ObjectSessionVerb(x.verb))

	if x.cnrSet || x.objSet {
		var addr refs.Address

		if x.cnrSet {
			var cnr refs.ContainerID
			x.cnr.WriteToV2(&cnr)

			addr.SetContainerID(&cnr)
		}

		if x.objSet {
			var obj refs.ObjectID
			x.obj.WriteToV2(&obj)

			addr.SetObjectID(&obj)
		}

		c.SetAddress(&addr)
	}

	return &c
}

// WriteToV2 writes Object to the session.Token message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Object) WriteToV2(m *session.Token) {
	x.writeToV2(m, x.writeContext)
}

// Marshal encodes Object into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x Object) Marshal() []byte {
	var m session.Token
	x.WriteToV2(&m)

	return x.marshal(x.writeContext)
}

// Unmarshal decodes NeoFS API protocol binary format into the Object
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *Object) Unmarshal(data []byte) error {
	return x.unmarshal(data, x.readContext)
}

// MarshalJSON encodes Object into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x Object) MarshalJSON() ([]byte, error) {
	return x.marshalJSON(x.writeContext)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Object
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *Object) UnmarshalJSON(data []byte) error {
	return x.unmarshalJSON(data, x.readContext)
}

// Sign calculates and writes signature of the Object data.
// Returns signature calculation errors.
//
// Zero Object is unsigned.
//
// Note that any Object mutation is likely to break the signature, so it is
// expected to be calculated as a final stage of Object formation.
//
// See also VerifySignature.
func (x *Object) Sign(key ecdsa.PrivateKey) error {
	return x.sign(key, x.writeContext)
}

// VerifySignature checks if Object signature is presented and valid.
//
// Zero Object fails the check.
//
// See also Sign.
func (x Object) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	return x.verifySignature(x.writeContext)
}

// BindContainer binds the Object session to a given container. Each session
// MUST be bound to exactly one container.
//
// See also AssertContainer.
func (x *Object) BindContainer(cnr cid.ID) {
	x.cnr = cnr
	x.cnrSet = true
}

// AssertContainer checks if Object session bound to a given container.
//
// Zero Object isn't bound to any container which is incorrect according to
// NeoFS API protocol.
//
// See also BindContainer.
func (x Object) AssertContainer(cnr cid.ID) bool {
	return x.cnr.Equals(cnr)
}

// LimitByObject limits session scope to a given object from the container
// to which Object session is bound.
//
// See also AssertObject.
func (x *Object) LimitByObject(obj oid.ID) {
	x.obj = obj
	x.objSet = true
}

// AssertObject checks if Object session is applied to a given object.
//
// Zero Object is applied to all objects in the container.
//
// See also LimitByObject.
func (x Object) AssertObject(obj oid.ID) bool {
	return !x.objSet || x.obj.Equals(obj)
}

// ObjectVerb enumerates object operations.
type ObjectVerb int8

const (
	_ ObjectVerb = iota

	VerbObjectPut       // Put rpc
	VerbObjectGet       // Get rpc
	VerbObjectHead      // Head rpc
	VerbObjectSearch    // Search rpc
	VerbObjectDelete    // Delete rpc
	VerbObjectRange     // GetRange rpc
	VerbObjectRangeHash // GetRangeHash rpc
)

// ForVerb specifies the object operation of the session scope. Each
// Object is related to the single operation.
//
// See also AssertVerb.
func (x *Object) ForVerb(verb ObjectVerb) {
	x.verb = verb
}

// AssertVerb checks if Object relates to one of the given object operations.
//
// Zero Object relates to zero (unspecified) verb.
//
// See also ForVerb.
func (x Object) AssertVerb(verbs ...ObjectVerb) bool {
	for i := range verbs {
		if verbs[i] == x.verb {
			return true
		}
	}

	return false
}

// ExpiredAt asserts "exp" claim.
//
// Zero Object is expired in any epoch.
//
// See also SetExp.
func (x Object) ExpiredAt(epoch uint64) bool {
	return x.expiredAt(epoch)
}
