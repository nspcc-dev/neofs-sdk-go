package pool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

// NewEphemeralKey creates new private key used for NeoFS session. It'll be
// removed after refactoring to use neo-go crypto library.
func NewEphemeralKey() (*ecdsa.PrivateKey, error) {
	c := elliptic.P256()
	priv, x, y, err := elliptic.GenerateKey(c, rand.Reader)
	if err != nil {
		return nil, err
	}
	key := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: c,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(priv),
	}
	return key, nil
}
