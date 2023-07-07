package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

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
