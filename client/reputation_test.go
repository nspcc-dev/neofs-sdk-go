package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/reputation"
	protoreputation "github.com/nspcc-dev/neofs-api-go/v2/reputation/grpc"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
)

// returns Client of Reputation service provided by given server.
func newTestReputationClient(t testing.TB, srv protoreputation.ReputationServiceServer) *Client {
	return newClient(t, testService{desc: &protoreputation.ReputationService_ServiceDesc, impl: srv})
}

type testAnnounceIntermediateReputationServer struct {
	protoreputation.UnimplementedReputationServiceServer
}

func (x *testAnnounceIntermediateReputationServer) AnnounceIntermediateResult(context.Context, *protoreputation.AnnounceIntermediateResultRequest,
) (*protoreputation.AnnounceIntermediateResultResponse, error) {
	var resp protoreputation.AnnounceIntermediateResultResponse

	var respV2 reputation.AnnounceIntermediateResultResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoreputation.AnnounceIntermediateResultResponse), nil
}

type testAnnounceLocalTrustServer struct {
	protoreputation.UnimplementedReputationServiceServer
}

func (x *testAnnounceLocalTrustServer) AnnounceLocalTrust(context.Context, *protoreputation.AnnounceLocalTrustRequest,
) (*protoreputation.AnnounceLocalTrustResponse, error) {
	var resp protoreputation.AnnounceLocalTrustResponse

	var respV2 reputation.AnnounceLocalTrustResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	if err := signServiceMessage(neofscryptotest.Signer(), &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoreputation.AnnounceLocalTrustResponse), nil
}
