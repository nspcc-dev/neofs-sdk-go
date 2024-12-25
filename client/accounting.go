package client

import (
	"context"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	protoaccounting "github.com/nspcc-dev/neofs-sdk-go/proto/accounting"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
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

	req := &protoaccounting.BalanceRequest{
		Body: &protoaccounting.BalanceRequest_Body{
			OwnerId: prm.account.ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	var res accounting.Decimal

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoaccounting.BalanceRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return res, err
	}

	resp, err := c.accounting.Balance(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return res, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return res, err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protoaccounting.BalanceResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return res, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return res, err
	}

	const fieldBalance = "balance"

	bal := resp.GetBody().GetBalance()
	if bal == nil {
		err = newErrMissingResponseField(fieldBalance)
		return res, err
	}

	err = res.FromProtoMessage(bal)
	if err != nil {
		err = newErrInvalidResponseField(fieldBalance, err)
		return res, err
	}

	return res, nil
}
