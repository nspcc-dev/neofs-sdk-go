package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	sigutil "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

var (
	errNilBearerToken     = errors.New("bearer token is not set")
	errNilBearerTokenBody = errors.New("bearer token body is not set")
	errNilBearerTokenEACL = errors.New("bearer token ContainerEACL table is not set")
)

type BearerToken struct {
	token acl.BearerToken
}

// ToV2 converts BearerToken to v2 BearerToken message.
//
// Nil BearerToken converts to nil.
func (b *BearerToken) ToV2() *acl.BearerToken {
	if b == nil {
		return nil
	}

	return &b.token
}

func (b *BearerToken) Empty() bool {
	return b == nil || b.token.GetBody() == nil && b.token.GetSignature() == nil
}

func (b *BearerToken) SetLifetime(exp, nbf, iat uint64) {
	body := b.token.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	lt := new(acl.TokenLifetime)
	lt.SetExp(exp)
	lt.SetNbf(nbf)
	lt.SetIat(iat)

	body.SetLifetime(lt)
	b.token.SetBody(body)
}

func (b BearerToken) Expiration() uint64 {
	return b.token.GetBody().GetLifetime().GetExp()
}

func (b BearerToken) NotBeforeTime() uint64 {
	return b.token.GetBody().GetLifetime().GetNbf()
}

func (b BearerToken) IssuedAt() uint64 {
	return b.token.GetBody().GetLifetime().GetIat()
}

func (b *BearerToken) SetEACLTable(table *eacl.Table) {
	body := b.token.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	body.SetEACL(table.ToV2())
	b.token.SetBody(body)
}

func (b BearerToken) EACLTable() *eacl.Table {
	return eacl.NewTableFromV2(b.token.GetBody().GetEACL())
}

func (b *BearerToken) SetOwner(id *owner.ID) {
	body := b.token.GetBody()
	if body == nil {
		body = new(acl.BearerTokenBody)
	}

	body.SetOwnerID(id.ToV2())
	b.token.SetBody(body)
}

func (b BearerToken) OwnerID() *owner.ID {
	return owner.NewIDFromV2(b.token.GetBody().GetOwnerID())
}

func (b *BearerToken) SignToken(key *ecdsa.PrivateKey) error {
	err := sanityCheck(b)
	if err != nil {
		return err
	}

	signWrapper := v2signature.StableMarshalerWrapper{SM: b.token.GetBody()}

	sig, err := sigutil.SignData(key, signWrapper)
	if err != nil {
		return err
	}

	b.token.SetSignature(sig.ToV2())
	return nil
}

func (b BearerToken) Signature() *signature.Signature {
	return signature.NewFromV2(b.token.GetSignature())
}

func (b BearerToken) VerifySignature() error {
	if b.Empty() {
		return nil
	}

	sigV2 := b.token.GetSignature()
	return sigutil.VerifyData(
		v2signature.StableMarshalerWrapper{SM: b.token.GetBody()},
		signature.NewFromV2(sigV2))
}

// Issuer returns owner.ID associated with the key that signed bearer token.
// To pass node validation it should be owner of requested container. Returns
// nil if token is not signed.
func (b *BearerToken) Issuer() *owner.ID {
	pub, _ := keys.NewPublicKeyFromBytes(b.token.GetSignature().GetKey(), elliptic.P256())
	if pub == nil {
		return nil
	}
	return owner.NewIDFromPublicKey((*ecdsa.PublicKey)(pub))
}

// NewBearerToken creates and initializes blank BearerToken.
//
// Defaults:
//  - signature: nil;
//  - eacl: nil;
//  - ownerID: nil;
//  - exp: 0;
//  - nbf: 0;
//  - iat: 0.
func NewBearerToken() *BearerToken {
	b := new(BearerToken)
	b.token = acl.BearerToken{}
	b.token.SetBody(new(acl.BearerTokenBody))

	return b
}

// ToV2 converts BearerToken to v2 BearerToken message.
func NewBearerTokenFromV2(v2 *acl.BearerToken) *BearerToken {
	if v2 == nil {
		v2 = new(acl.BearerToken)
	}

	return &BearerToken{
		token: *v2,
	}
}

// sanityCheck if bearer token is ready to be issued.
func sanityCheck(b *BearerToken) error {
	switch {
	case b == nil:
		return errNilBearerToken
	case b.token.GetBody() == nil:
		return errNilBearerTokenBody
	case b.token.GetBody().GetEACL() == nil:
		return errNilBearerTokenEACL
	}

	// consider checking ContainerEACL sanity there, lifetime correctness, etc.

	return nil
}

// Marshal marshals BearerToken into a protobuf binary form.
//
// Buffer is allocated when the argument is empty.
// Otherwise, the first buffer is used.
func (b *BearerToken) Marshal(bs ...[]byte) ([]byte, error) {
	var buf []byte
	if len(bs) > 0 {
		buf = bs[0]
	}

	return b.ToV2().
		StableMarshal(buf)
}

// Unmarshal unmarshals protobuf binary representation of BearerToken.
func (b *BearerToken) Unmarshal(data []byte) error {
	fV2 := new(acl.BearerToken)
	if err := fV2.Unmarshal(data); err != nil {
		return err
	}

	*b = *NewBearerTokenFromV2(fV2)

	return nil
}

// MarshalJSON encodes BearerToken to protobuf JSON format.
func (b *BearerToken) MarshalJSON() ([]byte, error) {
	return b.ToV2().
		MarshalJSON()
}

// UnmarshalJSON decodes BearerToken from protobuf JSON format.
func (b *BearerToken) UnmarshalJSON(data []byte) error {
	fV2 := new(acl.BearerToken)
	if err := fV2.UnmarshalJSON(data); err != nil {
		return err
	}

	*b = *NewBearerTokenFromV2(fV2)

	return nil
}
