package session

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Object represents token of the NeoFS Object session. A session is opened
// between any two sides of the system, and implements a mechanism for transferring
// the power of attorney of actions to another network member. The session has a
// limited validity period, and applies to a strictly defined set of operations.
// See methods for details.
//
// Object is mutually compatible with [session.Token] message. See
// [Object.ReadFromV2] / [Object.WriteToV2] methods.
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

func (x *Object) readFromV2(m *session.SessionToken, checkFieldPresence bool) error {
	err := x.commonData.readFromV2(m, checkFieldPresence)
	if err != nil {
		return err
	}

	var ctx *session.ObjectSessionContext
	if c := m.GetBody().GetContext(); c != nil {
		cc, ok := c.(*session.SessionToken_Body_Object)
		if !ok {
			return errors.New("wrong context field")
		}
		ctx = cc.Object
	} else if checkFieldPresence {
		return errors.New("missing session context")
	} else {
		x.cnrSet = false
		x.verb = 0
		x.objs = nil
		return nil
	}

	cnr := ctx.GetTarget().GetContainer()
	if x.cnrSet = cnr != nil; x.cnrSet {
		err := x.cnr.ReadFromV2(cnr)
		if err != nil {
			return fmt.Errorf("invalid context: invalid target container: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("invalid context: missing target container")
	}

	objs := ctx.GetTarget().GetObjects()
	if objs != nil {
		x.objs = make([]oid.ID, len(objs))

		for i := range objs {
			err = x.objs[i].ReadFromV2(objs[i])
			if err != nil {
				return fmt.Errorf("invalid context: invalid target object #%d: %w", i, err)
			}
		}
	} else {
		x.objs = nil
	}

	x.verb = ObjectVerb(ctx.GetVerb())

	return nil
}

// ReadFromV2 reads Object from the [session.SessionToken] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Object.WriteToV2].
func (x *Object) ReadFromV2(m *session.SessionToken) error {
	return x.readFromV2(m, true)
}

func (x Object) fillContext() *session.SessionToken_Body_Object {
	c := session.SessionToken_Body_Object{
		Object: &session.ObjectSessionContext{
			Verb: session.ObjectSessionContext_Verb(x.verb),
		},
	}
	if x.cnrSet {
		c.Object.Target = &session.ObjectSessionContext_Target{
			Container: new(refs.ContainerID),
		}
		x.cnr.WriteToV2(c.Object.Target.Container)
	}
	if x.objs != nil {
		if c.Object.Target == nil {
			c.Object.Target = new(session.ObjectSessionContext_Target)
		}
		c.Object.Target.Objects = make([]*refs.ObjectID, len(x.objs))
		for i := range x.objs {
			c.Object.Target.Objects[i] = new(refs.ObjectID)
			x.objs[i].WriteToV2(c.Object.Target.Objects[i])
		}
	}
	return &c
}

// WriteToV2 writes Object to the [session.SessionToken] message of the NeoFS
// API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Object.ReadFromV2].
func (x Object) WriteToV2(m *session.SessionToken) {
	x.writeToV2(m)
	m.Body.Context = x.fillContext()
}

// Marshal encodes Object into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also [Object.Unmarshal].
func (x Object) Marshal() []byte {
	var m session.SessionToken
	x.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Object. Returns an
// error describing a format violation of the specified fields. Unmarshal does
// not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Object.Marshal].
func (x *Object) Unmarshal(data []byte) error {
	var m session.SessionToken
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return x.readFromV2(&m, false)
}

// MarshalJSON encodes Object into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Object.UnmarshalJSON].
func (x Object) MarshalJSON() ([]byte, error) {
	var m session.SessionToken
	x.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Object (Protocol
// Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Object.MarshalJSON].
func (x *Object) UnmarshalJSON(data []byte) error {
	var m session.SessionToken
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}
	return x.readFromV2(&m, false)
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
	err := x.sig.Calculate(signer, x.SignedData())
	if err != nil {
		return err
	}
	x.sigSet = true
	return nil
}

// SignedData returns signed data of the Object.
//
// [Object.SetIssuer] must be called before.
//
// See also [Object.Sign], [Object.UnmarshalSignedData].
func (x Object) SignedData() []byte {
	body := x.fillBody()
	body.Context = x.fillContext()
	b := make([]byte, body.MarshaledSize())
	body.MarshalStable(b)
	return b
}

// UnmarshalSignedData is a reverse op to [Object.SignedData].
func (x *Object) UnmarshalSignedData(data []byte) error {
	var body session.SessionToken_Body
	err := proto.Unmarshal(data, &body)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return x.readFromV2(&session.SessionToken{Body: &body}, false)
}

// VerifySignature checks if Object signature is presented and valid.
//
// Zero Object fails the check.
//
// See also [Object.Sign], [Object.SetSignature].
func (x Object) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	return x.sigSet && x.sig.Verify(x.SignedData())
}

// BindContainer binds the Object session to a given container. Each session
// MUST be bound to exactly one container.
//
// See also [Object.AssertContainer].
func (x *Object) BindContainer(cnr cid.ID) {
	x.cnr = cnr
	x.cnrSet = true
}

// AssertContainer checks if Object session bound to a given container.
//
// Zero Object isn't bound to any container which is incorrect according to
// NeoFS API protocol.
//
// See also [Object.BindContainer].
func (x Object) AssertContainer(cnr cid.ID) bool {
	return x.cnrSet && x.cnr == cnr
}

// LimitByObjects limits session scope to the given objects from the container
// to which Object session is bound.
//
// Argument MUST NOT be mutated, make a copy first.
//
// See also [Object.AssertObject].
func (x *Object) LimitByObjects(objs []oid.ID) {
	x.objs = objs
}

// AssertObject checks if Object session is applied to a given object.
//
// Zero Object is applied to all objects in the container.
//
// See also [Object.LimitByObjects].
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
// See also [Object.AssertVerb].
func (x *Object) ForVerb(verb ObjectVerb) {
	x.verb = verb
}

// AssertVerb checks if Object relates to one of the given object operations.
//
// Zero Object relates to zero (unspecified) verb.
//
// See also [Object.ForVerb].
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
// See also [Object.SetExp].
func (x Object) ExpiredAt(epoch uint64) bool {
	return x.expiredAt(epoch)
}
