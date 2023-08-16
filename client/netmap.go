package client

import (
	"context"
	"fmt"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

var (
	// special variables for test purposes only, to overwrite real RPC calls.
	rpcAPINetworkInfo   = rpcapi.NetworkInfo
	rpcAPILocalNodeInfo = rpcapi.LocalNodeInfo
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
	defer func() {
		c.sendStatistic(stat.MethodEndpointInfo, err)()
	}()

	// form request
	var req v2netmap.LocalNodeInfoRequest

	// init call context

	var (
		cc  contextCall
		res ResEndpointInfo
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcAPILocalNodeInfo(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2netmap.LocalNodeInfoResponse)

		body := resp.GetBody()

		const fieldVersion = "version"

		verV2 := body.GetVersion()
		if verV2 == nil {
			cc.err = newErrMissingResponseField(fieldVersion)
			return
		}

		cc.err = res.version.ReadFromV2(*verV2)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldVersion, cc.err)
			return
		}

		const fieldNodeInfo = "node info"

		nodeInfoV2 := body.GetNodeInfo()
		if nodeInfoV2 == nil {
			cc.err = newErrMissingResponseField(fieldNodeInfo)
			return
		}

		cc.err = res.ni.ReadFromV2(*nodeInfoV2)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldNodeInfo, cc.err)
			return
		}
	}

	// process call
	if !cc.processCall() {
		err = cc.err
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
	defer func() {
		c.sendStatistic(stat.MethodNetworkInfo, err)()
	}()

	// form request
	var req v2netmap.NetworkInfoRequest

	// init call context

	var (
		cc  contextCall
		res netmap.NetworkInfo
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.call = func() (responseV2, error) {
		return rpcAPINetworkInfo(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2netmap.NetworkInfoResponse)

		const fieldNetInfo = "network info"

		netInfoV2 := resp.GetBody().GetNetworkInfo()
		if netInfoV2 == nil {
			cc.err = newErrMissingResponseField(fieldNetInfo)
			return
		}

		cc.err = res.ReadFromV2(*netInfoV2)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldNetInfo, cc.err)
			return
		}
	}

	// process call
	if !cc.processCall() {
		err = cc.err
		return netmap.NetworkInfo{}, cc.err
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
	defer func() {
		c.sendStatistic(stat.MethodNetMapSnapshot, err)()
	}()

	// form request body
	var body v2netmap.SnapshotRequestBody

	// form meta header
	var meta v2session.RequestMetaHeader

	// form request
	var req v2netmap.SnapshotRequest
	req.SetBody(&body)
	c.prepareRequest(&req, &meta)

	err = signServiceMessage(c.prm.signer, &req)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return netmap.NetMap{}, err
	}

	var resp *v2netmap.SnapshotResponse

	resp, err = c.server.netMapSnapshot(ctx, req)
	if err != nil {
		return netmap.NetMap{}, err
	}

	var res netmap.NetMap
	if err = c.processResponse(resp); err != nil {
		return netmap.NetMap{}, err
	}

	const fieldNetMap = "network map"

	netMapV2 := resp.GetBody().NetMap()
	if netMapV2 == nil {
		err = newErrMissingResponseField(fieldNetMap)
		return netmap.NetMap{}, err
	}

	err = res.ReadFromV2(*netMapV2)
	if err != nil {
		err = newErrInvalidResponseField(fieldNetMap, err)
		return netmap.NetMap{}, err
	}

	return res, nil
}
