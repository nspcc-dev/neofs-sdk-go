package client

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"time"
)

// clientMTLSServerName is the SNI for client-side mTLS; MUST match with node.
const clientMTLSServerName = "neofs.client.mtls"

// clientMTLSConfig builds the client-mTLS config. The cert is self-signed with
// key so its public key equals the request author key the node matches; key MUST
// be the one used to sign object requests. InsecureSkipVerify is fine because
// node identity stays authenticated by signed responses.
func clientMTLSConfig(key *ecdsa.PrivateKey) (*tls.Config, error) {
	cert, err := selfSignedCert(key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		ServerName:         clientMTLSServerName,
		MinVersion:         tls.VersionTLS12,
		NextProtos:         []string{"h2"},
		InsecureSkipVerify: true,
	}, nil
}

// selfSignedCert produces a self-signed client certificate for key.
func selfSignedCert(key *ecdsa.PrivateKey) (tls.Certificate, error) {
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(100 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create x509 certificate: %w", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}
