package neofscrypto

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// Signature represents a confirmation of data integrity received by the
// digital signature mechanism.
//
// Signature is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Signature
// message. See ReadFromV2 / WriteToV2 methods.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Signature(refs.Signature{}) // not recommended
type Signature refs.Signature

// ReadFromV2 reads Signature from the refs.Signature message.
//
// See also WriteToV2.
func (x *Signature) ReadFromV2(m refs.Signature) {
	*x = Signature(m)
}

// WriteToV2 writes Signature to the refs.Signature message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x Signature) WriteToV2(m *refs.Signature) {
	*m = (refs.Signature)(x)
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

	pub := signer.Public()

	key := make([]byte, pub.MaxEncodedSize())
	key = key[:pub.Encode(key)]

	m := (*refs.Signature)(x)

	m.SetScheme(refs.SignatureScheme(signer.Scheme()))
	m.SetSign(signature)
	m.SetKey(key)

	return nil
}

// Verify verifies data signature using encoded public key. True means valid
// signature.
//
// Verify fails if signature scheme is not supported (see RegisterScheme).
//
// See also Calculate.
func (x Signature) Verify(data []byte) bool {
	m := (*refs.Signature)(&x)

	f, ok := publicKeys[Scheme(m.GetScheme())]
	if !ok {
		return false
	}

	key := f()

	err := key.Decode(m.GetKey())
	if err != nil {
		return false
	}

	return key.Verify(data, m.GetSign())
}
