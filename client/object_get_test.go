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
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type getObjectHeaderServer struct {
	noOtherObjectCalls
	// client
	cnr          cid.ID
	obj          oid.ID
	clientSigner neofscrypto.Signer
	local        bool
	raw          bool
	session      *session.Object
	bearerToken  *bearer.Token
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	retSplitInfo   bool
	header         object.Header
	splitInfo      object.SplitInfo
	errTransport   error
	modifyResp     func(*apiobject.HeadResponse)
	corruptRespSig func(*apiobject.HeadResponse)
}

func (x getObjectHeaderServer) Head(ctx context.Context, req *apiobject.HeadRequest) (*apiobject.HeadResponse, error) {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.errTransport != nil {
		return nil, x.errTransport
	}
	var sts status.Status
	resp := apiobject.HeadResponse{
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
	} else if req.Body.Raw != x.raw {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong raw flag"
	} else if req.Body.MainOnly {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: main_only flag is set"
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
		resp.Body = new(apiobject.HeadResponse_Body)
		if x.retSplitInfo {
			var si apiobject.SplitInfo
			x.splitInfo.WriteToV2(&si)
			resp.Body.Head = &apiobject.HeadResponse_Body_SplitInfo{SplitInfo: &si}
		} else {
			var h apiobject.Header
			x.header.WriteToV2(&h)
			resp.Body.Head = &apiobject.HeadResponse_Body_Header{Header: &apiobject.HeaderWithSignature{Header: &h}}
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

func TestClient_GetObjectHeader(t *testing.T) {
	ctx := context.Background()
	var srv getObjectHeaderServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	usr, _ := usertest.TwoUsers()
	srv.clientSigner = usr
	srv.cnr = cidtest.ID()
	srv.obj = oidtest.ID()
	srv.header = objecttest.Header()
	srv.splitInfo = objecttest.SplitInfo()
	_dial := func(t testing.TB, srv *getObjectHeaderServer, assertErr func(error), customizeOpts func(*Options)) (*Client, *bool) {
		var opts Options
		var handlerCalled bool
		opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
			handlerCalled = true
			require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
			require.Equal(t, "localhost:8080", endpoint)
			require.Equal(t, stat.MethodObjectHead, op)
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
	dial := func(t testing.TB, srv *getObjectHeaderServer, assertErr func(error)) (*Client, *bool) {
		return _dial(t, srv, assertErr, nil)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		_, err = c.GetObjectHeader(ctx, srv.cnr, srv.obj, nil, GetObjectHeaderOptions{})
		require.ErrorIs(t, err, errMissingSigner)
	})
	t.Run("OK", func(t *testing.T) {
		for _, testCase := range []struct {
			name    string
			setOpts func(srv *getObjectHeaderServer, opts *GetObjectHeaderOptions)
		}{
			{name: "default", setOpts: func(srv *getObjectHeaderServer, opts *GetObjectHeaderOptions) {}},
			{name: "with session", setOpts: func(srv *getObjectHeaderServer, opts *GetObjectHeaderOptions) {
				so := sessiontest.Object()
				opts.WithinSession(so)
				srv.session = &so
			}},
			{name: "with bearer token", setOpts: func(srv *getObjectHeaderServer, opts *GetObjectHeaderOptions) {
				bt := bearertest.Token()
				opts.WithBearerToken(bt)
				srv.bearerToken = &bt
			}},
			{name: "no forwarding", setOpts: func(srv *getObjectHeaderServer, opts *GetObjectHeaderOptions) {
				srv.local = true
				opts.PreventForwarding()
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				srv := srv
				var opts GetObjectHeaderOptions
				testCase.setOpts(&srv, &opts)
				assertErr := func(err error) { require.NoError(t, err) }
				c, handlerCalled := dial(t, &srv, assertErr)
				res, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, opts)
				assertErr(err)
				if !assert.ObjectsAreEqual(srv.header, res) {
					// can be caused by gRPC service fields, binaries must still be equal
					require.Equal(t, srv.header.Marshal(), res.Marshal())
				}
				require.True(t, *handlerCalled)
			})
		}
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("split info", func(t *testing.T) {
			srv := srv
			srv.raw = true
			srv.retSplitInfo = true
			assertErr := func(err error) {
				var si object.SplitInfoError
				require.ErrorAs(t, err, &si)
				require.EqualValues(t, srv.splitInfo, si)
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			var opts GetObjectHeaderOptions
			opts.PreventAssembly()
			_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, opts)
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("sign request", func(t *testing.T) {
			srv := srv
			srv.sleepDur = 0
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, neofscryptotest.FailSigner(srv.clientSigner), GetObjectHeaderOptions{})
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
			_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
			assertErr(err)
			require.True(t, *handlerCalled)
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apiobject.HeadResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apiobject.HeadResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apiobject.HeadResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apiobject.HeadResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.HeadResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.HeadResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.HeadResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.HeadResponse) {
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
				_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			srv := srv
			srv.modifyResp = func(r *apiobject.HeadResponse) {
				r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
			}
			assertErr := func(err error) {
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			c, handlerCalled := dial(t, &srv, assertErr)
			_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
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
				{code: status.ObjectAlreadyRemoved, errConst: apistatus.ErrObjectAlreadyRemoved, errVar: new(apistatus.ObjectAlreadyRemoved)},
				{code: status.SessionTokenExpired, errConst: apistatus.ErrSessionTokenExpired, errVar: new(apistatus.SessionTokenExpired)},
			} {
				srv := srv
				srv.modifyResp = func(r *apiobject.HeadResponse) {
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
				_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) { require.EqualError(t, err, "invalid response: missing body") }
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apiobject.HeadResponse) { r.Body = nil }
				_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid oneof field", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: invalid field: unknown/invalid oneof field (*object.HeadResponse_Body_ShortHeader)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apiobject.HeadResponse) { r.Body.Head = new(apiobject.HeadResponse_Body_ShortHeader) }
				_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("missing header", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: missing required field (header)")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apiobject.HeadResponse) { r.Body.Head = new(apiobject.HeadResponse_Body_Header) }
				_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
				srv.modifyResp = func(r *apiobject.HeadResponse) {
					r.Body.Head = &apiobject.HeadResponse_Body_Header{Header: new(apiobject.HeaderWithSignature)}
				}
				_, err = c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
			})
			t.Run("invalid header", func(t *testing.T) {
				srv := srv
				for _, testCase := range invalidObjectHeaderTestCases {
					srv.modifyResp = func(r *apiobject.HeadResponse) {
						testCase.corrupt(r.Body.Head.(*apiobject.HeadResponse_Body_Header).Header.Header)
					}
					assertErr := func(err error) {
						if err.Error() != testCase.err {
							require.ErrorContains(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (header): %s", testCase.err))
						}
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
					assertErr(err)
					require.True(t, *handlerCalled, testCase)
				}
				for _, testCase := range invalidObjectHeaderTestCases {
					srv.modifyResp = func(r *apiobject.HeadResponse) {
						testCase.corrupt(r.Body.Head.(*apiobject.HeadResponse_Body_Header).Header.Header.Split.ParentHeader)
					}
					assertErr := func(err error) {
						if err.Error() != testCase.err {
							require.ErrorContains(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (header): invalid parent header: %s", testCase.err))
						}
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
					assertErr(err)
					require.True(t, *handlerCalled, testCase)
				}
			})
			t.Run("unexpected split info", func(t *testing.T) {
				srv := srv
				assertErr := func(err error) {
					require.EqualError(t, err, "invalid response: invalid body: server responded with split info which was not requested")
				}
				c, handlerCalled := dial(t, &srv, assertErr)
				srv.modifyResp = func(r *apiobject.HeadResponse) { r.Body.Head = new(apiobject.HeadResponse_Body_SplitInfo) }
				_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
				assertErr(err)
				require.True(t, *handlerCalled)
			})
			t.Run("invalid split info", func(t *testing.T) {
				for _, testCase := range []struct {
					err     string
					corrupt func(*apiobject.SplitInfo)
				}{
					{err: "invalid split ID length 15", corrupt: func(i *apiobject.SplitInfo) { i.SplitId = make([]byte, 15) }},
					{err: "both linking and last split-chain elements are missing", corrupt: func(i *apiobject.SplitInfo) {
						i.LastPart, i.Link = nil, nil
					}},
					{err: "invalid last split-chain element: missing value field", corrupt: func(i *apiobject.SplitInfo) { i.LastPart.Value = nil }},
					{err: "invalid last split-chain element: invalid value length 31", corrupt: func(i *apiobject.SplitInfo) { i.LastPart.Value = make([]byte, 31) }},
					{err: "invalid linking split-chain element: missing value field", corrupt: func(i *apiobject.SplitInfo) { i.Link.Value = nil }},
					{err: "invalid linking split-chain element: invalid value length 31", corrupt: func(i *apiobject.SplitInfo) { i.Link.Value = make([]byte, 31) }},
					{err: "invalid first split-chain element: missing value field", corrupt: func(i *apiobject.SplitInfo) { i.FirstPart.Value = nil }},
					{err: "invalid first split-chain element: invalid value length 31", corrupt: func(i *apiobject.SplitInfo) { i.FirstPart.Value = make([]byte, 31) }},
				} {
					srv := srv
					srv.raw = true
					srv.retSplitInfo = true
					srv.modifyResp = func(r *apiobject.HeadResponse) {
						testCase.corrupt(r.Body.Head.(*apiobject.HeadResponse_Body_SplitInfo).SplitInfo)
					}
					assertErr := func(err error) {
						require.EqualError(t, err, fmt.Sprintf("invalid response: invalid body: invalid field (split info): %s", testCase.err))
					}
					c, handlerCalled := dial(t, &srv, assertErr)
					var opts GetObjectHeaderOptions
					opts.PreventAssembly()
					_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, opts)
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
			_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
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
			_, err := c.GetObjectHeader(ctx, srv.cnr, srv.obj, srv.clientSigner, GetObjectHeaderOptions{})
			assertErr(err)
			require.True(t, respHandlerCalled)
			require.True(t, *reqHandlerCalled)
		})
	})
}
