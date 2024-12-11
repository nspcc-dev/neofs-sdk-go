package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type testDeleteObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonUnaryServerSettings[
		*protoobject.DeleteRequest_Body,
		*protoobject.DeleteRequest,
		*protoobject.DeleteResponse_Body,
		*protoobject.DeleteResponse,
	]
	testObjectSessionServerSettings
	testBearerTokenServerSettings
	testObjectAddressServerSettings
}

// returns [protoobject.ObjectServiceServer] supporting Delete method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestDeleteObjectServer() *testDeleteObjectServer { return new(testDeleteObjectServer) }

func (x *testDeleteObjectServer) verifyRequest(req *protoobject.DeleteRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	if req.MetaHeader.Ttl != 2 {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", req.MetaHeader.Ttl))
	}
	if err := x.verifySessionToken(req.MetaHeader.SessionToken); err != nil {
		return err
	}
	if err := x.verifyBearerToken(req.MetaHeader.BearerToken); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. address
	if err := x.verifyObjectAddress(body.Address); err != nil {
		return err
	}
	return nil
}

func (x *testDeleteObjectServer) Delete(_ context.Context, req *protoobject.DeleteRequest) (*protoobject.DeleteResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoobject.DeleteResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validFullDeleteObjectResponseBody).(*protoobject.DeleteResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_ObjectDelete(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectDelete
	anyCID := cidtest.ID()
	anyOID := oidtest.ID()
	anyValidSigner := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestDeleteObjectServer()
				c := newTestObjectClient(t, srv)

				srv.checkRequestObjectAddress(anyCID, anyOID)
				srv.authenticateRequest(anyValidSigner)
				_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, PrmObjectDelete{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestDeleteObjectServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, opts)
						return err
					})
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestDeleteObjectServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestDeleteObjectServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protoobject.DeleteResponse_Body
					}{
						{name: "min", body: validMinDeleteObjectResponseBody},
						{name: "full", body: validFullDeleteObjectResponseBody},
						{name: "invalid container ID", body: &protoobject.DeleteResponse_Body{
							Tombstone: &protorefs.Address{
								ContainerId: &protorefs.ContainerID{Value: []byte("any_invalid")},
								ObjectId:    proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
							},
						}},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestDeleteObjectServer()
							c := newTestObjectClient(t, srv)

							srv.respondWithBody(tc.body)
							id, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							require.NoError(t, checkObjectIDTransport(id, tc.body.GetTombstone().GetObjectId()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestDeleteObjectServer, newTestObjectClient, func(c *Client) error {
						_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "Delete", func(c *Client) error {
						_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestDeleteObjectServer, newTestObjectClient, func(c *Client) error {
						_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protoobject.DeleteResponse_Body]
					// missing fields
					tcs := []testcase{
						{name: "nil", body: nil, assertErr: func(t testing.TB, err error) {
							require.EqualError(t, err, "missing tombstone field in the response")
						}},
						{name: "empty", body: new(protoobject.DeleteResponse_Body), assertErr: func(t testing.TB, err error) {
							require.EqualError(t, err, "missing tombstone field in the response")
						}},
						{name: "tombstone address/object ID/missing", body: &protoobject.DeleteResponse_Body{
							Tombstone: new(protorefs.Address),
						}, assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, ErrMissingResponseField)
							require.EqualError(t, err, "missing tombstone field in the response")
						}},
					}
					// tombstone ID
					for _, tc := range invalidObjectIDProtoTestcases {
						id := proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID)
						tc.corrupt(id)
						tcs = append(tcs, testcase{
							name: "tombstone address/object ID/" + tc.name,
							body: &protoobject.DeleteResponse_Body{Tombstone: &protorefs.Address{ObjectId: id}},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid tombstone field in the response: "+tc.msg)
							},
						})
					}

					testInvalidResponseBodies(t, newTestDeleteObjectServer, newTestObjectClient, tcs, func(c *Client) error {
						_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ObjectDelete(ctx, anyCID, anyOID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestDeleteObjectServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, err := newClient(t).ObjectDelete(ctx, anyCID, anyOID, usertest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestDeleteObjectServer, newTestObjectClient, func(c *Client) error {
			_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/654")
		testUnaryResponseCallback(t, newTestDeleteObjectServer, newDefaultObjectService, func(c *Client) error {
			_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestDeleteObjectServer, newDefaultObjectService, stat.MethodObjectDelete,
			[]testedClientOp{func(c *Client) error {
				_, err := c.ObjectDelete(ctx, anyCID, anyOID, nil, anyValidOpts)
				return err
			}}, nil, func(c *Client) error {
				_, err := c.ObjectDelete(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
				return err
			},
		)
	})
}
