package client

import (
	"context"
	"fmt"
	"time"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// NetworkInfoExecutor describes methods to get network information.
type NetworkInfoExecutor interface {
	NetworkInfo(ctx context.Context, prm PrmNetworkInfo) (netmap.NetworkInfo, error)
}

// PrmEndpointInfo groups parameters of EndpointInfo operation.
type PrmEndpointInfo struct {
	prmCommonMeta
}

// ResEndpointInfo group resulting values of EndpointInfo operation.
type ResEndpointInfo struct {
	version version.Version

	ni netmap.NodeInfo
}

// NewResEndpointInfo is a constructor for ResEndpointInfo.
func NewResEndpointInfo(version version.Version, ni netmap.NodeInfo) ResEndpointInfo {
	return ResEndpointInfo{
		version: version,
		ni:      ni,
	}
}

// LatestVersion returns latest NeoFS API protocol's version in use.
func (x ResEndpointInfo) LatestVersion() version.Version {
	return x.version
}

// NodeInfo returns information about the NeoFS node served on the remote endpoint.
func (x ResEndpointInfo) NodeInfo() netmap.NodeInfo {
	return x.ni
}

// EndpointInfo requests information about the storage node served on the remote endpoint.
//
// Method can be used as a health check to see if node is alive and responds to requests.
//
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Exactly one return value is non-nil. Server status return is returned in ResEndpointInfo.
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
func (c *Client) EndpointInfo(ctx context.Context, prm PrmEndpointInfo) (*ResEndpointInfo, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodEndpointInfo, time.Since(startTime), err)
		}()
	}

	req := &protonetmap.LocalNodeInfoRequest{
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protonetmap.LocalNodeInfoRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return nil, err
	}

	resp, err := c.netmap.LocalNodeInfo(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return nil, err
	}

	if c.prm.cbRespInfo != nil {
		err = c.prm.cbRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		})
		if err != nil {
			err = fmt.Errorf("%w: %w", errResponseCallback, err)
			return nil, err
		}
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protonetmap.LocalNodeInfoResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return nil, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return nil, err
	}

	body := resp.GetBody()

	const fieldVersion = "version"

	mv := body.GetVersion()
	if mv == nil {
		err = newErrMissingResponseField(fieldVersion)
		return nil, err
	}

	var res ResEndpointInfo

	err = res.version.FromProtoMessage(mv)
	if err != nil {
		err = newErrInvalidResponseField(fieldVersion, err)
		return nil, err
	}

	const fieldNodeInfo = "node info"

	mn := body.GetNodeInfo()
	if mn == nil {
		err = newErrMissingResponseField(fieldNodeInfo)
		return nil, err
	}

	err = res.ni.FromProtoMessage(mn)
	if err != nil {
		err = newErrInvalidResponseField(fieldNodeInfo, err)
		return nil, err
	}

	return &res, nil
}

// PrmNetworkInfo groups parameters of NetworkInfo operation.
type PrmNetworkInfo struct {
	prmCommonMeta
}

// NetworkInfo requests information about the NeoFS network of which the remote server is a part.
//
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and must not be nil. It is used for network communication.
//
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
func (c *Client) NetworkInfo(ctx context.Context, prm PrmNetworkInfo) (netmap.NetworkInfo, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodNetworkInfo, time.Since(startTime), err)
		}()
	}

	req := &protonetmap.NetworkInfoRequest{
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}
	writeXHeadersToMeta(prm.xHeaders, req.MetaHeader)

	var res netmap.NetworkInfo

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protonetmap.NetworkInfoRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return res, err
	}

	resp, err := c.netmap.NetworkInfo(ctx, req)
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

	if err = neofscrypto.VerifyResponseWithBuffer[*protonetmap.NetworkInfoResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return res, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return res, err
	}

	const fieldNetInfo = "network info"

	mn := resp.GetBody().GetNetworkInfo()
	if mn == nil {
		err = newErrMissingResponseField(fieldNetInfo)
		return res, err
	}

	err = res.FromProtoMessage(mn)
	if err != nil {
		err = newErrInvalidResponseField(fieldNetInfo, err)
		return res, err
	}

	return res, nil
}

// PrmNetMapSnapshot groups parameters of NetMapSnapshot operation.
type PrmNetMapSnapshot struct {
}

// NetMapSnapshot requests current network view of the remote server.
//
// Any client's internal or transport errors are returned as `error`,
// see [apistatus] package for NeoFS-specific error types.
//
// Context is required and MUST NOT be nil. It is used for network communication.
//
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
func (c *Client) NetMapSnapshot(ctx context.Context, _ PrmNetMapSnapshot) (netmap.NetMap, error) {
	var err error
	if c.prm.statisticCallback != nil {
		startTime := time.Now()
		defer func() {
			c.sendStatistic(stat.MethodNetMapSnapshot, time.Since(startTime), err)
		}()
	}

	req := &protonetmap.NetmapSnapshotRequest{
		MetaHeader: &protosession.RequestMetaHeader{
			Version: version.Current().ProtoMessage(),
			Ttl:     defaultRequestTTL,
		},
	}

	var res netmap.NetMap

	buf := c.buffers.Get().(*[]byte)
	defer func() { c.buffers.Put(buf) }()

	req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protonetmap.NetmapSnapshotRequest_Body](c.prm.signer, req, *buf)
	if err != nil {
		err = fmt.Errorf("%w: %w", errSignRequest, err)
		return res, err
	}

	resp, err := c.netmap.NetmapSnapshot(ctx, req)
	if err != nil {
		err = rpcErr(err)
		return netmap.NetMap{}, err
	}

	if err = neofscrypto.VerifyResponseWithBuffer[*protonetmap.NetmapSnapshotResponse_Body](resp, *buf); err != nil {
		err = fmt.Errorf("%w: %w", errResponseSignatures, err)
		return res, err
	}

	if err = apistatus.ToError(resp.GetMetaHeader().GetStatus()); err != nil {
		return res, err
	}

	const fieldNetMap = "network map"

	mn := resp.GetBody().GetNetmap()
	if mn == nil {
		err = newErrMissingResponseField(fieldNetMap)
		return netmap.NetMap{}, err
	}

	err = res.FromProtoMessage(mn)
	if err != nil {
		err = newErrInvalidResponseField(fieldNetMap, err)
		return netmap.NetMap{}, err
	}

	return res, nil
}
