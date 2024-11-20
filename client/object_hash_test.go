package client

import (
	"context"
	"fmt"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

type testHashObjectPayloadRangesServer struct {
	protoobject.UnimplementedObjectServiceServer
}

func (x *testHashObjectPayloadRangesServer) GetRangeHash(context.Context, *protoobject.GetRangeHashRequest) (*protoobject.GetRangeHashResponse, error) {
	resp := protoobject.GetRangeHashResponse{
		Body: &protoobject.GetRangeHashResponse_Body{
			HashList: [][]byte{{1}},
		},
	}

	var respV2 v2object.GetRangeHashResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoobject.GetRangeHashResponse), nil
}

func TestClient_ObjectHash(t *testing.T) {
	c := newClient(t)

	t.Run("missing signer", func(t *testing.T) {
		var reqBody v2object.GetRangeHashRequestBody
		reqBody.SetRanges(make([]v2object.Range, 1))

		_, err := c.ObjectHash(context.Background(), cid.ID{}, oid.ID{}, nil, PrmObjectHash{
			body: reqBody,
		})

		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
