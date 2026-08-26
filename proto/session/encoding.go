package session

import (
	"fmt"

	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
)

// Field numbers of [ObjectSessionContext_Target] message.
const (
	_ = iota
	FieldObjectSessionContextTargetContainer
	FieldObjectSessionContextTargetObjects
)

// MarshaledSize returns size of the ObjectSessionContext_Target in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *ObjectSessionContext_Target) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeEmbedded(FieldObjectSessionContextTargetContainer, x.Container) +
			protoencoding.SizeRepeatedMessages(FieldObjectSessionContextTargetObjects, x.Objects)
	}
	return 0
}

// MarshalStable writes the ObjectSessionContext_Target in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ObjectSessionContext_Target.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ObjectSessionContext_Target) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldObjectSessionContextTargetContainer, x.Container)
		protoencoding.MarshalToRepeatedMessages(b[off:], FieldObjectSessionContextTargetObjects, x.Objects)
	}
}

// Field numbers of [ObjectSessionContext] message.
const (
	_ = iota
	FieldObjectSessionContextVerb
	FieldObjectSessionContextTarget
)

// MarshaledSize returns size of the ObjectSessionContext in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ObjectSessionContext) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeVarint(FieldObjectSessionContextVerb, x.Verb) +
			protoencoding.SizeEmbedded(FieldObjectSessionContextTarget, x.Target)
	}
	return sz
}

// MarshalStable writes the ObjectSessionContext in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [ObjectSessionContext.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ObjectSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldObjectSessionContextVerb, x.Verb)
		protoencoding.MarshalToEmbedded(b[off:], FieldObjectSessionContextTarget, x.Target)
	}
}

// Field numbers of [ContainerSessionContext] message.
const (
	_ = iota
	FieldContainerSessionContextVerb
	FieldContainerSessionContextWildcard
	FieldContainerSessionContextContainerID
)

// MarshaledSize returns size of the ContainerSessionContext in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *ContainerSessionContext) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeVarint(FieldContainerSessionContextVerb, x.Verb) +
			protoencoding.SizeBool(FieldContainerSessionContextWildcard, x.Wildcard) +
			protoencoding.SizeEmbedded(FieldContainerSessionContextContainerID, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the ContainerSessionContext in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ContainerSessionContext.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ContainerSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldContainerSessionContextVerb, x.Verb)
		off += protoencoding.MarshalToBool(b[off:], FieldContainerSessionContextWildcard, x.Wildcard)
		protoencoding.MarshalToEmbedded(b[off:], FieldContainerSessionContextContainerID, x.ContainerId)
	}
}

// Field numbers of [SessionToken_Body_TokenLifetime] message.
const (
	_ = iota
	FieldSessionTokenBodyTokenLifetimeExp
	FieldSessionTokenBodyTokenLifetimeNbf
	FieldSessionTokenBodyTokenLifetimeIat
)

// MarshaledSize returns size of the SessionToken_Body_TokenLifetime in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SessionToken_Body_TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeVarint(FieldSessionTokenBodyTokenLifetimeExp, x.Exp) +
			protoencoding.SizeVarint(FieldSessionTokenBodyTokenLifetimeNbf, x.Nbf) +
			protoencoding.SizeVarint(FieldSessionTokenBodyTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the SessionToken_Body_TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [SessionToken_Body_TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SessionToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldSessionTokenBodyTokenLifetimeExp, x.Exp)
		off += protoencoding.MarshalToVarint(b[off:], FieldSessionTokenBodyTokenLifetimeNbf, x.Nbf)
		protoencoding.MarshalToVarint(b[off:], FieldSessionTokenBodyTokenLifetimeIat, x.Iat)
	}
}

// Field numbers of [SessionToken_Body] message.
const (
	_ = iota
	FieldSessionTokenBodyID
	FieldSessionTokenBodyOwnerID
	FieldSessionTokenBodyLifetime
	FieldSessionTokenBodySessionKey
	FieldSessionTokenBodyObject
	FieldSessionTokenBodyContainer
)

// MarshaledSize returns size of the SessionToken_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionToken_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeBytes(FieldSessionTokenBodyID, x.Id) +
			protoencoding.SizeEmbedded(FieldSessionTokenBodyOwnerID, x.OwnerId) +
			protoencoding.SizeEmbedded(FieldSessionTokenBodyLifetime, x.Lifetime) +
			protoencoding.SizeBytes(FieldSessionTokenBodySessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			sz += protoencoding.SizeEmbedded(FieldSessionTokenBodyObject, c.Object)
		case *SessionToken_Body_Container:
			sz += protoencoding.SizeEmbedded(FieldSessionTokenBodyContainer, c.Container)
		}
	}
	return sz
}

// MarshalStable writes the SessionToken_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionToken_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SessionToken_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldSessionTokenBodyID, x.Id)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenBodyOwnerID, x.OwnerId)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenBodyLifetime, x.Lifetime)
		off += protoencoding.MarshalToBytes(b[off:], FieldSessionTokenBodySessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenBodyObject, c.Object)
		case *SessionToken_Body_Container:
			protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenBodyContainer, c.Container)
		}
	}
}

// Field numbers of [SessionToken] message.
const (
	_ = iota
	FieldSessionTokenBody
	FieldSessionTokenSignature
)

// MarshaledSize returns size of the SessionToken in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *SessionToken) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldSessionTokenBody, x.Body) +
			protoencoding.SizeEmbedded(FieldSessionTokenSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the SessionToken in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionToken.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SessionToken) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSessionTokenBody, x.Body)
		protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenSignature, x.Signature)
	}
}

// Field numbers of [CreateRequest_Body] message.
const (
	_ = iota
	FieldCreateRequestBodyOwnerID
	FieldCreateRequestBodyExpiration
)

// MarshaledSize returns size of the CreateRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *CreateRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldCreateRequestBodyOwnerID, x.OwnerId) +
			protoencoding.SizeVarint(FieldCreateRequestBodyExpiration, x.Expiration)
	}
	return sz
}

// MarshalStable writes the CreateRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [CreateRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *CreateRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldCreateRequestBodyOwnerID, x.OwnerId)
		protoencoding.MarshalToVarint(b[off:], FieldCreateRequestBodyExpiration, x.Expiration)
	}
}

// Field numbers of [CreateResponse_Body] message.
const (
	_ = iota
	FieldCreateResponseBodyID
	FieldCreateResponseBodySessionKey
)

// MarshaledSize returns size of the CreateResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *CreateResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeBytes(FieldCreateResponseBodyID, x.Id) +
			protoencoding.SizeBytes(FieldCreateResponseBodySessionKey, x.SessionKey)
	}
	return sz
}

// MarshalStable writes the CreateResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [CreateResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *CreateResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldCreateResponseBodyID, x.Id)
		protoencoding.MarshalToBytes(b[off:], FieldCreateResponseBodySessionKey, x.SessionKey)
	}
}

// Field numbers of [XHeader] message.
const (
	_ = iota
	FieldXHeaderKey
	FieldXHeaderValue
)

// CalculateXHeaderLength calculates length of X-header message with given
// fields.
func CalculateXHeaderLength(key string, value string) int {
	ln := protoencoding.SizeBytes(FieldXHeaderKey, key)
	ln += protoencoding.SizeBytes(FieldXHeaderValue, value)
	return ln
}

// MarshaledSize returns size of the XHeader in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *XHeader) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return CalculateXHeaderLength(x.Key, x.Value)
}

// WriteXHeader writes X-header message with given fields into buf. Returns
// number of bytes written.
func WriteXHeader(buf []byte, key string, value string) int {
	off := protoencoding.MarshalToBytes(buf, FieldXHeaderKey, key)
	off += protoencoding.MarshalToBytes(buf[off:], FieldXHeaderValue, value)
	return off
}

// MarshalStable writes the XHeader in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [XHeader.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *XHeader) MarshalStable(b []byte) {
	if x != nil {
		WriteXHeader(b, x.Key, x.Value)
	}
}

// Field numbers of [RequestMetaHeader] message.
const (
	_ = iota
	FieldRequestMetaHeaderVersion
	FieldRequestMetaHeaderEpoch
	FieldRequestMetaHeaderTTL
	FieldRequestMetaHeaderXHeaders
	FieldRequestMetaHeaderSessionToken
	FieldRequestMetaHeaderBearerToken
	FieldRequestMetaHeaderOrigin
	FieldRequestMetaHeaderMagicNumber
	FieldRequestMetaHeaderSessionTokenV2
)

// CalculateRequestMetaHeaderLength calculates length of request meta header
// message with given fields.
func CalculateRequestMetaHeaderLength(majorVersion uint32, minorVersion uint32, ttl uint32, xHdrNum int, xHdrLenFn protoencoding.RepeatedMessageLenFunc, sessionV1TokenLen int, bearerTokenLen int, magicNumber uint64, sessionV2TokenLen int) int {
	return calculateRequestMetaHeaderLength(majorVersion, minorVersion, 0, ttl, xHdrNum, xHdrLenFn, sessionV1TokenLen, bearerTokenLen, 0, magicNumber, sessionV2TokenLen)
}

func calculateRequestMetaHeaderLength(majorVersion uint32, minorVersion uint32, epoch uint64, ttl uint32, xHdrNum int, xHdrLenFn protoencoding.RepeatedMessageLenFunc, sessionV1TokenLen int, bearerTokenLen int, originLen int, magicNumber uint64, sessionV2TokenLen int) int {
	ln := protoencoding.SizeEmbeddedLENField(FieldRequestMetaHeaderVersion, refs.CalculateVersionLength(majorVersion, minorVersion))
	ln += protoencoding.SizeVarint(FieldRequestMetaHeaderEpoch, epoch)
	ln += protoencoding.SizeVarint(FieldRequestMetaHeaderTTL, ttl)
	ln += protoencoding.CalculateRepeatedFieldsLength(FieldRequestMetaHeaderXHeaders, xHdrNum, xHdrLenFn)
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestMetaHeaderSessionToken, sessionV1TokenLen)
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestMetaHeaderBearerToken, bearerTokenLen)
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestMetaHeaderOrigin, originLen)
	ln += protoencoding.SizeVarint(FieldRequestMetaHeaderMagicNumber, magicNumber)
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestMetaHeaderSessionTokenV2, sessionV2TokenLen)
	return ln
}

func (x *RequestMetaHeader) getXHeaderLen(i int) int {
	xh := x.XHeaders[i]
	if xh == nil {
		return 0
	}
	return CalculateXHeaderLength(xh.Key, xh.Value)
}

// MarshaledSize returns size of the RequestMetaHeader in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *RequestMetaHeader) MarshaledSize() int {
	if x == nil {
		return 0
	}
	sessionV1TokenLen := x.SessionToken.MarshaledSize()
	bearerTokenLen := x.BearerToken.MarshaledSize()
	originLen := x.Origin.MarshaledSize()
	sessionV2TokenLen := x.SessionTokenV2.MarshaledSize()
	return calculateRequestMetaHeaderLength(x.Version.GetMajor(), x.Version.GetMinor(), x.Epoch, x.Ttl, len(x.XHeaders), x.getXHeaderLen, sessionV1TokenLen, bearerTokenLen, originLen, x.MagicNumber, sessionV2TokenLen)
}

// WriteRequestMetaHeaderToRequest writes meta header field with given fields
// into buf. Returns number of bytes written.
func WriteRequestMetaHeaderToRequest(buf []byte, majorVersion uint32, minorVersion uint32, ttl uint32, xHdrNum int, xHdrLenFn protoencoding.RepeatedMessageLenFunc, writeXHdrFn protoencoding.WriteRepeatedMessageFunc,
	sessionV1TokenLen int, writeSessionV1TokenFn protoencoding.WriteMessageFunc, bearerTokenLen int, writeBearerTokenFn protoencoding.WriteMessageFunc, magicNumber uint64, sessionV2TokenLen int, writeSessionV2TokenFn protoencoding.WriteMessageFunc) int {
	ln := CalculateRequestMetaHeaderLength(majorVersion, minorVersion, ttl, xHdrNum, xHdrLenFn, sessionV1TokenLen, bearerTokenLen, magicNumber, sessionV2TokenLen)
	if ln == 0 {
		return 0
	}
	off := protoencoding.WriteRequestMetaHeaderTagAndLength(buf, ln)
	off += WriteRequestMetaHeader(buf[off:], majorVersion, minorVersion, ttl, xHdrNum, xHdrLenFn, writeXHdrFn, sessionV1TokenLen, writeSessionV1TokenFn, bearerTokenLen, writeBearerTokenFn, magicNumber, sessionV2TokenLen, writeSessionV2TokenFn)
	return off
}

// WriteRequestMetaHeader writes request meta header message with given fields
// into buf. Returns number of bytes written.
func WriteRequestMetaHeader(buf []byte, majorVersion uint32, minorVersion uint32, ttl uint32, xHdrNum int, xHdrLenFn protoencoding.RepeatedMessageLenFunc, writeXHdrFn protoencoding.WriteRepeatedMessageFunc,
	sessionV1TokenLen int, writeSessionV1TokenFn protoencoding.WriteMessageFunc, bearerTokenLen int, writeBearerTokenFn protoencoding.WriteMessageFunc, magicNumber uint64, sessionV2TokenLen int, writeSessionV2TokenFn protoencoding.WriteMessageFunc) int {
	return writeRequestMetaHeader(buf, majorVersion, minorVersion, 0, ttl, xHdrNum, xHdrLenFn, writeXHdrFn, sessionV1TokenLen, writeSessionV1TokenFn, bearerTokenLen, writeBearerTokenFn, 0, nil, magicNumber, sessionV2TokenLen, writeSessionV2TokenFn)
}

func writeRequestMetaHeader(buf []byte, majorVersion uint32, minorVersion uint32, epoch uint64, ttl uint32, xHdrNum int, xHdrLenFn protoencoding.RepeatedMessageLenFunc, writeXHdrFn protoencoding.WriteRepeatedMessageFunc,
	sessionV1TokenLen int, writeSessionV1TokenFn protoencoding.WriteMessageFunc, bearerTokenLen int, writeBearerTokenFn protoencoding.WriteMessageFunc, originLen int, writeOriginFn protoencoding.WriteMessageFunc, magicNumber uint64, sessionV2TokenLen int, writeSessionV2TokenFn protoencoding.WriteMessageFunc) int {
	off := refs.WriteVersionField(buf, FieldRequestMetaHeaderVersion, majorVersion, minorVersion)
	off += protoencoding.MarshalToVarint(buf[off:], FieldRequestMetaHeaderEpoch, epoch)
	off += protoencoding.MarshalToVarint(buf[off:], FieldRequestMetaHeaderTTL, ttl)
	off += protoencoding.WriteRepeatedFields(buf[off:], FieldRequestMetaHeaderXHeaders, xHdrNum, xHdrLenFn, writeXHdrFn)
	off += protoencoding.WriteMessageField(buf[off:], FieldRequestMetaHeaderSessionToken, sessionV1TokenLen, writeSessionV1TokenFn)
	off += protoencoding.WriteMessageField(buf[off:], FieldRequestMetaHeaderBearerToken, bearerTokenLen, writeBearerTokenFn)
	off += protoencoding.WriteMessageField(buf[off:], FieldRequestMetaHeaderOrigin, originLen, writeOriginFn)
	off += protoencoding.MarshalToVarint(buf[off:], FieldRequestMetaHeaderMagicNumber, magicNumber)
	off += protoencoding.WriteMessageField(buf[off:], FieldRequestMetaHeaderSessionTokenV2, sessionV2TokenLen, writeSessionV2TokenFn)
	return off
}

func (x *RequestMetaHeader) writeXHeader(buf []byte, i int) int {
	xh := x.XHeaders[i]
	if xh == nil {
		return 0
	}
	return WriteXHeader(buf, xh.Key, xh.Value)
}

// MarshalStable writes the RequestMetaHeader in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [RequestMetaHeader.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *RequestMetaHeader) MarshalStable(b []byte) {
	if x == nil {
		return
	}
	sessionV1TokenLen := x.SessionToken.MarshaledSize()
	writeSessionV1TokenFn := protoencoding.WriteStablyMarshalledMessageFunc(x.SessionToken)
	bearerTokenLen := x.BearerToken.MarshaledSize()
	writeBearerTokenFn := protoencoding.WriteStablyMarshalledMessageFunc(x.BearerToken)
	originLen := x.Origin.MarshaledSize()
	writeOriginFn := protoencoding.WriteStablyMarshalledMessageFunc(x.Origin)
	sessionV2TokenLen := x.SessionTokenV2.MarshaledSize()
	writeSessionV2TokenFn := protoencoding.WriteStablyMarshalledMessageFunc(x.SessionTokenV2)
	writeRequestMetaHeader(b, x.Version.GetMajor(), x.Version.GetMinor(), x.Epoch, x.Ttl, len(x.XHeaders), x.getXHeaderLen, x.writeXHeader, sessionV1TokenLen, writeSessionV1TokenFn, bearerTokenLen, writeBearerTokenFn, originLen, writeOriginFn, x.MagicNumber, sessionV2TokenLen, writeSessionV2TokenFn)
}

// Field numbers of [ResponseMetaHeader] message.
const (
	_ = iota
	FieldResponseMetaHeaderVersion
	FieldResponseMetaHeaderEpoch
	FieldResponseMetaHeaderTTL
	FieldResponseMetaHeaderXHeaders
	FieldResponseMetaHeaderOrigin
	FieldResponseMetaHeaderStatus
)

// MarshaledSize returns size of the ResponseMetaHeader in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ResponseMetaHeader) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeEmbedded(FieldResponseMetaHeaderVersion, x.Version) +
			protoencoding.SizeVarint(FieldResponseMetaHeaderEpoch, x.Epoch) +
			protoencoding.SizeVarint(FieldResponseMetaHeaderTTL, x.Ttl) +
			protoencoding.SizeEmbedded(FieldResponseMetaHeaderOrigin, x.Origin) +
			protoencoding.SizeEmbedded(FieldResponseMetaHeaderStatus, x.Status) +
			protoencoding.SizeRepeatedMessages(FieldResponseMetaHeaderXHeaders, x.XHeaders)
	}
	return 0
}

// MarshalStable writes the ResponseMetaHeader in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [ResponseMetaHeader.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ResponseMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldResponseMetaHeaderVersion, x.Version)
		off += protoencoding.MarshalToVarint(b[off:], FieldResponseMetaHeaderEpoch, x.Epoch)
		off += protoencoding.MarshalToVarint(b[off:], FieldResponseMetaHeaderTTL, x.Ttl)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldResponseMetaHeaderXHeaders, x.XHeaders)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldResponseMetaHeaderOrigin, x.Origin)
		protoencoding.MarshalToEmbedded(b[off:], FieldResponseMetaHeaderStatus, x.Status)
	}
}

// Field numbers of [RequestVerificationHeader] message.
const (
	_ = iota
	FieldRequestVerificationHeaderBodySignature
	FieldRequestVerificationHeaderMetaSignature
	FieldRequestVerificationHeaderOriginSignature
	FieldRequestVerificationHeaderOrigin
	FieldRequestVerificationHeaderRequestSignature
)

// CalculateMultiSignatureRequestVerificationHeaderLength calculates length of
// request verification header message with same public key and scheme, and
// given signatures.
func CalculateMultiSignatureRequestVerificationHeaderLength[SCHEME ~int32](pubKey []byte, scheme SCHEME, bodySig []byte, metaHdrSig []byte, originSig []byte) int {
	var originPubKey []byte
	var originScheme SCHEME
	if originSig != nil {
		originPubKey, originScheme = pubKey, scheme
	}
	return calculateRequestVerificationHeaderLength(pubKey, bodySig, scheme, pubKey, metaHdrSig, scheme, originPubKey, originSig, originScheme, 0, nil, nil, 0)
}

// CalculateSingleSignatureRequestVerificationHeaderLength calculates length of
// request verification with single request signature.
func CalculateSingleSignatureRequestVerificationHeaderLength[SCHEME ~int32](pubKey []byte, scheme SCHEME, reqSig []byte) int {
	return calculateRequestVerificationHeaderLength(nil, nil, 0, nil, nil, 0, nil, nil, 0, 0, pubKey, reqSig, scheme)
}

func calculateRequestVerificationHeaderLength[SCHEME ~int32](bodyPubKey []byte, bodySig []byte, bodyScheme SCHEME, metaPubKey []byte, metaSig []byte, metaScheme SCHEME, originPubKey []byte, originSig []byte, originScheme SCHEME, originLen int, reqPubKey []byte, reqSig []byte, reqScheme SCHEME) int {
	ln := protoencoding.SizeEmbeddedLENField(FieldRequestVerificationHeaderBodySignature, refs.CalculateSignatureLength(bodyPubKey, bodySig, bodyScheme))
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestVerificationHeaderMetaSignature, refs.CalculateSignatureLength(metaPubKey, metaSig, metaScheme))
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestVerificationHeaderOriginSignature, refs.CalculateSignatureLength(originPubKey, originSig, originScheme))
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestVerificationHeaderOrigin, originLen)
	ln += protoencoding.SizeEmbeddedLENField(FieldRequestVerificationHeaderRequestSignature, refs.CalculateSignatureLength(reqPubKey, reqSig, reqScheme))
	return ln
}

// MarshaledSize returns size of the RequestVerificationHeader in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *RequestVerificationHeader) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return calculateRequestVerificationHeaderLength(x.BodySignature.GetKey(), x.BodySignature.GetSign(), x.BodySignature.GetScheme(), x.MetaSignature.GetKey(), x.MetaSignature.GetSign(), x.MetaSignature.GetScheme(), x.OriginSignature.GetKey(), x.OriginSignature.GetSign(), x.OriginSignature.GetScheme(), x.Origin.MarshaledSize(), x.RequestSignature.GetKey(), x.RequestSignature.GetSign(), x.RequestSignature.GetScheme())
}

// WriteMultiSignatureRequestVerificationHeaderToRequest writes verification
// header field with same public key and scheme, and given signatures into buf.
// Returns number of bytes written.
func WriteMultiSignatureRequestVerificationHeaderToRequest[SCHEME ~int32](buf []byte, pubKey []byte, scheme SCHEME, bodySig []byte, metaHdrSig []byte, originSig []byte) int {
	ln := CalculateMultiSignatureRequestVerificationHeaderLength(pubKey, scheme, bodySig, metaHdrSig, originSig)
	if ln == 0 {
		return 0
	}
	off := protoencoding.WriteRequestVerificationHeaderTagAndLength(buf, ln)
	off += WriteMultiSignatureRequestVerificationHeader(buf[off:], pubKey, scheme, bodySig, metaHdrSig, originSig)
	return off
}

// WriteMultiSignatureRequestVerificationHeader writes request verification
// header message with same public key and scheme, and given signatures into
// buf. Returns number of bytes written.
func WriteMultiSignatureRequestVerificationHeader[SCHEME ~int32](buf []byte, pubKey []byte, scheme SCHEME, bodySig []byte, metaHdrSig []byte, originSig []byte) int {
	var originPubKey []byte
	var originScheme SCHEME
	if originSig != nil {
		originPubKey, originScheme = pubKey, scheme
	}
	return writeRequestVerificationHeader(buf, pubKey, bodySig, scheme, pubKey, metaHdrSig, scheme, originPubKey, originSig, originScheme, 0, nil, nil, nil, 0)
}

// WriteSingleSignatureRequestVerificationHeaderToRequest writes verification
// header field with single request signature into buf. Returns number of bytes
// written.
func WriteSingleSignatureRequestVerificationHeaderToRequest[SCHEME ~int32](buf []byte, pubKey []byte, scheme SCHEME, reqSig []byte) int {
	ln := CalculateSingleSignatureRequestVerificationHeaderLength(pubKey, scheme, reqSig)
	if ln == 0 {
		return 0
	}
	off := protoencoding.WriteRequestVerificationHeaderTagAndLength(buf, ln)
	off += WriteSingleSignatureRequestVerificationHeader(buf[off:], pubKey, scheme, reqSig)
	return off
}

// WriteSingleSignatureRequestVerificationHeader writes request verification
// header message with single request signature into buf. Returns number of
// bytes written.
func WriteSingleSignatureRequestVerificationHeader[SCHEME ~int32](buf []byte, pubKey []byte, scheme SCHEME, reqSig []byte) int {
	return writeRequestVerificationHeader(buf, nil, nil, 0, nil, nil, 0, nil, nil, 0, 0, nil, pubKey, reqSig, scheme)
}

func writeRequestVerificationHeader[SCHEME ~int32](buf []byte, bodyPubKey []byte, bodySig []byte, bodyScheme SCHEME, metaPubKey []byte, metaSig []byte, metaScheme SCHEME, originPubKey []byte, originSig []byte, originScheme SCHEME, originLen int, writeOriginFn protoencoding.WriteMessageFunc, reqPubKey []byte, reqSig []byte, reqScheme SCHEME) int {
	off := refs.WriteSignatureField(buf, FieldRequestVerificationHeaderBodySignature, bodyPubKey, bodySig, bodyScheme)
	off += refs.WriteSignatureField(buf[off:], FieldRequestVerificationHeaderMetaSignature, metaPubKey, metaSig, metaScheme)
	off += refs.WriteSignatureField(buf[off:], FieldRequestVerificationHeaderOriginSignature, originPubKey, originSig, originScheme)
	off += protoencoding.WriteMessageField(buf[off:], FieldRequestVerificationHeaderOrigin, originLen, writeOriginFn)
	off += refs.WriteSignatureField(buf[off:], FieldRequestVerificationHeaderRequestSignature, reqPubKey, reqSig, reqScheme)
	return off
}

// MarshalStable writes the RequestVerificationHeader in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [RequestVerificationHeader.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *RequestVerificationHeader) MarshalStable(b []byte) {
	if x == nil {
		return
	}
	writeRequestVerificationHeader(b, x.BodySignature.GetKey(), x.BodySignature.GetSign(), x.BodySignature.GetScheme(), x.MetaSignature.GetKey(), x.MetaSignature.GetSign(), x.MetaSignature.GetScheme(), x.OriginSignature.GetKey(), x.OriginSignature.GetSign(), x.OriginSignature.GetScheme(), x.Origin.MarshaledSize(), protoencoding.WriteStablyMarshalledMessageFunc(x.Origin), x.RequestSignature.GetKey(), x.RequestSignature.GetSign(), x.RequestSignature.GetScheme())
}

// Field numbers of [ResponseVerificationHeader] message.
const (
	_ = iota
	FieldResponseVerificationHeaderBodySignature
	FieldResponseVerificationHeaderMetaSignature
	FieldResponseVerificationHeaderOriginSignature
	FieldResponseVerificationHeaderOrigin
)

// MarshaledSize returns size of the ResponseVerificationHeader in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *ResponseVerificationHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldResponseVerificationHeaderBodySignature, x.BodySignature) +
			protoencoding.SizeEmbedded(FieldResponseVerificationHeaderMetaSignature, x.MetaSignature) +
			protoencoding.SizeEmbedded(FieldResponseVerificationHeaderOriginSignature, x.OriginSignature) +
			protoencoding.SizeEmbedded(FieldResponseVerificationHeaderOrigin, x.Origin)
	}
	return sz
}

// MarshalStable writes the ResponseVerificationHeader in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ResponseVerificationHeader.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ResponseVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldResponseVerificationHeaderBodySignature, x.BodySignature)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldResponseVerificationHeaderMetaSignature, x.MetaSignature)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldResponseVerificationHeaderOriginSignature, x.OriginSignature)
		protoencoding.MarshalToEmbedded(b[off:], FieldResponseVerificationHeaderOrigin, x.Origin)
	}
}
