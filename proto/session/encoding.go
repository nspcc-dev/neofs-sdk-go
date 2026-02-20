package session

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
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
		return proto.SizeEmbedded(FieldObjectSessionContextTargetContainer, x.Container) +
			proto.SizeRepeatedMessages(FieldObjectSessionContextTargetObjects, x.Objects)
	}
	return 0
}

// MarshalStable writes the ObjectSessionContext_Target in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ObjectSessionContext_Target.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ObjectSessionContext_Target) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldObjectSessionContextTargetContainer, x.Container)
		proto.MarshalToRepeatedMessages(b[off:], FieldObjectSessionContextTargetObjects, x.Objects)
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
		sz = proto.SizeVarint(FieldObjectSessionContextVerb, int32(x.Verb)) +
			proto.SizeEmbedded(FieldObjectSessionContextTarget, x.Target)
	}
	return sz
}

// MarshalStable writes the ObjectSessionContext in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [ObjectSessionContext.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ObjectSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldObjectSessionContextVerb, int32(x.Verb))
		proto.MarshalToEmbedded(b[off:], FieldObjectSessionContextTarget, x.Target)
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
		sz = proto.SizeVarint(FieldContainerSessionContextVerb, int32(x.Verb)) +
			proto.SizeBool(FieldContainerSessionContextWildcard, x.Wildcard) +
			proto.SizeEmbedded(FieldContainerSessionContextContainerID, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the ContainerSessionContext in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ContainerSessionContext.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ContainerSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldContainerSessionContextVerb, int32(x.Verb))
		off += proto.MarshalToBool(b[off:], FieldContainerSessionContextWildcard, x.Wildcard)
		proto.MarshalToEmbedded(b[off:], FieldContainerSessionContextContainerID, x.ContainerId)
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
		sz = proto.SizeVarint(FieldSessionTokenBodyTokenLifetimeExp, x.Exp) +
			proto.SizeVarint(FieldSessionTokenBodyTokenLifetimeNbf, x.Nbf) +
			proto.SizeVarint(FieldSessionTokenBodyTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the SessionToken_Body_TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [SessionToken_Body_TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SessionToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldSessionTokenBodyTokenLifetimeExp, x.Exp)
		off += proto.MarshalToVarint(b[off:], FieldSessionTokenBodyTokenLifetimeNbf, x.Nbf)
		proto.MarshalToVarint(b[off:], FieldSessionTokenBodyTokenLifetimeIat, x.Iat)
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
		sz = proto.SizeBytes(FieldSessionTokenBodyID, x.Id) +
			proto.SizeEmbedded(FieldSessionTokenBodyOwnerID, x.OwnerId) +
			proto.SizeEmbedded(FieldSessionTokenBodyLifetime, x.Lifetime) +
			proto.SizeBytes(FieldSessionTokenBodySessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			sz += proto.SizeEmbedded(FieldSessionTokenBodyObject, c.Object)
		case *SessionToken_Body_Container:
			sz += proto.SizeEmbedded(FieldSessionTokenBodyContainer, c.Container)
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
		off := proto.MarshalToBytes(b, FieldSessionTokenBodyID, x.Id)
		off += proto.MarshalToEmbedded(b[off:], FieldSessionTokenBodyOwnerID, x.OwnerId)
		off += proto.MarshalToEmbedded(b[off:], FieldSessionTokenBodyLifetime, x.Lifetime)
		off += proto.MarshalToBytes(b[off:], FieldSessionTokenBodySessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			proto.MarshalToEmbedded(b[off:], FieldSessionTokenBodyObject, c.Object)
		case *SessionToken_Body_Container:
			proto.MarshalToEmbedded(b[off:], FieldSessionTokenBodyContainer, c.Container)
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
		sz = proto.SizeEmbedded(FieldSessionTokenBody, x.Body) +
			proto.SizeEmbedded(FieldSessionTokenSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the SessionToken in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionToken.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SessionToken) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSessionTokenBody, x.Body)
		proto.MarshalToEmbedded(b[off:], FieldSessionTokenSignature, x.Signature)
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
		sz = proto.SizeEmbedded(FieldCreateRequestBodyOwnerID, x.OwnerId) +
			proto.SizeVarint(FieldCreateRequestBodyExpiration, x.Expiration)
	}
	return sz
}

// MarshalStable writes the CreateRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [CreateRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *CreateRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldCreateRequestBodyOwnerID, x.OwnerId)
		proto.MarshalToVarint(b[off:], FieldCreateRequestBodyExpiration, x.Expiration)
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
		sz = proto.SizeBytes(FieldCreateResponseBodyID, x.Id) +
			proto.SizeBytes(FieldCreateResponseBodySessionKey, x.SessionKey)
	}
	return sz
}

// MarshalStable writes the CreateResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [CreateResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *CreateResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldCreateResponseBodyID, x.Id)
		proto.MarshalToBytes(b[off:], FieldCreateResponseBodySessionKey, x.SessionKey)
	}
}

// Field numbers of [XHeader] message.
const (
	_ = iota
	FieldXHeaderKey
	FieldXHeaderValue
)

// MarshaledSize returns size of the XHeader in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *XHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldXHeaderKey, x.Key) +
			proto.SizeBytes(FieldXHeaderValue, x.Value)
	}
	return sz
}

// MarshalStable writes the XHeader in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [XHeader.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *XHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldXHeaderKey, x.Key)
		proto.MarshalToBytes(b[off:], FieldXHeaderValue, x.Value)
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
)

// MarshaledSize returns size of the RequestMetaHeader in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *RequestMetaHeader) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldRequestMetaHeaderVersion, x.Version) +
			proto.SizeVarint(FieldRequestMetaHeaderEpoch, x.Epoch) +
			proto.SizeVarint(FieldRequestMetaHeaderTTL, x.Ttl) +
			proto.SizeEmbedded(FieldRequestMetaHeaderSessionToken, x.SessionToken) +
			proto.SizeEmbedded(FieldRequestMetaHeaderBearerToken, x.BearerToken) +
			proto.SizeEmbedded(FieldRequestMetaHeaderOrigin, x.Origin) +
			proto.SizeVarint(FieldRequestMetaHeaderMagicNumber, x.MagicNumber) +
			proto.SizeRepeatedMessages(FieldRequestMetaHeaderXHeaders, x.XHeaders)
	}
	return 0
}

// MarshalStable writes the RequestMetaHeader in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [RequestMetaHeader.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *RequestMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldRequestMetaHeaderVersion, x.Version)
		off += proto.MarshalToVarint(b[off:], FieldRequestMetaHeaderEpoch, x.Epoch)
		off += proto.MarshalToVarint(b[off:], FieldRequestMetaHeaderTTL, x.Ttl)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldRequestMetaHeaderXHeaders, x.XHeaders)
		off += proto.MarshalToEmbedded(b[off:], FieldRequestMetaHeaderSessionToken, x.SessionToken)
		off += proto.MarshalToEmbedded(b[off:], FieldRequestMetaHeaderBearerToken, x.BearerToken)
		off += proto.MarshalToEmbedded(b[off:], FieldRequestMetaHeaderOrigin, x.Origin)
		proto.MarshalToVarint(b[off:], FieldRequestMetaHeaderMagicNumber, x.MagicNumber)
	}
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
		return proto.SizeEmbedded(FieldResponseMetaHeaderVersion, x.Version) +
			proto.SizeVarint(FieldResponseMetaHeaderEpoch, x.Epoch) +
			proto.SizeVarint(FieldResponseMetaHeaderTTL, x.Ttl) +
			proto.SizeEmbedded(FieldResponseMetaHeaderOrigin, x.Origin) +
			proto.SizeEmbedded(FieldResponseMetaHeaderStatus, x.Status) +
			proto.SizeRepeatedMessages(FieldResponseMetaHeaderXHeaders, x.XHeaders)
	}
	return 0
}

// MarshalStable writes the ResponseMetaHeader in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [ResponseMetaHeader.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ResponseMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldResponseMetaHeaderVersion, x.Version)
		off += proto.MarshalToVarint(b[off:], FieldResponseMetaHeaderEpoch, x.Epoch)
		off += proto.MarshalToVarint(b[off:], FieldResponseMetaHeaderTTL, x.Ttl)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldResponseMetaHeaderXHeaders, x.XHeaders)
		off += proto.MarshalToEmbedded(b[off:], FieldResponseMetaHeaderOrigin, x.Origin)
		proto.MarshalToEmbedded(b[off:], FieldResponseMetaHeaderStatus, x.Status)
	}
}

// Field numbers of [RequestVerificationHeader] message.
const (
	_ = iota
	FieldRequestVerificationHeaderBodySignature
	FieldRequestVerificationHeaderMetaSignature
	FieldRequestVerificationHeaderOriginSignature
	FieldRequestVerificationHeaderOrigin
)

// MarshaledSize returns size of the RequestVerificationHeader in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *RequestVerificationHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldRequestVerificationHeaderBodySignature, x.BodySignature) +
			proto.SizeEmbedded(FieldRequestVerificationHeaderMetaSignature, x.MetaSignature) +
			proto.SizeEmbedded(FieldRequestVerificationHeaderOriginSignature, x.OriginSignature) +
			proto.SizeEmbedded(FieldRequestVerificationHeaderOrigin, x.Origin)
	}
	return sz
}

// MarshalStable writes the RequestVerificationHeader in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [RequestVerificationHeader.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *RequestVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldRequestVerificationHeaderBodySignature, x.BodySignature)
		off += proto.MarshalToEmbedded(b[off:], FieldRequestVerificationHeaderMetaSignature, x.MetaSignature)
		off += proto.MarshalToEmbedded(b[off:], FieldRequestVerificationHeaderOriginSignature, x.OriginSignature)
		proto.MarshalToEmbedded(b[off:], FieldRequestVerificationHeaderOrigin, x.Origin)
	}
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
		sz = proto.SizeEmbedded(FieldResponseVerificationHeaderBodySignature, x.BodySignature) +
			proto.SizeEmbedded(FieldResponseVerificationHeaderMetaSignature, x.MetaSignature) +
			proto.SizeEmbedded(FieldResponseVerificationHeaderOriginSignature, x.OriginSignature) +
			proto.SizeEmbedded(FieldResponseVerificationHeaderOrigin, x.Origin)
	}
	return sz
}

// MarshalStable writes the ResponseVerificationHeader in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ResponseVerificationHeader.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ResponseVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldResponseVerificationHeaderBodySignature, x.BodySignature)
		off += proto.MarshalToEmbedded(b[off:], FieldResponseVerificationHeaderMetaSignature, x.MetaSignature)
		off += proto.MarshalToEmbedded(b[off:], FieldResponseVerificationHeaderOriginSignature, x.OriginSignature)
		proto.MarshalToEmbedded(b[off:], FieldResponseVerificationHeaderOrigin, x.Origin)
	}
}
