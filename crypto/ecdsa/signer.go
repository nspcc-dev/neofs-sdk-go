package neofsecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// Signer implements neofscrypto.Signer based on ECDSA.
//
// Supported schemes:
//  - neofscrypto.ECDSA_SHA512 (default)
//  - neofscrypto.ECDSA_DETERMINISTIC_SHA256
//
// Instances MUST be initialized with ecdsa.PrivateKey using SetKey.
type Signer struct {
	deterministic bool

	key ecdsa.PrivateKey
}

// SetKey specifies ecdsa.PrivateKey to be used for ECDSA signature calculation.
func (x *Signer) SetKey(key ecdsa.PrivateKey) {
	x.key = key
}

// MakeDeterministic makes Signer to use Deterministic ECDSA scheme
// (see neofscrypto.ECDSA_DETERMINISTIC_SHA256). By default,
// neofscrypto.ECDSA_SHA512 is used.
func (x *Signer) MakeDeterministic() {
	x.deterministic = true
}

// Scheme returns signature scheme depending on Signer state:
//  - neofscrypto.ECDSA_DETERMINISTIC_SHA256 if MakeDeterministic called
//  - neofscrypto.ECDSA_SHA512 otherwise.
func (x Signer) Scheme() neofscrypto.Scheme {
	if x.deterministic {
		return neofscrypto.ECDSA_DETERMINISTIC_SHA256
	}

	return neofscrypto.ECDSA_SHA512
}

// Sign signs data with algorithm depending on Signer state:
//  - Deterministic ECDSA with SHA-256 hashing if MakeDeterministic called
//  - ECDSA with SHA-512 hashing, otherwise
func (x Signer) Sign(data []byte) ([]byte, error) {
	if x.deterministic {
		p := keys.PrivateKey{PrivateKey: x.key}
		return p.Sign(data), nil
	}

	h := sha512.Sum512(data)
	r, s, err := ecdsa.Sign(rand.Reader, &x.key, h[:])
	if err != nil {
		return nil, err
	}

	return elliptic.Marshal(elliptic.P256(), r, s), nil
}

// MaxPublicKeyEncodedSize returns size of the compressed ECDSA public key
// of the Signer.
func (x Signer) MaxPublicKeyEncodedSize() int {
	return 33
}

// EncodePublicKey encodes ECDSA public key of the Signer in compressed form
// into buf. Uses exactly MaxPublicKeyEncodedSize bytes of the buf.
//
// EncodePublicKey panics if buf length is less than MaxPublicKeyEncodedSize.
//
// See also PublicKey.Decode.
func (x Signer) EncodePublicKey(buf []byte) int {
	if len(buf) < 33 {
		panic(fmt.Sprintf("too short buffer %d", len(buf)))
	}

	return copy(buf, (*keys.PublicKey)(&x.key.PublicKey).Bytes())
}