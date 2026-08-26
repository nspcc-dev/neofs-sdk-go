package client

import (
	"context"
	"fmt"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	igrpc "github.com/nspcc-dev/neofs-sdk-go/internal/grpc"
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	"github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
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
	clientStreamDesc = &grpc.StreamDesc{
		ServerStreams: false,
		ClientStreams: true,
	}
	grpcCallOptions = []grpc.CallOption{
		grpc.StaticMethod(),
		grpc.ForceCodecV2(protobuf.BufferedCodec{}),
	}
)

func openStream(ctx context.Context, conn *grpc.ClientConn, streamDesc *grpc.StreamDesc, method string) (grpc.ClientStream, error) {
	stream, err := conn.NewStream(ctx, streamDesc, method, grpcCallOptions...)
	if err != nil {
		return nil, fmt.Errorf("stream opening failed: %w", err)
	}

	return stream, nil
}

func sendRequest(ctx context.Context, conn *grpc.ClientConn, method string, streamDesc *grpc.StreamDesc, request mem.BufferSlice) (grpc.ClientStream, error) {
	stream, err := openStream(ctx, conn, streamDesc, method)
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

func appendVerificationHeader(signer neofscrypto.Signer, reqMemBuf *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, body []byte, metaHdr []byte, vers *protorefs.Version) (mem.BufferSlice, error) {
	if multipleReqSignatures(vers) {
		bodySig, metaHdrSig, originVerifHdrSig, err := signRequestParts(signer, body, metaHdr, needsOriginSig(vers))
		if err != nil {
			return nil, err
		}

		pubKeyBytes := neofscrypto.PublicKeyBytes(signer.Public())
		scheme := signer.Scheme()

		return appendVerificationHeaderSignatures2(reqMemBuf, reqBuf, bodyWithMetaHdrLen, pubKeyBytes, scheme, bodySig, metaHdrSig, originVerifHdrSig), nil
	}

	sigRaw, err := signer.Sign(reqBuf[:bodyWithMetaHdrLen])
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	var (
		pubKey = make([]byte, signer.Public().MaxEncodedSize())
		n      = signer.Public().Encode(pubKey)
	)

	return appendVerificationHeaderSignature(nil, reqMemBuf, reqBuf, bodyWithMetaHdrLen, pubKey[:n], sigRaw, signer.Scheme()), nil
}

func _appendVerificationHeaderSignatures(reqMemBuf *igrpc.MemBuffer, reqMemBuf2 *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, pubKey []byte, scheme neofscrypto.Scheme, bodySig, metaHdrSig, originVerifHdrSig []byte) mem.BufferSlice {
	// pre-calculate verification header message lengths
	verifHdrLen := protosession.CalculateMultiSignatureRequestVerificationHeaderLength(pubKey, scheme, bodySig, metaHdrSig, originVerifHdrSig)

	verifHdrFldLen := protoencoding.CalculateRequestVerificationHeaderFieldLength(verifHdrLen)

	// acquire buffer for verification header
	var verifHdrFldBuf []byte
	var reqBuffers mem.BufferSlice
	if len(reqBuf) >= bodyWithMetaHdrLen+verifHdrFldLen {
		verifHdrFldBuf = reqBuf[bodyWithMetaHdrLen:][:verifHdrFldLen]
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqBuf[:bodyWithMetaHdrLen+verifHdrFldLen]
			reqBuffers = mem.BufferSlice{reqMemBuf}
		} else if reqMemBuf2 != nil {
			reqMemBuf2.SetBounds(0, bodyWithMetaHdrLen+verifHdrFldLen)
			reqBuffers = mem.BufferSlice{reqMemBuf2}
		} else {
			reqBuffers = mem.BufferSlice{mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen+verifHdrFldLen])}
		}
	} else {
		verifHdrFldBuf = make([]byte, verifHdrFldLen)
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqBuf[:bodyWithMetaHdrLen]
			reqBuffers = mem.BufferSlice{reqMemBuf, mem.SliceBuffer(verifHdrFldBuf)}
		} else if reqMemBuf2 != nil {
			reqMemBuf2.SetBounds(0, bodyWithMetaHdrLen)
			reqBuffers = mem.BufferSlice{reqMemBuf2, mem.SliceBuffer(verifHdrFldBuf)}
		} else {
			reqBuffers = mem.BufferSlice{mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen]), mem.SliceBuffer(verifHdrFldBuf)}
		}
	}

	// encode verification header
	protosession.WriteMultiSignatureRequestVerificationHeaderToRequest(verifHdrFldBuf, pubKey, scheme, bodySig, metaHdrSig, originVerifHdrSig)

	return reqBuffers
}

func appendVerificationHeaderSignatures(reqMemBuf *igrpc.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, pubKey []byte, scheme neofscrypto.Scheme, bodySig, metaHdrSig, originVerifHdrSig []byte) mem.BufferSlice {
	return _appendVerificationHeaderSignatures(reqMemBuf, nil, reqBuf, bodyWithMetaHdrLen, pubKey, scheme, bodySig, metaHdrSig, originVerifHdrSig)
}

func appendVerificationHeaderSignatures2(reqMemBuf *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, pubKey []byte, scheme neofscrypto.Scheme, bodySig, metaHdrSig, originVerifHdrSig []byte) mem.BufferSlice {
	return _appendVerificationHeaderSignatures(nil, reqMemBuf, reqBuf, bodyWithMetaHdrLen, pubKey, scheme, bodySig, metaHdrSig, originVerifHdrSig)
}

func appendVerificationHeaderSignature(reqMemBuf *igrpc.MemBuffer, reqMemBuf2 *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, pubKey []byte, value []byte, scheme neofscrypto.Scheme) mem.BufferSlice {
	// pre-calculate verification header message lengths
	var (
		verifHdrLen    = protosession.CalculateSingleSignatureRequestVerificationHeaderLength(pubKey, scheme, value)
		verifHdrFldLen = protoencoding.CalculateRequestVerificationHeaderFieldLength(verifHdrLen)
	)

	// acquire buffer for verification header
	var verifHdrFldBuf []byte
	var reqBuffers mem.BufferSlice
	if len(reqBuf) >= bodyWithMetaHdrLen+verifHdrFldLen {
		verifHdrFldBuf = reqBuf[bodyWithMetaHdrLen:][:verifHdrFldLen]
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqBuf[:bodyWithMetaHdrLen+verifHdrFldLen]
			reqBuffers = mem.BufferSlice{reqMemBuf}
		} else if reqMemBuf2 != nil {
			reqMemBuf2.SetBounds(0, bodyWithMetaHdrLen+verifHdrFldLen)
			reqBuffers = mem.BufferSlice{reqMemBuf2}
		} else {
			reqBuffers = mem.BufferSlice{mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen+verifHdrFldLen])}
		}
	} else {
		verifHdrFldBuf = make([]byte, verifHdrFldLen)
		if reqMemBuf != nil {
			reqMemBuf.SliceBuffer = reqBuf[:bodyWithMetaHdrLen]
			reqBuffers = mem.BufferSlice{reqMemBuf, mem.SliceBuffer(verifHdrFldBuf)}
		} else if reqMemBuf2 != nil {
			reqMemBuf2.SetBounds(0, bodyWithMetaHdrLen)
			reqBuffers = mem.BufferSlice{reqMemBuf2, mem.SliceBuffer(verifHdrFldBuf)}
		} else {
			reqBuffers = mem.BufferSlice{mem.SliceBuffer(reqBuf[:bodyWithMetaHdrLen]), mem.SliceBuffer(verifHdrFldBuf)}
		}
	}

	// encode verification header
	protosession.WriteSingleSignatureRequestVerificationHeaderToRequest(verifHdrFldBuf, pubKey, scheme, value)

	return reqBuffers
}
