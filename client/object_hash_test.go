package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type testHashObjectPayloadRangesServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonUnaryServerSettings[
		*protoobject.GetRangeHashRequest_Body,
		v2object.GetRangeHashRequestBody,
		*v2object.GetRangeHashRequestBody,
		*protoobject.GetRangeHashRequest,
		v2object.GetRangeHashRequest,
		*v2object.GetRangeHashRequest,
		*protoobject.GetRangeHashResponse_Body,
		v2object.GetRangeHashResponseBody,
		*v2object.GetRangeHashResponseBody,
		*protoobject.GetRangeHashResponse,
		v2object.GetRangeHashResponse,
		*v2object.GetRangeHashResponse,
	]
	testCommonReadObjectRequestServerSettings
	reqHomo   bool
	reqRanges []uint64
	reqSalt   []byte
}

// returns [protoobject.ObjectServiceServer] supporting GetRangeHash method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestHashObjectServer() *testHashObjectPayloadRangesServer {
	return new(testHashObjectPayloadRangesServer)
}

// makes the server to assert that any request has given payload (offset,len)
// ranges. By default, and if nil, any valid ranges are accepted.
func (x *testHashObjectPayloadRangesServer) checkRequestRanges(rs []uint64) {
	if len(rs)%2 != 0 {
		panic("odd number of elements")
	}
	x.reqRanges = rs
}

// makes the server to assert that any request has given salt. By default, and
// if nil, salt must be empty.
func (x *testHashObjectPayloadRangesServer) checkRequestSalt(salt []byte) { x.reqSalt = salt }

// makes the server to assert that any request has homomorphic checksum type.
// By default, the type must be SHA-256.
func (x *testHashObjectPayloadRangesServer) checkRequestHomomorphic() { x.reqHomo = true }

func (x *testHashObjectPayloadRangesServer) verifyRequest(req *protoobject.GetRangeHashRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	if err := x.verifyMeta(req.MetaHeader); err != nil {
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
	// 2. ranges
	if len(body.Ranges) == 0 {
		return newErrMissingRequestBodyField("ranges")
	}
	if x.reqRanges != nil {
		if exp, act := len(x.reqRanges), 2*len(body.Ranges); exp != act {
			return newErrInvalidRequestField("ranges", fmt.Errorf("number of elements (client: %d, message: %d)", exp, act))
		}
		for i, r := range body.Ranges {
			if v1, v2 := r.GetOffset(), x.reqRanges[2*i]; v1 != v2 {
				return newErrInvalidRequestField("ranges", fmt.Errorf("element#%d: offset field (client: %v, message: %v)", i, v1, v2))
			}
			if v1, v2 := r.GetLength(), x.reqRanges[2*i+1]; v1 != v2 {
				return newErrInvalidRequestField("ranges", fmt.Errorf("element#%d: length field (client: %v, message: %v)", i, v1, v2))
			}
		}
	}
	// 3. salt
	if x.reqSalt != nil && !bytes.Equal(body.Salt, x.reqSalt) {
		return newErrInvalidRequestField("salt", fmt.Errorf("unexpected value (client: %x, message: %x)", x.reqSalt, body.Salt))
	}
	// 4. type
	var expType protorefs.ChecksumType
	if x.reqHomo {
		expType = protorefs.ChecksumType_TZ
	} else {
		expType = protorefs.ChecksumType_SHA256
	}
	if body.Type != expType {
		return newErrInvalidRequestField("type", fmt.Errorf("unexpected value (client: %v, message: %v)", expType, body.Type))
	}
	return nil
}

func (x *testHashObjectPayloadRangesServer) GetRangeHash(_ context.Context, req *protoobject.GetRangeHashRequest) (*protoobject.GetRangeHashResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoobject.GetRangeHashResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinObjectHashResponseBody).(*protoobject.GetRangeHashResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_ObjectHash(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectHash
	anyValidOpts.SetRangeList(0, 1)
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
				srv := newTestHashObjectServer()
				c := newTestObjectClient(t, srv)

				rs := []uint64{1, 2, 3, 4, 5, 6}
				var opts PrmObjectHash
				opts.SetRangeList(rs...)

				srv.checkRequestRanges(rs)
				srv.authenticateRequest(anyValidSigner)
				_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestHashObjectServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
						return err
					})
				})
				t.Run("salt", func(t *testing.T) {
					srv := newTestHashObjectServer()
					c := newTestObjectClient(t, srv)

					salt := []byte("any salt")
					opts := anyValidOpts
					opts.UseSalt(salt)

					srv.checkRequestSalt(salt)
					_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("homomorphic", func(t *testing.T) {
					srv := newTestHashObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.TillichZemorAlgo()

					srv.checkRequestHomomorphic()
					_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("local", func(t *testing.T) {
					srv := newTestHashObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkLocal()

					srv.checkRequestLocal()
					_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestHashObjectServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestHashObjectServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					bt.SetEACLTable(anyValidEACL) // TODO: drop after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protoobject.GetRangeHashResponse_Body
					}{
						{name: "min", body: validMinObjectHashResponseBody},
						{name: "full", body: validFullObjectHashResponseBody},
						{name: "type/negative", body: &protoobject.GetRangeHashResponse_Body{
							// https://github.com/nspcc-dev/neofs-sdk-go/issues/663
							Type: -1, HashList: validMinObjectHashResponseBody.GetHashList(),
						}},
						{name: "type/unsupported", body: &protoobject.GetRangeHashResponse_Body{
							Type: math.MaxInt32, HashList: validMinObjectHashResponseBody.GetHashList(),
						}},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestHashObjectServer()
							c := newTestObjectClient(t, srv)

							srv.respondWithBody(tc.body)
							res, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							require.Equal(t, tc.body.GetHashList(), res)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestHashObjectServer, newTestObjectClient, func(c *Client) error {
						_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "GetRangeHash", func(c *Client) error {
						_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestHashObjectServer, newTestObjectClient, func(c *Client) error {
						_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protoobject.GetRangeHashResponse_Body]
					// missing fields
					tcs := []testcase{
						{name: "nil", body: nil, assertErr: func(t testing.TB, err error) {
							require.EqualError(t, err, "missing hash list field in the response")
						}},
						{name: "empty", body: new(protoobject.GetRangeHashResponse_Body), assertErr: func(t testing.TB, err error) {
							require.EqualError(t, err, "missing hash list field in the response")
						}},
					}

					testInvalidResponseBodies(t, newTestHashObjectServer, newTestObjectClient, tcs, func(c *Client) error {
						_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ObjectHash(ctx, anyCID, anyOID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
		t.Run("missing ranges", func(t *testing.T) {
			var opts PrmObjectHash
			_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
			require.ErrorIs(t, err, ErrMissingRanges)

			opts = anyValidOpts
			opts.SetRangeList()
			_, err = c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, opts)
			require.ErrorIs(t, err, ErrMissingRanges)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestHashObjectServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, err := newClient(t).ObjectHash(ctx, anyCID, anyOID, usertest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestHashObjectServer, newTestObjectClient, func(c *Client) error {
			_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/653")
		testUnaryResponseCallback(t, newTestHashObjectServer, newDefaultObjectService, func(c *Client) error {
			_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestHashObjectServer, newDefaultObjectService, stat.MethodObjectHash,
			[]testedClientOp{func(c *Client) error {
				_, err := c.ObjectHash(ctx, anyCID, anyOID, nil, anyValidOpts)
				return err
			}}, []testedClientOp{func(c *Client) error {
				_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, PrmObjectHash{})
				return err
			}}, func(c *Client) error {
				_, err := c.ObjectHash(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
				return err
			},
		)
	})
}
