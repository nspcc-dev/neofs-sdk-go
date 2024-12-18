package session

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldTokenObjectTargetContainer
	fieldTokenObjectTargetIDs
)

// MarshaledSize returns size of the ObjectSessionContext_Target in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *ObjectSessionContext_Target) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldTokenObjectTargetContainer, x.Container) +
			proto.SizeRepeatedMessages(fieldTokenObjectTargetIDs, x.Objects)
	}
	return 0
}

// MarshalStable writes the ObjectSessionContext_Target in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ObjectSessionContext_Target.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ObjectSessionContext_Target) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldTokenObjectTargetContainer, x.Container)
		proto.MarshalToRepeatedMessages(b[off:], fieldTokenObjectTargetIDs, x.Objects)
	}
}

const (
	_ = iota
	fieldTokenObjectVerb
	fieldTokenObjectTarget
)

// MarshaledSize returns size of the ObjectSessionContext in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ObjectSessionContext) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenObjectVerb, int32(x.Verb)) +
			proto.SizeEmbedded(fieldTokenObjectTarget, x.Target)
	}
	return sz
}

// MarshalStable writes the ObjectSessionContext in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [ObjectSessionContext.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ObjectSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTokenObjectVerb, int32(x.Verb))
		proto.MarshalToEmbedded(b[off:], fieldTokenObjectTarget, x.Target)
	}
}

const (
	_ = iota
	fieldTokenContainerVerb
	fieldTokenContainerWildcard
	fieldTokenContainerID
)

// MarshaledSize returns size of the ContainerSessionContext in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *ContainerSessionContext) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenContainerVerb, int32(x.Verb)) +
			proto.SizeBool(fieldTokenContainerWildcard, x.Wildcard) +
			proto.SizeEmbedded(fieldTokenContainerID, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the ContainerSessionContext in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ContainerSessionContext.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ContainerSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTokenContainerVerb, int32(x.Verb))
		off += proto.MarshalToBool(b[off:], fieldTokenContainerWildcard, x.Wildcard)
		proto.MarshalToEmbedded(b[off:], fieldTokenContainerID, x.ContainerId)
	}
}

const (
	_ = iota
	fieldTokenExp
	fieldTokenNbf
	fieldTokenIat
)

// MarshaledSize returns size of the SessionToken_Body_TokenLifetime in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SessionToken_Body_TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenExp, x.Exp) +
			proto.SizeVarint(fieldTokenNbf, x.Nbf) +
			proto.SizeVarint(fieldTokenIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the SessionToken_Body_TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [SessionToken_Body_TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SessionToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTokenExp, x.Exp)
		off += proto.MarshalToVarint(b[off:], fieldTokenNbf, x.Nbf)
		proto.MarshalToVarint(b[off:], fieldTokenIat, x.Iat)
	}
}

const (
	_ = iota
	fieldTokenID
	fieldTokenOwner
	fieldTokenLifetime
	fieldTokenSessionKey
	fieldTokenContextObject
	fieldTokenContextContainer
)

// MarshaledSize returns size of the SessionToken_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionToken_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldTokenID, x.Id) +
			proto.SizeEmbedded(fieldTokenOwner, x.OwnerId) +
			proto.SizeEmbedded(fieldTokenLifetime, x.Lifetime) +
			proto.SizeBytes(fieldTokenSessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			sz += proto.SizeEmbedded(fieldTokenContextObject, c.Object)
		case *SessionToken_Body_Container:
			sz += proto.SizeEmbedded(fieldTokenContextContainer, c.Container)
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
		off := proto.MarshalToBytes(b, fieldTokenID, x.Id)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenOwner, x.OwnerId)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenLifetime, x.Lifetime)
		off += proto.MarshalToBytes(b[off:], fieldTokenSessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			proto.MarshalToEmbedded(b[off:], fieldTokenContextObject, c.Object)
		case *SessionToken_Body_Container:
			proto.MarshalToEmbedded(b[off:], fieldTokenContextContainer, c.Container)
		}
	}
}

const (
	_ = iota
	fieldTokenBody
	fieldTokenSignature
)

// MarshaledSize returns size of the SessionToken in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *SessionToken) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldTokenBody, x.Body) +
			proto.SizeEmbedded(fieldTokenSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the SessionToken in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionToken.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SessionToken) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldTokenBody, x.Body)
		proto.MarshalToEmbedded(b[off:], fieldTokenSignature, x.Signature)
	}
}

const (
	_ = iota
	fieldCreateReqUser
	fieldCreateReqExp
)

// MarshaledSize returns size of the CreateRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *CreateRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldCreateReqUser, x.OwnerId) +
			proto.SizeVarint(fieldCreateReqExp, x.Expiration)
	}
	return sz
}

// MarshalStable writes the CreateRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [CreateRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *CreateRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldCreateReqUser, x.OwnerId)
		proto.MarshalToVarint(b[off:], fieldCreateReqExp, x.Expiration)
	}
}

const (
	_ = iota
	fieldCreateRespID
	fieldCreateRespSessionKey
)

// MarshaledSize returns size of the CreateResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *CreateResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldCreateRespID, x.Id) +
			proto.SizeBytes(fieldCreateRespSessionKey, x.SessionKey)
	}
	return sz
}

// MarshalStable writes the CreateResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [CreateResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *CreateResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldCreateRespID, x.Id)
		proto.MarshalToBytes(b[off:], fieldCreateRespSessionKey, x.SessionKey)
	}
}

const (
	_ = iota
	fieldXHeaderKey
	fieldXHeaderValue
)

// MarshaledSize returns size of the XHeader in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *XHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldXHeaderKey, x.Key) +
			proto.SizeBytes(fieldXHeaderValue, x.Value)
	}
	return sz
}

// MarshalStable writes the XHeader in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [XHeader.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *XHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldXHeaderKey, x.Key)
		proto.MarshalToBytes(b[off:], fieldXHeaderValue, x.Value)
	}
}

const (
	_ = iota
	fieldReqMetaVersion
	fieldReqMetaEpoch
	fieldReqMetaTTL
	fieldReqMetaXHeaders
	fieldReqMetaSession
	fieldReqMetaBearer
	fieldReqMetaOrigin
	fieldReqMetaNetMagic
)

// MarshaledSize returns size of the RequestMetaHeader in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *RequestMetaHeader) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldReqMetaVersion, x.Version) +
			proto.SizeVarint(fieldReqMetaEpoch, x.Epoch) +
			proto.SizeVarint(fieldReqMetaTTL, x.Ttl) +
			proto.SizeEmbedded(fieldReqMetaSession, x.SessionToken) +
			proto.SizeEmbedded(fieldReqMetaBearer, x.BearerToken) +
			proto.SizeEmbedded(fieldReqMetaOrigin, x.Origin) +
			proto.SizeVarint(fieldReqMetaNetMagic, x.MagicNumber) +
			proto.SizeRepeatedMessages(fieldReqMetaXHeaders, x.XHeaders)
	}
	return 0
}

// MarshalStable writes the RequestMetaHeader in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [RequestMetaHeader.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *RequestMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldReqMetaVersion, x.Version)
		off += proto.MarshalToVarint(b[off:], fieldReqMetaEpoch, x.Epoch)
		off += proto.MarshalToVarint(b[off:], fieldReqMetaTTL, x.Ttl)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldReqMetaXHeaders, x.XHeaders)
		off += proto.MarshalToEmbedded(b[off:], fieldReqMetaSession, x.SessionToken)
		off += proto.MarshalToEmbedded(b[off:], fieldReqMetaBearer, x.BearerToken)
		off += proto.MarshalToEmbedded(b[off:], fieldReqMetaOrigin, x.Origin)
		proto.MarshalToVarint(b[off:], fieldReqMetaNetMagic, x.MagicNumber)
	}
}

const (
	_ = iota
	fieldRespMetaVersion
	fieldRespMetaEpoch
	fieldRespMetaTTL
	fieldRespMetaXHeaders
	fieldRespMetaOrigin
	fieldRespMetaStatus
)

// MarshaledSize returns size of the ResponseMetaHeader in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ResponseMetaHeader) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldRespMetaVersion, x.Version) +
			proto.SizeVarint(fieldRespMetaEpoch, x.Epoch) +
			proto.SizeVarint(fieldRespMetaTTL, x.Ttl) +
			proto.SizeEmbedded(fieldRespMetaOrigin, x.Origin) +
			proto.SizeEmbedded(fieldRespMetaStatus, x.Status) +
			proto.SizeRepeatedMessages(fieldRespMetaXHeaders, x.XHeaders)
	}
	return 0
}

// MarshalStable writes the ResponseMetaHeader in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [ResponseMetaHeader.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ResponseMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldRespMetaVersion, x.Version)
		off += proto.MarshalToVarint(b[off:], fieldRespMetaEpoch, x.Epoch)
		off += proto.MarshalToVarint(b[off:], fieldRespMetaTTL, x.Ttl)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldRespMetaXHeaders, x.XHeaders)
		off += proto.MarshalToEmbedded(b[off:], fieldRespMetaOrigin, x.Origin)
		proto.MarshalToEmbedded(b[off:], fieldRespMetaStatus, x.Status)
	}
}

const (
	_ = iota
	fieldReqVerifyBodySignature
	fieldReqVerifyMetaSignature
	fieldReqVerifyOriginSignature
	fieldReqVerifyOrigin
)

// MarshaledSize returns size of the RequestVerificationHeader in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *RequestVerificationHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldReqVerifyBodySignature, x.BodySignature) +
			proto.SizeEmbedded(fieldReqVerifyMetaSignature, x.MetaSignature) +
			proto.SizeEmbedded(fieldReqVerifyOriginSignature, x.OriginSignature) +
			proto.SizeEmbedded(fieldReqVerifyOrigin, x.Origin)
	}
	return sz
}

// MarshalStable writes the RequestVerificationHeader in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [RequestVerificationHeader.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *RequestVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldReqVerifyBodySignature, x.BodySignature)
		off += proto.MarshalToEmbedded(b[off:], fieldReqVerifyMetaSignature, x.MetaSignature)
		off += proto.MarshalToEmbedded(b[off:], fieldReqVerifyOriginSignature, x.OriginSignature)
		proto.MarshalToEmbedded(b[off:], fieldReqVerifyOrigin, x.Origin)
	}
}

const (
	_ = iota
	fieldRespVerifyBodySignature
	fieldRespVerifyMetaSignature
	fieldRespVerifyOriginSignature
	fieldRespVerifyOrigin
)

// MarshaledSize returns size of the ResponseVerificationHeader in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *ResponseVerificationHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldRespVerifyBodySignature, x.BodySignature) +
			proto.SizeEmbedded(fieldRespVerifyMetaSignature, x.MetaSignature) +
			proto.SizeEmbedded(fieldRespVerifyOriginSignature, x.OriginSignature) +
			proto.SizeEmbedded(fieldRespVerifyOrigin, x.Origin)
	}
	return sz
}

// MarshalStable writes the ResponseVerificationHeader in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [ResponseVerificationHeader.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *ResponseVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldRespVerifyBodySignature, x.BodySignature)
		off += proto.MarshalToEmbedded(b[off:], fieldRespVerifyMetaSignature, x.MetaSignature)
		off += proto.MarshalToEmbedded(b[off:], fieldRespVerifyOriginSignature, x.OriginSignature)
		proto.MarshalToEmbedded(b[off:], fieldRespVerifyOrigin, x.Origin)
	}
}
