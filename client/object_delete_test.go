package client

import (
	"context"
	"fmt"
	"testing"

	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

type testDeleteObjectServer struct {
	unimplementedNeoFSAPIServer
}

func (x testDeleteObjectServer) deleteObject(context.Context, apiobject.DeleteRequest) (*apiobject.DeleteResponse, error) {
	id := oidtest.ID()
	var idV2 refs.ObjectID
	id.WriteToV2(&idV2)
	var addr refs.Address
	addr.SetObjectID(&idV2)
	var body apiobject.DeleteResponseBody
	body.SetTombstone(&addr)
	var resp apiobject.DeleteResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

func TestClient_ObjectDelete(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t, nil)

		_, err := c.ObjectDelete(context.Background(), cid.ID{}, oid.ID{}, nil, PrmObjectDelete{})
		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
