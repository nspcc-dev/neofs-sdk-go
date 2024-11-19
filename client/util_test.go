package client

import (
	"context"
	"errors"

	"github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/reputation"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
)

type unimplementedNeoFSAPIServer struct{}

func (unimplementedNeoFSAPIServer) createSession(context.Context, session.CreateRequest) (*session.CreateResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getBalance(context.Context, accounting.BalanceRequest) (*accounting.BalanceResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) netMapSnapshot(context.Context, netmap.SnapshotRequest) (*netmap.SnapshotResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getNetworkInfo(context.Context, netmap.NetworkInfoRequest) (*netmap.NetworkInfoResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getNodeInfo(context.Context, netmap.LocalNodeInfoRequest) (*netmap.LocalNodeInfoResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) putContainer(context.Context, container.PutRequest) (*container.PutResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getContainer(context.Context, container.GetRequest) (*container.GetResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) deleteContainer(context.Context, container.DeleteRequest) (*container.DeleteResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) listContainers(context.Context, container.ListRequest) (*container.ListResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getEACL(context.Context, container.GetExtendedACLRequest) (*container.GetExtendedACLResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) setEACL(context.Context, container.SetExtendedACLRequest) (*container.SetExtendedACLResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) announceContainerSpace(context.Context, container.AnnounceUsedSpaceRequest) (*container.AnnounceUsedSpaceResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) announceIntermediateReputation(context.Context, reputation.AnnounceIntermediateResultRequest) (*reputation.AnnounceIntermediateResultResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) announceLocalTrust(context.Context, reputation.AnnounceLocalTrustRequest) (*reputation.AnnounceLocalTrustResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) putObject(context.Context) (objectWriter, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) deleteObject(context.Context, object.DeleteRequest) (*object.DeleteResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) hashObjectPayloadRanges(context.Context, object.GetRangeHashRequest) (*object.GetRangeHashResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) headObject(context.Context, object.HeadRequest) (*object.HeadResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getObject(context.Context, object.GetRequest) (getObjectResponseStream, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) getObjectPayloadRange(context.Context, object.GetRangeRequest) (getObjectPayloadRangeResponseStream, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) searchObjects(context.Context, object.SearchRequest) (searchObjectsResponseStream, error) {
	return nil, errors.New("unimplemented")
}
