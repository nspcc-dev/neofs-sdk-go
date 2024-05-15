package netmap

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldReplicaCount
	fieldReplicaSelector
)

func (x *Replica) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldReplicaCount, x.Count) +
			proto.SizeBytes(fieldReplicaSelector, x.Selector)
	}
	return sz
}

func (x *Replica) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldReplicaCount, x.Count)
		proto.MarshalBytes(b[off:], fieldReplicaSelector, x.Selector)
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

func (x *Selector) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldSelectorName, x.Name)
		off += proto.MarshalVarint(b[off:], fieldSelectorCount, x.Count)
		off += proto.MarshalVarint(b[off:], fieldSelectorClause, int32(x.Clause))
		off += proto.MarshalBytes(b[off:], fieldSelectorAttribute, x.Attribute)
		proto.MarshalBytes(b[off:], fieldSelectorFilter, x.Filter)
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

func (x *Filter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldFilterName, x.Name) +
			proto.SizeBytes(fieldFilterKey, x.Key) +
			proto.SizeVarint(fieldFilterOp, int32(x.Op)) +
			proto.SizeBytes(fieldFilterVal, x.Value)
		for i := range x.Filters {
			sz += proto.SizeNested(fieldFilterSubs, x.Filters[i])
		}
	}
	return sz
}

func (x *Filter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldFilterName, x.Name)
		off += proto.MarshalBytes(b[off:], fieldFilterKey, x.Key)
		off += proto.MarshalVarint(b[off:], fieldFilterOp, int32(x.Op))
		off += proto.MarshalBytes(b[off:], fieldFilterVal, x.Value)
		for i := range x.Filters {
			off += proto.MarshalNested(b[off:], fieldFilterSubs, x.Filters[i])
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

func (x *PlacementPolicy) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldPolicyBackupFactor, x.ContainerBackupFactor) +
			proto.SizeNested(fieldPolicySubnet, x.SubnetId)
		for i := range x.Replicas {
			sz += proto.SizeNested(fieldPolicyReplicas, x.Replicas[i])
		}
		for i := range x.Selectors {
			sz += proto.SizeNested(fieldPolicySelectors, x.Selectors[i])
		}
		for i := range x.Filters {
			sz += proto.SizeNested(fieldPolicyFilters, x.Filters[i])
		}
	}
	return sz
}

func (x *PlacementPolicy) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Replicas {
			off += proto.MarshalNested(b[off:], fieldPolicyReplicas, x.Replicas[i])
		}
		off += proto.MarshalVarint(b[off:], fieldPolicyBackupFactor, x.ContainerBackupFactor)
		for i := range x.Selectors {
			off += proto.MarshalNested(b[off:], fieldPolicySelectors, x.Selectors[i])
		}
		for i := range x.Filters {
			off += proto.MarshalNested(b[off:], fieldPolicyFilters, x.Filters[i])
		}
		proto.MarshalNested(b[off:], fieldPolicySubnet, x.SubnetId)
	}
}

const (
	_ = iota
	fieldNetPrmKey
	fieldNetPrmVal
)

func (x *NetworkConfig_Parameter) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldNetPrmKey, x.Key) +
			proto.SizeBytes(fieldNetPrmVal, x.Value)
	}
	return sz
}

func (x *NetworkConfig_Parameter) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldNetPrmKey, x.Key)
		proto.MarshalBytes(b[off:], fieldNetPrmVal, x.Value)
	}
}

const (
	_ = iota
	fieldNetConfigPrms
)

func (x *NetworkConfig) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.Parameters {
			sz += proto.SizeNested(fieldNetConfigPrms, x.Parameters[i])
		}
	}
	return sz
}

func (x *NetworkConfig) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Parameters {
			off += proto.MarshalNested(b[off:], fieldNetConfigPrms, x.Parameters[i])
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

func (x *NetworkInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldNetInfoCurEpoch, x.CurrentEpoch) +
			proto.SizeVarint(fieldNetInfoMagic, x.MagicNumber) +
			proto.SizeVarint(fieldNetInfoMSPerBlock, x.MsPerBlock) +
			proto.SizeNested(fieldNetInfoConfig, x.NetworkConfig)
	}
	return sz
}

func (x *NetworkInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldNetInfoCurEpoch, x.CurrentEpoch)
		off += proto.MarshalVarint(b[off:], fieldNetInfoMagic, x.MagicNumber)
		off += proto.MarshalVarint(b[off:], fieldNetInfoMSPerBlock, x.MsPerBlock)
		proto.MarshalNested(b[off:], fieldNetInfoConfig, x.NetworkConfig)
	}
}

const (
	_ = iota
	fieldNumNodeAttrKey
	fieldNumNodeAttrVal
	fieldNumNodeAttrParents
)

func (x *NodeInfo_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldNumNodeAttrKey, x.Key) +
			proto.SizeBytes(fieldNumNodeAttrVal, x.Value) +
			proto.SizeRepeatedBytes(fieldNumNodeAttrParents, x.Parents)
	}
	return sz
}

func (x *NodeInfo_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldNumNodeAttrKey, x.Key)
		off += proto.MarshalBytes(b[off:], fieldNumNodeAttrVal, x.Value)
		proto.MarshalRepeatedBytes(b[off:], fieldNumNodeAttrParents, x.Parents)
	}
}

const (
	_ = iota
	fieldNodeInfoPubKey
	fieldNodeInfoAddresses
	fieldNodeInfoAttributes
	fieldNodeInfoState
)

func (x *NodeInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldNodeInfoPubKey, x.PublicKey) +
			proto.SizeRepeatedBytes(fieldNodeInfoAddresses, x.Addresses) +
			proto.SizeVarint(fieldNodeInfoState, int32(x.State))
		for i := range x.Attributes {
			sz += proto.SizeNested(fieldNodeInfoAttributes, x.Attributes[i])
		}
	}
	return sz
}

func (x *NodeInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldNodeInfoPubKey, x.PublicKey)
		off += proto.MarshalRepeatedBytes(b[off:], fieldNodeInfoAddresses, x.Addresses)
		for i := range x.Attributes {
			off += proto.MarshalNested(b[off:], fieldNodeInfoAttributes, x.Attributes[i])
		}
		proto.MarshalVarint(b[off:], fieldNodeInfoState, int32(x.State))
	}
}

const (
	_ = iota
	fieldNetmapEpoch
	fieldNetmapNodes
)

func (x *Netmap) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldNetmapEpoch, x.Epoch)
		for i := range x.Nodes {
			sz += proto.SizeNested(fieldNetmapNodes, x.Nodes[i])
		}
	}
	return sz
}

func (x *Netmap) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldNetmapEpoch, x.Epoch)
		for i := range x.Nodes {
			off += proto.MarshalNested(b[off:], fieldNetmapNodes, x.Nodes[i])
		}
	}
}

func (x *LocalNodeInfoRequest_Body) MarshaledSize() int   { return 0 }
func (x *LocalNodeInfoRequest_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldNodeInfoRespVersion
	fieldNodeInfoRespInfo
)

func (x *LocalNodeInfoResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldNodeInfoRespVersion, x.Version) +
			proto.SizeNested(fieldNodeInfoRespInfo, x.NodeInfo)
	}
	return sz
}

func (x *LocalNodeInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldNodeInfoRespVersion, x.Version)
		proto.MarshalNested(b[off:], fieldNodeInfoRespInfo, x.NodeInfo)
	}
}

func (x *NetworkInfoRequest_Body) MarshaledSize() int   { return 0 }
func (x *NetworkInfoRequest_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldNetInfoRespInfo
)

func (x *NetworkInfoResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldNetInfoRespInfo, x.NetworkInfo)
	}
	return sz
}

func (x *NetworkInfoResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldNetInfoRespInfo, x.NetworkInfo)
	}
}

func (x *NetmapSnapshotRequest_Body) MarshaledSize() int   { return 0 }
func (x *NetmapSnapshotRequest_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldNetmapRespNetmap
)

func (x *NetmapSnapshotResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldNetmapRespNetmap, x.Netmap)
	}
	return sz
}

func (x *NetmapSnapshotResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldNetmapRespNetmap, x.Netmap)
	}
}
