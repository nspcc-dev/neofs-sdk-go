package client

import (
	"context"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/stretchr/testify/require"
)

func TestClient_Reputation(t *testing.T) {
	c := newClient(t, nil, nil)
	ctx := context.Background()

	t.Run("missing signer", func(t *testing.T) {
		t.Run("local", func(t *testing.T) {
			err := c.AnnounceLocalTrust(ctx, PrmAnnounceLocalTrust{epoch: 1, trusts: make([]reputation.Trust, 1)})
			require.ErrorIs(t, err, ErrMissingSigner)
		})

		t.Run("intermediate", func(t *testing.T) {
			err := c.AnnounceIntermediateTrust(ctx, PrmAnnounceIntermediateTrust{epoch: 1, trustSet: true})
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
}
