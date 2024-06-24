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

	apiacl "github.com/nspcc-dev/neofs-sdk-go/api/acl"
	apicontainer "github.com/nspcc-dev/neofs-sdk-go/api/container"
	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type noOtherContainerCalls struct{}

func (noOtherContainerCalls) Delete(context.Context, *apicontainer.DeleteRequest) (*apicontainer.DeleteResponse, error) {
	panic("must not be called")
}

func (noOtherContainerCalls) Get(context.Context, *apicontainer.GetRequest) (*apicontainer.GetResponse, error) {
	panic("must not be called")
}

func (noOtherContainerCalls) List(context.Context, *apicontainer.ListRequest) (*apicontainer.ListResponse, error) {
	panic("must not be called")
}

func (noOtherContainerCalls) SetExtendedACL(context.Context, *apicontainer.SetExtendedACLRequest) (*apicontainer.SetExtendedACLResponse, error) {
	panic("must not be called")
}

func (noOtherContainerCalls) GetExtendedACL(context.Context, *apicontainer.GetExtendedACLRequest) (*apicontainer.GetExtendedACLResponse, error) {
	panic("must not be called")
}

func (noOtherContainerCalls) AnnounceUsedSpace(context.Context, *apicontainer.AnnounceUsedSpaceRequest) (*apicontainer.AnnounceUsedSpaceResponse, error) {
	panic("must not be called")
}

func (noOtherContainerCalls) Put(context.Context, *apicontainer.PutRequest) (*apicontainer.PutResponse, error) {
	panic("must not be called")
}

type putContainerServer struct {
	noOtherContainerCalls
	// client
	cnrBin        []byte
	creatorSigner neofscrypto.Signer
	session       *session.Container
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	id             cid.ID
	errTransport   error
	modifyResp     func(*apicontainer.PutResponse)
	corruptRespSig func(*apicontainer.PutResponse)
}

func (x putContainerServer) Put(ctx context.Context, req *apicontainer.PutRequest) (*apicontainer.PutResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.PutResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var cnr container.Container
	sigScheme := refs.SignatureScheme(x.creatorSigner.Scheme())
	creatorPubKey := neofscrypto.PublicKeyBytes(x.creatorSigner.Public())
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
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.Container == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing container"
	} else if err = cnr.ReadFromV2(req.Body.Container); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid container field: %s", err)
	} else if !bytes.Equal(cnr.Marshal(), x.cnrBin) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong container"
	} else if req.Body.Signature == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing container signature"
	} else if !bytes.Equal(req.Body.Signature.Key, creatorPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] public key in request body differs with the creator's one"
	} else if !x.creatorSigner.Public().Verify(x.cnrBin, req.Body.Signature.Sign) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong container signature"
	} else if x.session != nil {
		var sc session.Container
		if req.MetaHeader == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing request meta header"
		} else if req.MetaHeader.SessionToken == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing session token"
		} else if err = sc.ReadFromV2(req.MetaHeader.SessionToken); err != nil {
			sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid meta header: invalid session token: %v", err)
		} else if !bytes.Equal(sc.Marshal(), x.session.Marshal()) {
			sts.Code, sts.Message = status.InternalServerError, "[test] session token in request differs with the input one"
		}
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	}
	if sts.Code == 0 {
		resp.MetaHeader.Status = nil
		resp.Body = &apicontainer.PutResponse_Body{ContainerId: new(refs.ContainerID)}
		x.id.WriteToV2(resp.Body.ContainerId)
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

func TestClient_PutContainer(t *testing.T) {
	ctx := context.Background()
	var srv putContainerServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	cnr := containertest.Container()
	srv.cnrBin = cnr.Marshal()
	srv.creatorSigner = neofscryptotest.RandomSignerRFC6979()
	srv.id = cidtest.ID()
	_dial := func(t testing.TB, srv *putContainerServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerPut, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *putContainerServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.PutContainer(ctx, cnr, nil, PutContainerOptions{})
		require.ErrorIs(t, err, errMissingSigner)
		_, err = c.PutContainer(ctx, cnr, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), PutContainerOptions{})
		require.EqualError(t, err, "wrong signature scheme: ECDSA_SHA512 instead of ECDSA_RFC6979_SHA256")
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
		assertErr(err)
		require.Equal(t, srv.id, res)
		require.True(t, *handlerCalled)
		t.Run("with session", func(t *testing.T) {
			srv := srv
			sc := sessiontest.Container()
			srv.session = &sc
			assertErr := func(err error) { require.NoError(t, err) }
			c, handlerCalled := dial(t, &srv, assertErr)
			var opts PutContainerOptions
			opts.WithinSession(sc)
			res, err := c.PutContainer(ctx, cnr, srv.creatorSigner, opts)
			assertErr(err)
			require.Equal(t, srv.id, res)
			require.True(t, *handlerCalled)
		})
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign container", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, "sign container") }
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.PutContainer(ctx, cnr, neofscryptotest.FailSigner(srv.creatorSigner), PutContainerOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.PutContainer(ctx, cnr, newDisposableSigner(srv.creatorSigner), PutContainerOptions{})
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
			_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.PutResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.PutResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.PutResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.PutResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.PutResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.PutResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.PutResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.PutResponse) {
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
				_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.PutResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
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
				srv.modifyResp = func(r *apicontainer.PutResponse) {
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
				_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apicontainer.PutResponse) { r.Body = nil }
				_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing container ID", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apicontainer.PutResponse) { r.Body.ContainerId = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (ID)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid container ID", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apicontainer.PutResponse) { r.Body.ContainerId.Value = make([]byte, 31) }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: invalid field (ID): invalid value length 31")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
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
			_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			srv := srv
			srv.epoch = 4386380643
			assertErr := func(err error) { require.EqualError(t, err, "intercept response info: some handler error") }
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
			_, err := c.PutContainer(ctx, cnr, srv.creatorSigner, PutContainerOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type getContainerServer struct {
	noOtherContainerCalls
	// client
	id              cid.ID
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	container      container.Container
	errTransport   error
	modifyResp     func(*apicontainer.GetResponse)
	corruptRespSig func(*apicontainer.GetResponse)
}

func (x getContainerServer) Get(ctx context.Context, req *apicontainer.GetRequest) (*apicontainer.GetResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.GetResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var id cid.ID
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
		sts.Code, sts.Message = status.InternalServerError, "meta header is set"
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "missing request body"
	} else if req.Body.ContainerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request body: missing ID"
	} else if err = id.ReadFromV2(req.Body.ContainerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request body: invalid ID field: %s", err)
	} else if id != x.id {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong ID"
	} else {
		resp.MetaHeader.Status = nil
		resp.Body = &apicontainer.GetResponse_Body{
			Container: new(apicontainer.Container),
			Signature: &refs.SignatureRFC6979{
				Key:  []byte("any_public_key"),
				Sign: []byte("any_signature"),
			},
			SessionToken: &apisession.SessionToken{
				Body: &apisession.SessionToken_Body{
					Id:      []byte("any_ID"),
					OwnerId: &refs.OwnerID{Value: []byte("any_owner")},
					Lifetime: &apisession.SessionToken_Body_TokenLifetime{
						Exp: rand.Uint64(),
						Nbf: rand.Uint64(),
						Iat: rand.Uint64(),
					},
					SessionKey: []byte("any_session_key"),
					Context: &apisession.SessionToken_Body_Object{
						Object: &apisession.ObjectSessionContext{
							Verb: apisession.ObjectSessionContext_Verb(rand.Int31()),
							Target: &apisession.ObjectSessionContext_Target{
								Container: &refs.ContainerID{Value: []byte("any_container")},
								Objects:   []*refs.ObjectID{{Value: []byte("any_object1")}, {Value: []byte("any_object2")}},
							},
						},
					},
				},
				Signature: &refs.Signature{
					Key:    []byte("any_public_key"),
					Sign:   []byte("any_signature"),
					Scheme: refs.SignatureScheme(rand.Int31()),
				},
			},
		}
		x.container.WriteToV2(resp.Body.Container)
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

func TestClient_GetContainer(t *testing.T) {
	ctx := context.Background()
	var srv getContainerServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.id = cidtest.ID()
	srv.container = containertest.Container()
	_dial := func(t testing.TB, srv *getContainerServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerGet, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *getContainerServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
		assertErr(err)
		if !assert.ObjectsAreEqual(srv.container, res) {
			// can be caused by gRPC service fields, binaries must still be equal
			require.Equal(t, srv.container.Marshal(), res.Marshal())
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
			_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
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
			_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.GetResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.GetResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.GetResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.GetResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.GetResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.GetResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.GetResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.GetResponse) {
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
				_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.GetResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
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
				{code: status.ContainerNotFound, errConst: apistatus.ErrContainerNotFound, errVar: new(apistatus.ContainerNotFound)},
			} {
				srv := srv
				srv.modifyResp = func(r *apicontainer.GetResponse) {
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
				_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apicontainer.GetResponse) { r.Body = nil }
				_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing container", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apicontainer.GetResponse) { r.Body.Container = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (container)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid container", func(t *testing.T) {
				testCases := []struct {
					name     string
					err      string
					contains bool
					corrupt  func(*apicontainer.Container)
				}{
					{name: "missing version", err: "missing version", corrupt: func(c *apicontainer.Container) {
						c.Version = nil
					}},
					{name: "missing owner", err: "missing owner", corrupt: func(c *apicontainer.Container) {
						c.OwnerId = nil
					}},
					{name: "nil nonce", err: "missing nonce", corrupt: func(c *apicontainer.Container) {
						c.Nonce = nil
					}},
					{name: "empty nonce", err: "missing nonce", corrupt: func(c *apicontainer.Container) {
						c.Nonce = []byte{}
					}},
					{name: "missing policy", err: "missing placement policy", corrupt: func(c *apicontainer.Container) {
						c.PlacementPolicy = nil
					}},
					{name: "owner/nil value", err: "invalid owner: missing value field", corrupt: func(c *apicontainer.Container) {
						c.OwnerId.Value = nil
					}},
					{name: "owner/empty value", err: "invalid owner: missing value field", corrupt: func(c *apicontainer.Container) {
						c.OwnerId.Value = []byte{}
					}},
					{name: "owner/wrong length", err: "invalid owner: invalid value length 24", corrupt: func(c *apicontainer.Container) {
						c.OwnerId.Value = make([]byte, 24)
					}},
					{name: "owner/wrong prefix", err: "invalid owner: invalid prefix byte 0x34, expected 0x35", corrupt: func(c *apicontainer.Container) {
						c.OwnerId.Value[0] = 0x34
					}},
					{name: "owner/checksum mismatch", err: "invalid owner: value checksum mismatch", corrupt: func(c *apicontainer.Container) {
						c.OwnerId.Value[24]++
					}},
					{name: "nonce/wrong length", err: "invalid nonce: invalid UUID (got 15 bytes)", corrupt: func(c *apicontainer.Container) {
						c.Nonce = make([]byte, 15)
					}},
					{name: "nonce/wrong version", err: "invalid nonce: wrong UUID version 3", corrupt: func(c *apicontainer.Container) {
						c.Nonce[6] = 3 << 4
					}},
					{name: "nonce/nil replicas", err: "invalid placement policy: missing replicas", corrupt: func(c *apicontainer.Container) {
						c.PlacementPolicy.Replicas = nil
					}},
					{name: "attributes/empty key", err: "invalid attribute #1: missing key", corrupt: func(c *apicontainer.Container) {
						c.Attributes = []*apicontainer.Container_Attribute{
							{Key: "key_valid", Value: "any"},
							{Key: "", Value: "any"},
						}
					}},
					{name: "attributes/repeated keys", err: "multiple attributes with key=k2", corrupt: func(c *apicontainer.Container) {
						c.Attributes = []*apicontainer.Container_Attribute{
							{Key: "k1", Value: "any"},
							{Key: "k2", Value: "1"},
							{Key: "k3", Value: "any"},
							{Key: "k2", Value: "2"},
						}
					}},
					{name: "attributes/empty value", err: "invalid attribute #1 (key2): missing value", corrupt: func(c *apicontainer.Container) {
						c.Attributes = []*apicontainer.Container_Attribute{
							{Key: "key1", Value: "any"},
							{Key: "key2", Value: ""},
						}
					}},
					{name: "attributes/invalid timestamp", err: "invalid timestamp attribute (#1): invalid integer", contains: true, corrupt: func(c *apicontainer.Container) {
						c.Attributes = []*apicontainer.Container_Attribute{
							{Key: "key1", Value: "any"},
							{Key: "Timestamp", Value: "not_a_number"},
						}
					}},
				}
				for i := range testCases {
					srv := srv
					srv.modifyResp = func(r *apicontainer.GetResponse) {
						testCases[i].corrupt(r.Body.Container)
					}
					assertErr := func(err error) {
						if testCases[i].contains {
							require.ErrorContains(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (container): %s", testCases[i].err))
						} else {
							require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (container): %s", testCases[i].err))
						}
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
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
			_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
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
			_, err := c.GetContainer(ctx, srv.id, GetContainerOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type listContainersServer struct {
	noOtherContainerCalls
	// client
	usr             user.ID
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	containers     []cid.ID
	errTransport   error
	modifyResp     func(*apicontainer.ListResponse)
	corruptRespSig func(*apicontainer.ListResponse)
}

func (x listContainersServer) List(ctx context.Context, req *apicontainer.ListRequest) (*apicontainer.ListResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.ListResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var usr user.ID
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
		sts.Code, sts.Message = status.InternalServerError, "meta header is set"
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "missing request body"
	} else if req.Body.OwnerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request body: missing user"
	} else if err = usr.ReadFromV2(req.Body.OwnerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request body: invalid user field: %s", err)
	} else if usr != x.usr {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong user"
	} else {
		resp.MetaHeader.Status = nil
		if len(x.containers) > 0 {
			resp.Body = &apicontainer.ListResponse_Body{ContainerIds: make([]*refs.ContainerID, len(x.containers))}
			for i := range x.containers {
				resp.Body.ContainerIds[i] = new(refs.ContainerID)
				x.containers[i].WriteToV2(resp.Body.ContainerIds[i])
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

func TestClient_ListContainers(t *testing.T) {
	ctx := context.Background()
	var srv listContainersServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.usr = usertest.ID()
	srv.containers = cidtest.NIDs(5)
	_dial := func(t testing.TB, srv *listContainersServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerList, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *listContainersServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
		assertErr(err)
		require.Equal(t, srv.containers, res)
		require.True(t, *handlerCalled)

		srv.containers = nil
		res, err = c.ListContainers(ctx, srv.usr, ListContainersOptions{})
		assertErr(err)
		require.Empty(t, res)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
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
			_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.ListResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.ListResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.ListResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.ListResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.ListResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.ListResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.ListResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.ListResponse) {
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
				_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.ListResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
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
				srv.modifyResp = func(r *apicontainer.ListResponse) {
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
				_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("invalid IDs", func(t *testing.T) {
				testCases := []struct {
					name    string
					err     string
					corrupt func([]*refs.ContainerID)
				}{
					// nil-ness is "lost" on gRPC transmission: element is decoded as zero structure,
					// therefore we won't reach 'nil element' error but the next one
					{name: "nil", err: "invalid element #1: missing value field", corrupt: func(ids []*refs.ContainerID) {
						ids[1] = nil
					}},
					{name: "wrong length", err: "invalid element #2: invalid value length 31", corrupt: func(ids []*refs.ContainerID) {
						ids[2].Value = make([]byte, 31)
					}},
				}
				for i := range testCases {
					srv := srv
					srv.modifyResp = func(r *apicontainer.ListResponse) {
						testCases[i].corrupt(r.Body.ContainerIds)
					}
					assertErr := func(err error) {
						require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (ID list): %s", testCases[i].err))
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
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
			_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
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
			_, err := c.ListContainers(ctx, srv.usr, ListContainersOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type deleteContainerServer struct {
	noOtherContainerCalls
	// client
	cnr           cid.ID
	removerSigner neofscrypto.Signer
	session       *session.Container
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	errTransport   error
	modifyResp     func(*apicontainer.DeleteResponse)
	corruptRespSig func(*apicontainer.DeleteResponse)
}

func (x deleteContainerServer) Delete(ctx context.Context, req *apicontainer.DeleteRequest) (*apicontainer.DeleteResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.DeleteResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var cnr cid.ID
	sigScheme := refs.SignatureScheme(x.removerSigner.Scheme())
	creatorPubKey := neofscrypto.PublicKeyBytes(x.removerSigner.Public())
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
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.ContainerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing ID"
	} else if err = cnr.ReadFromV2(req.Body.ContainerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid ID field: %s", err)
	} else if cnr != x.cnr {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong ID"
	} else if req.Body.Signature == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing ID signature"
	} else if !bytes.Equal(req.Body.Signature.Key, creatorPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] public key in request body differs with the creator's one"
	} else if !x.removerSigner.Public().Verify(cnr[:], req.Body.Signature.Sign) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong ID signature"
	} else if x.session != nil {
		var sc session.Container
		if req.MetaHeader == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing request meta header"
		} else if req.MetaHeader.SessionToken == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing session token"
		} else if err = sc.ReadFromV2(req.MetaHeader.SessionToken); err != nil {
			sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid meta header: invalid session token: %v", err)
		} else if !bytes.Equal(sc.Marshal(), x.session.Marshal()) {
			sts.Code, sts.Message = status.InternalServerError, "[test] session token in request differs with the input one"
		}
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	}
	if sts.Code == 0 {
		resp.MetaHeader.Status = nil
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

func TestClient_DeleteContainer(t *testing.T) {
	ctx := context.Background()
	var srv deleteContainerServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.cnr = cidtest.ID()
	srv.removerSigner = neofscryptotest.RandomSignerRFC6979()
	_dial := func(t testing.TB, srv *deleteContainerServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerDelete, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *deleteContainerServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		err = c.DeleteContainer(ctx, srv.cnr, nil, DeleteContainerOptions{})
		require.ErrorIs(t, err, errMissingSigner)
		err = c.DeleteContainer(ctx, srv.cnr, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), DeleteContainerOptions{})
		require.EqualError(t, err, "wrong signature scheme: ECDSA_SHA512 instead of ECDSA_RFC6979_SHA256")
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
		assertErr(err)
		require.True(t, *handlerCalled)
		t.Run("with session", func(t *testing.T) {
			srv := srv
			sc := sessiontest.Container()
			srv.session = &sc
			assertErr := func(err error) { require.NoError(t, err) }
			c, handlerCalled := dial(t, &srv, assertErr)
			var opts DeleteContainerOptions
			opts.WithinSession(sc)
			err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, opts)
			assertErr(err)
			require.True(t, *handlerCalled)
		})
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign container", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, "sign container") }
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.DeleteContainer(ctx, srv.cnr, neofscryptotest.FailSigner(srv.removerSigner), DeleteContainerOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.DeleteContainer(ctx, srv.cnr, newDisposableSigner(srv.removerSigner), DeleteContainerOptions{})
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
			err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.DeleteResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.DeleteResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.DeleteResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.DeleteResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.DeleteResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.DeleteResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.DeleteResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.DeleteResponse) {
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
				err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.DeleteResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
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
				srv.modifyResp = func(r *apicontainer.DeleteResponse) {
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
				err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
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
			err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
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
			err := c.DeleteContainer(ctx, srv.cnr, srv.removerSigner, DeleteContainerOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type getEACLServer struct {
	noOtherContainerCalls
	// client
	cnr             cid.ID
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	eacl           eacl.Table
	errTransport   error
	modifyResp     func(*apicontainer.GetExtendedACLResponse)
	corruptRespSig func(*apicontainer.GetExtendedACLResponse)
}

func (x getEACLServer) GetExtendedACL(ctx context.Context, req *apicontainer.GetExtendedACLRequest) (*apicontainer.GetExtendedACLResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.GetExtendedACLResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var cnr cid.ID
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
		sts.Code, sts.Message = status.InternalServerError, "meta header is set"
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "missing request body"
	} else if req.Body.ContainerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request body: missing container"
	} else if err = cnr.ReadFromV2(req.Body.ContainerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request body: invalid container field: %s", err)
	} else if cnr != x.cnr {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong container"
	} else {
		resp.MetaHeader.Status = nil
		resp.Body = &apicontainer.GetExtendedACLResponse_Body{
			Eacl: new(apiacl.EACLTable),
			Signature: &refs.SignatureRFC6979{
				Key:  []byte("any_public_key"),
				Sign: []byte("any_signature"),
			},
			SessionToken: &apisession.SessionToken{
				Body: &apisession.SessionToken_Body{
					Id:      []byte("any_ID"),
					OwnerId: &refs.OwnerID{Value: []byte("any_owner")},
					Lifetime: &apisession.SessionToken_Body_TokenLifetime{
						Exp: rand.Uint64(),
						Nbf: rand.Uint64(),
						Iat: rand.Uint64(),
					},
					SessionKey: []byte("any_session_key"),
					Context: &apisession.SessionToken_Body_Object{
						Object: &apisession.ObjectSessionContext{
							Verb: apisession.ObjectSessionContext_Verb(rand.Int31()),
							Target: &apisession.ObjectSessionContext_Target{
								Container: &refs.ContainerID{Value: []byte("any_container")},
								Objects:   []*refs.ObjectID{{Value: []byte("any_object1")}, {Value: []byte("any_object2")}},
							},
						},
					},
				},
				Signature: &refs.Signature{
					Key:    []byte("any_public_key"),
					Sign:   []byte("any_signature"),
					Scheme: refs.SignatureScheme(rand.Int31()),
				},
			},
		}
		x.eacl.WriteToV2(resp.Body.Eacl)
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

func TestClient_GetEACL(t *testing.T) {
	ctx := context.Background()
	var srv getEACLServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.cnr = cidtest.ID()
	srv.eacl = eacltest.Table()
	_dial := func(t testing.TB, srv *getEACLServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerEACL, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *getEACLServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
		assertErr(err)
		require.Equal(t, srv.eacl, res)
		require.True(t, *handlerCalled)
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			c.signer = neofscryptotest.FailSigner(c.signer)
			_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
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
			_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.GetExtendedACLResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.GetExtendedACLResponse) {
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
				_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.GetExtendedACLResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
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
				{code: status.ContainerNotFound, errConst: apistatus.ErrContainerNotFound, errVar: new(apistatus.ContainerNotFound)},
				{code: status.EACLNotFound, errConst: apistatus.ErrEACLNotFound, errVar: new(apistatus.EACLNotFound)},
			} {
				srv := srv
				srv.modifyResp = func(r *apicontainer.GetExtendedACLResponse) {
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
				_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apicontainer.GetExtendedACLResponse) { r.Body = nil }
				_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing eACL", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apicontainer.GetExtendedACLResponse) { r.Body.Eacl = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (eACL)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid eACL", func(t *testing.T) {
				testCases := []struct {
					name    string
					err     string
					corrupt func(*apiacl.EACLTable)
				}{
					{name: "container/empty", err: "invalid container: missing value field", corrupt: func(c *apiacl.EACLTable) {
						c.ContainerId = new(refs.ContainerID)
					}},
					{name: "container/wrong length", err: "invalid container: invalid value length 31", corrupt: func(c *apiacl.EACLTable) {
						c.ContainerId.Value = make([]byte, 31)
					}},
					{name: "records/nil", err: "missing records", corrupt: func(c *apiacl.EACLTable) {
						c.Records = nil
					}},
					{name: "records/empty", err: "missing records", corrupt: func(c *apiacl.EACLTable) {
						c.Records = []*apiacl.EACLRecord{}
					}},
					{name: "records/targets/nil", err: "invalid record #1: missing target subjects", corrupt: func(c *apiacl.EACLTable) {
						c.Records[1].Targets = nil
					}},
					{name: "records/targets/empty", err: "invalid record #1: missing target subjects", corrupt: func(c *apiacl.EACLTable) {
						c.Records[1].Targets = []*apiacl.EACLRecord_Target{}
					}},
					{name: "records/targets/neither keys nor role", err: "invalid record #1: invalid target #2: role and public keys are not mutually exclusive", corrupt: func(c *apiacl.EACLTable) {
						c.Records[1].Targets[2].Role, c.Records[1].Targets[2].Keys = 0, nil
					}},
					{name: "records/targets/key and role", err: "invalid record #1: invalid target #2: role and public keys are not mutually exclusive", corrupt: func(c *apiacl.EACLTable) {
						c.Records[1].Targets[2].Role, c.Records[1].Targets[2].Keys = 1, make([][]byte, 1)
					}},
					{name: "filters/missing key", err: "invalid record #1: invalid filter #2: missing key", corrupt: func(c *apiacl.EACLTable) {
						c.Records[1].Filters[2].Key = ""
					}},
				}
				for i := range testCases {
					srv := srv
					rs := eacltest.NRecords(3)
					rs[1].SetTargets(eacltest.NTargets(3))
					rs[1].SetFilters(eacltest.NFilters(3))
					srv.eacl.SetRecords(rs)
					srv.modifyResp = func(r *apicontainer.GetExtendedACLResponse) {
						testCases[i].corrupt(r.Body.Eacl)
					}
					assertErr := func(err error) {
						require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (eACL): %s", testCases[i].err))
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
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
			_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
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
			_, err := c.GetEACL(ctx, srv.cnr, GetEACLOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type setEACLServer struct {
	noOtherContainerCalls
	// client
	eacl         eacl.Table
	setterSigner neofscrypto.Signer
	session      *session.Container
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	errTransport   error
	modifyResp     func(*apicontainer.SetExtendedACLResponse)
	corruptRespSig func(*apicontainer.SetExtendedACLResponse)
}

func (x setEACLServer) SetExtendedACL(ctx context.Context, req *apicontainer.SetExtendedACLRequest) (*apicontainer.SetExtendedACLResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.SetExtendedACLResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var eACL eacl.Table
	sigScheme := refs.SignatureScheme(x.setterSigner.Scheme())
	creatorPubKey := neofscrypto.PublicKeyBytes(x.setterSigner.Public())
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
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.Eacl == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing eACL"
	} else if err = eACL.ReadFromV2(req.Body.Eacl); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid eACL: %s", err)
	} else if !bytes.Equal(eACL.Marshal(), x.eacl.Marshal()) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong eACL"
	} else if req.Body.Signature == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing ID signature"
	} else if !bytes.Equal(req.Body.Signature.Key, creatorPubKey) {
		sts.Code, sts.Message = status.InternalServerError, "[test] public key in request body differs with the creator's one"
	} else if !x.setterSigner.Public().Verify(eACL.Marshal(), req.Body.Signature.Sign) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong eACL signature"
	} else if x.session != nil {
		var sc session.Container
		if req.MetaHeader == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing request meta header"
		} else if req.MetaHeader.SessionToken == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing session token"
		} else if err = sc.ReadFromV2(req.MetaHeader.SessionToken); err != nil {
			sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid meta header: invalid session token: %v", err)
		} else if !bytes.Equal(sc.Marshal(), x.session.Marshal()) {
			sts.Code, sts.Message = status.InternalServerError, "[test] session token in request differs with the input one"
		}
	} else if req.MetaHeader != nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: meta header is set"
	}
	if sts.Code == 0 {
		resp.MetaHeader.Status = nil
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

func TestClient_SetEACL(t *testing.T) {
	ctx := context.Background()
	var srv setEACLServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.eacl = eacltest.Table()
	srv.setterSigner = neofscryptotest.RandomSignerRFC6979()
	_dial := func(t testing.TB, srv *setEACLServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerSetEACL, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *setEACLServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		err = c.SetEACL(ctx, srv.eacl, nil, SetEACLOptions{})
		require.ErrorIs(t, err, errMissingSigner)
		err = c.SetEACL(ctx, srv.eacl, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), SetEACLOptions{})
		require.EqualError(t, err, "wrong signature scheme: ECDSA_SHA512 instead of ECDSA_RFC6979_SHA256")
	})
	t.Run("unbound container", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		err = c.SetEACL(ctx, eacl.Table{}, srv.setterSigner, SetEACLOptions{})
		require.EqualError(t, err, "missing container in the eACL")
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
		assertErr(err)
		require.True(t, *handlerCalled)
		t.Run("with session", func(t *testing.T) {
			srv := srv
			sc := sessiontest.Container()
			srv.session = &sc
			assertErr := func(err error) { require.NoError(t, err) }
			c, handlerCalled := dial(t, &srv, assertErr)
			var opts SetEACLOptions
			opts.WithinSession(sc)
			err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, opts)
			assertErr(err)
			require.True(t, *handlerCalled)
		})
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign eACL", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, "sign eACL") }
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.SetEACL(ctx, srv.eacl, neofscryptotest.FailSigner(srv.setterSigner), SetEACLOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.SetEACL(ctx, srv.eacl, newDisposableSigner(srv.setterSigner), SetEACLOptions{})
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
			err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.SetExtendedACLResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.SetExtendedACLResponse) {
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
				err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.SetExtendedACLResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
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
				{code: status.ContainerNotFound, errConst: apistatus.ErrContainerNotFound, errVar: new(apistatus.ContainerNotFound)},
			} {
				srv := srv
				srv.modifyResp = func(r *apicontainer.SetExtendedACLResponse) {
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
				err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
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
			err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
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
			err := c.SetEACL(ctx, srv.eacl, srv.setterSigner, SetEACLOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}

type sendContainerSizeEstimationsServer struct {
	noOtherContainerCalls
	// client
	estimations     []container.SizeEstimation
	clientSigScheme neofscrypto.Scheme
	clientPubKey    []byte
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	errTransport   error
	modifyResp     func(*apicontainer.AnnounceUsedSpaceResponse)
	corruptRespSig func(*apicontainer.AnnounceUsedSpaceResponse)
}

func (x sendContainerSizeEstimationsServer) AnnounceUsedSpace(ctx context.Context, req *apicontainer.AnnounceUsedSpaceRequest) (*apicontainer.AnnounceUsedSpaceResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apicontainer.AnnounceUsedSpaceResponse{
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
		sts.Code, sts.Message = status.InternalServerError, "meta header is set"
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "missing request body"
	} else if len(req.Body.Announcements) == 0 {
		sts.Code, sts.Message = status.InternalServerError, "invalid request body: missing estimations"
	} else if len(req.Body.Announcements) != len(x.estimations) {
		sts.Code, sts.Message = status.InternalServerError, "[test] invalid request body: wrong number of estimations"
	} else {
		var est container.SizeEstimation
		for i := range req.Body.Announcements {
			if req.Body.Announcements[i] == nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("nil estimation #%d", i)
				break
			} else if err = est.ReadFromV2(req.Body.Announcements[i]); err != nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid estimation #%d: %v", i, err)
				break
			} else if est != x.estimations[i] {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] wrong estimation #%d", i)
				break
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

func TestClient_SendContainerSizeEstimations(t *testing.T) {
	ctx := context.Background()
	var srv sendContainerSizeEstimationsServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	srv.estimations = make([]container.SizeEstimation, 3)
	for i := range srv.estimations {
		srv.estimations[i] = containertest.SizeEstimation()
	}
	_dial := func(t testing.TB, srv *sendContainerSizeEstimationsServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodContainerAnnounceUsedSpace, op)
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
		apicontainer.RegisterContainerServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *sendContainerSizeEstimationsServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("missing estimations", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		err = c.SendContainerSizeEstimations(ctx, nil, SendContainerSizeEstimationsOptions{})
		require.EqualError(t, err, "missing estimations")
		err = c.SendContainerSizeEstimations(ctx, []container.SizeEstimation{}, SendContainerSizeEstimationsOptions{})
		require.EqualError(t, err, "missing estimations")
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
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
			err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
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
			err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apicontainer.AnnounceUsedSpaceResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apicontainer.AnnounceUsedSpaceResponse) {
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
				err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apicontainer.AnnounceUsedSpaceResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
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
				srv.modifyResp = func(r *apicontainer.AnnounceUsedSpaceResponse) {
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
				err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
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
			err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
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
			err := c.SendContainerSizeEstimations(ctx, srv.estimations, SendContainerSizeEstimationsOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
