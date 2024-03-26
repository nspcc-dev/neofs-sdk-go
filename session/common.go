package session

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

type commonData struct {
	idSet bool
	id    uuid.UUID

	issuerSet bool
	issuer    user.ID

	lifetimeSet   bool
	iat, nbf, exp uint64

	authKey []byte

	sigSet bool
	sig    refs.Signature
}

type contextReader func(session.TokenContext, bool) error

func (x commonData) copyTo(dst *commonData) {
	dst.idSet = x.idSet
	copy(dst.id[:], x.id[:])

	dst.issuerSet = x.issuerSet
	iss := x.issuer
	dst.issuer = iss

	dst.lifetimeSet = x.lifetimeSet
	dst.iat = x.iat
	dst.nbf = x.nbf
	dst.exp = x.exp
	dst.authKey = bytes.Clone(x.authKey)
	dst.sigSet = x.sigSet
	dst.sig.SetKey(bytes.Clone(x.sig.GetKey()))
	dst.sig.SetScheme(x.sig.GetScheme())
	dst.sig.SetSign(bytes.Clone(x.sig.GetSign()))
}

// reads commonData and custom context from the session.Token message.
// If checkFieldPresence is set, returns an error on absence of any protocol-required
// field. Verifies format of any presented field according to NeoFS API V2 protocol.
// Calls contextReader if session context is set. Passes checkFieldPresence into contextReader.
func (x *commonData) readFromV2(m session.Token, checkFieldPresence bool, r contextReader) error {
	var err error

	body := m.GetBody()
	if checkFieldPresence && body == nil {
		return errors.New("missing token body")
	}

	binID := body.GetID()
	if x.idSet = len(binID) > 0; x.idSet {
		err = x.id.UnmarshalBinary(binID)
		if err != nil {
			return fmt.Errorf("invalid session ID: %w", err)
		} else if ver := x.id.Version(); ver != 4 {
			return fmt.Errorf("invalid session UUID version %d", ver)
		}
	} else if checkFieldPresence {
		return errors.New("missing session ID")
	}

	issuer := body.GetOwnerID()
	if x.issuerSet = issuer != nil; x.issuerSet {
		err = x.issuer.ReadFromV2(*issuer)
		if err != nil {
			return fmt.Errorf("invalid session issuer: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing session issuer")
	}

	lifetime := body.GetLifetime()
	if x.lifetimeSet = lifetime != nil; x.lifetimeSet {
		x.iat = lifetime.GetIat()
		x.nbf = lifetime.GetNbf()
		x.exp = lifetime.GetExp()
	} else if checkFieldPresence {
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

	sig := m.GetSignature()
	if x.sigSet = sig != nil; sig != nil {
		x.sig = *sig
	} else if checkFieldPresence {
		return errors.New("missing body signature")
	}

	return nil
}

type contextWriter func() session.TokenContext

func (x commonData) fillBody(w contextWriter) *session.TokenBody {
	var body session.TokenBody

	if x.idSet {
		binID, err := x.id.MarshalBinary()
		if err != nil {
			panic(fmt.Sprintf("unexpected error from UUID.MarshalBinary: %v", err))
		}

		body.SetID(binID)
	}

	if x.issuerSet {
		var issuer refs.OwnerID
		x.issuer.WriteToV2(&issuer)

		body.SetOwnerID(&issuer)
	}

	if x.lifetimeSet {
		var lifetime session.TokenLifetime
		lifetime.SetIat(x.iat)
		lifetime.SetNbf(x.nbf)
		lifetime.SetExp(x.exp)

		body.SetLifetime(&lifetime)
	}

	body.SetSessionKey(x.authKey)

	body.SetContext(w())

	return &body
}

func (x commonData) writeToV2(m *session.Token, w contextWriter) {
	body := x.fillBody(w)

	m.SetBody(body)

	var sig *refs.Signature

	if x.sigSet {
		sig = &x.sig
	}

	m.SetSignature(sig)
}

func (x commonData) signedData(w contextWriter) []byte {
	return x.fillBody(w).StableMarshal(nil)
}

func (x *commonData) sign(signer neofscrypto.Signer, w contextWriter) error {
	var sig neofscrypto.Signature

	err := sig.Calculate(signer, x.signedData(w))
	if err != nil {
		return err
	}

	sig.WriteToV2(&x.sig)
	x.sigSet = true

	return nil
}

func (x commonData) verifySignature(w contextWriter) bool {
	if !x.sigSet {
		return false
	}

	var sig neofscrypto.Signature

	// TODO: (#233) check owner<->key relation
	return sig.ReadFromV2(x.sig) == nil && sig.Verify(x.signedData(w))
}

func (x commonData) marshal(w contextWriter) []byte {
	var m session.Token
	x.writeToV2(&m, w)

	return m.StableMarshal(nil)
}

func (x *commonData) unmarshal(data []byte, r contextReader) error {
	var m session.Token

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return x.readFromV2(m, false, r)
}

func (x commonData) marshalJSON(w contextWriter) ([]byte, error) {
	var m session.Token
	x.writeToV2(&m, w)

	return m.MarshalJSON()
}

func (x *commonData) unmarshalJSON(data []byte, r contextReader) error {
	var m session.Token

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return x.readFromV2(m, false, r)
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
	x.lifetimeSet = true
}

// SetNbf sets "nbf" (not before) claim which identifies the time (in NeoFS
// epochs) before which the session MUST NOT be accepted for processing.
// The processing of the "nbf" claim requires that the current date/time MUST be
// after or equal to the not-before date/time listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also InvalidAt.
func (x *commonData) SetNbf(nbf uint64) {
	x.nbf = nbf
	x.lifetimeSet = true
}

// SetIat sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the session was issued. This claim can be used to
// determine the age of the session.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also InvalidAt.
func (x *commonData) SetIat(iat uint64) {
	x.iat = iat
	x.lifetimeSet = true
}

func (x commonData) expiredAt(epoch uint64) bool {
	return !x.lifetimeSet || x.exp < epoch
}

// InvalidAt asserts "exp", "nbf" and "iat" claims.
//
// Zero session is invalid in any epoch.
//
// See also SetExp, SetNbf, SetIat.
func (x commonData) InvalidAt(epoch uint64) bool {
	return x.expiredAt(epoch) || x.nbf > epoch || x.iat > epoch
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
func (x *commonData) SetID(id uuid.UUID) {
	x.id = id
	x.idSet = true
}

// ID returns a unique identifier for the session.
//
// Zero session has empty UUID (all zeros, see uuid.Nil) which is legitimate
// but most likely not suitable.
//
// See also SetID.
func (x commonData) ID() uuid.UUID {
	if x.idSet {
		return x.id
	}

	return uuid.Nil
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
	x.issuerSet = true
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
	if x.issuerSet {
		return x.issuer
	}

	return user.ID{}
}

// IssuerPublicKeyBytes returns binary-encoded public key of the session issuer.
//
// IssuerPublicKeyBytes MUST NOT be called before ReadFromV2 or Sign methods.
func (x *commonData) IssuerPublicKeyBytes() []byte {
	if x.sigSet {
		return x.sig.GetKey()
	}

	return nil
}
