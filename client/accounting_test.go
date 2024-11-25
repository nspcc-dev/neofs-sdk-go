package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	protoaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// returns Client-compatible Accounting service handled by given server.
// Provided server must implement [protoaccounting.AccountingServiceServer]: the
// parameter is not of this type to support generics.
func newDefaultAccountingService(t testing.TB, srv any) testService {
	require.Implements(t, (*protoaccounting.AccountingServiceServer)(nil), srv)
	return testService{desc: &protoaccounting.AccountingService_ServiceDesc, impl: srv}
}

// returns Client of Accounting service provided by given server. Provided
// server must implement [protoaccounting.AccountingServiceServer]: the
// parameter is not of this type to support generics.
func newTestAccountingClient(t testing.TB, srv any) *Client {
	return newClient(t, newDefaultAccountingService(t, srv))
}

type testGetBalanceServer struct {
	protoaccounting.UnimplementedAccountingServiceServer
	testCommonUnaryServerSettings[
		*protoaccounting.BalanceRequest_Body,
		v2accounting.BalanceRequestBody,
		*v2accounting.BalanceRequestBody,
		*protoaccounting.BalanceRequest,
		v2accounting.BalanceRequest,
		*v2accounting.BalanceRequest,
		*protoaccounting.BalanceResponse_Body,
		v2accounting.BalanceResponseBody,
		*v2accounting.BalanceResponseBody,
		*protoaccounting.BalanceResponse,
		v2accounting.BalanceResponse,
		*v2accounting.BalanceResponse,
	]
	reqAcc *user.ID
}

// returns [protoaccounting.AccountingServiceServer] supporting Balance method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestGetBalanceServer() *testGetBalanceServer { return new(testGetBalanceServer) }

// makes the server to assert that any request is for the given
// account. By default, any account is accepted.
func (x *testGetBalanceServer) checkRequestAccount(acc user.ID) {
	x.reqAcc = &acc
}

func (x *testGetBalanceServer) verifyRequest(req *protoaccounting.BalanceRequest) error {
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
	switch {
	case body == nil:
		return newInvalidRequestBodyErr(errors.New("missing body"))
	case body.OwnerId == nil:
		return newErrMissingRequestBodyField("account")
	}
	if x.reqAcc != nil {
		if err := checkUserIDTransport(*x.reqAcc, body.OwnerId); err != nil {
			return newErrInvalidRequestField("account", err)
		}
	}
	return nil
}

func (x *testGetBalanceServer) Balance(_ context.Context, req *protoaccounting.BalanceRequest) (*protoaccounting.BalanceResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoaccounting.BalanceResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinBalanceResponseBody).(*protoaccounting.BalanceResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_BalanceGet(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	anyUsr := usertest.ID()
	var anyValidPrm PrmBalanceGet
	anyValidPrm.SetAccount(anyUsr)

	t.Run("invalid user input", func(t *testing.T) {
		t.Run("missing account", func(t *testing.T) {
			_, err := c.BalanceGet(context.Background(), PrmBalanceGet{})
			require.ErrorIs(t, err, ErrMissingAccount)
		})
	})
	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestGetBalanceServer()
				c := newTestAccountingClient(t, srv)

				var prm PrmBalanceGet
				prm.SetAccount(anyUsr)

				srv.checkRequestAccount(anyUsr)
				srv.authenticateRequest(c.prm.signer)
				_, err := c.BalanceGet(ctx, prm)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestGetBalanceServer, newTestAccountingClient, func(c *Client, xhs []string) error {
						opts := anyValidPrm
						opts.WithXHeaders(xhs...)
						_, err := c.BalanceGet(ctx, opts)
						return err
					})
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protoaccounting.BalanceResponse_Body
					}{
						{name: "min", body: validMinBalanceResponseBody},
						{name: "full", body: validFullBalanceResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestGetBalanceServer()
							c := newTestAccountingClient(t, srv)

							var prm PrmBalanceGet
							prm.SetAccount(anyUsr)

							srv.respondWithBody(tc.body)
							balance, err := c.BalanceGet(ctx, anyValidPrm)
							require.NoError(t, err)
							require.NoError(t, checkBalanceTransport(balance, tc.body.GetBalance()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestGetBalanceServer, newTestAccountingClient, func(c *Client) error {
						_, err := c.BalanceGet(ctx, anyValidPrm)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "accounting.AccountingService", "Balance", func(c *Client) error {
						_, err := c.BalanceGet(ctx, anyValidPrm)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestGetBalanceServer, newTestAccountingClient, func(c *Client) error {
						_, err := c.BalanceGet(ctx, anyValidPrm)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					tcs := []invalidResponseBodyTestcase[protoaccounting.BalanceResponse_Body]{
						{name: "missing", body: nil,
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, MissingResponseFieldErr{})
								require.EqualError(t, err, "missing balance field in the response")
							}},
						{name: "missing", body: new(protoaccounting.BalanceResponse_Body),
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, MissingResponseFieldErr{})
								require.EqualError(t, err, "missing balance field in the response")
							}},
					}

					testInvalidResponseBodies(t, newTestGetBalanceServer, newTestAccountingClient, tcs, func(c *Client) error {
						_, err := c.BalanceGet(ctx, anyValidPrm)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestGetBalanceServer, newTestAccountingClient, func(ctx context.Context, c *Client) error {
			_, err := c.BalanceGet(ctx, anyValidPrm)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.BalanceGet(ctx, anyValidPrm)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestGetBalanceServer, newTestAccountingClient, func(c *Client) error {
			_, err := c.BalanceGet(ctx, anyValidPrm)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestGetBalanceServer, newDefaultAccountingService, func(c *Client) error {
			_, err := c.BalanceGet(ctx, anyValidPrm)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestGetBalanceServer, newDefaultAccountingService, stat.MethodBalanceGet,
			nil,
			[]testedClientOp{func(c *Client) error {
				_, err := c.BalanceGet(ctx, PrmBalanceGet{})
				return err
			}}, func(c *Client) error {
				_, err := c.BalanceGet(ctx, anyValidPrm)
				return err
			},
		)
	})
}
