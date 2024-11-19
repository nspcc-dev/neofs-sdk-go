package client

import (
	"context"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type sessionAPIServer struct {
	unimplementedNeoFSAPIServer
	signer neofscrypto.Signer

	id  []byte
	key []byte
}

func (m sessionAPIServer) createSession(context.Context, session.CreateRequest) (*session.CreateResponse, error) {
	var body session.CreateResponseBody
	body.SetID(m.id)
	body.SetSessionKey(m.key)

	var resp session.CreateResponse
	resp.SetBody(&body)

	if err := signServiceMessage(m.signer, &resp, nil); err != nil {
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
		c.setNeoFSAPIServer(&sessionAPIServer{signer: usr, key: []byte{1}})

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session id field in the response", err.Error())
	})

	t.Run("missing session key", func(t *testing.T) {
		c.setNeoFSAPIServer(&sessionAPIServer{signer: usr, id: []byte{1}})

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session key field in the response", err.Error())
	})
}
