package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

type testPutObjectStream struct {
	denyAccess bool
}

func (t *testPutObjectStream) Write(req *v2object.PutRequest) error {
	switch req.GetBody().GetObjectPart().(type) {
	case *v2object.PutObjectPartInit:
		return nil
	case *v2object.PutObjectPartChunk:
		return io.EOF
	default:
		return errors.New("excuse me?")
	}
}

func (t *testPutObjectStream) Close() (*v2object.PutResponse, error) {
	m := new(v2session.ResponseMetaHeader)

	var v refs.Version
	version.Current().WriteToV2(&v)

	m.SetVersion(&v)
	if t.denyAccess {
		m.SetStatus(apistatus.ErrObjectAccessDenied.ErrorToV2())
	}

	var resp v2object.PutResponse
	resp.SetMetaHeader(m)
	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testPutObjectServer struct {
	unimplementedNeoFSAPIServer

	stream testPutObjectStream
}

func (x *testPutObjectServer) putObject(context.Context) (objectWriter, error) {
	return &x.stream, nil
}

func TestClient_ObjectPutInit(t *testing.T) {
	t.Run("EOF-on-status-return", func(t *testing.T) {
		srv := testPutObjectServer{
			stream: testPutObjectStream{
				denyAccess: true,
			},
		}
		c := newClient(t, &srv)
		usr := usertest.User()

		w, err := c.ObjectPutInit(context.Background(), object.Object{}, usr, PrmObjectPutInit{})
		require.NoError(t, err)

		n, err := w.Write([]byte{1})
		require.Zero(t, n)
		require.ErrorIs(t, err, new(apistatus.ObjectAccessDenied))

		err = w.Close()
		require.NoError(t, err)
	})
}
