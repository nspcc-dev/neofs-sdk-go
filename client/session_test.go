package client

import (
	"context"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/session"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

// returns Client of Session service provided by given server.
func newTestSessionClient(t testing.TB, srv protosession.SessionServiceServer) *Client {
	return newClient(t, testService{desc: &protosession.SessionService_ServiceDesc, impl: srv})
}

type testCreateSessionServer struct {
	protosession.UnimplementedSessionServiceServer
	signer neofscrypto.Signer

	unsetID  bool
	unsetKey bool
}

func (m *testCreateSessionServer) Create(context.Context, *protosession.CreateRequest) (*protosession.CreateResponse, error) {
	resp := protosession.CreateResponse{
		Body: new(protosession.CreateResponse_Body),
	}

	if !m.unsetID {
		resp.Body.Id = []byte{1}
	}
	if !m.unsetKey {
		resp.Body.SessionKey = []byte{1}
	}

	var respV2 session.CreateResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	signer := m.signer
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	if err := signServiceMessage(signer, &respV2, nil); err != nil {
		return nil, err
	}

	return respV2.ToGRPCMessage().(*protosession.CreateResponse), nil
}

func TestClient_SessionCreate(t *testing.T) {
	ctx := context.Background()
	usr := usertest.User()

	var prmSessionCreate PrmSessionCreate
	prmSessionCreate.SetExp(1)

	t.Run("missing session id", func(t *testing.T) {
		srv := testCreateSessionServer{signer: usr, unsetID: true}
		c := newTestSessionClient(t, &srv)

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session id field in the response", err.Error())
	})

	t.Run("missing session key", func(t *testing.T) {
		srv := testCreateSessionServer{signer: usr, unsetKey: true}
		c := newTestSessionClient(t, &srv)

		result, err := c.SessionCreate(ctx, usr, prmSessionCreate)
		require.Nil(t, result)
		require.ErrorIs(t, err, ErrMissingResponseField)
		require.Equal(t, "missing session key field in the response", err.Error())
	})
}
