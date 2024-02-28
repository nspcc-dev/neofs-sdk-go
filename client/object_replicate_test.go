package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"testing"

	objectgrpc "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	status "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

func BenchmarkPrepareReplicationMessage(b *testing.B) {
	bObj := make([]byte, 1<<10)
	_, err := rand.Read(bObj) // structure does not matter for
	require.NoError(b, err)

	var signer nopSigner

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = prepareReplicateMessage(bytes.NewReader(bObj), signer)
		require.NoError(b, err)
	}
}

type testReplicationServer struct {
	objectgrpc.UnimplementedObjectServiceServer

	clientSigner neofscrypto.Signer
	clientObj    object.Object

	respStatusCode uint32
}

func (x *testReplicationServer) Replicate(_ context.Context, req *objectgrpc.ReplicateRequest) (*objectgrpc.ReplicateResponse, error) {
	var resp objectgrpc.ReplicateResponse
	var st status.Status

	objMsg := req.GetObject()
	if objMsg == nil {
		st.Code = 1024 // internal error
		st.Message = "missing object field"
		resp.Status = &st
		return &resp, nil
	}

	sigMsg := req.GetSignature()
	if objMsg == nil {
		st.Code = 1024 // internal error
		st.Message = "missing signature field"
		resp.Status = &st
		return &resp, nil
	}

	var obj object.Object

	bObj, _ := proto.Marshal(objMsg)
	err := obj.Unmarshal(bObj)
	if err != nil {
		st.Code = 1024 // internal error
		st.Message = fmt.Sprintf("decode binary object: %v", err)
		resp.Status = &st
		return &resp, nil
	}

	bObjSent, _ := x.clientObj.Marshal()
	bObjRecv, _ := obj.Marshal()

	if !bytes.Equal(bObjSent, bObjRecv) {
		st.Code = 1024 // internal error
		st.Message = "received object differs from the sent one"
		resp.Status = &st
		return &resp, nil
	}

	if !bytes.Equal(sigMsg.GetKey(), neofscrypto.PublicKeyBytes(x.clientSigner.Public())) {
		st.Code = 1024 // internal error
		st.Message = "public key in the received signature differs with the client's one"
		resp.Status = &st
		return &resp, nil
	}

	if int32(sigMsg.GetScheme()) != int32(x.clientSigner.Scheme()) {
		st.Code = 1024 // internal error
		st.Message = "signature scheme in the received signature differs with the client's one"
		resp.Status = &st
		return &resp, nil
	}

	if !x.clientSigner.Public().Verify(bObjRecv, sigMsg.GetSign()) {
		st.Code = 1024 // internal error
		st.Message = "signature verification failed"
		resp.Status = &st
		return &resp, nil
	}

	resp.Status = &status.Status{Code: x.respStatusCode}
	return &resp, nil
}

func serveObjectReplication(tb testing.TB, clientSigner neofscrypto.Signer, clientObj object.Object) (*testReplicationServer, *Client) {
	lis := bufconn.Listen(1 << 10)

	var replicationSrv testReplicationServer

	gSrv := grpc.NewServer()
	objectgrpc.RegisterObjectServiceServer(gSrv, &replicationSrv)

	gConn, err := grpc.Dial("", grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(tb, err)

	tb.Cleanup(gSrv.Stop)

	go func() { _ = gSrv.Serve(lis) }()

	replicationSrv.clientObj = clientObj
	replicationSrv.clientSigner = clientSigner

	return &replicationSrv, &Client{
		c: *client.New(client.WithGRPCConn(gConn)),
	}
}

func TestClient_ReplicateObject(t *testing.T) {
	ctx := context.Background()
	signer := test.RandomSigner(t)
	obj := objecttest.Object(t)
	bObj, _ := obj.Marshal()

	t.Run("OK", func(t *testing.T) {
		srv, cli := serveObjectReplication(t, signer, obj)
		srv.respStatusCode = 0

		err := cli.ReplicateObject(ctx, bytes.NewReader(bObj), signer)
		require.NoError(t, err)
	})

	t.Run("invalid binary object", func(t *testing.T) {
		bObj := []byte("Hello, world!") // definitely incorrect binary object
		_, cli := serveObjectReplication(t, signer, obj)

		err := cli.ReplicateObject(ctx, bytes.NewReader(bObj), signer)
		require.Error(t, err)
	})

	t.Run("statuses", func(t *testing.T) {
		for _, tc := range []struct {
			code   uint32
			expErr error
			desc   string
		}{
			{code: 1024, expErr: apistatus.ErrServerInternal, desc: "internal server error"},
			{code: 2048, expErr: apistatus.ErrObjectAccessDenied, desc: "forbidden"},
			{code: 3072, expErr: apistatus.ErrContainerNotFound, desc: "container not found"},
		} {
			srv, cli := serveObjectReplication(t, signer, obj)
			srv.respStatusCode = tc.code

			err := cli.ReplicateObject(ctx, bytes.NewReader(bObj), signer)
			require.ErrorIs(t, err, tc.expErr, tc.desc)
		}
	})

	t.Run("demux", func(t *testing.T) {
		demuxObj := DemuxReplicatedObject(bytes.NewReader(bObj))
		_, cli := serveObjectReplication(t, signer, obj)

		err := cli.ReplicateObject(ctx, demuxObj, signer)
		require.NoError(t, err)

		msgCp := bytes.Clone(demuxObj.(*demuxReplicationMessage).msg)
		initBufPtr := &demuxObj.(*demuxReplicationMessage).msg[0]

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				err := cli.ReplicateObject(ctx, demuxObj, signer)
				fmt.Println(err)
				require.NoError(t, err)
			}()
		}

		wg.Wait()

		require.Equal(t, msgCp, demuxObj.(*demuxReplicationMessage).msg)
		require.Equal(t, initBufPtr, &demuxObj.(*demuxReplicationMessage).msg[0])
	})
}
