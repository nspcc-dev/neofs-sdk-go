package neofscrypto_test

import (
	"errors"
	"testing"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/stretchr/testify/require"
)

type mockSignerV2 struct {
	err error
}

func (m mockSignerV2) SignData([]byte) (neofscrypto.Signature, error) {
	return neofscrypto.Signature{}, m.err
}

func TestOverlapSigner(t *testing.T) {
	s2 := mockSignerV2{err: errors.New("any")}

	s := neofscrypto.OverlapSigner(s2)
	require.NotNil(t, s)

	require.Panics(t, func() { s.Scheme() })
	require.Panics(t, func() { s.Public() })
	require.Panics(t, func() { _, _ = s.Sign([]byte("any")) })

	require.Implements(t, (*neofscrypto.SignerV2)(nil), s)
	_, err := s.(neofscrypto.SignerV2).SignData([]byte("any"))
	require.Equal(t, s2.err, err)
}
