package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"

	objectgrpc "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	status "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func BenchmarkPrepareReplicationMessage(b *testing.B) {
	bObj := make([]byte, 1<<10)
	_, err := rand.Read(bObj) // structure does not matter for
	require.NoError(b, err)
	id := oidtest.ID()

	var signer nopSigner

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err = prepareReplicateMessage(id, bytes.NewReader(bObj), signer, true)
		require.NoError(b, err)
	}
}

var testDataToSign = []byte("requested data to sign")

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
	if sigMsg == nil {
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

	bObjSent := x.clientObj.Marshal()
	bObjRecv := obj.Marshal()

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

	id := obj.GetID()
	if id.IsZero() {
		st.Code = 1024 // internal error
		st.Message = "missing object ID"
		resp.Status = &st
		return &resp, nil
	}

	if !x.clientSigner.Public().Verify(id[:], sigMsg.GetSign()) {
		st.Code = 1024 // internal error
		st.Message = "signature verification failed"
		resp.Status = &st
		return &resp, nil
	}

	if req.GetSignObject() {
		var sig neofscrypto.Signature
		err = sig.Calculate(x.clientSigner, testDataToSign)
		if err != nil {
			st.Code = 1024
			st.Message = fmt.Sprintf("signing object information: %s", err)
			resp.Status = &st

			return &resp, nil
		}

		var sigV2 refs.Signature
		sig.WriteToV2(&sigV2)

		resp.ObjectSignature = sigV2.StableMarshal(nil)
	}

	resp.Status = &status.Status{Code: x.respStatusCode}
	return &resp, nil
}

func TestClient_ReplicateObject(t *testing.T) {
	ctx := context.Background()
	signer := neofscryptotest.Signer()
	obj := objecttest.Object()
	id := oidtest.ID()
	obj.SetID(id)
	bObj := obj.Marshal()

	t.Run("OK", func(t *testing.T) {
		srv := testReplicationServer{
			clientSigner: signer,
			clientObj:    obj,
		}
		cli := newTestObjectClient(t, &srv)

		_, err := cli.ReplicateObject(ctx, id, bytes.NewReader(bObj), signer, false)
		require.NoError(t, err)
	})

	t.Run("invalid binary object", func(t *testing.T) {
		bObj := []byte("Hello, world!") // definitely incorrect binary object
		cli := newClient(t)

		_, err := cli.ReplicateObject(ctx, id, bytes.NewReader(bObj), signer, false)
		require.Error(t, err)
	})

	t.Run("statuses", func(t *testing.T) {
		srv := testReplicationServer{
			clientSigner: signer,
			clientObj:    obj,
		}
		cli := newTestObjectClient(t, &srv)
		for _, tc := range []struct {
			code   uint32
			expErr error
			desc   string
		}{
			{code: 1024, expErr: apistatus.ErrServerInternal, desc: "internal server error"},
			{code: 2048, expErr: apistatus.ErrObjectAccessDenied, desc: "forbidden"},
			{code: 3072, expErr: apistatus.ErrContainerNotFound, desc: "container not found"},
		} {
			srv.respStatusCode = tc.code

			_, err := cli.ReplicateObject(ctx, id, bytes.NewReader(bObj), signer, false)
			require.ErrorIs(t, err, tc.expErr, tc.desc)
		}
	})

	t.Run("sign object data", func(t *testing.T) {
		srv := testReplicationServer{
			clientSigner: signer,
			clientObj:    obj,
		}
		cli := newTestObjectClient(t, &srv)

		sig, err := cli.ReplicateObject(ctx, id, bytes.NewReader(bObj), signer, true)
		require.NoError(t, err)
		require.True(t, sig.Verify(testDataToSign))
	})

	t.Run("demux", func(t *testing.T) {
		demuxObj := DemuxReplicatedObject(bytes.NewReader(bObj))
		srv := testReplicationServer{
			clientSigner: signer,
			clientObj:    obj,
		}
		cli := newTestObjectClient(t, &srv)

		_, err := cli.ReplicateObject(ctx, id, demuxObj, signer, false)
		require.NoError(t, err)

		msgCp := bytes.Clone(demuxObj.(*demuxReplicationMessage).msg)
		initBufPtr := &demuxObj.(*demuxReplicationMessage).msg[0]

		var wg sync.WaitGroup
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := cli.ReplicateObject(ctx, id, demuxObj, signer, false)
				require.NoError(t, err)
			}()
		}

		wg.Wait()

		require.Equal(t, msgCp, demuxObj.(*demuxReplicationMessage).msg)
		require.Equal(t, initBufPtr, &demuxObj.(*demuxReplicationMessage).msg[0])
	})
}
