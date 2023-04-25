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

func assertStatusErr(tb testing.TB, res interface{ Status() apistatus.Status }) {
	require.IsType(tb, &statusErr, res.Status())
	require.Equal(tb, statusErr.Message(), res.Status().(*apistatus.ServerInternal).Message())
}

func newClient(signer neofscrypto.Signer, server neoFSAPIServer) *Client {
	var prm PrmInit
	prm.SetDefaultSigner(signer)

	var c Client
	c.Init(prm)
	c.setNeoFSAPIServer(server)

	return &c
}

func TestClient_DialContext(t *testing.T) {
	var c Client

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
