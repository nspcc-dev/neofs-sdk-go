package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	protonetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// various sets of Netmap service testcases.
var (
	invalidNodeInfoProtoTestcases = []struct {
		name, msg string
		corrupt   func(valid *protonetmap.NodeInfo)
	}{
		{name: "public key/nil", msg: "missing public key", corrupt: func(valid *protonetmap.NodeInfo) {
			valid.PublicKey = nil
		}},
		{name: "public key/empty", msg: "missing public key", corrupt: func(valid *protonetmap.NodeInfo) {
			valid.PublicKey = []byte{}
		}},
		{name: "addresses/nil", msg: "missing network endpoints", corrupt: func(valid *protonetmap.NodeInfo) {
			valid.Addresses = nil
		}},
		{name: "addresses/empty", msg: "missing network endpoints", corrupt: func(valid *protonetmap.NodeInfo) {
			valid.Addresses = nil
		}},
		{name: "attributes/no key", msg: "empty key of the attribute #1", corrupt: func(valid *protonetmap.NodeInfo) {
			valid.Attributes = []*protonetmap.NodeInfo_Attribute{
				{Key: "k1", Value: "v1"}, {Key: "", Value: "v2"}, {Key: "k3", Value: "v3"},
			}
		}},
		{name: "attributes/no value", msg: "empty value of the attribute k2", corrupt: func(valid *protonetmap.NodeInfo) {
			valid.Attributes = []*protonetmap.NodeInfo_Attribute{
				{Key: "k1", Value: "v1"}, {Key: "k2", Value: ""}, {Key: "k3", Value: "v3"},
			}
		}},
		{name: "attributes/capacity", msg: "invalid Capacity attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
			corrupt: func(valid *protonetmap.NodeInfo) {
				valid.Attributes = []*protonetmap.NodeInfo_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "Capacity", Value: "foo"}, {Key: "k3", Value: "v3"},
				}
			}},
		{name: "attributes/price", msg: "invalid Price attribute: strconv.ParseUint: parsing \"foo\": invalid syntax",
			corrupt: func(valid *protonetmap.NodeInfo) {
				valid.Attributes = []*protonetmap.NodeInfo_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "Price", Value: "foo"}, {Key: "k3", Value: "v3"},
				}
			}},
		{name: "state/negative", msg: "negative state",
			corrupt: func(valid *protonetmap.NodeInfo) {
				valid.State = -1
			}},
	}
)

// returns Client-compatible Netmap service handled by given server. Provided
// server must implement [protocontainer.NetmapServiceServer]: the parameter is
// not of this type to support generics.
func newDefaultNetmapServiceDesc(t testing.TB, srv any) testService {
	require.Implements(t, (*protonetmap.NetmapServiceServer)(nil), srv)
	return testService{desc: &protonetmap.NetmapService_ServiceDesc, impl: srv}
}

// returns Client of Netmap service provided by given server. Provided server
// must implement [protonetmap.NetmapServiceServer]: the parameter is not of
// this type to support generics.
func newTestNetmapClient(t testing.TB, srv any) *Client {
	return newClient(t, newDefaultNetmapServiceDesc(t, srv))
}

type testNetmapSnapshotServer struct {
	protonetmap.UnimplementedNetmapServiceServer
	testCommonUnaryServerSettings[
		*protonetmap.NetmapSnapshotRequest_Body,
		v2netmap.SnapshotRequestBody,
		*v2netmap.SnapshotRequestBody,
		*protonetmap.NetmapSnapshotRequest,
		v2netmap.SnapshotRequest,
		*v2netmap.SnapshotRequest,
		*protonetmap.NetmapSnapshotResponse_Body,
		protonetmap.NetmapSnapshotResponse,
		v2netmap.SnapshotResponse,
		*v2netmap.SnapshotResponse,
	]
}

// returns [protonetmap.NetmapServiceServer] supporting NetmapSnapshot method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestNetmapSnapshotServer() *testNetmapSnapshotServer { return new(testNetmapSnapshotServer) }

// makes the server to always respond with the given network map. By default,
// any valid network map is returned.
//
// Overrides with respondWithBody.

func (x *testNetmapSnapshotServer) verifyRequest(req *protonetmap.NetmapSnapshotRequest) error {
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
	return nil
}

func (x *testNetmapSnapshotServer) NetmapSnapshot(_ context.Context, req *protonetmap.NetmapSnapshotRequest) (*protonetmap.NetmapSnapshotResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := protonetmap.NetmapSnapshotResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinNetmapResponseBody).(*protonetmap.NetmapSnapshotResponse_Body)
	}

	return x.signResponse(&resp)
}

type testGetNetworkInfoServer struct {
	protonetmap.UnimplementedNetmapServiceServer
	testCommonUnaryServerSettings[
		*protonetmap.NetworkInfoRequest_Body,
		v2netmap.NetworkInfoRequestBody,
		*v2netmap.NetworkInfoRequestBody,
		*protonetmap.NetworkInfoRequest,
		v2netmap.NetworkInfoRequest,
		*v2netmap.NetworkInfoRequest,
		*protonetmap.NetworkInfoResponse_Body,
		protonetmap.NetworkInfoResponse,
		v2netmap.NetworkInfoResponse,
		*v2netmap.NetworkInfoResponse,
	]
}

// returns [protonetmap.NetmapServiceServer] supporting NetworkInfo method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestNetworkInfoServer() *testGetNetworkInfoServer { return new(testGetNetworkInfoServer) }

func (x *testGetNetworkInfoServer) verifyRequest(req *protonetmap.NetworkInfoRequest) error {
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
	return nil
}

func (x *testGetNetworkInfoServer) NetworkInfo(_ context.Context, req *protonetmap.NetworkInfoRequest) (*protonetmap.NetworkInfoResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}
	resp := protonetmap.NetworkInfoResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinNetInfoResponseBody).(*protonetmap.NetworkInfoResponse_Body)
	}

	return x.signResponse(&resp)
}

type testGetNodeInfoServer struct {
	protonetmap.UnimplementedNetmapServiceServer
	testCommonUnaryServerSettings[
		*protonetmap.LocalNodeInfoRequest_Body,
		v2netmap.LocalNodeInfoRequestBody,
		*v2netmap.LocalNodeInfoRequestBody,
		*protonetmap.LocalNodeInfoRequest,
		v2netmap.LocalNodeInfoRequest,
		*v2netmap.LocalNodeInfoRequest,
		*protonetmap.LocalNodeInfoResponse_Body,
		protonetmap.LocalNodeInfoResponse,
		v2netmap.LocalNodeInfoResponse,
		*v2netmap.LocalNodeInfoResponse,
	]
}

// returns [protonetmap.NetmapServiceServer] supporting LocalNodeInfo method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestGetNodeInfoServer() *testGetNodeInfoServer { return new(testGetNodeInfoServer) }

// makes the server to always respond with the given node public key. By
// default, any valid key is returned.
//
// Overrides respondWithBody.
func (x *testGetNodeInfoServer) respondWithNodePublicKey(pub []byte) {
	b := proto.Clone(validMinNodeInfoResponseBody).(*protonetmap.LocalNodeInfoResponse_Body)
	b.NodeInfo.PublicKey = pub
	x.respondWithBody(b)
}

func (x *testGetNodeInfoServer) verifyRequest(req *protonetmap.LocalNodeInfoRequest) error {
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
	return nil
}

func (x *testGetNodeInfoServer) LocalNodeInfo(_ context.Context, req *protonetmap.LocalNodeInfoRequest) (*protonetmap.LocalNodeInfoResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := protonetmap.LocalNodeInfoResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinNodeInfoResponseBody).(*protonetmap.LocalNodeInfoResponse_Body)
	}

	return x.signResponse(&resp)
}

func TestClient_EndpointInfo(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmEndpointInfo

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestGetNodeInfoServer()
				c := newTestNetmapClient(t, srv)

				srv.authenticateRequest(c.prm.signer)
				_, err := c.EndpointInfo(ctx, PrmEndpointInfo{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestGetNodeInfoServer, newTestNetmapClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.EndpointInfo(ctx, opts)
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
						body *protonetmap.LocalNodeInfoResponse_Body
					}{
						{name: "min", body: validMinNodeInfoResponseBody},
						{name: "full", body: validFullNodeInfoResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestGetNodeInfoServer()
							c := newTestNetmapClient(t, srv)

							srv.respondWithBody(tc.body)
							res, err := c.EndpointInfo(ctx, anyValidOpts)
							require.NoError(t, err)
							require.NotNil(t, res)
							require.NoError(t, checkVersionTransport(res.LatestVersion(), tc.body.GetVersion()))
							require.NoError(t, checkNodeInfoTransport(res.NodeInfo(), tc.body.GetNodeInfo()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestGetNodeInfoServer, newTestNetmapClient, func(c *Client) error {
						_, err := c.EndpointInfo(ctx, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "netmap.NetmapService", "LocalNodeInfo", func(c *Client) error {
						_, err := c.EndpointInfo(ctx, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestGetNodeInfoServer, newTestNetmapClient, func(c *Client) error {
						_, err := c.EndpointInfo(ctx, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protonetmap.LocalNodeInfoResponse_Body]
					tcs := []testcase{{name: "missing", body: nil, assertErr: func(t testing.TB, err error) {
						require.EqualError(t, err, "missing version field in the response")
						// TODO: worth clarifying that body is completely missing
					}}}

					type corruptedBodyTestcase = struct {
						name      string
						corrupt   func(valid *protonetmap.LocalNodeInfoResponse_Body)
						assertErr func(testing.TB, error)
					}
					// missing fields
					ctcs := []corruptedBodyTestcase{
						{name: "version/missing", corrupt: func(valid *protonetmap.LocalNodeInfoResponse_Body) { valid.Version = nil },
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, MissingResponseFieldErr{})
								require.EqualError(t, err, "missing version field in the response")
							}},
						{name: "node info/missing", corrupt: func(valid *protonetmap.LocalNodeInfoResponse_Body) { valid.NodeInfo = nil },
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, MissingResponseFieldErr{})
								require.EqualError(t, err, "missing node info field in the response")
							}},
					}
					// invalid node info
					for _, tc := range invalidNodeInfoProtoTestcases {
						ctcs = append(ctcs, corruptedBodyTestcase{
							name:    "node info/" + tc.name,
							corrupt: func(valid *protonetmap.LocalNodeInfoResponse_Body) { tc.corrupt(valid.NodeInfo) },
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid node info field in the response: "+tc.msg)
							},
						})
					}

					for _, tc := range ctcs {
						body := proto.Clone(validMinNodeInfoResponseBody).(*protonetmap.LocalNodeInfoResponse_Body)
						tc.corrupt(body)
						tcs = append(tcs, testcase{name: tc.name, body: body, assertErr: tc.assertErr})
					}

					testInvalidResponseBodies(t, newTestGetNodeInfoServer, newTestNetmapClient, tcs, func(c *Client) error {
						_, err := c.EndpointInfo(ctx, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestGetNodeInfoServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.EndpointInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.EndpointInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestGetNodeInfoServer, newTestNetmapClient, func(c *Client) error {
			_, err := c.EndpointInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestGetNodeInfoServer, newDefaultNetmapServiceDesc, func(c *Client) error {
			_, err := c.EndpointInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestGetNodeInfoServer, newDefaultNetmapServiceDesc, stat.MethodEndpointInfo,
			nil, nil, func(c *Client) error {
				_, err := c.EndpointInfo(ctx, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_NetMapSnapshot(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmNetMapSnapshot

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestNetmapSnapshotServer()
				c := newTestNetmapClient(t, srv)

				srv.authenticateRequest(c.prm.signer)
				_, err := c.NetMapSnapshot(ctx, PrmNetMapSnapshot{})
				require.NoError(t, err)
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protonetmap.NetmapSnapshotResponse_Body
					}{
						{name: "min", body: validMinNetmapResponseBody},
						{name: "full", body: validFullNetmapResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestNetmapSnapshotServer()
							c := newTestNetmapClient(t, srv)

							srv.respondWithBody(tc.body)
							res, err := c.NetMapSnapshot(ctx, anyValidOpts)
							require.NoError(t, err)
							require.NoError(t, checkNetmapTransport(res, tc.body.GetNetmap()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestNetmapSnapshotServer, newTestNetmapClient, func(c *Client) error {
						_, err := c.NetMapSnapshot(ctx, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "netmap.NetmapService", "NetmapSnapshot", func(c *Client) error {
						_, err := c.NetMapSnapshot(ctx, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestNetmapSnapshotServer, newTestNetmapClient, func(c *Client) error {
						_, err := c.NetMapSnapshot(ctx, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protonetmap.NetmapSnapshotResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "missing network map field in the response")
								// TODO: worth clarifying that body is completely missing
							}},
						{name: "empty", body: new(protonetmap.NetmapSnapshotResponse_Body),
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, ErrMissingResponseField)
								require.EqualError(t, err, "missing network map field in the response")
							}},
					}

					// 1. network map
					for _, tc := range invalidNodeInfoProtoTestcases {
						body := &protonetmap.NetmapSnapshotResponse_Body{
							Netmap: proto.Clone(validFullProtoNetmap).(*protonetmap.Netmap),
						}
						tc.corrupt(body.Netmap.Nodes[1])
						tcs = append(tcs, testcase{
							name: "network map/node info/" + tc.name,
							body: body,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid network map field in the response: invalid node info: "+tc.msg)
							},
						})
					}

					testInvalidResponseBodies(t, newTestNetmapSnapshotServer, newTestNetmapClient, tcs, func(c *Client) error {
						_, err := c.NetMapSnapshot(ctx, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestNetmapSnapshotServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.NetMapSnapshot(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.NetMapSnapshot(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestNetmapSnapshotServer, newTestNetmapClient, func(c *Client) error {
			_, err := c.NetMapSnapshot(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestNetmapSnapshotServer, newDefaultNetmapServiceDesc, func(c *Client) error {
			_, err := c.NetMapSnapshot(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestNetmapSnapshotServer, newDefaultNetmapServiceDesc, stat.MethodNetMapSnapshot,
			nil, nil, func(c *Client) error {
				_, err := c.NetMapSnapshot(ctx, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_NetworkInfo(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmNetworkInfo

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestNetworkInfoServer()
				c := newTestNetmapClient(t, srv)

				srv.authenticateRequest(c.prm.signer)
				_, err := c.NetworkInfo(ctx, anyValidOpts)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				testRequestXHeaders(t, newTestNetworkInfoServer, newTestNetmapClient, func(c *Client, xhs []string) error {
					opts := anyValidOpts
					opts.WithXHeaders(xhs...)
					_, err := c.NetworkInfo(ctx, opts)
					return err
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protonetmap.NetworkInfoResponse_Body
					}{
						{name: "min", body: validMinNetInfoResponseBody},
						{name: "full", body: validFullNetInfoResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestNetworkInfoServer()
							c := newTestNetmapClient(t, srv)

							srv.respondWithBody(tc.body)
							res, err := c.NetworkInfo(ctx, anyValidOpts)
							require.NoError(t, err)
							require.NoError(t, checkNetInfoTransport(res, tc.body.GetNetworkInfo()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestNetworkInfoServer, newTestNetmapClient, func(c *Client) error {
						_, err := c.NetworkInfo(ctx, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "netmap.NetmapService", "NetworkInfo", func(c *Client) error {
						_, err := c.NetworkInfo(ctx, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestNetworkInfoServer, newTestNetmapClient, func(c *Client) error {
						_, err := c.NetworkInfo(ctx, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protonetmap.NetworkInfoResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "missing network info field in the response")
								// TODO: worth clarifying that body is completely missing
							}},
						{name: "empty", body: new(protonetmap.NetworkInfoResponse_Body),
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, ErrMissingResponseField)
								require.EqualError(t, err, "missing network info field in the response")
							}},
					}

					// 1. net info
					for _, tc := range []struct {
						name, msg string
						corrupt   func(valid *protonetmap.NetworkInfo)
					}{} {
						body := &protonetmap.NetworkInfoResponse_Body{
							NetworkInfo: proto.Clone(validFullProtoNetInfo).(*protonetmap.NetworkInfo),
						}
						tc.corrupt(body.NetworkInfo)
						tcs = append(tcs, testcase{
							name: "network info/" + tc.name,
							body: body,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid network info field in the response: "+tc.msg)
							},
						})
					}

					testInvalidResponseBodies(t, newTestNetworkInfoServer, newTestNetmapClient, tcs, func(c *Client) error {
						_, err := c.NetworkInfo(ctx, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestNetworkInfoServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.NetworkInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.NetworkInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestNetworkInfoServer, newTestNetmapClient, func(c *Client) error {
			_, err := c.NetworkInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestNetworkInfoServer, newDefaultNetmapServiceDesc, func(c *Client) error {
			_, err := c.NetworkInfo(ctx, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestNetworkInfoServer, newDefaultNetmapServiceDesc, stat.MethodNetworkInfo,
			nil, nil, func(c *Client) error {
				_, err := c.NetworkInfo(ctx, anyValidOpts)
				return err
			},
		)
	})
}
