package bearer

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Token represents bearer token for object service operations.
//
// Token is mutually compatible with [protoacl.BearerToken] message. See
// [Token.FromProtoMessage] / [Token.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type Token struct {
	targetUser user.ID

	issuer user.ID

	eaclTableSet bool
	eaclTable    eacl.Table

	iat, nbf, exp uint64

	sigSet bool
	sig    neofscrypto.Signature
}

// reads Token from the acl.BearerToken message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field.
func (b *Token) fromProtoMessage(m *protoacl.BearerToken, checkFieldPresence bool) error {
	var err error

	body := m.Body
	if checkFieldPresence && body == nil {
		return errors.New("missing token body")
	}

	eaclTable := body.GetEaclTable()
	if b.eaclTableSet = eaclTable != nil; b.eaclTableSet {
		if err = b.eaclTable.FromProtoMessage(eaclTable); err != nil {
			return fmt.Errorf("invalid eACL: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing eACL table")
	}

	targetUser := body.GetOwnerId()
	if targetUser != nil {
		err = b.targetUser.FromProtoMessage(targetUser)
		if err != nil {
			return fmt.Errorf("invalid target user: %w", err)
		}
	} else {
		b.targetUser = user.ID{}
	}

	issuer := body.GetIssuer()
	if issuer != nil {
		err = b.issuer.FromProtoMessage(issuer)
		if err != nil {
			return fmt.Errorf("invalid issuer: %w", err)
		}
	} else {
		b.issuer = user.ID{}
	}

	lifetime := body.GetLifetime()
	if checkFieldPresence && lifetime == nil {
		return errors.New("missing token lifetime")
	}

	if b.sigSet = m.Signature != nil; b.sigSet {
		if err = b.sig.FromProtoMessage(m.Signature); err != nil {
			return fmt.Errorf("invalid body signature: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing body signature")
	}

	b.iat = lifetime.GetIat()
	b.nbf = lifetime.GetNbf()
	b.exp = lifetime.GetExp()

	return nil
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// b from it.
//
// See also [Token.ProtoMessage].
func (b *Token) FromProtoMessage(m *protoacl.BearerToken) error {
	return b.fromProtoMessage(m, true)
}

func (b Token) fillBody() *protoacl.BearerToken_Body {
	lifetimeSet := b.iat != 0 || b.nbf != 0 || b.exp != 0
	if !b.eaclTableSet && b.targetUser.IsZero() && !lifetimeSet && b.issuer.IsZero() {
		return nil
	}

	var body protoacl.BearerToken_Body

	if b.eaclTableSet {
		body.EaclTable = b.eaclTable.ProtoMessage()
	}

	if !b.targetUser.IsZero() {
		body.OwnerId = b.targetUser.ProtoMessage()
	}

	if !b.issuer.IsZero() {
		body.Issuer = b.issuer.ProtoMessage()
	}

	if lifetimeSet {
		body.Lifetime = &protoacl.BearerToken_Body_TokenLifetime{Exp: b.exp, Nbf: b.nbf, Iat: b.iat}
	}

	return &body
}

func (b Token) signedData() []byte {
	return neofsproto.MarshalMessage(b.fillBody())
}

// ProtoMessage converts sg into message to transmit using the NeoFS API
// protocol.
//
// See also [Token.FromProtoMessage].
func (b Token) ProtoMessage() *protoacl.BearerToken {
	m := &protoacl.BearerToken{
		Body: b.fillBody(),
	}
	if b.sigSet {
		m.Signature = b.sig.ProtoMessage()
	}
	return m
}

// SetExp sets "exp" (expiration time) claim which identifies the
// expiration time (in NeoFS epochs) after which the Token MUST NOT be
// accepted for processing. The processing of the "exp" claim requires
// that the current epoch MUST be before or equal to the expiration epoch
// listed in the "exp" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also [Token.ValidAt].
func (b *Token) SetExp(exp uint64) {
	b.exp = exp
}

// Exp returns "exp" (expiration time) claim.
func (b Token) Exp() uint64 {
	return b.exp
}

// SetNbf sets "nbf" (not before) claim which identifies the time (in
// NeoFS epochs) before which the Token MUST NOT be accepted for processing. The
// processing of the "nbf" claim requires that the current epoch MUST be
// after or equal to the not-before epoch listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also [Token.ValidAt].
func (b *Token) SetNbf(nbf uint64) {
	b.nbf = nbf
}

// Nbf returns "nbf" (not before) claim.
func (b Token) Nbf() uint64 {
	return b.nbf
}

// SetIat sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the Token was issued. This claim can be used to determine
// the age of the Token.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also [Token.ValidAt].
func (b *Token) SetIat(iat uint64) {
	b.iat = iat
}

// Iat returns "iat" (issued at) claim.
func (b Token) Iat() uint64 {
	return b.iat
}

// ValidAt checks whether the Token is still valid at the given epoch according
// to its lifetime claims.
func (b Token) ValidAt(epoch uint64) bool {
	return b.Nbf() <= epoch && b.Iat() <= epoch && b.Exp() >= epoch
}

// InvalidAt asserts "exp", "nbf" and "iat" claims for the given epoch.
//
// Zero Container is invalid in any epoch.
//
// See also SetExp, SetNbf, SetIat.
// Deprecated: use inverse [Token.ValidAt] instead.
func (b Token) InvalidAt(epoch uint64) bool { return !b.ValidAt(epoch) }

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

	err := b.sig.Calculate(signer, b.signedData())
	if err == nil {
		b.sigSet = true
	}
	return err
}

// SignedData returns actual payload to sign.
//
// See also [Token.Sign], [Token.UnmarshalSignedData].
func (b *Token) SignedData() []byte {
	return b.signedData()
}

// UnmarshalSignedData is a reverse op to [Token.SignedData].
func (b *Token) UnmarshalSignedData(data []byte) error {
	var body protoacl.BearerToken_Body
	err := neofsproto.UnmarshalMessage(data, &body)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	return b.fromProtoMessage(&protoacl.BearerToken{Body: &body}, false)
}

// AttachSignature attaches given signature to the Token. Use [Token.SignedData]
// for calculation. If signature instance itself is not needed, use
// [Token.Sign].
func (b *Token) AttachSignature(sig neofscrypto.Signature) {
	b.sig, b.sigSet = sig, true
}

// Signature returns Token signature. If the signature is missing, false is
// returned. Use [Token.SignedData] for verification. If signature instance
// itself is not needed, use [Token.VerifySignature].
func (b Token) Signature() (neofscrypto.Signature, bool) {
	return b.sig, b.sigSet
}

// VerifySignature checks if Token signature is presented and valid.
//
// Zero Token fails the check.
//
// See also Sign.
func (b Token) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	return b.sigSet && b.sig.Verify(b.signedData())
}

// Marshal encodes Token into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also Unmarshal.
func (b Token) Marshal() []byte {
	return neofsproto.Marshal(b)
}

// Unmarshal decodes NeoFS API protocol binary data into the Token
// (Protocol Buffers V3 with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (b *Token) Unmarshal(data []byte) error {
	return neofsproto.UnmarshalOptional(data, b, (*Token).fromProtoMessage)
}

// MarshalJSON encodes Token into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also UnmarshalJSON.
func (b Token) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(b)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Token
// (Protocol Buffers V3 JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (b *Token) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSONOptional(data, b, (*Token).fromProtoMessage)
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
// Deprecated: use [Token.Signature] instead.
func (b Token) SigningKeyBytes() []byte {
	if sig, ok := b.Signature(); ok {
		return sig.PublicKeyBytes()
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
	if b.sigSet {
		if err := idFromKey(&usr, b.sig.PublicKeyBytes()); err != nil {
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
