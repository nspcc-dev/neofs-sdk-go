package client

import (
	"context"
	"encoding/binary"
	"fmt"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	igrpc "github.com/nspcc-dev/neofs-sdk-go/internal/grpc"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
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
		bodySig, metaHdrSig, originVerifHdrSig, err := calculateRequestSignatures(signer, body, metaHdr, vers)
		if err != nil {
			return nil, err
		}
		return appendVerificationHeaderSignatures2(reqMemBuf, reqBuf, bodyWithMetaHdrLen, bodySig, metaHdrSig, originVerifHdrSig), nil
	}

	sigRaw, err := signer.Sign(reqBuf[:bodyWithMetaHdrLen])
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	var (
		pubKey = make([]byte, signer.Public().MaxEncodedSize())
		n      = signer.Public().Encode(pubKey)
		reqSig = neofscrypto.NewSignatureFromRawKey(signer.Scheme(), pubKey[:n], sigRaw)
	)

	return appendVerificationHeaderSignature(nil, reqMemBuf, reqBuf, bodyWithMetaHdrLen, reqSig), nil
}

func _appendVerificationHeaderSignatures(reqMemBuf *igrpc.MemBuffer, reqMemBuf2 *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, bodySig, metaHdrSig, originVerifHdrSig neofscrypto.Signature) mem.BufferSlice {
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
	writeRequestVerificationHeader(verifHdrFldBuf, verifHdrLen, bodySigMsgLen, bodySig, metaHdrSigMsgLen, metaHdrSig, originVerifHdrSigMsgLen, originVerifHdrSig)

	return reqBuffers
}

func appendVerificationHeaderSignatures(reqMemBuf *igrpc.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, bodySig, metaHdrSig, originVerifHdrSig neofscrypto.Signature) mem.BufferSlice {
	return _appendVerificationHeaderSignatures(reqMemBuf, nil, reqBuf, bodyWithMetaHdrLen, bodySig, metaHdrSig, originVerifHdrSig)
}

func appendVerificationHeaderSignatures2(reqMemBuf *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, bodySig, metaHdrSig, originVerifHdrSig neofscrypto.Signature) mem.BufferSlice {
	return _appendVerificationHeaderSignatures(nil, reqMemBuf, reqBuf, bodyWithMetaHdrLen, bodySig, metaHdrSig, originVerifHdrSig)
}

func appendVerificationHeaderSignature(reqMemBuf *igrpc.MemBuffer, reqMemBuf2 *protobuf.MemBuffer, reqBuf []byte, bodyWithMetaHdrLen int, reqSig neofscrypto.Signature) mem.BufferSlice {
	// pre-calculate verification header message lengths
	var (
		sigMsgLen      = calculateSignatureFieldLength(reqSig)
		verifHdrLen    = neofsproto.SizeEmbeddedLENField(protosession.FieldRequestVerificationHeaderRequestSignature, sigMsgLen)
		verifHdrFldLen = neofsproto.SizeEmbeddedLENField(protobuf.FieldRequestVerificationHeader, verifHdrLen)
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
	off := binary.PutUvarint(verifHdrFldBuf, protobuf.TagBytes3)
	off += binary.PutUvarint(verifHdrFldBuf[off:], uint64(verifHdrLen))
	writeEmbeddedSignatureField(verifHdrFldBuf[off:], protosession.FieldRequestVerificationHeaderRequestSignature, sigMsgLen, reqSig)

	return reqBuffers
}
