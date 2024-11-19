package client

import (
	"context"
	"fmt"
	"testing"

	apiacl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	apicontainer "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

type testPutContainerServer struct {
	unimplementedNeoFSAPIServer
}

func (x testPutContainerServer) putContainer(context.Context, apicontainer.PutRequest) (*apicontainer.PutResponse, error) {
	id := cidtest.ID()
	var idV2 refs.ContainerID
	id.WriteToV2(&idV2)
	var body apicontainer.PutResponseBody
	body.SetContainerID(&idV2)
	var resp apicontainer.PutResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testGetContainerServer struct {
	unimplementedNeoFSAPIServer
}

func (x testGetContainerServer) getContainer(context.Context, apicontainer.GetRequest) (*apicontainer.GetResponse, error) {
	cnr := containertest.Container()
	var cnrV2 apicontainer.Container
	cnr.WriteToV2(&cnrV2)
	var body apicontainer.GetResponseBody
	body.SetContainer(&cnrV2)
	var resp apicontainer.GetResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testListContainersServer struct {
	unimplementedNeoFSAPIServer
}

func (x testListContainersServer) listContainers(context.Context, apicontainer.ListRequest) (*apicontainer.ListResponse, error) {
	var resp apicontainer.ListResponse

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testDeleteContainerServer struct {
	unimplementedNeoFSAPIServer
}

func (x testDeleteContainerServer) deleteContainer(context.Context, apicontainer.DeleteRequest) (*apicontainer.DeleteResponse, error) {
	var resp apicontainer.DeleteResponse

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testGetEACLServer struct {
	unimplementedNeoFSAPIServer
}

func (x testGetEACLServer) getEACL(context.Context, apicontainer.GetExtendedACLRequest) (*apicontainer.GetExtendedACLResponse, error) {
	var body apicontainer.GetExtendedACLResponseBody
	body.SetEACL(new(apiacl.Table))
	var resp apicontainer.GetExtendedACLResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testSetEACLServer struct {
	unimplementedNeoFSAPIServer
}

func (x testSetEACLServer) setEACL(context.Context, apicontainer.SetExtendedACLRequest) (*apicontainer.SetExtendedACLResponse, error) {
	var resp apicontainer.SetExtendedACLResponse

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testAnnounceContainerSpaceServer struct {
	unimplementedNeoFSAPIServer
}

func (x testAnnounceContainerSpaceServer) announceContainerSpace(context.Context, apicontainer.AnnounceUsedSpaceRequest) (*apicontainer.AnnounceUsedSpaceResponse, error) {
	var resp apicontainer.AnnounceUsedSpaceResponse

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

func TestClient_Container(t *testing.T) {
	c := newClient(t, nil)
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
