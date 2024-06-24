package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	apiaccounting "github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// GetBalanceOptions groups optional parameters of [Client.GetBalance].
type GetBalanceOptions struct{}

// GetBalance requests current balance of the NeoFS account.
func (c *Client) GetBalance(ctx context.Context, usr user.ID, _ GetBalanceOptions) (accounting.Decimal, error) {
	var res accounting.Decimal
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodBalanceGet, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apiaccounting.BalanceRequest{
		Body: &apiaccounting.BalanceRequest_Body{OwnerId: new(refs.OwnerID)},
	}
	usr.WriteToV2(req.Body.OwnerId)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%v: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.accounting.Balance(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return res, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return res, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return res, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return res, err
	}
	if sts != nil {
		err = sts // for closure above
		return res, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldBalance = "balance"
	if resp.Body.Balance == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldBalance) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.Balance); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldBalance, err) // for closure above
		return res, err
	}
	return res, nil
}
