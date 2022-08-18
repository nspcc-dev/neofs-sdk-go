package neofsecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// Signer wraps ecdsa.PrivateKey and represents signer based on ECDSA with
// SHA-512 hashing. Provides neofscrypto.Signer interface.
//
// Instances MUST be initialized from ecdsa.PrivateKey using type conversion.
type Signer ecdsa.PrivateKey

// Scheme returns neofscrypto.ECDSA_SHA512.
// Implements neofscrypto.Signer.
func (x Signer) Scheme() neofscrypto.Scheme {
	return neofscrypto.ECDSA_SHA512
}

// Sign signs data using ECDSA algorithm with SHA-512 hashing.
// Implements neofscrypto.Signer.
func (x Signer) Sign(data []byte) ([]byte, error) {
	h := sha512.Sum512(data)
	r, s, err := ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(&x), h[:])
	if err != nil {
		return nil, err
	}

	params := elliptic.P256().Params()
	curveOrderByteSize := params.P.BitLen() / 8

	buf := make([]byte, 1+curveOrderByteSize*2)
	buf[0] = 4

	_ = r.FillBytes(buf[1 : 1+curveOrderByteSize])
	_ = s.FillBytes(buf[1+curveOrderByteSize:])

	return buf, nil
}

// Public initializes PublicKey and returns it as neofscrypto.PublicKey.
// Implements neofscrypto.Signer.
func (x Signer) Public() neofscrypto.PublicKey {
	return (*PublicKey)(&x.PublicKey)
}

// SignerRFC6979 wraps ecdsa.PrivateKey and represents signer based on deterministic
// ECDSA with SHA-256 hashing (RFC 6979). Provides neofscrypto.Signer interface.
//
// Instances SHOULD be initialized from ecdsa.PrivateKey using type conversion.
type SignerRFC6979 ecdsa.PrivateKey

// Scheme returns neofscrypto.ECDSA_DETERMINISTIC_SHA256.
// Implements neofscrypto.Signer.
func (x SignerRFC6979) Scheme() neofscrypto.Scheme {
	return neofscrypto.ECDSA_DETERMINISTIC_SHA256
}

// Sign signs data using deterministic ECDSA algorithm with SHA-256 hashing.
// Implements neofscrypto.Signer.
//
// See also RFC 6979.
func (x SignerRFC6979) Sign(data []byte) ([]byte, error) {
	p := keys.PrivateKey{PrivateKey: (ecdsa.PrivateKey)(x)}
	return p.Sign(data), nil
}

// Public initializes PublicKeyRFC6979 and returns it as neofscrypto.PublicKey.
// Implements neofscrypto.Signer.
func (x SignerRFC6979) Public() neofscrypto.PublicKey {
	return (*PublicKeyRFC6979)(&x.PublicKey)
}
