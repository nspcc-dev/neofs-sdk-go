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

func (x *ObjectSessionContext_Target) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldTokenObjectTargetContainer, x.Container)
		for i := range x.Objects {
			sz += proto.SizeNested(fieldTokenObjectTargetIDs, x.Objects[i])
		}
	}
	return sz
}

func (x *ObjectSessionContext_Target) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldTokenObjectTargetContainer, x.Container)
		for i := range x.Objects {
			off += proto.MarshalNested(b[off:], fieldTokenObjectTargetIDs, x.Objects[i])
		}
	}
}

const (
	_ = iota
	fieldTokenObjectVerb
	fieldTokenObjectTarget
)

func (x *ObjectSessionContext) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenObjectVerb, int32(x.Verb)) +
			proto.SizeNested(fieldTokenObjectTarget, x.Target)
	}
	return sz
}

func (x *ObjectSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldTokenObjectVerb, int32(x.Verb))
		proto.MarshalNested(b[off:], fieldTokenObjectTarget, x.Target)
	}
}

const (
	_ = iota
	fieldTokenContainerVerb
	fieldTokenContainerWildcard
	fieldTokenContainerID
)

func (x *ContainerSessionContext) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenContainerVerb, int32(x.Verb)) +
			proto.SizeBool(fieldTokenContainerWildcard, x.Wildcard) +
			proto.SizeNested(fieldTokenContainerID, x.ContainerId)
	}
	return sz
}

func (x *ContainerSessionContext) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldTokenContainerVerb, int32(x.Verb))
		off += proto.MarshalBool(b[off:], fieldTokenContainerWildcard, x.Wildcard)
		proto.MarshalNested(b[off:], fieldTokenContainerID, x.ContainerId)
	}
}

const (
	_ = iota
	fieldTokenExp
	fieldTokenNbf
	fieldTokenIat
)

func (x *SessionToken_Body_TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenExp, x.Exp) +
			proto.SizeVarint(fieldTokenNbf, x.Nbf) +
			proto.SizeVarint(fieldTokenIat, x.Iat)

	}
	return sz
}

func (x *SessionToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldTokenExp, x.Exp)
		off += proto.MarshalVarint(b[off:], fieldTokenNbf, x.Nbf)
		proto.MarshalVarint(b[off:], fieldTokenIat, x.Iat)
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

func (x *SessionToken_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldTokenID, x.Id) +
			proto.SizeNested(fieldTokenOwner, x.OwnerId) +
			proto.SizeNested(fieldTokenLifetime, x.Lifetime) +
			proto.SizeBytes(fieldTokenSessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			sz += proto.SizeNested(fieldTokenContextObject, c.Object)
		case *SessionToken_Body_Container:
			sz += proto.SizeNested(fieldTokenContextContainer, c.Container)
		}
	}
	return sz
}

func (x *SessionToken_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldTokenID, x.Id)
		off += proto.MarshalNested(b[off:], fieldTokenOwner, x.OwnerId)
		off += proto.MarshalNested(b[off:], fieldTokenLifetime, x.Lifetime)
		off += proto.MarshalBytes(b[off:], fieldTokenSessionKey, x.SessionKey)
		switch c := x.Context.(type) {
		default:
			panic(fmt.Sprintf("unexpected context %T", x.Context))
		case nil:
		case *SessionToken_Body_Object:
			proto.MarshalNested(b[off:], fieldTokenContextObject, c.Object)
		case *SessionToken_Body_Container:
			proto.MarshalNested(b[off:], fieldTokenContextContainer, c.Container)
		}
	}
}

const (
	_ = iota
	fieldTokenBody
	fieldTokenSignature
)

func (x *SessionToken) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldTokenBody, x.Body) +
			proto.SizeNested(fieldTokenSignature, x.Signature)
	}
	return sz
}

func (x *SessionToken) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldTokenBody, x.Body)
		proto.MarshalNested(b[off:], fieldTokenSignature, x.Signature)
	}
}

const (
	_ = iota
	fieldCreateReqUser
	fieldCreateReqExp
)

func (x *CreateRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldCreateReqUser, x.OwnerId) +
			proto.SizeVarint(fieldCreateReqExp, x.Expiration)
	}
	return sz
}

func (x *CreateRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldCreateReqUser, x.OwnerId)
		proto.MarshalVarint(b[off:], fieldCreateReqExp, x.Expiration)
	}
}

const (
	_ = iota
	fieldCreateRespID
	fieldCreateRespSessionKey
)

func (x *CreateResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldCreateRespID, x.Id) +
			proto.SizeBytes(fieldCreateRespSessionKey, x.SessionKey)
	}
	return sz
}

func (x *CreateResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldCreateRespID, x.Id)
		proto.MarshalBytes(b[off:], fieldCreateRespSessionKey, x.SessionKey)
	}
}

const (
	_ = iota
	fieldXHeaderKey
	fieldXHeaderValue
)

func (x *XHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldXHeaderKey, x.Key) +
			proto.SizeBytes(fieldXHeaderValue, x.Value)
	}
	return sz
}

func (x *XHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldXHeaderKey, x.Key)
		proto.MarshalBytes(b[off:], fieldXHeaderValue, x.Value)
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

func (x *RequestMetaHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldReqMetaVersion, x.Version) +
			proto.SizeVarint(fieldReqMetaEpoch, x.Epoch) +
			proto.SizeVarint(fieldReqMetaTTL, x.Ttl) +
			proto.SizeNested(fieldReqMetaSession, x.SessionToken) +
			proto.SizeNested(fieldReqMetaBearer, x.BearerToken) +
			proto.SizeNested(fieldReqMetaOrigin, x.Origin) +
			proto.SizeVarint(fieldReqMetaNetMagic, x.MagicNumber)
		for i := range x.XHeaders {
			sz += proto.SizeNested(fieldReqMetaXHeaders, x.XHeaders[i])
		}
	}
	return sz
}

func (x *RequestMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldReqMetaVersion, x.Version)
		off += proto.MarshalVarint(b[off:], fieldReqMetaEpoch, x.Epoch)
		off += proto.MarshalVarint(b[off:], fieldReqMetaTTL, x.Ttl)
		for i := range x.XHeaders {
			off += proto.MarshalNested(b[off:], fieldReqMetaXHeaders, x.XHeaders[i])
		}
		off += proto.MarshalNested(b[off:], fieldReqMetaSession, x.SessionToken)
		off += proto.MarshalNested(b[off:], fieldReqMetaBearer, x.BearerToken)
		off += proto.MarshalNested(b[off:], fieldReqMetaOrigin, x.Origin)
		off += proto.MarshalVarint(b[off:], fieldReqMetaNetMagic, x.MagicNumber)
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

func (x *ResponseMetaHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldRespMetaVersion, x.Version) +
			proto.SizeVarint(fieldRespMetaEpoch, x.Epoch) +
			proto.SizeVarint(fieldRespMetaTTL, x.Ttl) +
			proto.SizeNested(fieldRespMetaOrigin, x.Origin) +
			proto.SizeNested(fieldRespMetaStatus, x.Status)
		for i := range x.XHeaders {
			sz += proto.SizeNested(fieldRespMetaXHeaders, x.XHeaders[i])
		}
	}
	return sz
}

func (x *ResponseMetaHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldRespMetaVersion, x.Version)
		off += proto.MarshalVarint(b[off:], fieldRespMetaEpoch, x.Epoch)
		off += proto.MarshalVarint(b[off:], fieldRespMetaTTL, x.Ttl)
		for i := range x.XHeaders {
			off += proto.MarshalNested(b[off:], fieldRespMetaXHeaders, x.XHeaders[i])
		}
		off += proto.MarshalNested(b[off:], fieldRespMetaOrigin, x.Origin)
		off += proto.MarshalNested(b[off:], fieldRespMetaStatus, x.Status)
	}
}

const (
	_ = iota
	fieldReqVerifyBodySignature
	fieldReqVerifyMetaSignature
	fieldReqVerifyOriginSignature
	fieldReqVerifyOrigin
)

func (x *RequestVerificationHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldReqVerifyBodySignature, x.BodySignature) +
			proto.SizeNested(fieldReqVerifyMetaSignature, x.MetaSignature) +
			proto.SizeNested(fieldReqVerifyOriginSignature, x.OriginSignature) +
			proto.SizeNested(fieldReqVerifyOrigin, x.Origin)
	}
	return sz
}

func (x *RequestVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldReqVerifyBodySignature, x.BodySignature)
		off += proto.MarshalNested(b[off:], fieldReqVerifyMetaSignature, x.MetaSignature)
		off += proto.MarshalNested(b[off:], fieldReqVerifyOriginSignature, x.OriginSignature)
		proto.MarshalNested(b[off:], fieldReqVerifyOrigin, x.Origin)
	}
}

const (
	_ = iota
	fieldRespVerifyBodySignature
	fieldRespVerifyMetaSignature
	fieldRespVerifyOriginSignature
	fieldRespVerifyOrigin
)

func (x *ResponseVerificationHeader) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldRespVerifyBodySignature, x.BodySignature) +
			proto.SizeNested(fieldRespVerifyMetaSignature, x.MetaSignature) +
			proto.SizeNested(fieldRespVerifyOriginSignature, x.OriginSignature) +
			proto.SizeNested(fieldRespVerifyOrigin, x.Origin)
	}
	return sz
}

func (x *ResponseVerificationHeader) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldRespVerifyBodySignature, x.BodySignature)
		off += proto.MarshalNested(b[off:], fieldRespVerifyMetaSignature, x.MetaSignature)
		off += proto.MarshalNested(b[off:], fieldRespVerifyOriginSignature, x.OriginSignature)
		proto.MarshalNested(b[off:], fieldRespVerifyOrigin, x.Origin)
	}
}
