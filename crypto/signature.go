package neofscrypto

import (
	"fmt"
	"math"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// StablyMarshallable describes structs which can be marshalled transparently.
// Deprecated: avoid using it.
type StablyMarshallable interface {
	StableMarshal([]byte) []byte
	StableSize() int
}

// Signature represents a confirmation of data integrity received by the
// digital signature mechanism.
//
// Signature is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Signature
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances should be constructed using one of the constructors.
type Signature struct {
	scheme Scheme
	pub    []byte
	val    []byte
}

// NewSignatureFromRawKey constructs new Signature instance.
func NewSignatureFromRawKey(scheme Scheme, pub []byte, value []byte) Signature {
	return Signature{scheme: scheme, pub: pub, val: value}
}

// NewSignature is a Signature instance constructor.
func NewSignature(scheme Scheme, publicKey PublicKey, value []byte) Signature {
	return NewSignatureFromRawKey(scheme, PublicKeyBytes(publicKey), value)
}

// CalculateDataSignature calculates signature of the given data. Use
// [IsValidDataSignature] for verification.
func CalculateDataSignature(signer Signer, data []byte) (Signature, error) {
	sig, err := signer.Sign(data)
	if err != nil {
		return Signature{}, fmt.Errorf("signer %T failure: %w", signer, err)
	}
	return NewSignature(signer.Scheme(), signer.Public(), sig), nil
}

// IsValidDataSignature checks whether sig is a correct signature of the given
// data. IsValidDataSignature returns false if scheme is unsupported. Use
// [CalculateDataSignature] for calculation.
func IsValidDataSignature(sig Signature, data []byte) bool {
	if pub, err := DecodePublicKey(sig.scheme, sig.PublicKeyBytes()); err == nil {
		return pub.Verify(data, sig.Value())
	}
	return false
}

// ReadFromV2 reads Signature from the refs.Signature message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *Signature) ReadFromV2(m refs.Signature) error {
	scheme := m.GetScheme()
	if scheme > math.MaxInt32 { // max value of Scheme type
		return fmt.Errorf("scheme %d overflows int32", scheme)
	}
	x.scheme = Scheme(scheme)
	x.pub = m.GetKey()
	x.val = m.GetSign()
	return nil
}

// WriteToV2 writes Signature to the refs.Signature message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Signature) WriteToV2(m *refs.Signature) {
	m.SetScheme(refs.SignatureScheme(x.scheme))
	m.SetKey(x.pub)
	m.SetSign(x.val)
}

// Calculate signs data using Signer and encodes public key for subsequent
// verification.
//
// Signer MUST NOT be nil.
//
// See also Verify.
// Deprecated: use [CalculateDataSignature] instead.
func (x *Signature) Calculate(signer Signer, data []byte) error {
	s, err := CalculateDataSignature(signer, data)
	if err == nil {
		*x = s
	}
	return err
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
// Deprecated: encode manually and use [CalculateDataSignature] instead.
func (x *Signature) CalculateMarshalled(signer Signer, obj StablyMarshallable, buf []byte) error {
	if static, ok := signer.(*StaticSigner); ok {
		*x = NewSignature(signer.Scheme(), signer.Public(), static.sig)
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
// Deprecated: use [IsValidDataSignature] instead.
func (x Signature) Verify(data []byte) bool { return IsValidDataSignature(x, data) }

// SetScheme sets signature scheme used by signer to calculate the signature.
func (x *Signature) SetScheme(s Scheme) {
	x.scheme = s
}

// Scheme returns signature scheme used by signer to calculate the signature.
func (x Signature) Scheme() Scheme {
	return x.scheme
}

// PublicKey returns public key of the signer which calculated the signature.
//
// PublicKey MUST NOT be called before [NewSignature], [Signature.ReadFromV2] or
// [Signature.Calculate] methods.
//
// See also [Signature.PublicKeyBytes].
// Deprecated: use [DecodePublicKey] instead.
func (x Signature) PublicKey() PublicKey {
	key, _ := DecodePublicKey(x.scheme, x.pub)
	return key
}

// SetPublicKeyBytes returns binary-encoded public key of the signer which
// calculated the signature.
func (x *Signature) SetPublicKeyBytes(pub []byte) {
	x.pub = pub
}

// PublicKeyBytes returns binary-encoded public key of the signer which
// calculated the signature.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x Signature) PublicKeyBytes() []byte {
	return x.pub
}

// SetValue sets calculated digital signature.
func (x *Signature) SetValue(v []byte) {
	x.val = v
}

// Value returns calculated digital signature.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x Signature) Value() []byte {
	return x.val
}

// DecodePublicKey attempts to decode b into PublicKey according to the given
// scheme. Use [RegisterScheme] to support custom encoding scheme.
func DecodePublicKey(scheme Scheme, b []byte) (PublicKey, error) {
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
