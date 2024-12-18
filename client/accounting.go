package client

import (
	"context"
	"time"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	protoaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmBalanceGet groups parameters of BalanceGet operation.
type PrmBalanceGet struct {
	prmCommonMeta

	account user.ID
}

// SetAccount sets identifier of the NeoFS account for which the balance is requested.
// Required parameter.
func (x *PrmBalanceGet) SetAccount(id user.ID) {
	x.account = id
}

// BalanceGet requests current balance of the NeoFS account.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrMissingAccount]
func (c *Client) BalanceGet(ctx context.Context, prm PrmBalanceGet) (accounting.Decimal, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodBalanceGet, time.Since(startTime), err)
		}()
	}

	switch {
	case prm.account.IsZero():
		err = ErrMissingAccount
		return accounting.Decimal{}, err
	}

	// form request body
	var accountV2 refs.OwnerID
	prm.account.WriteToV2(&accountV2)

	var body v2accounting.BalanceRequestBody
	body.SetOwnerID(&accountV2)

	// form request
	var req v2accounting.BalanceRequest

	req.SetBody(&body)

	// init call context

	var (
		cc  contextCall
		res accounting.Decimal
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		resp, err := c.accounting.Balance(ctx, req.ToGRPCMessage().(*protoaccounting.BalanceRequest))
		if err != nil {
			return nil, rpcErr(err)
		}
		var respV2 v2accounting.BalanceResponse
		if err = respV2.FromGRPCMessage(resp); err != nil {
			return nil, err
		}
		return &respV2, nil
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2accounting.BalanceResponse)

		const fieldBalance = "balance"

		bal := resp.GetBody().GetBalance()
		if bal == nil {
			cc.err = newErrMissingResponseField(fieldBalance)
			return
		}

		cc.err = res.ReadFromV2(*bal)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldBalance, cc.err)
		}
	}

	// process call
	if !cc.processCall() {
		err = cc.err
		return accounting.Decimal{}, cc.err
	}

	return res, nil
}
