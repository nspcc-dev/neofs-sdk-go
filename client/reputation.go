package client

import (
	"context"

	v2reputation "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	protoreputation "github.com/nspcc-dev/neofs-api-go/v2/reputation/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

// PrmAnnounceLocalTrust groups optional parameters of AnnounceLocalTrust operation.
type PrmAnnounceLocalTrust struct {
	prmCommonMeta
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
//
// Parameter epoch must not be zero.
// Parameter trusts must not be empty.
func (c *Client) AnnounceLocalTrust(ctx context.Context, epoch uint64, trusts []reputation.Trust, prm PrmAnnounceLocalTrust) error {
	var err error
	defer func() {
		c.sendStatistic(stat.MethodAnnounceLocalTrust, err)()
	}()

	// check parameters
	switch {
	case epoch == 0:
		err = ErrZeroEpoch
		return err
	case len(trusts) == 0:
		err = ErrMissingTrusts
		return err
	}

	// form request body
	reqBody := new(v2reputation.AnnounceLocalTrustRequestBody)
	reqBody.SetEpoch(epoch)

	trustList := make([]v2reputation.Trust, len(trusts))

	for i := range trusts {
		trusts[i].WriteToV2(&trustList[i])
	}

	reqBody.SetTrusts(trustList)

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
		resp, err := c.reputation.AnnounceLocalTrust(ctx, req.ToGRPCMessage().(*protoreputation.AnnounceLocalTrustRequest))
		if err != nil {
			return nil, rpcErr(err)
		}
		var respV2 v2reputation.AnnounceLocalTrustResponse
		if err = respV2.FromGRPCMessage(resp); err != nil {
			return nil, err
		}
		return &respV2, nil
	}

	// process call
	if !cc.processCall() {
		err = cc.err
		return err
	}

	return nil
}

// PrmAnnounceIntermediateTrust groups optional parameters of AnnounceIntermediateTrust operation.
type PrmAnnounceIntermediateTrust struct {
	prmCommonMeta

	iter uint32
}

// SetIteration sets current sequence number of the client's calculation algorithm.
// By default, corresponds to initial (zero) iteration.
func (x *PrmAnnounceIntermediateTrust) SetIteration(iter uint32) {
	x.iter = iter
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
//
// Parameter epoch must not be zero.
func (c *Client) AnnounceIntermediateTrust(ctx context.Context, epoch uint64, trust reputation.PeerToPeerTrust, prm PrmAnnounceIntermediateTrust) error {
	var err error
	defer func() {
		c.sendStatistic(stat.MethodAnnounceIntermediateTrust, err)()
	}()

	if epoch == 0 {
		err = ErrZeroEpoch
		return err
	}

	var v2Trust v2reputation.PeerToPeerTrust
	trust.WriteToV2(&v2Trust)

	// form request body
	reqBody := new(v2reputation.AnnounceIntermediateResultRequestBody)
	reqBody.SetEpoch(epoch)
	reqBody.SetIteration(prm.iter)
	reqBody.SetTrust(&v2Trust)

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
		resp, err := c.reputation.AnnounceIntermediateResult(ctx, req.ToGRPCMessage().(*protoreputation.AnnounceIntermediateResultRequest))
		if err != nil {
			return nil, rpcErr(err)
		}
		var respV2 v2reputation.AnnounceIntermediateResultResponse
		if err = respV2.FromGRPCMessage(resp); err != nil {
			return nil, err
		}
		return &respV2, nil
	}

	// process call
	if !cc.processCall() {
		err = cc.err
		return err
	}

	return nil
}
