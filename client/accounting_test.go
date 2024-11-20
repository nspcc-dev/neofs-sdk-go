package client

import (
	"context"
	"fmt"
	"testing"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	protoaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting/grpc"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

// returns Client of Accounting service provided by given server.
func newTestAccountingClient(t testing.TB, srv protoaccounting.AccountingServiceServer) *Client {
	return newClient(t, testService{desc: &protoaccounting.AccountingService_ServiceDesc, impl: srv})
}

type testGetBalanceServer struct {
	protoaccounting.UnimplementedAccountingServiceServer
}

func (x *testGetBalanceServer) Balance(context.Context, *protoaccounting.BalanceRequest) (*protoaccounting.BalanceResponse, error) {
	resp := protoaccounting.BalanceResponse{
		Body: &protoaccounting.BalanceResponse_Body{
			Balance: new(protoaccounting.Decimal),
		},
	}

	var respV2 v2accounting.BalanceResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoaccounting.BalanceResponse), nil
}

func TestClient_BalanceGet(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	t.Run("missing", func(t *testing.T) {
		t.Run("account", func(t *testing.T) {
			_, err := c.BalanceGet(ctx, PrmBalanceGet{})
			require.ErrorIs(t, err, ErrMissingAccount)
		})
	})
}
