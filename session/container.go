package session

import (
	"errors"
	"fmt"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Container represents token of the NeoFS Container session. A session is opened
// between any two sides of the system, and implements a mechanism for transferring
// the power of attorney of actions to another network member. The session has a
// limited validity period, and applies to a strictly defined set of operations.
// See methods for details.
//
// Container is mutually compatible with [protosession.SessionToken] message.
// See [Container.FromProtoMessage] / [Container.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type Container struct {
	commonData

	verb ContainerVerb

	cnr cid.ID
}

// CopyTo writes deep copy of the [Container] to dst.
func (x Container) CopyTo(dst *Container) {
	x.commonData.copyTo(&dst.commonData)

	dst.verb = x.verb
	dst.cnr = x.cnr
}

// readContext is a contextReader needed for commonData methods.
func (x *Container) readContext(c any, checkFieldPresence bool) error {
	cc, ok := c.(*protosession.SessionToken_Body_Container)
	if !ok || cc == nil {
		return fmt.Errorf("invalid context %T", c)
	}
	cCnr := cc.Container

	cnr := cCnr.GetContainerId()

	if !cCnr.Wildcard {
		if cnr != nil {
			err := x.cnr.FromProtoMessage(cnr)
			if err != nil {
				return fmt.Errorf("invalid container ID: %w", err)
			}
		} else if checkFieldPresence {
			return errors.New("missing container or wildcard flag")
		} else {
			x.cnr = cid.ID{}
		}
	} else if cnr != nil {
		return errors.New("container conflicts with wildcard flag")
	}

	verb := cCnr.GetVerb()
	if verb < 0 {
		return fmt.Errorf("negative verb %d", verb)
	}
	x.verb = ContainerVerb(cCnr.Verb)

	return nil
}

func (x *Container) fromProtoMessage(m *protosession.SessionToken, checkFieldPresence bool) error {
	return x.commonData.fromProtoMessage(m, checkFieldPresence, x.readContext)
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [Container.ProtoMessage].
func (x *Container) FromProtoMessage(m *protosession.SessionToken) error {
	return x.fromProtoMessage(m, true)
}

func (x Container) writeContext(m *protosession.SessionToken_Body) {
	c := &protosession.ContainerSessionContext{
		Verb:     protosession.ContainerSessionContext_Verb(x.verb),
		Wildcard: x.cnr.IsZero(),
	}

	if !c.Wildcard {
		c.ContainerId = x.cnr.ProtoMessage()
	}

	m.Context = &protosession.SessionToken_Body_Container{Container: c}
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [Container.FromProtoMessage].
func (x Container) ProtoMessage() *protosession.SessionToken {
	return x.protoMessage(x.writeContext)
}

// Marshal encodes Container into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x Container) Marshal() []byte {
	return x.marshal(x.writeContext)
}

// Unmarshal decodes NeoFS API protocol binary format into the Container
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *Container) Unmarshal(data []byte) error {
	return x.unmarshal(data, x.readContext)
}

// MarshalJSON encodes Container into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x Container) MarshalJSON() ([]byte, error) {
	return x.marshalJSON(x.writeContext)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Container
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *Container) UnmarshalJSON(data []byte) error {
	return x.unmarshalJSON(data, x.readContext)
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
	if x.issuer.IsZero() {
		return user.ErrZeroID
	}
	return x.SignIssued(signer)
}

// SignIssued allows to sign Container like [Container.Sign] but without
// issuer setting.
func (x *Container) SignIssued(signer neofscrypto.Signer) error {
	return x.sign(signer, x.writeContext)
}

// SignedData returns actual payload to sign.
//
// Using this method require to set issuer via [Container.SetIssuer] before SignedData call.
//
// See also [Container.Sign], [Container.UnmarshalSignedData].
func (x *Container) SignedData() []byte {
	return x.signedData(x.writeContext)
}

// UnmarshalSignedData is a reverse op to [Container.SignedData].
func (x *Container) UnmarshalSignedData(data []byte) error {
	var body protosession.SessionToken_Body
	err := neofsproto.UnmarshalMessage(data, &body)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	return x.fromProtoMessage(&protosession.SessionToken{Body: &body}, false)
}

// VerifySignature checks if Container signature is presented and valid.
//
// Zero Container fails the check.
//
// See also Sign.
func (x Container) VerifySignature() bool {
	return x.verifySignature(x.writeContext)
}

// ApplyOnlyTo limits session scope to a given author container.
//
// See also AppliedTo.
func (x *Container) ApplyOnlyTo(cnr cid.ID) {
	x.cnr = cnr
}

// AppliedTo checks if the session is propagated to the given container.
//
// Zero Container is applied to all author's containers.
//
// See also ApplyOnlyTo.
func (x Container) AppliedTo(cnr cid.ID) bool {
	return x.cnr.IsZero() || x.cnr == cnr
}

// ContainerVerb enumerates container operations.
type ContainerVerb int32

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
	x.verb = verb
}

// AssertVerb checks if Container relates to the given container operation.
//
// Zero Container relates to zero (unspecified) verb.
//
// See also ForVerb.
func (x Container) AssertVerb(verb ContainerVerb) bool {
	return x.verb == verb
}

// IssuedBy checks if Container session is issued by the given user.
//
// See also Container.Issuer.
func IssuedBy(cnr Container, id user.ID) bool {
	return cnr.Issuer() == id
}

// VerifySessionDataSignature verifies signature of the session data. In practice,
// the method is used to authenticate an operation with session data.
func (x Container) VerifySessionDataSignature(data, signature []byte) bool {
	sig := neofscrypto.NewSignatureFromRawKey(neofscrypto.ECDSA_DETERMINISTIC_SHA256, x.authKey, signature)
	return sig.Verify(data)
}
