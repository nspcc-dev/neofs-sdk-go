package neofscrypto

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// StablyMarshallable describes structs which can be marshalled transparently.
type StablyMarshallable interface {
	StableMarshal([]byte) []byte
}

// Signature represents a confirmation of data integrity received by the
// digital signature mechanism.
//
// Signature is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Signature
// message. See ReadFromV2 / WriteToV2 methods.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
//
//	_ = Signature(refs.Signature{}) // not recommended
type Signature refs.Signature

// ReadFromV2 reads Signature from the refs.Signature message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *Signature) ReadFromV2(m refs.Signature) error {
	bPubKey := m.GetKey()
	if len(bPubKey) == 0 {
		return errors.New("missing public key")
	}

	sig := m.GetSign()
	if len(sig) == 0 {
		return errors.New("missing signature")
	}

	_, err := decodePublicKey(m)
	if err != nil {
		return err
	}

	*x = Signature(m)

	return nil
}

// WriteToV2 writes Signature to the refs.Signature message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Signature) WriteToV2(m *refs.Signature) {
	*m = refs.Signature(x)
}

// Calculate signs data using Signer and encodes public key for subsequent
// verification.
//
// Signer MUST NOT be nil.
//
// See also Verify.
func (x *Signature) Calculate(signer Signer, data []byte) error {
	signature, err := signer.Sign(data)
	if err != nil {
		return fmt.Errorf("signer %T failure: %w", signer, err)
	}

	x.fillSignature(signer, signature)

	return nil
}

// CalculateMarshalled signs data using Signer and encodes public key for subsequent verification.
// If signer is a StaticSigner, just sets prepared signature.
//
// Signer MUST NOT be nil.
//
// See also Verify.
func (x *Signature) CalculateMarshalled(signer Signer, obj StablyMarshallable) error {
	if static, ok := signer.(*StaticSigner); ok {
		x.fillSignature(signer, static.sig)
		return nil
	}

	var data []byte
	if obj != nil {
		data = obj.StableMarshal(nil)
	}

	return x.Calculate(signer, data)
}

// Verify verifies data signature using encoded public key. True means valid
// signature.
//
// Verify fails if signature scheme is not supported (see RegisterScheme).
//
// See also Calculate.
func (x Signature) Verify(data []byte) bool {
	m := refs.Signature(x)

	key, err := decodePublicKey(m)

	return err == nil && key.Verify(data, m.GetSign())
}

func (x *Signature) fillSignature(signer Signer, signature []byte) {
	m := (*refs.Signature)(x)
	m.SetScheme(refs.SignatureScheme(signer.Scheme()))
	m.SetSign(signature)
	m.SetKey(PublicKeyBytes(signer.Public()))
}

// Scheme returns signature scheme used by signer to calculate the signature.
//
// Scheme MUST NOT be called before [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
func (x Signature) Scheme() Scheme {
	return Scheme((*refs.Signature)(&x).GetScheme())
}

// PublicKey returns public key of the signer which calculated the signature.
//
// PublicKey MUST NOT be called before [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
//
// See also [Signature.PublicKeyBytes].
func (x Signature) PublicKey() PublicKey {
	key, _ := decodePublicKey(refs.Signature(x))
	return key
}

// PublicKeyBytes returns binary-encoded public key of the signer which
// calculated the signature.
//
// PublicKeyBytes MUST NOT be called before [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Signature.PublicKey].
func (x Signature) PublicKeyBytes() []byte {
	return (*refs.Signature)(&x).GetKey()
}

// Value returns calculated digital signature.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// Value MUST NOT be called before [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
func (x Signature) Value() []byte {
	return (*refs.Signature)(&x).GetSign()
}

func decodePublicKey(m refs.Signature) (PublicKey, error) {
	scheme := Scheme(m.GetScheme())

	newPubKey, ok := publicKeys[scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme %d", scheme)
	}

	pubKey := newPubKey()

	err := pubKey.Decode(m.GetKey())
	if err != nil {
		return nil, fmt.Errorf("decode public key from binary: %w", err)
	}

	return pubKey, nil
}
