package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	apireputation "github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

// SendLocalTrustsOptions groups optional parameters of [Client.SendLocalTrusts]
// operation.
type SendLocalTrustsOptions struct{}

// SendLocalTrusts sends client's trust values to the NeoFS network participants
// collected for a given epoch. The trust set must not be empty.
//
// SendLocalTrusts is used for system needs and is not intended to be called by
// regular users.
func (c *Client) SendLocalTrusts(ctx context.Context, epoch uint64, trusts []reputation.Trust, _ SendLocalTrustsOptions) error {
	if len(trusts) == 0 {
		return errors.New("missing trusts")
	}

	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodAnnounceLocalTrust, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apireputation.AnnounceLocalTrustRequest{
		Body: &apireputation.AnnounceLocalTrustRequest_Body{
			Epoch:  epoch,
			Trusts: make([]*apireputation.Trust, len(trusts)),
		},
	}
	for i := range trusts {
		req.Body.Trusts[i] = new(apireputation.Trust)
		trusts[i].WriteToV2(req.Body.Trusts[i])
	}
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return err
	}

	// send request
	resp, err := c.transport.reputation.AnnounceLocalTrust(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
	} else if sts != nil {
		err = sts // for closure above
	}
	return err
}

// SendIntermediateTrustOptions groups optional parameters of [Client.SendIntermediateTrust]
// operation.
type SendIntermediateTrustOptions struct{}

// SendIntermediateTrust sends global trust value calculated for the specified
// NeoFS network participant at given stage of client's iteration algorithm.
//
// SendIntermediateTrust is used for system needs and is not intended to be
// called by regular users.
func (c *Client) SendIntermediateTrust(ctx context.Context, epoch uint64, iter uint32, trust reputation.PeerToPeerTrust, _ SendIntermediateTrustOptions) error {
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodAnnounceIntermediateTrust, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := &apireputation.AnnounceIntermediateResultRequest{
		Body: &apireputation.AnnounceIntermediateResultRequest_Body{
			Epoch:     epoch,
			Iteration: iter,
			Trust:     new(apireputation.PeerToPeerTrust),
		},
	}
	trust.WriteToV2(req.Body.Trust)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return err
	}

	// send request
	resp, err := c.transport.reputation.AnnounceIntermediateResult(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
	} else if sts != nil {
		err = sts // for closure above
	}
	return err
}
