package usertest

import (
	"crypto/ecdsa"
	"errors"

	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ID returns random user.ID.
func ID() user.ID {
	return user.NewFromScriptHash(testutil.RandScriptHash())
}

// OtherID returns random user.ID other than any given one.
func OtherID(vs ...user.ID) user.ID {
loop:
	for {
		v := ID()
		for i := range vs {
			if v == vs[i] {
				continue loop
			}
		}
		return v
	}
}

// IDs returns n random user.ID instances.
func IDs(n int) []user.ID {
	res := make([]user.ID, n)
	for i := range res {
		res[i] = ID()
	}
	return res
}

// UserSigner represents NeoFS user credentials.
type UserSigner struct {
	ID              user.ID
	ECDSAPrivateKey ecdsa.PrivateKey
	// Components calculated for ECDSAPrivateKey.
	PublicKeyBytes []byte
	user.Signer
	RFC6979       user.Signer
	WalletConnect user.Signer
}

func User() UserSigner {
	cs := neofscryptotest.Signer()
	s := user.NewAutoIDSigner(cs.ECDSAPrivateKey)
	return UserSigner{
		ID:              s.UserID(),
		ECDSAPrivateKey: cs.ECDSAPrivateKey,
		PublicKeyBytes:  cs.PublicKeyBytes,
		Signer:          s,
		RFC6979:         user.NewAutoIDSignerRFC6979(cs.ECDSAPrivateKey),
		WalletConnect:   user.NewSigner(cs.WalletConnect, s.UserID()),
	}
}

type failedSigner struct {
	user.Signer
}

func (x failedSigner) Sign([]byte) ([]byte, error) { return nil, errors.New("[test] failed to sign") }

// FailSigner returns wraps s to always return error from Sign method.
func FailSigner(s user.Signer) user.Signer {
	return failedSigner{s}
}
