package neofscrypto

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
)

// StablyMarshallable describes structs which can be marshalled transparently.
type StablyMarshallable interface {
	StableMarshal([]byte) []byte
	StableSize() int
}

// Signature represents a confirmation of data integrity received by the
// digital signature mechanism.
//
// Signature is mutually compatible with github.com/nspcc-dev/neofs-sdk-go/api/refs.Signature
// message. See ReadFromV2 / WriteToV2 methods.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
//
//	_ = Signature(refs.Signature{}) // not recommended
type Signature struct {
	scheme Scheme
	pubKey []byte
	val    []byte
}

// NewSignature is a Signature instance constructor.
func NewSignature(scheme Scheme, publicKey PublicKey, value []byte) Signature {
	var s Signature
	s.setFields(scheme, publicKey, value)
	return s
}

// CopyTo writes deep copy of the Signature to dst.
func (x Signature) CopyTo(dst *Signature) {
	dst.scheme = x.scheme
	dst.pubKey = bytes.Clone(x.pubKey)
	dst.val = bytes.Clone(x.val)
}

// ReadFromV2 reads Signature from the refs.Signature message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *Signature) ReadFromV2(m *refs.Signature) error {
	bPubKey := m.GetKey()
	if len(bPubKey) == 0 {
		return errors.New("missing public key")
	}

	sig := m.GetSign()
	if len(sig) == 0 {
		return errors.New("missing signature")
	}

	_, err := decodePublicKey(Scheme(m.Scheme), m.Key)
	if err != nil {
		return err
	}

	x.scheme = Scheme(m.Scheme)
	x.pubKey = m.Key
	x.val = m.Sign

	return nil
}

// WriteToV2 writes Signature to the refs.Signature message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Signature) WriteToV2(m *refs.Signature) {
	m.Scheme = refs.SignatureScheme(x.scheme)
	m.Key = x.pubKey
	m.Sign = x.val
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
// Pre-allocated byte slice can be passed in buf parameter to avoid new allocations. In ideal case buf length should be
// StableSize length. If buffer length shorter than StableSize or nil, new slice will be allocated.
//
// Signer MUST NOT be nil.
//
// See also Verify.
func (x *Signature) CalculateMarshalled(signer Signer, obj StablyMarshallable, buf []byte) error {
	if static, ok := signer.(*StaticSigner); ok {
		x.fillSignature(signer, static.sig)
		return nil
	}

	var data []byte
	if obj != nil {
		if len(buf) >= obj.StableSize() {
			data = obj.StableMarshal(buf[0:obj.StableSize()])
		} else {
			data = obj.StableMarshal(nil)
		}
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
	key, err := decodePublicKey(x.scheme, x.pubKey)

	return err == nil && key.Verify(data, x.val)
}

func (x *Signature) fillSignature(signer Signer, signature []byte) {
	x.setFields(signer.Scheme(), signer.Public(), signature)
}

func (x *Signature) setFields(scheme Scheme, publicKey PublicKey, value []byte) {
	x.scheme = scheme
	x.val = value
	x.pubKey = PublicKeyBytes(publicKey)
}

// Scheme returns signature scheme used by signer to calculate the signature.
//
// Scheme MUST NOT be called before [NewSignature], [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
func (x Signature) Scheme() Scheme {
	return x.scheme
}

// PublicKey returns public key of the signer which calculated the signature.
//
// PublicKey MUST NOT be called before [NewSignature], [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
//
// See also [Signature.PublicKeyBytes].
func (x Signature) PublicKey() PublicKey {
	key, _ := decodePublicKey(x.scheme, x.pubKey)
	return key
}

// PublicKeyBytes returns binary-encoded public key of the signer which
// calculated the signature.
//
// PublicKeyBytes MUST NOT be called before [NewSignature],
// [Signature.ReadFromV2] or [Signature.Calculate] methods.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Signature.PublicKey].
func (x Signature) PublicKeyBytes() []byte {
	return x.pubKey
}

// Value returns calculated digital signature.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// Value MUST NOT be called before [NewSignature], [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
func (x Signature) Value() []byte {
	return x.val
}

func decodePublicKey(scheme Scheme, b []byte) (PublicKey, error) {
	newPubKey, ok := publicKeys[scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported scheme %d", scheme)
	}

	pubKey := newPubKey()

	err := pubKey.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("decode public key from binary: %w", err)
	}

	return pubKey, nil
}
