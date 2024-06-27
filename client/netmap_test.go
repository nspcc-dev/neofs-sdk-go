package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

type serverNetMap struct {
	errTransport error

	signResponse bool

	statusOK bool

	setNetMap bool
	netMap    v2netmap.NetMap

	signer neofscrypto.Signer
}

func (x *serverNetMap) createSession(*client.Client, *session.CreateRequest, ...client.CallOption) (*session.CreateResponse, error) {
	return nil, nil
}

func (x *serverNetMap) netMapSnapshot(_ context.Context, req v2netmap.SnapshotRequest) (*v2netmap.SnapshotResponse, error) {
	err := verifyServiceMessage(&req)
	if err != nil {
		return nil, err
	}

	if x.errTransport != nil {
		return nil, x.errTransport
	}

	var body v2netmap.SnapshotResponseBody

	if x.setNetMap {
		body.SetNetMap(&x.netMap)
	}

	var meta session.ResponseMetaHeader

	if !x.statusOK {
		meta.SetStatus(statusErr.ErrorToV2())
	}

	var resp v2netmap.SnapshotResponse
	resp.SetBody(&body)
	resp.SetMetaHeader(&meta)

	if x.signResponse {
		err = signServiceMessage(x.signer, &resp, nil)
		if err != nil {
			panic(fmt.Sprintf("sign response: %v", err))
		}
	}

	return &resp, nil
}

func TestClient_NetMapSnapshot(t *testing.T) {
	var err error
	var prm PrmNetMapSnapshot
	var res netmap.NetMap
	var srv serverNetMap

	signer := test.RandomSignerRFC6979()

	srv.signer = signer

	c := newClient(t, &srv)
	ctx := context.Background()

	// request signature
	srv.errTransport = errors.New("any error")

	_, err = c.NetMapSnapshot(ctx, prm)
	require.ErrorIs(t, err, srv.errTransport)

	srv.errTransport = nil

	// unsigned response
	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)

	srv.signResponse = true

	// failure error
	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)
	require.ErrorIs(t, err, apistatus.ErrServerInternal)

	srv.statusOK = true

	// missing netmap field
	_, err = c.NetMapSnapshot(ctx, prm)
	require.Error(t, err)

	srv.setNetMap = true

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
