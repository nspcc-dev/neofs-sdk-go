package acl

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldEACLVersion
	fieldEACLContainer
	fieldEACLRecords
)

func (x *EACLTable) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldEACLVersion, x.Version) +
			proto.SizeNested(fieldEACLContainer, x.ContainerId)
		for i := range x.Records {
			sz += proto.SizeNested(fieldEACLRecords, x.Records[i])
		}
	}
	return sz
}

func (x *EACLTable) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldEACLVersion, x.Version)
		off += proto.MarshalNested(b[off:], fieldEACLContainer, x.ContainerId)
		for i := range x.Records {
			off += proto.MarshalNested(b[off:], fieldEACLRecords, x.Records[i])
		}
	}
}

const (
	_ = iota
	fieldEACLOp
	fieldEACLAction
	fieldEACLFilters
	fieldEACLTargets
)

func (x *EACLRecord) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldEACLOp, int32(x.Operation)) +
			proto.SizeVarint(fieldEACLAction, int32(x.Action))
		for i := range x.Filters {
			sz += proto.SizeNested(fieldEACLFilters, x.Filters[i])
		}
		for i := range x.Targets {
			sz += proto.SizeNested(fieldEACLTargets, x.Targets[i])
		}
	}
	return sz
}

func (x *EACLRecord) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldEACLOp, int32(x.Operation))
		off += proto.MarshalVarint(b[off:], fieldEACLAction, int32(x.Action))
		for i := range x.Filters {
			off += proto.MarshalNested(b[off:], fieldEACLFilters, x.Filters[i])
		}
		for i := range x.Targets {
			off += proto.MarshalNested(b[off:], fieldEACLTargets, x.Targets[i])
		}
	}
}

const (
	_ = iota
	fieldEACLHeader
	fieldEACLMatcher
	fieldEACLKey
	fieldEACLValue
)

func (x *EACLRecord_Filter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldEACLHeader, int32(x.HeaderType)) +
			proto.SizeVarint(fieldEACLMatcher, int32(x.MatchType)) +
			proto.SizeBytes(fieldEACLKey, x.Key) +
			proto.SizeBytes(fieldEACLValue, x.Value)
	}
	return sz
}

func (x *EACLRecord_Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldEACLHeader, int32(x.HeaderType))
		off += proto.MarshalVarint(b[off:], fieldEACLMatcher, int32(x.MatchType))
		off += proto.MarshalBytes(b[off:], fieldEACLKey, x.Key)
		proto.MarshalBytes(b[off:], fieldEACLValue, x.Value)
	}
}

const (
	_ = iota
	fieldEACLRole
	fieldEACLTargetKeys
)

func (x *EACLRecord_Target) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldEACLRole, int32(x.Role)) +
			proto.SizeRepeatedBytes(fieldEACLTargetKeys, x.Keys)
	}
	return sz
}

func (x *EACLRecord_Target) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldEACLRole, int32(x.Role))
		proto.MarshalRepeatedBytes(b[off:], fieldEACLTargetKeys, x.Keys)
	}
}

const (
	_ = iota
	fieldBearerExp
	fieldBearerNbf
	fieldBearerIat
)

func (x *BearerToken_Body_TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldBearerExp, x.Exp) +
			proto.SizeVarint(fieldBearerNbf, x.Nbf) +
			proto.SizeVarint(fieldBearerIat, x.Iat)
	}
	return sz
}

func (x *BearerToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldBearerExp, x.Exp)
		off += proto.MarshalVarint(b[off:], fieldBearerNbf, x.Nbf)
		proto.MarshalVarint(b[off:], fieldBearerIat, x.Iat)
	}
}

const (
	_ = iota
	fieldBearerEACL
	fieldBearerOwner
	fieldBearerLifetime
	fieldBearerIssuer
)

func (x *BearerToken_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldBearerEACL, x.EaclTable) +
			proto.SizeNested(fieldBearerOwner, x.OwnerId) +
			proto.SizeNested(fieldBearerLifetime, x.Lifetime) +
			proto.SizeNested(fieldBearerIssuer, x.Issuer)
	}
	return sz
}

func (x *BearerToken_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldBearerEACL, x.EaclTable)
		off += proto.MarshalNested(b[off:], fieldBearerOwner, x.OwnerId)
		off += proto.MarshalNested(b[off:], fieldBearerLifetime, x.Lifetime)
		proto.MarshalNested(b[off:], fieldBearerIssuer, x.Issuer)
	}
}

const (
	_ = iota
	fieldBearerBody
	fieldBearerSignature
)

func (x *BearerToken) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldBearerBody, x.Body) +
			proto.SizeNested(fieldBearerSignature, x.Signature)
	}
	return sz
}

func (x *BearerToken) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldBearerBody, x.Body)
		proto.MarshalNested(b[off:], fieldBearerSignature, x.Signature)
	}
}
