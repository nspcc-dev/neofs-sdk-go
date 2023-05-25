package client

import (
	"context"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func TestClient_BalanceGet(t *testing.T) {
	c := newClient(t, nil, nil)
	ctx := context.Background()

	t.Run("missing", func(t *testing.T) {
		t.Run("account", func(t *testing.T) {
			_, err := c.BalanceGet(ctx, PrmBalanceGet{})
			require.ErrorIs(t, err, ErrMissingAccount)
		})

		t.Run("signer", func(t *testing.T) {
			var prm PrmBalanceGet
			prm.SetAccount(user.ID{})

			_, err := c.BalanceGet(ctx, prm)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
}
