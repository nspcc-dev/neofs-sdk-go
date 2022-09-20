package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

/*
File contains common functionality used for client package testing.
*/

var key, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

var statusErr apistatus.ServerInternal

func init() {
	statusErr.SetMessage("test status error")
}

func assertStatusErr(tb testing.TB, res interface{ Status() apistatus.Status }) {
	require.IsType(tb, &statusErr, res.Status())
	require.Equal(tb, statusErr.Message(), res.Status().(*apistatus.ServerInternal).Message())
}

func newClient(server neoFSAPIServer) *Client {
	var prm PrmInit
	prm.SetDefaultPrivateKey(*key)

	var c Client
	c.Init(prm)
	c.setNeoFSAPIServer(server)

	return &c
}
