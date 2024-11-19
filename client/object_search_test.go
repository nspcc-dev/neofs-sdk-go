package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
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
		require.True(t, errors.Is(err, expectedErr), "got: %v", err)
		require.Equal(t, ids[:2], actual)
	})
}

func TestClient_ObjectSearch(t *testing.T) {
	c := newClient(t, nil)

	t.Run("missing signer", func(t *testing.T) {
		_, err := c.ObjectSearchInit(context.Background(), cid.ID{}, nil, PrmObjectSearch{})
		require.ErrorIs(t, err, ErrMissingSigner)
	})
}

func newTestSearchObjectsStreamWithEndErr(t *testing.T, endError error, idList ...[]oid.ID) *ObjectListReader {
	usr := usertest.User()
	srv := testSearchObjectsServer{
		stream: testSearchObjectsResponseStream{
			signer:   usr,
			endError: endError,
			idList:   idList,
		},
	}
	stream, err := newClient(t, &srv).ObjectSearchInit(context.Background(), cidtest.ID(), usr, PrmObjectSearch{})
	require.NoError(t, err)
	return stream
}

func newTestSearchObjectsStream(t *testing.T, idList ...[]oid.ID) *ObjectListReader {
	return newTestSearchObjectsStreamWithEndErr(t, nil, idList...)
}

type testSearchObjectsResponseStream struct {
	signer   neofscrypto.Signer
	n        int
	endError error
	idList   [][]oid.ID
}

func (s *testSearchObjectsResponseStream) Read(resp *v2object.SearchResponse) error {
	if len(s.idList) == 0 || s.n == len(s.idList) {
		if s.endError != nil {
			return s.endError
		}
		return io.EOF
	}

	var body v2object.SearchResponseBody

	if s.idList[s.n] != nil {
		ids := make([]refs.ObjectID, len(s.idList[s.n]))
		for i := range s.idList[s.n] {
			s.idList[s.n][i].WriteToV2(&ids[i])
		}
		body.SetIDList(ids)
	}
	resp.SetBody(&body)

	signer := s.signer
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	err := signServiceMessage(s.signer, resp, nil)
	if err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}

	s.n++
	return nil
}

type testSearchObjectsServer struct {
	unimplementedNeoFSAPIServer

	stream testSearchObjectsResponseStream
}

func (x *testSearchObjectsServer) searchObjects(context.Context, apiobject.SearchRequest) (searchObjectsResponseStream, error) {
	x.stream.n = 0
	return &x.stream, nil
}
