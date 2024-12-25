package neofscrypto

import (
	"fmt"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
)

// StablyMarshallable describes structs which can be marshalled transparently.
type StablyMarshallable = neofsproto.Message

// Signature represents a confirmation of data integrity received by the
// digital signature mechanism.
//
// Signature is mutually compatible with [refs.Signature] message. See
// [Signature.FromProtoMessage] / [Signature.ProtoMessage] methods.
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

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [Signature.ProtoMessage].
func (x *Signature) FromProtoMessage(m *refs.Signature) error {
	if m.Scheme < 0 {
		return fmt.Errorf("negative scheme %d", m.Scheme)
	}
	x.scheme = Scheme(m.Scheme)
	x.pub = m.Key
	x.val = m.Sign
	return nil
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [Signature.FromProtoMessage].
func (x Signature) ProtoMessage() *refs.Signature {
	return &refs.Signature{
		Key:    x.pub,
		Sign:   x.val,
		Scheme: refs.SignatureScheme(x.scheme),
	}
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

	*x = NewSignature(signer.Scheme(), signer.Public(), signature)

	return nil
}

// Verify verifies data signature using encoded public key. True means valid
// signature.
//
// Verify fails if signature scheme is not supported (see RegisterScheme).
//
// See also Calculate.
func (x Signature) Verify(data []byte) bool {
	key, err := decodePublicKey(x.scheme, x.pub)

	return err == nil && key.Verify(data, x.val)
}

// Scheme returns signature scheme used by signer to calculate the signature.
//
// Scheme MUST NOT be called before [NewSignature], [Signature.FromProtoMessage]
// or [Signature.Calculate] methods.
func (x Signature) Scheme() Scheme {
	return x.scheme
}

// SetScheme sets signature scheme used by signer to calculate the signature.
func (x *Signature) SetScheme(s Scheme) {
	x.scheme = s
}

// PublicKey returns public key of the signer which calculated the signature.
//
// PublicKey MUST NOT be called before [NewSignature],
// [Signature.FromProtoMessage] or [Signature.Calculate] methods.
//
// See also [Signature.PublicKeyBytes].
func (x Signature) PublicKey() PublicKey {
	key, _ := decodePublicKey(x.scheme, x.pub)
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
// PublicKeyBytes MUST NOT be called before [NewSignature],
// [Signature.FromProtoMessage] or [Signature.Calculate] methods.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Signature.PublicKey].
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
//
// Value MUST NOT be called before [NewSignature], [Signature.FromProtoMessage]
// or [Signature.Calculate] methods.
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

// Marshal encodes x transmitted via NeoFS API protocol into a dynamically
// allocated buffer.
func (x Signature) Marshal() []byte {
	return neofsproto.Marshal(x)
}

// Unmarshal decodes x transmitted via NeoFS API protocol from data.
func (x *Signature) Unmarshal(b []byte) error {
	return neofsproto.Unmarshal(b, x)
}
