package client

import (
	"context"
	"fmt"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	v2signature "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// EndpointInfo represents versioned information about the node
// specified in the client.
type EndpointInfo struct {
	version *version.Version

	ni *netmap.NodeInfo
}

// LatestVersion returns latest NeoFS API version in use.
func (e *EndpointInfo) LatestVersion() *version.Version {
	return e.version
}

// NodeInfo returns information about the NeoFS node.
func (e *EndpointInfo) NodeInfo() *netmap.NodeInfo {
	return e.ni
}

type EndpointInfoRes struct {
	statusRes

	info *EndpointInfo
}

func (x EndpointInfoRes) Info() *EndpointInfo {
	return x.info
}

func (x *EndpointInfoRes) setInfo(info *EndpointInfo) {
	x.info = info
}

// EndpointInfo returns attributes, address and public key of the node, specified
// in client constructor via address or open connection. This can be used as a
// health check to see if node is alive and responses to requests.
//
// Any client's internal or transport errors are returned as error,
// NeoFS status codes are included in the returned results.
func (c *Client) EndpointInfo(ctx context.Context, opts ...CallOption) (*EndpointInfoRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2netmap.LocalNodeInfoRequestBody)

	req := new(v2netmap.LocalNodeInfoRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.LocalNodeInfo(c.Raw(), req)
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	var (
		res     = new(EndpointInfoRes)
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

	body := resp.GetBody()

	res.setInfo(&EndpointInfo{
		version: version.NewFromV2(body.GetVersion()),
		ni:      netmap.NewNodeInfoFromV2(body.GetNodeInfo()),
	})

	return res, nil
}

type NetworkInfoRes struct {
	statusRes

	info *netmap.NetworkInfo
}

func (x NetworkInfoRes) Info() *netmap.NetworkInfo {
	return x.info
}

func (x *NetworkInfoRes) setInfo(info *netmap.NetworkInfo) {
	x.info = info
}

// NetworkInfo returns information about the NeoFS network of which the remote server is a part.
//
// Any client's internal or transport errors are returned as error,
// NeoFS status codes are included in the returned results.
func (c *Client) NetworkInfo(ctx context.Context, opts ...CallOption) (*NetworkInfoRes, error) {
	// apply all available options
	callOptions := c.defaultCallOptions()

	for i := range opts {
		opts[i](callOptions)
	}

	reqBody := new(v2netmap.NetworkInfoRequestBody)

	req := new(v2netmap.NetworkInfoRequest)
	req.SetBody(reqBody)
	req.SetMetaHeader(v2MetaHeaderFromOpts(callOptions))

	err := v2signature.SignServiceMessage(callOptions.key, req)
	if err != nil {
		return nil, err
	}

	resp, err := rpcapi.NetworkInfo(c.Raw(), req, client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("v2 NetworkInfo RPC failure: %w", err)
	}

	var (
		res     = new(NetworkInfoRes)
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

	res.setInfo(netmap.NewNetworkInfoFromV2(resp.GetBody().GetNetworkInfo()))

	return res, nil
}
