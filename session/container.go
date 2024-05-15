package session

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Container represents token of the NeoFS Container session. A session is opened
// between any two sides of the system, and implements a mechanism for transferring
// the power of attorney of actions to another network member. The session has a
// limited validity period, and applies to a strictly defined set of operations.
// See methods for details.
//
// Container is mutually compatible with [session.SessionToken] message. See
// [Container.ReadFromV2] / [Container.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type Container struct {
	commonData

	verb ContainerVerb

	cnrSet bool
	cnr    cid.ID
}

// CopyTo writes deep copy of the [Container] to dst.
func (x Container) CopyTo(dst *Container) {
	x.commonData.copyTo(&dst.commonData)

	dst.verb = x.verb

	dst.cnrSet = x.cnrSet
	contID := x.cnr
	dst.cnr = contID
}

func (x *Container) readFromV2(m *session.SessionToken, checkFieldPresence bool) error {
	err := x.commonData.readFromV2(m, checkFieldPresence)
	if err != nil {
		return err
	}

	var ctx *session.ContainerSessionContext
	if c := m.GetBody().GetContext(); c != nil {
		cc, ok := c.(*session.SessionToken_Body_Container)
		if !ok {
			return errors.New("wrong context field")
		}
		ctx = cc.Container
	} else if checkFieldPresence {
		return errors.New("missing session context")
	} else {
		x.cnrSet = false
		x.verb = 0
		return nil
	}

	x.cnrSet = !ctx.GetWildcard()
	cnr := ctx.GetContainerId()

	if x.cnrSet {
		if cnr != nil {
			err := x.cnr.ReadFromV2(cnr)
			if err != nil {
				return fmt.Errorf("invalid context: invalid container ID: %w", err)
			}
		} else {
			return errors.New("invalid context: missing container or wildcard flag")
		}
	} else if cnr != nil {
		return errors.New("invalid context: container conflicts with wildcard flag")
	}

	x.verb = ContainerVerb(ctx.GetVerb())

	return nil
}

// ReadFromV2 reads Container from the [session.SessionToken] message. Returns
// an error if the message is malformed according to the NeoFS API V2 protocol.
// The message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Container.WriteToV2].
func (x *Container) ReadFromV2(m *session.SessionToken) error {
	return x.readFromV2(m, true)
}

func (x Container) fillContext() *session.SessionToken_Body_Container {
	c := session.SessionToken_Body_Container{
		Container: &session.ContainerSessionContext{
			Verb:     session.ContainerSessionContext_Verb(x.verb),
			Wildcard: !x.cnrSet,
		},
	}
	if x.cnrSet {
		c.Container.ContainerId = new(refs.ContainerID)
		x.cnr.WriteToV2(c.Container.ContainerId)
	}
	return &c
}

// WriteToV2 writes Container to the [session.SessionToken] message of the NeoFS
// API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Container.ReadFromV2].
func (x Container) WriteToV2(m *session.SessionToken) {
	x.writeToV2(m)
	m.Body.Context = x.fillContext()
}

// Marshal encodes Container into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also [Container.Unmarshal].
func (x Container) Marshal() []byte {
	var m session.SessionToken
	x.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Container. Returns
// an error describing a format violation of the specified fields. Unmarshal
// does not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Container.Marshal].
func (x *Container) Unmarshal(data []byte) error {
	var m session.SessionToken
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return x.readFromV2(&m, false)
}

// MarshalJSON encodes Container into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Container.UnmarshalJSON].
func (x Container) MarshalJSON() ([]byte, error) {
	var m session.SessionToken
	x.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Container
// (Protocol Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Container.MarshalJSON].
func (x *Container) UnmarshalJSON(data []byte) error {
	var m session.SessionToken
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}
	return x.readFromV2(&m, false)
}

// Sign calculates and writes signature of the [Container] data along with
// issuer ID using signer. Returns signature calculation errors.
//
// Zero [Container] is unsigned.
//
// Note that any [Container] mutation is likely to break the signature, so it is
// expected to be calculated as a final stage of [Container] formation.
//
// See also [Container.VerifySignature], [Container.SignedData].
func (x *Container) Sign(signer user.Signer) error {
	x.issuer = signer.UserID()
	x.issuerSet = true
	return x.SetSignature(signer)
}

// SetSignature allows to sign Container like [Container.Sign] but without
// issuer setting.
func (x *Container) SetSignature(signer neofscrypto.Signer) error {
	err := x.sig.Calculate(signer, x.SignedData())
	if err != nil {
		return err
	}
	x.sigSet = true
	return nil
}

// SignedData returns signed data of the Container.
//
// [Container.SetIssuer] must be called before.
//
// See also [Container.Sign], [Container.UnmarshalSignedData].
func (x Container) SignedData() []byte {
	body := x.fillBody()
	body.Context = x.fillContext()
	b := make([]byte, body.MarshaledSize())
	body.MarshalStable(b)
	return b
}

// UnmarshalSignedData is a reverse op to [Container.SignedData].
func (x *Container) UnmarshalSignedData(data []byte) error {
	var body session.SessionToken_Body
	err := proto.Unmarshal(data, &body)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return x.readFromV2(&session.SessionToken{Body: &body}, false)
}

// VerifySignature checks if Container signature is presented and valid.
//
// Zero Container fails the check.
//
// See also [Container.Sign], [Container.SetSignature].
func (x Container) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	return x.sigSet && x.sig.Verify(x.SignedData())
}

// ApplyOnlyTo limits session scope to a given author container.
//
// See also [Container.AppliedTo].
func (x *Container) ApplyOnlyTo(cnr cid.ID) {
	x.cnr = cnr
	x.cnrSet = true
}

// AppliedTo checks if the session is propagated to the given container.
//
// Zero Container is applied to all author's containers.
//
// See also [Container.ApplyOnlyTo].
func (x Container) AppliedTo(cnr cid.ID) bool {
	return !x.cnrSet || x.cnr == cnr
}

// ContainerVerb enumerates container operations.
type ContainerVerb uint8

const (
	_ ContainerVerb = iota

	VerbContainerPut     // Put rpc
	VerbContainerDelete  // Delete rpc
	VerbContainerSetEACL // SetExtendedACL rpc
)

// ForVerb specifies the container operation of the session scope. Each
// Container is related to the single operation.
//
// See also [Container.AssertVerb].
func (x *Container) ForVerb(verb ContainerVerb) {
	x.verb = verb
}

// AssertVerb checks if Container relates to the given container operation.
//
// Zero Container relates to zero (unspecified) verb.
//
// See also [Container.ForVerb].
func (x Container) AssertVerb(verb ContainerVerb) bool {
	return x.verb == verb
}

// IssuedBy checks if Container session is issued by the given user.
//
// See also [Container.Issuer].
func IssuedBy(cnr Container, id user.ID) bool {
	return cnr.Issuer() == id
}

// VerifySessionDataSignature verifies signature of the session data. In practice,
// the method is used to authenticate an operation with session data.
func (x Container) VerifySessionDataSignature(data, signature []byte) bool {
	var pubKey neofsecdsa.PublicKeyRFC6979
	return pubKey.Decode(x.authKey) == nil && pubKey.Verify(data, signature)
}
