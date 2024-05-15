package usertest

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ID returns random user.ID.
func ID() user.ID {
	var h util.Uint160
	rand.Read(h[:])
	return user.NewID(h)
}

// ChangeID returns user ID other than the given one.
func ChangeID(id user.ID) user.ID {
	id[0]++
	return id
}

// NIDs returns n random user.ID instances.
func NIDs(n int) []user.ID {
	res := make([]user.ID, n)
	for i := range res {
		res[i] = ID()
	}
	return res
}

// User represents NeoFS user credentials.
type User struct {
	user.Signer
	SignerRFC6979, SignerWalletConnect user.Signer

	ID             user.ID
	PrivateKey     ecdsa.PrivateKey
	PublicKeyBytes []byte
}

// TwoUsers returns two packs of different static user ECDSA credentials.
func TwoUsers() (User, User) {
	const strUser1 = "NPo3isCPDA6S7EVfpATq1NaBkzVhxWX9FQ"
	const hexPrivKey1 = "ebd6154e7e1b85050647a480ee3a97355a6ea7fe6e80ce7b4c27dbf00d599a2a"
	const hexPubKey1 = "03ffb26e9e499ae96024730b5b73d32182d97a46a51a68ffcf0a5db6b56a67057f"
	const strUser2 = "NZrDLV77VcxTCWhpjR2DuD8zg5iyn8rYtW"
	const hexPrivKey2 = "23d2ba98afa05c06dc9186efc40c568b0333c33f790865ef68e7aa20cde5354d"
	const hexPubKey2 = "021e6c4449951ff3170a92ba9d9af88facceedf1dfcb42d0d2afa0e72b3d67c372"

	k1, err := keys.NewPrivateKeyFromHex(hexPrivKey1)
	if err != nil {
		panic(fmt.Errorf("unexpected decode private key failure: %w", err))
	}
	k2, err := keys.NewPrivateKeyFromHex(hexPrivKey2)
	if err != nil {
		panic(fmt.Errorf("unexpected decode private key failure: %w", err))
	}

	bPubKey1, err := hex.DecodeString(hexPubKey1)
	if err != nil {
		panic(fmt.Errorf("unexpected decode HEX failure: %w", err))
	}
	bPubKey2, err := hex.DecodeString(hexPubKey2)
	if err != nil {
		panic(fmt.Errorf("unexpected decode HEX failure: %w", err))
	}

	var usr1, usr2 user.ID
	if err = usr1.DecodeString(strUser1); err != nil {
		panic(fmt.Errorf("unexpected decode string user ID failure: %w", err))
	}
	if err = usr2.DecodeString(strUser2); err != nil {
		panic(fmt.Errorf("unexpected decode string user ID failure: %w", err))
	}

	return User{
			ID:                  user.ResolveFromECDSAPublicKey(k1.PrivateKey.PublicKey),
			PrivateKey:          k1.PrivateKey,
			PublicKeyBytes:      bPubKey1,
			Signer:              user.NewAutoIDSigner(k1.PrivateKey),
			SignerRFC6979:       user.NewAutoIDSignerRFC6979(k1.PrivateKey),
			SignerWalletConnect: user.NewSigner(neofsecdsa.SignerWalletConnect(k1.PrivateKey), usr1),
		}, User{
			ID:                  user.ResolveFromECDSAPublicKey(k2.PrivateKey.PublicKey),
			PrivateKey:          k2.PrivateKey,
			PublicKeyBytes:      bPubKey2,
			Signer:              user.NewAutoIDSigner(k2.PrivateKey),
			SignerRFC6979:       user.NewAutoIDSignerRFC6979(k2.PrivateKey),
			SignerWalletConnect: user.NewSigner(neofsecdsa.SignerWalletConnect(k2.PrivateKey), usr2),
		}
}

type failedSigner struct {
	user.Signer
}

func (x failedSigner) Sign([]byte) ([]byte, error) { return nil, errors.New("failed to sign") }

// FailSigner returns wraps s to always return error from Sign method.
func FailSigner(s user.Signer) user.Signer {
	return failedSigner{s}
}
