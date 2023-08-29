package client

import (
	"context"
	"errors"
	"io"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

type testPutStreamAccessDenied struct {
	resp   *v2object.PutResponse
	signer user.Signer
	t      *testing.T
}

func (t *testPutStreamAccessDenied) Write(req *v2object.PutRequest) error {
	switch req.GetBody().GetObjectPart().(type) {
	case *v2object.PutObjectPartInit:
		return nil
	case *v2object.PutObjectPartChunk:
		return io.EOF
	default:
		return errors.New("excuse me?")
	}
}

func (t *testPutStreamAccessDenied) Close() error {
	m := new(v2session.ResponseMetaHeader)

	var v refs.Version
	version.Current().WriteToV2(&v)

	m.SetVersion(&v)
	m.SetStatus(apistatus.ErrObjectAccessDenied.ErrorToV2())

	t.resp.SetMetaHeader(m)
	require.NoError(t.t, signServiceMessage(t.signer, t.resp))

	return nil
}

func TestClient_ObjectPutInit(t *testing.T) {
	t.Run("EOF-on-status-return", func(t *testing.T) {
		c := newClient(t, nil)
		signer := test.RandomSignerRFC6979(t)

		rpcAPIPutObject = func(cli *client.Client, r *v2object.PutResponse, o ...client.CallOption) (objectWriter, error) {
			return &testPutStreamAccessDenied{resp: r, signer: signer, t: t}, nil
		}

		w, err := c.ObjectPutInit(context.Background(), object.Object{}, signer, PrmObjectPutInit{})
		require.NoError(t, err)

		n, err := w.Write([]byte{1})
		require.Zero(t, n)
		require.ErrorIs(t, err, new(apistatus.ObjectAccessDenied))

		err = w.Close()
		require.NoError(t, err)
	})
}
