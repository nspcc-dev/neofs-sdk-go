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
			err := c.AnnounceLocalTrust(ctx, 1, make([]reputation.Trust, 1), PrmAnnounceLocalTrust{})
			require.ErrorIs(t, err, ErrMissingSigner)
		})

		t.Run("intermediate", func(t *testing.T) {
			err := c.AnnounceIntermediateTrust(ctx, 1, reputation.PeerToPeerTrust{}, PrmAnnounceIntermediateTrust{})
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
}
