package client

import (
	"context"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/reputation"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
)

type testAnnounceIntermediateReputationServer struct {
	unimplementedNeoFSAPIServer
}

func (x testAnnounceIntermediateReputationServer) announceIntermediateReputation(context.Context, reputation.AnnounceIntermediateResultRequest) (*reputation.AnnounceIntermediateResultResponse, error) {
	var resp reputation.AnnounceIntermediateResultResponse

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}

type testAnnounceLocalTrustServer struct {
	unimplementedNeoFSAPIServer
}

func (x testAnnounceLocalTrustServer) announceLocalTrust(context.Context, reputation.AnnounceLocalTrustRequest) (*reputation.AnnounceLocalTrustResponse, error) {
	var resp reputation.AnnounceLocalTrustResponse

	if err := signServiceMessage(neofscryptotest.Signer(), &resp, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return &resp, nil
}
