package owner

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// NEO3Wallet represents NEO3 wallet address.
type NEO3Wallet [NEO3WalletSize]byte

// NEO3WalletSize contains size of neo3 wallet.
const NEO3WalletSize = 25

// ErrEmptyPublicKey when PK passed to Verify method is nil.
var ErrEmptyPublicKey = errors.New("empty public key")

// NEO3WalletFromPublicKey converts public key to NEO3 wallet address.
func NEO3WalletFromPublicKey(key *ecdsa.PublicKey) (*NEO3Wallet, error) {
	if key == nil {
		return nil, ErrEmptyPublicKey
	}

	neoPublicKey := (*keys.PublicKey)(key)

	d, err := base58.Decode(neoPublicKey.Address())
	if err != nil {
		return nil, fmt.Errorf("can't decode neo3 address from key: %w", err)
	}

	w := new(NEO3Wallet)
	copy(w.Bytes(), d)

	return w, nil
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
