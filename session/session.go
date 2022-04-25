package session

import (
	"crypto/ecdsa"

	"github.com/nspcc-dev/neofs-api-go/v2/session"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	sigutil "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

// Token represents NeoFS API v2-compatible
// session token.
type Token session.Token

// NewTokenFromV2 wraps session.Token message structure
// into Token.
//
// Nil session.Token converts to nil.
func NewTokenFromV2(tV2 *session.Token) *Token {
	return (*Token)(tV2)
}

// NewToken creates and returns blank Token.
//
// Defaults:
//  - body: nil;
//  - id: nil;
//  - ownerId: nil;
//  - sessionKey: nil;
//  - exp: 0;
//  - iat: 0;
//  - nbf: 0;
func NewToken() *Token {
	return NewTokenFromV2(new(session.Token))
}

// ToV2 converts Token to session.Token message structure.
//
// Nil Token converts to nil.
func (t *Token) ToV2() *session.Token {
	return (*session.Token)(t)
}

func (t *Token) setBodyField(setter func(*session.TokenBody)) {
	token := (*session.Token)(t)
	body := token.GetBody()

	if body == nil {
		body = new(session.TokenBody)
		token.SetBody(body)
	}

	setter(body)
}

// ID returns Token identifier.
func (t *Token) ID() []byte {
	return (*session.Token)(t).
		GetBody().
		GetID()
}

// SetID sets Token identifier.
func (t *Token) SetID(v []byte) {
	t.setBodyField(func(body *session.TokenBody) {
		body.SetID(v)
	})
}

// OwnerID returns Token's owner identifier.
func (t *Token) OwnerID() *owner.ID {
	return owner.NewIDFromV2(
		(*session.Token)(t).
			GetBody().
			GetOwnerID(),
	)
}

// SetOwnerID sets Token's owner identifier.
func (t *Token) SetOwnerID(v *owner.ID) {
	t.setBodyField(func(body *session.TokenBody) {
		body.SetOwnerID(v.ToV2())
	})
}

// SessionKey returns public key of the session
// in a binary format.
func (t *Token) SessionKey() []byte {
	return (*session.Token)(t).
		GetBody().
		GetSessionKey()
}

// SetSessionKey sets public key of the session
// in a binary format.
func (t *Token) SetSessionKey(v []byte) {
	t.setBodyField(func(body *session.TokenBody) {
		body.SetSessionKey(v)
	})
}

func (t *Token) setLifetimeField(f func(*session.TokenLifetime)) {
	t.setBodyField(func(body *session.TokenBody) {
		lt := body.GetLifetime()
		if lt == nil {
			lt = new(session.TokenLifetime)
			body.SetLifetime(lt)
		}

		f(lt)
	})
}

// Exp returns epoch number of the token expiration.
func (t *Token) Exp() uint64 {
	return (*session.Token)(t).
		GetBody().
		GetLifetime().
		GetExp()
}

// SetExp sets epoch number of the token expiration.
func (t *Token) SetExp(exp uint64) {
	t.setLifetimeField(func(lt *session.TokenLifetime) {
		lt.SetExp(exp)
	})
}

// Nbf returns starting epoch number of the token.
func (t *Token) Nbf() uint64 {
	return (*session.Token)(t).
		GetBody().
		GetLifetime().
		GetNbf()
}

// SetNbf sets starting epoch number of the token.
func (t *Token) SetNbf(nbf uint64) {
	t.setLifetimeField(func(lt *session.TokenLifetime) {
		lt.SetNbf(nbf)
	})
}

// Iat returns starting epoch number of the token.
func (t *Token) Iat() uint64 {
	return (*session.Token)(t).
		GetBody().
		GetLifetime().
		GetIat()
}

// SetIat sets the number of the epoch in which the token was issued.
func (t *Token) SetIat(iat uint64) {
	t.setLifetimeField(func(lt *session.TokenLifetime) {
		lt.SetIat(iat)
	})
}

// Sign calculates and writes signature of the Token data.
//
// Returns signature calculation errors.
func (t *Token) Sign(key *ecdsa.PrivateKey) error {
	tV2 := (*session.Token)(t)

	signedData := v2signature.StableMarshalerWrapper{
		SM: tV2.GetBody(),
	}

	sig, err := sigutil.SignData(key, signedData)
	if err != nil {
		return err
	}

	tV2.SetSignature(sig.ToV2())
	return nil
}

// VerifySignature checks if token signature is
// presented and valid.
func (t *Token) VerifySignature() bool {
	tV2 := (*session.Token)(t)

	signedData := v2signature.StableMarshalerWrapper{
		SM: tV2.GetBody(),
	}

	return sigutil.VerifyData(signedData, t.Signature()) == nil
}

// Signature returns Token signature.
func (t *Token) Signature() *signature.Signature {
	return signature.NewFromV2(
		(*session.Token)(t).
			GetSignature(),
	)
}

// SetContext sets context of the Token.
//
// Supported contexts:
//  - *ContainerContext,
//  - *ObjectContext.
//
// Resets context if it is not supported.
func (t *Token) SetContext(v interface{}) {
	var cV2 session.TokenContext

	switch c := v.(type) {
	case *ContainerContext:
		cV2 = c.ToV2()
	case *ObjectContext:
		cV2 = c.ToV2()
	}

	t.setBodyField(func(body *session.TokenBody) {
		body.SetContext(cV2)
	})
}

// Context returns context of the Token.
//
// Supports same contexts as SetContext.
//
// Returns nil if context is not supported.
func (t *Token) Context() interface{} {
	switch v := (*session.Token)(t).
		GetBody().
		GetContext(); c := v.(type) {
	default:
		return nil
	case *session.ContainerSessionContext:
		return NewContainerContextFromV2(c)
	case *session.ObjectSessionContext:
		return NewObjectContextFromV2(c)
	}
}

// GetContainerContext is a helper function that casts
// Token context to ContainerContext.
//
// Returns nil if context is not a ContainerContext.
func GetContainerContext(t *Token) *ContainerContext {
	c, _ := t.Context().(*ContainerContext)
	return c
}

// Marshal marshals Token into a protobuf binary form.
func (t *Token) Marshal() []byte {
	return (*session.Token)(t).
		StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Token.
func (t *Token) Unmarshal(data []byte) error {
	return (*session.Token)(t).
		Unmarshal(data)
}

// MarshalJSON encodes Token to protobuf JSON format.
func (t *Token) MarshalJSON() ([]byte, error) {
	return (*session.Token)(t).
		MarshalJSON()
}

// UnmarshalJSON decodes Token from protobuf JSON format.
func (t *Token) UnmarshalJSON(data []byte) error {
	return (*session.Token)(t).
		UnmarshalJSON(data)
}
