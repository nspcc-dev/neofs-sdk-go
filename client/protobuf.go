package client

import (
	"encoding/binary"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	iproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/protobuf"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protowire"
)

// TODO: This code is also in demand by https://github.com/nspcc-dev/neofs-node/issues/4005.
//  Think of the best place for it and share.

// set by init.
var currentVersionRequestMetaHeader []byte

func initCurrentVersionRequestMetaHeader() {
	ver := version.Current()
	verMsg := ver.ProtoMessage()
	verMsgLen := verMsg.MarshaledSize()

	verFldLen := iproto.SizeEmbeddedLENField(protosession.FieldRequestMetaHeaderVersion, verMsgLen)
	if verFldLen == 0 {
		return
	}

	b := make([]byte, verFldLen)
	iproto.MarshalToEmbeddedLength(b, protosession.FieldRequestMetaHeaderVersion, verMsgLen, verMsg)
	currentVersionRequestMetaHeader = b
}

func calculateRequestVerificationHeaderLength(bodySigLen int, metaSigLen int, originSigLen int) int {
	ln := iproto.SizeEmbeddedLENField(protosession.FieldRequestVerificationHeaderBodySignature, bodySigLen)
	ln += iproto.SizeEmbeddedLENField(protosession.FieldRequestVerificationHeaderMetaSignature, metaSigLen)
	ln += iproto.SizeEmbeddedLENField(protosession.FieldRequestVerificationHeaderOriginSignature, originSigLen)
	return ln
}

func calculateXHeaderLength(key string, value string) int {
	ln := iproto.SizeBytes(protosession.FieldXHeaderKey, key)
	ln += iproto.SizeBytes(protosession.FieldXHeaderValue, value)
	return ln
}

func writeXHeader(buf []byte, tag uint64, key string, value string) int {
	xln := calculateXHeaderLength(key, value)
	if xln == 0 {
		return 0
	}
	off := binary.PutUvarint(buf, tag)
	off += binary.PutUvarint(buf[off:], uint64(xln))
	off += iproto.MarshalToBytes(buf[off:], protosession.FieldXHeaderKey, key)
	off += iproto.MarshalToBytes(buf[off:], protosession.FieldXHeaderValue, value)
	return off
}

func calculateRequestXHeadersLength(xHeaders []string) int {
	var ln int
	for i := 0; i < len(xHeaders); i += 2 {
		lni := calculateXHeaderLength(xHeaders[i], xHeaders[i+1])
		ln += iproto.SizeEmbeddedLENField(protosession.FieldRequestMetaHeaderXHeaders, lni)
	}
	return ln
}

func writeXHeaders(buf []byte, num int, xHeaders []string) int {
	var off int
	tag := protowire.EncodeTag(protowire.Number(num), protowire.BytesType)
	for i := 0; i < len(xHeaders); i += 2 {
		off += writeXHeader(buf[off:], tag, xHeaders[i], xHeaders[i+1])
	}
	return off
}

func localFlagToTTL(local bool) uint32 {
	if local {
		return localRequestTTL
	}
	return defaultRequestTTL
}

func calculateRequestMetaHeaderFieldLengths(local bool, xHeadersLen int, sessionV1TokenLen int, bearerTokenLen int, sessionV2TokenLen int) int {
	ln := len(currentVersionRequestMetaHeader)
	ln += iproto.SizeVarint(protosession.FieldRequestMetaHeaderTTL, localFlagToTTL(local))
	ln += xHeadersLen
	ln += iproto.SizeEmbeddedLENField(protosession.FieldRequestMetaHeaderSessionToken, sessionV1TokenLen)
	ln += iproto.SizeEmbeddedLENField(protosession.FieldRequestMetaHeaderBearerToken, bearerTokenLen)
	ln += iproto.SizeEmbeddedLENField(protosession.FieldRequestMetaHeaderSessionTokenV2, sessionV2TokenLen)
	return ln
}

func writeRequestMetaHeader(buf []byte, ln int, local bool, xHeaders []string, sessionV1TokenLen int, sessionV1TokenMsg *protosession.SessionToken, bearerTokenLen int, bearerTokenMsg *protoacl.BearerToken, sessionV2TokenLen int, sessionV2TokenMsg *protosession.SessionTokenV2) int {
	off := binary.PutUvarint(buf, protobuf.TagBytes2)
	off += binary.PutUvarint(buf[off:], uint64(ln))
	off += copy(buf[off:], currentVersionRequestMetaHeader)
	off += iproto.MarshalToVarint(buf[off:], protosession.FieldRequestMetaHeaderTTL, localFlagToTTL(local))
	off += writeXHeaders(buf[off:], protosession.FieldRequestMetaHeaderXHeaders, xHeaders)
	off += iproto.MarshalToEmbeddedLength(buf[off:], protosession.FieldRequestMetaHeaderSessionToken, sessionV1TokenLen, sessionV1TokenMsg)
	off += iproto.MarshalToEmbeddedLength(buf[off:], protosession.FieldRequestMetaHeaderBearerToken, bearerTokenLen, bearerTokenMsg)
	off += iproto.MarshalToEmbeddedLength(buf[off:], protosession.FieldRequestMetaHeaderSessionTokenV2, sessionV2TokenLen, sessionV2TokenMsg)
	return off
}

func calculateSignatureFieldLength(sig neofscrypto.Signature) int {
	ln := iproto.SizeEmbeddedLENField(protorefs.FieldSignatureKey, len(sig.PublicKeyBytes()))
	ln += iproto.SizeEmbeddedLENField(protorefs.FieldSignatureValue, len(sig.Value()))
	ln += iproto.SizeVarint(protorefs.FieldSignatureScheme, sig.Scheme())
	return ln
}

func writeEmbeddedSignatureField(buf []byte, num int, ln int, sig neofscrypto.Signature) int {
	if ln == 0 {
		return 0
	}
	off := binary.PutUvarint(buf, protowire.EncodeTag(protowire.Number(num), protowire.BytesType))
	off += binary.PutUvarint(buf[off:], uint64(ln))
	off += iproto.MarshalToBytes(buf[off:], protorefs.FieldSignatureKey, sig.PublicKeyBytes())
	off += iproto.MarshalToBytes(buf[off:], protorefs.FieldSignatureValue, sig.Value())
	off += iproto.MarshalToVarint(buf[off:], protorefs.FieldSignatureScheme, sig.Scheme())
	return off
}

func writeEmbeddedObjectAddressField(buf []byte, num int, cnr cid.ID, obj oid.ID) int {
	off := binary.PutUvarint(buf, protowire.EncodeTag(protowire.Number(num), protowire.BytesType))
	buf[off] = protobuf.ObjectAddressLength
	off++
	buf[off] = protobuf.TagBytes1
	off++
	buf[off] = protobuf.ContainerIDLength
	off++
	off += iproto.MarshalToBytes(buf[off:], protorefs.FieldContainerIDValue, cnr[:])
	buf[off] = protobuf.TagBytes2
	off++
	buf[off] = protobuf.ObjectIDLength
	off++
	off += iproto.MarshalToBytes(buf[off:], protorefs.FieldObjectIDValue, obj[:])
	return off
}

func writeRequestVerificationHeader(buf []byte, ln int, bodySigFldLen int, bodySig neofscrypto.Signature, metaHdrSigFldLen int, metaHdrSig neofscrypto.Signature, originVerifHdrSigFldLen int, originVerifHdrSig neofscrypto.Signature) int {
	off := binary.PutUvarint(buf, protobuf.TagBytes3)
	off += binary.PutUvarint(buf[off:], uint64(ln))
	off += writeEmbeddedSignatureField(buf[off:], protosession.FieldRequestVerificationHeaderBodySignature, bodySigFldLen, bodySig)
	off += writeEmbeddedSignatureField(buf[off:], protosession.FieldRequestVerificationHeaderMetaSignature, metaHdrSigFldLen, metaHdrSig)
	off += writeEmbeddedSignatureField(buf[off:], protosession.FieldRequestVerificationHeaderOriginSignature, originVerifHdrSigFldLen, originVerifHdrSig)
	return off
}

func calculateGetObjectRequestBodyLength(raw bool, rngLen int, payloadOnly bool, extendedRngLen int) int {
	ln := iproto.SizeEmbeddedLENField(protoobject.FieldGetRequestBodyAddress, protobuf.ObjectAddressLength)
	ln += iproto.SizeBool(protoobject.FieldGetRequestBodyRaw, raw)
	ln += iproto.SizeEmbeddedLENField(protoobject.FieldGetRequestBodyRange, rngLen)
	ln += iproto.SizeBool(protoobject.FieldGetRequestBodyPayloadOnly, payloadOnly)
	ln += iproto.SizeEmbeddedLENField(protoobject.FieldGetRequestBodyExtendedRange, extendedRngLen)
	return ln
}

func writeGetRequestBody(buf []byte, ln int, cnr cid.ID, obj oid.ID, raw bool, rngLen int, rng *protoobject.Range, payloadOnly bool, extRngLen int, extRng *protoobject.ExtendedRange) int {
	if ln == 0 {
		return 0
	}
	buf[0] = protobuf.TagBytes1
	off := 1 + binary.PutUvarint(buf[1:], uint64(ln))
	off += writeEmbeddedObjectAddressField(buf[off:], protoobject.FieldGetRequestBodyAddress, cnr, obj)
	off += iproto.MarshalToBool(buf[off:], protoobject.FieldGetRequestBodyRaw, raw)
	off += iproto.MarshalToEmbeddedLength(buf[off:], protoobject.FieldGetRequestBodyRange, rngLen, rng)
	off += iproto.MarshalToBool(buf[off:], protoobject.FieldGetRequestBodyPayloadOnly, payloadOnly)
	off += iproto.MarshalToEmbeddedLength(buf[off:], protoobject.FieldGetRequestBodyExtendedRange, extRngLen, extRng)
	return off
}
