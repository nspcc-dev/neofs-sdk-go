package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// GetEndpointInfoOptions groups optional parameters of [Client.GetEndpointInfo].
type GetEndpointInfoOptions struct{}

// EndpointInfo is a result of [Client.GetEndpointInfo] operation.
type EndpointInfo struct {
	// The latest NeoFS API protocol's version in use.
	LatestVersion version.Version
	// Information about the NeoFS node served on the remote endpoint.
	Node netmap.NodeInfo
}

// GetEndpointInfo requests information about the storage node served on the
// remote endpoint. GetEndpointInfo can be used as a health check to see if node
// is alive and responds to requests.
func (c *Client) GetEndpointInfo(ctx context.Context, _ GetEndpointInfoOptions) (EndpointInfo, error) {
	var res EndpointInfo
	var err error
	if c.serverPubKey != nil && c.handleAPIOpResult != nil {
		// serverPubKey can be nil since it's initialized using EndpointInfo itself on dial
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodEndpointInfo, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := new(apinetmap.LocalNodeInfoRequest)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.netmap.LocalNodeInfo(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return res, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return res, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return res, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return res, err
	}
	if sts != nil {
		err = sts // for closure above
		return res, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldVersion = "version"
	if resp.Body.Version == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldVersion) // for closure above
		return res, err
	}
	if err = res.LatestVersion.ReadFromV2(resp.Body.Version); err != nil {
		err = fmt.Errorf("%s (%s)", errInvalidResponseBodyField, fieldVersion) // for closure above
		return res, err
	}
	const fieldNodeInfo = "node info"
	if resp.Body.NodeInfo == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldNodeInfo) // for closure above
		return res, err
	} else if err = res.Node.ReadFromV2(resp.Body.NodeInfo); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldNodeInfo, err) // for closure above
		return res, err
	}
	return res, nil
}

// GetNetworkInfoOptions groups optional parameters of [Client.GetNetworkInfo].
type GetNetworkInfoOptions struct{}

// GetNetworkInfo requests information about the NeoFS network of which the remote
// server is a part.
func (c *Client) GetNetworkInfo(ctx context.Context, _ GetNetworkInfoOptions) (netmap.NetworkInfo, error) {
	var res netmap.NetworkInfo
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodNetworkInfo, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := new(apinetmap.NetworkInfoRequest)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.netmap.NetworkInfo(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return res, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return res, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return res, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return res, err
	}
	if sts != nil {
		err = sts // for closure above
		return res, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldNetworkInfo = "network info"
	if resp.Body.NetworkInfo == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldNetworkInfo) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.NetworkInfo); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldNetworkInfo, err) // for closure above
		return res, err
	}
	return res, nil
}

// GetCurrentNetmapOptions groups optional parameters of [Client.GetCurrentNetmap] operation.
type GetCurrentNetmapOptions struct{}

// GetCurrentNetmap requests current network map from the remote server.
func (c *Client) GetCurrentNetmap(ctx context.Context, _ GetCurrentNetmapOptions) (netmap.NetMap, error) {
	var res netmap.NetMap
	var err error
	if c.handleAPIOpResult != nil {
		defer func(start time.Time) {
			c.handleAPIOpResult(c.serverPubKey, c.endpoint, stat.MethodNetMapSnapshot, time.Since(start), err)
		}(time.Now())
	}

	// form request
	req := new(apinetmap.NetmapSnapshotRequest)
	// FIXME: balance requests need small fixed-size buffers for encoding, its makes
	// no sense to mosh them with other buffers
	buf := c.signBuffers.Get().(*[]byte)
	defer c.signBuffers.Put(buf)
	if req.VerifyHeader, err = neofscrypto.SignRequest(c.signer, req, req.Body, *buf); err != nil {
		err = fmt.Errorf("%s: %w", errSignRequest, err) // for closure above
		return res, err
	}

	// send request
	resp, err := c.transport.netmap.NetmapSnapshot(ctx, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", errTransport, err) // for closure above
		return res, err
	}

	// intercept response info
	if c.interceptAPIRespInfo != nil {
		if err = c.interceptAPIRespInfo(ResponseMetaInfo{
			key:   resp.GetVerifyHeader().GetBodySignature().GetKey(),
			epoch: resp.GetMetaHeader().GetEpoch(),
		}); err != nil {
			err = fmt.Errorf("%s: %w", errInterceptResponseInfo, err) // for closure above
			return res, err
		}
	}

	// verify response integrity
	if err = neofscrypto.VerifyResponse(resp, resp.Body); err != nil {
		err = fmt.Errorf("%s: %w", errResponseSignature, err) // for closure above
		return res, err
	}
	sts, err := apistatus.ErrorFromV2(resp.GetMetaHeader().GetStatus())
	if err != nil {
		err = fmt.Errorf("%s: %w", errInvalidResponseStatus, err) // for closure above
		return res, err
	}
	if sts != nil {
		err = sts // for closure above
		return res, err
	}

	// decode response payload
	if resp.Body == nil {
		err = errors.New(errMissingResponseBody) // for closure above
		return res, err
	}
	const fieldNetmap = "network map"
	if resp.Body.Netmap == nil {
		err = fmt.Errorf("%s (%s)", errMissingResponseBodyField, fieldNetmap) // for closure above
		return res, err
	} else if err = res.ReadFromV2(resp.Body.Netmap); err != nil {
		err = fmt.Errorf("%s (%s): %w", errInvalidResponseBodyField, fieldNetmap, err) // for closure above
		return res, err
	}
	return res, nil
}
