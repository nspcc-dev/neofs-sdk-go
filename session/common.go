package session

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/google/uuid"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

type commonData struct {
	idSet bool
	id    uuid.UUID

	issuer user.ID

	iat, nbf, exp uint64

	authKey []byte

	sigSet bool
	sig    neofscrypto.Signature
}

type contextReader func(any, bool) error

func (x commonData) copyTo(dst *commonData) {
	dst.idSet = x.idSet
	dst.id = x.id

	dst.issuer = x.issuer

	dst.iat = x.iat
	dst.nbf = x.nbf
	dst.exp = x.exp
	dst.authKey = bytes.Clone(x.authKey)
	dst.sigSet = x.sigSet
	dst.sig = neofscrypto.NewSignatureFromRawKey(x.sig.Scheme(), bytes.Clone(x.sig.PublicKeyBytes()), bytes.Clone(x.sig.Value()))
}

// reads commonData and custom context from the session.Token message.
// If checkFieldPresence is set, returns an error on absence of any protocol-required
// field. Verifies format of any presented field according to NeoFS API V2 protocol.
// Calls contextReader if session context is set. Passes checkFieldPresence into contextReader.
func (x *commonData) fromProtoMessage(m *protosession.SessionToken, checkFieldPresence bool, r contextReader) error {
	return x.fromProtoMessageWithVersion(m, checkFieldPresence, r, nil)
}

// minLifetimeVersion is the minimum version of NeoFS API protocol that supports session tokens lifetime.
var minLifetimeVersion = version.New(2, 12)

// isVersionWithMandatoryLifetime checks whether the given version is supported by
// this package and has mandatory lifetime fields.
func isVersionWithMandatoryLifetime(version *version.Version) bool {
	return version == nil ||
		version.Major() > minLifetimeVersion.Major() ||
		(version.Major() == minLifetimeVersion.Major() && version.Minor() >= minLifetimeVersion.Minor())
}

// reads commonData and custom context from the session.Token message considering object version.
// If checkFieldPresence is set, returns an error on absence of any protocol-required
// field. Verifies format of any presented field according to NeoFS API V2 protocol.
// If version is not nil, some field validations may be skipped for older versions.
// Calls contextReader if session context is set. Passes checkFieldPresence into contextReader.
func (x *commonData) fromProtoMessageWithVersion(m *protosession.SessionToken, checkFieldPresence bool, r contextReader, version *version.Version) error {
	var err error

	body := m.GetBody()
	if checkFieldPresence && body == nil {
		return errors.New("missing token body")
	}

	binID := body.GetId()
	if x.idSet = len(binID) > 0; x.idSet {
		err = x.id.UnmarshalBinary(binID)
		if err != nil {
			return fmt.Errorf("invalid session ID: %w", err)
		} else if ver := x.id.Version(); ver != 4 {
			return fmt.Errorf("invalid session ID: wrong UUID version %d, expected 4", ver)
		}
	} else if checkFieldPresence {
		return errors.New("missing session ID")
	} else {
		x.id = uuid.Nil
	}

	issuer := body.GetOwnerId()
	if issuer != nil {
		err = x.issuer.FromProtoMessage(issuer)
		if err != nil {
			return fmt.Errorf("invalid session issuer: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing session issuer")
	} else {
		x.issuer = user.ID{}
	}

	lifetime := body.GetLifetime()
	if checkFieldPresence && lifetime == nil && isVersionWithMandatoryLifetime(version) {
		return errors.New("missing token lifetime")
	}

	x.authKey = body.GetSessionKey()
	if checkFieldPresence && len(x.authKey) == 0 {
		return errors.New("missing session public key")
	}

	c := body.GetContext()
	if c != nil {
		err = r(c, checkFieldPresence)
		if err != nil {
			return fmt.Errorf("invalid context: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing session context")
	}

	if x.sigSet = m.Signature != nil; x.sigSet {
		if err = x.sig.FromProtoMessage(m.Signature); err != nil {
			return fmt.Errorf("invalid body signature: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing body signature")
	}

	x.iat = lifetime.GetIat()
	x.nbf = lifetime.GetNbf()
	x.exp = lifetime.GetExp()

	return nil
}

type contextWriter func(body *protosession.SessionToken_Body)

func (x commonData) fillBody(w contextWriter) *protosession.SessionToken_Body {
	var body protosession.SessionToken_Body

	if x.id != uuid.Nil {
		body.Id = x.id[:]
	}

	if !x.issuer.IsZero() {
		body.OwnerId = x.issuer.ProtoMessage()
	}

	if x.iat != 0 || x.nbf != 0 || x.exp != 0 {
		body.Lifetime = &protosession.SessionToken_Body_TokenLifetime{
			Exp: x.exp,
			Nbf: x.nbf,
			Iat: x.iat,
		}
	}

	body.SessionKey = x.authKey

	w(&body)

	return &body
}

func (x commonData) protoMessage(w contextWriter) *protosession.SessionToken {
	m := &protosession.SessionToken{
		Body: x.fillBody(w),
	}

	if x.sigSet {
		m.Signature = x.sig.ProtoMessage()
	}

	return m
}

func (x commonData) signedData(w contextWriter) []byte {
	return neofsproto.MarshalMessage(x.fillBody(w))
}

func (x *commonData) sign(signer neofscrypto.Signer, w contextWriter) error {
	x.sigSet = true
	return x.sig.Calculate(signer, x.signedData(w))
}

func (x commonData) verifySignature(w contextWriter) bool {
	// TODO: (#233) check owner<->key relation
	return x.sigSet && x.sig.Verify(x.signedData(w))
}

func (x commonData) marshal(w contextWriter) []byte {
	return neofsproto.MarshalMessage(x.protoMessage(w))
}

func (x *commonData) unmarshal(data []byte, r contextReader) error {
	var m protosession.SessionToken

	err := neofsproto.UnmarshalMessage(data, &m)
	if err != nil {
		return err
	}

	return x.fromProtoMessage(&m, false, r)
}

func (x commonData) marshalJSON(w contextWriter) ([]byte, error) {
	return neofsproto.MarshalMessageJSON(x.protoMessage(w))
}

func (x *commonData) unmarshalJSON(data []byte, r contextReader) error {
	var m protosession.SessionToken

	err := neofsproto.UnmarshalMessageJSON(data, &m)
	if err != nil {
		return err
	}

	return x.fromProtoMessage(&m, false, r)
}

// SetExp sets "exp" (expiration time) claim which identifies the expiration
// time (in NeoFS epochs) after which the session MUST NOT be accepted for
// processing. The processing of the "exp" claim requires that the current
// epoch MUST be before or equal to the expiration epoch listed in the "exp"
// claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also ExpiredAt.
func (x *commonData) SetExp(exp uint64) {
	x.exp = exp
}

// Exp returns "exp" (expiration time) claim.
func (x commonData) Exp() uint64 {
	return x.exp
}

// SetNbf sets "nbf" (not before) claim which identifies the time (in NeoFS
// epochs) before which the session MUST NOT be accepted for processing.
// The processing of the "nbf" claim requires that the current date/time MUST be
// after or equal to the not-before date/time listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also ValidAt.
func (x *commonData) SetNbf(nbf uint64) {
	x.nbf = nbf
}

// Nbf returns "nbf" (not before) claim.
func (x commonData) Nbf() uint64 {
	return x.nbf
}

// SetIat sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the session was issued. This claim can be used to
// determine the age of the session.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also ValidAt.
func (x *commonData) SetIat(iat uint64) {
	x.iat = iat
}

// Iat returns "iat" (issued at) claim.
func (x commonData) Iat() uint64 {
	return x.iat
}

func (x commonData) expiredAt(epoch uint64) bool {
	return x.Exp() < epoch
}

// ValidAt checks whether the token is still valid at the given epoch according
// to its lifetime claims.
func (x commonData) ValidAt(epoch uint64) bool {
	return x.Nbf() <= epoch && x.Iat() <= epoch && x.Exp() >= epoch
}

// InvalidAt asserts "exp", "nbf" and "iat" claims.
//
// Zero session is invalid in any epoch.
//
// See also SetExp, SetNbf, SetIat.
// Deprecated: use inverse [Token.ValidAt] instead.
func (x commonData) InvalidAt(epoch uint64) bool { return !x.ValidAt(epoch) }

// SetID sets a unique identifier for the session. The identifier value MUST be
// assigned in a manner that ensures that there is a negligible probability
// that the same value will be accidentally assigned to a different session.
//
// ID format MUST be UUID version 4 (random). uuid.New can be used to generate
// a new ID. See https://datatracker.ietf.org/doc/html/rfc4122 and
// github.com/google/uuid package docs for details.
//
// See also ID.
func (x *commonData) SetID(id uuid.UUID) {
	x.id, x.idSet = id, true
}

// ID returns a unique identifier for the session.
//
// Zero session has empty UUID (all zeros, see uuid.Nil) which is legitimate
// but most likely not suitable.
//
// See also SetID.
func (x commonData) ID() uuid.UUID {
	return x.id
}

// SetAuthKey public key corresponding to the private key bound to the session.
//
// See also AssertAuthKey.
func (x *commonData) SetAuthKey(key neofscrypto.PublicKey) {
	x.authKey = neofscrypto.PublicKeyBytes(key)
}

// SetIssuer allows to set issuer before Sign call.
// Using this method is not required when Sign is used (issuer will be derived from the signer automatically).
// When using it please ensure that the token is signed with the same signer as the issuer passed here.
func (x *commonData) SetIssuer(id user.ID) {
	x.issuer = id
}

// AssertAuthKey asserts public key bound to the session.
//
// Zero session fails the check.
//
// See also SetAuthKey.
func (x commonData) AssertAuthKey(key neofscrypto.PublicKey) bool {
	return bytes.Equal(neofscrypto.PublicKeyBytes(key), x.authKey)
}

// Issuer returns user ID of the session issuer.
//
// Makes sense only for signed session instances. For unsigned instances,
// Issuer returns zero user.ID.
//
// See also Sign.
func (x commonData) Issuer() user.ID {
	return x.issuer
}

// IssuerPublicKeyBytes returns binary-encoded public key of the session issuer.
//
// IssuerPublicKeyBytes MUST NOT be called before ProtoMessage or Sign methods.
// Deprecated: use Signature method.
func (x *commonData) IssuerPublicKeyBytes() []byte {
	if sig, ok := x.Signature(); ok {
		return sig.PublicKeyBytes()
	}
	return nil
}

// AttachSignature attaches given signature to the token. Use SignedData method
// for calculation. If signature instance itself is not needed, use Sign method.
func (x *commonData) AttachSignature(sig neofscrypto.Signature) {
	x.sig, x.sigSet = sig, true
}

// Signature returns token signature. If the signature is missing, false is
// returned. Use SignedData method for verification. If signature instance
// itself is not needed, use VerifySignature method.
func (x commonData) Signature() (neofscrypto.Signature, bool) {
	return x.sig, x.sigSet
}
