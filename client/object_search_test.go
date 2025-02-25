package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"slices"
	"strconv"
	"testing"
	"time"

	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func readAllObjectIDs(r *ObjectListReader) ([]oid.ID, error) {
	buf := make([]oid.ID, 32)
	var collected []oid.ID
	for {
		n, err := r.Read(buf)
		collected = append(collected, buf[:n]...)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return collected, err
		}
	}
}

func setChunkInSearchResponse(b *protoobject.SearchResponse_Body, c []oid.ID) *protoobject.SearchResponse_Body {
	b = proto.Clone(b).(*protoobject.SearchResponse_Body)
	b.IdList = make([]*protorefs.ObjectID, len(c))
	for i := range c {
		b.IdList[i] = &protorefs.ObjectID{Value: c[i][:]}
	}
	return b
}

type commonSearchObjectsServerSettings struct {
	protoobject.UnimplementedObjectServiceServer
	testObjectSessionServerSettings
	testBearerTokenServerSettings
	testRequiredContainerIDServerSettings
	testLocalRequestServerSettings
	reqFilters []object.SearchFilter
}

type testSearchObjectsServer struct {
	commonSearchObjectsServerSettings
	testCommonServerStreamServerSettings[
		*protoobject.SearchRequest_Body,
		*protoobject.SearchRequest,
		*protoobject.SearchResponse_Body,
		*protoobject.SearchResponse,
	]
	chunk []oid.ID
}

func TestObjectIterate(t *testing.T) {
	ids := oidtest.IDs(3)
	newTestSearchObjectsStream := func(t testing.TB, code uint32, chunks [][]oid.ID) *ObjectListReader {
		srv := newTestSearchObjectsServer()
		c := newTestObjectClient(t, srv)

		srv.respondWithChunks(chunks)
		srv.respondWithStatus(uint(len(chunks)-1), &protostatus.Status{Code: code})

		r, err := c.ObjectSearchInit(context.Background(), cidtest.ID(), usertest.User(), PrmObjectSearch{})
		require.NoError(t, err)
		return r
	}

	t.Run("no objects", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t, 0, nil)

		var actual []oid.ID
		require.NoError(t, stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Empty(t, actual)
	})
	t.Run("iterate all sequence", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t, 0, [][]oid.ID{ids[0:2], nil, ids[2:3]})

		var actual []oid.ID
		require.NoError(t, stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		}))
		require.Equal(t, ids[:3], actual)
	})
	t.Run("stop by return value", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t, 0, [][]oid.ID{ids})
		var actual []oid.ID
		require.NoError(t, stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return len(actual) == 2
		}))
		require.Equal(t, ids[:2], actual)
	})
	t.Run("stop after error", func(t *testing.T) {
		stream := newTestSearchObjectsStream(t, 1024, [][]oid.ID{ids[:2], ids[2:]})

		var actual []oid.ID
		err := stream.Iterate(func(id oid.ID) bool {
			actual = append(actual, id)
			return false
		})
		require.ErrorIs(t, err, apistatus.ErrServerInternal)
		require.Equal(t, ids[:2], actual)
	})
}

// returns [protoobject.ObjectServiceServer] supporting Search method only.
// Default implementation performs common verification of any request, and
// responds with any valid message stream. Some methods allow to tune the
// behavior.
func newTestSearchObjectsServer() *testSearchObjectsServer { return new(testSearchObjectsServer) }

// makes the server to assert that any request carries given filter set. By
// default, and if nil, any set is accepted.
func (x *commonSearchObjectsServerSettings) checkRequestFilters(fs []object.SearchFilter) {
	x.reqFilters = fs
}

// makes the server to return given chunk of IDs in any response. By default,
// and if nil, some non-empty data is returned.
func (x *testSearchObjectsServer) respondWithChunk(chunk []oid.ID) { x.chunk = chunk }

// makes the server to respond with given chunk responses.
//
// Overrides configured len(chunks) responses.
func (x *testSearchObjectsServer) respondWithChunks(chunks [][]oid.ID) {
	if len(chunks) == 0 {
		x.respondWithBody(0, validMinSearchResponseBody)
		return
	}
	for i := range chunks {
		b := setChunkInSearchResponse(validFullSearchResponseBody, chunks[i])
		x.respondWithBody(uint(i), b)
	}
}

func (x *testSearchObjectsServer) verifyRequest(req *protoobject.SearchRequest) error {
	if err := x.testCommonServerStreamServerSettings.verifyRequest(req); err != nil {
		return err
	}
	return x.commonSearchObjectsServerSettings.verifyRequest(req.MetaHeader, req.Body)
}

func (x commonSearchObjectsServerSettings) verifyRequest(mh *protosession.RequestMetaHeader, body interface {
	GetContainerId() *protorefs.ContainerID
	GetVersion() uint32
	GetFilters() []*protoobject.SearchFilter
}) error {
	// meta header
	// TTL
	if err := x.verifyTTL(mh); err != nil {
		return err
	}
	// session token
	if err := x.verifySessionToken(mh.GetSessionToken()); err != nil {
		return err
	}
	// bearer token
	if err := x.verifyBearerToken(mh.GetBearerToken()); err != nil {
		return err
	}
	// body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. address
	if err := x.verifyRequestContainerID(body.GetContainerId()); err != nil {
		return err
	}
	// 2. version
	if v := body.GetVersion(); v != 1 {
		return newErrInvalidRequestField("version", fmt.Errorf("wrong value (client: 1, message: %d)", v))
	}
	// 3. filters
	if x.reqFilters != nil {
		if err := checkObjectSearchFiltersTransport(x.reqFilters, body.GetFilters()); err != nil {
			return newErrInvalidRequestField("filters", err)
		}
	}
	return nil
}

func (x *testSearchObjectsServer) Search(req *protoobject.SearchRequest, stream protoobject.ObjectService_SearchServer) error {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return err
	}
	if x.handlerErrForced {
		return x.handlerErr
	}
	lastRespInd := uint(1)
	if x.resps != nil {
		lastRespInd = 0
	}
	for n := range x.resps {
		if n > lastRespInd {
			lastRespInd = n
		}
	}
	if x.respErrN > lastRespInd {
		lastRespInd = x.respErrN
	}
	chunk := x.chunk
	if chunk == nil {
		chunk = oidtest.IDs(3)
	}
	for n := range lastRespInd + 1 {
		s := x.resps[n]
		resp := &protoobject.SearchResponse{
			MetaHeader: s.respMeta,
		}
		if s.respBodyForced {
			resp.Body = s.respBody
		} else {
			resp.Body = setChunkInSearchResponse(validFullSearchResponseBody, chunk)
		}
		var err error
		resp.VerifyHeader, err = s.signResponse(resp)
		if err != nil {
			return fmt.Errorf("sign response: %w", err)
		}
		if err := stream.Send(resp); err != nil {
			return fmt.Errorf("send response #%d: %w", n, err)
		}
		if x.respErrN > 0 && n >= x.respErrN-1 {
			return x.respErr
		}
	}
	return nil
}

func TestClient_ObjectSearch(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectSearch
	anyCID := cidtest.ID()
	anyValidSigner := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestSearchObjectsServer()
				c := newTestObjectClient(t, srv)

				srv.checkRequestContainerID(anyCID)
				srv.authenticateRequest(anyValidSigner)
				r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, PrmObjectSearch{})
				require.NoError(t, err)
				_, err = readAllObjectIDs(r)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestSearchObjectsServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, opts)
						if err == nil {
							_, err = readAllObjectIDs(r)
						}
						return err
					})
				})
				t.Run("local", func(t *testing.T) {
					srv := newTestSearchObjectsServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkLocal()

					srv.checkRequestLocal()
					r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = readAllObjectIDs(r)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestSearchObjectsServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = readAllObjectIDs(r)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestSearchObjectsServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = readAllObjectIDs(r)
					require.NoError(t, err)
				})
				t.Run("filters", func(t *testing.T) {
					srv := newTestSearchObjectsServer()
					c := newTestObjectClient(t, srv)

					fs := make(object.SearchFilters, 10)
					fs.AddFilter("k1", "v1", object.MatchStringEqual)
					fs.AddFilter("k1", "v2", object.MatchStringNotEqual)
					fs.AddFilter("k3", "v3", object.MatchNotPresent)
					fs.AddFilter("k4", "v4", object.MatchCommonPrefix)
					fs.AddFilter("k5", "v5", object.MatchNumGT)
					fs.AddFilter("k6", "v6", object.MatchNumGE)
					fs.AddFilter("k7", "v7", object.MatchNumLT)
					fs.AddFilter("k8", "v8", object.MatchNumLE)
					fs.AddFilter("k_max", "v_max", math.MaxInt32)

					opts := anyValidOpts
					opts.SetFilters(fs)

					srv.checkRequestFilters(fs)
					r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = readAllObjectIDs(r)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					const bigChunkSize = (3<<20 + 500<<10) / oid.Size
					bigChunkTwice := oidtest.IDs(bigChunkSize * 2)
					smallChunk := oidtest.IDs(10)
					for _, tc := range []struct {
						name   string
						chunks [][]oid.ID
					}{
						{name: "empty"},
						{name: "with single ID chunk", chunks: [][]oid.ID{smallChunk}},
						{name: "with multiple ID chunks",
							chunks: [][]oid.ID{bigChunkTwice[:bigChunkSize], smallChunk, {}, bigChunkTwice[bigChunkSize:]}},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestSearchObjectsServer()
							c := newTestObjectClient(t, srv)

							srv.respondWithChunks(tc.chunks)
							r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							res, err := readAllObjectIDs(r)
							require.NoError(t, err)
							require.Equal(t, join(tc.chunks), res)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					test := func(t testing.TB, code uint32, assert func(testing.TB, error)) {
						srv := newTestSearchObjectsServer()
						c := newTestObjectClient(t, srv)

						chunks := [][]oid.ID{oidtest.IDs(3), oidtest.IDs(5), oidtest.IDs(4)}

						srv.respondWithChunks(chunks)
						srv.respondWithStatus(uint(len(chunks))-1, &protostatus.Status{Code: code})
						r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
						require.NoError(t, err)
						_, err = readAllObjectIDs(r)
						assert(t, err)
					}
					t.Run("OK", func(t *testing.T) {
						test(t, 0, func(t testing.TB, err error) { require.NoError(t, err) })
					})
					t.Run("failure", func(t *testing.T) {
						var code uint32
						for code == 0 || code == 1024 {
							code = rand.Uint32()
						}
						test(t, code, func(t testing.TB, err error) {
							require.EqualError(t, err, "status: code = unrecognized")
							// TODO: replace after https://github.com/nspcc-dev/neofs-sdk-go/issues/648
							// require.ErrorIs(t, err, apistatus.Error)
						})
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "Search", func(c *Client) error {
						r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
						if err == nil {
							_, err = readAllObjectIDs(r)
						}
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					srv := newTestSearchObjectsServer()
					c := newTestObjectClient(t, srv)

					const n = 10
					chunks := make([][]oid.ID, n)
					for i := range chunks {
						chunks[i] = oidtest.IDs(20)
					}

					srv.respondWithChunks(chunks)
					srv.respondWithoutSigning(n - 1)
					r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
					require.NoError(t, err)
					read, err := readAllObjectIDs(r)
					require.ErrorContains(t, err, "invalid response signature")
					require.Equal(t, join(chunks[:n-1]), read)
				})
				t.Run("payloads", func(t *testing.T) {
					t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/657")
					type testcase = struct {
						name, msg string
						corrupt   func(valid *protoobject.SearchResponse_Body) // with 3 valid IDs
					}
					tcs := []testcase{
						{name: "IDs/nil element", msg: "invalid length 0", corrupt: func(valid *protoobject.SearchResponse_Body) {
							valid.IdList[1] = nil
						}},
					}
					for _, tc := range invalidObjectIDProtoTestcases {
						tcs = append(tcs, testcase{name: "IDs/element/" + tc.name, msg: "invalid ID #1: " + tc.msg,
							corrupt: func(valid *protoobject.SearchResponse_Body) { tc.corrupt(valid.IdList[1]) },
						})
					}

					for _, tc := range tcs {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestSearchObjectsServer()
							c := newTestObjectClient(t, srv)

							b := proto.Clone(validFullSearchResponseBody).(*protoobject.SearchResponse_Body)
							tc.corrupt(b)

							srv.respondWithBody(0, b)
							r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							_, err = readAllObjectIDs(r)
							require.EqualError(t, err, tc.msg)
						})
					}
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ObjectSearchInit(ctx, anyCID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
		t.Run("empty buffer", func(t *testing.T) {
			srv := newTestSearchObjectsServer()
			c := newTestObjectClient(t, srv)

			r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			require.PanicsWithValue(t, "empty buffer in ObjectListReader.ReadList", func() { _, _ = r.Read(nil) })
			require.PanicsWithValue(t, "empty buffer in ObjectListReader.ReadList", func() { _, _ = r.Read([]oid.ID{}) })
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestSearchObjectsServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, err := newClient(t).ObjectSearchInit(ctx, anyCID, usertest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		t.Run("on payload transmission", func(t *testing.T) {
			for _, n := range []uint{0, 2, 10} {
				t.Run(fmt.Sprintf("after %d successes", n), func(t *testing.T) {
					srv := newTestSearchObjectsServer()
					c := newTestObjectClient(t, srv)

					chunk := oidtest.IDs(10)
					transportErr := errors.New("any transport failure")

					srv.respondWithChunk(chunk)
					srv.abortHandlerAfterResponse(n, transportErr)
					r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
					require.NoError(t, err)
					for range n * uint(len(chunk)) {
						_, err = r.Read([]oid.ID{{}})
						require.NoError(t, err)
					}
					_, err = r.Read([]oid.ID{{}})
					assertObjectStreamTransportErr(t, transportErr, err)
				})
			}
		})
		t.Run("too large chunk message", func(t *testing.T) {
			srv := newTestSearchObjectsServer()
			c := newTestObjectClient(t, srv)

			b := setChunkInSearchResponse(validFullSearchResponseBody, make([]oid.ID, 4194304/oid.Size))

			srv.respondWithBody(0, b)
			r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = r.Read([]oid.ID{{}})
			st, ok := status.FromError(err)
			require.True(t, ok, err)
			require.Equal(t, codes.ResourceExhausted, st.Code())
			require.Contains(t, st.Message(), "grpc: received message larger than max (")
			require.Contains(t, st.Message(), " vs. 4194304)")
		})
	})
	t.Run("response callback", func(t *testing.T) {
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/653")
		// TODO: implement
	})
	t.Run("exec statistics", func(t *testing.T) {
		type collectedItem struct {
			pub      []byte
			endpoint string
			mtd      stat.Method
			dur      time.Duration
			err      error
		}
		bind := func() (*testSearchObjectsServer, *Client, *[]collectedItem) {
			srv := newTestSearchObjectsServer()
			svc := newDefaultObjectService(t, srv)
			var collected []collectedItem
			handler := func(pub []byte, endpoint string, mtd stat.Method, dur time.Duration, err error) {
				collected = append(collected, collectedItem{pub: pub, endpoint: endpoint, mtd: mtd, dur: dur, err: err})
			}
			c := newCustomClient(t, func(prm *PrmInit) { prm.SetStatisticCallback(handler) }, svc)
			// [Client.EndpointInfo] is always called to dial the server: this is also submitted
			require.Len(t, collected, 1)
			require.Nil(t, collected[0].pub) // server key is not yet received
			require.Equal(t, testServerEndpoint, collected[0].endpoint)
			require.Equal(t, stat.MethodEndpointInfo, collected[0].mtd)
			require.Positive(t, collected[0].dur)
			require.NoError(t, collected[0].err)
			collected = nil
			return srv, c, &collected
		}
		assertCommon := func(c *[]collectedItem) {
			collected := *c
			for i := range collected {
				require.Equal(t, testServerStateOnDial.pub, collected[i].pub)
				require.Equal(t, testServerEndpoint, collected[i].endpoint)
				require.Positive(t, collected[i].dur)
			}
		}
		t.Run("missing signer", func(t *testing.T) {
			_, c, cl := bind()
			_, err := c.ObjectSearchInit(ctx, anyCID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
			assertCommon(cl)
			collected := *cl
			require.Len(t, *cl, 1)
			require.Equal(t, stat.MethodObjectSearch, collected[0].mtd)
			require.NoError(t, collected[0].err)
		})
		t.Run("sign request", func(t *testing.T) {
			_, c, cl := bind()
			_, err := c.ObjectSearchInit(ctx, anyCID, usertest.FailSigner(anyValidSigner), anyValidOpts)
			assertSignRequestErr(t, err)
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 1)
			require.Equal(t, stat.MethodObjectSearch, collected[0].mtd)
			require.Equal(t, err, collected[0].err)
		})
		t.Run("transport failure", func(t *testing.T) {
			srv, c, cl := bind()
			transportErr := errors.New("any transport failure")
			srv.abortHandlerAfterResponse(3, transportErr)

			r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			for err == nil {
				_, err = r.Read([]oid.ID{{}})
			}
			assertObjectStreamTransportErr(t, transportErr, err)
			assertCommon(cl)
			collected := *cl
			require.Equal(t, stat.MethodObjectSearch, collected[0].mtd)
			require.NoError(t, collected[0].err)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/656")
			require.Len(t, collected, 2) // move upper
			require.Equal(t, stat.MethodObjectSearchStream, collected[1].mtd)
			require.Equal(t, err, collected[1].err)
		})
		t.Run("OK", func(t *testing.T) {
			srv, c, cl := bind()
			const sleepDur = 100 * time.Millisecond
			// duration is pretty short overall, but most likely larger than the exec time w/o sleep
			srv.setSleepDuration(sleepDur)

			r, err := c.ObjectSearchInit(ctx, anyCID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			for err == nil {
				_, err = r.Read([]oid.ID{{}})
			}
			require.ErrorIs(t, err, io.EOF)
			assertCommon(cl)
			collected := *cl
			require.Equal(t, stat.MethodObjectSearch, collected[0].mtd)
			require.NoError(t, collected[0].err)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/656")
			require.Len(t, collected, 2) // move upper
			require.Equal(t, stat.MethodObjectSearchStream, collected[1].mtd)
			require.NoError(t, err, collected[1].err)
			require.Greater(t, collected[1].dur, sleepDur)
		})
	})
}

type testSearchObjectsV2Server struct {
	commonSearchObjectsServerSettings
	testCommonUnaryServerSettings[
		*protoobject.SearchV2Request_Body,
		*protoobject.SearchV2Request,
		*protoobject.SearchV2Response_Body,
		*protoobject.SearchV2Response,
	]
	count     *uint32
	reqCursor *string // response also has cursor
	attrs     []string
}

// returns [protoobject.ObjectServiceServer] supporting SearchV2 method only.
// Default implementation performs common verification of any request and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestSearchObjectsV2Server() *testSearchObjectsV2Server { return new(testSearchObjectsV2Server) }

// makes the server to assert that any request carries given count. By default,
// any valid count is accepted.
func (x *testSearchObjectsV2Server) checkRequestCount(count uint32) { x.count = &count }

// makes the server to assert that any request carries given cursor. By default,
// any cursor is accepted.
func (x *testSearchObjectsV2Server) checkRequestCursor(cursor string) { x.reqCursor = &cursor }

// makes the server to assert that any request carries given attribute set. By
// default, and if nil, any valid filters are accepted.
func (x *testSearchObjectsV2Server) checkRequestAttributes(as []string) { x.attrs = as }

func (x *testSearchObjectsV2Server) verifyRequest(req *protoobject.SearchV2Request) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	if err := x.commonSearchObjectsServerSettings.verifyRequest(req.MetaHeader, req.Body); err != nil {
		return err
	}
	body := req.Body
	// 4. count
	if body.Count == 0 {
		return newErrInvalidRequestField("count", errors.New("zero"))
	}
	if body.Count > 1000 {
		return newErrInvalidRequestField("count", errors.New("limit exceeded"))
	}
	if x.count != nil {
		var expCount uint32
		if *x.count != 0 {
			expCount = *x.count
		} else {
			expCount = 1000
		}
		if body.Count != expCount {
			return newErrInvalidRequestField("count", fmt.Errorf("wrong value (client: %d, message: %d)", expCount, body.Count))
		}
	}
	// 5. cursor
	if x.reqCursor != nil && body.Cursor != *x.reqCursor {
		return newErrInvalidRequestField("cursor", fmt.Errorf("wrong value (client: %q, message: %q)", *x.reqCursor, body.Cursor))
	}
	// 6. attributes
	if x.attrs != nil && !slices.Equal(body.Attributes, x.attrs) {
		return newErrInvalidRequestField("attributes", fmt.Errorf("wrong value (client: %v, message: %v)", x.attrs, body.Attributes))
	}
	for i := range body.Attributes {
		if body.Attributes[i] == "" {
			return newErrInvalidRequestField("attributes", fmt.Errorf("empty element #%d", i))
		}
		for j := i + 1; j < len(body.Attributes); j++ {
			if body.Attributes[i] == body.Attributes[j] {
				return newErrInvalidRequestField("attributes", fmt.Errorf("duplicated attribute %q", body.Attributes[i]))
			}
		}
	}
	if len(body.Attributes) > 0 {
		if !slices.ContainsFunc(body.Filters, func(f *protoobject.SearchFilter) bool { return f.GetKey() == body.Attributes[0] }) {
			return newErrInvalidRequestField("attributes", fmt.Errorf("attribute %q is requested but not filtered", body.Attributes[0]))
		}
	}
	return nil
}

func (x *testSearchObjectsV2Server) SearchV2(_ context.Context, req *protoobject.SearchV2Request) (*protoobject.SearchV2Response, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoobject.SearchV2Response{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		count := min(uint32(len(validProtoObjectIDs)), req.Body.Count)
		resp.Body = &protoobject.SearchV2Response_Body{
			Result: make([]*protoobject.SearchV2Response_OIDWithMeta, count),
			Cursor: req.Body.Cursor + "_next",
		}
		for i := range resp.Body.Result {
			resp.Body.Result[i] = &protoobject.SearchV2Response_OIDWithMeta{
				Id:         proto.Clone(validProtoObjectIDs[i]).(*protorefs.ObjectID),
				Attributes: make([]string, len(req.Body.Attributes)),
			}
			for j := range req.Body.Attributes {
				resp.Body.Result[i].Attributes[j] = "val_" + strconv.Itoa(i) + "_" + strconv.Itoa(j)
			}
		}
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func assertSearchV2ResponseTransport(t testing.TB, body *protoobject.SearchV2Response_Body, items []SearchResultItem, cursor string) {
	require.Equal(t, body.GetCursor(), cursor)
	r := body.GetResult()
	require.Len(t, items, len(r))
	for i := range r {
		require.NoError(t, checkObjectIDTransport(items[i].ID, r[i].Id), i)
		require.Equal(t, r[i].Attributes, items[i].Attributes)
	}
}

func TestClient_SearchObjects(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts SearchObjectsOptions
	anyCID := cidtest.ID()
	const anyRequestCursor = ""
	var anyValidFilters object.SearchFilters
	var anyValidAttrs []string
	anyValidSigner := usertest.User()
	okConn := newTestObjectClient(t, newTestSearchObjectsV2Server())

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestSearchObjectsV2Server()
				c := newTestObjectClient(t, srv)
				const anyRequestCursor = "any_request_cursor"

				reqAttrs := make([]string, 4)
				for i := range reqAttrs {
					reqAttrs[i] = "attr_" + strconv.Itoa(i)
				}

				var fs object.SearchFilters
				fs.AddFilter(reqAttrs[0], "any_val", 100)
				fs.AddFilter("k1", "v1", object.MatchStringEqual)
				fs.AddFilter("k1", "v2", object.MatchStringNotEqual)
				fs.AddFilter("k3", "v3", object.MatchNotPresent)
				fs.AddFilter("k4", "v4", object.MatchCommonPrefix)
				fs.AddFilter("k5", "v5", object.MatchNumGT)
				fs.AddFilter("k6", "v6", object.MatchNumGE)
				fs.AddFilter("k7", "v7", object.MatchNumLT)

				nItems := min(len(validProtoObjectIDs), 1000)
				respBody := &protoobject.SearchV2Response_Body{
					Result: make([]*protoobject.SearchV2Response_OIDWithMeta, nItems),
					Cursor: "any_response_cursor",
				}
				for i := range respBody.Result {
					respBody.Result[i] = &protoobject.SearchV2Response_OIDWithMeta{
						Id:         validProtoObjectIDs[i],
						Attributes: make([]string, len(reqAttrs)),
					}
					for j := range respBody.Result[i].Attributes {
						respBody.Result[i].Attributes[j] = "val_" + strconv.Itoa(i) + "_" + strconv.Itoa(j)
					}
				}

				srv.respondWithBody(respBody)

				srv.checkRequestContainerID(anyCID)
				srv.checkRequestFilters(fs)
				srv.checkRequestAttributes(reqAttrs)
				srv.checkRequestCursor(anyRequestCursor)
				srv.authenticateRequest(anyValidSigner)
				items, cursor, err := c.SearchObjects(ctx, anyCID, fs, reqAttrs, anyRequestCursor, anyValidSigner, SearchObjectsOptions{})
				require.NoError(t, err)
				assertSearchV2ResponseTransport(t, respBody, items, cursor)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestSearchObjectsV2Server, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, opts)
						return err
					})
				})
				t.Run("disable forwarding", func(t *testing.T) {
					srv := newTestSearchObjectsV2Server()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.DisableForwarding()

					srv.checkRequestLocal()
					_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestSearchObjectsV2Server()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithSessionToken(st)

					srv.checkRequestSessionToken(st)
					_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestSearchObjectsV2Server()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("count", func(t *testing.T) {
					srv := newTestSearchObjectsV2Server()
					c := newTestObjectClient(t, srv)
					count := rand.Uint32() % 1001

					opts := anyValidOpts
					opts.SetCount(count)

					srv.checkRequestCount(count)
					_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protoobject.SearchV2Response_Body
					}{
						{name: "nil", body: nil},
						{name: "min", body: validMinSearchV2ResponseBody},
						{name: "full", body: validFullSearchV2ResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestSearchObjectsV2Server()
							c := newTestObjectClient(t, srv)

							var as []string
							var fs object.SearchFilters
							if r := tc.body.GetResult(); len(r) > 0 {
								if n := len(r[0].GetAttributes()); n > 0 {
									as = make([]string, n)
									for i := range as {
										as[i] = "attr_" + strconv.Itoa(i)
									}

									fs.AddFilter(as[0], "any_val", 100)
								}
							}

							srv.respondWithBody(tc.body)
							items, cursor, err := c.SearchObjects(ctx, anyCID, fs, as, anyRequestCursor, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							assertSearchV2ResponseTransport(t, tc.body, items, cursor)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestSearchObjectsV2Server, newTestObjectClient, func(c *Client) error {
						_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "SearchV2", func(c *Client) error {
						_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestSearchObjectsV2Server, newTestObjectClient, func(c *Client) error {
						_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = struct {
						name, msg  string
						count      uint32
						reqCursor  string
						respCursor string
						attrs      []string
						items      []*protoobject.SearchV2Response_OIDWithMeta
					}
					tcs := []testcase{
						{name: "cursor/without items", msg: "invalid cursor field in the response: set while result is empty",
							respCursor: "any_cursor", items: nil,
						},
						{name: "cursor/repeated", msg: "invalid cursor field in the response: repeats the initial one",
							reqCursor: "req_cursor", respCursor: "req_cursor", items: []*protoobject.SearchV2Response_OIDWithMeta{
								{Id: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID)}, nil,
							},
						},
						{name: "items/limit exceeded", msg: "invalid result field in the response: more items than requested: 3",
							count: 2, items: validFullSearchV2ResponseBody.Result, // 3 items
						},
						{name: "items/nil element", msg: "invalid result field in the response: invalid element #1: missing ID",
							items: []*protoobject.SearchV2Response_OIDWithMeta{
								{Id: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID)}, nil,
							},
						},
						{name: "items/element/missing ID", msg: "invalid result field in the response: invalid element #1: missing ID",
							items: []*protoobject.SearchV2Response_OIDWithMeta{
								{Id: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID)}, {},
							},
						},
						{name: "items/element/lack of attributes", msg: "invalid result field in the response: invalid element #1: wrong attribute count 1",
							attrs: []string{"a1", "a2"}, items: []*protoobject.SearchV2Response_OIDWithMeta{
								{Id: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID), Attributes: []string{"val_1_1", "val_1_2"}},
								{Id: proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID), Attributes: []string{"val_2_1"}},
							},
						},
						{name: "items/element/excess of attributes", msg: "invalid result field in the response: invalid element #1: wrong attribute count 3",
							attrs: []string{"a1", "a2"}, items: []*protoobject.SearchV2Response_OIDWithMeta{
								{Id: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID), Attributes: []string{"val_1_1", "val_1_2"}},
								{Id: proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID), Attributes: []string{"val_2_1", "val_2_2", "val_2_3"}},
							},
						},
					}
					for _, tc := range invalidObjectIDProtoTestcases {
						id := proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID)
						tc.corrupt(id)
						tcs = append(tcs, testcase{
							name: "items/element/ID/" + tc.name,
							msg:  "invalid result field in the response: invalid element #1: invalid ID: " + tc.msg,
							items: []*protoobject.SearchV2Response_OIDWithMeta{
								{Id: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID)}, {Id: id}},
						})
					}

					for _, tc := range tcs {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestSearchObjectsV2Server()
							c := newTestObjectClient(t, srv)

							if tc.count == 0 {
								tc.count = max(uint32(len(tc.items)), 1)
							}

							opts := anyValidOpts
							opts.SetCount(tc.count)

							var fs object.SearchFilters
							if len(tc.attrs) > 0 {
								fs.AddFilter(tc.attrs[0], "any_val", 100)
							}

							srv.respondWithBody(&protoobject.SearchV2Response_Body{
								Result: tc.items,
								Cursor: tc.respCursor,
							})
							_, _, err := c.SearchObjects(ctx, anyCID, fs, tc.attrs, tc.reqCursor, anyValidSigner, opts)
							require.EqualError(t, err, tc.msg)
						})
					}
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		t.Run("zero container", func(t *testing.T) {
			_, _, err := okConn.SearchObjects(ctx, cid.ID{}, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
			require.ErrorIs(t, err, cid.ErrZero)
		})
		t.Run("count", func(t *testing.T) {
			opts := anyValidOpts
			opts.SetCount(1001)
			_, _, err := okConn.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, opts)
			require.EqualError(t, err, "count is out of [1, 1000] range")
		})
		t.Run("missing signer", func(t *testing.T) {
			_, _, err := okConn.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
		t.Run("filters", func(t *testing.T) {
			t.Run("limit exceeded", func(t *testing.T) {
				_, _, err := okConn.SearchObjects(ctx, anyCID, make(object.SearchFilters, 9), anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "more than 8 filters")
			})
			t.Run("missing 1st requested attribute", func(t *testing.T) {
				as := []string{"a1", "a2", "a3"}
				var fs object.SearchFilters
				for i := range as {
					fs.AddFilter(as[i], "any_val", 100)
				}

				_, _, err := okConn.SearchObjects(ctx, anyCID, fs, as[1:], anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, `1st attribute "a2" is requested but not filtered 1st`)
				_, _, err = okConn.SearchObjects(ctx, anyCID, nil, as[2:], anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, `1st attribute "a3" is requested but not filtered 1st`)
			})
			t.Run("prohibited", func(t *testing.T) {
				var fs object.SearchFilters
				fs.AddFilter("attr", "val", object.MatchStringEqual)
				fs.AddObjectContainerIDFilter(object.SearchMatchType(rand.Int31()), cidtest.ID())
				_, _, err := okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: prohibited attribute $Object:containerID")
				fs = fs[:1]
				fs.AddObjectIDFilter(object.SearchMatchType(rand.Int31()), oidtest.ID())
				_, _, err = okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: prohibited attribute $Object:objectID")
			})
			t.Run("empty", func(t *testing.T) {
				var fs object.SearchFilters
				fs.AddFilter("attr", "val", object.MatchStringEqual)
				fs.AddFilter("", "val", object.SearchMatchType(rand.Int31()))
				_, _, err := okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: missing attribute")
			})
			t.Run("key-only", func(t *testing.T) {
				var fs object.SearchFilters
				fs.AddFilter("attr", "val", object.MatchStringEqual)
				fs.AddFilter(object.FilterRoot, "", 123)
				_, _, err := okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: non-zero matcher 123 for attribute $Object:ROOT")
				fs = fs[:1]
				fs.AddFilter(object.FilterRoot, "val", 0)
				_, _, err = okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: value for attribute $Object:ROOT is prohibited")
				fs = fs[:1]
				fs.AddFilter(object.FilterPhysical, "", 123)
				_, _, err = okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: non-zero matcher 123 for attribute $Object:PHY")
				fs = fs[:1]
				fs.AddFilter(object.FilterPhysical, "val", 0)
				_, _, err = okConn.SearchObjects(ctx, anyCID, fs, []string{"attr"}, anyRequestCursor, anyValidSigner, anyValidOpts)
				require.EqualError(t, err, "invalid filter #1: value for attribute $Object:PHY is prohibited")
			})
		})
		t.Run("attributes", func(t *testing.T) {
			for _, tc := range []struct {
				name, err string
				as        []string
			}{
				{name: "empty", err: "empty attribute #1", as: []string{"a1", "", "a3"}},
				{name: "duplicated", err: `duplicated attribute "a2"`, as: []string{"a1", "a2", "a3", "a2"}},
				{name: "limit exceeded", err: "more than 8 attributes", as: []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9"}},
				{name: "prohibited/CID", err: "prohibited attribute $Object:containerID", as: []string{object.FilterContainerID}},
				{name: "prohibited/OID", err: "prohibited attribute $Object:objectID", as: []string{object.FilterID}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					_, _, err := okConn.SearchObjects(ctx, anyCID, anyValidFilters, tc.as, anyRequestCursor, anyValidSigner, anyValidOpts)
					require.EqualError(t, err, tc.err)
				})
			}
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestSearchObjectsV2Server, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, _, err := okConn.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor,
			neofscryptotest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestSearchObjectsV2Server, newTestObjectClient, func(c *Client) error {
			_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestSearchObjectsV2Server, newDefaultObjectService, func(c *Client) error {
			_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		var statFailures []testedClientOp
		for _, in := range []struct {
			cnr     cid.ID
			count   uint32
			signer  neofscrypto.Signer
			filters object.SearchFilters
			attrs   []string
		}{
			{cnr: cid.ID{}, signer: anyValidSigner},
			{cnr: anyCID, count: 1001, signer: anyValidSigner},
			{cnr: anyCID, signer: neofscryptotest.FailSigner(anyValidSigner)},
			{cnr: anyCID, signer: anyValidSigner, filters: make(object.SearchFilters, 9)},
			{cnr: anyCID, signer: anyValidSigner, attrs: []string{"a1", "a2", "a3", "a4", "a5"}},
			{cnr: anyCID, signer: anyValidSigner, attrs: []string{"a1", "", "a3"}},
			{cnr: anyCID, signer: anyValidSigner, attrs: []string{"a1", "a2", "a3", "a2"}},
			{cnr: anyCID, signer: anyValidSigner, filters: nil, attrs: []string{"a1", "a2"}},
		} {
			statFailures = append(statFailures, func(c *Client) error {
				opts := anyValidOpts
				opts.SetCount(in.count)
				_, _, err := c.SearchObjects(ctx, in.cnr, in.filters, in.attrs, anyRequestCursor, in.signer, opts)
				return err
			})
		}

		testStatistic(t, newTestSearchObjectsV2Server, newDefaultObjectService, stat.MethodObjectSearchV2, []testedClientOp{
			func(c *Client) error {
				_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, nil, anyValidOpts)
				return err
			},
		}, statFailures, func(c *Client) error {
			_, _, err := c.SearchObjects(ctx, anyCID, anyValidFilters, anyValidAttrs, anyRequestCursor, anyValidSigner, anyValidOpts)
			return err
		})
	})
}
