package neofsecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// PublicKey implements neofscrypto.PublicKey based on ECDSA.
//
// Supported schemes:
//  - neofscrypto.ECDSA_SHA512 (default)
//  - neofscrypto.ECDSA_DETERMINISTIC_SHA256
type PublicKey struct {
	deterministic bool

	key ecdsa.PublicKey
}

// MakeDeterministic makes PublicKey to use Deterministic ECDSA scheme
// (see neofscrypto.ECDSA_DETERMINISTIC_SHA256). By default,
// neofscrypto.ECDSA_SHA512 is used.
func (x *PublicKey) MakeDeterministic() {
	x.deterministic = true
}

// Decode decodes compressed binary representation of the PublicKey.
//
// See also Signer.EncodePublicKey.
func (x *PublicKey) Decode(data []byte) error {
	pub, err := keys.NewPublicKeyFromBytes(data, elliptic.P256())
	if err != nil {
		return err
	}

	x.key = (ecdsa.PublicKey)(*pub)

	return nil
}

// similar to elliptic.Unmarshal but without IsOnCurve check.
func unmarshalXY(data []byte) (x *big.Int, y *big.Int) {
	if len(data) != 65 {
		return
	} else if data[0] != 4 { // uncompressed form
		return
	}

	p := elliptic.P256().Params().P
	x = new(big.Int).SetBytes(data[1:33])
	y = new(big.Int).SetBytes(data[33:])

	if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
		x, y = nil, nil
	}

	return
}

// Verify verifies data signature calculated by algorithm depending on
// PublicKey state:
//  - Deterministic ECDSA with SHA-256 hashing if MakeDeterministic called
//  - ECDSA with SHA-512 hashing, otherwise
func (x PublicKey) Verify(data, signature []byte) bool {
	if x.deterministic {
		h := sha256.Sum256(data)
		return (*keys.PublicKey)(&x.key).Verify(signature, h[:])
	}

	h := sha512.Sum512(data)
	r, s := unmarshalXY(signature)

	return r != nil && s != nil && ecdsa.Verify(&x.key, h[:], r, s)
}
