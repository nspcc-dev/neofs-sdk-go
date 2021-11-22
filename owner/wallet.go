package owner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"errors"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

// NEO3Wallet represents NEO3 wallet address.
type NEO3Wallet [NEO3WalletSize]byte

// NEO3WalletSize contains size of neo3 wallet.
const NEO3WalletSize = 25

const addressPrefixN3 = 0x35

// ErrEmptyPublicKey when PK passed to Verify method is nil.
var ErrEmptyPublicKey = errors.New("empty public key")

// NEO3WalletFromPublicKey converts public key to NEO3 wallet address.
func NEO3WalletFromPublicKey(key *ecdsa.PublicKey) (*NEO3Wallet, error) {
	if key == nil {
		return nil, ErrEmptyPublicKey
	}

	b := elliptic.MarshalCompressed(key.Curve, key.X, key.Y)
	script := []byte{0x0C /* PUSHDATA1 */, byte(len(b)) /* 33 */}
	script = append(script, b...)
	script = append(script, 0x41 /* SYSCALL */)
	h := sha256.Sum256([]byte("System.Crypto.CheckSig"))
	script = append(script, h[:4]...)

	h1 := sha256.Sum256(script)
	rw := ripemd160.New()
	rw.Write(h1[:])
	h160 := rw.Sum(nil)

	var w NEO3Wallet
	w[0] = addressPrefixN3
	copy(w[1:21], h160)
	copy(w[21:], addressChecksum(w[:21]))

	return &w, nil
}

func addressChecksum(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:4]
}

// String implements fmt.Stringer.
func (w *NEO3Wallet) String() string {
	if w != nil {
		return base58.Encode(w[:])
	}

	return ""
}

// Bytes returns slice of NEO3 wallet address bytes.
func (w *NEO3Wallet) Bytes() []byte {
	if w != nil {
		return w[:]
	}

	return nil
}
