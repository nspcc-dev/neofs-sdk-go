package client

import (
	"context"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmBalanceGet groups parameters of BalanceGet operation.
type PrmBalanceGet struct {
	prmCommonMeta

	accountSet bool
	account    user.ID
}

// SetAccount sets identifier of the NeoFS account for which the balance is requested.
// Required parameter.
func (x *PrmBalanceGet) SetAccount(id user.ID) {
	x.account = id
	x.accountSet = true
}

// ResBalanceGet groups resulting values of BalanceGet operation.
type ResBalanceGet struct {
	amount accounting.Decimal
}

// Amount returns current amount of funds on the NeoFS account as decimal number.
func (x ResBalanceGet) Amount() accounting.Decimal {
	return x.amount
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
func (c *Client) BalanceGet(ctx context.Context, prm PrmBalanceGet) (*ResBalanceGet, error) {
	switch {
	case !prm.accountSet:
		return nil, ErrMissingAccount
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
		res ResBalanceGet
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.Balance(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2accounting.BalanceResponse)

		const fieldBalance = "balance"

		bal := resp.GetBody().GetBalance()
		if bal == nil {
			cc.err = newErrMissingResponseField(fieldBalance)
			return
		}

		cc.err = res.amount.ReadFromV2(*bal)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldBalance, cc.err)
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
