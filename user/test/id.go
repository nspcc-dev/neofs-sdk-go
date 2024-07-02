package usertest

import (
	"crypto/ecdsa"
	"errors"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

// ID returns random user.ID.
func ID() user.ID {
	var h util.Uint160
	//nolint:staticcheck
	rand.Read(h[:])
	var res user.ID
	res.SetScriptHash(h)
	return res
}

// OtherID returns random user.ID other than any given one.
func OtherID(vs ...user.ID) user.ID {
loop:
	for {
		v := ID()
		for i := range vs {
			if v.Equals(vs[i]) {
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

// SignedComponent contains data to be signed and its signature to be verified.
type SignedComponent interface {
	SignedData() []byte
	Sign(user.Signer) error
	VerifySignature() bool
}

// TestSignedData SignedDataComponentUser signing and verification of
// [SignedComponent.SignedData].
func TestSignedData(tb testing.TB, signer user.Signer, cmp SignedComponent) {
	data := cmp.SignedData()

	sig, err := signer.Sign(data)
	require.NoError(tb, err)

	static := neofscrypto.NewStaticSigner(signer.Scheme(), sig, signer.Public())

	err = cmp.Sign(user.NewSigner(static, signer.UserID()))
	require.NoError(tb, err)

	require.True(tb, cmp.VerifySignature())
}
