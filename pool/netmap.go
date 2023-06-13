package pool

import (
	"context"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// NetworkInfo requests information about the NeoFS network of which the remote server is a part.
//
// See details in [client.Client.EndpointInfo].
func (p *Pool) NetworkInfo(ctx context.Context, prm client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	c, err := p.sdkClient()
	if err != nil {
		return netmap.NetworkInfo{}, err
	}

	return c.NetworkInfo(ctx, prm)
}

// NetMapSnapshot requests current network view of the remote server.
//
// See details in [client.Client.NetMapSnapshot].
func (p *Pool) NetMapSnapshot(ctx context.Context, prm client.PrmNetMapSnapshot) (netmap.NetMap, error) {
	c, err := p.sdkClient()
	if err != nil {
		return netmap.NetMap{}, err
	}

	return c.NetMapSnapshot(ctx, prm)
}
