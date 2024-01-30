package client

import (
	"context"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/stretchr/testify/require"
)

/*
File contains common functionality used for client package testing.
*/

var statusErr apistatus.ServerInternal

func init() {
	statusErr.SetMessage("test status error")
}

func newClient(t *testing.T, server neoFSAPIServer) *Client {
	var prm PrmInit

	c, err := New(prm)
	require.NoError(t, err)
	c.setNeoFSAPIServer(server)

	return c
}

func TestClient_DialContext(t *testing.T) {
	var prmInit PrmInit

	c, err := New(prmInit)
	require.NoError(t, err)

	// try to connect to any host
	var prm PrmDial
	prm.SetServerURI("localhost:8080")

	assert := func(ctx context.Context, errExpected error) {
		// use the particular context
		prm.SetContext(ctx)

		// expect particular context error according to Dial docs
		require.ErrorIs(t, c.Dial(prm), errExpected)
	}

	// create pre-abandoned context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert(ctx, context.Canceled)

	// create "pre-deadlined" context
	ctx, cancel = context.WithTimeout(context.Background(), 0)
	defer cancel()

	assert(ctx, context.DeadlineExceeded)
}

type nopPublicKey struct{}

func (x nopPublicKey) MaxEncodedSize() int     { return 10 }
func (x nopPublicKey) Encode(buf []byte) int   { return copy(buf, "public_key") }
func (x nopPublicKey) Decode([]byte) error     { return nil }
func (x nopPublicKey) Verify(_, _ []byte) bool { return true }

type nopSigner struct{}

func (nopSigner) Scheme() neofscrypto.Scheme      { return neofscrypto.ECDSA_SHA512 }
func (nopSigner) Sign([]byte) ([]byte, error)     { return []byte("signature"), nil }
func (x nopSigner) Public() neofscrypto.PublicKey { return nopPublicKey{} }
