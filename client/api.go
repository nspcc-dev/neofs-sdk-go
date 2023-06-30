package client

import (
	"context"
	"fmt"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	rpcapi "github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
)

var (
	// special variables for test purposes only, to overwrite real RPC calls.
	rpcAPINetMapSnapshot = rpcapi.NetMapSnapshot
	rpcAPICreateSession  = rpcapi.CreateSession
)

// interface of NeoFS API server. Exists for test purposes only.
type neoFSAPIServer interface {
	createSession(cli *client.Client, req *session.CreateRequest, opts ...client.CallOption) (*session.CreateResponse, error)

	netMapSnapshot(context.Context, v2netmap.SnapshotRequest) (*v2netmap.SnapshotResponse, error)
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
func (x *coreServer) netMapSnapshot(ctx context.Context, req v2netmap.SnapshotRequest) (*v2netmap.SnapshotResponse, error) {
	resp, err := rpcAPINetMapSnapshot((*client.Client)(x), &req, client.WithContext(ctx))
	if err != nil {
		return nil, rpcErr(err)
	}

	return resp, nil
}

func (x *coreServer) createSession(cli *client.Client, req *session.CreateRequest, opts ...client.CallOption) (*session.CreateResponse, error) {
	resp, err := rpcAPICreateSession(cli, req, opts...)
	if err != nil {
		return nil, rpcErr(err)
	}

	return resp, nil
}
