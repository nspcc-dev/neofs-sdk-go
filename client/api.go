package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/reputation"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/common"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
)

type getObjectResponseStream interface {
	Read(*object.GetResponse) error
}

type getObjectPayloadRangeResponseStream interface {
	Read(*object.GetRangeResponse) error
}

type searchObjectsResponseStream interface {
	Read(*object.SearchResponse) error
}

// interface of NeoFS API server. Exists for test purposes only.
type neoFSAPIServer interface {
	createSession(context.Context, session.CreateRequest) (*session.CreateResponse, error)
	getBalance(context.Context, accounting.BalanceRequest) (*accounting.BalanceResponse, error)
	netMapSnapshot(context.Context, netmap.SnapshotRequest) (*netmap.SnapshotResponse, error)
	getNetworkInfo(context.Context, netmap.NetworkInfoRequest) (*netmap.NetworkInfoResponse, error)
	getNodeInfo(context.Context, netmap.LocalNodeInfoRequest) (*netmap.LocalNodeInfoResponse, error)
	putContainer(context.Context, container.PutRequest) (*container.PutResponse, error)
	getContainer(context.Context, container.GetRequest) (*container.GetResponse, error)
	deleteContainer(context.Context, container.DeleteRequest) (*container.DeleteResponse, error)
	listContainers(context.Context, container.ListRequest) (*container.ListResponse, error)
	getEACL(context.Context, container.GetExtendedACLRequest) (*container.GetExtendedACLResponse, error)
	setEACL(context.Context, container.SetExtendedACLRequest) (*container.SetExtendedACLResponse, error)
	announceContainerSpace(context.Context, container.AnnounceUsedSpaceRequest) (*container.AnnounceUsedSpaceResponse, error)
	announceIntermediateReputation(context.Context, reputation.AnnounceIntermediateResultRequest) (*reputation.AnnounceIntermediateResultResponse, error)
	announceLocalTrust(context.Context, reputation.AnnounceLocalTrustRequest) (*reputation.AnnounceLocalTrustResponse, error)
	putObject(context.Context) (objectWriter, error)
	deleteObject(context.Context, object.DeleteRequest) (*object.DeleteResponse, error)
	hashObjectPayloadRanges(context.Context, object.GetRangeHashRequest) (*object.GetRangeHashResponse, error)
	headObject(context.Context, object.HeadRequest) (*object.HeadResponse, error)
	getObject(context.Context, object.GetRequest) (getObjectResponseStream, error)
	getObjectPayloadRange(context.Context, object.GetRangeRequest) (getObjectPayloadRangeResponseStream, error)
	searchObjects(context.Context, object.SearchRequest) (searchObjectsResponseStream, error)
}

// wrapper over real client connection which communicates over NeoFS API protocol.
// Provides neoFSAPIServer for Client instances used in real applications.
type coreServer client.Client

// unifies errors of all RPC.
func rpcErr(e error) error {
	return fmt.Errorf("rpc failure: %w", e)
}

// executes NetmapService.NetmapSnapshot RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) netMapSnapshot(ctx context.Context, req netmap.SnapshotRequest) (*netmap.SnapshotResponse, error) {
	resp, err := rpcapi.NetMapSnapshot((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}

	return resp, nil
}

// executes NetmapService.NetworkInfo RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getNetworkInfo(ctx context.Context, req netmap.NetworkInfoRequest) (*netmap.NetworkInfoResponse, error) {
	resp, err := rpcapi.NetworkInfo((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes NetmapService.LocalNodeInfo RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getNodeInfo(ctx context.Context, req netmap.LocalNodeInfoRequest) (*netmap.LocalNodeInfoResponse, error) {
	resp, err := rpcapi.LocalNodeInfo((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes AccountingService.Balance RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getBalance(ctx context.Context, req accounting.BalanceRequest) (*accounting.BalanceResponse, error) {
	resp, err := rpcapi.Balance((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes SessionService.Create RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) createSession(ctx context.Context, req session.CreateRequest) (*session.CreateResponse, error) {
	resp, err := rpcapi.CreateSession((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}

	return resp, nil
}

// executes ContainerService.Put RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) putContainer(ctx context.Context, req container.PutRequest) (*container.PutResponse, error) {
	resp, err := rpcapi.PutContainer((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ContainerService.Get RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getContainer(ctx context.Context, req container.GetRequest) (*container.GetResponse, error) {
	resp, err := rpcapi.GetContainer((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ContainerService.Delete RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) deleteContainer(ctx context.Context, req container.DeleteRequest) (*container.DeleteResponse, error) {
	// rpcapi.DeleteContainer returns wrong response type
	resp := new(container.DeleteResponse)
	err := client.SendUnary((*client.Client)(x),
		common.CallMethodInfoUnary("neo.fs.v2.container.ContainerService", "Delete"),
		&req, resp, client.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// executes ContainerService.List RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) listContainers(ctx context.Context, req container.ListRequest) (*container.ListResponse, error) {
	resp, err := rpcapi.ListContainers((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ContainerService.GetExtendedACL RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getEACL(ctx context.Context, req container.GetExtendedACLRequest) (*container.GetExtendedACLResponse, error) {
	resp, err := rpcapi.GetEACL((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ContainerService.SetExtendedACL RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) setEACL(ctx context.Context, req container.SetExtendedACLRequest) (*container.SetExtendedACLResponse, error) {
	// rpcapi.SetEACL returns wrong response type
	resp := new(container.SetExtendedACLResponse)
	err := client.SendUnary((*client.Client)(x),
		common.CallMethodInfoUnary("neo.fs.v2.container.ContainerService", "SetExtendedACL"),
		&req, resp, client.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// executes ContainerService.AnnounceUsedSpace RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) announceContainerSpace(ctx context.Context, req container.AnnounceUsedSpaceRequest) (*container.AnnounceUsedSpaceResponse, error) {
	// rpcapi.AnnounceUsedSpace returns wrong response type
	resp := new(container.AnnounceUsedSpaceResponse)
	err := client.SendUnary((*client.Client)(x),
		common.CallMethodInfoUnary("neo.fs.v2.container.ContainerService", "AnnounceUsedSpace"),
		&req, resp, client.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// executes ReputationService.AnnounceIntermediateResult RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) announceIntermediateReputation(ctx context.Context, req reputation.AnnounceIntermediateResultRequest) (*reputation.AnnounceIntermediateResultResponse, error) {
	resp, err := rpcapi.AnnounceIntermediateResult((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ReputationService.AnnounceLocalTrust RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) announceLocalTrust(ctx context.Context, req reputation.AnnounceLocalTrustRequest) (*reputation.AnnounceLocalTrustResponse, error) {
	resp, err := rpcapi.AnnounceLocalTrust((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

type corePutObjectStream struct {
	reqWriter *rpcapi.PutRequestWriter
	resp      object.PutResponse
}

func (x *corePutObjectStream) Write(req *object.PutRequest) error {
	return x.reqWriter.Write(req)
}

func (x *corePutObjectStream) Close() (*object.PutResponse, error) {
	if err := x.reqWriter.Close(); err != nil {
		return nil, err
	}
	return &x.resp, nil
}

// executes ObjectService.Put RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) putObject(ctx context.Context) (objectWriter, error) {
	var err error
	var stream corePutObjectStream
	stream.reqWriter, err = rpcapi.PutObject((*client.Client)(x), &stream.resp, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return &stream, nil
}

// executes ObjectService.Delete RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) deleteObject(ctx context.Context, req object.DeleteRequest) (*object.DeleteResponse, error) {
	resp, err := rpcapi.DeleteObject((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ObjectService.GetRangeHash RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) hashObjectPayloadRanges(ctx context.Context, req object.GetRangeHashRequest) (*object.GetRangeHashResponse, error) {
	resp, err := rpcapi.HashObjectRange((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ObjectService.Head RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) headObject(ctx context.Context, req object.HeadRequest) (*object.HeadResponse, error) {
	resp, err := rpcapi.HeadObject((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return resp, nil
}

// executes ObjectService.Get RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getObject(ctx context.Context, req object.GetRequest) (getObjectResponseStream, error) {
	stream, err := rpcapi.GetObject((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return stream, nil
}

// executes ObjectService.GetRange RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) getObjectPayloadRange(ctx context.Context, req object.GetRangeRequest) (getObjectPayloadRangeResponseStream, error) {
	stream, err := rpcapi.GetObjectRange((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return stream, nil
}

// executes ObjectService.Search RPC declared in NeoFS API protocol
// using underlying client.Client.
func (x *coreServer) searchObjects(ctx context.Context, req object.SearchRequest) (searchObjectsResponseStream, error) {
	stream, err := rpcapi.SearchObjects((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}
	return stream, nil
}
