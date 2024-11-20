package client

import (
	"testing"

	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
)

// returns Client of Object service provided by given server.
func newTestObjectClient(t testing.TB, srv protoobject.ObjectServiceServer) *Client {
	return newClient(t, testService{desc: &protoobject.ObjectService_ServiceDesc, impl: srv})
}
