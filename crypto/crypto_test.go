package neofscrypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/io"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	data := make([]byte, 512)
	//nolint:staticcheck
	rand.Read(data)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)

	var s neofscrypto.Signature

	for _, f := range []func() neofscrypto.Signer{
		func() neofscrypto.Signer {
			return neofsecdsa.Signer(k.PrivateKey)
		},
		func() neofscrypto.Signer {
			return neofsecdsa.SignerRFC6979(k.PrivateKey)
		},
		func() neofscrypto.Signer {
			return neofsecdsa.SignerWalletConnect(k.PrivateKey)
		},
	} {
		signer := f()

		err := s.Calculate(signer, data)
		require.NoError(t, err)

		m := s.ProtoMessage()

		require.NoError(t, s.FromProtoMessage(m))

		valid := s.Verify(data)
		require.True(t, valid, "type %T", signer)
	}
}

type nFailedSigner struct {
	neofscrypto.Signer
	n, count int
}

func (x *nFailedSigner) Sign(data []byte) ([]byte, error) {
	x.count++
	if x.count < x.n {
		return x.Signer.Sign(data)
	}
	return nil, errors.New("test signer forcefully fails")
}

func newNFailedSigner(s neofscrypto.Signer, n int) neofscrypto.Signer {
	return &nFailedSigner{Signer: s, n: n}
}

func verifyECDSAWithSHA512Signature(t testing.TB, pub *ecdsa.PublicKey, h, sig []byte) {
	require.Len(t, sig, 65)

	require.EqualValues(t, 4, sig[0])

	r := new(big.Int).SetBytes(sig[1:33])
	s := new(big.Int).SetBytes(sig[33:])

	p := elliptic.P256().Params().P
	require.Negative(t, r.Cmp(p))
	require.Negative(t, s.Cmp(p))

	require.True(t, ecdsa.Verify(pub, h, r, s))
}

func verifyECDSAWithSHA256RFC6979Signature(t testing.TB, pub *ecdsa.PublicKey, h, sig []byte) {
	require.Len(t, sig, 64)
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	require.True(t, ecdsa.Verify(pub, h, r, s))
}

func verifyWalletConnectSignature(t testing.TB, pub *ecdsa.PublicKey, data, sigSalt []byte) {
	require.Len(t, sigSalt, 80)

	data = base64.StdEncoding.AppendEncode(nil, data)
	sig, salt := sigSalt[:64], sigSalt[64:]
	sln := uint64(hex.EncodedLen(len(salt)) + len(data))

	// we could +io.GetVarSize(sln), but https://github.com/nspcc-dev/neo-go/issues/3766
	msg := make([]byte, 0, 4+binary.MaxVarintLen64)
	msg = append(msg, 0x01, 0x00, 0x01, 0xf0)
	msg = msg[:cap(msg)]
	msg = msg[:4+io.PutVarUint(msg[4:], sln)]

	msg = hex.AppendEncode(msg, salt)
	msg = append(msg, data...)
	msg = append(msg, 0x00, 0x00)

	h := sha256.Sum256(msg)
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	require.True(t, ecdsa.Verify(pub, h[:], r, s))
}
