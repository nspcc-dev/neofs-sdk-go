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

	"github.com/google/uuid"
	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type noOtherSessionCalls struct{}

func (noOtherSessionCalls) Create(context.Context, *apisession.CreateRequest) (*apisession.CreateResponse, error) {
	panic("must not be called")
}

type startSessionServer struct {
	noOtherSessionCalls
	// client
	issuer user.Signer
	exp    uint64
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	id             uuid.UUID
	sessionPubKey  neofscrypto.PublicKey
	errTransport   error
	modifyResp     func(*apisession.CreateResponse)
	corruptRespSig func(*apisession.CreateResponse)
}

func (x startSessionServer) Create(ctx context.Context, req *apisession.CreateRequest) (*apisession.CreateResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apisession.CreateResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var usr user.ID
	sigScheme := refs.SignatureScheme(x.issuer.Scheme())
	creatorPubKey := neofscrypto.PublicKeyBytes(x.issuer.Public())
	if ctx == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil context"
	} else if req == nil {
		sts.Code, sts.Message = status.InternalServerError, "nil request"
	} else if err = neofscrypto.VerifyRequest(req, req.Body); err != nil {
		sts.Code, sts.Message = status.SignatureVerificationFail, err.Error()
	} else if req.VerifyHeader.BodySignature.Scheme != sigScheme ||
		!bytes.Equal(req.VerifyHeader.BodySignature.Key, creatorPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request body signature credentials"
	} else if req.VerifyHeader.MetaSignature.Scheme != sigScheme ||
		!bytes.Equal(req.VerifyHeader.MetaSignature.Key, creatorPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected request meta header signature credentials"
	} else if req.VerifyHeader.OriginSignature.Scheme != sigScheme ||
		!bytes.Equal(req.VerifyHeader.OriginSignature.Key, creatorPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] unexpected origin request verification header signature credentials"
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.Expiration != x.exp {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong expiration timestamp"
	} else if req.Body.OwnerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing issuer"
	} else if err = usr.ReadFromV2(req.Body.OwnerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid issuer: %s", err)
	} else if usr != x.issuer.UserID() {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong issuer"
	} else {
		resp.Body = &apisession.CreateResponse_Body{
			Id:         x.id[:],
			SessionKey: neofscrypto.PublicKeyBytes(x.sessionPubKey),
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

func TestClient_StartSession(t *testing.T) {
	ctx := context.Background()
	var srv startSessionServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.issuer, _ = usertest.TwoUsers()
	srv.exp = rand.Uint64()
	srv.id = uuid.New()
	sessionKey := neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey())
	srv.sessionPubKey = sessionKey.Public()
	_dial := func(t testing.TB, srv *startSessionServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodSessionCreate, op)
			require.Greater(t, dur, srv.sleepDur)
			assertErr(err)
		})
		if customizeOpts != nil {
			customizeOpts(&opts)
		}

		c, err := New(anyValidURI, opts)
		require.NoError(t, err)

		conn := bufconn.Listen(10 << 10)
		gs := grpc.NewServer()
		apinetmap.RegisterNetmapServiceServer(gs, srv)
		apisession.RegisterSessionServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *startSessionServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.StartSession(ctx, nil, srv.exp, StartSessionOptions{})
		require.ErrorIs(t, err, errMissingSigner)
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
		assertErr(err)
		require.Equal(t, srv.id, res.ID)
		require.Equal(t, srv.sessionPubKey, res.PublicKey)
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.StartSession(ctx, usertest.FailSigner(srv.issuer), srv.exp, StartSessionOptions{})
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
			_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apisession.CreateResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apisession.CreateResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apisession.CreateResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apisession.CreateResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apisession.CreateResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apisession.CreateResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apisession.CreateResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apisession.CreateResponse) {
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
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apisession.CreateResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
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
				srv.modifyResp = func(r *apisession.CreateResponse) {
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
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body = nil }
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing ID", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.Id = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (ID)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)

				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.Id = []byte{} }
				_, err = c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
			})
			t.Run("invalid ID", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.Id = make([]byte, 17) }
				assertErr := func(err error) {
					require.ErrorContains(t, err, "invalid response: invalid body: invalid field (ID)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.ErrorContains(t, err, "invalid UUID (got 17 bytes)")
				require.True(t, *handlerCalled)

				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.Id[6] = 1 << 4 }
				_, err = c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.ErrorContains(t, err, "wrong UUID version 1")
			})
			t.Run("missing public session key", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.SessionKey = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (public session key)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)

				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.SessionKey = []byte{} }
				_, err = c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
			})
			t.Run("invalid public session key", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.SessionKey = r.Body.SessionKey[:32] }
				assertErr := func(err error) {
					require.ErrorContains(t, err, "invalid response: invalid body: invalid field (public session key)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.ErrorContains(t, err, "unexpected EOF")
				require.True(t, *handlerCalled)

				srv.modifyResp = func(r *apisession.CreateResponse) { r.Body.SessionKey[0] = 255 }
				_, err = c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
				assertErr(err)
				require.ErrorContains(t, err, "invalid prefix 255")
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
			_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
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
			_, err := c.StartSession(ctx, srv.issuer, srv.exp, StartSessionOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
