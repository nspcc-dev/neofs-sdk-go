package pool

import (
	"context"

	"github.com/nspcc-dev/neofs-sdk-go/client"
)

// SessionCreate opens a session with the node server on the remote endpoint.
//
// See details in [client.Client.SessionCreate].
func (p *Pool) SessionCreate(ctx context.Context, prm client.PrmSessionCreate) (*client.ResSessionCreate, error) {
	c, err := p.sdkClient()
	if err != nil {
		return nil, err
	}

	return c.SessionCreate(ctx, prm)
}
