package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type noOtherNetmapCalls struct{}

func (noOtherNetmapCalls) LocalNodeInfo(context.Context, *apinetmap.LocalNodeInfoRequest) (*apinetmap.LocalNodeInfoResponse, error) {
	panic("must not be called")
}

func (noOtherNetmapCalls) NetworkInfo(context.Context, *apinetmap.NetworkInfoRequest) (*apinetmap.NetworkInfoResponse, error) {
	panic("must not be called")
}

func (noOtherNetmapCalls) NetmapSnapshot(context.Context, *apinetmap.NetmapSnapshotRequest) (*apinetmap.NetmapSnapshotResponse, error) {
	panic("must not be called")
}

// implements [apinetmap.NetmapServiceServer] with simplified LocalNodeInfo to
// be used on [Client.Dial] while testing other methods.
type endpointInfoOnDialServer struct {
	noOtherNetmapCalls
	epoch         uint64
	serverSigner  neofscrypto.Signer
	latestVersion version.Version
	nodeInfo      netmap.NodeInfo
}

func (x endpointInfoOnDialServer) LocalNodeInfo(_ context.Context, _ *apinetmap.LocalNodeInfoRequest) (*apinetmap.LocalNodeInfoResponse, error) {
	resp := apinetmap.LocalNodeInfoResponse{
		Body: &apinetmap.LocalNodeInfoResponse_Body{
			Version:  new(refs.Version),
			NodeInfo: new(apinetmap.NodeInfo),
		},
		MetaHeader: &apisession.ResponseMetaHeader{
			Epoch: x.epoch,
		},
	}
	x.latestVersion.WriteToV2(resp.Body.Version)
	x.nodeInfo.WriteToV2(resp.Body.NodeInfo)
	var err error
	resp.VerifyHeader, err = neofscrypto.SignResponse(x.serverSigner, &resp, resp.Body, nil)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return &resp, nil
}

type endpointInfoServer struct {
	// client
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	called   bool // to distinguish on-dial call
	sleepDur time.Duration
	endpointInfoOnDialServer
	errTransport   error
	modifyResp     func(*apinetmap.LocalNodeInfoResponse)
	corruptRespSig func(*apinetmap.LocalNodeInfoResponse)
}

func (x *endpointInfoServer) LocalNodeInfo(ctx context.Context, req *apinetmap.LocalNodeInfoRequest) (*apinetmap.LocalNodeInfoResponse, error) {
	defer func() { x.called = true }()
	if !x.called {
		return x.endpointInfoOnDialServer.LocalNodeInfo(ctx, req)
	}
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apinetmap.LocalNodeInfoResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	if ctx == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil context"
	} else if req == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil request"
	} else if err = neofscrypto.VerifyRequest(req, req.Body); err != nil {
		sts.Code, sts.Message = status.SignatureVerificationFail, err.Error()
	} else if req.VerifyHeader.BodySignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.BodySignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request body signature credentials"
	} else if req.VerifyHeader.MetaSignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.MetaSignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request meta header signature credentials"
	} else if req.VerifyHeader.OriginSignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.OriginSignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected origin request verification header signature credentials"
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	} else {
		resp.MetaHeader.Status = nil
		resp.Body = &apinetmap.LocalNodeInfoResponse_Body{
			Version:  new(refs.Version),
			NodeInfo: new(apinetmap.NodeInfo),
		}
		x.latestVersion.WriteToV2(resp.Body.Version)
		x.nodeInfo.WriteToV2(resp.Body.NodeInfo)
	}
	if x.modifyResp != nil {
		x.modifyResp(&resp)
	}
	resp.VerifyHeader, err = neofscrypto.SignResponse(x.serverSigner, &resp, resp.Body, nil)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	if x.corruptRespSig != nil {
		x.corruptRespSig(&resp)
	}
	return &resp, nil
}

func TestClient_GetEndpointInfo(t *testing.T) {
	ctx := context.Background()
	var srv endpointInfoServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	_dial := func(t testing.TB, srv *endpointInfoServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodEndpointInfo, op)
			require.Greater(t, dur, srv.sleepDur)
			assertErr(err)
		})
		if customizeOpts != nil {
			customizeOpts(&opts)
		}

		c, err := New(anyValidURI, opts)
		require.NoError(t, err)
		srv.clientSigScheme = c.signer.Scheme()
		srv.clientPubKey = neofscrypto.PublicKeyBytes(c.signer.Public())

		conn := bufconn.Listen(10 << 10)
		gs := grpc.NewServer()
		apinetmap.RegisterNetmapServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *endpointInfoServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
		assertErr(err)
		require.Equal(t, srv.latestVersion, res.LatestVersion)
		require.Equal(t, srv.nodeInfo, res.Node)
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("transport", func(t *testing.T) {
			srv := srv
			srv.errTransport = errors.New("any transport failure")
			assertErr := func(err error) {
				require.ErrorContains(t, err, errTransport)
				require.ErrorContains(t, err, "any transport failure")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apinetmap.LocalNodeInfoResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apinetmap.LocalNodeInfoResponse) {
						r.VerifyHeader.OriginSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
			} {
				srv := srv
				srv.corruptRespSig = testCase.corrupt
				assertErr := func(err error) {
					require.ErrorContains(t, err, errResponseSignature, [2]any{i, testCase})
					require.ErrorContains(t, err, testCase.err, [2]any{i, testCase})
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apinetmap.LocalNodeInfoResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("status errors", func(t *testing.T) {
			for _, testCase := range []struct {
				code     uint32
				errConst error
				errVar   any
			}{
				{code: 1 << 32 / 2},
				{code: status.InternalServerError, errConst: apistatus.ErrServerInternal, errVar: new(apistatus.InternalServerError)},
				{code: status.SignatureVerificationFail, errConst: apistatus.ErrSignatureVerification, errVar: new(apistatus.SignatureVerificationFailure)},
			} {
				srv := srv
				srv.modifyResp = func(r *apinetmap.LocalNodeInfoResponse) {
					r.MetaHeader.Status = &status.Status{Code: testCase.code, Message: "any message"}
				}
				assertErr := func(err error) {
					require.ErrorIs(t, err, apistatus.Error, testCase)
					require.ErrorContains(t, err, "any message", testCase)
					if testCase.errConst != nil {
						require.ErrorIs(t, err, testCase.errConst, testCase)
					}
					if testCase.errVar != nil {
						require.ErrorAs(t, err, testCase.errVar, testCase)
					}
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apinetmap.LocalNodeInfoResponse) { r.Body = nil }
				_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing version", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apinetmap.LocalNodeInfoResponse) { r.Body.Version = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (version)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing node info", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apinetmap.LocalNodeInfoResponse) { r.Body.NodeInfo = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (node info)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid node info", func(t *testing.T) {
				for _, testCase := range []struct {
					name     string
					err      string
					contains bool
					corrupt  func(info *apinetmap.NodeInfo)
				}{
					{name: "nil public key", err: "missing public key",
						corrupt: func(n *apinetmap.NodeInfo) { n.PublicKey = nil }},
					{name: "empty public key", err: "missing public key",
						corrupt: func(n *apinetmap.NodeInfo) { n.PublicKey = nil }},
					{name: "nil addresses", err: "missing network endpoints",
						corrupt: func(n *apinetmap.NodeInfo) { n.Addresses = nil }},
					{name: "empty addresses", err: "missing network endpoints",
						corrupt: func(n *apinetmap.NodeInfo) { n.Addresses = []string{} }},
					{name: "empty address", err: "empty network endpoint #1",
						corrupt: func(n *apinetmap.NodeInfo) { n.Addresses = []string{"any", "", "any"} }},
					{name: "attributes/missing key", err: "invalid attribute #1: missing key",
						corrupt: func(n *apinetmap.NodeInfo) {
							n.Attributes = []*apinetmap.NodeInfo_Attribute{
								{Key: "key_valid", Value: "any"},
								{Key: "", Value: "any"},
							}
						}},
					{name: "attributes/repeated keys", err: "multiple attributes with key=k2",
						corrupt: func(n *apinetmap.NodeInfo) {
							n.Attributes = []*apinetmap.NodeInfo_Attribute{
								{Key: "k1", Value: "any"},
								{Key: "k2", Value: "1"},
								{Key: "k3", Value: "any"},
								{Key: "k2", Value: "2"},
							}
						}},
					{name: "attributes/missing value", err: "invalid attribute #1 (key2): missing value",
						corrupt: func(n *apinetmap.NodeInfo) {
							n.Attributes = []*apinetmap.NodeInfo_Attribute{
								{Key: "key1", Value: "any"},
								{Key: "key2", Value: ""},
							}
						}},
					{name: "attributes/price", err: "invalid price attribute (#1): invalid integer", contains: true,
						corrupt: func(n *apinetmap.NodeInfo) {
							n.Attributes = []*apinetmap.NodeInfo_Attribute{
								{Key: "any", Value: "any"},
								{Key: "Price", Value: "not_a_number"},
							}
						}},
					{name: "attributes/capacity", err: "invalid capacity attribute (#1): invalid integer", contains: true,
						corrupt: func(n *apinetmap.NodeInfo) {
							n.Attributes = []*apinetmap.NodeInfo_Attribute{
								{Key: "any", Value: "any"},
								{Key: "Capacity", Value: "not_a_number"},
							}
						}},
				} {
					srv := srv
					srv.modifyResp = func(r *apinetmap.LocalNodeInfoResponse) { testCase.corrupt(r.Body.NodeInfo) }
					assertErr := func(err error) {
						if testCase.contains {
							require.ErrorContains(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (node info): %s", testCase.err))
						} else {
							require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (node info): %s", testCase.err))
						}
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
					assertErr(err)
					require.True(t, *handlerCalled, testCase)
				}
			})
		})
	})
	t.Run("response info handler", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			srv := srv
			srv.epoch = 3598503
			assertErr := func(err error) { require.NoError(t, err) }
			respHandlerCalled := false
			c, reqHandlerCalled := _dial(t, &srv, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					respHandlerCalled = true
					require.EqualValues(t, 3598503, info.Epoch())
					require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), info.ResponderKey())
					return nil
				})
			})
			_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			srv := srv
			srv.epoch = 4386380643
			assertErr := func(err error) { require.ErrorContains(t, err, "intercept response info: some handler error") }
			respHandlerCalled := false
			c, reqHandlerCalled := _dial(t, &srv, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					if !respHandlerCalled { // dial
						respHandlerCalled = true
						return nil
					}
					require.EqualValues(t, 4386380643, info.Epoch())
					require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), info.ResponderKey())
					return errors.New("some handler error")
				})
			})
			_, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type networkInfoServer struct {
	// client
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	called   bool // to distinguish on-dial call
	sleepDur time.Duration
	endpointInfoOnDialServer
	netInfo        netmap.NetworkInfo
	errTransport   error
	modifyResp     func(*apinetmap.NetworkInfoResponse)
	corruptRespSig func(*apinetmap.NetworkInfoResponse)
}

func (x *networkInfoServer) NetworkInfo(ctx context.Context, req *apinetmap.NetworkInfoRequest) (*apinetmap.NetworkInfoResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apinetmap.NetworkInfoResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	if ctx == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil context"
	} else if req == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil request"
	} else if err = neofscrypto.VerifyRequest(req, req.Body); err != nil {
		sts.Code, sts.Message = status.SignatureVerificationFail, err.Error()
	} else if req.VerifyHeader.BodySignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.BodySignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request body signature credentials"
	} else if req.VerifyHeader.MetaSignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.MetaSignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request meta header signature credentials"
	} else if req.VerifyHeader.OriginSignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.OriginSignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected origin request verification header signature credentials"
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	} else {
		resp.MetaHeader.Status = nil
		resp.Body = &apinetmap.NetworkInfoResponse_Body{
			NetworkInfo: new(apinetmap.NetworkInfo),
		}
		x.netInfo.WriteToV2(resp.Body.NetworkInfo)
	}
	if x.modifyResp != nil {
		x.modifyResp(&resp)
	}
	resp.VerifyHeader, err = neofscrypto.SignResponse(x.serverSigner, &resp, resp.Body, nil)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	if x.corruptRespSig != nil {
		x.corruptRespSig(&resp)
	}
	return &resp, nil
}

func TestClient_GetNetworkInfo(t *testing.T) {
	ctx := context.Background()
	var srv networkInfoServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.netInfo = netmaptest.NetworkInfo()
	_dial := func(t testing.TB, srv *networkInfoServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodNetworkInfo, op)
			require.Greater(t, dur, srv.sleepDur)
			assertErr(err)
		})
		if customizeOpts != nil {
			customizeOpts(&opts)
		}

		c, err := New(anyValidURI, opts)
		require.NoError(t, err)
		srv.clientSigScheme = c.signer.Scheme()
		srv.clientPubKey = neofscrypto.PublicKeyBytes(c.signer.Public())

		conn := bufconn.Listen(10 << 10)
		gs := grpc.NewServer()
		apinetmap.RegisterNetmapServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *networkInfoServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
		assertErr(err)
		if !assert.ObjectsAreEqual(srv.netInfo, res) {
			// can be caused by gRPC service fields, binaries must still be equal
			require.Equal(t, srv.netInfo.Marshal(), res.Marshal())
		}
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("transport", func(t *testing.T) {
			srv := srv
			srv.errTransport = errors.New("any transport failure")
			assertErr := func(err error) {
				require.ErrorContains(t, err, errTransport)
				require.ErrorContains(t, err, "any transport failure")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apinetmap.NetworkInfoResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apinetmap.NetworkInfoResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apinetmap.NetworkInfoResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apinetmap.NetworkInfoResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apinetmap.NetworkInfoResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apinetmap.NetworkInfoResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apinetmap.NetworkInfoResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apinetmap.NetworkInfoResponse) {
						r.VerifyHeader.OriginSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
			} {
				srv := srv
				srv.corruptRespSig = testCase.corrupt
				assertErr := func(err error) {
					require.ErrorContains(t, err, errResponseSignature, [2]any{i, testCase})
					require.ErrorContains(t, err, testCase.err, [2]any{i, testCase})
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apinetmap.NetworkInfoResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("status errors", func(t *testing.T) {
			for _, testCase := range []struct {
				code     uint32
				errConst error
				errVar   any
			}{
				{code: 1 << 32 / 2},
				{code: status.InternalServerError, errConst: apistatus.ErrServerInternal, errVar: new(apistatus.InternalServerError)},
				{code: status.SignatureVerificationFail, errConst: apistatus.ErrSignatureVerification, errVar: new(apistatus.SignatureVerificationFailure)},
			} {
				srv := srv
				srv.modifyResp = func(r *apinetmap.NetworkInfoResponse) {
					r.MetaHeader.Status = &status.Status{Code: testCase.code, Message: "any message"}
				}
				assertErr := func(err error) {
					require.ErrorIs(t, err, apistatus.Error, testCase)
					require.ErrorContains(t, err, "any message", testCase)
					if testCase.errConst != nil {
						require.ErrorIs(t, err, testCase.errConst, testCase)
					}
					if testCase.errVar != nil {
						require.ErrorAs(t, err, testCase.errVar, testCase)
					}
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apinetmap.NetworkInfoResponse) { r.Body = nil }
				_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing network info", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apinetmap.NetworkInfoResponse) { r.Body.NetworkInfo = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (network info)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid network info", func(t *testing.T) {
				testCases := []struct {
					name     string
					err      string
					contains bool
					prm      apinetmap.NetworkConfig_Parameter
				}{
					{name: "nil key", err: "invalid network parameter #1: missing name", prm: apinetmap.NetworkConfig_Parameter{
						Key: nil, Value: []byte("any"),
					}},
					{name: "empty key", err: "invalid network parameter #1: missing name", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte{}, Value: []byte("any"),
					}},
					{name: "nil value", err: "invalid network parameter #1: missing value", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("any"), Value: nil,
					}},
					{name: "repeated keys", err: "multiple network parameters with name=any_key", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("any_key"), Value: []byte("any"),
					}},
					{name: "audit fee format", err: "invalid network parameter #1 (AuditFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("AuditFee"), Value: []byte("Hello, world!"),
					}},
					{name: "storage price format", err: "invalid network parameter #1 (BasicIncomeRate): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("BasicIncomeRate"), Value: []byte("Hello, world!"),
					}},
					{name: "container fee format", err: "invalid network parameter #1 (ContainerFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("ContainerFee"), Value: []byte("Hello, world!"),
					}},
					{name: "named container fee format", err: "invalid network parameter #1 (ContainerAliasFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("ContainerAliasFee"), Value: []byte("Hello, world!"),
					}},
					{name: "num of EigenTrust iterations format", err: "invalid network parameter #1 (EigenTrustIterations): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("EigenTrustIterations"), Value: []byte("Hello, world!"),
					}},
					{name: "epoch duration format", err: "invalid network parameter #1 (EpochDuration): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("EpochDuration"), Value: []byte("Hello, world!"),
					}},
					{name: "IR candidate fee format", err: "invalid network parameter #1 (InnerRingCandidateFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("InnerRingCandidateFee"), Value: []byte("Hello, world!"),
					}},
					{name: "max object size format", err: "invalid network parameter #1 (MaxObjectSize): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("MaxObjectSize"), Value: []byte("Hello, world!"),
					}},
					{name: "withdrawal fee format", err: "invalid network parameter #1 (WithdrawFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("WithdrawFee"), Value: []byte("Hello, world!"),
					}},
					{name: "EigenTrust alpha format", err: "invalid network parameter #1 (EigenTrustAlpha): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("EigenTrustAlpha"), Value: []byte("Hello, world!"),
					}},
					{name: "negative EigenTrust alpha", err: "invalid network parameter #1 (EigenTrustAlpha): EigenTrust alpha value -3.14 is out of range [0, 1]", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("EigenTrustAlpha"), Value: []byte{31, 133, 235, 81, 184, 30, 9, 192},
					}},
					{name: "negative EigenTrust alpha", err: "invalid network parameter #1 (EigenTrustAlpha): EigenTrust alpha value 1.10 is out of range [0, 1]", prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("EigenTrustAlpha"), Value: []byte{154, 153, 153, 153, 153, 153, 241, 63},
					}},
					{name: "disable homomorphic hashing format", err: "invalid network parameter #1 (HomomorphicHashingDisabled): invalid bool parameter", contains: true, prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("HomomorphicHashingDisabled"), Value: make([]byte, 32+1), // max 32
					}},
					{name: "allow maintenance mode format", err: "invalid network parameter #1 (MaintenanceModeAllowed): invalid bool parameter", contains: true, prm: apinetmap.NetworkConfig_Parameter{
						Key: []byte("MaintenanceModeAllowed"), Value: make([]byte, 32+1), // max 32
					}},
				}
				for i := range testCases {
					srv := srv
					srv.modifyResp = func(r *apinetmap.NetworkInfoResponse) {
						r.Body.NetworkInfo.NetworkConfig.Parameters = []*apinetmap.NetworkConfig_Parameter{
							{Key: []byte("any_key"), Value: []byte("any_val")},
							&testCases[i].prm,
						}
					}
					assertErr := func(err error) {
						if testCases[i].contains {
							require.ErrorContains(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (network info): %s", testCases[i].err))
						} else {
							require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (network info): %s", testCases[i].err))
						}
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
					assertErr(err)
					require.True(t, *handlerCalled, testCases[i].name)
				}
			})
		})
	})
	t.Run("response info handler", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			srv := srv
			srv.epoch = 3598503
			assertErr := func(err error) { require.NoError(t, err) }
			respHandlerCalled := false
			c, reqHandlerCalled := _dial(t, &srv, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					respHandlerCalled = true
					require.EqualValues(t, 3598503, info.Epoch())
					require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), info.ResponderKey())
					return nil
				})
			})
			_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			srv := srv
			srv.epoch = 4386380643
			assertErr := func(err error) { require.ErrorContains(t, err, "intercept response info: some handler error") }
			respHandlerCalled := false
			c, reqHandlerCalled := _dial(t, &srv, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					if !respHandlerCalled { // dial
						respHandlerCalled = true
						return nil
					}
					require.EqualValues(t, 4386380643, info.Epoch())
					require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), info.ResponderKey())
					return errors.New("some handler error")
				})
			})
			_, err := c.GetNetworkInfo(ctx, GetNetworkInfoOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type currentNetmapServer struct {
	// client
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	called   bool // to distinguish on-dial call
	sleepDur time.Duration
	endpointInfoOnDialServer
	curNetmap      netmap.NetMap
	errTransport   error
	modifyResp     func(*apinetmap.NetmapSnapshotResponse)
	corruptRespSig func(*apinetmap.NetmapSnapshotResponse)
}

func (x *currentNetmapServer) NetmapSnapshot(ctx context.Context, req *apinetmap.NetmapSnapshotRequest) (*apinetmap.NetmapSnapshotResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apinetmap.NetmapSnapshotResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	if ctx == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil context"
	} else if req == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil request"
	} else if err = neofscrypto.VerifyRequest(req, req.Body); err != nil {
		sts.Code, sts.Message = status.SignatureVerificationFail, err.Error()
	} else if req.VerifyHeader.BodySignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.BodySignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request body signature credentials"
	} else if req.VerifyHeader.MetaSignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.MetaSignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request meta header signature credentials"
	} else if req.VerifyHeader.OriginSignature.Scheme != refs.SignatureScheme(x.clientSigScheme) ||
		!bytes.Equal(req.VerifyHeader.OriginSignature.Key, x.clientPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected origin request verification header signature credentials"
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	} else {
		resp.MetaHeader.Status = nil
		resp.Body = &apinetmap.NetmapSnapshotResponse_Body{
			Netmap: new(apinetmap.Netmap),
		}
		x.curNetmap.WriteToV2(resp.Body.Netmap)
	}
	if x.modifyResp != nil {
		x.modifyResp(&resp)
	}
	resp.VerifyHeader, err = neofscrypto.SignResponse(x.serverSigner, &resp, resp.Body, nil)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	if x.corruptRespSig != nil {
		x.corruptRespSig(&resp)
	}
	return &resp, nil
}

func TestClient_GetCurrentNetmap(t *testing.T) {
	ctx := context.Background()
	var srv currentNetmapServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.curNetmap = netmaptest.Netmap()
	_dial := func(t testing.TB, srv *currentNetmapServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodNetMapSnapshot, op)
			require.Greater(t, dur, srv.sleepDur)
			assertErr(err)
		})
		if customizeOpts != nil {
			customizeOpts(&opts)
		}

		c, err := New(anyValidURI, opts)
		require.NoError(t, err)
		srv.clientSigScheme = c.signer.Scheme()
		srv.clientPubKey = neofscrypto.PublicKeyBytes(c.signer.Public())

		conn := bufconn.Listen(10 << 10)
		gs := grpc.NewServer()
		apinetmap.RegisterNetmapServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *currentNetmapServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
		assertErr(err)
		require.EqualValues(t, srv.curNetmap, res)
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("transport", func(t *testing.T) {
			srv := srv
			srv.errTransport = errors.New("any transport failure")
			assertErr := func(err error) {
				require.ErrorContains(t, err, errTransport)
				require.ErrorContains(t, err, "any transport failure")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(request *apinetmap.NetmapSnapshotResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apinetmap.NetmapSnapshotResponse) {
						r.VerifyHeader.OriginSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
			} {
				srv := srv
				srv.corruptRespSig = testCase.corrupt
				assertErr := func(err error) {
					require.ErrorContains(t, err, errResponseSignature, [2]any{i, testCase})
					require.ErrorContains(t, err, testCase.err, [2]any{i, testCase})
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apinetmap.NetmapSnapshotResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("status errors", func(t *testing.T) {
			for _, testCase := range []struct {
				code     uint32
				errConst error
				errVar   any
			}{
				{code: 1 << 32 / 2},
				{code: status.InternalServerError, errConst: apistatus.ErrServerInternal, errVar: new(apistatus.InternalServerError)},
				{code: status.SignatureVerificationFail, errConst: apistatus.ErrSignatureVerification, errVar: new(apistatus.SignatureVerificationFailure)},
			} {
				srv := srv
				srv.modifyResp = func(r *apinetmap.NetmapSnapshotResponse) {
					r.MetaHeader.Status = &status.Status{Code: testCase.code, Message: "any message"}
				}
				assertErr := func(err error) {
					require.ErrorIs(t, err, apistatus.Error, testCase)
					require.ErrorContains(t, err, "any message", testCase)
					if testCase.errConst != nil {
						require.ErrorIs(t, err, testCase.errConst, testCase)
					}
					if testCase.errVar != nil {
						require.ErrorAs(t, err, testCase.errVar, testCase)
					}
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apinetmap.NetmapSnapshotResponse) { r.Body = nil }
				_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing network map", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apinetmap.NetmapSnapshotResponse) { r.Body.Netmap = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (network map)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid network map", func(t *testing.T) {
				t.Run("invalid node", func(t *testing.T) {
					for _, testCase := range []struct {
						name     string
						err      string
						contains bool
						corrupt  func(*apinetmap.NodeInfo)
					}{
						{name: "nil public key", err: "missing public key",
							corrupt: func(n *apinetmap.NodeInfo) { n.PublicKey = nil }},
						{name: "empty public key", err: "missing public key",
							corrupt: func(n *apinetmap.NodeInfo) { n.PublicKey = nil }},
						{name: "nil addresses", err: "missing network endpoints",
							corrupt: func(n *apinetmap.NodeInfo) { n.Addresses = nil }},
						{name: "empty addresses", err: "missing network endpoints",
							corrupt: func(n *apinetmap.NodeInfo) { n.Addresses = []string{} }},
						{name: "empty address", err: "empty network endpoint #1",
							corrupt: func(n *apinetmap.NodeInfo) { n.Addresses = []string{"any", "", "any"} }},
						{name: "attributes/missing key", err: "invalid attribute #1: missing key",
							corrupt: func(n *apinetmap.NodeInfo) {
								n.Attributes = []*apinetmap.NodeInfo_Attribute{
									{Key: "key_valid", Value: "any"},
									{Key: "", Value: "any"},
								}
							}},
						{name: "attributes/repeated keys", err: "multiple attributes with key=k2",
							corrupt: func(n *apinetmap.NodeInfo) {
								n.Attributes = []*apinetmap.NodeInfo_Attribute{
									{Key: "k1", Value: "any"},
									{Key: "k2", Value: "1"},
									{Key: "k3", Value: "any"},
									{Key: "k2", Value: "2"},
								}
							}},
						{name: "attributes/missing value", err: "invalid attribute #1 (key2): missing value",
							corrupt: func(n *apinetmap.NodeInfo) {
								n.Attributes = []*apinetmap.NodeInfo_Attribute{
									{Key: "key1", Value: "any"},
									{Key: "key2", Value: ""},
								}
							}},
						{name: "attributes/price", err: "invalid price attribute (#1): invalid integer", contains: true,
							corrupt: func(n *apinetmap.NodeInfo) {
								n.Attributes = []*apinetmap.NodeInfo_Attribute{
									{Key: "any", Value: "any"},
									{Key: "Price", Value: "not_a_number"},
								}
							}},
						{name: "attributes/capacity", err: "invalid capacity attribute (#1): invalid integer", contains: true,
							corrupt: func(n *apinetmap.NodeInfo) {
								n.Attributes = []*apinetmap.NodeInfo_Attribute{
									{Key: "any", Value: "any"},
									{Key: "Capacity", Value: "not_a_number"},
								}
							}},
					} {
						srv := srv
						srv.modifyResp = func(r *apinetmap.NetmapSnapshotResponse) {
							r.Body.Netmap.Nodes = make([]*apinetmap.NodeInfo, 2)
							for i := range r.Body.Netmap.Nodes {
								r.Body.Netmap.Nodes[i] = new(apinetmap.NodeInfo)
								netmaptest.NodeInfo().WriteToV2(r.Body.Netmap.Nodes[i])
							}
							testCase.corrupt(r.Body.Netmap.Nodes[1])
						}
						assertErr := func(err error) {
							if testCase.contains {
								require.ErrorContains(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (network map): invalid node info #1: %s", testCase.err))
							} else {
								require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (network map): invalid node info #1: %s", testCase.err))
							}
						}
						c, handlerCalled := dial(t, &srv, assertErr)
						_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
						assertErr(err)
						require.True(t, *handlerCalled, testCase)
					}
				})
			})
		})
	})
	t.Run("response info handler", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			srv := srv
			srv.epoch = 3598503
			assertErr := func(err error) { require.NoError(t, err) }
			respHandlerCalled := false
			c, reqHandlerCalled := _dial(t, &srv, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					respHandlerCalled = true
					require.EqualValues(t, 3598503, info.Epoch())
					require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), info.ResponderKey())
					return nil
				})
			})
			_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			srv := srv
			srv.epoch = 4386380643
			assertErr := func(err error) { require.ErrorContains(t, err, "intercept response info: some handler error") }
			respHandlerCalled := false
			c, reqHandlerCalled := _dial(t, &srv, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					if !respHandlerCalled { // dial
						respHandlerCalled = true
						return nil
					}
					require.EqualValues(t, 4386380643, info.Epoch())
					require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), info.ResponderKey())
					return errors.New("some handler error")
				})
			})
			_, err := c.GetCurrentNetmap(ctx, GetCurrentNetmapOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
