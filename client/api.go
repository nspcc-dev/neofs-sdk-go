package client

import (
	apiaccounting "github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	apireputation "github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"google.golang.org/grpc"
)

// cross-RPC non-status errors.
var (
	errSignRequest              = "sign request"
	errTransport                = "transport failure"
	errInterceptResponseInfo    = "intercept response info"
	errResponseSignature        = "verify response signature"
	errInvalidResponse          = "invalid response"
	errInvalidResponseStatus    = errInvalidResponse + ": invalid status"
	errMissingResponseBody      = errInvalidResponse + ": missing body"
	errInvalidResponseBody      = errInvalidResponse + ": invalid body"
	errMissingResponseBodyField = errInvalidResponseBody + ": missing required field"
	errInvalidResponseBodyField = errInvalidResponseBody + ": invalid field"
)

// unites all NeoFS services served over gRPC.
type grpcTransport struct {
	accounting apiaccounting.AccountingServiceClient
	container  apicontainer.ContainerServiceClient
	netmap     apinetmap.NetmapServiceClient
	object     apiobject.ObjectServiceClient
	reputation apireputation.ReputationServiceClient
	session    apisession.SessionServiceClient
}

func newGRPCTransport(con *grpc.ClientConn) grpcTransport {
	return grpcTransport{
		accounting: apiaccounting.NewAccountingServiceClient(con),
		container:  apicontainer.NewContainerServiceClient(con),
		netmap:     apinetmap.NewNetmapServiceClient(con),
		object:     apiobject.NewObjectServiceClient(con),
		reputation: apireputation.NewReputationServiceClient(con),
		session:    apisession.NewSessionServiceClient(con),
	}
}
