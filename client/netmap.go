package client

import (
	"context"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// PrmEndpointInfo groups parameters of EndpointInfo operation.
type PrmEndpointInfo struct {
	prmCommonMeta
}

// ResEndpointInfo group resulting values of EndpointInfo operation.
type ResEndpointInfo struct {
	statusRes

	version version.Version

	ni netmap.NodeInfo
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
// Any client's internal or transport errors are returned as `error`.
// If PrmInit.ResolveNeoFSFailures has been called, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmEndpointInfo docs).
// Context is required and must not be nil. It is used for network communication.
//
// Exactly one return value is non-nil. Server status return is returned in ResEndpointInfo.
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
//
// Return statuses:
//   - global (see Client docs).
func (c *Client) EndpointInfo(ctx context.Context, prm PrmEndpointInfo) (*ResEndpointInfo, error) {
	// check context
	if ctx == nil {
		panic(panicMsgMissingContext)
	}

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
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.LocalNodeInfo(&c.c, &req, client.WithContext(ctx))
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
		return nil, cc.err
	}

	return &res, nil
}

// PrmNetworkInfo groups parameters of NetworkInfo operation.
type PrmNetworkInfo struct {
	prmCommonMeta
}

// ResNetworkInfo groups resulting values of NetworkInfo operation.
type ResNetworkInfo struct {
	statusRes

	info netmap.NetworkInfo
}

// Info returns structured information about the NeoFS network.
func (x ResNetworkInfo) Info() netmap.NetworkInfo {
	return x.info
}

// NetworkInfo requests information about the NeoFS network of which the remote server is a part.
//
// Any client's internal or transport errors are returned as `error`.
// If PrmInit.ResolveNeoFSFailures has been called, unsuccessful
// NeoFS status codes are returned as `error`, otherwise, are included
// in the returned result structure.
//
// Immediately panics if parameters are set incorrectly (see PrmNetworkInfo docs).
// Context is required and must not be nil. It is used for network communication.
//
// Exactly one return value is non-nil. Server status return is returned in ResNetworkInfo.
// Reflects all internal errors in second return value (transport problems, response processing, etc.).
//
// Return statuses:
//   - global (see Client docs).
func (c *Client) NetworkInfo(ctx context.Context, prm PrmNetworkInfo) (*ResNetworkInfo, error) {
	// check context
	if ctx == nil {
		panic(panicMsgMissingContext)
	}

	// form request
	var req v2netmap.NetworkInfoRequest

	// init call context

	var (
		cc  contextCall
		res ResNetworkInfo
	)

	c.initCallContext(&cc)
	cc.meta = prm.prmCommonMeta
	cc.req = &req
	cc.statusRes = &res
	cc.call = func() (responseV2, error) {
		return rpcapi.NetworkInfo(&c.c, &req, client.WithContext(ctx))
	}
	cc.result = func(r responseV2) {
		resp := r.(*v2netmap.NetworkInfoResponse)

		const fieldNetInfo = "network info"

		netInfoV2 := resp.GetBody().GetNetworkInfo()
		if netInfoV2 == nil {
			cc.err = newErrMissingResponseField(fieldNetInfo)
			return
		}

		cc.err = res.info.ReadFromV2(*netInfoV2)
		if cc.err != nil {
			cc.err = newErrInvalidResponseField(fieldNetInfo, cc.err)
			return
		}
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
