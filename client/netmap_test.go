package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	protonetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// returns Client of Netmap service provided by given server.
func newTestNetmapClient(t testing.TB, srv protonetmap.NetmapServiceServer) *Client {
	return newClient(t, testService{desc: &protonetmap.NetmapService_ServiceDesc, impl: srv})
}

type testNetmapSnapshotServer struct {
	protonetmap.UnimplementedNetmapServiceServer

	errTransport error

	unsignedResponse bool

	statusFail bool

	unsetNetMap bool
	netMap      *protonetmap.Netmap

	signer neofscrypto.Signer
}

func (x *testNetmapSnapshotServer) NetmapSnapshot(_ context.Context, req *protonetmap.NetmapSnapshotRequest) (*protonetmap.NetmapSnapshotResponse, error) {
	var reqV2 v2netmap.SnapshotRequest
	if err := reqV2.FromGRPCMessage(req); err != nil {
		panic(err)
	}

	err := verifyServiceMessage(&reqV2)
	if err != nil {
		return nil, err
	}

	if x.errTransport != nil {
		return nil, x.errTransport
	}

	var nm *protonetmap.Netmap
	if !x.unsetNetMap {
		if x.netMap != nil {
			nm = x.netMap
		} else {
			nm = new(protonetmap.Netmap)
		}
	}
	resp := protonetmap.NetmapSnapshotResponse{
		Body: &protonetmap.NetmapSnapshotResponse_Body{
			Netmap: nm,
		},
	}
	if x.statusFail {
		resp.MetaHeader = &protosession.ResponseMetaHeader{
			Status: statusErr.ErrorToV2().ToGRPCMessage().(*protostatus.Status),
		}
	}

	var respV2 v2netmap.SnapshotResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	signer := x.signer
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	if !x.unsignedResponse {
		err = signServiceMessage(signer, &respV2, nil)
		if err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}
	}

	return respV2.ToGRPCMessage().(*protonetmap.NetmapSnapshotResponse), nil
}

type testGetNetworkInfoServer struct {
	protonetmap.UnimplementedNetmapServiceServer
}

func (x *testGetNetworkInfoServer) NetworkInfo(context.Context, *protonetmap.NetworkInfoRequest) (*protonetmap.NetworkInfoResponse, error) {
	resp := protonetmap.NetworkInfoResponse{
		Body: &protonetmap.NetworkInfoResponse_Body{
			NetworkInfo: &protonetmap.NetworkInfo{
				NetworkConfig: &protonetmap.NetworkConfig{
					Parameters: []*protonetmap.NetworkConfig_Parameter{
						{Value: []byte("any")},
					},
				},
			},
		},
	}

	var respV2 v2netmap.NetworkInfoResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protonetmap.NetworkInfoResponse), nil
}

type testGetNodeInfoServer struct {
	protonetmap.UnimplementedNetmapServiceServer
}

func (x *testGetNodeInfoServer) LocalNodeInfo(context.Context, *protonetmap.LocalNodeInfoRequest) (*protonetmap.LocalNodeInfoResponse, error) {
	resp := protonetmap.LocalNodeInfoResponse{
		Body: &protonetmap.LocalNodeInfoResponse_Body{
			Version: new(protorefs.Version),
			NodeInfo: &protonetmap.NodeInfo{
				PublicKey: []byte("any"),
				Addresses: []string{"any"},
			},
		},
	}

	var respV2 v2netmap.LocalNodeInfoResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protonetmap.LocalNodeInfoResponse), nil
}

func TestClient_NetMapSnapshot(t *testing.T) {
	var err error
	var prm PrmNetMapSnapshot
	var res netmap.NetMap
	var srv testNetmapSnapshotServer

	signer := neofscryptotest.Signer()

	srv.signer = signer

	c := newTestNetmapClient(t, &srv)
	ctx := context.Background()

	// transport error
	srv.errTransport = errors.New("any error")

	_, err = c.NetMapSnapshot(ctx, prm)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unknown, st.Code())
	require.Contains(t, st.Message(), srv.errTransport.Error())

	srv.errTransport = nil

	// unsigned response
	srv.unsignedResponse = true
	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)
	srv.unsignedResponse = false

	// failure error
	srv.statusFail = true
	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)
	require.ErrorIs(t, err, apistatus.ErrServerInternal)

	srv.statusFail = false

	// missing netmap field
	srv.unsetNetMap = true
	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)

	srv.unsetNetMap = false

	// invalid network map
	srv.netMap = &protonetmap.Netmap{
		Nodes: []*protonetmap.NodeInfo{new(protonetmap.NodeInfo)},
	}

	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)

	// correct network map
	// TODO: #260 use instance normalizer
	srv.netMap.Nodes[0].PublicKey = []byte{1, 2, 3}
	srv.netMap.Nodes[0].Addresses = []string{"1", "2", "3"}

	res, err = c.NetMapSnapshot(ctx, prm)
	require.NoError(t, err)

	require.Zero(t, res.Epoch())
	ns := res.Nodes()
	require.Len(t, ns, 1)
	node := ns[0]
	require.False(t, node.IsOnline())
	require.False(t, node.IsOffline())
	require.False(t, node.IsMaintenance())
	require.Zero(t, node.NumberOfAttributes())
	require.Equal(t, srv.netMap.Nodes[0].PublicKey, node.PublicKey())
	var es []string
	netmap.IterateNetworkEndpoints(node, func(e string) { es = append(es, e) })
	require.Equal(t, srv.netMap.Nodes[0].Addresses, es)
}
