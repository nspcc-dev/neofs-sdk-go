package client

import (
	"context"
	"fmt"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

type testHashObjectPayloadRangesServer struct {
	unimplementedNeoFSAPIServer
}

func (x *testHashObjectPayloadRangesServer) hashObjectPayloadRanges(context.Context, v2object.GetRangeHashRequest) (*v2object.GetRangeHashResponse, error) {
	var body v2object.GetRangeHashResponseBody
	body.SetHashList([][]byte{{1}})
	var resp v2object.GetRangeHashResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

func TestClient_ObjectHash(t *testing.T) {
	c := newClient(t, nil)

	t.Run("missing signer", func(t *testing.T) {
		var reqBody v2object.GetRangeHashRequestBody
		reqBody.SetRanges(make([]v2object.Range, 1))

		_, err := c.ObjectHash(context.Background(), cid.ID{}, oid.ID{}, nil, PrmObjectHash{
			body: reqBody,
		})

		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
