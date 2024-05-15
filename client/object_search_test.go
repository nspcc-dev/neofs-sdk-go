package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
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
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestAllObjectsQuery(t *testing.T) {
	require.Empty(t, AllObjectsQuery())
}

type searchObjectsServer struct {
	noOtherObjectCalls
	// client
	clientSigner neofscrypto.Signer
	cnr          cid.ID
	filters      []object.SearchFilter
	local        bool
	session      *session.Object
	bearerToken  *bearer.Token
	// server
	sleepDur time.Duration
	endpointInfoOnDialServer
	emptyStream bool
	idLists     [][]oid.ID
	// allows to modify n-th response in the stream and replace it with returned
	// transport error
	modifyResp func(n int, r *apiobject.SearchResponse) error
	// allows to corrupt signature of the n-th response in the stream
	corruptRespSig func(n int, r *apiobject.SearchResponse)
}

func (x searchObjectsServer) sendResponse(stream apiobject.ObjectService_SearchServer, n int, resp *apiobject.SearchResponse) error {
	var err error
	if x.modifyResp != nil {
		if err = x.modifyResp(n, resp); err != nil {
			return err
		}
	}
	resp.VerifyHeader, err = neofscrypto.SignResponse(x.serverSigner, resp, resp.Body, nil)
	if err != nil {
		return fmt.Errorf("sign response: %w", err)
	}
	if x.corruptRespSig != nil {
		x.corruptRespSig(n, resp)
	}
	return stream.Send(resp)
}

func (x searchObjectsServer) Search(req *apiobject.SearchRequest, stream apiobject.ObjectService_SearchServer) error {
	if x.sleepDur > 0 {
		time.Sleep(x.sleepDur)
	}
	if x.emptyStream {
		return nil
	}
	var sts status.Status
	var err error
	var cnr cid.ID
	sigScheme := refs.SignatureScheme(x.clientSigner.Scheme())
	creatorPubKey := neofscrypto.PublicKeyBytes(x.clientSigner.Public())
	if req == nil {
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
	} else if req.Body.Version > 0 {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: query version set"
	} else if req.Body.ContainerId == nil {
		sts.Code, sts.Message = status.InternalServerError, "invalid request: invalid body: missing address"
	} else if err = cnr.ReadFromV2(req.Body.ContainerId); err != nil {
		sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid container: %s", err)
	} else if cnr != x.cnr {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong container"
	} else if len(req.Body.Filters) != len(x.filters) {
		sts.Code, sts.Message = status.InternalServerError, "[test] wrong number of filters"
	} else {
		var sf object.SearchFilter
		for i := range req.Body.Filters {
			if req.Body.Filters[i] == nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: nil filter #%d", i)
			} else if err = sf.ReadFromV2(req.Body.Filters[i]); err != nil {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("invalid request: invalid body: invalid filter #%d: %v", i, err)
			} else if object.FilterOp(req.Body.Filters[i].MatchType) != x.filters[i].Operation() {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] wrong filter #%d op", i)
			} else if req.Body.Filters[i].Key != x.filters[i].Key() {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] wrong filter #%d key", i)
			} else if req.Body.Filters[i].Value != x.filters[i].Value() {
				sts.Code, sts.Message = status.InternalServerError, fmt.Sprintf("[test] wrong filter #%d value", i)
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
	metaHdr := &apisession.ResponseMetaHeader{Status: &sts, Epoch: x.epoch}
	if sts.Code != 0 {
		return x.sendResponse(stream, 0, &apiobject.SearchResponse{MetaHeader: metaHdr})
	}
	if len(x.idLists) == 0 {
		x.idLists = [][]oid.ID{{}} // to return empty list
	}
	for i := range x.idLists {
		resp := apiobject.SearchResponse{
			Body: &apiobject.SearchResponse_Body{
				IdList: make([]*refs.ObjectID, len(x.idLists[i])),
			},
			MetaHeader: metaHdr,
		}
		for j := range x.idLists[i] {
			resp.Body.IdList[j] = new(refs.ObjectID)
			x.idLists[i][j].WriteToV2(resp.Body.IdList[j])
		}
		if err = x.sendResponse(stream, i, &resp); err != nil {
			return err
		}
	}
	return nil
}

func bindClientServerForObjectSearchWithOpts(t testing.TB, assertErr func(error), customizeOpts func(*Options)) (*searchObjectsServer, *Client, *bool) {
	var srv searchObjectsServer
	srv.sleepDur = 10 * time.Millisecond
	srv.serverSigner = neofscryptotest.RandomSigner()
	srv.latestVersion = versiontest.Version()
	srv.nodeInfo = netmaptest.NodeInfo()
	srv.nodeInfo.SetPublicKey(neofscrypto.PublicKeyBytes(srv.serverSigner.Public()))
	usr, _ := usertest.TwoUsers()
	srv.clientSigner = usr
	srv.filters = make([]object.SearchFilter, 5)
	for i := range srv.filters {
		si := strconv.Itoa(i)
		srv.filters[i] = object.NewSearchFilter("k"+si, object.FilterOp(i), "v"+si)
	}
	ids := oidtest.NIDs(10)
	srv.idLists = [][]oid.ID{ids[:2], ids[2:6], ids[6:9], ids[9:]}

	var opts Options
	var handlerCalled bool
	opts.SetAPIRequestResultHandler(func(nodeKey []byte, endpoint string, op stat.Method, dur time.Duration, err error) {
		handlerCalled = true
		require.Equal(t, srv.nodeInfo.PublicKey(), nodeKey)
		require.Equal(t, "localhost:8080", endpoint)
		require.Equal(t, stat.MethodObjectSearch, op)
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
	apinetmap.RegisterNetmapServiceServer(gs, &srv)
	apiobject.RegisterObjectServiceServer(gs, &srv)
	go func() { _ = gs.Serve(conn) }()
	t.Cleanup(gs.Stop)

	c.dial = func(context.Context, string) (net.Conn, error) { return conn.Dial() }
	require.NoError(t, c.Dial(context.Background()))

	return &srv, c, &handlerCalled
}

func bindClientServerForObjectSearch(t testing.TB, assertErr func(error)) (*searchObjectsServer, *Client, *bool) {
	return bindClientServerForObjectSearchWithOpts(t, assertErr, nil)
}

func testObjectSearchingMethod(t *testing.T, call func(context.Context, *Client, cid.ID, neofscrypto.Signer, SelectObjectsOptions, []object.SearchFilter) ([]oid.ID, error)) {
	ctx := context.Background()
	firstNIDLists := func(all [][]oid.ID, n int) []oid.ID {
		var res []oid.ID
		for i := 0; i < n; i++ {
			res = append(res, all[i]...)
		}
		return res
	}
	checkRes := func(t testing.TB, all [][]oid.ID, nResps int, res []oid.ID) {
		require.Equal(t, firstNIDLists(all, nResps), res)
	}
	checkEmptyRes := func(t testing.TB, res []oid.ID) {
		checkRes(t, nil, 0, res)
	}
	t.Run("invalid signer", func(t *testing.T) {
		c, err := New(anyValidURI, Options{})
		require.NoError(t, err)
		res, err := call(ctx, c, cidtest.ID(), nil, SelectObjectsOptions{}, nil)
		require.ErrorIs(t, err, errMissingSigner)
		checkEmptyRes(t, res)
	})
	t.Run("OK", func(t *testing.T) {
		for _, testCase := range []struct {
			name    string
			setOpts func(srv *searchObjectsServer, opts *SelectObjectsOptions)
		}{
			{name: "default", setOpts: func(srv *searchObjectsServer, opts *SelectObjectsOptions) {}},
			{name: "with session", setOpts: func(srv *searchObjectsServer, opts *SelectObjectsOptions) {
				so := sessiontest.Object()
				opts.WithinSession(so)
				srv.session = &so
			}},
			{name: "with bearer token", setOpts: func(srv *searchObjectsServer, opts *SelectObjectsOptions) {
				bt := bearertest.Token()
				opts.WithBearerToken(bt)
				srv.bearerToken = &bt
			}},
			{name: "no forwarding", setOpts: func(srv *searchObjectsServer, opts *SelectObjectsOptions) {
				srv.local = true
				opts.PreventForwarding()
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				assertErr := func(err error) { require.NoError(t, err) }
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				var opts SelectObjectsOptions
				testCase.setOpts(srv, &opts)
				res, err := call(ctx, c, srv.cnr, srv.clientSigner, opts, srv.filters)
				assertErr(err)
				checkRes(t, srv.idLists, len(srv.idLists), res)
				require.True(t, *handlerCalled)

				srv.idLists = nil
				res, err = call(ctx, c, srv.cnr, srv.clientSigner, opts, srv.filters)
				assertErr(err)
				checkEmptyRes(t, res)
			})
		}
	})
	t.Run("fail", func(t *testing.T) {
		t.Run("wrong stream flow", func(t *testing.T) {
			t.Run("empty stream", func(t *testing.T) {
				assertErr := func(err error) { require.EqualError(t, err, "stream ended without a status response") }
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				srv.emptyStream = true
				res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				checkEmptyRes(t, res)
				require.True(t, *handlerCalled)
			})
			t.Run("message after status error", func(t *testing.T) {
				assertErr := func(err error) {
					require.EqualError(t, err, "stream is not completed after the message #2 which must be the last one")
				}
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == 2 {
						for r.MetaHeader.Status.Code == 0 {
							r.MetaHeader.Status.Code = rand.Uint32()
						}
					}
					return nil
				}
				res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				require.True(t, *handlerCalled)
				checkRes(t, srv.idLists, 2, res)
			})
			t.Run("message after empty payload", func(t *testing.T) {
				assertErr := func(err error) {
					require.EqualError(t, err, "stream is not completed after the message #0 which must be the last one")
				}
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == 0 {
						r.Body = nil
					}
					return nil
				}
				res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				checkEmptyRes(t, res)
				require.True(t, *handlerCalled)
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == 0 {
						r.Body.IdList = nil
					}
					return nil
				}
				res, err = call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				checkEmptyRes(t, res)
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == 0 {
						r.Body.IdList = []*refs.ObjectID{}
					}
					return nil
				}
				res, err = call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				checkEmptyRes(t, res)
			})
		})
		t.Run("sign request", func(t *testing.T) {
			assertErr := func(err error) { require.ErrorContains(t, err, errSignRequest) }
			srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
			srv.sleepDur = 0
			res, err := call(ctx, c, srv.cnr, neofscryptotest.FailSigner(srv.clientSigner), SelectObjectsOptions{}, srv.filters)
			assertErr(err)
			checkEmptyRes(t, res)
			require.True(t, *handlerCalled)
		})
		t.Run("transport", func(t *testing.T) {
			assertErr := func(err error) { require.ErrorContains(t, err, errTransport) }
			srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
			srv.sleepDur = 0
			require.NoError(t, c.conn.Close())
			res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
			assertErr(err)
			checkEmptyRes(t, res)
			require.True(t, *handlerCalled)

			for nResp := range srv.idLists {
				assertErr = func(err error) {
					require.ErrorContains(t, err, errTransport)
					require.ErrorContains(t, err, "any transport error")
					require.ErrorContains(t, err, fmt.Sprintf("while reading response #%d", nResp))
				}
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				srv.sleepDur = 0
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == nResp {
						return errors.New("any transport error")
					}
					return nil
				}
				res, err = call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				checkRes(t, srv.idLists, nResp, res)
				require.True(t, *handlerCalled)
			}
		})
		t.Run("invalid response signature", func(t *testing.T) {
			for i, testCase := range []struct {
				err     string
				corrupt func(*apiobject.SearchResponse)
			}{
				{err: "missing verification header",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader = nil },
				},
				{err: "missing body signature",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.BodySignature = nil },
				},
				{err: "missing signature of the meta header",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.MetaSignature = nil },
				},
				{err: "missing signature of the origin verification header",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.OriginSignature = nil },
				},
				{err: "verify body signature: missing public key",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.BodySignature.Key = nil },
				},
				{err: "verify signature of the meta header: missing public key",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.MetaSignature.Key = nil },
				},
				{err: "verify signature of the origin verification header: missing public key",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.OriginSignature.Key = nil },
				},
				{err: "verify body signature: decode public key from binary",
					corrupt: func(r *apiobject.SearchResponse) {
						r.VerifyHeader.BodySignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the meta header: decode public key from binary",
					corrupt: func(r *apiobject.SearchResponse) {
						r.VerifyHeader.MetaSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify signature of the origin verification header: decode public key from binary",
					corrupt: func(r *apiobject.SearchResponse) {
						r.VerifyHeader.OriginSignature.Key = []byte("not a public key")
					},
				},
				{err: "verify body signature: invalid scheme -1",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.BodySignature.Scheme = -1 },
				},
				{err: "verify body signature: unsupported scheme 3",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.BodySignature.Scheme = 3 },
				},
				{err: "verify signature of the meta header: unsupported scheme 3",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.MetaSignature.Scheme = 3 },
				},
				{err: "verify signature of the origin verification header: unsupported scheme 3",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.OriginSignature.Scheme = 3 },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.BodySignature.Sign[0]++ },
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.MetaSignature.Sign[0]++ },
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.SearchResponse) { r.VerifyHeader.OriginSignature.Sign[0]++ },
				},
				{err: "verify body signature: signature mismatch",
					corrupt: func(r *apiobject.SearchResponse) {
						r.VerifyHeader.BodySignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the meta header: signature mismatch",
					corrupt: func(r *apiobject.SearchResponse) {
						r.VerifyHeader.MetaSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
				{err: "verify signature of the origin verification header: signature mismatch",
					corrupt: func(r *apiobject.SearchResponse) {
						r.VerifyHeader.OriginSignature.Key = neofscrypto.PublicKeyBytes(neofscryptotest.RandomSigner().Public())
					},
				},
			} {
				assertErr := func(err error) {
					require.ErrorContains(t, err, "invalid response")
					require.ErrorContains(t, err, errResponseSignature, [2]any{i, testCase.err})
					require.ErrorContains(t, err, testCase.err, [2]any{i, testCase.err})
				}
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				for nResp := range srv.idLists {
					srv.corruptRespSig = func(n int, r *apiobject.SearchResponse) {
						if n == nResp {
							testCase.corrupt(r)
						}
					}
					res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
					assertErr(err)
					require.ErrorContains(t, err, fmt.Sprintf("invalid response #%d", nResp))
					checkRes(t, srv.idLists, nResp, res)
					require.True(t, *handlerCalled)
				}
			}
		})
		t.Run("invalid response status", func(t *testing.T) {
			assertErr := func(err error) {
				require.ErrorContains(t, err, "invalid response")
				require.ErrorContains(t, err, errInvalidResponseStatus)
				require.ErrorContains(t, err, "details attached but not supported")
			}
			srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
			for nResp := range srv.idLists {
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == nResp {
						r.MetaHeader.Status = &status.Status{Code: status.InternalServerError, Details: make([]*status.Status_Detail, 1)}
					}
					return nil
				}
				res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				require.ErrorContains(t, err, fmt.Sprintf("invalid response #%d", nResp))
				checkRes(t, srv.idLists, nResp, res)
				require.True(t, *handlerCalled)
			}
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
				{code: status.SessionTokenExpired, errConst: apistatus.ErrSessionTokenExpired, errVar: new(apistatus.SessionTokenExpired)},
			} {
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
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
					if n == len(srv.idLists)-1 {
						r.MetaHeader.Status = &status.Status{Code: testCase.code, Message: "any message"}
					}
					return nil
				}
				res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
				assertErr(err)
				checkRes(t, srv.idLists, len(srv.idLists)-1, res)
				require.True(t, *handlerCalled, testCase)
			}
		})
		t.Run("response body", func(t *testing.T) {
			t.Run("missing", func(t *testing.T) {
				assertErr := func(err error) {
					require.ErrorContains(t, err, "invalid response")
					require.ErrorContains(t, err, "empty object ID list is only allowed in the first stream message")
				}
				srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
				for nResp := 1; nResp < len(srv.idLists); nResp++ {
					srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
						if n == nResp {
							r.Body = nil
						}
						return nil
					}
					res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
					assertErr(err)
					require.ErrorContains(t, err, fmt.Sprintf("invalid response #%d", nResp))
					checkRes(t, srv.idLists, nResp, res)
					require.True(t, *handlerCalled)
					srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
						if n == nResp {
							r.Body.IdList = nil
						}
						return nil
					}
					res, err = call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
					assertErr(err)
					checkRes(t, srv.idLists, nResp, res)
					srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
						if n == nResp {
							r.Body.IdList = []*refs.ObjectID{}
						}
						return nil
					}
					res, err = call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
					assertErr(err)
					checkRes(t, srv.idLists, nResp, res)
				}
			})
			t.Run("ID list", func(t *testing.T) {
				validID1 := oidtest.ID()
				validID2 := oidtest.ChangeID(validID1)
				var mValidID1, mValidID2 refs.ObjectID
				validID1.WriteToV2(&mValidID1)
				validID2.WriteToV2(&mValidID2)
				for i, testCase := range []struct {
					err     string
					corrupt func(*apiobject.SearchResponse_Body)
				}{
					{err: "invalid element #1: missing value field", corrupt: func(body *apiobject.SearchResponse_Body) {
						body.IdList = []*refs.ObjectID{&mValidID1, nil, &mValidID2}
					}},
					{err: "invalid element #1: missing value field", corrupt: func(body *apiobject.SearchResponse_Body) {
						body.IdList = []*refs.ObjectID{&mValidID1, {Value: nil}, &mValidID2}
					}},
					{err: "invalid element #1: missing value field", corrupt: func(body *apiobject.SearchResponse_Body) {
						body.IdList = []*refs.ObjectID{&mValidID1, {Value: []byte{}}, &mValidID2}
					}},
					{err: "invalid element #1: invalid value length 31", corrupt: func(body *apiobject.SearchResponse_Body) {
						body.IdList = []*refs.ObjectID{&mValidID1, {Value: make([]byte, 31)}, &mValidID2}
					}},
				} {
					assertErr := func(err error) {
						require.ErrorContains(t, err, "invalid response", i)
						require.ErrorContains(t, err, fmt.Sprintf(": invalid body: invalid field (object ID list): %s", testCase.err), i)
					}
					srv, c, handlerCalled := bindClientServerForObjectSearch(t, assertErr)
					for nResp := range srv.idLists {
						srv.modifyResp = func(n int, r *apiobject.SearchResponse) error {
							if n == nResp {
								testCase.corrupt(r.Body)
							}
							return nil
						}
						res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
						assertErr(err)
						require.ErrorContains(t, err, fmt.Sprintf("invalid response #%d", nResp), i)
						require.Equal(t, append(firstNIDLists(srv.idLists, nResp), validID1), res)
						require.NotContains(t, res, validID2)
						require.True(t, *handlerCalled, testCase.err, i)
					}
				}
			})
		})
	})
	t.Run("response info handler", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			assertErr := func(err error) { require.NoError(t, err) }
			var r ResponseMetaInfo
			callCounter := 0
			srv, c, reqHandlerCalled := bindClientServerForObjectSearchWithOpts(t, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					callCounter++
					r = info
					return nil
				})
			})
			srv.epoch = 3598503
			res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
			assertErr(err)
			checkRes(t, srv.idLists, len(srv.idLists), res)
			require.EqualValues(t, 2, callCounter) // + on dial
			require.Equal(t, srv.epoch, r.Epoch())
			require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), r.ResponderKey())
			require.True(t, *reqHandlerCalled)
		})
		t.Run("fail", func(t *testing.T) {
			assertErr := func(err error) { require.EqualError(t, err, "intercept response info: some handler error") }
			callCounter := 0
			var r ResponseMetaInfo
			srv, c, handlerCalled := bindClientServerForObjectSearchWithOpts(t, assertErr, func(opts *Options) {
				opts.SetAPIResponseInfoInterceptor(func(info ResponseMetaInfo) error {
					callCounter++
					if callCounter == 1 { // dial
						return nil
					}
					r = info
					return errors.New("some handler error")
				})
			})
			srv.epoch = 4386380643
			res, err := call(ctx, c, srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters)
			assertErr(err)
			checkEmptyRes(t, res)
			require.EqualValues(t, 2, callCounter) // + on dial
			require.Equal(t, srv.epoch, r.Epoch())
			require.Equal(t, neofscrypto.PublicKeyBytes(srv.serverSigner.Public()), r.ResponderKey())
			require.True(t, *handlerCalled)
		})
	})
}

func TestClient_SelectObjects(t *testing.T) {
	testObjectSearchingMethod(t, func(ctx context.Context, c *Client, cnr cid.ID, signer neofscrypto.Signer, opts SelectObjectsOptions, filters []object.SearchFilter) ([]oid.ID, error) {
		return c.SelectObjects(ctx, cnr, signer, opts, filters)
	})
}

func TestClient_ForEachSelectedObject(t *testing.T) {
	testObjectSearchingMethod(t, func(ctx context.Context, c *Client, cnr cid.ID, signer neofscrypto.Signer, opts SelectObjectsOptions, filters []object.SearchFilter) ([]oid.ID, error) {
		var res []oid.ID
		return res, c.ForEachSelectedObject(ctx, cnr, signer, opts, filters, func(id oid.ID) bool {
			res = append(res, id)
			return true
		})
	})
	t.Run("break", func(t *testing.T) {
		assertErr := func(err error) { require.NoError(t, err) }
		srv, c, _ := bindClientServerForObjectSearch(t, assertErr)
		srv.idLists = [][]oid.ID{oidtest.NIDs(2), oidtest.NIDs(3), oidtest.NIDs(1)}
		callCounter := 0
		err := c.ForEachSelectedObject(context.Background(), srv.cnr, srv.clientSigner, SelectObjectsOptions{}, srv.filters, func(id oid.ID) bool {
			switch callCounter {
			default:
				return false // break
			case 0:
				require.Equal(t, srv.idLists[0][0], id)
			case 1:
				require.Equal(t, srv.idLists[0][1], id)
			case 2:
				require.Equal(t, srv.idLists[1][0], id)
			case 3:
				require.Equal(t, srv.idLists[1][1], id)
			}
			callCounter++
			return true // continue
		})
		assertErr(err)
		require.EqualValues(t, 4, callCounter)
	})
}
