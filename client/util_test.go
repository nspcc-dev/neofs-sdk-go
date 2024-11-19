package client

import (
	"context"
	"errors"

	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
)

type unimplementedNeoFSAPIServer struct{}

func (unimplementedNeoFSAPIServer) createSession(context.Context, session.CreateRequest) (*session.CreateResponse, error) {
	return nil, errors.New("unimplemented")
}
func (unimplementedNeoFSAPIServer) netMapSnapshot(context.Context, netmap.SnapshotRequest) (*netmap.SnapshotResponse, error) {
	return nil, errors.New("unimplemented")
}
