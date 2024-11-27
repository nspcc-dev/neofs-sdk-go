package client

import (
	"context"
	"fmt"
	"testing"

	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

type testGetObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
}

func (x *testGetObjectServer) Get(_ *protoobject.GetRequest, stream protoobject.ObjectService_GetServer) error {
	resp := protoobject.GetResponse{
		Body: &protoobject.GetResponse_Body{
			ObjectPart: &protoobject.GetResponse_Body_Init_{
				Init: new(protoobject.GetResponse_Body_Init),
			},
		},
	}

	var respV2 apiobject.GetResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}

	return stream.SendMsg(respV2.ToGRPCMessage().(*protoobject.GetResponse))
}

type testGetObjectPayloadRangeServer struct {
	protoobject.UnimplementedObjectServiceServer
}

func (x *testGetObjectPayloadRangeServer) GetRange(req *protoobject.GetRangeRequest, stream protoobject.ObjectService_GetRangeServer) error {
	ln := req.GetBody().GetRange().GetLength()
	if ln == 0 {
		return nil
	}

	resp := protoobject.GetRangeResponse{
		Body: &protoobject.GetRangeResponse_Body{
			RangePart: &protoobject.GetRangeResponse_Body_Chunk{
				Chunk: make([]byte, ln),
			},
		},
	}

	var respV2 apiobject.GetRangeResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}

	return stream.SendMsg(respV2.ToGRPCMessage().(*protoobject.GetRangeResponse))
}

type testHeadObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
}

func (x *testHeadObjectServer) Head(context.Context, *protoobject.HeadRequest) (*protoobject.HeadResponse, error) {
	resp := protoobject.HeadResponse{
		Body: &protoobject.HeadResponse_Body{
			Head: &protoobject.HeadResponse_Body_Header{
				Header: new(protoobject.HeaderWithSignature),
			},
		},
	}

	var respV2 apiobject.HeadResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoobject.HeadResponse), nil
}

func TestClient_Get(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t)
		ctx := context.Background()

		var nonilAddr v2refs.Address
		nonilAddr.SetObjectID(new(v2refs.ObjectID))
		nonilAddr.SetContainerID(new(v2refs.ContainerID))

		tt := []struct {
			name       string
			methodCall func() error
		}{
			{
				"get",
				func() error {
					_, _, err := c.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, nil, PrmObjectGet{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
			{
				"get_range",
				func() error {
					_, err := c.ObjectRangeInit(ctx, cid.ID{}, oid.ID{}, 0, 1, nil, PrmObjectRange{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
			{
				"get_head",
				func() error {
					_, err := c.ObjectHead(ctx, cid.ID{}, oid.ID{}, nil, PrmObjectHead{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
		}

		for _, test := range tt {
			t.Run(test.name, func(t *testing.T) {
				require.ErrorIs(t, test.methodCall(), ErrMissingSigner)
			})
		}
	})
}
