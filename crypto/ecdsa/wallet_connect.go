package neofsecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/io"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

const saltLen = 16

// SignerWalletConnect is similar to SignerRFC6979 with 2 changes:
// 1. The data is base64 encoded before signing/verifying.
// 2. The signature is a concatenation of the signature itself and 16-byte salt.
//
// Instances MUST be initialized from ecdsa.PrivateKey using type conversion.
type SignerWalletConnect ecdsa.PrivateKey

// Scheme returns neofscrypto.ECDSA_WALLETCONNECT.
// Implements neofscrypto.Signer.
func (x SignerWalletConnect) Scheme() neofscrypto.Scheme {
	return neofscrypto.ECDSA_WALLETCONNECT
}

// Sign signs data using ECDSA algorithm with SHA-512 hashing.
// Implements neofscrypto.Signer.
func (x SignerWalletConnect) Sign(data []byte) ([]byte, error) {
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(b64, data)
	var salt [saltLen]byte
	_, err := rand.Read(salt[:])
	if err != nil {
		return nil, fmt.Errorf("randomize salt: %w", err)
	}
	sig, err := SignerRFC6979(x).Sign(saltMessageWalletConnect(b64, salt[:]))
	if err != nil {
		return nil, err
	}
	return append(sig, salt[:]...), nil
}

// Public initializes PublicKey and returns it as neofscrypto.PublicKey.
// Implements neofscrypto.Signer.
func (x SignerWalletConnect) Public() neofscrypto.PublicKey {
	return (*PublicKeyWalletConnect)(&x.PublicKey)
}

// PublicKeyWalletConnect is a wrapper over ecdsa.PublicKey used for NeoFS needs.
// Provides neofscrypto.PublicKey interface.
//
// Instances MUST be initialized from ecdsa.PublicKey using type conversion.
type PublicKeyWalletConnect ecdsa.PublicKey

// MaxEncodedSize returns size of the compressed ECDSA public key.
func (x PublicKeyWalletConnect) MaxEncodedSize() int {
	return 33
}

// Encode encodes ECDSA public key in compressed form into buf.
// Uses exactly MaxEncodedSize bytes of the buf.
//
// Encode panics if buf length is less than MaxEncodedSize.
//
// See also Decode.
func (x PublicKeyWalletConnect) Encode(buf []byte) int {
	if len(buf) < 33 {
		panic(fmt.Sprintf("too short buffer %d", len(buf)))
	}

	return copy(buf, (*keys.PublicKey)(&x).Bytes())
}

// Decode decodes compressed binary representation of the PublicKeyWalletConnect.
//
// See also Encode.
func (x *PublicKeyWalletConnect) Decode(data []byte) error {
	pub, err := keys.NewPublicKeyFromBytes(data, elliptic.P256())
	if err != nil {
		return err
	}

	*x = (PublicKeyWalletConnect)(*pub)

	return nil
}

// Verify verifies data signature calculated by ECDSA algorithm with SHA-512 hashing.
func (x PublicKeyWalletConnect) Verify(data, signature []byte) bool {
	if len(signature) != keys.SignatureLen+saltLen {
		return false
	}
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(b64, data)
	return verifyWalletConnect((*ecdsa.PublicKey)(&x), b64, signature[:keys.SignatureLen], signature[keys.SignatureLen:])
}

func verifyWalletConnect(key *ecdsa.PublicKey, data, sig, salt []byte) bool {
	h := sha256.Sum256(saltMessageWalletConnect(data, salt))
	var r, s big.Int
	r.SetBytes(sig[:keys.SignatureLen/2])
	s.SetBytes(sig[keys.SignatureLen/2:])
	return ecdsa.Verify(key, h[:], &r, &s)
}

// saltMessageWalletConnect calculates signed message for given data and salt
// according to WalletConnect.
func saltMessageWalletConnect(data, salt []byte) []byte {
	saltedLen := hex.EncodedLen(len(salt)) + len(data)
	b := make([]byte, 4+getVarIntSize(saltedLen)+saltedLen+2)
	b[0], b[1], b[2], b[3] = 0x01, 0x00, 0x01, 0xf0
	n := 4 + io.PutVarUint(b[4:], uint64(saltedLen))
	n += hex.Encode(b[n:], salt)
	n += copy(b[n:], data)
	b[n], b[n+1] = 0x00, 0x00
	return b
}

// copy-paste from neo-go.
func getVarIntSize(value int) int {
	if value < 0xFD {
		return 1
	} else if value <= 0xFFFF {
		return 3
	}
	return 5
}
