package pool

import (
	"context"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// NetworkInfo requests information about the NeoFS network of which the remote server is a part.
//
// See details in [client.Client.NetworkInfo].
func (p *Pool) NetworkInfo(ctx context.Context, prm client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return netmap.NetworkInfo{}, err
	}

	start := time.Now()
	info, err := c.NetworkInfo(ctx, prm)
	statUpdater.incRequests(time.Since(start), methodNetworkInfo)
	statUpdater.updateErrorRate(err)

	return info, err
}

// NetMapSnapshot requests current network view of the remote server.
//
// See details in [client.Client.NetMapSnapshot].
func (p *Pool) NetMapSnapshot(ctx context.Context, prm client.PrmNetMapSnapshot) (netmap.NetMap, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return netmap.NetMap{}, err
	}

	start := time.Now()
	netMap, err := c.NetMapSnapshot(ctx, prm)
	statUpdater.incRequests(time.Since(start), methodNetMapSnapshot)
	statUpdater.updateErrorRate(err)

	return netMap, err
}
