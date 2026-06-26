package client

import (
	"bytes"
	"context"
	"io"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func BenchmarkPayloadReaderWriteTo(b *testing.B) {
	ctx := context.Background()
	var opts PrmObjectGet
	opts.SkipChecksumVerification()
	cnr := cidtest.ID()
	obj := oidtest.ID()
	signer := usertest.User()

	for _, tc := range []struct {
		name      string
		chunkSize int
		chunks    int
	}{
		{name: "4KiB", chunkSize: 4 << 10, chunks: 1},
		{name: "1MiB", chunkSize: 32 << 10, chunks: 32},
	} {
		b.Run(tc.name, func(b *testing.B) {
			payload := bytes.Repeat([]byte{1}, tc.chunkSize)
			chunks := make([][]byte, tc.chunks)
			for i := range chunks {
				chunks[i] = payload
			}

			srv := newTestGetObjectServer()
			srv.respondWithObject(proto.Clone(validFullHeadingObjectGetResponseBody.GetInit()).(*protoobject.GetResponse_Body_Init), chunks)
			c := newTestObjectClient(b, srv)

			b.SetBytes(int64(tc.chunkSize * tc.chunks))
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, r, err := c.ObjectGetInit(ctx, cnr, obj, signer, opts)
				require.NoError(b, err)
				n, err := r.WriteTo(io.Discard)
				require.NoError(b, err)
				require.Equal(b, int64(tc.chunkSize*tc.chunks), n)
			}
		})
	}
}
