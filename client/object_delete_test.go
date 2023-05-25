package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func randOID(t *testing.T) oid.ID {
	var id oid.ID
	id.SetSHA256(randSHA256Checksum(t))

	return id
}

func randCID(t *testing.T) cid.ID {
	var id cid.ID
	id.SetSHA256(randSHA256Checksum(t))

	return id
}

func randSHA256Checksum(t *testing.T) (cs [sha256.Size]byte) {
	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	return
}

func TestClient_ObjectDelete(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t, nil, nil)

		_, err := c.ObjectDelete(context.Background(), cid.ID{}, oid.ID{}, PrmObjectDelete{})
		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
