package client

import (
	"errors"
	"io"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	signatureV2 "github.com/nspcc-dev/neofs-api-go/v2/signature"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestObjectSearch(t *testing.T) {
	ids := make([]oid.ID, 20)
	for i := range ids {
		ids[i] = *oidtest.ID()
	}

	resp, setID := testListReaderResponse(t)

	buf := make([]oid.ID, 2)
	checkRead := func(t *testing.T, expected []oid.ID) {
		n, ok := resp.Read(buf)
		require.True(t, ok == (len(expected) == len(buf)), "expected no error")
		require.Equal(t, len(expected), n, "expected %d items to be read", len(expected))
		require.Equal(t, expected, buf[:len(expected)])
	}

	// nil panic
	require.Panics(t, func() { resp.Read(nil) })

	// both ID fetched
	setID(ids[:3])
	checkRead(t, ids[:2])

	// one ID cached, second fetched
	setID(ids[3:6])
	checkRead(t, ids[2:4])

	// both ID cached
	resp.ctxCall.resp = nil
	checkRead(t, ids[4:6])

	// both ID fetched in 2 requests, with empty one in the middle
	var n int
	resp.ctxCall.rResp = func() error {
		switch n {
		case 0:
			setID(ids[6:7])
		case 1:
			setID(nil)
		case 2:
			setID(ids[7:8])
		default:
			t.FailNow()
		}
		n++
		return nil
	}
	checkRead(t, ids[6:8])

	// read from tail multiple times
	resp.ctxCall.rResp = nil
	setID(ids[8:11])
	buf = buf[:1]
	checkRead(t, ids[8:9])
	checkRead(t, ids[9:10])
	checkRead(t, ids[10:11])

	// handle EOF
	buf = buf[:2]
	n = 0
	resp.ctxCall.rResp = func() error {
		if n > 0 {
			return io.EOF
		}
		n++
		setID(ids[11:12])
		return nil
	}
	checkRead(t, ids[11:12])
}

func TestObjectIterate(t *testing.T) {
	ids := make([]oid.ID, 3)
	for i := range ids {
		ids[i] = *oidtest.ID()
	}

	t.Run("iterate all sequence", func(t *testing.T) {
		resp, setID := testListReaderResponse(t)

		// Iterate over all sequence
		var n int
		resp.ctxCall.rResp = func() error {
			switch n {
			case 0:
				setID(ids[0:2])
			case 1:
				setID(nil)
			case 2:
				setID(ids[2:3])
			default:
				return io.EOF
			}
			n++
			return nil
		}

		var actual []oid.ID
		require.NoError(t, resp.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Equal(t, ids[:3], actual)
	})
	t.Run("stop by return value", func(t *testing.T) {
		resp, setID := testListReaderResponse(t)

		var actual []oid.ID
		setID(ids)
		require.NoError(t, resp.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return len(actual) == 2
		}))
		require.Equal(t, ids[:2], actual)
	})
	t.Run("stop after error", func(t *testing.T) {
		resp, setID := testListReaderResponse(t)
		expectedErr := errors.New("test error")

		var actual []oid.ID
		var n int
		resp.ctxCall.rResp = func() error {
			switch n {
			case 0:
				setID(ids[:2])
			default:
				return expectedErr
			}
			n++
			return nil
		}

		err := resp.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		})
		require.True(t, errors.Is(err, expectedErr), "got: %v", err)
		require.Equal(t, ids[:2], actual)
	})
}

func testListReaderResponse(t *testing.T) (*ObjectListReader, func(id []oid.ID) *object.SearchResponse) {
	p, err := keys.NewPrivateKey()
	require.NoError(t, err)

	obj := &ObjectListReader{
		cancelCtxStream: func() {},
		ctxCall: contextCall{
			closer:    func() error { return nil },
			result:    func(v2 responseV2) {},
			statusRes: new(ResObjectSearch),
		},
		reqWritten: true,
		bodyResp:   object.SearchResponseBody{},
		tail:       nil,
	}

	return obj, func(id []oid.ID) *object.SearchResponse {
		resp := new(object.SearchResponse)
		resp.SetBody(new(object.SearchResponseBody))

		v2id := make([]*refs.ObjectID, len(id))
		for i := range id {
			v2id[i] = id[i].ToV2()
		}
		resp.GetBody().SetIDList(v2id)
		err := signatureV2.SignServiceMessage(&p.PrivateKey, resp)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		obj.ctxCall.resp = resp
		obj.bodyResp = *resp.GetBody()
		return resp
	}
}
