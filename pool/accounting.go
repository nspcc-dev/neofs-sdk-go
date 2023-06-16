package pool

import (
	"context"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/client"
)

// BalanceGet requests current balance of the NeoFS account.
//
// See details in [client.Client.BalanceGet].
func (p *Pool) BalanceGet(ctx context.Context, prm client.PrmBalanceGet) (accounting.Decimal, error) {
	c, statUpdater, err := p.sdkClient()
	if err != nil {
		return accounting.Decimal{}, err
	}

	start := time.Now()
	acc, err := c.BalanceGet(ctx, prm)
	statUpdater.incRequests(time.Since(start), methodBalanceGet)
	statUpdater.updateErrorRate(err)

	return acc, err
}
