package client

import (
	"context"
	"errors"
	"testing"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type sessionAPIServer struct {
	signer  neofscrypto.Signer
	setBody func(body *session.CreateResponseBody)
}

func (m sessionAPIServer) netMapSnapshot(context.Context, v2netmap.SnapshotRequest) (*v2netmap.SnapshotResponse, error) {
	return nil, errors.New("unimplemented")
}

func (m sessionAPIServer) createSession(context.Context, session.CreateRequest) (*session.CreateResponse, error) {
	var body session.CreateResponseBody
	m.setBody(&body)

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
		c.setNeoFSAPIServer(&sessionAPIServer{signer: usr, setBody: func(body *session.CreateResponseBody) {
			body.SetSessionKey([]byte{1})
		}})

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session id field in the response", err.Error())
	})

	t.Run("missing session key", func(t *testing.T) {
		c.setNeoFSAPIServer(&sessionAPIServer{signer: usr, setBody: func(body *session.CreateResponseBody) {
			body.SetID([]byte{1})
		}})

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session key field in the response", err.Error())
	})
}
