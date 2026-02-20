package acl

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Field numbers of [EACLTable] message.
const (
	_ = iota
	FieldEACLTableVersion
	FieldEACLTableContainerID
	FieldEACLTableRecords
)

// MarshaledSize returns size of the EACLTable in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *EACLTable) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldEACLTableVersion, x.Version) +
			proto.SizeEmbedded(FieldEACLTableContainerID, x.ContainerId) +
			proto.SizeRepeatedMessages(FieldEACLTableRecords, x.Records)
	}
	return 0
}

// MarshalStable writes the EACLTable in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLTable.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *EACLTable) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldEACLTableVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], FieldEACLTableContainerID, x.ContainerId)
		proto.MarshalToRepeatedMessages(b[off:], FieldEACLTableRecords, x.Records)
	}
}

// Field numbers of [EACLRecord] message.
const (
	_ = iota
	FieldEACLRecordOperation
	FieldEACLRecordAction
	FieldEACLRecordFilters
	FieldEACLRecordTargets
)

// MarshaledSize returns size of the EACLRecord in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *EACLRecord) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldEACLRecordOperation, int32(x.Operation)) +
			proto.SizeVarint(FieldEACLRecordAction, int32(x.Action)) +
			proto.SizeRepeatedMessages(FieldEACLRecordFilters, x.Filters) +
			proto.SizeRepeatedMessages(FieldEACLRecordTargets, x.Targets)
	}
	return 0
}

// MarshalStable writes the EACLRecord in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *EACLRecord) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldEACLRecordOperation, int32(x.Operation))
		off += proto.MarshalToVarint(b[off:], FieldEACLRecordAction, int32(x.Action))
		off += proto.MarshalToRepeatedMessages(b[off:], FieldEACLRecordFilters, x.Filters)
		proto.MarshalToRepeatedMessages(b[off:], FieldEACLRecordTargets, x.Targets)
	}
}

// Field numbers of [EACLRecord_Filter] message.
const (
	_ = iota
	FieldEACLRecordFilterHeaderType
	FieldEACLRecordFilterMatchType
	FieldEACLRecordFilterKey
	FieldEACLRecordFilterValue
)

// MarshaledSize returns size of the EACLRecord_Filter in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *EACLRecord_Filter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldEACLRecordFilterHeaderType, int32(x.HeaderType)) +
			proto.SizeVarint(FieldEACLRecordFilterMatchType, int32(x.MatchType)) +
			proto.SizeBytes(FieldEACLRecordFilterKey, x.Key) +
			proto.SizeBytes(FieldEACLRecordFilterValue, x.Value)
	}
	return sz
}

// MarshalStable writes the EACLRecord_Filter in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord_Filter.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *EACLRecord_Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldEACLRecordFilterHeaderType, int32(x.HeaderType))
		off += proto.MarshalToVarint(b[off:], FieldEACLRecordFilterMatchType, int32(x.MatchType))
		off += proto.MarshalToBytes(b[off:], FieldEACLRecordFilterKey, x.Key)
		proto.MarshalToBytes(b[off:], FieldEACLRecordFilterValue, x.Value)
	}
}

// Field numbers of [EACLRecord_Target] message.
const (
	_ = iota
	FieldEACLRecordTargetRole
	FieldEACLRecordTargetKeys
)

// MarshaledSize returns size of the EACLRecord_Target in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *EACLRecord_Target) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldEACLRecordTargetRole, int32(x.Role)) +
			proto.SizeRepeatedBytes(FieldEACLRecordTargetKeys, x.Keys)
	}
	return sz
}

// MarshalStable writes the EACLRecord_Target in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord_Target.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *EACLRecord_Target) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldEACLRecordTargetRole, int32(x.Role))
		proto.MarshalToRepeatedBytes(b[off:], FieldEACLRecordTargetKeys, x.Keys)
	}
}

// Field numbers of [BearerToken_Body_TokenLifetime] message.
const (
	_ = iota
	FieldBearerTokenBodyTokenLifetimeExp
	FieldBearerTokenBodyTokenLifetimeNbf
	FieldBearerTokenBodyTokenLifetimeIat
)

// MarshaledSize returns size of the BearerToken_Body_TokenLifetime in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *BearerToken_Body_TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldBearerTokenBodyTokenLifetimeExp, x.Exp) +
			proto.SizeVarint(FieldBearerTokenBodyTokenLifetimeNbf, x.Nbf) +
			proto.SizeVarint(FieldBearerTokenBodyTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the BearerToken_Body_TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [BearerToken_Body_TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *BearerToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldBearerTokenBodyTokenLifetimeExp, x.Exp)
		off += proto.MarshalToVarint(b[off:], FieldBearerTokenBodyTokenLifetimeNbf, x.Nbf)
		proto.MarshalToVarint(b[off:], FieldBearerTokenBodyTokenLifetimeIat, x.Iat)
	}
}

// Field numbers of [BearerToken_Body] message.
const (
	_ = iota
	FieldBearerTokenBodyEACLTable
	FieldBearerTokenBodyOwnerID
	FieldBearerTokenBodyLifetime
	FieldBearerTokenBodyIssuer
)

// MarshaledSize returns size of the BearerToken_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *BearerToken_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldBearerTokenBodyEACLTable, x.EaclTable) +
			proto.SizeEmbedded(FieldBearerTokenBodyOwnerID, x.OwnerId) +
			proto.SizeEmbedded(FieldBearerTokenBodyLifetime, x.Lifetime) +
			proto.SizeEmbedded(FieldBearerTokenBodyIssuer, x.Issuer)
	}
	return sz
}

// MarshalStable writes the BearerToken_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [BearerToken_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *BearerToken_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldBearerTokenBodyEACLTable, x.EaclTable)
		off += proto.MarshalToEmbedded(b[off:], FieldBearerTokenBodyOwnerID, x.OwnerId)
		off += proto.MarshalToEmbedded(b[off:], FieldBearerTokenBodyLifetime, x.Lifetime)
		proto.MarshalToEmbedded(b[off:], FieldBearerTokenBodyIssuer, x.Issuer)
	}
}

// Field numbers of [BearerToken] message.
const (
	_ = iota
	FieldBearerTokenBody
	FieldBearerTokenSignature
)

// MarshaledSize returns size of the BearerToken in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *BearerToken) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldBearerTokenBody, x.Body) +
			proto.SizeEmbedded(FieldBearerTokenSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the BearerToken in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [BearerToken.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *BearerToken) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldBearerTokenBody, x.Body)
		proto.MarshalToEmbedded(b[off:], FieldBearerTokenSignature, x.Signature)
	}
}
