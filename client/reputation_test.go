package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	protoreputation "github.com/nspcc-dev/neofs-sdk-go/proto/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// returns Client-compatible Reputation service handled by given server.
// Provided server must implement [protoreputation.ReputationServiceServer]: the
// parameter is not of this type to support generics.
func newDefaultReputationServiceDesc(t testing.TB, srv any) testService {
	require.Implements(t, (*protoreputation.ReputationServiceServer)(nil), srv)
	return testService{desc: &protoreputation.ReputationService_ServiceDesc, impl: srv}
}

// returns Client of Reputation service provided by given server. Provided
// server must implement [protoreputation.ReputationServiceServer]: the
// parameter is not of this type to support generics.
func newTestReputationClient(t testing.TB, srv any) *Client {
	return newClient(t, newDefaultReputationServiceDesc(t, srv))
}

type testAnnounceIntermediateReputationServer struct {
	protoreputation.UnimplementedReputationServiceServer
	testCommonUnaryServerSettings[
		*protoreputation.AnnounceIntermediateResultRequest_Body,
		*protoreputation.AnnounceIntermediateResultRequest,
		*protoreputation.AnnounceIntermediateResultResponse_Body,
		*protoreputation.AnnounceIntermediateResultResponse,
	]
	reqEpoch *uint64
	reqIter  uint32
	reqTrust *reputation.PeerToPeerTrust
}

// returns [protoreputation.ReputationServiceServer] supporting
// AnnounceIntermediateResult method only. Default implementation performs
// common verification of any request, and responds with any valid message. Some
// methods allow to tune the behavior.
func newTestAnnounceIntermediateReputationServer() *testAnnounceIntermediateReputationServer {
	return new(testAnnounceIntermediateReputationServer)
}

// makes the server to assert that any request is for the given epoch. By
// default, any epoch is accepted.
func (x *testAnnounceIntermediateReputationServer) checkRequestEpoch(epoch uint64) {
	x.reqEpoch = &epoch
}

// makes the server to assert that any request is for the given iteration. By
// default, iteration must be unset.
func (x *testAnnounceIntermediateReputationServer) checkRequestIteration(iter uint32) {
	x.reqIter = iter
}

// makes the server to assert that any request has given trust. By default,
// any valid trust is accepted.
func (x *testAnnounceIntermediateReputationServer) checkRequestTrust(t reputation.PeerToPeerTrust) {
	x.reqTrust = &t
}

func (x *testAnnounceIntermediateReputationServer) verifyRequest(req *protoreputation.AnnounceIntermediateResultRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. epoch
	if body.Epoch == 0 {
		return newErrInvalidRequestField("epoch", errors.New("zero"))
	}
	if x.reqEpoch != nil && body.Epoch != *x.reqEpoch {
		return newErrInvalidRequestField("epoch", errors.New("mismatches the test input"))
	}
	// 2. iteration
	if body.Iteration != x.reqIter {
		return newErrInvalidRequestField("iteration", errors.New("mismatches the test input"))
	}
	// 3. trust
	if body.Trust == nil {
		return newErrMissingRequestBodyField("trust")
	}
	if x.reqTrust != nil {
		if err := checkP2PTrustTransport(*x.reqTrust, body.Trust); err != nil {
			return newErrInvalidRequestField("trust", err)
		}
	}
	return nil
}

func (x *testAnnounceIntermediateReputationServer) AnnounceIntermediateResult(_ context.Context, req *protoreputation.AnnounceIntermediateResultRequest,
) (*protoreputation.AnnounceIntermediateResultResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoreputation.AnnounceIntermediateResultResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinAnnounceIntermediateRepResponseBody).(*protoreputation.AnnounceIntermediateResultResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testAnnounceLocalTrustServer struct {
	protoreputation.UnimplementedReputationServiceServer
	testCommonUnaryServerSettings[
		*protoreputation.AnnounceLocalTrustRequest_Body,
		*protoreputation.AnnounceLocalTrustRequest,
		*protoreputation.AnnounceLocalTrustResponse_Body,
		*protoreputation.AnnounceLocalTrustResponse,
	]
	reqEpoch  *uint64
	reqTrusts []reputation.Trust
}

// returns [protoreputation.ReputationServiceServer] supporting
// AnnounceLocalTrust method only. Default implementation performs common
// verification of any request, and responds with any valid message. Some
// methods allow to tune the behavior.
func newTestAnnounceLocalTrustServer() *testAnnounceLocalTrustServer {
	return new(testAnnounceLocalTrustServer)
}

// makes the server to assert that any request is for the given epoch. By
// default, any epoch is accepted.
func (x *testAnnounceLocalTrustServer) checkRequestEpoch(epoch uint64) { x.reqEpoch = &epoch }

// makes the server to assert that any request has given trust. By default, and
// if nil, any valid trusts are accepted.
func (x *testAnnounceLocalTrustServer) checkRequestTrusts(ts []reputation.Trust) {
	x.reqTrusts = ts
}

func (x *testAnnounceLocalTrustServer) verifyRequest(req *protoreputation.AnnounceLocalTrustRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. epoch
	if body.Epoch == 0 {
		return newErrInvalidRequestField("epoch", errors.New("zero"))
	}
	if x.reqEpoch != nil && body.Epoch != *x.reqEpoch {
		return newErrInvalidRequestField("epoch", errors.New("mismatches the test input"))
	}
	// 2. trusts
	if len(body.Trusts) == 0 {
		return newErrMissingRequestBodyField("trusts")
	}
	if x.reqTrusts != nil {
		if v1, v2 := len(x.reqTrusts), len(body.Trusts); v1 != v2 {
			return fmt.Errorf("number of trusts (client: %d, message: %d)", v1, v2)
		}
		for i := range x.reqTrusts {
			if err := checkTrustTransport(x.reqTrusts[i], body.Trusts[i]); err != nil {
				return newErrInvalidRequestField("trusts", fmt.Errorf("element #%d: %w", i, err))
			}
		}
	}
	return nil
}

func (x *testAnnounceLocalTrustServer) AnnounceLocalTrust(_ context.Context, req *protoreputation.AnnounceLocalTrustRequest,
) (*protoreputation.AnnounceLocalTrustResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoreputation.AnnounceLocalTrustResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinAnnounceLocalTrustResponseBody).(*protoreputation.AnnounceLocalTrustResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_AnnounceIntermediateTrust(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmAnnounceIntermediateTrust
	const anyValidEpoch = 123
	anyValidTrust := reputationtest.PeerToPeerTrust()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestAnnounceIntermediateReputationServer()
				c := newTestReputationClient(t, srv)

				srv.checkRequestEpoch(anyValidEpoch)
				srv.checkRequestTrust(anyValidTrust)
				srv.authenticateRequest(c.prm.signer)
				err := c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestAnnounceIntermediateReputationServer, newTestReputationClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, opts)
					})
				})
				t.Run("iteration", func(t *testing.T) {
					srv := newTestAnnounceIntermediateReputationServer()
					c := newTestReputationClient(t, srv)

					iter := rand.Uint32()
					opts := anyValidOpts
					opts.SetIteration(iter)

					srv.checkRequestIteration(iter)
					err := c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protoreputation.AnnounceIntermediateResultResponse_Body
					}{
						{name: "min", body: validMinAnnounceIntermediateRepResponseBody},
						{name: "full", body: validFullAnnounceIntermediateRepResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestAnnounceIntermediateReputationServer()
							c := newTestReputationClient(t, srv)

							srv.respondWithBody(tc.body)
							err := c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
							require.NoError(t, err)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestAnnounceIntermediateReputationServer, newTestReputationClient, func(c *Client) error {
						return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "reputation.ReputationService", "AnnounceIntermediateResult", func(c *Client) error {
						return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("zero epoch", func(t *testing.T) {
			err := c.AnnounceIntermediateTrust(ctx, 0, anyValidTrust, anyValidOpts)
			require.ErrorIs(t, err, ErrZeroEpoch)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestAnnounceIntermediateReputationServer, newTestReputationClient, func(ctx context.Context, c *Client) error {
			return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestAnnounceIntermediateReputationServer, newTestReputationClient, func(c *Client) error {
			return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestAnnounceIntermediateReputationServer, newDefaultReputationServiceDesc, func(c *Client) error {
			return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestAnnounceIntermediateReputationServer, newDefaultReputationServiceDesc, stat.MethodAnnounceIntermediateTrust,
			nil, []testedClientOp{func(c *Client) error {
				return c.AnnounceIntermediateTrust(ctx, 0, anyValidTrust, anyValidOpts)
			},
			}, func(c *Client) error {
				return c.AnnounceIntermediateTrust(ctx, anyValidEpoch, anyValidTrust, anyValidOpts)
			},
		)
	})
}

func TestClient_AnnounceLocalTrust(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmAnnounceLocalTrust
	const anyValidEpoch = 123
	anyValidTrusts := []reputation.Trust{reputationtest.Trust(), reputationtest.Trust()}

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestAnnounceLocalTrustServer()
				c := newTestReputationClient(t, srv)

				srv.checkRequestEpoch(anyValidEpoch)
				srv.checkRequestTrusts(anyValidTrusts)
				srv.authenticateRequest(c.prm.signer)
				err := c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestAnnounceLocalTrustServer, newTestReputationClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, opts)
					})
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protoreputation.AnnounceLocalTrustResponse_Body
					}{
						{name: "min", body: validMinAnnounceLocalTrustResponseBody},
						{name: "full", body: validFullAnnounceLocalTrustRepResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestAnnounceLocalTrustServer()
							c := newTestReputationClient(t, srv)

							srv.respondWithBody(tc.body)
							err := c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
							require.NoError(t, err)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestAnnounceLocalTrustServer, newTestReputationClient, func(c *Client) error {
						return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "reputation.ReputationService", "AnnounceLocalTrust", func(c *Client) error {
						return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("zero epoch", func(t *testing.T) {
			err := c.AnnounceLocalTrust(ctx, 0, anyValidTrusts, anyValidOpts)
			require.ErrorIs(t, err, ErrZeroEpoch)
		})
		t.Run("empty trusts", func(t *testing.T) {
			err := c.AnnounceLocalTrust(ctx, anyValidEpoch, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingTrusts)
			err = c.AnnounceLocalTrust(ctx, anyValidEpoch, []reputation.Trust{}, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingTrusts)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestAnnounceLocalTrustServer, newTestReputationClient, func(ctx context.Context, c *Client) error {
			return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestAnnounceLocalTrustServer, newTestReputationClient, func(c *Client) error {
			return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestAnnounceLocalTrustServer, newDefaultReputationServiceDesc, func(c *Client) error {
			return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestAnnounceLocalTrustServer, newDefaultReputationServiceDesc, stat.MethodAnnounceLocalTrust,
			nil, []testedClientOp{func(c *Client) error {
				return c.AnnounceLocalTrust(ctx, 0, anyValidTrusts, anyValidOpts)
			}, func(c *Client) error {
				return c.AnnounceLocalTrust(ctx, anyValidEpoch, nil, anyValidOpts)
			}}, func(c *Client) error {
				return c.AnnounceLocalTrust(ctx, anyValidEpoch, anyValidTrusts, anyValidOpts)
			},
		)
	})
}
