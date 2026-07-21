package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	"google.golang.org/grpc"
	grpcproto "google.golang.org/grpc/encoding/proto"
)

var (
	getStreamDesc = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: false,
	}
)

func callServerStream(ctx context.Context, conn *grpc.ClientConn, method string, streamDesc *grpc.StreamDesc, request any) (grpc.ClientStream, error) {
	stream, err := conn.NewStream(ctx, streamDesc, method,
		grpc.StaticMethod(),
		grpc.ForceCodecV2(protobuf.BufferedCodec{}),
		grpc.CallContentSubtype(grpcproto.Name), // TODO: check the need
	)
	if err != nil {
		return nil, fmt.Errorf("stream opening failed: %w", err)
	}

	if err = stream.SendMsg(request); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if err = stream.CloseSend(); err != nil {
		return nil, fmt.Errorf("close send: %w", err)
	}

	return stream, nil
}
