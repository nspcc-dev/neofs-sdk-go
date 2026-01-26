package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// returns Client-compatible Session service handled by given server. Provided
// server must implement [protosession.SessionServiceServer]: the parameter is
// not of this type to support generics.
func newDefaultSessionServiceDesc(t testing.TB, srv any) testService {
	require.Implements(t, (*protosession.SessionServiceServer)(nil), srv)
	return testService{desc: &protosession.SessionService_ServiceDesc, impl: srv}
}

// returns Client of Session service provided by given server. Provided server
// must implement [protosession.SessionServiceServer]: the parameter is not of
// this type to support generics.
func newTestSessionClient(t testing.TB, srv any) *Client {
	return newClient(t, newDefaultSessionServiceDesc(t, srv))
}

type testCreateSessionServer struct {
	protosession.UnimplementedSessionServiceServer
	testCommonUnaryServerSettings[
		*protosession.CreateRequest_Body,
		*protosession.CreateRequest,
		*protosession.CreateResponse_Body,
		*protosession.CreateResponse,
	]
	reqUsr *user.ID
	reqExp uint64
}

// returns [protosession.SessionServiceServer] supporting Create method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestCreateSessionInfoServer() *testCreateSessionServer { return new(testCreateSessionServer) }

// makes the server to assert that any request is for the given user. By
// default, any user is accepted.
func (x *testCreateSessionServer) checkRequestAccount(usr user.ID) { x.reqUsr = &usr }

// makes the server to assert that any request has given expiration epoch. By
// default, expiration must be unset.
func (x *testCreateSessionServer) checkRequestExpirationEpoch(epoch uint64) { x.reqExp = epoch }

func (x *testCreateSessionServer) verifyRequest(req *protosession.CreateRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil && metaHdr.SessionTokenV2 != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("both session token and session token v2 are set"))
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
	// 1. user
	if body.OwnerId == nil {
		return newErrMissingRequestBodyField("user")
	}
	if x.reqUsr != nil {
		if err := checkUserIDTransport(*x.reqUsr, body.OwnerId); err != nil {
			return newErrInvalidRequestField("user", err)
		}
	}
	// 2. expiration epoch
	if body.Expiration != x.reqExp {
		return newErrInvalidRequestField("expiration epoch", errors.New("mismatches the test input"))
	}
	return nil
}

func (x *testCreateSessionServer) Create(_ context.Context, req *protosession.CreateRequest) (*protosession.CreateResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protosession.CreateResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinCreateSessionResponseBody).(*protosession.CreateResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_SessionCreate(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmSessionCreate
	anyUsr := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestCreateSessionInfoServer()
				c := newTestSessionClient(t, srv)

				srv.checkRequestAccount(anyUsr.ID)
				srv.authenticateRequest(anyUsr)
				_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestCreateSessionInfoServer, newTestSessionClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.SessionCreate(ctx, anyUsr, opts)
						return err
					})
				})
				t.Run("expiration epoch", func(t *testing.T) {
					srv := newTestCreateSessionInfoServer()
					c := newTestSessionClient(t, srv)

					epoch := rand.Uint64()
					var opts PrmSessionCreate
					opts.SetExp(epoch)

					srv.checkRequestExpirationEpoch(epoch)
					_, err := c.SessionCreate(ctx, anyUsr, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protosession.CreateResponse_Body
					}{
						{name: "min", body: validMinCreateSessionResponseBody},
						{name: "full", body: validFullCreateSessionResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestCreateSessionInfoServer()
							c := newTestSessionClient(t, srv)

							srv.respondWithBody(tc.body)
							res, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
							require.NoError(t, err)
							require.NotNil(t, res)
							require.Equal(t, validFullCreateSessionResponseBody.Id, res.ID())
							require.Equal(t, validFullCreateSessionResponseBody.SessionKey, res.PublicKey())
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestCreateSessionInfoServer, newTestSessionClient, func(c *Client) error {
						_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "session.SessionService", "Create", func(c *Client) error {
						_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protosession.CreateResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil, assertErr: func(t testing.TB, err error) {
							require.EqualError(t, err, "missing session id field in the response")
						}},
						{name: "ID/missing", body: &protosession.CreateResponse_Body{
							SessionKey: validFullCreateSessionResponseBody.SessionKey,
						}, assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, MissingResponseFieldErr{})
							require.EqualError(t, err, "missing session id field in the response")
						}},
						{name: "session public key/missing", body: &protosession.CreateResponse_Body{
							Id: validFullCreateSessionResponseBody.Id,
						}, assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, MissingResponseFieldErr{})
							require.EqualError(t, err, "missing session key field in the response")
						}},
					}

					testInvalidResponseBodies(t, newTestCreateSessionInfoServer, newTestSessionClient, tcs, func(c *Client) error {
						_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.SessionCreate(ctx, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestCreateSessionInfoServer, newTestSessionClient, func(ctx context.Context, c *Client) error {
			_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, err := newClient(t).SessionCreate(ctx, usertest.FailSigner(anyUsr), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestCreateSessionInfoServer, newTestSessionClient, func(c *Client) error {
			_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestCreateSessionInfoServer, newDefaultSessionServiceDesc, func(c *Client) error {
			_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestCreateSessionInfoServer, newDefaultSessionServiceDesc, stat.MethodSessionCreate,
			[]testedClientOp{
				func(c *Client) error {
					_, err := c.SessionCreate(ctx, nil, anyValidOpts)
					return err
				},
			}, nil, func(c *Client) error {
				_, err := c.SessionCreate(ctx, anyUsr, anyValidOpts)
				return err
			},
		)
	})
}
