package client

import (
	"context"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

// GetBalancePrm groups parameters of GetBalance operation.
type GetBalancePrm struct {
	ownerSet bool
	ownerID  owner.ID
}

// SetAccount sets identifier of the NeoFS account for which the balance is requested.
// Required parameter. Must be a valid ID according to NeoFS API protocol.
func (x *GetBalancePrm) SetAccount(id owner.ID) {
	x.ownerID = id
	x.ownerSet = true
}

// GetBalanceRes groups resulting values of GetBalance operation.
type GetBalanceRes struct {
	statusRes

	amount *accounting.Decimal
}

func (x *GetBalanceRes) setAmount(v *accounting.Decimal) {
	x.amount = v
}

// Amount returns current amount of funds on the NeoFS account as decimal number.
//
// Client doesn't retain value so modification is safe.
func (x GetBalanceRes) Amount() *accounting.Decimal {
	return x.amount
}

// GetBalance requests current balance of the NeoFS account.
//
// Any client's internal or transport errors are returned as `error`,
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see GetBalancePrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Exactly one return value is non-nil. Server status return is returned in GetBalanceRes.
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
func (c *Client) GetBalance(ctx context.Context, prm GetBalancePrm) (*GetBalanceRes, error) {
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
		res GetBalanceRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.Balance(c.Raw(), &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2accounting.BalanceResponse)
		res.setAmount(accounting.NewDecimalFromV2(resp.GetBody().GetBalance()))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
