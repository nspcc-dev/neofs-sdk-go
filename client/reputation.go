package client

import (
	"context"

	v2reputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

// AnnounceLocalTrustPrm groups parameters of AnnounceLocalTrust operation.
type AnnounceLocalTrustPrm struct {
	epoch uint64

	trusts []reputation.Trust
}

// SetEpoch sets number of NeoFS epoch in which the trust was assessed.
// Required parameter, must not be zero.
func (x *AnnounceLocalTrustPrm) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// SetValues sets values describing trust of the client to the NeoFS network participants.
// Required parameter. Must not be empty.
//
// Must not be mutated before the end of the operation.
func (x *AnnounceLocalTrustPrm) SetValues(trusts []reputation.Trust) {
	x.trusts = trusts
}

// AnnounceLocalTrustRes groups results of AnnounceLocalTrust operation.
type AnnounceLocalTrustRes struct {
	statusRes
}

// AnnounceLocalTrust sends client's trust values to the NeoFS network participants.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see AnnounceLocalTrustPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) AnnounceLocalTrust(ctx context.Context, prm AnnounceLocalTrustPrm) (*AnnounceLocalTrustRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.epoch == 0:
		panic("zero epoch")
	case len(prm.trusts) == 0:
		panic("missing trusts")
	}

	// form request body
	reqBody := new(v2reputation.AnnounceLocalTrustRequestBody)
	reqBody.SetEpoch(prm.epoch)

	trusts := make([]*reputation.Trust, 0, len(prm.trusts))

	for i := range prm.trusts {
		trusts = append(trusts, &prm.trusts[i])
	}

	reqBody.SetTrusts(reputation.TrustsToV2(trusts))

	// form request
	var req v2reputation.AnnounceLocalTrustRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res AnnounceLocalTrustRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceLocalTrust(c.Raw(), &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}

// AnnounceIntermediateTrustPrm groups parameters of AnnounceIntermediateTrust operation.
type AnnounceIntermediateTrustPrm struct {
	epoch uint64

	iter uint32

	trustSet bool
	trust    reputation.PeerToPeerTrust
}

// SetEpoch sets number of NeoFS epoch with which client's calculation algorithm is initialized.
// Required parameter, must not be zero.
func (x *AnnounceIntermediateTrustPrm) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// SetIteration sets current sequence number of the client's calculation algorithm.
// By default, corresponds to initial (zero) iteration.
func (x *AnnounceIntermediateTrustPrm) SetIteration(iter uint32) {
	x.iter = iter
}

// SetCurrentValue sets current global trust value computed at the specified iteration
//of the client's calculation algorithm. Required parameter.
func (x *AnnounceIntermediateTrustPrm) SetCurrentValue(trust reputation.PeerToPeerTrust) {
	x.trust = trust
	x.trustSet = true
}

// AnnounceIntermediateTrustRes groups results of AnnounceIntermediateTrust operation.
type AnnounceIntermediateTrustRes struct {
	statusRes
}

// AnnounceIntermediateTrust sends global trust values calculated for the specified NeoFS network participants
// at some stage of client's calculation algorithm.
//
// Exactly one return value is non-nil. By default, server status is returned in res structure.
// Any client's internal or transport errors are returned as `error`.
// If WithNeoFSErrorParsing option has been provided, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see AnnounceIntermediateTrustPrm docs).
// Context is required and must not be nil. It is used for network communication.
//
// Return statuses:
//  - global (see Client docs).
func (c *Client) AnnounceIntermediateTrust(ctx context.Context, prm AnnounceIntermediateTrustPrm) (*AnnounceIntermediateTrustRes, error) {
	// check parameters
	switch {
	case ctx == nil:
		panic(panicMsgMissingContext)
	case prm.epoch == 0:
		panic("zero epoch")
	case !prm.trustSet:
		panic("current trust value not set")
	}

	// form request body
	reqBody := new(v2reputation.AnnounceIntermediateResultRequestBody)
	reqBody.SetEpoch(prm.epoch)
	reqBody.SetIteration(prm.iter)
	reqBody.SetTrust(prm.trust.ToV2())

	// form request
	var req v2reputation.AnnounceIntermediateResultRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc  contextCall
		res AnnounceIntermediateTrustRes
	)

	c.initCallContext(&cc)
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceIntermediateResult(c.Raw(), &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
