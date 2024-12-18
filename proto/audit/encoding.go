package audit

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldResultVersion
	fieldResultEpoch
	fieldResultContainer
	fieldResultPubKey
	fieldResultCompleted
	fieldResultRequests
	fieldResultRetries
	fieldResultPassedSG
	fieldResultFailedSG
	fieldResultHits
	fieldResultMisses
	fieldResultFailures
	fieldResultPassedNodes
	fieldResultFailedNodes
)

// MarshaledSize returns size of the DataAuditResult in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DataAuditResult) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldResultVersion, x.Version) +
			proto.SizeFixed64(fieldResultEpoch, x.AuditEpoch) +
			proto.SizeEmbedded(fieldResultContainer, x.ContainerId) +
			proto.SizeBytes(fieldResultPubKey, x.PublicKey) +
			proto.SizeBool(fieldResultCompleted, x.Complete) +
			proto.SizeVarint(fieldResultRequests, x.Requests) +
			proto.SizeVarint(fieldResultRetries, x.Retries) +
			proto.SizeVarint(fieldResultHits, x.Hit) +
			proto.SizeVarint(fieldResultMisses, x.Miss) +
			proto.SizeVarint(fieldResultFailures, x.Fail) +
			proto.SizeRepeatedBytes(fieldResultPassedNodes, x.PassNodes) +
			proto.SizeRepeatedBytes(fieldResultFailedNodes, x.FailNodes) +
			proto.SizeRepeatedMessages(fieldResultPassedSG, x.PassSg) +
			proto.SizeRepeatedMessages(fieldResultFailedSG, x.FailSg)
	}
	return 0
}

// MarshalStable writes the DataAuditResult in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [DataAuditResult.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *DataAuditResult) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldResultVersion, x.Version)
		off += proto.MarshalToFixed64(b[off:], fieldResultEpoch, x.AuditEpoch)
		off += proto.MarshalToEmbedded(b[off:], fieldResultContainer, x.ContainerId)
		off += proto.MarshalToBytes(b[off:], fieldResultPubKey, x.PublicKey)
		off += proto.MarshalToBool(b[off:], fieldResultCompleted, x.Complete)
		off += proto.MarshalToVarint(b[off:], fieldResultRequests, x.Requests)
		off += proto.MarshalToVarint(b[off:], fieldResultRetries, x.Retries)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldResultPassedSG, x.PassSg)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldResultFailedSG, x.FailSg)
		off += proto.MarshalToVarint(b[off:], fieldResultHits, x.Hit)
		off += proto.MarshalToVarint(b[off:], fieldResultMisses, x.Miss)
		off += proto.MarshalToVarint(b[off:], fieldResultFailures, x.Fail)
		off += proto.MarshalToRepeatedBytes(b[off:], fieldResultPassedNodes, x.PassNodes)
		proto.MarshalToRepeatedBytes(b[off:], fieldResultFailedNodes, x.FailNodes)
	}
}
