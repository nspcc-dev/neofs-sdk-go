package client

import (
	"errors"
	"testing"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/stretchr/testify/require"
)

const anyValidURI = "grpc://localhost:8080"

type disposableSigner struct {
	neofscrypto.Signer
	signed bool
}

// returns signer which uses s only once and then only fails.
func newDisposableSigner(s neofscrypto.Signer) neofscrypto.Signer {
	return &disposableSigner{Signer: s}
}

func (x *disposableSigner) Sign(data []byte) ([]byte, error) {
	if x.signed {
		return nil, errors.New("already signed")
	}
	x.signed = true
	return x.Signer.Sign(data)
}

func TestParseURI(t *testing.T) {
	addr, isTLS, err := parseURI(anyValidURI)
	require.NoError(t, err)
	require.False(t, isTLS)
	require.Equal(t, "localhost:8080", addr)
}
