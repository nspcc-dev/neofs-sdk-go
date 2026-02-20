package audit

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

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
		return proto.SizeEmbedded(FieldDataAuditResultVersion, x.Version) +
			proto.SizeFixed64(FieldDataAuditResultAuditEpoch, x.AuditEpoch) +
			proto.SizeEmbedded(FieldDataAuditResultContainerID, x.ContainerId) +
			proto.SizeBytes(FieldDataAuditResultPublicKey, x.PublicKey) +
			proto.SizeBool(FieldDataAuditResultComplete, x.Complete) +
			proto.SizeVarint(FieldDataAuditResultRequests, x.Requests) +
			proto.SizeVarint(FieldDataAuditResultRetries, x.Retries) +
			proto.SizeVarint(FieldDataAuditResultHit, x.Hit) +
			proto.SizeVarint(FieldDataAuditResultMiss, x.Miss) +
			proto.SizeVarint(FieldDataAuditResultFail, x.Fail) +
			proto.SizeRepeatedBytes(FieldDataAuditResultPassNodes, x.PassNodes) +
			proto.SizeRepeatedBytes(FieldDataAuditResultFailNodes, x.FailNodes) +
			proto.SizeRepeatedMessages(FieldDataAuditResultPassSG, x.PassSg) +
			proto.SizeRepeatedMessages(FieldDataAuditResultFailSG, x.FailSg)
	}
	return 0
}

// MarshalStable writes the DataAuditResult in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [DataAuditResult.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *DataAuditResult) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldDataAuditResultVersion, x.Version)
		off += proto.MarshalToFixed64(b[off:], FieldDataAuditResultAuditEpoch, x.AuditEpoch)
		off += proto.MarshalToEmbedded(b[off:], FieldDataAuditResultContainerID, x.ContainerId)
		off += proto.MarshalToBytes(b[off:], FieldDataAuditResultPublicKey, x.PublicKey)
		off += proto.MarshalToBool(b[off:], FieldDataAuditResultComplete, x.Complete)
		off += proto.MarshalToVarint(b[off:], FieldDataAuditResultRequests, x.Requests)
		off += proto.MarshalToVarint(b[off:], FieldDataAuditResultRetries, x.Retries)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldDataAuditResultPassSG, x.PassSg)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldDataAuditResultFailSG, x.FailSg)
		off += proto.MarshalToVarint(b[off:], FieldDataAuditResultHit, x.Hit)
		off += proto.MarshalToVarint(b[off:], FieldDataAuditResultMiss, x.Miss)
		off += proto.MarshalToVarint(b[off:], FieldDataAuditResultFail, x.Fail)
		off += proto.MarshalToRepeatedBytes(b[off:], FieldDataAuditResultPassNodes, x.PassNodes)
		proto.MarshalToRepeatedBytes(b[off:], FieldDataAuditResultFailNodes, x.FailNodes)
	}
}
