package client

import (
	"context"

	v2reputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
)

// PrmAnnounceLocalTrust groups parameters of AnnounceLocalTrust operation.
type PrmAnnounceLocalTrust struct {
	prmCommonMeta

	epoch uint64

	trusts []reputation.Trust
}

// SetEpoch sets number of NeoFS epoch in which the trust was assessed.
// Required parameter, must not be zero.
func (x *PrmAnnounceLocalTrust) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// SetValues sets values describing trust of the client to the NeoFS network participants.
// Required parameter. Must not be empty.
//
// Must not be mutated before the end of the operation.
func (x *PrmAnnounceLocalTrust) SetValues(trusts []reputation.Trust) {
	x.trusts = trusts
}

// AnnounceLocalTrust sends client's trust values to the NeoFS network participants.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrZeroEpoch]
//   - [ErrMissingTrusts]
//   - [ErrMissingSigner]
func (c *Client) AnnounceLocalTrust(ctx context.Context, prm PrmAnnounceLocalTrust) error {
	// check parameters
	switch {
	case prm.epoch == 0:
		return ErrZeroEpoch
	case len(prm.trusts) == 0:
		return ErrMissingTrusts
	}

	if c.prm.signer == nil {
		return ErrMissingSigner
	}

	// form request body
	reqBody := new(v2reputation.AnnounceLocalTrustRequestBody)
	reqBody.SetEpoch(prm.epoch)

	trusts := make([]v2reputation.Trust, len(prm.trusts))

	for i := range prm.trusts {
		prm.trusts[i].WriteToV2(&trusts[i])
	}

	reqBody.SetTrusts(trusts)

	// form request
	var req v2reputation.AnnounceLocalTrustRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc contextCall
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceLocalTrust(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return cc.err
	}

	return nil
}

// PrmAnnounceIntermediateTrust groups parameters of AnnounceIntermediateTrust operation.
type PrmAnnounceIntermediateTrust struct {
	prmCommonMeta

	epoch uint64

	iter uint32

	trustSet bool
	trust    reputation.PeerToPeerTrust
}

// SetEpoch sets number of NeoFS epoch with which client's calculation algorithm is initialized.
// Required parameter, must not be zero.
func (x *PrmAnnounceIntermediateTrust) SetEpoch(epoch uint64) {
	x.epoch = epoch
}

// SetIteration sets current sequence number of the client's calculation algorithm.
// By default, corresponds to initial (zero) iteration.
func (x *PrmAnnounceIntermediateTrust) SetIteration(iter uint32) {
	x.iter = iter
}

// SetCurrentValue sets current global trust value computed at the specified iteration
// of the client's calculation algorithm. Required parameter.
func (x *PrmAnnounceIntermediateTrust) SetCurrentValue(trust reputation.PeerToPeerTrust) {
	x.trust = trust
	x.trustSet = true
}

// AnnounceIntermediateTrust sends global trust values calculated for the specified NeoFS network participants
// at some stage of client's calculation algorithm.
//
// Any errors (local or remote, including returned status codes) are returned as Go errors,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Return errors:
//   - [ErrZeroEpoch]
//   - [ErrMissingTrust]
//   - [ErrMissingSigner]
func (c *Client) AnnounceIntermediateTrust(ctx context.Context, prm PrmAnnounceIntermediateTrust) error {
	// check parameters
	switch {
	case prm.epoch == 0:
		return ErrZeroEpoch
	case !prm.trustSet:
		return ErrMissingTrust
	}

	if c.prm.signer == nil {
		return ErrMissingSigner
	}

	var trust v2reputation.PeerToPeerTrust
	prm.trust.WriteToV2(&trust)

	// form request body
	reqBody := new(v2reputation.AnnounceIntermediateResultRequestBody)
	reqBody.SetEpoch(prm.epoch)
	reqBody.SetIteration(prm.iter)
	reqBody.SetTrust(&trust)

	// form request
	var req v2reputation.AnnounceIntermediateResultRequest

	req.SetBody(reqBody)

	// init call context

	var (
		cc contextCall
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcapi.AnnounceIntermediateResult(&c.c, &req, client.WithContext(ctx))
	}

	// process call
	if !cc.processCall() {
		return cc.err
	}

	return nil
}
