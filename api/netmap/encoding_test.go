package netmap_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLocalNodeInfoRequest_Body(t *testing.T) {
	var v netmap.LocalNodeInfoRequest_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}

func TestLocalNodeInfoResponse_Body(t *testing.T) {
	v := &netmap.LocalNodeInfoResponse_Body{
		Version: &refs.Version{Major: 1, Minor: 2},
		NodeInfo: &netmap.NodeInfo{
			PublicKey: []byte("any_key"),
			Addresses: []string{"addr1", "addr2"},
			Attributes: []*netmap.NodeInfo_Attribute{
				{Key: "key1", Value: "val1", Parents: []string{"par1", "par2"}},
				{Key: "key2", Value: "val2", Parents: []string{"par3", "par4"}},
			},
			State: 3,
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res netmap.LocalNodeInfoResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Version, res.Version)
	require.Equal(t, v.NodeInfo, res.NodeInfo)
}

func TestNetworkInfoRequest_Body(t *testing.T) {
	var v netmap.NetworkInfoRequest_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}

func TestNetworkInfoResponse_Body(t *testing.T) {
	v := &netmap.NetworkInfoResponse_Body{
		NetworkInfo: &netmap.NetworkInfo{
			CurrentEpoch: 1, MagicNumber: 2, MsPerBlock: 3,
			NetworkConfig: &netmap.NetworkConfig{
				Parameters: []*netmap.NetworkConfig_Parameter{
					{Key: []byte("key1"), Value: []byte("val1")},
					{Key: []byte("key2"), Value: []byte("val2")},
				},
			},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res netmap.NetworkInfoResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.NetworkInfo, res.NetworkInfo)
}

func TestNetmapSnapshotRequest_Body(t *testing.T) {
	var v netmap.NetworkInfoRequest_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}

func TestNetmapSnapshotResponse_Body(t *testing.T) {
	v := &netmap.NetmapSnapshotResponse_Body{
		Netmap: &netmap.Netmap{
			Epoch: 1,
			Nodes: []*netmap.NodeInfo{
				{
					PublicKey: []byte("any_key1"),
					Addresses: []string{"addr1", "addr2"},
					Attributes: []*netmap.NodeInfo_Attribute{
						{Key: "key1", Value: "val1", Parents: []string{"par1", "par2"}},
						{Key: "key2", Value: "val2", Parents: []string{"par3", "par4"}},
					},
					State: 2,
				},
				{
					PublicKey: []byte("any_key2"),
					Addresses: []string{"addr3", "addr4"},
					Attributes: []*netmap.NodeInfo_Attribute{
						{Key: "key3", Value: "val3", Parents: []string{"par5", "par6"}},
						{Key: "key4", Value: "val4", Parents: []string{"par7", "par8"}},
					},
					State: 3,
				},
			},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res netmap.NetmapSnapshotResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Netmap, res.Netmap)
}
