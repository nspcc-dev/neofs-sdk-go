package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	objectgrpc "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-api-go/v2/signature"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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
	require.NoError(t.t, signServiceMessage(t.signer, t.resp, nil))

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

type discardBinaryMessageWriter struct{}

func (discardBinaryMessageWriter) Write([]byte) error {
	return nil
}

type binaryObjectStreamChecker struct {
	bSignerPubKey []byte

	seenInit bool

	restoredObject object.Object
}

func newBinaryObjectStreamChecker(bSignerPubKey []byte) *binaryObjectStreamChecker {
	return &binaryObjectStreamChecker{
		bSignerPubKey: bSignerPubKey,
	}
}

func (x *binaryObjectStreamChecker) Write(msg []byte) error {
	var gRes objectgrpc.PutRequest

	err := proto.Unmarshal(msg, &gRes)
	if err != nil {
		return fmt.Errorf("decode binary object PUT request: %w", err)
	}

	var req v2object.PutRequest

	err = req.FromGRPCMessage(&gRes)
	if err != nil {
		return fmt.Errorf("convert request message: %w", err)
	}

	err = signature.VerifyServiceMessage(&req)
	if err != nil {
		return fmt.Errorf("verify request: %w", err)
	}

	metaHdr := req.GetMetaHeader()
	if metaHdr == nil {
		return errors.New("missing meta header in the request")
	}

	ver := metaHdr.GetVersion()
	curVer := version.Current()
	if ver.GetMajor() != curVer.Major() || ver.GetMinor() != curVer.Minor() {
		return errors.New("wrong protocol version in the meta header")
	}

	ttl := metaHdr.GetTTL()
	if ttl != 1 {
		return fmt.Errorf("unexpected TTL in the meta header: %d instead of 1", ttl)
	}

	switch part := req.GetBody().GetObjectPart().(type) {
	default:
		return fmt.Errorf("unexpected stream message type %T", part)
	case *v2object.PutObjectPartInit:
		if x.seenInit {
			return errors.New("duplicated init message in the stream")
		}

		x.seenInit = true

		copyNum := part.GetCopiesNumber()
		if copyNum != 0 {
			return fmt.Errorf("unexpected copies number field: %d instead of 0", copyNum)
		}

		var objMsg v2object.Object
		objMsg.SetObjectID(part.GetObjectID())
		objMsg.SetSignature(part.GetSignature())
		objMsg.SetHeader(part.GetHeader())

		x.restoredObject = *object.NewFromV2(&objMsg)
	case *v2object.PutObjectPartChunk:
		if !x.seenInit {
			return errors.New("payload chunk message before the initial one")
		}

		chunk := part.GetChunk()
		if len(chunk) == 0 {
			return errors.New("empty payload chunk")
		}

		x.restoredObject.SetPayload(append(x.restoredObject.Payload(), chunk...))
	}

	return nil
}

func TestStreamBinaryObject(t *testing.T) {
	f := func(withPayload, singleBuffer bool) {
		t.Run(fmt.Sprintf("with_payload=%t,single_buffer=%t", withPayload, singleBuffer), func(t *testing.T) {
			signer := test.RandomSigner(t)
			obj := objecttest.Object(t)

			var payload []byte

			if withPayload {
				payload = make([]byte, 3*defaultBufferSize)
				_, err := rand.Read(payload)
				require.NoError(t, err)
			}

			obj.SetPayload(payload)

			b, err := obj.Marshal()
			require.NoError(t, err)

			stream := newBinaryObjectStreamChecker(neofscrypto.PublicKeyBytes(signer.Public()))

			err = streamBinaryObject(stream, bytes.NewReader(b), signer, singleBuffer)
			require.NoError(t, err)
			require.Equal(t, obj, stream.restoredObject)
		})
	}

	f(false, false)
	f(false, true)
	f(true, false)
	f(true, true)
}

func BenchmarkStreamBinaryObject(b *testing.B) {
	for _, payloadSize := range []int{
		0,
		100,
		32 << 10,
		defaultBufferSize,
		10 * defaultBufferSize,
	} {
		b.Run(fmt.Sprintf("payload_size=%d", payloadSize), func(b *testing.B) {
			signer := test.RandomSigner(b)
			obj := objecttest.Object(b)

			payload := make([]byte, payloadSize)
			_, err := rand.Read(payload)
			require.NoError(b, err)
			obj.SetPayload(payload)

			bObj, err := obj.Marshal()
			require.NoError(b, err)

			f := func(singleBuffer bool) {
				b.Run(fmt.Sprintf("single_buffer=%t", singleBuffer), func(b *testing.B) {
					var stream discardBinaryMessageWriter

					b.ReportAllocs()
					b.ResetTimer()

					for i := 0; i < b.N; i++ {
						_ = streamBinaryObject(stream, bytes.NewReader(bObj), signer, singleBuffer)
					}
				})
			}

			f(false)
			f(true)
		})
	}
}
