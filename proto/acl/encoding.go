package acl

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
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
		return protoencoding.SizeEmbedded(FieldEACLTableVersion, x.Version) +
			protoencoding.SizeEmbedded(FieldEACLTableContainerID, x.ContainerId) +
			protoencoding.SizeRepeatedMessages(FieldEACLTableRecords, x.Records)
	}
	return 0
}

// MarshalStable writes the EACLTable in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLTable.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *EACLTable) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldEACLTableVersion, x.Version)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldEACLTableContainerID, x.ContainerId)
		protoencoding.MarshalToRepeatedMessages(b[off:], FieldEACLTableRecords, x.Records)
	}
}

// Field numbers of [EACLRecord] message.
const (
	_ = iota
	FieldEACLRecordOperation
	FieldEACLRecordAction
	FieldEACLRecordFilters
	FieldEACLRecordTargets
	FieldEACLRecordComment
)

// MarshaledSize returns size of the EACLRecord in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *EACLRecord) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeVarint(FieldEACLRecordOperation, x.Operation) +
			protoencoding.SizeVarint(FieldEACLRecordAction, x.Action) +
			protoencoding.SizeRepeatedMessages(FieldEACLRecordFilters, x.Filters) +
			protoencoding.SizeRepeatedMessages(FieldEACLRecordTargets, x.Targets) +
			protoencoding.SizeBytes(FieldEACLRecordComment, x.Comment)
	}
	return 0
}

// MarshalStable writes the EACLRecord in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *EACLRecord) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldEACLRecordOperation, x.Operation)
		off += protoencoding.MarshalToVarint(b[off:], FieldEACLRecordAction, x.Action)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldEACLRecordFilters, x.Filters)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldEACLRecordTargets, x.Targets)
		protoencoding.MarshalToBytes(b[off:], FieldEACLRecordComment, x.Comment)
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
		sz = protoencoding.SizeVarint(FieldEACLRecordFilterHeaderType, x.HeaderType) +
			protoencoding.SizeVarint(FieldEACLRecordFilterMatchType, x.MatchType) +
			protoencoding.SizeBytes(FieldEACLRecordFilterKey, x.Key) +
			protoencoding.SizeBytes(FieldEACLRecordFilterValue, x.Value)
	}
	return sz
}

// MarshalStable writes the EACLRecord_Filter in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord_Filter.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *EACLRecord_Filter) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldEACLRecordFilterHeaderType, x.HeaderType)
		off += protoencoding.MarshalToVarint(b[off:], FieldEACLRecordFilterMatchType, x.MatchType)
		off += protoencoding.MarshalToBytes(b[off:], FieldEACLRecordFilterKey, x.Key)
		protoencoding.MarshalToBytes(b[off:], FieldEACLRecordFilterValue, x.Value)
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
		sz = protoencoding.SizeVarint(FieldEACLRecordTargetRole, x.Role) +
			protoencoding.SizeRepeatedBytes(FieldEACLRecordTargetKeys, x.Keys)
	}
	return sz
}

// MarshalStable writes the EACLRecord_Target in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [EACLRecord_Target.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *EACLRecord_Target) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldEACLRecordTargetRole, x.Role)
		protoencoding.MarshalToRepeatedBytes(b[off:], FieldEACLRecordTargetKeys, x.Keys)
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
		sz = protoencoding.SizeVarint(FieldBearerTokenBodyTokenLifetimeExp, x.Exp) +
			protoencoding.SizeVarint(FieldBearerTokenBodyTokenLifetimeNbf, x.Nbf) +
			protoencoding.SizeVarint(FieldBearerTokenBodyTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the BearerToken_Body_TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [BearerToken_Body_TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *BearerToken_Body_TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldBearerTokenBodyTokenLifetimeExp, x.Exp)
		off += protoencoding.MarshalToVarint(b[off:], FieldBearerTokenBodyTokenLifetimeNbf, x.Nbf)
		protoencoding.MarshalToVarint(b[off:], FieldBearerTokenBodyTokenLifetimeIat, x.Iat)
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
		sz = protoencoding.SizeEmbedded(FieldBearerTokenBodyEACLTable, x.EaclTable) +
			protoencoding.SizeEmbedded(FieldBearerTokenBodyOwnerID, x.OwnerId) +
			protoencoding.SizeEmbedded(FieldBearerTokenBodyLifetime, x.Lifetime) +
			protoencoding.SizeEmbedded(FieldBearerTokenBodyIssuer, x.Issuer)
	}
	return sz
}

// MarshalStable writes the BearerToken_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [BearerToken_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *BearerToken_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldBearerTokenBodyEACLTable, x.EaclTable)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldBearerTokenBodyOwnerID, x.OwnerId)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldBearerTokenBodyLifetime, x.Lifetime)
		protoencoding.MarshalToEmbedded(b[off:], FieldBearerTokenBodyIssuer, x.Issuer)
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
		sz = protoencoding.SizeEmbedded(FieldBearerTokenBody, x.Body) +
			protoencoding.SizeEmbedded(FieldBearerTokenSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the BearerToken in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [BearerToken.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *BearerToken) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldBearerTokenBody, x.Body)
		protoencoding.MarshalToEmbedded(b[off:], FieldBearerTokenSignature, x.Signature)
	}
}
