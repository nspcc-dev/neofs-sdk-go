package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apireputation "github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type noOtherReputationCalls struct{}

func (noOtherReputationCalls) AnnounceIntermediateResult(context.Context, *apireputation.AnnounceIntermediateResultRequest) (*apireputation.AnnounceIntermediateResultResponse, error) {
	panic("must not be called")
}

func (x noOtherReputationCalls) AnnounceLocalTrust(context.Context, *apireputation.AnnounceLocalTrustRequest) (*apireputation.AnnounceLocalTrustResponse, error) {
	panic("must not be called")
}

type sendLocalTrustsServer struct {
	noOtherReputationCalls
	// client
	epoch           uint64
	trusts          []reputation.Trust
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	errTransport   error
	modifyResp     func(*apireputation.AnnounceLocalTrustResponse)
	corruptRespSig func(*apireputation.AnnounceLocalTrustResponse)
}

func (x sendLocalTrustsServer) AnnounceLocalTrust(ctx context.Context, req *apireputation.AnnounceLocalTrustRequest) (*apireputation.AnnounceLocalTrustResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apireputation.AnnounceLocalTrustResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.endpointInfoOnDialServer.epoch},
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
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.Epoch != x.epoch {
		sts.Code, sts.Message = status.InternalServerError, "[test] invalid request: invalid body: wrong epoch"
	} else if len(req.Body.Trusts) == 0 {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing trusts"
	} else {
		var tr reputation.Trust
		for i := range req.Body.Trusts {
			if req.Body.Trusts[i] == nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: nil trust #%d", i)
			} else if err = tr.ReadFromV2(req.Body.Trusts[i]); err != nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid trust #%d: %v", i, err)
			} else if tr != x.trusts[i] {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] invalid request: invalid body: wrong trust #%d", i)
			}
		}
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

func TestClient_SendLocalTrusts(t *testing.T) {
	ctx := context.Background()
	var srv sendLocalTrustsServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.epoch = rand.Uint64()
	srv.trusts = reputationtest.NTrusts(3)
	_dial := func(t testing.TB, srv *sendLocalTrustsServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodAnnounceLocalTrust, op)
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
		apireputation.RegisterReputationServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *sendLocalTrustsServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("missing trusts", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		err = c.SendLocalTrusts(ctx, srv.epoch, nil, SendLocalTrustsOptions{})
		require.EqualError(t, err, "missing trusts")
		err = c.SendLocalTrusts(ctx, srv.epoch, []reputation.Trust{}, SendLocalTrustsOptions{})
		require.EqualError(t, err, "missing trusts")
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
		assertErr(err)
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
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
			err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apireputation.AnnounceLocalTrustResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceLocalTrustResponse) {
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
				err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apireputation.AnnounceLocalTrustResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
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
				srv.modifyResp = func(r *apireputation.AnnounceLocalTrustResponse) {
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
				err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
	})
	t.Run("response info handler", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			srv := srv
			srv.endpointInfoOnDialServer.epoch = 3598503
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
			err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			srv := srv
			srv.endpointInfoOnDialServer.epoch = 4386380643
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
			err := c.SendLocalTrusts(ctx, srv.epoch, srv.trusts, SendLocalTrustsOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type sendIntermediateTrustServer struct {
	noOtherReputationCalls
	// client
	epoch           uint64
	iter            uint32
	trust           reputation.PeerToPeerTrust
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	errTransport   error
	modifyResp     func(*apireputation.AnnounceIntermediateResultResponse)
	corruptRespSig func(*apireputation.AnnounceIntermediateResultResponse)
}

func (x sendIntermediateTrustServer) AnnounceIntermediateResult(ctx context.Context, req *apireputation.AnnounceIntermediateResultRequest) (*apireputation.AnnounceIntermediateResultResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apireputation.AnnounceIntermediateResultResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.endpointInfoOnDialServer.epoch},
	}
	var err error
	var tr reputation.PeerToPeerTrust
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
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.Epoch != x.epoch {
		sts.Code, sts.Message = status.InternalServerError, "[test] invalid request: invalid body: wrong epoch"
	} else if req.Body.Iteration != x.iter {
		sts.Code, sts.Message = status.InternalServerError, "[test] invalid request: invalid body: wrong iteration"
	} else if req.Body.Trust == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing trust"
	} else if err = tr.ReadFromV2(req.Body.Trust); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] invalid request: invalid body: invalid trust: %v", err)
	} else if tr != x.trust {
		sts.Code, sts.Message = status.InternalServerError, "[test] invalid request: invalid body: wrong trust"
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

func TestClient_SendIntermediateTrust(t *testing.T) {
	ctx := context.Background()
	var srv sendIntermediateTrustServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.epoch = rand.Uint64()
	srv.iter = rand.Uint32()
	srv.trust = reputationtest.PeerToPeerTrust()
	_dial := func(t testing.TB, srv *sendIntermediateTrustServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodAnnounceIntermediateTrust, op)
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
		apireputation.RegisterReputationServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *sendIntermediateTrustServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("missing trusts", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		err = c.SendLocalTrusts(ctx, srv.epoch, nil, SendLocalTrustsOptions{})
		require.EqualError(t, err, "missing trusts")
		err = c.SendLocalTrusts(ctx, srv.epoch, []reputation.Trust{}, SendLocalTrustsOptions{})
		require.EqualError(t, err, "missing trusts")
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
		assertErr(err)
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
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
			err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apireputation.AnnounceIntermediateResultResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apireputation.AnnounceIntermediateResultResponse) {
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
				err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apireputation.AnnounceIntermediateResultResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
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
				srv.modifyResp = func(r *apireputation.AnnounceIntermediateResultResponse) {
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
				err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
	})
	t.Run("response info handler", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			srv := srv
			srv.endpointInfoOnDialServer.epoch = 3598503
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
			err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			srv := srv
			srv.endpointInfoOnDialServer.epoch = 4386380643
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
			err := c.SendIntermediateTrust(ctx, srv.epoch, srv.iter, srv.trust, SendIntermediateTrustOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
