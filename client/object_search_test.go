package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"testing"
	"time"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
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

type testSearchObjectsServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonServerStreamServerSettings[
		*protoobject.SearchRequest_Body,
		v2object.SearchRequestBody,
		*v2object.SearchRequestBody,
		*protoobject.SearchRequest,
		v2object.SearchRequest,
		*v2object.SearchRequest,
		*protoobject.SearchResponse_Body,
		v2object.SearchResponseBody,
		*v2object.SearchResponseBody,
		*protoobject.SearchResponse,
		v2object.SearchResponse,
		*v2object.SearchResponse,
	]
	testObjectSessionServerSettings
	testBearerTokenServerSettings
	testRequiredContainerIDServerSettings
	testLocalRequestServerSettings
	chunk      []oid.ID
	reqFilters []object.SearchFilter
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
func (x *testSearchObjectsServer) checkRequestFilters(fs []object.SearchFilter) { x.reqFilters = fs }

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
	// meta header
	// TTL
	if err := x.verifyTTL(req.MetaHeader); err != nil {
		return err
	}
	// session token
	if err := x.verifySessionToken(req.MetaHeader.GetSessionToken()); err != nil {
		return err
	}
	// bearer token
	if err := x.verifyBearerToken(req.MetaHeader.GetBearerToken()); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. address
	if err := x.verifyRequestContainerID(body.ContainerId); err != nil {
		return err
	}
	// 2. version
	if body.Version != 1 {
		return newErrInvalidRequestField("version", fmt.Errorf("wrong value (client: 1, message: %d)", body.Version))
	}
	// 3. filters
	if x.reqFilters != nil {
		if err := checkObjectSearchFiltersTransport(x.reqFilters, body.Filters); err != nil {
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
					bt.SetEACLTable(anyValidEACL) // TODO: drop after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
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
					t.Skip("")
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
