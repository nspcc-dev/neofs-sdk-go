package bearer

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Token represents bearer token for object service operations.
//
// Token is mutually compatible with [acl.BearerToken] message. See
// [Token.ReadFromV2] / [Token.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type Token struct {
	targetUserSet bool
	targetUser    user.ID

	issuerSet bool
	issuer    user.ID

	eaclTableSet bool
	eaclTable    eacl.Table

	lifetimeSet   bool
	iat, nbf, exp uint64

	sigSet bool
	sig    neofscrypto.Signature
}

func (b *Token) readFromV2(m *acl.BearerToken, checkFieldPresence bool) error {
	var err error

	if checkFieldPresence && m.Body == nil {
		return errors.New("missing token body")
	}

	bodySet := m.Body != nil
	if b.eaclTableSet = bodySet && m.Body.EaclTable != nil; b.eaclTableSet {
		err = b.eaclTable.ReadFromV2(m.Body.EaclTable)
		if err != nil {
			return fmt.Errorf("invalid eACL table: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing eACL table")
	}

	if b.targetUserSet = bodySet && m.Body.OwnerId != nil; b.targetUserSet {
		err = b.targetUser.ReadFromV2(m.Body.OwnerId)
		if err != nil {
			return fmt.Errorf("invalid target user: %w", err)
		}
	}

	if b.issuerSet = bodySet && m.Body.Issuer != nil; b.issuerSet {
		err = b.issuer.ReadFromV2(m.Body.Issuer)
		if err != nil {
			return fmt.Errorf("invalid issuer: %w", err)
		}
	}

	if b.lifetimeSet = bodySet && m.Body.Lifetime != nil; b.lifetimeSet {
		b.iat = m.Body.Lifetime.Iat
		b.nbf = m.Body.Lifetime.Nbf
		b.exp = m.Body.Lifetime.Exp
	} else if checkFieldPresence {
		return errors.New("missing token lifetime")
	}

	if b.sigSet = m.Signature != nil; b.sigSet {
		err = b.sig.ReadFromV2(m.Signature)
		if err != nil {
			return fmt.Errorf("invalid body signature: %w", err)
		}
	}

	return nil
}

// ReadFromV2 reads Token from the [acl.BearerToken] message. Returns an error
// if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Token.WriteToV2].
func (b *Token) ReadFromV2(m *acl.BearerToken) error {
	return b.readFromV2(m, true)
}

func (b Token) fillBody() *acl.BearerToken_Body {
	if !b.eaclTableSet && !b.targetUserSet && !b.lifetimeSet && !b.issuerSet {
		return nil
	}

	var body acl.BearerToken_Body

	if b.eaclTableSet {
		body.EaclTable = new(acl.EACLTable)
		b.eaclTable.WriteToV2(body.EaclTable)
	}

	if b.targetUserSet {
		body.OwnerId = new(refs.OwnerID)
		b.targetUser.WriteToV2(body.OwnerId)
	}

	if b.issuerSet {
		body.Issuer = new(refs.OwnerID)
		b.issuer.WriteToV2(body.Issuer)
	}

	if b.lifetimeSet {
		body.Lifetime = new(acl.BearerToken_Body_TokenLifetime)
		body.Lifetime.Iat = b.iat
		body.Lifetime.Nbf = b.nbf
		body.Lifetime.Exp = b.exp
	}

	return &body
}

func (b Token) signedData() []byte {
	m := b.fillBody()
	bs := make([]byte, m.MarshaledSize())
	m.MarshalStable(bs)
	return bs
}

// WriteToV2 writes Table to the [acl.BearerToken] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Token.ReadFromV2].
func (b Token) WriteToV2(m *acl.BearerToken) {
	m.Body = b.fillBody()
	if b.sigSet {
		m.Signature = new(refs.Signature)
		b.sig.WriteToV2(m.Signature)
	}
}

// SetExp sets "exp" (expiration time) claim which identifies the
// expiration time (in NeoFS epochs) after which the Token MUST NOT be
// accepted for processing. The processing of the "exp" claim requires
// that the current epoch MUST be before or equal to the expiration epoch
// listed in the "exp" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also [Token.InvalidAt].
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
// See also [Token.InvalidAt].
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
// See also [Token.InvalidAt].
func (b *Token) SetIat(iat uint64) {
	b.iat = iat
	b.lifetimeSet = true
}

// InvalidAt asserts "exp", "nbf" and "iat" claims for the given epoch.
//
// Zero Container is invalid in any epoch.
//
// See also [Token.SetExp], [Token.SetNbf], [Token.SetIat].
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
// See also [Token.EACLTable], [Token.AssertContainer].
func (b *Token) SetEACLTable(table eacl.Table) {
	b.eaclTable = table
	b.eaclTableSet = true
}

// EACLTable returns extended ACL table set by SetEACLTable. Second value
// indicates whether the eACL is set.
//
// Zero Token has zero eacl.Table.
func (b Token) EACLTable() (eacl.Table, bool) {
	return b.eaclTable, b.eaclTableSet
}

// AssertContainer checks if the token is valid within the given container.
//
// Note: cnr is assumed to refer to the issuer's container, otherwise the check
// is meaningless.
//
// Zero Token is valid in any container.
//
// See also [Token.SetEACLTable].
func (b Token) AssertContainer(cnr cid.ID) bool {
	if !b.eaclTableSet {
		return true
	}

	cnrTable := b.eaclTable.LimitedContainer()
	return cnrTable.IsZero() || cnrTable == cnr
}

// ForUser specifies ID of the user who can use the Token for the operations
// within issuer's container(s).
//
// Optional: by default, any user has access to Token usage.
//
// See also [Token.AssertUser].
func (b *Token) ForUser(id user.ID) {
	b.targetUser = id
	b.targetUserSet = true
}

// AssertUser checks if the Token is issued to the given user.
//
// Zero Token is available to any user.
//
// See also [Token.ForUser].
func (b Token) AssertUser(id user.ID) bool {
	return !b.targetUserSet || b.targetUser == id
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
	if err != nil {
		return err
	}
	b.sigSet = true

	return nil
}

// SignedData returns signed data of the Token.
//
// See also [Token.Sign], [Token.UnmarshalSignedData].
func (b *Token) SignedData() []byte {
	return b.signedData()
}

// UnmarshalSignedData is a reverse op to [Token.SignedData].
func (b *Token) UnmarshalSignedData(data []byte) error {
	var body acl.BearerToken_Body
	err := proto.Unmarshal(data, &body)
	if err != nil {
		return fmt.Errorf("decode body: %w", err)
	}

	return b.readFromV2(&acl.BearerToken{Body: &body}, false)
}

// VerifySignature checks if Token signature is presented and valid.
//
// Zero Token fails the check.
//
// See also [Token.Sign].
func (b Token) VerifySignature() bool {
	// TODO: (#233) check owner<->key relation
	return b.sigSet && b.sig.Verify(b.signedData())
}

// Marshal encodes Token into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also [Token.Unmarshal].
func (b Token) Marshal() []byte {
	var m acl.BearerToken
	b.WriteToV2(&m)
	bs := make([]byte, m.MarshaledSize())
	m.MarshalStable(bs)
	return bs
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Table. Returns an
// error describing a format violation of the specified fields. Unmarshal does
// not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Token.Marshal].
func (b *Token) Unmarshal(data []byte) error {
	var m acl.BearerToken
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return b.readFromV2(&m, false)
}

// MarshalJSON encodes Token into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Token.UnmarshalJSON].
func (b Token) MarshalJSON() ([]byte, error) {
	var m acl.BearerToken
	b.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Token (Protocol
// Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Table.MarshalJSON].
func (b *Token) UnmarshalJSON(data []byte) error {
	var m acl.BearerToken
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}

	return b.readFromV2(&m, false)
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
		return b.sig.PublicKeyBytes()
	}

	return nil
}

// SetIssuer sets NeoFS user ID of the Token issuer.
//
// See also [Token.Issuer], [Token.Sign].
func (b *Token) SetIssuer(usr user.ID) {
	b.issuerSet = true
	b.issuer = usr
}

// Issuer returns NeoFS user ID of the explicitly set Token issuer. Zero value
// means unset issuer. In this case, [Token.ResolveIssuer] can be used to get ID
// resolved from signer's public key.
//
// See also [Token.SetIssuer], [Token.Sign].
func (b Token) Issuer() user.ID {
	if b.issuerSet {
		return b.issuer
	}
	return user.ID{}
}

// ResolveIssuer works like [Token.Issuer] with fallback to the public key
// resolution when explicit issuer ID is unset. Returns zero [user.ID] when
// neither issuer is set nor key resolution succeeds.
//
// See also [Token.SigningKeyBytes], [Token.Sign].
func (b Token) ResolveIssuer() user.ID {
	if b.issuerSet {
		return b.issuer
	}

	var usr user.ID
	binKey := b.SigningKeyBytes()

	if len(binKey) != 0 {
		var pk keys.PublicKey
		if err := pk.DecodeBytes(binKey); err == nil {
			usr = user.ResolveFromECDSAPublicKey(ecdsa.PublicKey(pk))
		}
	}

	return usr
}
