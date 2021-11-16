package client

import (
	"context"
	"fmt"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// Accounting contains methods related to balance querying.
type Accounting interface {
	// GetBalance returns balance of provided account.
	GetBalance(context.Context, *owner.ID, ...CallOption) (*BalanceOfRes, error)
}

type BalanceOfRes struct {
	statusRes

	amount *accounting.Decimal
}

func (x *BalanceOfRes) setAmount(v *accounting.Decimal) {
	x.amount = v
}

func (x BalanceOfRes) Amount() *accounting.Decimal {
	return x.amount
}

func (c *clientImpl) GetBalance(ctx context.Context, owner *owner.ID, opts ...CallOption) (*BalanceOfRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2accounting.BalanceRequestBody)
	reqBody.SetOwnerID(owner.ToV2())

	req := new(v2accounting.BalanceRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.Balance(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(BalanceOfRes)
		procPrm processResponseV2Prm
		procRes processResponseV2Res
	)

	procPrm.callOpts = callOptions
	procPrm.resp = resp

	procRes.statusRes = res

	// process response in general
	if c.processResponseV2(&procRes, procPrm) {
		if procRes.cliErr != nil {
			return nil, procRes.cliErr
		}

		return res, nil
	}

	res.setAmount(accounting.NewDecimalFromV2(resp.GetBody().GetBalance()))

	return res, nil
}
