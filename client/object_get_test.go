package client

import (
	"context"
	"fmt"
	"io"
	"testing"

	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

type testGetObjectResponseStream struct {
	sent bool
}

func (x *testGetObjectResponseStream) Read(resp *apiobject.GetResponse) error {
	if x.sent {
		return io.EOF
	}

	var body apiobject.GetResponseBody
	body.SetObjectPart(new(apiobject.GetObjectPartInit))
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), resp, nil); err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}

	x.sent = true
	return nil
}

type testGetObjectServer struct {
	unimplementedNeoFSAPIServer

	stream testGetObjectResponseStream
}

func (x *testGetObjectServer) getObject(context.Context, apiobject.GetRequest) (getObjectResponseStream, error) {
	x.stream.sent = false
	return &x.stream, nil
}

type testGetObjectPayloadRangeServer struct {
	unimplementedNeoFSAPIServer

	stream testGetObjectPayloadRangeResponseStream
}

func (x *testGetObjectPayloadRangeServer) getObjectPayloadRange(_ context.Context, req apiobject.GetRangeRequest) (getObjectPayloadRangeResponseStream, error) {
	x.stream.sent = false
	x.stream.ln = req.GetBody().GetRange().GetLength()
	return &x.stream, nil
}

type testGetObjectPayloadRangeResponseStream struct {
	ln   uint64
	sent bool
}

func (x *testGetObjectPayloadRangeResponseStream) Read(resp *apiobject.GetRangeResponse) error {
	if x.sent {
		return io.EOF
	}

	var rngPart apiobject.GetRangePartChunk
	rngPart.SetChunk(make([]byte, x.ln))
	var body apiobject.GetRangeResponseBody
	body.SetRangePart(&rngPart)
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), resp, nil); err != nil {
		return fmt.Errorf("sign response message: %w", err)
	}

	x.sent = true
	return nil
}

type testHeadObjectServer struct {
	unimplementedNeoFSAPIServer
}

func (x *testHeadObjectServer) headObject(context.Context, apiobject.HeadRequest) (*apiobject.HeadResponse, error) {
	var body apiobject.HeadResponseBody
	body.SetHeaderPart(new(apiobject.HeaderWithSignature))
	var resp apiobject.HeadResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

func TestClient_Get(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t, nil)
		ctx := context.Background()

		var nonilAddr v2refs.Address
		nonilAddr.SetObjectID(new(v2refs.ObjectID))
		nonilAddr.SetContainerID(new(v2refs.ContainerID))

		tt := []struct {
			name       string
			methodCall func() error
		}{
			{
				"get",
				func() error {
					_, _, err := c.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, nil, PrmObjectGet{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
			{
				"get_range",
				func() error {
					_, err := c.ObjectRangeInit(ctx, cid.ID{}, oid.ID{}, 0, 1, nil, PrmObjectRange{prmObjectRead: prmObjectRead{}})
					return err
				},
			},
			{
				"get_head",
				func() error {
					_, err := c.ObjectHead(ctx, cid.ID{}, oid.ID{}, nil, PrmObjectHead{prmObjectRead: prmObjectRead{}})
					return err
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
