package client

import (
	"context"
	"fmt"
	"time"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	protoreputation "github.com/nspcc-dev/neofs-sdk-go/proto/reputation"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
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
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodAnnounceLocalTrust, time.Since(startTime), err)
		}()
	}

	// check parameters
	switch {
	case epoch == 0:
		err = ErrZeroEpoch
		return err
	case len(trusts) == 0:
		err = ErrMissingTrusts
		return err
	}

	req := &protoreputation.AnnounceLocalTrustRequest{
		Body: &protoreputation.AnnounceLocalTrustRequest_Body{
			Epoch:  epoch,
			Trusts: make([]*protoreputation.Trust, len(trusts)),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	for i := range trusts {
		req.Body.Trusts[i] = trusts[i].ProtoMessage()
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoreputation.AnnounceLocalTrustRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	resp, err := c.reputation.AnnounceLocalTrust(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   c.nodeKey,
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return err
		}
	}

	err = apistatus.ToError(resp.GetMetaHeader().GetStatus())
	return err
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
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodAnnounceIntermediateTrust, time.Since(startTime), err)
		}()
	}

	if epoch == 0 {
		err = ErrZeroEpoch
		return err
	}

	req := &protoreputation.AnnounceIntermediateResultRequest{
		Body: &protoreputation.AnnounceIntermediateResultRequest_Body{
			Epoch:     epoch,
			Iteration: prm.iter,
			Trust:     trust.ProtoMessage(),
		},
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoreputation.AnnounceIntermediateResultRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return err
	}

	resp, err := c.reputation.AnnounceIntermediateResult(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   c.nodeKey,
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return err
		}
	}

	err = apistatus.ToError(resp.GetMetaHeader().GetStatus())
	return err
}
