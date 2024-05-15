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
	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
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

type hashObjectPayloadRangesServer struct {
	noOtherObjectCalls
	// client
	cnr          cid.ID
	obj          oid.ID
	clientSigner neofscrypto.Signer
	typ          checksum.Type
	ranges       []object.Range
	salt         []byte
	local        bool
	session      *session.Object
	bearerToken  *bearer.Token
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	hashes         [][]byte
	errTransport   error
	modifyResp     func(*apiobject.GetRangeHashResponse)
	corruptRespSig func(*apiobject.GetRangeHashResponse)
}

func (x hashObjectPayloadRangesServer) GetRangeHash(ctx context.Context, req *apiobject.GetRangeHashRequest) (*apiobject.GetRangeHashResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apiobject.GetRangeHashResponse{
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
	} else if x.local && req.MetaHeader.Ttl != 1 || !x.local && req.MetaHeader.Ttl != 2 {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid meta header: invalid TTL %d", req.MetaHeader.Ttl)
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
	} else if !bytes.Equal(req.Body.Salt, x.salt) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong salt"
	} else if req.Body.Type != refs.ChecksumType(x.typ) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong checksum type"
	} else if len(req.Body.Ranges) == 0 {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing ranges"
	} else if len(req.Body.Ranges) != len(x.ranges) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong number of ranges"
	} else {
		for i := range req.Body.Ranges {
			if req.Body.Ranges[i] == nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: nil range #%d", i)
			} else if req.Body.Ranges[i].Length != x.ranges[i].Length || req.Body.Ranges[i].Offset != x.ranges[i].Offset {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] wrong range #%d", i)
			}
		}
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
		resp.Body = &apiobject.GetRangeHashResponse_Body{HashList: x.hashes}
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

func TestClient_HashObjectPayloadRanges(t *testing.T) {
	ctx := context.Background()
	var srv hashObjectPayloadRangesServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	usr, _ := usertest.TwoUsers()
	srv.clientSigner = usr
	srv.cnr = cidtest.ID()
	srv.obj = oidtest.ID()
	srv.typ = checksum.Type(rand.Uint32() % 256)
	if srv.typ == 0 {
		srv.typ++
	}
	srv.ranges = []object.Range{{1, 2}, {3, 4}}
	srv.hashes = [][]byte{[]byte("hello"), []byte("world")}
	_dial := func(t testing.TB, srv *hashObjectPayloadRangesServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodObjectHash, op)
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
	dial := func(t testing.TB, srv *hashObjectPayloadRangesServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, nil, HashObjectPayloadRangesOptions{}, srv.ranges)
		require.ErrorIs(t, err, errMissingSigner)
	})
	t.Run("invalid checksum type", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, 0, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
		require.EqualError(t, err, "zero checksum type")
	})
	t.Run("invalid ranges", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, nil)
		require.EqualError(t, err, "missing ranges")
		_, err = c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, []object.Range{})
		require.EqualError(t, err, "missing ranges")
		_, err = c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, []object.Range{
			{1, 2}, {3, 0},
		})
		require.EqualError(t, err, "zero length of range #1")
	})
	t.Run("OK", func(t *testing.T) {
		for _, testCase := range []struct {
			name    string
			setOpts func(srv *hashObjectPayloadRangesServer, opts *HashObjectPayloadRangesOptions)
		}{
			{name: "default", setOpts: func(srv *hashObjectPayloadRangesServer, opts *HashObjectPayloadRangesOptions) {}},
			{name: "with session", setOpts: func(srv *hashObjectPayloadRangesServer, opts *HashObjectPayloadRangesOptions) {
				so := sessiontest.Object()
				opts.WithinSession(so)
				srv.session = &so
			}},
			{name: "with bearer token", setOpts: func(srv *hashObjectPayloadRangesServer, opts *HashObjectPayloadRangesOptions) {
				bt := bearertest.Token()
				opts.WithBearerToken(bt)
				srv.bearerToken = &bt
			}},
			{name: "no forwarding", setOpts: func(srv *hashObjectPayloadRangesServer, opts *HashObjectPayloadRangesOptions) {
				srv.local = true
				opts.PreventForwarding()
			}},
			{name: "with salt", setOpts: func(srv *hashObjectPayloadRangesServer, opts *HashObjectPayloadRangesOptions) {
				srv.salt = []byte("any_salt")
				opts.WithSalt(srv.salt)
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				srv := srv
				var opts HashObjectPayloadRangesOptions
				testCase.setOpts(&srv, &opts)
				assertErr := func(err error) { require.NoError(t, err) }
				c, handlerCalled := dial(t, &srv, assertErr)
				res, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, opts, srv.ranges)
				assertErr(err)
				require.Equal(t, srv.hashes, res)
				require.True(t, *handlerCalled)
			})
		}
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, neofscryptotest.FailSigner(srv.clientSigner),
				HashObjectPayloadRangesOptions{}, srv.ranges)
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
			_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apiobject.GetRangeHashResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apiobject.GetRangeHashResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apiobject.GetRangeHashResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apiobject.GetRangeHashResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.GetRangeHashResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.GetRangeHashResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.GetRangeHashResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.GetRangeHashResponse) {
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
				_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apiobject.GetRangeHashResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
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
				{code: status.ObjectNotFound, errConst: apistatus.ErrObjectNotFound, errVar: new(apistatus.ObjectNotFound)},
				{code: status.ObjectAccessDenied, errConst: apistatus.ErrObjectAccessDenied, errVar: new(apistatus.ObjectAccessDenied)},
				{code: status.OutOfRange, errConst: apistatus.ErrObjectOutOfRange, errVar: new(apistatus.ObjectOutOfRange)},
				{code: status.SessionTokenExpired, errConst: apistatus.ErrSessionTokenExpired, errVar: new(apistatus.SessionTokenExpired)},
			} {
				srv := srv
				srv.modifyResp = func(r *apiobject.GetRangeHashResponse) {
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
				_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apiobject.GetRangeHashResponse) { r.Body = nil }
				_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
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
			_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
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
			_, err := c.HashObjectPayloadRanges(ctx, srv.cnr, srv.obj, srv.typ, srv.clientSigner, HashObjectPayloadRangesOptions{}, srv.ranges)
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
