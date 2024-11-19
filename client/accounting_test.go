package client

import (
	"context"
	"fmt"
	"testing"

	apiaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/stretchr/testify/require"
)

type testGetBalanceServer struct {
	unimplementedNeoFSAPIServer
}

func (x testGetBalanceServer) getBalance(context.Context, apiaccounting.BalanceRequest) (*apiaccounting.BalanceResponse, error) {
	var body apiaccounting.BalanceResponseBody
	body.SetBalance(new(apiaccounting.Decimal))
	var resp apiaccounting.BalanceResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

func TestClient_BalanceGet(t *testing.T) {
	c := newClient(t, nil)
	ctx := context.Background()

	t.Run("missing", func(t *testing.T) {
		t.Run("account", func(t *testing.T) {
			_, err := c.BalanceGet(ctx, PrmBalanceGet{})
			require.ErrorIs(t, err, ErrMissingAccount)
		})
	})
}
