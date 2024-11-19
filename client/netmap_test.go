package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

type testNetmapSnapshotServer struct {
	unimplementedNeoFSAPIServer

	errTransport error

	unsignedResponse bool

	statusFail bool

	unsetNetMap bool
	netMap      v2netmap.NetMap

	signer neofscrypto.Signer
}

func (x *testNetmapSnapshotServer) netMapSnapshot(_ context.Context, req v2netmap.SnapshotRequest) (*v2netmap.SnapshotResponse, error) {
	err := verifyServiceMessage(&req)
	if err != nil {
		return nil, err
	}

	if x.errTransport != nil {
		return nil, x.errTransport
	}

	var body v2netmap.SnapshotResponseBody

	if !x.unsetNetMap {
		body.SetNetMap(&x.netMap)
	}

	var meta session.ResponseMetaHeader

	if x.statusFail {
		meta.SetStatus(statusErr.ErrorToV2())
	}

	var resp v2netmap.SnapshotResponse
	resp.SetBody(&body)
	resp.SetMetaHeader(&meta)

	signer := x.signer
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	if !x.unsignedResponse {
		err = signServiceMessage(signer, &resp, nil)
		if err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}
	}

	return &resp, nil
}

type testGetNetworkInfoServer struct {
	unimplementedNeoFSAPIServer
}

func (x testGetNetworkInfoServer) getNetworkInfo(context.Context, v2netmap.NetworkInfoRequest) (*v2netmap.NetworkInfoResponse, error) {
	var netPrm v2netmap.NetworkParameter
	netPrm.SetValue([]byte("any"))
	var netCfg v2netmap.NetworkConfig
	netCfg.SetParameters(netPrm)
	var netInfo v2netmap.NetworkInfo
	netInfo.SetNetworkConfig(&netCfg)
	var body v2netmap.NetworkInfoResponseBody
	body.SetNetworkInfo(&netInfo)
	var resp v2netmap.NetworkInfoResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testGetNodeInfoServer struct {
	unimplementedNeoFSAPIServer
}

func (x testGetNodeInfoServer) getNodeInfo(context.Context, v2netmap.LocalNodeInfoRequest) (*v2netmap.LocalNodeInfoResponse, error) {
	var nodeInfo v2netmap.NodeInfo
	nodeInfo.SetPublicKey([]byte("any"))
	nodeInfo.SetAddresses("any")
	var body v2netmap.LocalNodeInfoResponseBody
	body.SetVersion(new(refs.Version))
	body.SetNodeInfo(&nodeInfo)
	var resp v2netmap.LocalNodeInfoResponse
	resp.SetBody(&body)

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

func TestClient_NetMapSnapshot(t *testing.T) {
	var err error
	var prm PrmNetMapSnapshot
	var res netmap.NetMap
	var srv testNetmapSnapshotServer

	signer := neofscryptotest.Signer()

	srv.signer = signer

	c := newClient(t, &srv)
	ctx := context.Background()

	// request signature
	srv.errTransport = errors.New("any error")

	_, err = c.NetMapSnapshot(ctx, prm)
	require.ErrorIs(t, err, srv.errTransport)

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
	var netMap netmap.NetMap

	var node netmap.NodeInfo
	// TODO: #260 use instance corrupter

	var nodeV2 v2netmap.NodeInfo

	node.WriteToV2(&nodeV2)
	require.Error(t, new(netmap.NodeInfo).ReadFromV2(nodeV2))

	netMap.SetNodes([]netmap.NodeInfo{node})
	netMap.WriteToV2(&srv.netMap)

	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)

	// correct network map
	// TODO: #260 use instance normalizer
	node.SetPublicKey([]byte{1, 2, 3})
	node.SetNetworkEndpoints("1", "2", "3")

	node.WriteToV2(&nodeV2)
	require.NoError(t, new(netmap.NodeInfo).ReadFromV2(nodeV2))

	netMap.SetNodes([]netmap.NodeInfo{node})
	netMap.WriteToV2(&srv.netMap)

	res, err = c.NetMapSnapshot(ctx, prm)
	require.NoError(t, err)
	require.Equal(t, netMap, res)
}
