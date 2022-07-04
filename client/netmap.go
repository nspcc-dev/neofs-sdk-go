package client

import (
	"context"
	"errors"
	"fmt"

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

	version *version.Version

	ni *netmap.NodeInfo
}

// LatestVersion returns latest NeoFS API protocol's version in use.
//
// Client doesn't retain value so modification is safe.
func (x ResEndpointInfo) LatestVersion() *version.Version {
	return x.version
}

func (x *ResEndpointInfo) setLatestVersion(ver *version.Version) {
	x.version = ver
}

// NodeInfo returns information about the NeoFS node served on the remote endpoint.
//
// Client doesn't retain value so modification is safe.
func (x ResEndpointInfo) NodeInfo() *netmap.NodeInfo {
	return x.ni
}

func (x *ResEndpointInfo) setNodeInfo(info *netmap.NodeInfo) {
	x.ni = info
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
//  - global (see Client docs).
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

		var ver version.Version
		if v2ver := body.GetVersion(); v2ver != nil {
			ver.ReadFromV2(*v2ver)
		}
		res.setLatestVersion(&ver)

		nodeV2 := body.GetNodeInfo()
		if nodeV2 == nil {
			cc.err = errors.New("missing node info in response body")
			return
		}

		var node netmap.NodeInfo

		cc.err = node.ReadFromV2(*nodeV2)
		if cc.err != nil {
			cc.err = fmt.Errorf("invalid node info: %w", cc.err)
			return
		}

		res.setNodeInfo(&node)
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

	info *netmap.NetworkInfo
}

// Info returns structured information about the NeoFS network.
//
// Client doesn't retain value so modification is safe.
func (x ResNetworkInfo) Info() *netmap.NetworkInfo {
	return x.info
}

func (x *ResNetworkInfo) setInfo(info *netmap.NetworkInfo) {
	x.info = info
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
//  - global (see Client docs).
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

		netInfoV2 := resp.GetBody().GetNetworkInfo()
		if netInfoV2 == nil {
			cc.err = errors.New("missing network info in response body")
			return
		}

		var netInfo netmap.NetworkInfo

		cc.err = netInfo.ReadFromV2(*netInfoV2)
		if cc.err != nil {
			cc.err = fmt.Errorf("invalid network info: %w", cc.err)
			return
		}

		res.setInfo(&netInfo)
	}

	// process call
	if !cc.processCall() {
		return nil, cc.err
	}

	return &res, nil
}
