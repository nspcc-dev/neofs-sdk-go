package netmap

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
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
		sz = protoencoding.SizeVarint(FieldReplicaCount, x.Count) +
			protoencoding.SizeBytes(FieldReplicaSelector, x.Selector)
	}
	return sz
}

// MarshalStable writes the Replica in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Replica.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Replica) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldReplicaCount, x.Count)
		protoencoding.MarshalToBytes(b[off:], FieldReplicaSelector, x.Selector)
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
		sz = protoencoding.SizeBytes(FieldSelectorName, x.Name) +
			protoencoding.SizeVarint(FieldSelectorCount, x.Count) +
			protoencoding.SizeVarint(FieldSelectorClause, x.Clause) +
			protoencoding.SizeBytes(FieldSelectorAttribute, x.Attribute) +
			protoencoding.SizeBytes(FieldSelectorFilter, x.Filter)
	}
	return sz
}

// MarshalStable writes the Selector in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Selector.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Selector) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldSelectorName, x.Name)
		off += protoencoding.MarshalToVarint(b[off:], FieldSelectorCount, x.Count)
		off += protoencoding.MarshalToVarint(b[off:], FieldSelectorClause, x.Clause)
		off += protoencoding.MarshalToBytes(b[off:], FieldSelectorAttribute, x.Attribute)
		protoencoding.MarshalToBytes(b[off:], FieldSelectorFilter, x.Filter)
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
		return protoencoding.SizeBytes(FieldFilterName, x.Name) +
			protoencoding.SizeBytes(FieldFilterKey, x.Key) +
			protoencoding.SizeVarint(FieldFilterOp, x.Op) +
			protoencoding.SizeBytes(FieldFilterValue, x.Value) +
			protoencoding.SizeRepeatedMessages(FieldFilterFilters, x.Filters)
	}
	return 0
}

// MarshalStable writes the Filter in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Filter.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Filter) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldFilterName, x.Name)
		off += protoencoding.MarshalToBytes(b[off:], FieldFilterKey, x.Key)
		off += protoencoding.MarshalToVarint(b[off:], FieldFilterOp, x.Op)
		off += protoencoding.MarshalToBytes(b[off:], FieldFilterValue, x.Value)
		protoencoding.MarshalToRepeatedMessages(b[off:], FieldFilterFilters, x.Filters)
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
		return protoencoding.SizeVarint(FieldPlacementPolicyECRuleDataPartNum, x.DataPartNum) +
			protoencoding.SizeVarint(FieldPlacementPolicyECRuleParityPartNum, x.ParityPartNum) +
			protoencoding.SizeBytes(FieldPlacementPolicyECRuleSelector, x.Selector)
	}
	return 0
}

// MarshalStable writes the PlacementPolicy_ECRule in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [PlacementPolicy_ECRule.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *PlacementPolicy_ECRule) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldPlacementPolicyECRuleDataPartNum, x.DataPartNum)
		off += protoencoding.MarshalToVarint(b[off:], FieldPlacementPolicyECRuleParityPartNum, x.ParityPartNum)
		protoencoding.MarshalToBytes(b[off:], FieldPlacementPolicyECRuleSelector, x.Selector)
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
	FieldPlacementPolicyInitial
)

// MarshaledSize returns size of the PlacementPolicy in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PlacementPolicy) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeVarint(FieldPlacementPolicyContainerBackupFactor, x.ContainerBackupFactor) +
			protoencoding.SizeEmbedded(FieldPlacementPolicySubnetID, x.SubnetId) +
			protoencoding.SizeRepeatedMessages(FieldPlacementPolicyReplicas, x.Replicas) +
			protoencoding.SizeRepeatedMessages(FieldPlacementPolicySelectors, x.Selectors) +
			protoencoding.SizeRepeatedMessages(FieldPlacementPolicyFilters, x.Filters) +
			protoencoding.SizeRepeatedMessages(FieldPlacementPolicyECRules, x.EcRules) +
			protoencoding.SizeEmbedded(FieldPlacementPolicyInitial, x.Initial)
	}
	return 0
}

// MarshalStable writes the PlacementPolicy in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PlacementPolicy.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PlacementPolicy) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToRepeatedMessages(b, FieldPlacementPolicyReplicas, x.Replicas)
		off += protoencoding.MarshalToVarint(b[off:], FieldPlacementPolicyContainerBackupFactor, x.ContainerBackupFactor)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldPlacementPolicySelectors, x.Selectors)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldPlacementPolicyFilters, x.Filters)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldPlacementPolicySubnetID, x.SubnetId)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldPlacementPolicyECRules, x.EcRules)
		protoencoding.MarshalToEmbedded(b[off:], FieldPlacementPolicyInitial, x.Initial)
	}
}

// Field numbers of [InitialPlacementPolicy] message.
const (
	_ = iota
	FieldInitialPlacementPolicyReplicaLimits
	FieldInitialPlacementPolicyMaxReplicas
	FieldInitialPlacementPolicyPreferLocal
)

// MarshaledSize returns size of x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *PlacementPolicy_Initial) MarshaledSize() int {
	if x != nil {
		return protoencoding.SizeRepeatedVarint(FieldInitialPlacementPolicyReplicaLimits, x.ReplicaLimits) +
			protoencoding.SizeVarint(FieldInitialPlacementPolicyMaxReplicas, x.MaxReplicas) +
			protoencoding.SizeBool(FieldInitialPlacementPolicyPreferLocal, x.PreferLocal)
	}
	return 0
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [InitialPlacementPolicy.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *PlacementPolicy_Initial) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToRepeatedVarint(b, FieldInitialPlacementPolicyReplicaLimits, x.ReplicaLimits)
		off += protoencoding.MarshalToVarint(b[off:], FieldInitialPlacementPolicyMaxReplicas, x.MaxReplicas)
		protoencoding.MarshalToBool(b[off:], FieldInitialPlacementPolicyPreferLocal, x.PreferLocal)
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
		sz = protoencoding.SizeBytes(FieldNetworkConfigParameterKey, x.Key) +
			protoencoding.SizeBytes(FieldNetworkConfigParameterValue, x.Value)
	}
	return sz
}

// MarshalStable writes the NetworkConfig_Parameter in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkConfig_Parameter.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkConfig_Parameter) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldNetworkConfigParameterKey, x.Key)
		protoencoding.MarshalToBytes(b[off:], FieldNetworkConfigParameterValue, x.Value)
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
		return protoencoding.SizeRepeatedMessages(FieldNetworkConfigParameters, x.Parameters)
	}
	return 0
}

// MarshalStable writes the NetworkConfig in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NetworkConfig.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NetworkConfig) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToRepeatedMessages(b, FieldNetworkConfigParameters, x.Parameters)
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
		sz = protoencoding.SizeVarint(FieldNetworkInfoCurrentEpoch, x.CurrentEpoch) +
			protoencoding.SizeVarint(FieldNetworkInfoMagicNumber, x.MagicNumber) +
			protoencoding.SizeVarint(FieldNetworkInfoMSPerBlock, x.MsPerBlock) +
			protoencoding.SizeEmbedded(FieldNetworkInfoConfig, x.NetworkConfig)
	}
	return sz
}

// MarshalStable writes the NetworkInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NetworkInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NetworkInfo) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldNetworkInfoCurrentEpoch, x.CurrentEpoch)
		off += protoencoding.MarshalToVarint(b[off:], FieldNetworkInfoMagicNumber, x.MagicNumber)
		off += protoencoding.MarshalToVarint(b[off:], FieldNetworkInfoMSPerBlock, x.MsPerBlock)
		protoencoding.MarshalToEmbedded(b[off:], FieldNetworkInfoConfig, x.NetworkConfig)
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
		sz = protoencoding.SizeBytes(FieldNodeInfoAttributeKey, x.Key) +
			protoencoding.SizeBytes(FieldNodeInfoAttributeValue, x.Value) +
			protoencoding.SizeRepeatedBytes(FieldNodeInfoAttributeParents, x.Parents)
	}
	return sz
}

// MarshalStable writes the NodeInfo_Attribute in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [NodeInfo_Attribute.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *NodeInfo_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldNodeInfoAttributeKey, x.Key)
		off += protoencoding.MarshalToBytes(b[off:], FieldNodeInfoAttributeValue, x.Value)
		protoencoding.MarshalToRepeatedBytes(b[off:], FieldNodeInfoAttributeParents, x.Parents)
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
		return protoencoding.SizeBytes(FieldNodeInfoPublicKey, x.PublicKey) +
			protoencoding.SizeRepeatedBytes(FieldNodeInfoAddresses, x.Addresses) +
			protoencoding.SizeVarint(FieldNodeInfoState, x.State) +
			protoencoding.SizeRepeatedMessages(FieldNodeInfoAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes the NodeInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [NodeInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *NodeInfo) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldNodeInfoPublicKey, x.PublicKey)
		off += protoencoding.MarshalToRepeatedBytes(b[off:], FieldNodeInfoAddresses, x.Addresses)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldNodeInfoAttributes, x.Attributes)
		protoencoding.MarshalToVarint(b[off:], FieldNodeInfoState, x.State)
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
		return protoencoding.SizeVarint(FieldNetmapEpoch, x.Epoch) +
			protoencoding.SizeRepeatedMessages(FieldNetmapNodes, x.Nodes)
	}
	return 0
}

// MarshalStable writes the Netmap in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Netmap.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Netmap) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldNetmapEpoch, x.Epoch)
		protoencoding.MarshalToRepeatedMessages(b[off:], FieldNetmapNodes, x.Nodes)
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
		sz = protoencoding.SizeEmbedded(FieldLocalNodeInfoResponseBodyVersion, x.Version) +
			protoencoding.SizeEmbedded(FieldLocalNodeInfoResponseBodyNodeInfo, x.NodeInfo)
	}
	return sz
}

// MarshalStable writes the LocalNodeInfoResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [LocalNodeInfoResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *LocalNodeInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldLocalNodeInfoResponseBodyVersion, x.Version)
		protoencoding.MarshalToEmbedded(b[off:], FieldLocalNodeInfoResponseBodyNodeInfo, x.NodeInfo)
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
		sz = protoencoding.SizeEmbedded(FieldNetworkInfoResponseBodyInfo, x.NetworkInfo)
	}
	return sz
}

// MarshalStable writes the NetworkInfoResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetworkInfoResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetworkInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldNetworkInfoResponseBodyInfo, x.NetworkInfo)
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
		sz = protoencoding.SizeEmbedded(FieldNetmapSnapshotResponseBodyNetmap, x.Netmap)
	}
	return sz
}

// MarshalStable writes the NetmapSnapshotResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [NetmapSnapshotResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *NetmapSnapshotResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldNetmapSnapshotResponseBodyNetmap, x.Netmap)
	}
}
