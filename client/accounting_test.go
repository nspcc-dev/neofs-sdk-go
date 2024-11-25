package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	protoaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

// returns Client-compatible Accounting service handler by given srv. Provided
// server must implement [protoaccounting.AccountingServiceServer]: the
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
	testCommonServerSettings[
		*protoaccounting.BalanceRequest,
		v2accounting.BalanceRequest,
		*v2accounting.BalanceRequest,
		protoaccounting.BalanceResponse_Body,
		protoaccounting.BalanceResponse,
		v2accounting.BalanceResponse,
		*v2accounting.BalanceResponse,
	]
	reqAcc []byte
}

// returns [protoaccounting.AccountingServiceServer] supporting Balance method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestGetBalanceServer() *testGetBalanceServer { return new(testGetBalanceServer) }

// makes the server to assert that any request is for the given
// account. By default, any account is accepted.
func (x *testGetBalanceServer) checkRequestAccount(acc user.ID) {
	x.reqAcc = acc[:]
}

// makes the server to always respond with the given balance. By default, any
// valid balance is returned.
//
// Conflicts with respondWithBody.
func (x *testGetBalanceServer) respondWithBalance(balance *protoaccounting.Decimal) {
	x.respondWithBody(&protoaccounting.BalanceResponse_Body{
		Balance: balance,
	})
}

func (x *testGetBalanceServer) verifyRequest(req *protoaccounting.BalanceRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	if req.MetaHeader.SessionToken != nil {
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	}
	// body
	body := req.Body
	switch {
	case body == nil:
		return newInvalidRequestBodyErr(errors.New("missing body"))
	case body.OwnerId == nil:
		return newErrMissingRequestBodyField("account")
	}
	if x.reqAcc != nil && !bytes.Equal(body.OwnerId.Value, x.reqAcc[:]) {
		return newErrInvalidRequestField("account", fmt.Errorf("test input mismatch"))
	}
	return nil
}

func (x *testGetBalanceServer) Balance(_ context.Context, req *protoaccounting.BalanceRequest) (*protoaccounting.BalanceResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protoaccounting.BalanceResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = &protoaccounting.BalanceResponse_Body{
			Balance: &protoaccounting.Decimal{
				Value:     rand.Int63(),
				Precision: rand.Uint32(),
			},
		}
	}

	return x.signResponse(&resp)
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
	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			val := rand.Int63()
			precision := rand.Uint32()

			srv := newTestGetBalanceServer()
			srv.checkRequestAccount(anyUsr)
			srv.respondWithBalance(&protoaccounting.Decimal{
				Value:     val,
				Precision: precision,
			})
			c := newTestAccountingClient(t, srv)

			var prm PrmBalanceGet
			prm.SetAccount(anyUsr)
			balance, err := c.BalanceGet(ctx, anyValidPrm)
			require.NoError(t, err)
			require.EqualValues(t, val, balance.Value())
			require.EqualValues(t, precision, balance.Precision())
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
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestGetBalanceServer, newTestAccountingClient, func(c *Client) error {
				_, err := c.BalanceGet(ctx, anyValidPrm)
				return err
			})
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
	t.Run("response message type mismatch", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "accounting.AccountingService", "Balance", func(c *Client) error {
			_, err := c.BalanceGet(ctx, anyValidPrm)
			return err
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestGetBalanceServer, newTestAccountingClient, func(c *Client) error {
			_, err := c.BalanceGet(ctx, anyValidPrm)
			return err
		})
	})
	t.Run("invalid response body", func(t *testing.T) {
		tcs := []invalidResponseBodyTestcase[protoaccounting.BalanceResponse_Body]{
			{name: "missing", body: nil,
				assertErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, MissingResponseFieldErr{})
					require.EqualError(t, err, "missing balance field in the response")
					// TODO: worth clarifying that body is completely missing
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
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestGetBalanceServer, newDefaultAccountingService, func(c *Client) error {
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
