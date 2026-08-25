package audit

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

// Field numbers of [DataAuditResult] message.
const (
	_ = iota
	FieldDataAuditResultVersion
	FieldDataAuditResultAuditEpoch
	FieldDataAuditResultContainerID
	FieldDataAuditResultPublicKey
	FieldDataAuditResultComplete
	FieldDataAuditResultRequests
	FieldDataAuditResultRetries
	FieldDataAuditResultPassSG
	FieldDataAuditResultFailSG
	FieldDataAuditResultHit
	FieldDataAuditResultMiss
	FieldDataAuditResultFail
	FieldDataAuditResultPassNodes
	FieldDataAuditResultFailNodes
)

// MarshaledSize returns size of the DataAuditResult in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DataAuditResult) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeEmbedded(FieldDataAuditResultVersion, x.Version) +
			protoencoding.SizeFixed64(FieldDataAuditResultAuditEpoch, x.AuditEpoch) +
			protoencoding.SizeEmbedded(FieldDataAuditResultContainerID, x.ContainerId) +
			protoencoding.SizeBytes(FieldDataAuditResultPublicKey, x.PublicKey) +
			protoencoding.SizeBool(FieldDataAuditResultComplete, x.Complete) +
			protoencoding.SizeVarint(FieldDataAuditResultRequests, x.Requests) +
			protoencoding.SizeVarint(FieldDataAuditResultRetries, x.Retries) +
			protoencoding.SizeVarint(FieldDataAuditResultHit, x.Hit) +
			protoencoding.SizeVarint(FieldDataAuditResultMiss, x.Miss) +
			protoencoding.SizeVarint(FieldDataAuditResultFail, x.Fail) +
			protoencoding.SizeRepeatedBytes(FieldDataAuditResultPassNodes, x.PassNodes) +
			protoencoding.SizeRepeatedBytes(FieldDataAuditResultFailNodes, x.FailNodes) +
			protoencoding.SizeRepeatedMessages(FieldDataAuditResultPassSG, x.PassSg) +
			protoencoding.SizeRepeatedMessages(FieldDataAuditResultFailSG, x.FailSg)
	}
	return 0
}

// MarshalStable writes the DataAuditResult in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [DataAuditResult.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *DataAuditResult) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldDataAuditResultVersion, x.Version)
		off += protoencoding.MarshalToFixed64(b[off:], FieldDataAuditResultAuditEpoch, x.AuditEpoch)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldDataAuditResultContainerID, x.ContainerId)
		off += protoencoding.MarshalToBytes(b[off:], FieldDataAuditResultPublicKey, x.PublicKey)
		off += protoencoding.MarshalToBool(b[off:], FieldDataAuditResultComplete, x.Complete)
		off += protoencoding.MarshalToVarint(b[off:], FieldDataAuditResultRequests, x.Requests)
		off += protoencoding.MarshalToVarint(b[off:], FieldDataAuditResultRetries, x.Retries)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldDataAuditResultPassSG, x.PassSg)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldDataAuditResultFailSG, x.FailSg)
		off += protoencoding.MarshalToVarint(b[off:], FieldDataAuditResultHit, x.Hit)
		off += protoencoding.MarshalToVarint(b[off:], FieldDataAuditResultMiss, x.Miss)
		off += protoencoding.MarshalToVarint(b[off:], FieldDataAuditResultFail, x.Fail)
		off += protoencoding.MarshalToRepeatedBytes(b[off:], FieldDataAuditResultPassNodes, x.PassNodes)
		protoencoding.MarshalToRepeatedBytes(b[off:], FieldDataAuditResultFailNodes, x.FailNodes)
	}
}
