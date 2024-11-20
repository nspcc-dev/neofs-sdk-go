package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestObjectSearch(t *testing.T) {
	ids := oidtest.IDs(20)

	buf := make([]oid.ID, 2)
	checkRead := func(t *testing.T, r *ObjectListReader, expected []oid.ID, expectedErr error) {
		n, err := r.Read(buf)
		if expectedErr == nil {
			require.NoError(t, err)
			require.True(t, len(expected) == len(buf), "expected the same length")
		} else {
			require.Error(t, err)
			require.True(t, len(expected) != len(buf), "expected different length")
		}

		require.Equal(t, len(expected), n, "expected %d items to be read", len(expected))
		require.Equal(t, expected, buf[:len(expected)])
	}

	// no data
	stream := newTestSearchObjectsStream(t, []oid.ID{})
	checkRead(t, stream, []oid.ID{}, io.EOF)

	stream = newTestSearchObjectsStream(t, ids[:3], ids[3:6], ids[6:7], nil, ids[7:8])

	// both ID fetched
	checkRead(t, stream, ids[:2], nil)

	// one ID cached, second fetched
	checkRead(t, stream, ids[2:4], nil)

	// both ID cached
	streamCp := stream.stream
	stream.stream = nil // shouldn't be called, panic if so
	checkRead(t, stream, ids[4:6], nil)
	stream.stream = streamCp

	// both ID fetched in 2 requests, with empty one in the middle
	checkRead(t, stream, ids[6:8], nil)

	// read from tail multiple times
	stream = newTestSearchObjectsStream(t, ids[8:11])
	buf = buf[:1]
	checkRead(t, stream, ids[8:9], nil)
	checkRead(t, stream, ids[9:10], nil)
	checkRead(t, stream, ids[10:11], nil)

	// handle EOF
	buf = buf[:2]
	stream = newTestSearchObjectsStream(t, ids[11:12])
	checkRead(t, stream, ids[11:12], io.EOF)
}

func TestObjectIterate(t *testing.T) {
	ids := oidtest.IDs(3)

	t.Run("no objects", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t)

		var actual []oid.ID
		require.NoError(t, stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Len(t, actual, 0)
	})
	t.Run("iterate all sequence", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t, ids[0:2], nil, ids[2:3])

		var actual []oid.ID
		require.NoError(t, stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Equal(t, ids[:3], actual)
	})
	t.Run("stop by return value", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t, ids)
		var actual []oid.ID
		require.NoError(t, stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return len(actual) == 2
		}))
		require.Equal(t, ids[:2], actual)
	})
	t.Run("stop after error", func(t *testing.T) {
		expectedErr := errors.New("test error")

		stream := newTestSearchObjectsStreamWithEndErr(t, expectedErr, ids[:2])

		var actual []oid.ID
		err := stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		})
		require.ErrorIs(t, err, apistatus.ErrServerInternal)
		require.Equal(t, ids[:2], actual)
	})
}

func TestClient_ObjectSearch(t *testing.T) {
	c := newClient(t)

	t.Run("missing signer", func(t *testing.T) {
		_, err := c.ObjectSearchInit(context.Background(), cid.ID{}, nil, PrmObjectSearch{})
		require.ErrorIs(t, err, ErrMissingSigner)
	})
}

func newTestSearchObjectsStreamWithEndErr(t *testing.T, endError error, idList ...[]oid.ID) *ObjectListReader {
	usr := usertest.User()
	srv := testSearchObjectsServer{
		signer:    usr,
		endStatus: apistatus.ErrorToV2(endError).ToGRPCMessage().(*protostatus.Status),
		idList:    idList,
	}
	stream, err := newTestObjectClient(t, &srv).ObjectSearchInit(context.Background(), cidtest.ID(), usr, PrmObjectSearch{})
	require.NoError(t, err)
	return stream
}

func newTestSearchObjectsStream(t *testing.T, idList ...[]oid.ID) *ObjectListReader {
	return newTestSearchObjectsStreamWithEndErr(t, nil, idList...)
}

type testSearchObjectsServer struct {
	protoobject.UnimplementedObjectServiceServer

	signer    neofscrypto.Signer
	endStatus *protostatus.Status
	idList    [][]oid.ID
}

func (x *testSearchObjectsServer) Search(_ *protoobject.SearchRequest, stream protoobject.ObjectService_SearchServer) error {
	signer := x.signer
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	for i := range x.idList {
		resp := protoobject.SearchResponse{
			Body: &protoobject.SearchResponse_Body{
				IdList: make([]*protorefs.ObjectID, len(x.idList[i])),
			},
		}
		for j := range x.idList[i] {
			resp.Body.IdList[j] = &protorefs.ObjectID{Value: x.idList[i][j][:]}
		}

		var respV2 v2object.SearchResponse
		if err := respV2.FromGRPCMessage(&resp); err != nil {
			panic(err)
		}
		if err := signServiceMessage(signer, &respV2, nil); err != nil {
			return fmt.Errorf("sign response message: %w", err)
		}
		if err := stream.Send(respV2.ToGRPCMessage().(*protoobject.SearchResponse)); err != nil {
			return err
		}
	}

	if x.endStatus == nil {
		return nil
	}

	resp := protoobject.SearchResponse{
		MetaHeader: &protosession.ResponseMetaHeader{
			Status: x.endStatus,
		},
	}

	var respV2 v2object.SearchResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(signer, &respV2, nil); err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}
	if err := stream.Send(respV2.ToGRPCMessage().(*protoobject.SearchResponse)); err != nil {
		return err
	}

	return stream.Send(respV2.ToGRPCMessage().(*protoobject.SearchResponse))
}
