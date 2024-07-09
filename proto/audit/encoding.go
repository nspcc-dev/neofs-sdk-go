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
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldResultVersion, x.Version)
		sz += proto.SizeFixed64(fieldResultEpoch, x.AuditEpoch)
		sz += proto.SizeEmbedded(fieldResultContainer, x.ContainerId)
		sz += proto.SizeBytes(fieldResultPubKey, x.PublicKey)
		sz += proto.SizeBool(fieldResultCompleted, x.Complete)
		sz += proto.SizeVarint(fieldResultRequests, x.Requests)
		sz += proto.SizeVarint(fieldResultRetries, x.Retries)
		sz += proto.SizeVarint(fieldResultHits, x.Hit)
		sz += proto.SizeVarint(fieldResultMisses, x.Miss)
		sz += proto.SizeVarint(fieldResultFailures, x.Fail)
		sz += proto.SizeRepeatedBytes(fieldResultPassedNodes, x.PassNodes)
		sz += proto.SizeRepeatedBytes(fieldResultFailedNodes, x.FailNodes)
		for i := range x.PassSg {
			sz += proto.SizeEmbedded(fieldResultPassedSG, x.PassSg[i])
		}
		for i := range x.FailSg {
			sz += proto.SizeEmbedded(fieldResultFailedSG, x.FailSg[i])
		}
	}
	return sz
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
		for i := range x.PassSg {
			off += proto.MarshalToEmbedded(b[off:], fieldResultPassedSG, x.PassSg[i])
		}
		for i := range x.FailSg {
			off += proto.MarshalToEmbedded(b[off:], fieldResultFailedSG, x.FailSg[i])
		}
		off += proto.MarshalToVarint(b[off:], fieldResultHits, x.Hit)
		off += proto.MarshalToVarint(b[off:], fieldResultMisses, x.Miss)
		off += proto.MarshalToVarint(b[off:], fieldResultFailures, x.Fail)
		off += proto.MarshalToRepeatedBytes(b[off:], fieldResultPassedNodes, x.PassNodes)
		proto.MarshalToRepeatedBytes(b[off:], fieldResultFailedNodes, x.FailNodes)
	}
}
