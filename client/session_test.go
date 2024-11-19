package client

import (
	"context"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type testCreateSessionServer struct {
	unimplementedNeoFSAPIServer
	signer neofscrypto.Signer

	unsetID  bool
	unsetKey bool
}

func (m testCreateSessionServer) createSession(context.Context, session.CreateRequest) (*session.CreateResponse, error) {
	var body session.CreateResponseBody
	if !m.unsetID {
		body.SetID([]byte{1})
	}
	if !m.unsetKey {
		body.SetSessionKey([]byte{2})
	}

	var resp session.CreateResponse
	resp.SetBody(&body)

	signer := m.signer
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	if err := signServiceMessage(signer, &resp, nil); err != nil {
		return nil, err
	}

	return &resp, nil
}

func TestClient_SessionCreate(t *testing.T) {
	ctx := context.Background()
	usr := usertest.User()

	c := newClient(t, nil)

	var prmSessionCreate PrmSessionCreate
	prmSessionCreate.SetExp(1)

	t.Run("missing session id", func(t *testing.T) {
		c.setNeoFSAPIServer(&testCreateSessionServer{signer: usr, unsetID: true})

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session id field in the response", err.Error())
	})

	t.Run("missing session key", func(t *testing.T) {
		c.setNeoFSAPIServer(&testCreateSessionServer{signer: usr, unsetKey: true})

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session key field in the response", err.Error())
	})
}
