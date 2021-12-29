package client

import (
	"context"

	v2reputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

// AnnounceLocalTrustPrm groups parameters of AnnounceLocalTrust operation.
type AnnounceLocalTrustPrm struct {
	epoch uint64

	trusts []*reputation.Trust
}

// Epoch returns epoch in which the trust was assessed.
func (x AnnounceLocalTrustPrm) Epoch() uint64 {
	return x.epoch
}

// SetEpoch sets epoch in which the trust was assessed.
func (x *AnnounceLocalTrustPrm) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// Trusts returns list of local trust values.
func (x AnnounceLocalTrustPrm) Trusts() []*reputation.Trust {
	return x.trusts
}

// SetTrusts sets list of local trust values.
func (x *AnnounceLocalTrustPrm) SetTrusts(trusts []*reputation.Trust) {
	x.trusts = trusts
}

// AnnounceLocalTrustRes groups results of AnnounceLocalTrust operation.
type AnnounceLocalTrustRes struct {
	statusRes
}

// AnnounceLocalTrust announces node's local trusts through NeoFS API call.
//
// Any client's internal or transport errors are returned as error,
// NeoFS status codes are included in the returned results.
func (c *Client) AnnounceLocalTrust(ctx context.Context, prm AnnounceLocalTrustPrm, opts ...CallOption) (*AnnounceLocalTrustRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2reputation.AnnounceLocalTrustRequestBody)
	reqBody.SetEpoch(prm.Epoch())
	reqBody.SetTrusts(reputation.TrustsToV2(prm.Trusts()))

	req := new(v2reputation.AnnounceLocalTrustRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.AnnounceLocalTrust(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var (
		res     = new(AnnounceLocalTrustRes)
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

	return res, nil
}

// AnnounceIntermediateTrustPrm groups parameters of AnnounceIntermediateTrust operation.
type AnnounceIntermediateTrustPrm struct {
	epoch uint64

	iter uint32

	trust *reputation.PeerToPeerTrust
}

func (x *AnnounceIntermediateTrustPrm) Epoch() uint64 {
	return x.epoch
}

func (x *AnnounceIntermediateTrustPrm) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// Iteration returns sequence number of the iteration.
func (x AnnounceIntermediateTrustPrm) Iteration() uint32 {
	return x.iter
}

// SetIteration sets sequence number of the iteration.
func (x *AnnounceIntermediateTrustPrm) SetIteration(iter uint32) {
	x.iter = iter
}

// Trust returns current global trust value computed at the specified iteration.
func (x AnnounceIntermediateTrustPrm) Trust() *reputation.PeerToPeerTrust {
	return x.trust
}

// SetTrust sets current global trust value computed at the specified iteration.
func (x *AnnounceIntermediateTrustPrm) SetTrust(trust *reputation.PeerToPeerTrust) {
	x.trust = trust
}

// AnnounceIntermediateTrustRes groups results of AnnounceIntermediateTrust operation.
type AnnounceIntermediateTrustRes struct {
	statusRes
}

// AnnounceIntermediateTrust announces node's intermediate trusts through NeoFS API call.
//
// Any client's internal or transport errors are returned as error,
// NeoFS status codes are included in the returned results.
func (c *Client) AnnounceIntermediateTrust(ctx context.Context, prm AnnounceIntermediateTrustPrm, opts ...CallOption) (*AnnounceIntermediateTrustRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2reputation.AnnounceIntermediateResultRequestBody)
	reqBody.SetEpoch(prm.Epoch())
	reqBody.SetIteration(prm.Iteration())
	reqBody.SetTrust(prm.Trust().ToV2())

	req := new(v2reputation.AnnounceIntermediateResultRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.AnnounceIntermediateResult(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	var (
		res     = new(AnnounceIntermediateTrustRes)
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

	return res, nil
}
