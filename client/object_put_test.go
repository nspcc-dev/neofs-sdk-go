package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

type testPutObjectServer struct {
	protoobject.UnimplementedObjectServiceServer

	denyAccess bool
}

func (x *testPutObjectServer) Put(stream protoobject.ObjectService_PutServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		switch req.GetBody().GetObjectPart().(type) {
		case *protoobject.PutRequest_Body_Init_,
			*protoobject.PutRequest_Body_Chunk:
		default:
			return errors.New("excuse me?")
		}
	}

	var v refs.Version
	version.Current().WriteToV2(&v)
	id := oidtest.ID()
	resp := protoobject.PutResponse{
		Body: &protoobject.PutResponse_Body{
			ObjectId: &protorefs.ObjectID{Value: id[:]},
		},
		MetaHeader: &protosession.ResponseMetaHeader{
			Version: v.ToGRPCMessage().(*protorefs.Version),
		},
	}

	if x.denyAccess {
		resp.MetaHeader.Status = apistatus.ErrObjectAccessDenied.ErrorToV2().ToGRPCMessage().(*protostatus.Status)
	}

	var respV2 v2object.PutResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}

	return stream.SendAndClose(respV2.ToGRPCMessage().(*protoobject.PutResponse))
}

func TestClient_ObjectPutInit(t *testing.T) {
	t.Run("EOF-on-status-return", func(t *testing.T) {
		srv := testPutObjectServer{
			denyAccess: true,
		}
		c := newTestObjectClient(t, &srv)
		usr := usertest.User()

		w, err := c.ObjectPutInit(context.Background(), object.Object{}, usr, PrmObjectPutInit{})
		require.NoError(t, err)

		err = w.Close()
		require.ErrorIs(t, err, apistatus.ErrObjectAccessDenied)
	})
}
