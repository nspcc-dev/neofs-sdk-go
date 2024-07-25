package bearer

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Token represents bearer token for object service operations.
//
// Token is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/acl.BearerToken
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
type Token struct {
	targetUser user.ID

	issuer user.ID

	eaclTableSet bool
	eaclTable    eacl.Table

	lifetimeSet   bool
	iat, nbf, exp uint64

	sigSet bool
	sig    refs.Signature
}

// reads Token from the acl.BearerToken message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field.
func (b *Token) readFromV2(m acl.BearerToken, checkFieldPresence bool) error {
	var err error

	body := m.GetBody()
	if checkFieldPresence && body == nil {
		return errors.New("missing token body")
	}

	eaclTable := body.GetEACL()
	if b.eaclTableSet = eaclTable != nil; b.eaclTableSet {
		if err = b.eaclTable.ReadFromV2(*eaclTable); err != nil {
			return fmt.Errorf("invalid eACL")
		}
	} else if checkFieldPresence {
		return errors.New("missing eACL table")
	}

	targetUser := body.GetOwnerID()
	if targetUser != nil {
		err = b.targetUser.ReadFromV2(*targetUser)
		if err != nil {
			return fmt.Errorf("invalid target user: %w", err)
		}
	} else {
		b.targetUser = user.ID{}
	}

	issuer := body.GetIssuer()
	if issuer != nil {
		err = b.issuer.ReadFromV2(*issuer)
		if err != nil {
			return fmt.Errorf("invalid issuer: %w", err)
		}
	} else {
		b.issuer = user.ID{}
	}

	lifetime := body.GetLifetime()
	if b.lifetimeSet = lifetime != nil; b.lifetimeSet {
		b.iat = lifetime.GetIat()
		b.nbf = lifetime.GetNbf()
		b.exp = lifetime.GetExp()
	} else if checkFieldPresence {
		return errors.New("missing token lifetime")
	}

	sig := m.GetSignature()
	if b.sigSet = sig != nil; sig != nil {
		b.sig = *sig
	} else if checkFieldPresence {
		return errors.New("missing body signature")
	}

	return nil
}

// ReadFromV2 reads Token from the acl.BearerToken message.
//
// See also WriteToV2.
func (b *Token) ReadFromV2(m acl.BearerToken) error {
	return b.readFromV2(m, true)
}

func (b Token) fillBody() *acl.BearerTokenBody {
	if !b.eaclTableSet && b.targetUser.IsZero() && !b.lifetimeSet && b.issuer.IsZero() {
		return nil
	}

	var body acl.BearerTokenBody

	if b.eaclTableSet {
		body.SetEACL(b.eaclTable.ToV2())
	}

	if !b.targetUser.IsZero() {
		var targetUser refs.OwnerID
		b.targetUser.WriteToV2(&targetUser)

		body.SetOwnerID(&targetUser)
	}

	if !b.issuer.IsZero() {
		var issuer refs.OwnerID
		b.issuer.WriteToV2(&issuer)

		body.SetIssuer(&issuer)
	}

	if b.lifetimeSet {
		var lifetime acl.TokenLifetime
		lifetime.SetIat(b.iat)
		lifetime.SetNbf(b.nbf)
		lifetime.SetExp(b.exp)

		body.SetLifetime(&lifetime)
	}

	return &body
}

func (b Token) signedData() []byte {
	return b.fillBody().StableMarshal(nil)
}

// WriteToV2 writes Token to the acl.BearerToken message.
// The message must not be nil.
//
// See also ReadFromV2.
func (b Token) WriteToV2(m *acl.BearerToken) {
	m.SetBody(b.fillBody())

	var sig *refs.Signature

	if b.sigSet {
		sig = &b.sig
	}

	m.SetSignature(sig)
}

// SetExp sets "exp" (expiration time) claim which identifies the
// expiration time (in NeoFS epochs) after which the Token MUST NOT be
// accepted for processing. The processing of the "exp" claim requires
// that the current epoch MUST be before or equal to the expiration epoch
// listed in the "exp" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also InvalidAt.
func (b *Token) SetExp(exp uint64) {
	b.exp = exp
	b.lifetimeSet = true
}

// SetNbf sets "nbf" (not before) claim which identifies the time (in
// NeoFS epochs) before which the Token MUST NOT be accepted for processing. The
// processing of the "nbf" claim requires that the current epoch MUST be
// after or equal to the not-before epoch listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also InvalidAt.
func (b *Token) SetNbf(nbf uint64) {
	b.nbf = nbf
	b.lifetimeSet = true
}

// SetIat sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the Token was issued. This claim can be used to determine
// the age of the Token.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also InvalidAt.
func (b *Token) SetIat(iat uint64) {
	b.iat = iat
	b.lifetimeSet = true
}

// InvalidAt asserts "exp", "nbf" and "iat" claims for the given epoch.
//
// Zero Container is invalid in any epoch.
//
// See also SetExp, SetNbf, SetIat.
func (b Token) InvalidAt(epoch uint64) bool {
	return !b.lifetimeSet || b.nbf > epoch || b.iat > epoch || b.exp < epoch
}

// SetEACLTable sets eacl.Table that replaces the one from the issuer's
// container. If table has specified container, bearer token can be used only
// for operations within this specific container. Otherwise, Token can be used
// within any issuer's container.
//
// SetEACLTable MUST be called if Token is going to be transmitted over
// NeoFS API V2 protocol.
//
// See also EACLTable, AssertContainer.
func (b *Token) SetEACLTable(table eacl.Table) {
	b.eaclTable = table
	b.eaclTableSet = true
}

// EACLTable returns extended ACL table set by SetEACLTable.
//
// Zero Token has zero eacl.Table.
func (b Token) EACLTable() eacl.Table {
	if b.eaclTableSet {
		return b.eaclTable
	}

	return eacl.Table{}
}

// AssertContainer checks if the token is valid within the given container.
//
// Note: cnr is assumed to refer to the issuer's container, otherwise the check
// is meaningless.
//
// Zero Token is valid in any container.
//
// See also SetEACLTable.
func (b Token) AssertContainer(cnr cid.ID) bool {
	if !b.eaclTableSet {
		return true
	}

	cnrTable := b.eaclTable.GetCID()
	return cnrTable.IsZero() || cnrTable == cnr
}

// ForUser specifies ID of the user who can use the Token for the operations
// within issuer's container(s).
//
// Optional: by default, any user has access to Token usage.
//
// See also AssertUser.
func (b *Token) ForUser(id user.ID) {
	b.targetUser = id
}

// AssertUser checks if the Token is issued to the given user.
//
// Zero Token is available to any user.
//
// See also ForUser.
func (b Token) AssertUser(id user.ID) bool {
	return b.targetUser.IsZero() || b.targetUser == id
}

// Sign calculates and writes signature of the [Token] data along with issuer ID
// using signer. Returns signature calculation errors.
//
// Sign MUST be called if [Token] is going to be transmitted over
// NeoFS API V2 protocol.
//
// Note that any [Token] mutation is likely to break the signature, so it is
// expected to be calculated as a final stage of Token formation.
//
// See also [Token.VerifySignature], [Token.Issuer], [Token.SignedData].
func (b *Token) Sign(signer user.Signer) error {
	b.SetIssuer(signer.UserID())

	var sig neofscrypto.Signature

	err := sig.Calculate(signer, b.signedData())
	if err != nil {
		return err
	}

	sig.WriteToV2(&b.sig)
	b.sigSet = true

	return nil
}

// SignedData returns actual payload to sign.
//
// See also [Token.Sign], [Token.UnmarshalSignedData].
func (b *Token) SignedData() []byte {
	return b.signedData()
}

// UnmarshalSignedData is a reverse op to [Token.SignedData].
func (b *Token) UnmarshalSignedData(data []byte) error {
	var body acl.BearerTokenBody
	err := body.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	var tok acl.BearerToken
	tok.SetBody(&body)
	return b.readFromV2(tok, false)
}

// VerifySignature checks if Token signature is presented and valid.
//
// Zero Token fails the check.
//
// See also Sign.
func (b Token) VerifySignature() bool {
	if !b.sigSet {
		return false
	}

	var sig neofscrypto.Signature

	// TODO: (#233) check owner<->key relation
	return sig.ReadFromV2(b.sig) == nil && sig.Verify(b.signedData())
}

// Marshal encodes Token into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also Unmarshal.
func (b Token) Marshal() []byte {
	var m acl.BearerToken
	b.WriteToV2(&m)

	return m.StableMarshal(nil)
}

// Unmarshal decodes NeoFS API protocol binary data into the Token
// (Protocol Buffers V3 with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (b *Token) Unmarshal(data []byte) error {
	var m acl.BearerToken

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return b.readFromV2(m, false)
}

// MarshalJSON encodes Token into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also UnmarshalJSON.
func (b Token) MarshalJSON() ([]byte, error) {
	var m acl.BearerToken
	b.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Token
// (Protocol Buffers V3 JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (b *Token) UnmarshalJSON(data []byte) error {
	var m acl.BearerToken

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return b.readFromV2(m, false)
}

// SigningKeyBytes returns issuer's public key in a binary format of
// NeoFS API protocol.
//
// Unsigned [Token] has empty key.
//
// The resulting slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Token.ResolveIssuer].
func (b Token) SigningKeyBytes() []byte {
	if b.sigSet {
		return b.sig.GetKey()
	}

	return nil
}

// SetIssuer sets NeoFS user ID of the Token issuer.
//
// See also [Token.Issuer], [Token.Sign].
func (b *Token) SetIssuer(usr user.ID) {
	b.issuer = usr
}

// Issuer returns NeoFS user ID of the explicitly set Token issuer. Zero value
// means unset issuer. In this case, [Token.ResolveIssuer] can be used to get ID
// resolved from signer's public key.
//
// See also [Token.SetIssuer], [Token.Sign].
func (b Token) Issuer() user.ID {
	return b.issuer
}

// ResolveIssuer works like [Token.Issuer] with fallback to the public key
// resolution when explicit issuer ID is unset. Returns zero [user.ID] when
// neither issuer is set nor key resolution succeeds.
//
// See also [Token.SigningKeyBytes], [Token.Sign].
func (b Token) ResolveIssuer() user.ID {
	if !b.issuer.IsZero() {
		return b.issuer
	}

	var usr user.ID
	binKey := b.SigningKeyBytes()

	if len(binKey) != 0 {
		if err := idFromKey(&usr, binKey); err != nil {
			usr = user.ID{}
		}
	}

	return usr
}

func idFromKey(id *user.ID, key []byte) error {
	var pk keys.PublicKey
	if err := pk.DecodeBytes(key); err != nil {
		return fmt.Errorf("decode owner failed: %w", err)
	}

	id.SetScriptHash(pk.GetScriptHash())
	return nil
}
