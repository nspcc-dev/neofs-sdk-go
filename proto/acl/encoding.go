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

// MarshaledSize returns size of the EACLTable in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *EACLTable) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldEACLVersion, x.Version) +
			proto.SizeEmbedded(fieldEACLContainer, x.ContainerId) +
			proto.SizeRepeatedMessages(fieldEACLRecords, x.Records)
	}
	return 0
}

// MarshalStable writes the EACLTable in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLTable.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *EACLTable) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldEACLVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], fieldEACLContainer, x.ContainerId)
		proto.MarshalToRepeatedMessages(b[off:], fieldEACLRecords, x.Records)
	}
}

const (
	_ = iota
	fieldEACLOp
	fieldEACLAction
	fieldEACLFilters
	fieldEACLTargets
)

// MarshaledSize returns size of the EACLRecord in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *EACLRecord) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(fieldEACLOp, int32(x.Operation)) +
			proto.SizeVarint(fieldEACLAction, int32(x.Action)) +
			proto.SizeRepeatedMessages(fieldEACLFilters, x.Filters) +
			proto.SizeRepeatedMessages(fieldEACLTargets, x.Targets)
	}
	return 0
}

// MarshalStable writes the EACLRecord in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *EACLRecord) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldEACLOp, int32(x.Operation))
		off += proto.MarshalToVarint(b[off:], fieldEACLAction, int32(x.Action))
		off += proto.MarshalToRepeatedMessages(b[off:], fieldEACLFilters, x.Filters)
		proto.MarshalToRepeatedMessages(b[off:], fieldEACLTargets, x.Targets)
	}
}

const (
	_ = iota
	fieldEACLHeader
	fieldEACLMatcher
	fieldEACLKey
	fieldEACLValue
)

// MarshaledSize returns size of the EACLRecord_Filter in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
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

// MarshalStable writes the EACLRecord_Filter in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord_Filter.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *EACLRecord_Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldEACLHeader, int32(x.HeaderType))
		off += proto.MarshalToVarint(b[off:], fieldEACLMatcher, int32(x.MatchType))
		off += proto.MarshalToBytes(b[off:], fieldEACLKey, x.Key)
		proto.MarshalToBytes(b[off:], fieldEACLValue, x.Value)
	}
}

const (
	_ = iota
	fieldEACLRole
	fieldEACLTargetKeys
)

// MarshaledSize returns size of the EACLRecord_Target in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *EACLRecord_Target) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldEACLRole, int32(x.Role)) +
			proto.SizeRepeatedBytes(fieldEACLTargetKeys, x.Keys)
	}
	return sz
}

// MarshalStable writes the EACLRecord_Target in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord_Target.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *EACLRecord_Target) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldEACLRole, int32(x.Role))
		proto.MarshalToRepeatedBytes(b[off:], fieldEACLTargetKeys, x.Keys)
	}
}

const (
	_ = iota
	fieldBearerExp
	fieldBearerNbf
	fieldBearerIat
)

// MarshaledSize returns size of the BearerToken_Body_TokenLifetime in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *BearerToken_Body_TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldBearerExp, x.Exp) +
			proto.SizeVarint(fieldBearerNbf, x.Nbf) +
			proto.SizeVarint(fieldBearerIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the BearerToken_Body_TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [BearerToken_Body_TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *BearerToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldBearerExp, x.Exp)
		off += proto.MarshalToVarint(b[off:], fieldBearerNbf, x.Nbf)
		proto.MarshalToVarint(b[off:], fieldBearerIat, x.Iat)
	}
}

const (
	_ = iota
	fieldBearerEACL
	fieldBearerOwner
	fieldBearerLifetime
	fieldBearerIssuer
)

// MarshaledSize returns size of the BearerToken_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *BearerToken_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldBearerEACL, x.EaclTable) +
			proto.SizeEmbedded(fieldBearerOwner, x.OwnerId) +
			proto.SizeEmbedded(fieldBearerLifetime, x.Lifetime) +
			proto.SizeEmbedded(fieldBearerIssuer, x.Issuer)
	}
	return sz
}

// MarshalStable writes the BearerToken_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [BearerToken_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *BearerToken_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldBearerEACL, x.EaclTable)
		off += proto.MarshalToEmbedded(b[off:], fieldBearerOwner, x.OwnerId)
		off += proto.MarshalToEmbedded(b[off:], fieldBearerLifetime, x.Lifetime)
		proto.MarshalToEmbedded(b[off:], fieldBearerIssuer, x.Issuer)
	}
}

const (
	_ = iota
	fieldBearerBody
	fieldBearerSignature
)

// MarshaledSize returns size of the BearerToken in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *BearerToken) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldBearerBody, x.Body) +
			proto.SizeEmbedded(fieldBearerSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the BearerToken in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [BearerToken.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *BearerToken) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldBearerBody, x.Body)
		proto.MarshalToEmbedded(b[off:], fieldBearerSignature, x.Signature)
	}
}
