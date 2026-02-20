package netmap

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Field numbers of [Replica] message.
const (
	_ = iota
	FieldReplicaCount
	FieldReplicaSelector
)

// MarshaledSize returns size of the Replica in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Replica) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldReplicaCount, x.Count) +
			proto.SizeBytes(FieldReplicaSelector, x.Selector)
	}
	return sz
}

// MarshalStable writes the Replica in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Replica.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Replica) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldReplicaCount, x.Count)
		proto.MarshalToBytes(b[off:], FieldReplicaSelector, x.Selector)
	}
}

// Field numbers of [Selector] message.
const (
	_ = iota
	FieldSelectorName
	FieldSelectorCount
	FieldSelectorClause
	FieldSelectorAttribute
	FieldSelectorFilter
)

// MarshaledSize returns size of the Selector in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Selector) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldSelectorName, x.Name) +
			proto.SizeVarint(FieldSelectorCount, x.Count) +
			proto.SizeVarint(FieldSelectorClause, int32(x.Clause)) +
			proto.SizeBytes(FieldSelectorAttribute, x.Attribute) +
			proto.SizeBytes(FieldSelectorFilter, x.Filter)
	}
	return sz
}

// MarshalStable writes the Selector in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Selector.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Selector) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldSelectorName, x.Name)
		off += proto.MarshalToVarint(b[off:], FieldSelectorCount, x.Count)
		off += proto.MarshalToVarint(b[off:], FieldSelectorClause, int32(x.Clause))
		off += proto.MarshalToBytes(b[off:], FieldSelectorAttribute, x.Attribute)
		proto.MarshalToBytes(b[off:], FieldSelectorFilter, x.Filter)
	}
}

// Field numbers of [Filter] message.
const (
	_ = iota
	FieldFilterName
	FieldFilterKey
	FieldFilterOp
	FieldFilterValue
	FieldFilterFilters
)

// MarshaledSize returns size of the Filter in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Filter) MarshaledSize() int {
	if x != nil {
		return proto.SizeBytes(FieldFilterName, x.Name) +
			proto.SizeBytes(FieldFilterKey, x.Key) +
			proto.SizeVarint(FieldFilterOp, int32(x.Op)) +
			proto.SizeBytes(FieldFilterValue, x.Value) +
			proto.SizeRepeatedMessages(FieldFilterFilters, x.Filters)
	}
	return 0
}

// MarshalStable writes the Filter in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Filter.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldFilterName, x.Name)
		off += proto.MarshalToBytes(b[off:], FieldFilterKey, x.Key)
		off += proto.MarshalToVarint(b[off:], FieldFilterOp, int32(x.Op))
		off += proto.MarshalToBytes(b[off:], FieldFilterValue, x.Value)
		proto.MarshalToRepeatedMessages(b[off:], FieldFilterFilters, x.Filters)
	}
}

// Field numbers of [PlacementPolicy_ECRule] message.
const (
	_ = iota
	FieldPlacementPolicyECRuleDataPartNum
	FieldPlacementPolicyECRuleParityPartNum
	FieldPlacementPolicyECRuleSelector
)

// MarshaledSize returns size of the PlacementPolicy_ECRule in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *PlacementPolicy_ECRule) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldPlacementPolicyECRuleDataPartNum, x.DataPartNum) +
			proto.SizeVarint(FieldPlacementPolicyECRuleParityPartNum, x.ParityPartNum) +
			proto.SizeBytes(FieldPlacementPolicyECRuleSelector, x.Selector)
	}
	return 0
}

// MarshalStable writes the PlacementPolicy_ECRule in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [PlacementPolicy_ECRule.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *PlacementPolicy_ECRule) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldPlacementPolicyECRuleDataPartNum, x.DataPartNum)
		off += proto.MarshalToVarint(b[off:], FieldPlacementPolicyECRuleParityPartNum, x.ParityPartNum)
		proto.MarshalToBytes(b[off:], FieldPlacementPolicyECRuleSelector, x.Selector)
	}
}

// Field numbers of [PlacementPolicy] message.
const (
	_ = iota
	FieldPlacementPolicyReplicas
	FieldPlacementPolicyContainerBackupFactor
	FieldPlacementPolicySelectors
	FieldPlacementPolicyFilters
	FieldPlacementPolicySubnetID
	FieldPlacementPolicyECRules
)

// MarshaledSize returns size of the PlacementPolicy in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PlacementPolicy) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldPlacementPolicyContainerBackupFactor, x.ContainerBackupFactor) +
			proto.SizeEmbedded(FieldPlacementPolicySubnetID, x.SubnetId) +
			proto.SizeRepeatedMessages(FieldPlacementPolicyReplicas, x.Replicas) +
			proto.SizeRepeatedMessages(FieldPlacementPolicySelectors, x.Selectors) +
			proto.SizeRepeatedMessages(FieldPlacementPolicyFilters, x.Filters) +
			proto.SizeRepeatedMessages(FieldPlacementPolicyECRules, x.EcRules)
	}
	return 0
}

// MarshalStable writes the PlacementPolicy in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PlacementPolicy.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PlacementPolicy) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToRepeatedMessages(b, FieldPlacementPolicyReplicas, x.Replicas)
		off += proto.MarshalToVarint(b[off:], FieldPlacementPolicyContainerBackupFactor, x.ContainerBackupFactor)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldPlacementPolicySelectors, x.Selectors)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldPlacementPolicyFilters, x.Filters)
		off += proto.MarshalToEmbedded(b[off:], FieldPlacementPolicySubnetID, x.SubnetId)
		proto.MarshalToRepeatedMessages(b[off:], FieldPlacementPolicyECRules, x.EcRules)
	}
}

// Field numbers of [NetworkConfig_Parameter] message.
const (
	_ = iota
	FieldNetworkConfigParameterKey
	FieldNetworkConfigParameterValue
)

// MarshaledSize returns size of the NetworkConfig_Parameter in Protocol Buffers
// V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetworkConfig_Parameter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldNetworkConfigParameterKey, x.Key) +
			proto.SizeBytes(FieldNetworkConfigParameterValue, x.Value)
	}
	return sz
}

// MarshalStable writes the NetworkConfig_Parameter in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkConfig_Parameter.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkConfig_Parameter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldNetworkConfigParameterKey, x.Key)
		proto.MarshalToBytes(b[off:], FieldNetworkConfigParameterValue, x.Value)
	}
}

// Field numbers of [NetworkConfig] message.
const (
	_ = iota
	FieldNetworkConfigParameters
)

// MarshaledSize returns size of the NetworkConfig in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *NetworkConfig) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(FieldNetworkConfigParameters, x.Parameters)
	}
	return 0
}

// MarshalStable writes the NetworkConfig in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NetworkConfig.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NetworkConfig) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, FieldNetworkConfigParameters, x.Parameters)
	}
}

// Field numbers of [NetworkInfo] message.
const (
	_ = iota
	FieldNetworkInfoCurrentEpoch
	FieldNetworkInfoMagicNumber
	FieldNetworkInfoMSPerBlock
	FieldNetworkInfoConfig
)

// MarshaledSize returns size of the NetworkInfo in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *NetworkInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldNetworkInfoCurrentEpoch, x.CurrentEpoch) +
			proto.SizeVarint(FieldNetworkInfoMagicNumber, x.MagicNumber) +
			proto.SizeVarint(FieldNetworkInfoMSPerBlock, x.MsPerBlock) +
			proto.SizeEmbedded(FieldNetworkInfoConfig, x.NetworkConfig)
	}
	return sz
}

// MarshalStable writes the NetworkInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NetworkInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NetworkInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldNetworkInfoCurrentEpoch, x.CurrentEpoch)
		off += proto.MarshalToVarint(b[off:], FieldNetworkInfoMagicNumber, x.MagicNumber)
		off += proto.MarshalToVarint(b[off:], FieldNetworkInfoMSPerBlock, x.MsPerBlock)
		proto.MarshalToEmbedded(b[off:], FieldNetworkInfoConfig, x.NetworkConfig)
	}
}

// Field numbers of [NodeInfo_Attribute] message.
const (
	_ = iota
	FieldNodeInfoAttributeKey
	FieldNodeInfoAttributeValue
	FieldNodeInfoAttributeParents
)

// MarshaledSize returns size of the NodeInfo_Attribute in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *NodeInfo_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldNodeInfoAttributeKey, x.Key) +
			proto.SizeBytes(FieldNodeInfoAttributeValue, x.Value) +
			proto.SizeRepeatedBytes(FieldNodeInfoAttributeParents, x.Parents)
	}
	return sz
}

// MarshalStable writes the NodeInfo_Attribute in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [NodeInfo_Attribute.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *NodeInfo_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldNodeInfoAttributeKey, x.Key)
		off += proto.MarshalToBytes(b[off:], FieldNodeInfoAttributeValue, x.Value)
		proto.MarshalToRepeatedBytes(b[off:], FieldNodeInfoAttributeParents, x.Parents)
	}
}

// Field numbers of [NodeInfo] message.
const (
	_ = iota
	FieldNodeInfoPublicKey
	FieldNodeInfoAddresses
	FieldNodeInfoAttributes
	FieldNodeInfoState
)

// MarshaledSize returns size of the NodeInfo in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *NodeInfo) MarshaledSize() int {
	if x != nil {
		return proto.SizeBytes(FieldNodeInfoPublicKey, x.PublicKey) +
			proto.SizeRepeatedBytes(FieldNodeInfoAddresses, x.Addresses) +
			proto.SizeVarint(FieldNodeInfoState, int32(x.State)) +
			proto.SizeRepeatedMessages(FieldNodeInfoAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes the NodeInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NodeInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NodeInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldNodeInfoPublicKey, x.PublicKey)
		off += proto.MarshalToRepeatedBytes(b[off:], FieldNodeInfoAddresses, x.Addresses)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldNodeInfoAttributes, x.Attributes)
		proto.MarshalToVarint(b[off:], FieldNodeInfoState, int32(x.State))
	}
}

// Field numbers of [Netmap] message.
const (
	_ = iota
	FieldNetmapEpoch
	FieldNetmapNodes
)

// MarshaledSize returns size of the Netmap in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Netmap) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldNetmapEpoch, x.Epoch) +
			proto.SizeRepeatedMessages(FieldNetmapNodes, x.Nodes)
	}
	return 0
}

// MarshalStable writes the Netmap in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Netmap.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Netmap) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldNetmapEpoch, x.Epoch)
		proto.MarshalToRepeatedMessages(b[off:], FieldNetmapNodes, x.Nodes)
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

// Field numbers of [LocalNodeInfoResponse_Body] message.
const (
	_ = iota
	FieldLocalNodeInfoResponseBodyVersion
	FieldLocalNodeInfoResponseBodyNodeInfo
)

// MarshaledSize returns size of the LocalNodeInfoResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *LocalNodeInfoResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldLocalNodeInfoResponseBodyVersion, x.Version) +
			proto.SizeEmbedded(FieldLocalNodeInfoResponseBodyNodeInfo, x.NodeInfo)
	}
	return sz
}

// MarshalStable writes the LocalNodeInfoResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [LocalNodeInfoResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *LocalNodeInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldLocalNodeInfoResponseBodyVersion, x.Version)
		proto.MarshalToEmbedded(b[off:], FieldLocalNodeInfoResponseBodyNodeInfo, x.NodeInfo)
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

// Field numbers of [NetworkInfoResponse_Body] message.
const (
	_ = iota
	FieldNetworkInfoResponseBodyInfo
)

// MarshaledSize returns size of the NetworkInfoResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetworkInfoResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldNetworkInfoResponseBodyInfo, x.NetworkInfo)
	}
	return sz
}

// MarshalStable writes the NetworkInfoResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkInfoResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldNetworkInfoResponseBodyInfo, x.NetworkInfo)
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

// Field numbers of [NetmapSnapshotResponse_Body] message.
const (
	_ = iota
	FieldNetmapSnapshotResponseBodyNetmap
)

// MarshaledSize returns size of the NetmapSnapshotResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *NetmapSnapshotResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldNetmapSnapshotResponseBodyNetmap, x.Netmap)
	}
	return sz
}

// MarshalStable writes the NetmapSnapshotResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetmapSnapshotResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetmapSnapshotResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldNetmapSnapshotResponseBodyNetmap, x.Netmap)
	}
}
