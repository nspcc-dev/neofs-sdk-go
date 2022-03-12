package client

import (
	"context"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// PrmBalanceGet groups parameters of BalanceGet operation.
type PrmBalanceGet struct {
	prmCommonMeta

	ownerSet bool
	ownerID  owner.ID
}

// SetAccount sets identifier of the NeoFS account for which the balance is requested.
// Required parameter. Must be a valid ID according to NeoFS API protocol.
func (x *PrmBalanceGet) SetAccount(id owner.ID) {
	x.ownerID = id
	x.ownerSet = true
}

// ResBalanceGet groups resulting values of BalanceGet operation.
type ResBalanceGet struct {
	statusRes

	amount *accounting.Decimal
}

func (x *ResBalanceGet) setAmount(v *accounting.Decimal) {
	x.amount = v
}

// Amount returns current amount of funds on the NeoFS account as decimal number.
//
// Client doesn't retain value so modification is safe.
func (x ResBalanceGet) Amount() *accounting.Decimal {
	return x.amount
}

// BalanceGet requests current balance of the NeoFS account.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`,
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmBalanceGet docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) BalanceGet(ctx context.Context, prm PrmBalanceGet) (*ResBalanceGet, error) {
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case !prm.ownerSet:
		panic("account not set")
	case !prm.ownerID.Valid():
		panic("invalid account ID")
	}

	// form request body
	var body v2accounting.BalanceRequestBody

	body.SetOwnerID(prm.ownerID.ToV2())

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
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.Balance(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2accounting.BalanceResponse)

		if bal := resp.GetBody().GetBalance(); bal != nil {
			var d accounting.Decimal
			d.ReadFromMessageV2(*bal)

			res.setAmount(&d)
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
