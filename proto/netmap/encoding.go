package netmap

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldReplicaCount
	fieldReplicaSelector
)

// MarshaledSize returns size of the Replica in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Replica) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldReplicaCount, x.Count) +
			proto.SizeBytes(fieldReplicaSelector, x.Selector)
	}
	return sz
}

// MarshalStable writes the Replica in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Replica.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Replica) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldReplicaCount, x.Count)
		proto.MarshalToBytes(b[off:], fieldReplicaSelector, x.Selector)
	}
}

const (
	_ = iota
	fieldSelectorName
	fieldSelectorCount
	fieldSelectorClause
	fieldSelectorAttribute
	fieldSelectorFilter
)

// MarshaledSize returns size of the Selector in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Selector) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSelectorName, x.Name) +
			proto.SizeVarint(fieldSelectorCount, x.Count) +
			proto.SizeVarint(fieldSelectorClause, int32(x.Clause)) +
			proto.SizeBytes(fieldSelectorAttribute, x.Attribute) +
			proto.SizeBytes(fieldSelectorFilter, x.Filter)
	}
	return sz
}

// MarshalStable writes the Selector in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Selector.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Selector) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldSelectorName, x.Name)
		off += proto.MarshalToVarint(b[off:], fieldSelectorCount, x.Count)
		off += proto.MarshalToVarint(b[off:], fieldSelectorClause, int32(x.Clause))
		off += proto.MarshalToBytes(b[off:], fieldSelectorAttribute, x.Attribute)
		proto.MarshalToBytes(b[off:], fieldSelectorFilter, x.Filter)
	}
}

const (
	_ = iota
	fieldFilterName
	fieldFilterKey
	fieldFilterOp
	fieldFilterVal
	fieldFilterSubs
)

// MarshaledSize returns size of the Filter in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Filter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldFilterName, x.Name) +
			proto.SizeBytes(fieldFilterKey, x.Key) +
			proto.SizeVarint(fieldFilterOp, int32(x.Op)) +
			proto.SizeBytes(fieldFilterVal, x.Value)
		for i := range x.Filters {
			sz += proto.SizeEmbedded(fieldFilterSubs, x.Filters[i])
		}
	}
	return sz
}

// MarshalStable writes the Filter in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Filter.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldFilterName, x.Name)
		off += proto.MarshalToBytes(b[off:], fieldFilterKey, x.Key)
		off += proto.MarshalToVarint(b[off:], fieldFilterOp, int32(x.Op))
		off += proto.MarshalToBytes(b[off:], fieldFilterVal, x.Value)
		for i := range x.Filters {
			off += proto.MarshalToEmbedded(b[off:], fieldFilterSubs, x.Filters[i])
		}
	}
}

const (
	_ = iota
	fieldPolicyReplicas
	fieldPolicyBackupFactor
	fieldPolicySelectors
	fieldPolicyFilters
	fieldPolicySubnet
)

// MarshaledSize returns size of the PlacementPolicy in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PlacementPolicy) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldPolicyBackupFactor, x.ContainerBackupFactor) +
			proto.SizeEmbedded(fieldPolicySubnet, x.SubnetId)
		for i := range x.Replicas {
			sz += proto.SizeEmbedded(fieldPolicyReplicas, x.Replicas[i])
		}
		for i := range x.Selectors {
			sz += proto.SizeEmbedded(fieldPolicySelectors, x.Selectors[i])
		}
		for i := range x.Filters {
			sz += proto.SizeEmbedded(fieldPolicyFilters, x.Filters[i])
		}
	}
	return sz
}

// MarshalStable writes the PlacementPolicy in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PlacementPolicy.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PlacementPolicy) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Replicas {
			off += proto.MarshalToEmbedded(b[off:], fieldPolicyReplicas, x.Replicas[i])
		}
		off += proto.MarshalToVarint(b[off:], fieldPolicyBackupFactor, x.ContainerBackupFactor)
		for i := range x.Selectors {
			off += proto.MarshalToEmbedded(b[off:], fieldPolicySelectors, x.Selectors[i])
		}
		for i := range x.Filters {
			off += proto.MarshalToEmbedded(b[off:], fieldPolicyFilters, x.Filters[i])
		}
		proto.MarshalToEmbedded(b[off:], fieldPolicySubnet, x.SubnetId)
	}
}

const (
	_ = iota
	fieldNetPrmKey
	fieldNetPrmVal
)

// MarshaledSize returns size of the NetworkConfig_Parameter in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetworkConfig_Parameter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldNetPrmKey, x.Key) +
			proto.SizeBytes(fieldNetPrmVal, x.Value)
	}
	return sz
}

// MarshalStable writes the NetworkConfig_Parameter in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkConfig_Parameter.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkConfig_Parameter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldNetPrmKey, x.Key)
		proto.MarshalToBytes(b[off:], fieldNetPrmVal, x.Value)
	}
}

const (
	_ = iota
	fieldNetConfigPrms
)

// MarshaledSize returns size of the NetworkConfig in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *NetworkConfig) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.Parameters {
			sz += proto.SizeEmbedded(fieldNetConfigPrms, x.Parameters[i])
		}
	}
	return sz
}

// MarshalStable writes the NetworkConfig in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NetworkConfig.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NetworkConfig) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Parameters {
			off += proto.MarshalToEmbedded(b[off:], fieldNetConfigPrms, x.Parameters[i])
		}
	}
}

const (
	_ = iota
	fieldNetInfoCurEpoch
	fieldNetInfoMagic
	fieldNetInfoMSPerBlock
	fieldNetInfoConfig
)

// MarshaledSize returns size of the NetworkInfo in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *NetworkInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldNetInfoCurEpoch, x.CurrentEpoch) +
			proto.SizeVarint(fieldNetInfoMagic, x.MagicNumber) +
			proto.SizeVarint(fieldNetInfoMSPerBlock, x.MsPerBlock) +
			proto.SizeEmbedded(fieldNetInfoConfig, x.NetworkConfig)
	}
	return sz
}

// MarshalStable writes the NetworkInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NetworkInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NetworkInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldNetInfoCurEpoch, x.CurrentEpoch)
		off += proto.MarshalToVarint(b[off:], fieldNetInfoMagic, x.MagicNumber)
		off += proto.MarshalToVarint(b[off:], fieldNetInfoMSPerBlock, x.MsPerBlock)
		proto.MarshalToEmbedded(b[off:], fieldNetInfoConfig, x.NetworkConfig)
	}
}

const (
	_ = iota
	fieldNumNodeAttrKey
	fieldNumNodeAttrVal
	fieldNumNodeAttrParents
)

// MarshaledSize returns size of the NodeInfo_Attribute in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *NodeInfo_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldNumNodeAttrKey, x.Key) +
			proto.SizeBytes(fieldNumNodeAttrVal, x.Value) +
			proto.SizeRepeatedBytes(fieldNumNodeAttrParents, x.Parents)
	}
	return sz
}

// MarshalStable writes the NodeInfo_Attribute in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [NodeInfo_Attribute.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *NodeInfo_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldNumNodeAttrKey, x.Key)
		off += proto.MarshalToBytes(b[off:], fieldNumNodeAttrVal, x.Value)
		proto.MarshalToRepeatedBytes(b[off:], fieldNumNodeAttrParents, x.Parents)
	}
}

const (
	_ = iota
	fieldNodeInfoPubKey
	fieldNodeInfoAddresses
	fieldNodeInfoAttributes
	fieldNodeInfoState
)

// MarshaledSize returns size of the NodeInfo in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *NodeInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldNodeInfoPubKey, x.PublicKey) +
			proto.SizeRepeatedBytes(fieldNodeInfoAddresses, x.Addresses) +
			proto.SizeVarint(fieldNodeInfoState, int32(x.State))
		for i := range x.Attributes {
			sz += proto.SizeEmbedded(fieldNodeInfoAttributes, x.Attributes[i])
		}
	}
	return sz
}

// MarshalStable writes the NodeInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NodeInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NodeInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldNodeInfoPubKey, x.PublicKey)
		off += proto.MarshalToRepeatedBytes(b[off:], fieldNodeInfoAddresses, x.Addresses)
		for i := range x.Attributes {
			off += proto.MarshalToEmbedded(b[off:], fieldNodeInfoAttributes, x.Attributes[i])
		}
		proto.MarshalToVarint(b[off:], fieldNodeInfoState, int32(x.State))
	}
}

const (
	_ = iota
	fieldNetmapEpoch
	fieldNetmapNodes
)

// MarshaledSize returns size of the Netmap in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Netmap) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldNetmapEpoch, x.Epoch)
		for i := range x.Nodes {
			sz += proto.SizeEmbedded(fieldNetmapNodes, x.Nodes[i])
		}
	}
	return sz
}

// MarshalStable writes the Netmap in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Netmap.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Netmap) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldNetmapEpoch, x.Epoch)
		for i := range x.Nodes {
			off += proto.MarshalToEmbedded(b[off:], fieldNetmapNodes, x.Nodes[i])
		}
	}
}

// MarshaledSize returns size of the LocalNodeInfoRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *LocalNodeInfoRequest_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the LocalNodeInfoRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [LocalNodeInfoRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *LocalNodeInfoRequest_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldNodeInfoRespVersion
	fieldNodeInfoRespInfo
)

// MarshaledSize returns size of the LocalNodeInfoResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *LocalNodeInfoResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldNodeInfoRespVersion, x.Version) +
			proto.SizeEmbedded(fieldNodeInfoRespInfo, x.NodeInfo)
	}
	return sz
}

// MarshalStable writes the LocalNodeInfoResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [LocalNodeInfoResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *LocalNodeInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldNodeInfoRespVersion, x.Version)
		proto.MarshalToEmbedded(b[off:], fieldNodeInfoRespInfo, x.NodeInfo)
	}
}

// MarshaledSize returns size of the NetworkInfoRequest_Body in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetworkInfoRequest_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the NetworkInfoRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkInfoRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkInfoRequest_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldNetInfoRespInfo
)

// MarshaledSize returns size of the NetworkInfoResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetworkInfoResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldNetInfoRespInfo, x.NetworkInfo)
	}
	return sz
}

// MarshalStable writes the NetworkInfoResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkInfoResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldNetInfoRespInfo, x.NetworkInfo)
	}
}

// MarshaledSize returns size of the NetmapSnapshotRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetmapSnapshotRequest_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the NetmapSnapshotRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetmapSnapshotRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetmapSnapshotRequest_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldNetmapRespNetmap
)

// MarshaledSize returns size of the NetmapSnapshotResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetmapSnapshotResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldNetmapRespNetmap, x.Netmap)
	}
	return sz
}

// MarshalStable writes the NetmapSnapshotResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetmapSnapshotResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetmapSnapshotResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldNetmapRespNetmap, x.Netmap)
	}
}
