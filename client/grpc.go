package client

import (
	"context"
	"fmt"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/mem"
	"google.golang.org/protobuf/proto"
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

func callUnary(ctx context.Context, conn *grpc.ClientConn, method string, request any, response proto.Message) error {
	return conn.Invoke(ctx, method, request, response,
		grpc.StaticMethod(),
		grpc.ForceCodecV2(protobuf.BufferedCodec{}),
	)
}

func appendVerificationHeader(signer neofscrypto.Signer, reqBuf []byte, bodyWithMetaHdrLen int, body []byte, metaHdr []byte, vers *protorefs.Version) (mem.BufferSlice, error) {
	bodySig, metaHdrSig, originVerifHdrSig, err := calculateRequestSignatures(signer, body, metaHdr, vers)
	if err != nil {
		return nil, err
	}

	// pre-calculate verification header message lengths
	bodySigMsgLen := calculateSignatureFieldLength(bodySig)
	metaHdrSigMsgLen := calculateSignatureFieldLength(metaHdrSig)
	originVerifHdrSigMsgLen := calculateSignatureFieldLength(originVerifHdrSig)

	verifHdrLen := calculateRequestVerificationHeaderLength(bodySigMsgLen, metaHdrSigMsgLen, originVerifHdrSigMsgLen)

	verifHdrFldLen := neofsproto.SizeEmbeddedLENField(protobuf.FieldRequestVerificationHeader, verifHdrLen)

	// acquire buffer for verification header
	var verifHdrFldBuf []byte
	var reqBuffers mem.BufferSlice
	if len(reqBuf) >= bodyWithMetaHdrLen+verifHdrFldLen {
		verifHdrFldBuf = reqBuf[bodyWithMetaHdrLen:][:verifHdrFldLen]
		reqBuffers = mem.BufferSlice{mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen+verifHdrFldLen])}
	} else {
		verifHdrFldBuf = make([]byte, verifHdrFldLen)
		reqBuffers = mem.BufferSlice{mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen]), mem.SliceBuffer(verifHdrFldBuf)}
	}

	// encode verification header
	writeRequestVerificationHeader(verifHdrFldBuf, verifHdrLen, bodySigMsgLen, bodySig, metaHdrSigMsgLen, metaHdrSig, originVerifHdrSigMsgLen, originVerifHdrSig)

	return reqBuffers, nil
}
