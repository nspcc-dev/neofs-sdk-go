package client

import (
	"context"
	"fmt"
	"testing"

	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

type testDeleteObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
}

func (x *testDeleteObjectServer) Delete(context.Context, *protoobject.DeleteRequest) (*protoobject.DeleteResponse, error) {
	id := oidtest.ID()
	resp := protoobject.DeleteResponse{
		Body: &protoobject.DeleteResponse_Body{
			Tombstone: &protorefs.Address{
				ObjectId: &protorefs.ObjectID{Value: id[:]},
			},
		},
	}

	var respV2 apiobject.DeleteResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoobject.DeleteResponse), nil
}

func TestClient_ObjectDelete(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t)

		_, err := c.ObjectDelete(context.Background(), cid.ID{}, oid.ID{}, nil, PrmObjectDelete{})
		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
