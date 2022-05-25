package session

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
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
	lt session.TokenLifetime

	obj refs.Address

	c session.ObjectSessionContext

	body session.TokenBody

	sig neofscrypto.Signature
}

// ReadFromV2 reads Object from the session.Token message.
//
// See also WriteToV2.
func (x *Object) ReadFromV2(m session.Token) error {
	b := m.GetBody()
	if b == nil {
		return errors.New("missing body")
	}

	bID := b.GetID()
	var id uuid.UUID

	err := id.UnmarshalBinary(bID)
	if err != nil {
		return fmt.Errorf("invalid binary ID: %w", err)
	} else if ver := id.Version(); ver != 4 {
		return fmt.Errorf("invalid UUID version %s", ver)
	}

	c, ok := b.GetContext().(*session.ObjectSessionContext)
	if !ok || c == nil {
		return fmt.Errorf("invalid context %T", b.GetContext())
	}

	x.body = *b

	x.c = *c

	obj := c.GetAddress()

	cnr := obj.GetContainerID()
	if cnr == nil {
		return errors.New("missing bound container")
	}

	x.obj = *obj

	lt := b.GetLifetime()
	if lt != nil {
		x.lt = *lt
	} else {
		x.lt = session.TokenLifetime{}
	}

	sig := m.GetSignature()
	if sig != nil {
		x.sig.ReadFromV2(*sig)
	} else {
		x.sig = neofscrypto.Signature{}
	}

	return nil
}

// WriteToV2 writes Object to the session.Token message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Object) WriteToV2(m *session.Token) {
	var sig refs.Signature
	x.sig.WriteToV2(&sig)

	m.SetBody(&x.body)
	m.SetSignature(&sig)
}

// Marshal encodes Object into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x Object) Marshal() []byte {
	var m session.Token
	x.WriteToV2(&m)

	data, err := m.StableMarshal(nil)
	if err != nil {
		panic(fmt.Sprintf("unexpected error from Token.StableMarshal: %v", err))
	}

	return data
}

// Unmarshal decodes NeoFS API protocol binary format into the Object
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *Object) Unmarshal(data []byte) error {
	var m session.Token

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return x.ReadFromV2(m)
}

// MarshalJSON encodes Object into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x Object) MarshalJSON() ([]byte, error) {
	var m session.Token
	x.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Object
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *Object) UnmarshalJSON(data []byte) error {
	var m session.Token

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return x.ReadFromV2(m)
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
	var idUser user.ID
	user.IDFromKey(&idUser, key.PublicKey)

	var idUserV2 refs.OwnerID
	idUser.WriteToV2(&idUserV2)

	x.c.SetAddress(&x.obj)

	x.body.SetOwnerID(&idUserV2)
	x.body.SetLifetime(&x.lt)
	x.body.SetContext(&x.c)

	data, err := x.body.StableMarshal(nil)
	if err != nil {
		panic(fmt.Sprintf("unexpected error from Token.StableMarshal: %v", err))
	}

	return x.sig.Calculate(neofsecdsa.Signer(key), data)
}

// VerifySignature checks if Object signature is presented and valid.
//
// Zero Object fails the check.
//
// See also Sign.
func (x Object) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	data, err := x.body.StableMarshal(nil)
	if err != nil {
		panic(fmt.Sprintf("unexpected error from Token.StableMarshal: %v", err))
	}

	return x.sig.Verify(data)
}

// BindContainer binds the Object session to a given container. Each session
// MUST be bound to exactly one container.
//
// See also AssertContainer.
func (x *Object) BindContainer(cnr cid.ID) {
	var cnrV2 refs.ContainerID
	cnr.WriteToV2(&cnrV2)

	x.obj.SetContainerID(&cnrV2)
}

// AssertContainer checks if Object session bound to a given container.
//
// Zero Object isn't bound to any container which is incorrect according to
// NeoFS API protocol.
//
// See also BindContainer.
func (x Object) AssertContainer(cnr cid.ID) bool {
	cnrV2 := x.obj.GetContainerID()
	if cnrV2 == nil {
		return false
	}

	var cnr2 cid.ID

	err := cnr2.ReadFromV2(*cnrV2)

	return err == nil && cnr2.Equals(cnr)
}

// LimitByObject limits session scope to a given object from the container
// to which Object session is bound.
//
// See also AssertObject.
func (x *Object) LimitByObject(obj oid.ID) {
	var objV2 refs.ObjectID
	obj.WriteToV2(&objV2)

	x.obj.SetObjectID(&objV2)
}

// AssertObject checks if Object session is applied to a given object.
//
// Zero Object is applied to all objects in the container.
//
// See also LimitByObject.
func (x Object) AssertObject(obj oid.ID) bool {
	objV2 := x.obj.GetObjectID()
	if objV2 == nil {
		return true
	}

	var obj2 oid.ID

	err := obj2.ReadFromV2(*objV2)

	return err == nil && obj2.Equals(obj)
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
	x.c.SetVerb(session.ObjectSessionVerb(verb))
}

// AssertVerb checks if Object relates to one of the given object operations.
//
// Zero Object relates to zero (unspecified) verb.
//
// See also ForVerb.
func (x Object) AssertVerb(verbs ...ObjectVerb) bool {
	verb := ObjectVerb(x.c.GetVerb())

	for i := range verbs {
		if verbs[i] == verb {
			return true
		}
	}

	return false
}

// SetExp sets "exp" (expiration time) claim which identifies the expiration time
// (in NeoFS epochs) on or after which the Object MUST NOT be accepted for
// processing.  The processing of the "exp" claim requires that the current
// epoch MUST be before the expiration epoch listed in the "exp" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also ExpiredAt.
func (x *Object) SetExp(exp uint64) {
	x.lt.SetExp(exp)
}

// ExpiredAt asserts "exp" claim.
//
// Zero Object is expired in any epoch.
//
// See also SetExp.
func (x Object) ExpiredAt(epoch uint64) bool {
	return x.lt.GetExp() <= epoch
}

// SetNbf sets "nbf" (not before) claim which identifies the time (in NeoFS
// epochs) before which the Object MUST NOT be accepted for processing.
// The processing of the "nbf" claim requires that the current date/time MUST be
// after or equal to the not-before date/time listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also InvalidAt.
func (x *Object) SetNbf(nbf uint64) {
	x.lt.SetNbf(nbf)
}

// SetIat sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the Object was issued. This claim can be used to
// determine the age of the Object.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also InvalidAt.
func (x *Object) SetIat(iat uint64) {
	x.lt.SetIat(iat)
}

// InvalidAt asserts "exp", "nbf" and "iat" claims.
//
// Zero Object is invalid in any epoch.
//
// See also SetExp, SetNbf, SetIat.
func (x Object) InvalidAt(epoch uint64) bool {
	return x.lt.GetNbf() > epoch || x.lt.GetIat() > epoch || x.ExpiredAt(epoch)
}

// SetID sets a unique identifier for the session. The identifier value MUST be
// assigned in a manner that ensures that there is a negligible probability
// that the same value will be accidentally assigned to a different session.
//
// ID format MUST be UUID version 4 (random). uuid.New can be used to generate
// a new ID. See https://datatracker.ietf.org/doc/html/rfc4122 and
// github.com/google/uuid package docs for details.
//
// See also ID.
func (x *Object) SetID(id uuid.UUID) {
	x.body.SetID(id[:])
}

// ID returns a unique identifier for the session.
//
// Zero Object has empty UUID (all zeros, see uuid.Nil) which is legitimate
// but most likely not suitable.
//
// See also SetID.
func (x Object) ID() uuid.UUID {
	data := x.body.GetID()
	if data == nil {
		return uuid.Nil
	}

	var id uuid.UUID

	err := id.UnmarshalBinary(x.body.GetID())
	if err != nil {
		panic(fmt.Sprintf("unexpected error from UUID.UnmarshalBinary: %v", err))
	}

	return id
}

// SetAuthKey public key corresponding to the private key bound to the session.
//
// See also AssertAuthKey.
func (x *Object) SetAuthKey(key neofscrypto.PublicKey) {
	bKey := make([]byte, key.MaxEncodedSize())
	bKey = bKey[:key.Encode(bKey)]

	x.body.SetSessionKey(bKey)
}

// AssertAuthKey asserts public key bound to the session.
//
// Zero Object fails the check.
//
// See also SetAuthKey.
func (x Object) AssertAuthKey(key neofscrypto.PublicKey) bool {
	bKey := make([]byte, key.MaxEncodedSize())
	bKey = bKey[:key.Encode(bKey)]

	return bytes.Equal(bKey, x.body.GetSessionKey())
}
