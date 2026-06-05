package client

import (
	"crypto/ecdsa"
	"crypto/x509"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/stretchr/testify/require"
)

func TestClientMTLSConfig(t *testing.T) {
	pk, err := keys.NewPrivateKey()
	require.NoError(t, err)

	cfg, err := clientMTLSConfig(&pk.PrivateKey)
	require.NoError(t, err)

	require.Equal(t, clientMTLSServerName, cfg.ServerName)
	require.Equal(t, []string{"h2"}, cfg.NextProtos)
	require.True(t, cfg.InsecureSkipVerify)
	require.Len(t, cfg.Certificates, 1)

	cert, err := x509.ParseCertificate(cfg.Certificates[0].Certificate[0])
	require.NoError(t, err)
	certPub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	require.True(t, ok)
	require.True(t, certPub.Equal(&pk.PrivateKey.PublicKey),
		"certificate public key must equal the request-signing key")
}
