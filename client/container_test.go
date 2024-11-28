package client

import (
	"context"
	"fmt"
	"testing"

	protoacl "github.com/nspcc-dev/neofs-api-go/v2/acl/grpc"
	apicontainer "github.com/nspcc-dev/neofs-api-go/v2/container"
	protocontainer "github.com/nspcc-dev/neofs-api-go/v2/container/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

// returns Client of Container service provided by given server.
func newTestContainerClient(t testing.TB, srv protocontainer.ContainerServiceServer) *Client {
	return newClient(t, testService{desc: &protocontainer.ContainerService_ServiceDesc, impl: srv})
}

type testPutContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testPutContainerServer) Put(context.Context, *protocontainer.PutRequest) (*protocontainer.PutResponse, error) {
	id := cidtest.ID()
	resp := protocontainer.PutResponse{
		Body: &protocontainer.PutResponse_Body{
			ContainerId: &protorefs.ContainerID{Value: id[:]},
		},
	}

	var respV2 apicontainer.PutResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.PutResponse), nil
}

type testGetContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testGetContainerServer) Get(context.Context, *protocontainer.GetRequest) (*protocontainer.GetResponse, error) {
	cnr := containertest.Container()
	var cnrV2 apicontainer.Container
	cnr.WriteToV2(&cnrV2)
	resp := protocontainer.GetResponse{
		Body: &protocontainer.GetResponse_Body{
			Container: cnrV2.ToGRPCMessage().(*protocontainer.Container),
		},
	}

	var respV2 apicontainer.GetResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.GetResponse), nil
}

type testListContainersServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testListContainersServer) List(context.Context, *protocontainer.ListRequest) (*protocontainer.ListResponse, error) {
	var resp protocontainer.ListResponse

	var respV2 apicontainer.ListResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.ListResponse), nil
}

type testDeleteContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testDeleteContainerServer) Delete(context.Context, *protocontainer.DeleteRequest) (*protocontainer.DeleteResponse, error) {
	var resp protocontainer.DeleteResponse

	var respV2 apicontainer.DeleteResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.DeleteResponse), nil
}

type testGetEACLServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testGetEACLServer) GetExtendedACL(context.Context, *protocontainer.GetExtendedACLRequest) (*protocontainer.GetExtendedACLResponse, error) {
	resp := protocontainer.GetExtendedACLResponse{
		Body: &protocontainer.GetExtendedACLResponse_Body{
			Eacl: new(protoacl.EACLTable),
		},
	}

	var respV2 apicontainer.GetExtendedACLResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.GetExtendedACLResponse), nil
}

type testSetEACLServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testSetEACLServer) SetExtendedACL(context.Context, *protocontainer.SetExtendedACLRequest) (*protocontainer.SetExtendedACLResponse, error) {
	var resp protocontainer.SetExtendedACLResponse

	var respV2 apicontainer.SetExtendedACLResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.SetExtendedACLResponse), nil
}

type testAnnounceContainerSpaceServer struct {
	protocontainer.UnimplementedContainerServiceServer
}

func (x *testAnnounceContainerSpaceServer) AnnounceUsedSpace(context.Context, *protocontainer.AnnounceUsedSpaceRequest) (*protocontainer.AnnounceUsedSpaceResponse, error) {
	var resp protocontainer.AnnounceUsedSpaceResponse

	var respV2 apicontainer.AnnounceUsedSpaceResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protocontainer.AnnounceUsedSpaceResponse), nil
}

func TestClient_Container(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	t.Run("missing signer", func(t *testing.T) {
		tt := []struct {
			name       string
			methodCall func() error
		}{
			{
				"put",
				func() error {
					_, err := c.ContainerPut(ctx, container.Container{}, nil, PrmContainerPut{})
					return err
				},
			},
			{
				"delete",
				func() error {
					return c.ContainerDelete(ctx, cid.ID{}, nil, PrmContainerDelete{})
				},
			},
			{
				"set_eacl",
				func() error {
					return c.ContainerSetEACL(ctx, eacl.Table{}, nil, PrmContainerSetEACL{})
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
