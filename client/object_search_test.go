package client

import (
	"context"
	"errors"
	"io"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestObjectSearch(t *testing.T) {
	ids := make([]oid.ID, 20)
	for i := range ids {
		ids[i] = oidtest.ID()
	}

	p, resp := testListReaderResponse(t)

	buf := make([]oid.ID, 2)
	checkRead := func(t *testing.T, expected []oid.ID, expectedErr error) {
		n, err := resp.Read(buf)
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

	// nil panic
	require.Panics(t, func() {
		_, _ = resp.Read(nil)
	})

	// no data
	resp.stream = newSearchStream(p, io.EOF, []oid.ID{})
	checkRead(t, []oid.ID{}, io.EOF)

	// both ID fetched
	resp.stream = newSearchStream(p, nil, ids[:3])
	checkRead(t, ids[:2], nil)

	// one ID cached, second fetched
	resp.stream = newSearchStream(p, nil, ids[3:6])
	checkRead(t, ids[2:4], nil)

	// both ID cached
	resp.stream = nil // shouldn't be called, panic if so
	checkRead(t, ids[4:6], nil)

	// both ID fetched in 2 requests, with empty one in the middle
	resp.stream = newSearchStream(p, nil, ids[6:7], nil, ids[7:8])
	checkRead(t, ids[6:8], nil)

	// read from tail multiple times
	resp.stream = newSearchStream(p, nil, ids[8:11])
	buf = buf[:1]
	checkRead(t, ids[8:9], nil)
	checkRead(t, ids[9:10], nil)
	checkRead(t, ids[10:11], nil)

	// handle EOF
	buf = buf[:2]
	resp.stream = newSearchStream(p, io.EOF, ids[11:12])
	checkRead(t, ids[11:12], io.EOF)
}

func TestObjectIterate(t *testing.T) {
	ids := make([]oid.ID, 3)
	for i := range ids {
		ids[i] = oidtest.ID()
	}

	t.Run("no objects", func(t *testing.T) {
		p, resp := testListReaderResponse(t)

		resp.stream = newSearchStream(p, io.EOF, []oid.ID{})

		var actual []oid.ID
		require.NoError(t, resp.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Len(t, actual, 0)
	})
	t.Run("iterate all sequence", func(t *testing.T) {
		p, resp := testListReaderResponse(t)

		resp.stream = newSearchStream(p, io.EOF, ids[0:2], nil, ids[2:3])

		var actual []oid.ID
		require.NoError(t, resp.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Equal(t, ids[:3], actual)
	})
	t.Run("stop by return value", func(t *testing.T) {
		p, resp := testListReaderResponse(t)

		var actual []oid.ID
		resp.stream = &singleStreamResponder{signer: p, idList: [][]oid.ID{ids}}
		require.NoError(t, resp.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return len(actual) == 2
		}))
		require.Equal(t, ids[:2], actual)
	})
	t.Run("stop after error", func(t *testing.T) {
		p, resp := testListReaderResponse(t)
		expectedErr := errors.New("test error")

		resp.stream = newSearchStream(p, expectedErr, ids[:2])

		var actual []oid.ID
		err := resp.Iterate(func(id oid.ID) bool {
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

func testListReaderResponse(t *testing.T) (neofscrypto.Signer, *ObjectListReader) {
	return test.RandomSigner(t), &ObjectListReader{
		cancelCtxStream: func() {},
		client:          &Client{},
		tail:            nil,
	}
}

func newSearchStream(signer neofscrypto.Signer, endError error, idList ...[]oid.ID) *singleStreamResponder {
	return &singleStreamResponder{
		signer:   signer,
		endError: endError,
		idList:   idList,
	}
}

type singleStreamResponder struct {
	signer   neofscrypto.Signer
	n        int
	endError error
	idList   [][]oid.ID
}

func (s *singleStreamResponder) Read(resp *v2object.SearchResponse) error {
	if s.n >= len(s.idList) {
		if s.endError != nil {
			return s.endError
		}
		return ErrUnexpectedReadCall
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

	err := signServiceMessage(s.signer, resp, nil)
	if err != nil {
		return err
	}

	s.n++
	return nil
}
