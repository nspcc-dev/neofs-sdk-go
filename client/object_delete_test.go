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
	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type deleteObjectServer struct {
	noOtherObjectCalls
	// client
	cnr          cid.ID
	obj          oid.ID
	clientSigner neofscrypto.Signer
	session      *session.Object
	bearerToken  *bearer.Token
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	tomb           oid.ID
	errTransport   error
	modifyResp     func(*apiobject.DeleteResponse)
	corruptRespSig func(*apiobject.DeleteResponse)
}

func (x deleteObjectServer) Delete(ctx context.Context, req *apiobject.DeleteRequest) (*apiobject.DeleteResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apiobject.DeleteResponse{
		MetaHeader: &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch},
	}
	var err error
	var cnr cid.ID
	var obj oid.ID
	sigScheme := refs.SignatureScheme(x.clientSigner.Scheme())
	creatorPubKey := neofscrypto.PublicKeyBytes(x.clientSigner.Public())
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
	} else if req.MetaHeader == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing meta header"
	} else if req.MetaHeader.Ttl != 2 {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid TTL %d, expected 2", req.MetaHeader.Ttl)
	} else if req.Body == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: missing body"
	} else if req.Body.Address == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing address"
	} else if req.Body.Address.ObjectId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: invalid address: missing ID"
	} else if err = obj.ReadFromV2(req.Body.Address.ObjectId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid address: invalid ID: %s", err)
	} else if obj != x.obj {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong ID"
	} else if req.Body.Address.ContainerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: invalid address: missing container"
	} else if err = cnr.ReadFromV2(req.Body.Address.ContainerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid address: invalid container: %s", err)
	} else if cnr != x.cnr {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong container"
	}
	if sts.Code == 0 && x.session != nil {
		var so session.Object
		if req.MetaHeader.SessionToken == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing session token"
		} else if err = so.ReadFromV2(req.MetaHeader.SessionToken); err != nil {
			sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid meta header: invalid session token: %v", err)
		} else if !bytes.Equal(so.Marshal(), x.session.Marshal()) {
			sts.Code, sts.Message = status.InternalServerError, "[test] session token in request differs with the input one"
		}
	}
	if sts.Code == 0 && x.bearerToken != nil {
		var bt bearer.Token
		if req.MetaHeader.BearerToken == nil {
			sts.Code, sts.Message = status.InternalServerError, "[test] missing bearer token"
		} else if err = bt.ReadFromV2(req.MetaHeader.BearerToken); err != nil {
			sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid meta header: invalid bearer token: %v", err)
		} else if !bytes.Equal(bt.Marshal(), x.bearerToken.Marshal()) {
			sts.Code, sts.Message = status.InternalServerError, "[test] bearer token in request differs with the input one"
		}
	}
	if sts.Code == 0 {
		resp.MetaHeader.Status = nil
		resp.Body = &apiobject.DeleteResponse_Body{Tombstone: &refs.Address{ObjectId: new(refs.ObjectID)}}
		x.tomb.WriteToV2(resp.Body.Tombstone.ObjectId)
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

func TestClient_DeleteObject(t *testing.T) {
	ctx := context.Background()
	var srv deleteObjectServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	usr, _ := usertest.TwoUsers()
	srv.clientSigner = usr
	srv.cnr = cidtest.ID()
	srv.obj = oidtest.ID()
	srv.tomb = oidtest.ID()
	_dial := func(t testing.TB, srv *deleteObjectServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodObjectDelete, op)
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
		apiobject.RegisterObjectServiceServer(gs, srv)
		go func() { _ = gs.Serve(conn) }()
		t.Cleanup(gs.Stop)

		c.dial = func(ctx context.Context, _ string) (net.Conn, error) { return conn.DialContext(ctx) }
		require.NoError(t, c.Dial(ctx))

		return c, &handlerCalled
	}
	dial := func(t testing.TB, srv *deleteObjectServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.DeleteObject(ctx, srv.cnr, srv.obj, nil, DeleteObjectOptions{})
		require.ErrorIs(t, err, errMissingSigner)
	})
	t.Run("OK", func(t *testing.T) {
		srv := srv
		assertErr := func(err error) { require.NoError(t, err) }
		c, handlerCalled := dial(t, &srv, assertErr)
		res, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
		assertErr(err)
		require.Equal(t, srv.tomb, res)
		require.True(t, *handlerCalled)
		t.Run("with session", func(t *testing.T) {
			srv := srv
			so := sessiontest.Object()
			srv.session = &so
			assertErr := func(err error) { require.NoError(t, err) }
			c, handlerCalled := dial(t, &srv, assertErr)
			var opts DeleteObjectOptions
			opts.WithinSession(so)
			res, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, opts)
			assertErr(err)
			require.Equal(t, srv.tomb, res)
			require.True(t, *handlerCalled)
		})
		t.Run("with bearer token", func(t *testing.T) {
			srv := srv
			bt := bearertest.Token()
			srv.bearerToken = &bt
			assertErr := func(err error) { require.NoError(t, err) }
			c, handlerCalled := dial(t, &srv, assertErr)
			var opts DeleteObjectOptions
			opts.WithBearerToken(bt)
			res, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, opts)
			assertErr(err)
			require.Equal(t, srv.tomb, res)
			require.True(t, *handlerCalled)
		})
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, neofscryptotest.FailSigner(srv.clientSigner), DeleteObjectOptions{})
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
			_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apiobject.DeleteResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apiobject.DeleteResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apiobject.DeleteResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apiobject.DeleteResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.DeleteResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.DeleteResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.DeleteResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.DeleteResponse) {
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
				_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apiobject.DeleteResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
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
				{code: status.ObjectAccessDenied, errConst: apistatus.ErrObjectAccessDenied, errVar: new(apistatus.ObjectAccessDenied)},
				{code: status.ObjectLocked, errConst: apistatus.ErrObjectLocked, errVar: new(apistatus.ObjectLocked)},
				{code: status.SessionTokenNotFound, errConst: apistatus.ErrSessionTokenNotFound, errVar: new(apistatus.SessionTokenNotFound)},
				{code: status.SessionTokenExpired, errConst: apistatus.ErrSessionTokenExpired, errVar: new(apistatus.SessionTokenExpired)},
			} {
				srv := srv
				srv.modifyResp = func(r *apiobject.DeleteResponse) {
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
				_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apiobject.DeleteResponse) { r.Body = nil }
				_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing tombstone", func(t *testing.T) {
				srv := srv
				srv.modifyResp = func(r *apiobject.DeleteResponse) { r.Body.Tombstone = nil }
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (tombstone)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid tombstone", func(t *testing.T) {
				for _, testCase := range []struct {
					err     string
					corrupt func(*refs.Address)
				}{
					{err: "missing ID field", corrupt: func(a *refs.Address) { a.ObjectId = nil }},
					{err: "invalid ID: missing value field", corrupt: func(a *refs.Address) { a.ObjectId.Value = nil }},
					{err: "invalid ID: invalid value length 31", corrupt: func(a *refs.Address) { a.ObjectId.Value = make([]byte, 31) }},
				} {
					srv := srv
					srv.modifyResp = func(r *apiobject.DeleteResponse) { testCase.corrupt(r.Body.Tombstone) }
					assertErr := func(err error) {
						require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (tombstone): %s", testCase.err))
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
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
			_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
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
			_, err := c.DeleteObject(ctx, srv.cnr, srv.obj, srv.clientSigner, DeleteObjectOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
