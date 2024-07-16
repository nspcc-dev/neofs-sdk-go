package netmap_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
)

// returns random netmap.NetworkConfig_Parameter with all non-zero fields.
func randNetworkConfigParameter() *netmap.NetworkConfig_Parameter {
	return &netmap.NetworkConfig_Parameter{
		Key: prototest.RandBytes(), Value: prototest.RandBytes(),
	}
}

// returns non-empty list of netmap.NetworkConfig_Parameter up to 10 elements.
// Each element may be nil and pointer to zero.
func randNetworkConfigParameters() []*netmap.NetworkConfig_Parameter {
	return prototest.RandRepeated(randNetworkConfigParameter)
}

// returns random netmap.NetworkConfig with all non-zero fields.
func randNetworkConfig() *netmap.NetworkConfig {
	return &netmap.NetworkConfig{
		Parameters: randNetworkConfigParameters(),
	}
}

// returns random netmap.NetworkInfo with all non-zero fields.
func randNetworkInfo() *netmap.NetworkInfo {
	return &netmap.NetworkInfo{
		CurrentEpoch:  prototest.RandUint64(),
		MagicNumber:   prototest.RandUint64(),
		MsPerBlock:    prototest.RandInt64(),
		NetworkConfig: randNetworkConfig(),
	}
}

// returns random netmap.NodeInfo_Attribute with all non-zero fields.
func randNodeAttribute() *netmap.NodeInfo_Attribute {
	return &netmap.NodeInfo_Attribute{
		Key:     prototest.RandString(),
		Value:   prototest.RandString(),
		Parents: prototest.RandStrings(),
	}
}

// returns non-empty list of netmap.NodeInfo_Attribute up to 10 elements. Each
// element may be nil and pointer to zero.
func randNodeAttributes() []*netmap.NodeInfo_Attribute {
	return prototest.RandRepeated(randNodeAttribute)
}

// returns random netmap.NodeInfo with all non-zero fields.
func randNode() *netmap.NodeInfo {
	return &netmap.NodeInfo{
		PublicKey:  prototest.RandBytes(),
		Addresses:  prototest.RandStrings(),
		Attributes: randNodeAttributes(),
		State:      prototest.RandInteger[netmap.NodeInfo_State](),
	}
}

// returns non-empty list of netmap.NodeInfo up to 10 elements. Each element may
// be nil and pointer to zero.
func randNodes() []*netmap.NodeInfo { return prototest.RandRepeated(randNode) }

// returns random netmap.Netmap with all non-zero fields.
func randNetmap() *netmap.Netmap {
	return &netmap.Netmap{
		Epoch: prototest.RandUint64(),
		Nodes: randNodes(),
	}
}

func TestReplica_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.Replica{
		prototest.RandPlacementReplica(),
	})
}

func TestSelector_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.Selector{
		prototest.RandPlacementSelector(),
	})
}

func TestFilter_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.Filter{
		prototest.RandPlacementFilter(),
	})
}

func TestPlacement_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.PlacementPolicy{
		prototest.RandPlacementPolicy(),
	})
}

func TestNetworkConfig_Parameter_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetworkConfig_Parameter{
		randNetworkConfigParameter(),
	})
}

func TestNetworkConfig_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetworkConfig{
		randNetworkConfig(),
	})
}

func TestNetworkInfo_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetworkInfo{
		randNetworkInfo(),
	})
}

func TestNodeInfo_Attribute_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NodeInfo_Attribute{
		randNodeAttribute(),
	})
}

func TestNodeInfo_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NodeInfo{
		randNode(),
	})
}

func TestNetmap_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.Netmap{
		randNetmap(),
	})
}

func TestLocalNodeInfoRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.LocalNodeInfoRequest_Body{})
}

func TestLocalNodeInfoResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.LocalNodeInfoResponse_Body{
		{
			Version:  prototest.RandVersion(),
			NodeInfo: randNode(),
		},
	})
}

func TestNetworkInfoRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetworkInfoRequest_Body{})
}

func TestNetworkInfoResponse_Body_MarshaledSize(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetworkInfoResponse_Body{
		{NetworkInfo: randNetworkInfo()},
	})
}

func TestNetmapSnapshotRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetmapSnapshotRequest_Body{})
}

func TestNetmapSnapshotResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*netmap.NetmapSnapshotResponse_Body{
		{Netmap: randNetmap()},
	})
}
