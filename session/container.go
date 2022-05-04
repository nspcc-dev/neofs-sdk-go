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
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Container represents token of the NeoFS Container session. A session is opened
// between any two sides of the system, and implements a mechanism for transferring
// the power of attorney of actions to another network member. The session has a
// limited validity period, and applies to a strictly defined set of operations.
// See methods for details.
//
// Container is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/session.Token
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
type Container struct {
	cnrSet bool

	lt session.TokenLifetime

	c session.ContainerSessionContext

	body session.TokenBody

	sig neofscrypto.Signature
}

// ReadFromV2 reads Container from the session.Token message.
//
// See also WriteToV2.
func (x *Container) ReadFromV2(m session.Token) error {
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

	c, ok := b.GetContext().(*session.ContainerSessionContext)
	if !ok {
		return fmt.Errorf("invalid context %T", b.GetContext())
	}

	cnr := c.ContainerID()
	x.cnrSet = !c.Wildcard()

	if x.cnrSet && cnr == nil {
		return errors.New("container is not specified with unset wildcard")
	}

	x.body = *b

	if c != nil {
		x.c = *c
	} else {
		x.c = session.ContainerSessionContext{}
	}

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

// WriteToV2 writes Container to the session.Token message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Container) WriteToV2(m *session.Token) {
	var sig refs.Signature
	x.sig.WriteToV2(&sig)

	m.SetBody(&x.body)
	m.SetSignature(&sig)
}

// Marshal encodes Container into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x Container) Marshal() []byte {
	var m session.Token
	x.WriteToV2(&m)

	return m.StableMarshal(nil)
}

// Unmarshal decodes NeoFS API protocol binary format into the Container
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *Container) Unmarshal(data []byte) error {
	var m session.Token

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return x.ReadFromV2(m)
}

// MarshalJSON encodes Container into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x Container) MarshalJSON() ([]byte, error) {
	var m session.Token
	x.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Container
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *Container) UnmarshalJSON(data []byte) error {
	var m session.Token

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return x.ReadFromV2(m)
}

// Sign calculates and writes signature of the Container data.
// Returns signature calculation errors.
//
// Zero Container is unsigned.
//
// Note that any Container mutation is likely to break the signature, so it is
// expected to be calculated as a final stage of Container formation.
//
// See also VerifySignature.
func (x *Container) Sign(key ecdsa.PrivateKey) error {
	var idUser user.ID
	user.IDFromKey(&idUser, key.PublicKey)

	var idUserV2 refs.OwnerID
	idUser.WriteToV2(&idUserV2)

	x.c.SetWildcard(!x.cnrSet)

	x.body.SetOwnerID(&idUserV2)
	x.body.SetLifetime(&x.lt)
	x.body.SetContext(&x.c)

	data := x.body.StableMarshal(nil)

	return x.sig.Calculate(neofsecdsa.Signer(key), data)
}

// VerifySignature checks if Container signature is presented and valid.
//
// Zero Container fails the check.
//
// See also Sign.
func (x Container) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	data := x.body.StableMarshal(nil)

	return x.sig.Verify(data)
}

// ApplyOnlyTo limits session scope to a given author container.
//
// See also AppliedTo.
func (x *Container) ApplyOnlyTo(cnr cid.ID) {
	var cnrv2 refs.ContainerID
	cnr.WriteToV2(&cnrv2)

	x.c.SetContainerID(&cnrv2)
	x.cnrSet = true
}

// AppliedTo checks if the session is propagated to the given container.
//
// Zero Container is applied to all author's containers.
//
// See also ApplyOnlyTo.
func (x Container) AppliedTo(cnr cid.ID) bool {
	if !x.cnrSet {
		return true
	}

	var cnr2 cid.ID

	if err := cnr2.ReadFromV2(*x.c.ContainerID()); err != nil {
		// NPE and error must never happen
		panic(fmt.Sprintf("unexpected error from cid.ReadFromV2: %v", err))
	}

	return cnr2.Equals(cnr)
}

// ContainerVerb enumerates container operations.
type ContainerVerb int8

const (
	_ ContainerVerb = iota

	VerbContainerPut     // Put rpc
	VerbContainerDelete  // Delete rpc
	VerbContainerSetEACL // SetExtendedACL rpc
)

// ForVerb specifies the container operation of the session scope. Each
// Container is related to the single operation.
//
// See also AssertVerb.
func (x *Container) ForVerb(verb ContainerVerb) {
	x.c.SetVerb(session.ContainerSessionVerb(verb))
}

// AssertVerb checks if Container relates to the given container operation.
//
// Zero Container relates to zero (unspecified) verb.
//
// See also ForVerb.
func (x Container) AssertVerb(verb ContainerVerb) bool {
	return verb == ContainerVerb(x.c.Verb())
}

// SetExp sets "exp" (expiration time) claim which identifies the expiration time
// (in NeoFS epochs) on or after which the Container MUST NOT be accepted for
// processing.  The processing of the "exp" claim requires that the current
// epoch MUST be before the expiration epoch listed in the "exp" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also ExpiredAt.
func (x *Container) SetExp(exp uint64) {
	x.lt.SetExp(exp)
}

// ExpiredAt asserts "exp" claim.
//
// Zero Container is expired in any epoch.
//
// See also SetExp.
func (x Container) ExpiredAt(epoch uint64) bool {
	return x.lt.GetExp() <= epoch
}

// SetNbf sets "nbf" (not before) claim which identifies the time (in NeoFS
// epochs) before which the Container MUST NOT be accepted for processing.
// The processing of the "nbf" claim requires that the current date/time MUST be
// after or equal to the not-before date/time listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also InvalidAt.
func (x *Container) SetNbf(nbf uint64) {
	x.lt.SetNbf(nbf)
}

// SetIat sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the Container was issued. This claim can be used to
// determine the age of the Container.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also InvalidAt.
func (x *Container) SetIat(iat uint64) {
	x.lt.SetIat(iat)
}

// InvalidAt asserts "exp", "nbf" and "iat" claims.
//
// Zero Container is invalid in any epoch.
//
// See also SetExp, SetNbf, SetIat.
func (x Container) InvalidAt(epoch uint64) bool {
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
func (x *Container) SetID(id uuid.UUID) {
	x.body.SetID(id[:])
}

// ID returns a unique identifier for the session.
//
// Zero Container has empty UUID (all zeros, see uuid.Nil) which is legitimate
// but most likely not suitable.
//
// See also SetID.
func (x Container) ID() uuid.UUID {
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
func (x *Container) SetAuthKey(key neofscrypto.PublicKey) {
	bKey := make([]byte, key.MaxEncodedSize())
	bKey = bKey[:key.Encode(bKey)]

	x.body.SetSessionKey(bKey)
}

// AssertAuthKey asserts public key bound to the session.
//
// Zero Container fails the check.
//
// See also SetAuthKey.
func (x Container) AssertAuthKey(key neofscrypto.PublicKey) bool {
	bKey := make([]byte, key.MaxEncodedSize())
	bKey = bKey[:key.Encode(bKey)]

	return bytes.Equal(bKey, x.body.GetSessionKey())
}
