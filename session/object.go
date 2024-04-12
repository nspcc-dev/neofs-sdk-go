package session

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
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

	objs []oid.ID
}

// CopyTo writes deep copy of the [Container] to dst.
func (x Object) CopyTo(dst *Object) {
	x.commonData.copyTo(&dst.commonData)

	dst.verb = x.verb

	dst.cnrSet = x.cnrSet
	contID := x.cnr
	dst.cnr = contID

	if objs := x.objs; objs != nil {
		dst.objs = make([]oid.ID, len(x.objs))
		copy(dst.objs, x.objs)
	} else {
		dst.objs = nil
	}
}

func (x *Object) readContext(c session.TokenContext, checkFieldPresence bool) error {
	cObj, ok := c.(*session.ObjectSessionContext)
	if !ok || cObj == nil {
		return fmt.Errorf("invalid context %T", c)
	}

	var err error

	cnr := cObj.GetContainer()
	if x.cnrSet = cnr != nil; x.cnrSet {
		err := x.cnr.ReadFromV2(*cnr)
		if err != nil {
			return fmt.Errorf("invalid container ID: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing target container")
	}

	objs := cObj.GetObjects()
	if objs != nil {
		x.objs = make([]oid.ID, len(objs))

		for i := range objs {
			err = x.objs[i].ReadFromV2(objs[i])
			if err != nil {
				return fmt.Errorf("invalid target object: %w", err)
			}
		}
	} else {
		x.objs = nil
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

	if x.cnrSet || len(x.objs) > 0 {
		var cnr *refs.ContainerID

		if x.cnrSet {
			cnr = new(refs.ContainerID)
			x.cnr.WriteToV2(cnr)
		}

		var objs []refs.ObjectID

		if x.objs != nil {
			objs = make([]refs.ObjectID, len(x.objs))

			for i := range x.objs {
				x.objs[i].WriteToV2(&objs[i])
			}
		}

		c.SetTarget(cnr, objs...)
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
	x.issuerSet = true
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
	var body session.TokenBody
	err := body.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	var tok session.Token
	tok.SetBody(&body)
	return x.readFromV2(tok, false)
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
		if x.objs[i].Equals(obj) {
			return true
		}
	}

	return false
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
