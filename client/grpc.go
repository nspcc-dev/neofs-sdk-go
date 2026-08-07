package client

import (
	"context"
	"fmt"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	igrpc "github.com/nspcc-dev/neofs-sdk-go/internal/grpc"
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
	unaryStreamDesc = &grpc.StreamDesc{
		ServerStreams: false,
		ClientStreams: false,
	}
	grpcCallOptions = []grpc.CallOption{
		grpc.StaticMethod(),
		grpc.ForceCodecV2(protobuf.BufferedCodec{}),
	}
)

func sendRequest(ctx context.Context, conn *grpc.ClientConn, method string, streamDesc *grpc.StreamDesc, request mem.BufferSlice) (grpc.ClientStream, error) {
	stream, err := conn.NewStream(ctx, streamDesc, method, grpcCallOptions...)
	if err != nil {
		request.Free()
		return nil, fmt.Errorf("stream opening failed: %w", err)
	}

	if err = stream.SendMsg(request); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	return stream, nil
}

func callServerStream(ctx context.Context, conn *grpc.ClientConn, method string, streamDesc *grpc.StreamDesc, request mem.BufferSlice) (grpc.ClientStream, error) {
	stream, err := sendRequest(ctx, conn, method, streamDesc, request)
	if err != nil {
		return nil, err
	}

	if err = stream.CloseSend(); err != nil {
		return nil, fmt.Errorf("close send: %w", err)
	}

	return stream, nil
}

func callUnary(ctx context.Context, conn *grpc.ClientConn, method string, request mem.BufferSlice, response proto.Message) error {
	stream, err := sendRequest(ctx, conn, method, unaryStreamDesc, request)
	if err != nil {
		return err
	}

	return stream.RecvMsg(response)
}

func appendVerificationHeader(signer neofscrypto.Signer, reqMemBuf *igrpc.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, body []byte, metaHdr []byte, vers *protorefs.Version) (mem.BufferSlice, error) {
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
		reqSliceBuf := mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen+verifHdrFldLen])
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqSliceBuf
			reqBuffers = mem.BufferSlice{reqMemBuf}
		} else {
			reqBuffers = mem.BufferSlice{reqSliceBuf}
		}
	} else {
		verifHdrFldBuf = make([]byte, verifHdrFldLen)
		reqSliceBuf := mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen])
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqSliceBuf
			reqBuffers = mem.BufferSlice{reqMemBuf, mem.SliceBuffer(verifHdrFldBuf)}
		} else {
			reqBuffers = mem.BufferSlice{reqSliceBuf, mem.SliceBuffer(verifHdrFldBuf)}
		}
	}

	// encode verification header
	writeRequestVerificationHeader(verifHdrFldBuf, verifHdrLen, bodySigMsgLen, bodySig, metaHdrSigMsgLen, metaHdrSig, originVerifHdrSigMsgLen, originVerifHdrSig)

	return reqBuffers, nil
}
