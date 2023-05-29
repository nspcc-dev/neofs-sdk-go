package client

import (
	"context"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestClient_ObjectHash(t *testing.T) {
	c := newClient(t, nil, nil)

	t.Run("missing signer", func(t *testing.T) {
		var reqBody v2object.GetRangeHashRequestBody
		reqBody.SetRanges(make([]v2object.Range, 1))

		_, err := c.ObjectHash(context.Background(), cid.ID{}, oid.ID{}, PrmObjectHash{
			body: reqBody,
		})

		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
