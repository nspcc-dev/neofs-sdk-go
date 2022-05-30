package bearer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

var (
	errNilBearerTokenBody = errors.New("bearer token body is not set")
	errNilBearerTokenEACL = errors.New("bearer token ContainerEACL table is not set")
)

// Token represents bearer token for object service operations.
//
// Token is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/acl.BearerToken
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Token(acl.BearerToken{}) // not recommended
type Token acl.BearerToken

// ReadFromV2 reads Token from the acl.BearerToken message.
//
// See also WriteToV2.
func (b *Token) ReadFromV2(m acl.BearerToken) {
	*b = Token(m)
}

// WriteToV2 writes Token to the acl.BearerToken message.
// The message must not be nil.
//
// See also ReadFromV2.
func (b Token) WriteToV2(m *acl.BearerToken) {
	*m = (acl.BearerToken)(b)
}

// SetExpiration sets "exp" (expiration time) claim which identifies the
// expiration time (in NeoFS epochs) on or after which the Token MUST NOT be
// accepted for processing. The processing of the "exp" claim requires that the
// current epoch MUST be before the expiration epoch listed in the "exp" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4.
//
// See also Expiration.
func (b *Token) SetExpiration(exp uint64) {
	v2token := (*acl.BearerToken)(b)

	body := v2token.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	lt := new(acl.TokenLifetime)
	lt.SetExp(exp)
	lt.SetNbf(body.GetLifetime().GetNbf())
	lt.SetIat(body.GetLifetime().GetIat())

	body.SetLifetime(lt)
	v2token.SetBody(body)
}

// Expiration returns "exp" claim.
//
// Empty Token has zero "exp".
//
// See also SetExpiration.
func (b Token) Expiration() uint64 {
	v2token := (acl.BearerToken)(b)
	return v2token.GetBody().GetLifetime().GetExp()
}

// SetNotBefore sets "nbf" (not before) claim which identifies the time (in
// NeoFS epochs) before which the Token MUST NOT be accepted for processing. The
// processing of the "nbf" claim requires that the current epoch MUST be
// after or equal to the not-before epoch listed in the "nbf" claim.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5.
//
// See also NotBefore.
func (b *Token) SetNotBefore(nbf uint64) {
	v2token := (*acl.BearerToken)(b)

	body := v2token.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	lt := new(acl.TokenLifetime)
	lt.SetExp(body.GetLifetime().GetExp())
	lt.SetNbf(nbf)
	lt.SetIat(body.GetLifetime().GetIat())

	body.SetLifetime(lt)
	v2token.SetBody(body)
}

// NotBefore returns "nbf" claim.
//
// Empty Token has zero "nbf".
//
// See also SetNotBefore.
func (b Token) NotBefore() uint64 {
	v2token := (acl.BearerToken)(b)
	return v2token.GetBody().GetLifetime().GetNbf()
}

// SetIssuedAt sets "iat" (issued at) claim which identifies the time (in NeoFS
// epochs) at which the Token was issued. This claim can be used to determine
// the age of the Token.
//
// Naming is inspired by https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6.
//
// See also IssuedAt.
func (b *Token) SetIssuedAt(iat uint64) {
	v2token := (*acl.BearerToken)(b)

	body := v2token.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	lt := new(acl.TokenLifetime)
	lt.SetExp(body.GetLifetime().GetExp())
	lt.SetNbf(body.GetLifetime().GetNbf())
	lt.SetIat(iat)

	body.SetLifetime(lt)
	v2token.SetBody(body)
}

// IssuedAt returns "iat" claim.
//
// Empty Token has zero "iat".
//
// See also SetIssuedAt.
func (b Token) IssuedAt() uint64 {
	v2token := (acl.BearerToken)(b)
	return v2token.GetBody().GetLifetime().GetIat()
}

// SetEACLTable sets extended ACL table that should be used during object
// service request processing with bearer token.
//
// See also EACLTable.
func (b *Token) SetEACLTable(table eacl.Table) {
	v2 := (*acl.BearerToken)(b)

	body := v2.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	body.SetEACL(table.ToV2())
	v2.SetBody(body)
}

// EACLTable returns extended ACL table that should be used during object
// service request processing with bearer token.
//
// See also SetEACLTable.
func (b Token) EACLTable() eacl.Table {
	v2 := (acl.BearerToken)(b)
	return *eacl.NewTableFromV2(v2.GetBody().GetEACL())
}

// SetOwnerID sets user.ID value of the user who can attach bearer token to
// its requests.
//
// See also OwnerID.
func (b *Token) SetOwnerID(id user.ID) {
	v2 := (*acl.BearerToken)(b)

	body := v2.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	var idV2 refs.OwnerID
	id.WriteToV2(&idV2)

	body.SetOwnerID(&idV2)
	v2.SetBody(body)
}

// OwnerID returns user.ID value of the user who can attach bearer token to
// its requests.
//
// See also SetOwnerID.
func (b Token) OwnerID() user.ID {
	v2 := (acl.BearerToken)(b)

	var id user.ID

	idV2 := v2.GetBody().GetOwnerID()
	if idV2 != nil {
		_ = id.ReadFromV2(*idV2)
	}

	return id
}

// Sign signs bearer token. This method should be invoked with the private
// key of container owner to allow overriding extended ACL table of the container
// included in this token.
//
// See also Signature.
func (b *Token) Sign(key ecdsa.PrivateKey) error {
	err := sanityCheck(b)
	if err != nil {
		return err
	}

	m := (*acl.BearerToken)(b)

	var sig neofscrypto.Signature

	err = sig.Calculate(neofsecdsa.Signer(key), m.GetBody().StableMarshal(nil))
	if err != nil {
		return fmt.Errorf("calculate signature: %w", err)
	}

	var sigV2 refs.Signature
	sig.WriteToV2(&sigV2)

	m.SetSignature(&sigV2)

	return nil
}

// VerifySignature returns nil if bearer token contains correct signature.
func (b Token) VerifySignature() error {
	if b.isEmpty() {
		return nil
	}

	m := (acl.BearerToken)(b)

	sigV2 := m.GetSignature()
	if sigV2 == nil {
		return errors.New("missing signature")
	}

	var sig neofscrypto.Signature
	sig.ReadFromV2(*sigV2)

	if !sig.Verify(m.GetBody().StableMarshal(nil)) {
		return errors.New("wrong signature")
	}

	return nil
}

// Issuer returns user.ID associated with the key that signed bearer token.
// To pass node validation it should be owner of requested container.
//
// If token is not signed, Issuer returns empty owner ID and false `ok` flag.
//
// See also Sign.
func (b Token) Issuer() (id user.ID, ok bool) {
	v2 := (acl.BearerToken)(b)

	pub, _ := keys.NewPublicKeyFromBytes(v2.GetSignature().GetKey(), elliptic.P256())

	ok = pub != nil
	if ok {
		user.IDFromKey(&id, (ecdsa.PublicKey)(*pub))
	}

	return
}

// sanityCheck if bearer token is ready to be issued.
func sanityCheck(b *Token) error {
	v2 := (*acl.BearerToken)(b)

	switch {
	case v2.GetBody() == nil:
		return errNilBearerTokenBody
	case v2.GetBody().GetEACL() == nil:
		return errNilBearerTokenEACL
	}

	// consider checking ContainerEACL sanity there, lifetime correctness, etc.

	return nil
}

// Marshal marshals Token into a canonical NeoFS binary format (proto3
// with direct field order).
//
// See also Unmarshal.
func (b Token) Marshal() []byte {
	return (*acl.BearerToken)(&b).StableMarshal(nil)
}

// Unmarshal unmarshals Token from canonical NeoFS binary format (proto3
// with direct field order).
//
// See also Marshal.
func (b *Token) Unmarshal(data []byte) error {
	v2 := (*acl.BearerToken)(b)
	return v2.Unmarshal(data)
}

// MarshalJSON encodes Token to protobuf JSON format.
//
// See also UnmarshalJSON.
func (b Token) MarshalJSON() ([]byte, error) {
	v2 := (acl.BearerToken)(b)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes Token from protobuf JSON format.
//
// See also MarshalJSON.
func (b *Token) UnmarshalJSON(data []byte) error {
	v2 := (*acl.BearerToken)(b)
	return v2.UnmarshalJSON(data)
}

func (b Token) isEmpty() bool {
	v2token := (acl.BearerToken)(b)
	return v2token.GetBody() == nil && v2token.GetSignature() == nil
}
