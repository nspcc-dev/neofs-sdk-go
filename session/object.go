package session

import (
	"errors"
	"fmt"
	"slices"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Object represents token of the NeoFS Object session. A session is opened
// between any two sides of the system, and implements a mechanism for transferring
// the power of attorney of actions to another network member. The session has a
// limited validity period, and applies to a strictly defined set of operations.
// See methods for details.
//
// Object is mutually compatible with [protosession.SessionToken] message. See
// [Object.FromProtoMessage] / [Object.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type Object struct {
	commonData

	verb ObjectVerb

	cnr cid.ID

	objs []oid.ID
}

// CopyTo writes deep copy of the [Container] to dst.
func (x Object) CopyTo(dst *Object) {
	x.commonData.copyTo(&dst.commonData)

	dst.verb = x.verb
	dst.cnr = x.cnr
	dst.objs = slices.Clone(x.objs)
}

func (x *Object) readContext(c any, checkFieldPresence bool) error {
	cc, ok := c.(*protosession.SessionToken_Body_Object)
	if !ok || cc == nil {
		return fmt.Errorf("invalid context %T", c)
	}
	cObj := cc.Object

	var err error

	cnr := cObj.Target.GetContainer()
	if cnr != nil {
		err := x.cnr.FromProtoMessage(cnr)
		if err != nil {
			return fmt.Errorf("invalid container ID: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing target container")
	} else {
		x.cnr = cid.ID{}
	}

	objs := cObj.Target.GetObjects()
	if objs != nil {
		x.objs = make([]oid.ID, len(objs))

		for i := range objs {
			if objs[i] == nil {
				return fmt.Errorf("nil target object #%d", i)
			}
			err = x.objs[i].FromProtoMessage(objs[i])
			if err != nil {
				return fmt.Errorf("invalid target object: %w", err)
			}
		}
	} else {
		x.objs = nil
	}

	verb := cObj.GetVerb()
	if verb < 0 {
		return fmt.Errorf("negative verb %d", verb)
	}
	x.verb = ObjectVerb(verb)

	return nil
}

func (x *Object) fromProtoMessage(m *protosession.SessionToken, checkFieldPresence bool) error {
	return x.fromProtoMessageWithVersion(m, checkFieldPresence, nil)
}

func (x *Object) fromProtoMessageWithVersion(m *protosession.SessionToken, checkFieldPresence bool, version *version.Version) error {
	return x.commonData.fromProtoMessageWithVersion(m, checkFieldPresence, x.readContext, version)
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [Object.ProtoMessage].
func (x *Object) FromProtoMessage(m *protosession.SessionToken) error {
	return x.fromProtoMessage(m, true)
}

// FromProtoMessageWithVersion validates m according to the NeoFS API protocol and restores
// x from it, taking into account the object version for backward compatibility.
//
// See also [Object.ProtoMessage].
func (x *Object) FromProtoMessageWithVersion(m *protosession.SessionToken, version *version.Version) error {
	return x.fromProtoMessageWithVersion(m, true, version)
}

func (x Object) writeContext(m *protosession.SessionToken_Body) {
	c := &protosession.ObjectSessionContext{
		Verb: protosession.ObjectSessionContext_Verb(x.verb),
	}

	if !x.cnr.IsZero() || len(x.objs) > 0 {
		c.Target = new(protosession.ObjectSessionContext_Target)

		if !x.cnr.IsZero() {
			c.Target.Container = x.cnr.ProtoMessage()
		}

		if x.objs != nil {
			c.Target.Objects = make([]*refs.ObjectID, len(x.objs))

			for i := range x.objs {
				c.Target.Objects[i] = x.objs[i].ProtoMessage()
			}
		}
	}

	m.Context = &protosession.SessionToken_Body_Object{Object: c}
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [Object.FromProtoMessage].
func (x Object) ProtoMessage() *protosession.SessionToken {
	return x.protoMessage(x.writeContext)
}

// Marshal encodes Object into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x Object) Marshal() []byte {
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

// Sign calculates and writes signature of the [Object] data along with issuer
// ID using signer. Returns signature calculation errors.
//
// Zero [Object] is unsigned.
//
// Note that any [Object] mutation is likely to break the signature, so it is
// expected to be calculated as a final stage of [Object] formation.
//
// See also [Object.VerifySignature], [Object.SignedData].
func (x *Object) Sign(signer user.Signer) error {
	x.issuer = signer.UserID()
	if x.issuer.IsZero() {
		return user.ErrZeroID
	}
	return x.SetSignature(signer)
}

// SetSignature allows to sign Object like [Object.Sign] but without issuer
// setting.
func (x *Object) SetSignature(signer neofscrypto.Signer) error {
	return x.sign(signer, x.writeContext)
}

// SignedData returns actual payload to sign.
//
// See also [Object.Sign], [Object.UnmarshalSignedData].
func (x *Object) SignedData() []byte {
	return x.signedData(x.writeContext)
}

// UnmarshalSignedData is a reverse op to [Object.SignedData].
func (x *Object) UnmarshalSignedData(data []byte) error {
	var body protosession.SessionToken_Body
	err := neofsproto.UnmarshalMessage(data, &body)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	return x.fromProtoMessage(&protosession.SessionToken{Body: &body}, false)
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
}

// AssertContainer checks if Object session bound to a given container.
//
// Zero Object isn't bound to any container which is incorrect according to
// NeoFS API protocol.
//
// See also BindContainer.
func (x Object) AssertContainer(cnr cid.ID) bool {
	return x.cnr == cnr
}

// LimitByObjects limits session scope to the given objects from the container
// to which Object session is bound.
//
// Argument MUST NOT be mutated, make a copy first.
//
// See also AssertObject.
func (x *Object) LimitByObjects(objs ...oid.ID) {
	x.objs = objs
}

// AssertObject checks if Object session is applied to a given object.
//
// Zero Object is applied to all objects in the container.
//
// See also LimitByObjects.
func (x Object) AssertObject(obj oid.ID) bool {
	if len(x.objs) == 0 {
		return true
	}

	for i := range x.objs {
		if x.objs[i] == obj {
			return true
		}
	}

	return false
}

// ObjectVerb enumerates object operations.
type ObjectVerb int32

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
