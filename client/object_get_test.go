package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"testing"
	"testing/iotest"
	"time"

	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	protoobject "github.com/nspcc-dev/neofs-api-go/v2/object/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func setPayloadLengthInHeadingGetResponse(b *protoobject.GetResponse_Body, ln uint64) *protoobject.GetResponse_Body {
	b = proto.Clone(b).(*protoobject.GetResponse_Body)
	in := b.GetInit()
	if in == nil {
		in = new(protoobject.GetResponse_Body_Init)
		b.ObjectPart = &protoobject.GetResponse_Body_Init_{Init: in}
	}
	h := in.GetHeader()
	if h == nil {
		h = new(protoobject.Header)
		in.Header = h
	}
	h.PayloadLength = ln
	return b
}

func setChunkInGetResponse(b *protoobject.GetResponse_Body, c []byte) *protoobject.GetResponse_Body {
	b = proto.Clone(b).(*protoobject.GetResponse_Body)
	b.ObjectPart.(*protoobject.GetResponse_Body_Chunk).Chunk = c
	return b
}

func setChunkInRangeResponse(b *protoobject.GetRangeResponse_Body, c []byte) *protoobject.GetRangeResponse_Body {
	b = proto.Clone(b).(*protoobject.GetRangeResponse_Body)
	b.RangePart.(*protoobject.GetRangeResponse_Body_Chunk).Chunk = c
	return b
}

func checkSuccessfulGetObjectTransport(t testing.TB, hb *protoobject.GetResponse_Body, payload []byte, h object.Object, r io.Reader, err error) {
	require.NoError(t, err)
	require.NoError(t, iotest.TestReader(r, payload))
	id := h.GetID()
	require.False(t, id.IsZero())
	in := hb.GetInit()
	require.NoError(t, checkObjectIDTransport(id, in.GetObjectId()))
	require.NoError(t, checkObjectHeaderWithSignatureTransport(h, &protoobject.HeaderWithSignature{
		Header:    in.GetHeader(),
		Signature: in.GetSignature(),
	}))
}

type testCommonReadObjectRequestServerSettings struct {
	testObjectSessionServerSettings
	testBearerTokenServerSettings
	testObjectAddressServerSettings
	testLocalRequestServerSettings
	reqRaw bool
}

// makes the server to assert that any request is with set raw flag. By default,
// the flag must be unset.
func (x *testCommonReadObjectRequestServerSettings) checkRequestRaw() { x.reqRaw = true }

func (x *testCommonReadObjectRequestServerSettings) verifyRawFlag(raw bool) error {
	if x.reqRaw != raw {
		return newErrInvalidRequestField("raw flag", fmt.Errorf("unexpected value (client: %t, message: %t)",
			x.reqRaw, raw))
	}
	return nil
}

func (x *testCommonReadObjectRequestServerSettings) verifyMeta(m *protosession.RequestMetaHeader) error {
	// TTL
	if err := x.verifyTTL(m); err != nil {
		return err
	}
	// session token
	if err := x.verifySessionToken(m.GetSessionToken()); err != nil {
		return err
	}
	// bearer token
	if err := x.verifyBearerToken(m.GetBearerToken()); err != nil {
		return err
	}
	return nil
}

type testGetObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonServerStreamServerSettings[
		*protoobject.GetRequest_Body,
		apiobject.GetRequestBody,
		*apiobject.GetRequestBody,
		*protoobject.GetRequest,
		apiobject.GetRequest,
		*apiobject.GetRequest,
		*protoobject.GetResponse_Body,
		apiobject.GetResponseBody,
		*apiobject.GetResponseBody,
		*protoobject.GetResponse,
		apiobject.GetResponse,
		*apiobject.GetResponse,
	]
	testCommonReadObjectRequestServerSettings
	chunk []byte
}

// returns [protoobject.ObjectServiceServer] supporting Get method only. Default
// implementation performs common verification of any request, and responds with
// any valid message stream. Some methods allow to tune the behavior.
func newTestGetObjectServer() *testGetObjectServer { return new(testGetObjectServer) }

// makes the server to return given chunk in any chunk response. By default, and
// if nil, some non-empty data chunk is returned.
func (x *testGetObjectServer) respondWithChunk(chunk []byte) { x.chunk = chunk }

// makes the server to respond with given heading part and chunk responses.
// Returns heading response message.
//
// Overrides configured len(chunks)+1 responses.
func (x *testGetObjectServer) respondWithObject(h *protoobject.GetResponse_Body_Init, chunks [][]byte) *protoobject.GetResponse_Body {
	var ln uint64
	for i := range chunks {
		b := setChunkInGetResponse(validFullChunkObjectGetResponseBody, chunks[i])
		x.respondWithBody(uint(i)+1, b)
		ln += uint64(len(chunks[i]))
	}
	b := setPayloadLengthInHeadingGetResponse(&protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_Init_{Init: h},
	}, ln)
	x.respondWithBody(0, b)
	return b
}

func (x *testGetObjectServer) verifyRequest(req *protoobject.GetRequest) error {
	if err := x.testCommonServerStreamServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	if err := x.verifyMeta(req.MetaHeader); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. address
	if err := x.verifyObjectAddress(body.Address); err != nil {
		return err
	}
	// 2. raw
	return x.verifyRawFlag(body.Raw)
}

func (x *testGetObjectServer) Get(req *protoobject.GetRequest, stream protoobject.ObjectService_GetServer) error {
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
		chunk = []byte("Hello, world!")
	}
	for n := range lastRespInd + 1 {
		s := x.resps[n]
		resp := &protoobject.GetResponse{
			MetaHeader: s.respMeta,
		}
		if s.respBodyForced {
			resp.Body = s.respBody
		} else {
			if n == 0 {
				resp.Body = proto.Clone(validFullHeadingObjectGetResponseBody).(*protoobject.GetResponse_Body)
				if lastRespInd > 0 {
					resp.Body = setPayloadLengthInHeadingGetResponse(resp.Body, uint64(lastRespInd-1)*uint64(len(chunk)))
				}
			} else {
				resp.Body = setChunkInGetResponse(validFullChunkObjectGetResponseBody, chunk)
			}
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

type testGetObjectPayloadRangeServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonServerStreamServerSettings[
		*protoobject.GetRangeRequest_Body,
		apiobject.GetRangeRequestBody,
		*apiobject.GetRangeRequestBody,
		*protoobject.GetRangeRequest,
		apiobject.GetRangeRequest,
		*apiobject.GetRangeRequest,
		*protoobject.GetRangeResponse_Body,
		apiobject.GetRangeResponseBody,
		*apiobject.GetRangeResponseBody,
		*protoobject.GetRangeResponse,
		apiobject.GetRangeResponse,
		*apiobject.GetRangeResponse,
	]
	testCommonReadObjectRequestServerSettings
	chunk  []byte
	reqRng *protoobject.Range
}

// returns [protoobject.ObjectServiceServer] supporting GetRange method only.
// Default implementation performs common verification of any request, and
// responds with any valid message stream. Some methods allow to tune the
// behavior.
func newTestObjectPayloadRangeServer() *testGetObjectPayloadRangeServer {
	return new(testGetObjectPayloadRangeServer)
}

// makes the server to assert that any request carries given range. By default,
// any valid range is accepted.
func (x *testGetObjectPayloadRangeServer) checkRequestRange(off, ln uint64) {
	x.reqRng = &protoobject.Range{Offset: off, Length: ln}
}

// makes the server to return given chunk in any chunk response. By default, and
// if nil, some non-empty data chunk is returned.
func (x *testGetObjectPayloadRangeServer) respondWithChunk(chunk []byte) { x.chunk = chunk }

// makes the server to respond with given chunk responses.
//
// Overrides configured len(chunks) responses.
func (x *testGetObjectPayloadRangeServer) respondWithChunks(chunks [][]byte) {
	for i := range chunks {
		b := setChunkInRangeResponse(validFullChunkObjectRangeResponseBody, chunks[i])
		x.respondWithBody(uint(i), b)
	}
}

func (x *testGetObjectPayloadRangeServer) verifyRequest(req *protoobject.GetRangeRequest) error {
	if err := x.testCommonServerStreamServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	if err := x.verifyMeta(req.MetaHeader); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. address
	if err := x.verifyObjectAddress(body.Address); err != nil {
		return err
	}
	// 2. range
	if body.Range == nil {
		return newErrMissingRequestBodyField("range")
	}
	if body.Range.Length == 0 {
		return newErrInvalidRequestField("range", errors.New("zero length"))
	}
	if x.reqRng != nil {
		if v1, v2 := x.reqRng.GetOffset(), body.Range.GetOffset(); v1 != v2 {
			return newErrInvalidRequestField("range", fmt.Errorf("offset (client: %d, message: %d)", v1, v2))
		}
		if v1, v2 := x.reqRng.GetLength(), body.Range.GetLength(); v1 != v2 {
			return newErrInvalidRequestField("range", fmt.Errorf("length (client: %d, message: %d)", v1, v2))
		}
	}
	// 3. raw
	return x.verifyRawFlag(body.Raw)
}

func (x *testGetObjectPayloadRangeServer) GetRange(req *protoobject.GetRangeRequest, stream protoobject.ObjectService_GetRangeServer) error {
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
		chunk = []byte("Hello, world!")
	}
	for n := range lastRespInd + 1 {
		s := x.resps[n]
		resp := &protoobject.GetRangeResponse{
			MetaHeader: s.respMeta,
		}
		if s.respBodyForced {
			resp.Body = s.respBody
		} else {
			resp.Body = setChunkInRangeResponse(validFullChunkObjectRangeResponseBody, chunk)
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

type testHeadObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonUnaryServerSettings[
		*protoobject.HeadRequest_Body,
		apiobject.HeadRequestBody,
		*apiobject.HeadRequestBody,
		*protoobject.HeadRequest,
		apiobject.HeadRequest,
		*apiobject.HeadRequest,
		*protoobject.HeadResponse_Body,
		apiobject.HeadResponseBody,
		*apiobject.HeadResponseBody,
		*protoobject.HeadResponse,
		apiobject.HeadResponse,
		*apiobject.HeadResponse,
	]
	testCommonReadObjectRequestServerSettings
}

// returns [protoobject.ObjectServiceServer] supporting Head method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestHeadObjectServer() *testHeadObjectServer {
	return new(testHeadObjectServer)
}

func (x *testHeadObjectServer) verifyRequest(req *protoobject.HeadRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	if err := x.verifyMeta(req.MetaHeader); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// 1. address
	if err := x.verifyObjectAddress(body.Address); err != nil {
		return err
	}
	// 2. main only
	if body.MainOnly {
		return newErrInvalidRequestField("main only flag", fmt.Errorf("unexpected value (client: %t, message: %t)", false, body.MainOnly))
	}
	// 3. raw
	return x.verifyRawFlag(body.Raw)
}

func (x *testHeadObjectServer) Head(_ context.Context, req *protoobject.HeadRequest) (*protoobject.HeadResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protoobject.HeadResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinObjectHeadResponseBody).(*protoobject.HeadResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_ObjectHead(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectHead
	anyCID := cidtest.ID()
	anyOID := oidtest.ID()
	anyValidSigner := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestHeadObjectServer()
				c := newTestObjectClient(t, srv)

				srv.checkRequestObjectAddress(anyCID, anyOID)
				srv.authenticateRequest(anyValidSigner)
				_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, PrmObjectHead{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestHeadObjectServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, opts)
						return err
					})
				})
				t.Run("local", func(t *testing.T) {
					srv := newTestHeadObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkLocal()

					srv.checkRequestLocal()
					_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("raw", func(t *testing.T) {
					srv := newTestHeadObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkRaw()

					srv.checkRequestRaw()
					_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestHeadObjectServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestHeadObjectServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					bt.SetEACLTable(anyValidEACL) // TODO: drop after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					type testcase = struct {
						name   string
						body   *protoobject.HeadResponse_Body
						assert func(testing.TB, *protoobject.HeadResponse_Body, object.Object, error)
					}
					var tcs []testcase
					for _, tc := range []struct {
						name string
						body *protoobject.HeadResponse_Body
					}{
						{name: "min", body: validMinObjectSplitInfoHeadResponseBody},
						{name: "full", body: validFullObjectSplitInfoHeadResponseBody},
					} {
						tcs = append(tcs, testcase{name: "split info/" + tc.name, body: tc.body,
							assert: func(t testing.TB, body *protoobject.HeadResponse_Body, _ object.Object, err error) {
								var e *object.SplitInfoError
								require.ErrorAs(t, err, &e)
								require.NoError(t, checkSplitInfoTransport(*e.SplitInfo(), body.GetSplitInfo()))
							}})
					}
					for _, tc := range []struct {
						name string
						body *protoobject.HeadResponse_Body
					}{
						{name: "min", body: validMinObjectHeadResponseBody},
						{name: "full", body: validFullObjectHeadResponseBody},
					} {
						tcs = append(tcs, testcase{name: "header with signature/" + tc.name, body: tc.body,
							assert: func(t testing.TB, body *protoobject.HeadResponse_Body, hdr object.Object, err error) {
								require.NoError(t, err)
								require.NoError(t, checkObjectHeaderWithSignatureTransport(hdr, body.GetHeader()))
							}})
					}
					for _, tc := range tcs {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestHeadObjectServer()
							c := newTestObjectClient(t, srv)

							srv.respondWithBody(tc.body)
							hdr, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
							if err != nil {
								tc.assert(t, tc.body, object.Object{}, err)
							} else {
								tc.assert(t, tc.body, *hdr, err)
							}
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestHeadObjectServer, newTestObjectClient, func(c *Client) error {
						_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "Head", func(c *Client) error {
						_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestHeadObjectServer, newTestObjectClient, func(c *Client) error {
						_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protoobject.HeadResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "unexpected header type <nil>")
							}},
						{name: "empty", body: new(protoobject.HeadResponse_Body),
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "unexpected header type <nil>")
							}},
						{name: "short header oneof/nil", body: &protoobject.HeadResponse_Body{Head: (*protoobject.HeadResponse_Body_ShortHeader)(nil)},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "unexpected header type <nil>")
							}},
						{name: "short header oneof/empty", body: &protoobject.HeadResponse_Body{Head: new(protoobject.HeadResponse_Body_ShortHeader)},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "unexpected header type *object.ShortHeader")
							}},
						{name: "split info oneof/nil", body: &protoobject.HeadResponse_Body{Head: (*protoobject.HeadResponse_Body_SplitInfo)(nil)},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "unexpected header type <nil>")
							}},
						{name: "split info oneof/empty", body: &protoobject.HeadResponse_Body{Head: new(protoobject.HeadResponse_Body_SplitInfo)},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid split info: neither link object ID nor last part object ID is set")
							}},
						{name: "header oneof/nil", body: &protoobject.HeadResponse_Body{Head: (*protoobject.HeadResponse_Body_Header)(nil)},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "unexpected header type <nil>")
							}},
						{name: "header oneof/empty", body: &protoobject.HeadResponse_Body{Head: new(protoobject.HeadResponse_Body_Header)},
							assertErr: func(t testing.TB, err error) {
								require.ErrorAs(t, err, new(MissingResponseFieldErr))
								require.EqualError(t, err, "missing signature field in the response")
							}},
						{name: "header oneof/missing header", body: &protoobject.HeadResponse_Body{Head: &protoobject.HeadResponse_Body_Header{
							Header: &protoobject.HeaderWithSignature{
								Signature: proto.Clone(validMinProtoSignature).(*protorefs.Signature),
							},
						}},
							assertErr: func(t testing.TB, err error) {
								require.ErrorAs(t, err, new(MissingResponseFieldErr))
								require.EqualError(t, err, "missing header field in the response")
							}},
						{name: "header oneof/missing signature", body: &protoobject.HeadResponse_Body{
							Head: &protoobject.HeadResponse_Body_Header{
								Header: &protoobject.HeaderWithSignature{
									Header: proto.Clone(validMinObjectHeader).(*protoobject.Header),
								},
							}},
							assertErr: func(t testing.TB, err error) {
								require.ErrorAs(t, err, new(MissingResponseFieldErr))
								require.EqualError(t, err, "missing signature field in the response")
							}},
					}
					for _, tc := range invalidObjectHeaderProtoTestcases {
						hdr := proto.Clone(validFullObjectHeader).(*protoobject.Header)
						tc.corrupt(hdr)
						tcs = append(tcs, testcase{
							name: "header oneof/header/" + tc.name,
							body: &protoobject.HeadResponse_Body{Head: &protoobject.HeadResponse_Body_Header{
								Header: &protoobject.HeaderWithSignature{
									Header:    hdr,
									Signature: proto.Clone(validMinProtoSignature).(*protorefs.Signature),
								},
							}},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid header response: invalid header: "+tc.msg)
							},
						})
					}
					for _, tc := range invalidSignatureProtoTestcases {
						sig := proto.Clone(validFullProtoSignature).(*protorefs.Signature)
						tc.corrupt(sig)
						tcs = append(tcs, testcase{
							name: "header oneof/signature/" + tc.name,
							body: &protoobject.HeadResponse_Body{Head: &protoobject.HeadResponse_Body_Header{
								Header: &protoobject.HeaderWithSignature{
									Header:    proto.Clone(validMinObjectHeader).(*protoobject.Header),
									Signature: sig,
								},
							}},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid header response: invalid header: "+tc.msg)
							},
						})
					}
					for _, tc := range invalidObjectSplitInfoProtoTestcases {
						si := proto.Clone(validFullSplitInfo).(*protoobject.SplitInfo)
						tc.corrupt(si)
						tcs = append(tcs, testcase{
							name: "split info oneof/split info/" + tc.name,
							body: &protoobject.HeadResponse_Body{Head: &protoobject.HeadResponse_Body_SplitInfo{
								SplitInfo: si,
							}},
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid split info: "+tc.msg)
							},
						})
					}

					testInvalidResponseBodies(t, newTestHeadObjectServer, newTestObjectClient, tcs, func(c *Client) error {
						_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ObjectHead(ctx, anyCID, anyOID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestHeadObjectServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, err := newClient(t).ObjectHead(ctx, anyCID, anyOID, usertest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestHeadObjectServer, newTestObjectClient, func(c *Client) error {
			_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/654")
		testUnaryResponseCallback(t, newTestHeadObjectServer, newDefaultObjectService, func(c *Client) error {
			_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestHeadObjectServer, newDefaultObjectService, stat.MethodObjectHead,
			[]testedClientOp{func(c *Client) error {
				_, err := c.ObjectHead(ctx, anyCID, anyOID, nil, anyValidOpts)
				return err
			}}, nil, func(c *Client) error {
				_, err := c.ObjectHead(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_ObjectGetInit(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectGet
	anyCID := cidtest.ID()
	anyOID := oidtest.ID()
	anyValidSigner := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestGetObjectServer()
				c := newTestObjectClient(t, srv)

				srv.checkRequestObjectAddress(anyCID, anyOID)
				srv.authenticateRequest(anyValidSigner)
				_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, PrmObjectGet{})
				require.NoError(t, err)
				_, err = io.Copy(io.Discard, r)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestGetObjectServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, opts)
						return err
					})
				})
				t.Run("local", func(t *testing.T) {
					srv := newTestGetObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkLocal()

					srv.checkRequestLocal()
					_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
				t.Run("raw", func(t *testing.T) {
					srv := newTestGetObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkRaw()

					srv.checkRequestRaw()
					_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestGetObjectServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestGetObjectServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					bt.SetEACLTable(anyValidEACL) // TODO: drop after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					t.Run("split info", func(t *testing.T) {
						for _, tc := range []struct {
							name string
							body *protoobject.GetResponse_Body
						}{
							{name: "min", body: validMinObjectSplitInfoGetResponseBody},
							{name: "full", body: validFullObjectSplitInfoGetResponseBody},
						} {
							t.Run(tc.name, func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								srv.respondWithBody(0, tc.body)
								_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								var e *object.SplitInfoError
								require.ErrorAs(t, err, &e)
								require.NoError(t, checkSplitInfoTransport(*e.SplitInfo(), tc.body.GetSplitInfo()))
							})
						}
					})
					t.Run("header", func(t *testing.T) {
						const bigChunkSize = 4<<20 - object.MaxHeaderLen
						bigChunkTwice := make([]byte, 2*bigChunkSize)
						//nolint:staticcheck // OK for this test
						rand.Read(bigChunkTwice)
						type bodies = struct {
							heading *protoobject.GetResponse_Body
							chunks  [][]byte
						}
						type testcase = struct {
							name string
							bodies
							assert func(testing.TB, bodies, object.Object, io.Reader, error)
						}
						var tcs []testcase
						assertObject := func(t testing.TB, bs bodies, hdr object.Object, r io.Reader, err error) {
							checkSuccessfulGetObjectTransport(t, bs.heading, join(bs.chunks), hdr, r, err)
						}
						for _, tc := range []struct {
							name    string
							heading *protoobject.GetResponse_Body
						}{
							{name: "min", heading: validMinHeadingObjectGetResponseBody},
							{name: "full", heading: validFullHeadingObjectGetResponseBody},
						} {
							tcs = append(tcs,
								testcase{
									name: tc.name + " without payload", bodies: bodies{heading: tc.heading},
									assert: assertObject,
								},
								testcase{
									name: tc.name + " with single payload chunk", bodies: bodies{
										heading: tc.heading,
										chunks:  [][]byte{[]byte("Hello, world!")},
									}, assert: assertObject,
								},
								testcase{name: tc.name + " with multiple payload chunks", bodies: bodies{
									heading: tc.heading,
									chunks:  [][]byte{bigChunkTwice[:bigChunkSize], []byte("small"), {}, bigChunkTwice[bigChunkSize:]},
								}, assert: assertObject},
							)
						}
						for _, tc := range tcs {
							t.Run(tc.name, func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								h := srv.respondWithObject(proto.Clone(tc.heading.GetInit()).(*protoobject.GetResponse_Body_Init), tc.chunks)
								hdr, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								tc.assert(t, bodies{heading: h, chunks: tc.chunks}, hdr, r, err)
							})
						}
					})
				})
				t.Run("statuses", func(t *testing.T) {
					t.Run("no payload", func(t *testing.T) {
						t.Run("header", func(t *testing.T) {
							t.Run("OK", func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								hb := srv.respondWithObject(validFullHeadingObjectGetResponseBody.GetInit(), nil)
								hdr, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								checkSuccessfulGetObjectTransport(t, hb, nil, hdr, r, err)
							})
							t.Run("not OK", func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								var code uint32
								for code == 0 {
									code = rand.Uint32()
								}

								srv.respondWithObject(validFullHeadingObjectGetResponseBody.GetInit(), nil)
								srv.respondWithStatus(0, &protostatus.Status{Code: code})
								//nolint:staticcheck // drop with t.Skip()
								_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								t.Skip("currently ignores header and returns status error")
								require.EqualError(t, err, fmt.Sprintf("split info response returned with non-OK status code = %d", code))
							})
						})
						t.Run("split info", func(t *testing.T) {
							t.Run("OK", func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								body := validFullObjectSplitInfoGetResponseBody
								srv.respondWithBody(0, body)
								_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								var e *object.SplitInfoError
								require.ErrorAs(t, err, &e)
								require.NoError(t, checkSplitInfoTransport(*e.SplitInfo(), body.GetSplitInfo()))
							})
							t.Run("not OK", func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								var code uint32
								for code == 0 {
									code = rand.Uint32()
								}

								srv.respondWithBody(0, validFullObjectSplitInfoGetResponseBody)
								srv.respondWithStatus(0, &protostatus.Status{Code: code})
								//nolint:staticcheck // drop with t.Skip()
								_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								t.Skip("currently ignores split info and returns status error")
								require.EqualError(t, err, fmt.Sprintf("split info response returned with non-OK status code = %d", code))
							})
						})
					})
					t.Run("with payload", func(t *testing.T) {
						test := func(t testing.TB, code uint32,
							assert func(testing.TB, *protoobject.GetResponse_Body, []byte, object.Object, io.Reader, error)) {
							srv := newTestGetObjectServer()
							c := newTestObjectClient(t, srv)

							chunks := [][]byte{[]byte("one"), []byte("two"), []byte("three")}
							payload := join(chunks)

							hb := srv.respondWithObject(validFullHeadingObjectGetResponseBody.GetInit(), chunks)
							srv.respondWithStatus(uint(len(chunks)), &protostatus.Status{Code: code})
							hdr, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
							assert(t, hb, payload, hdr, r, err)
						}
						t.Run("OK", func(t *testing.T) {
							test(t, 0, func(t testing.TB, hb *protoobject.GetResponse_Body, payload []byte, hdr object.Object, r io.Reader, err error) {
								checkSuccessfulGetObjectTransport(t, hb, payload, hdr, r, err)
							})
						})
						t.Run("failure", func(t *testing.T) {
							test := func(t testing.TB, code uint32, assert func(t testing.TB, err error)) {
								test(t, code, func(t testing.TB, _ *protoobject.GetResponse_Body, _ []byte, _ object.Object, r io.Reader, err error) {
									require.NoError(t, err)
									_, err = io.ReadAll(r)
									assert(t, err)
								})
							}
							t.Run("internal server error", func(t *testing.T) {
								test(t, 1024, func(t testing.TB, err error) {
									require.ErrorAs(t, err, new(*apistatus.ServerInternal))
								})
							})
							t.Run("any other failure", func(t *testing.T) {
								var code uint32
								for code == 0 || code == 1024 {
									code = rand.Uint32()
								}
								test(t, code, func(t testing.TB, err error) {
									t.Skip("client sees no problem and just returns status")
									require.EqualError(t, err, fmt.Sprintf("unexpected status code = %d while reading payload", code))
								})
							})
						})
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "Get", func(c *Client) error {
						_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					t.Run("heading message", func(t *testing.T) {
						srv := newTestGetObjectServer()
						srv.respondWithoutSigning(0)
						c := newTestObjectClient(t, srv)
						_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						require.ErrorContains(t, err, "invalid response signature")
					})
					t.Run("payload chunk message", func(t *testing.T) {
						srv := newTestGetObjectServer()
						c := newTestObjectClient(t, srv)

						const n = 10
						chunks := make([][]byte, n)
						for i := range chunks {
							chunks[i] = []byte(fmt.Sprintf("chunk#%d", i))
						}

						srv.respondWithObject(validFullHeadingObjectGetResponseBody.GetInit(), chunks)
						srv.respondWithoutSigning(n) // remember that 1st message is heading
						_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
						require.NoError(t, err)
						read, err := io.ReadAll(r)
						require.ErrorContains(t, err, "invalid response signature")
						require.Equal(t, join(chunks[:n-1]), read)
					})
				})
				t.Run("payloads", func(t *testing.T) {
					t.Run("split info", func(t *testing.T) {
						for _, tc := range invalidObjectSplitInfoProtoTestcases {
							t.Run("split info/"+tc.name, func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								b := proto.Clone(validFullObjectSplitInfoGetResponseBody).(*protoobject.GetResponse_Body)
								tc.corrupt(b.GetSplitInfo())

								srv.respondWithBody(0, b)
								_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								require.EqualError(t, err, "header: invalid split info: "+tc.msg)
							})
						}
					})
					t.Run("heading", func(t *testing.T) {
						type testcase = struct {
							name, msg string
							corrupt   func(valid *protoobject.GetResponse_Body)
						}
						tcs := []testcase{
							{name: "nil", msg: "missing object ID field in the response", corrupt: func(valid *protoobject.GetResponse_Body) {
								valid.ObjectPart.(*protoobject.GetResponse_Body_Init_).Init = nil
							}},
							{name: "nil oneof", msg: "missing object ID field in the response", corrupt: func(valid *protoobject.GetResponse_Body) {
								valid.ObjectPart = &protoobject.GetResponse_Body_Init_{}
							}},
						}
						type initTescase = struct {
							name, msg string
							corrupt   func(valid *protoobject.GetResponse_Body_Init)
						}
						itcs := []initTescase{
							{name: "object ID/missing", msg: "missing object ID field in the response", corrupt: func(valid *protoobject.GetResponse_Body_Init) {
								valid.ObjectId = nil
							}},
							{name: "signature/missing", msg: "missing signature field in the response", corrupt: func(valid *protoobject.GetResponse_Body_Init) {
								valid.Signature = nil
							}},
							{name: "header/missing", msg: "missing header field in the response", corrupt: func(valid *protoobject.GetResponse_Body_Init) {
								valid.Header = nil
							}},
						}
						for _, tc := range invalidObjectIDProtoTestcases {
							itcs = append(itcs, initTescase{
								name: "object ID/" + tc.name, msg: "invalid ID: " + tc.msg,
								corrupt: func(valid *protoobject.GetResponse_Body_Init) { tc.corrupt(valid.ObjectId) },
							})
						}
						for _, tc := range invalidSignatureProtoTestcases {
							itcs = append(itcs, initTescase{
								name: "signature/" + tc.name, msg: "invalid signature: " + tc.msg,
								corrupt: func(valid *protoobject.GetResponse_Body_Init) { tc.corrupt(valid.Signature) },
							})
						}
						for _, tc := range invalidObjectHeaderProtoTestcases {
							itcs = append(itcs, initTescase{
								name: "header/" + tc.name, msg: "invalid header: " + tc.msg,
								corrupt: func(valid *protoobject.GetResponse_Body_Init) { tc.corrupt(valid.Header) },
							})
						}

						for _, tc := range itcs {
							tcs = append(tcs, testcase{
								name: tc.name, msg: tc.msg,
								corrupt: func(valid *protoobject.GetResponse_Body) {
									tc.corrupt(valid.ObjectPart.(*protoobject.GetResponse_Body_Init_).Init)
								},
							})
						}

						for _, tc := range tcs {
							t.Run(tc.name, func(t *testing.T) {
								srv := newTestGetObjectServer()
								c := newTestObjectClient(t, srv)

								b := proto.Clone(validFullHeadingObjectGetResponseBody).(*protoobject.GetResponse_Body)
								tc.corrupt(b)

								srv.respondWithBody(0, b)
								_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
								require.EqualError(t, err, "header: "+tc.msg)
							})
						}
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestGetObjectServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, _, err := newClient(t).ObjectGetInit(ctx, anyCID, anyOID, usertest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		t.Run("on stream init", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			transportErr := errors.New("any transport failure")

			srv.setHandlerError(transportErr)
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			assertObjectStreamTransportErr(t, transportErr, err)
		})
		t.Run("after heading response", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			transportErr := errors.New("any transport failure")

			srv.abortHandlerAfterResponse(1, transportErr)
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = r.Read([]byte{1})
			assertObjectStreamTransportErr(t, transportErr, err)
		})
		t.Run("on payload transmission", func(t *testing.T) {
			for _, n := range []uint{0, 2, 10} {
				t.Run(fmt.Sprintf("after %d successes", n), func(t *testing.T) {
					srv := newTestGetObjectServer()
					c := newTestObjectClient(t, srv)

					chunk := []byte("Hello, world!")
					transportErr := errors.New("any transport failure")

					srv.respondWithChunk(chunk)
					srv.abortHandlerAfterResponse(1+n, transportErr)
					_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
					require.NoError(t, err)
					for range n * uint(len(chunk)) {
						_, err = r.Read([]byte{1})
						require.NoError(t, err)
					}
					_, err = r.Read([]byte{1})
					assertObjectStreamTransportErr(t, transportErr, err)
				})
			}
		})
		t.Run("too large chunk message", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			cb := setChunkInGetResponse(validFullChunkObjectGetResponseBody, make([]byte, 4194305))

			srv.respondWithBody(1, cb)
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			st, ok := status.FromError(err)
			require.True(t, ok, err)
			require.Equal(t, codes.ResourceExhausted, st.Code())
			require.Contains(t, st.Message(), "grpc: received message larger than max (")
			require.Contains(t, st.Message(), " vs. 4194304)")
		})
	})
	t.Run("invalid message sequence", func(t *testing.T) {
		t.Run("no messages", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			srv.setHandlerError(nil)
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			_, ok := status.FromError(err)
			require.False(t, ok)
			require.EqualError(t, err, "header: %!w(<nil>)")
		})
		t.Run("chunk message first", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			srv.respondWithBody(0, proto.Clone(validFullChunkObjectGetResponseBody).(*protoobject.GetResponse_Body))
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.EqualError(t, err, "header: unexpected message instead of heading part: *object.GetObjectPartChunk")
		})
		t.Run("repeated heading message", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			srv.respondWithBody(2, proto.Clone(validMinHeadingObjectGetResponseBody).(*protoobject.GetResponse_Body))
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			require.EqualError(t, err, "unexpected message instead of chunk part: *object.GetObjectPartInit")
		})
		t.Run("non-first split info message", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			srv.respondWithBody(2, validMinObjectSplitInfoGetResponseBody)
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			require.EqualError(t, err, "unexpected message instead of chunk part: *object.SplitInfo")
		})
		t.Run("chunk after split info", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			srv.respondWithBody(1, validMinObjectSplitInfoGetResponseBody)
			srv.respondWithBody(2, validFullChunkObjectGetResponseBody)
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			//nolint:staticcheck // drop with t.Skip()
			_, err = io.Copy(io.Discard, r)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/659")
			require.EqualError(t, err, "unexpected message after split info response")
		})
		t.Run("cut payload", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			chunk := []byte("Hello, world!")
			hb := setPayloadLengthInHeadingGetResponse(validFullHeadingObjectGetResponseBody, uint64(len(chunk)+1))
			cb := setChunkInGetResponse(validFullChunkObjectGetResponseBody, chunk)

			srv.respondWithBody(0, hb)
			srv.respondWithBody(1, cb)
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})
		t.Run("payload size overflow", func(t *testing.T) {
			srv := newTestGetObjectServer()
			c := newTestObjectClient(t, srv)

			chunk := []byte("Hello, world!")
			hb := setPayloadLengthInHeadingGetResponse(validFullHeadingObjectGetResponseBody, uint64(len(chunk)-1))
			cb := setChunkInGetResponse(validFullChunkObjectGetResponseBody, chunk)

			srv.respondWithBody(0, hb)
			srv.respondWithBody(1, cb)
			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			//nolint:staticcheck // drop with t.Skip()
			_, err = io.Copy(io.Discard, r)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/658")
			require.EqualError(t, err, "payload size overflow")
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
		bind := func() (*testGetObjectServer, *Client, *[]collectedItem) {
			srv := newTestGetObjectServer()
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
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
			assertCommon(cl)
			collected := *cl
			require.Len(t, *cl, 1)
			require.Equal(t, stat.MethodObjectGet, collected[0].mtd)
			require.NoError(t, collected[0].err)
		})
		t.Run("sign request", func(t *testing.T) {
			_, c, cl := bind()
			_, _, err := c.ObjectGetInit(ctx, anyCID, anyOID, usertest.FailSigner(anyValidSigner), anyValidOpts)
			assertSignRequestErr(t, err)
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 1)
			require.Equal(t, stat.MethodObjectGet, collected[0].mtd)
			require.Equal(t, err, collected[0].err)
		})
		t.Run("transport failure", func(t *testing.T) {
			srv, c, cl := bind()
			transportErr := errors.New("any transport failure")
			srv.abortHandlerAfterResponse(3, transportErr)

			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			for err == nil {
				_, err = r.Read([]byte{1})
			}
			assertObjectStreamTransportErr(t, transportErr, err)
			assertCommon(cl)
			collected := *cl
			require.Equal(t, stat.MethodObjectGet, collected[0].mtd)
			require.NoError(t, collected[0].err)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/656")
			require.Len(t, collected, 2) // move upper
			require.Equal(t, stat.MethodObjectGetStream, collected[1].mtd)
			require.Equal(t, err, collected[1].err)
		})
		t.Run("OK", func(t *testing.T) {
			srv, c, cl := bind()
			const sleepDur = 100 * time.Millisecond
			// duration is pretty short overall, but most likely larger than the exec time w/o sleep
			srv.setSleepDuration(sleepDur)

			_, r, err := c.ObjectGetInit(ctx, anyCID, anyOID, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			require.NoError(t, err)
			assertCommon(cl)
			collected := *cl
			require.Equal(t, stat.MethodObjectGet, collected[0].mtd)
			require.NoError(t, collected[0].err)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/656")
			require.Len(t, collected, 2) // move upper
			require.Equal(t, stat.MethodObjectGetStream, collected[1].mtd)
			require.NoError(t, collected[1].err)
			require.Greater(t, collected[1].dur, sleepDur)
		})
	})
}

func TestClient_ObjectRangeInit(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectRange
	anyCID := cidtest.ID()
	anyOID := oidtest.ID()
	anyValidOff, anyValidLn := uint64(1), uint64(2)
	anyValidSigner := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestObjectPayloadRangeServer()
				c := newTestObjectClient(t, srv)

				srv.checkRequestObjectAddress(anyCID, anyOID)
				srv.checkRequestRange(anyValidOff, anyValidLn)
				srv.authenticateRequest(anyValidSigner)
				r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, PrmObjectRange{})
				require.NoError(t, err)
				_, err = io.Copy(io.Discard, r)
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestObjectPayloadRangeServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, opts)
						if err == nil {
							_, err = io.Copy(io.Discard, r)
						}
						return err
					})
				})
				t.Run("local", func(t *testing.T) {
					srv := newTestObjectPayloadRangeServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkLocal()

					srv.checkRequestLocal()
					r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
				t.Run("raw", func(t *testing.T) {
					srv := newTestObjectPayloadRangeServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkRaw()

					srv.checkRequestRaw()
					r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestObjectPayloadRangeServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newTestObjectPayloadRangeServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					bt.SetEACLTable(anyValidEACL) // TODO: drop after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = io.Copy(io.Discard, r)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					t.Run("split info", func(t *testing.T) {
						for _, tc := range []struct {
							name string
							body *protoobject.GetRangeResponse_Body
						}{
							{name: "min", body: validMinObjectSplitInfoRangeResponseBody},
							{name: "full", body: validFullObjectSplitInfoRangeResponseBody},
						} {
							t.Run(tc.name, func(t *testing.T) {
								srv := newTestObjectPayloadRangeServer()
								c := newTestObjectClient(t, srv)

								srv.respondWithBody(0, tc.body)
								r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
								require.NoError(t, err)
								_, err = r.Read([]byte{1})
								var e *object.SplitInfoError
								require.ErrorAs(t, err, &e)
								require.NoError(t, checkSplitInfoTransport(*e.SplitInfo(), tc.body.GetSplitInfo()))
							})
						}
					})
					t.Run("header", func(t *testing.T) {
						const bigChunkSize = 4<<20 - object.MaxHeaderLen
						bigChunkTwice := make([]byte, 2*bigChunkSize)
						//nolint:staticcheck // OK for this test
						rand.Read(bigChunkTwice)
						for _, tc := range []struct {
							name   string
							chunks [][]byte
						}{
							{name: "with single payload chunk", chunks: [][]byte{[]byte("Hello, world!")}},
							{name: "with multiple payload chunks",
								chunks: [][]byte{bigChunkTwice[:bigChunkSize], []byte("small"), {}, bigChunkTwice[bigChunkSize:]}},
						} {
							t.Run(tc.name, func(t *testing.T) {
								srv := newTestObjectPayloadRangeServer()
								c := newTestObjectClient(t, srv)

								payload := join(tc.chunks)

								srv.respondWithChunks(tc.chunks)
								r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, uint64(len(payload)), anyValidSigner, anyValidOpts)
								require.NoError(t, err)
								require.NoError(t, iotest.TestReader(r, payload))
							})
						}
					})
				})
				t.Run("statuses", func(t *testing.T) {
					t.Run("split info", func(t *testing.T) {
						t.Run("OK", func(t *testing.T) {
							srv := newTestObjectPayloadRangeServer()
							c := newTestObjectClient(t, srv)

							body := validFullObjectSplitInfoRangeResponseBody
							srv.respondWithBody(0, body)
							r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							_, err = r.Read([]byte{1})
							var e *object.SplitInfoError
							require.ErrorAs(t, err, &e)
							require.NoError(t, checkSplitInfoTransport(*e.SplitInfo(), body.GetSplitInfo()))
						})
						t.Run("not OK", func(t *testing.T) {
							srv := newTestObjectPayloadRangeServer()
							c := newTestObjectClient(t, srv)

							var code uint32
							for code == 0 {
								code = rand.Uint32()
							}

							srv.respondWithBody(0, validFullObjectSplitInfoRangeResponseBody)
							srv.respondWithStatus(0, &protostatus.Status{Code: code})
							r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							//nolint:staticcheck // drop with t.Skip()
							_, err = r.Read([]byte{1})
							t.Skip("currently ignores split info and returns status error")
							require.EqualError(t, err, fmt.Sprintf("split info response returned with non-OK status code = %d", code))
						})
					})
					t.Run("payload", func(t *testing.T) {
						test := func(t testing.TB, code uint32,
							assert func(testing.TB, []byte, io.Reader, error)) {
							srv := newTestObjectPayloadRangeServer()
							c := newTestObjectClient(t, srv)

							chunks := [][]byte{[]byte("one"), []byte("two"), []byte("three")}
							payload := join(chunks)

							srv.respondWithChunks(chunks)
							srv.respondWithStatus(uint(len(chunks))-1, &protostatus.Status{Code: code})
							r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, uint64(len(payload)), anyValidSigner, anyValidOpts)
							assert(t, payload, r, err)
						}
						t.Run("OK", func(t *testing.T) {
							test(t, 0, func(t testing.TB, payload []byte, r io.Reader, err error) {
								require.NoError(t, err)
								require.NoError(t, iotest.TestReader(r, payload))
							})
						})
						t.Run("failure", func(t *testing.T) {
							test := func(t testing.TB, code uint32, assert func(t testing.TB, err error)) {
								test(t, code, func(t testing.TB, _ []byte, r io.Reader, err error) {
									require.NoError(t, err)
									_, err = io.ReadAll(r)
									assert(t, err)
								})
							}
							t.Run("internal server error", func(t *testing.T) {
								test(t, 1024, func(t testing.TB, err error) {
									require.ErrorAs(t, err, new(*apistatus.ServerInternal))
								})
							})
							t.Run("any other failure", func(t *testing.T) {
								var code uint32
								for code == 0 || code == 1024 {
									code = rand.Uint32()
								}
								test(t, code, func(t testing.TB, err error) {
									t.Skip("client sees no problem and just returns status")
									require.EqualError(t, err, fmt.Sprintf("unexpected status code = %d while reading payload", code))
								})
							})
						})
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "GetRange", func(c *Client) error {
						r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
						for err == nil {
							_, err = r.Read([]byte{1})
						}
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					srv := newTestObjectPayloadRangeServer()
					c := newTestObjectClient(t, srv)

					const n = 10
					chunks := make([][]byte, n)
					for i := range chunks {
						chunks[i] = []byte(fmt.Sprintf("chunk#%d", i))
					}

					srv.respondWithChunks(chunks)
					srv.respondWithoutSigning(n - 1)
					r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, uint64(n*len(chunks)), anyValidSigner, anyValidOpts)
					require.NoError(t, err)
					read, err := io.ReadAll(r)
					require.ErrorContains(t, err, "invalid response signature")
					require.Equal(t, join(chunks[:n-1]), read)
				})
				t.Run("payloads", func(t *testing.T) {
					t.Run("split info", func(t *testing.T) {
						type testcase = struct {
							name, msg string
							splitInfo *protoobject.SplitInfo
						}
						tcs := []testcase{{
							name: "missing",
							msg:  "invalid split info: neither link object ID nor last part object ID is set",
							// nil becomes a zero-pointer after transport
							splitInfo: nil,
						}}
						for _, tc := range invalidObjectSplitInfoProtoTestcases {
							si := proto.Clone(validFullSplitInfo).(*protoobject.SplitInfo)
							tc.corrupt(si)
							tcs = append(tcs, testcase{name: tc.name, msg: "invalid split info: " + tc.msg, splitInfo: si})
						}
						for _, tc := range tcs {
							t.Run("split info/"+tc.name, func(t *testing.T) {
								srv := newTestObjectPayloadRangeServer()
								c := newTestObjectClient(t, srv)

								b := proto.Clone(validFullObjectSplitInfoRangeResponseBody).(*protoobject.GetRangeResponse_Body)
								b.RangePart.(*protoobject.GetRangeResponse_Body_SplitInfo).SplitInfo = tc.splitInfo

								srv.respondWithBody(0, b)
								r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
								require.NoError(t, err)
								_, err = r.Read([]byte{1})
								require.EqualError(t, err, tc.msg)
							})
						}
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
		t.Run("zero length", func(t *testing.T) {
			_, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, 0, anyValidSigner, anyValidOpts)
			require.ErrorIs(t, err, ErrZeroRangeLength)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestObjectPayloadRangeServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		_, err := newClient(t).ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, usertest.FailSigner(anyValidSigner), anyValidOpts)
		assertSignRequestErr(t, err)
	})
	t.Run("transport failure", func(t *testing.T) {
		t.Run("on stream init", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			transportErr := errors.New("any transport failure")

			srv.setHandlerError(transportErr)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = r.Read([]byte{1})
			assertObjectStreamTransportErr(t, transportErr, err)
		})
		t.Run("on payload transmission", func(t *testing.T) {
			for _, n := range []uint{0, 2, 10} {
				t.Run(fmt.Sprintf("after %d successes", n), func(t *testing.T) {
					srv := newTestObjectPayloadRangeServer()
					c := newTestObjectClient(t, srv)

					chunk := []byte("Hello, world!")
					transportErr := errors.New("any transport failure")

					srv.respondWithChunk(chunk)
					srv.abortHandlerAfterResponse(n, transportErr)
					r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, uint64(len(chunk))*uint64(n+1), anyValidSigner, anyValidOpts)
					require.NoError(t, err)
					for range n * uint(len(chunk)) {
						_, err = r.Read([]byte{1})
						require.NoError(t, err)
					}
					_, err = r.Read([]byte{1})
					assertObjectStreamTransportErr(t, transportErr, err)
				})
			}
		})
		t.Run("too large chunk message", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			cb := setChunkInRangeResponse(validFullChunkObjectRangeResponseBody, make([]byte, 4194305))

			srv.respondWithBody(1, cb)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			st, ok := status.FromError(err)
			require.True(t, ok, err)
			require.Equal(t, codes.ResourceExhausted, st.Code())
			require.Contains(t, st.Message(), "grpc: received message larger than max (")
			require.Contains(t, st.Message(), " vs. 4194304)")
		})
	})
	t.Run("invalid message sequence", func(t *testing.T) {
		t.Run("no messages", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			srv.setHandlerError(nil)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = r.Read([]byte{1})
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})
		t.Run("non-first split info message", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			srv.respondWithBody(1, validMinObjectSplitInfoRangeResponseBody)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			//nolint:staticcheck // drop with t.Skip()
			_, err = io.Copy(io.Discard, r)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/659")
			require.EqualError(t, err, "unexpected message instead of chunk part: *object.SplitInfo")
		})
		t.Run("chunk after split info", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			srv.respondWithBody(0, validMinObjectSplitInfoRangeResponseBody)
			srv.respondWithBody(1, validFullChunkObjectRangeResponseBody)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			//nolint:staticcheck // drop with t.Skip()
			_, err = io.Copy(io.Discard, r)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/659")
			require.EqualError(t, err, "unexpected message after split info response")
		})
		t.Run("cut payload", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			chunk := []byte("Hello, world!")
			cb := setChunkInRangeResponse(validFullChunkObjectRangeResponseBody, chunk)

			srv.respondWithBody(0, cb)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, uint64(len(chunk))+1, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})
		t.Run("payload size overflow", func(t *testing.T) {
			srv := newTestObjectPayloadRangeServer()
			c := newTestObjectClient(t, srv)

			chunk := []byte("Hello, world!")
			cb := setChunkInRangeResponse(validFullChunkObjectRangeResponseBody, chunk)

			srv.respondWithBody(0, cb)
			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, uint64(len(chunk))-1, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			//nolint:staticcheck // drop with t.Skip()
			_, err = io.Copy(io.Discard, r)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/658")
			require.EqualError(t, err, "payload size overflow")
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
		bind := func() (*testGetObjectPayloadRangeServer, *Client, *[]collectedItem) {
			srv := newTestObjectPayloadRangeServer()
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
		t.Run("zero range length", func(t *testing.T) {
			_, c, cl := bind()
			_, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, 0, anyValidSigner, anyValidOpts)
			require.ErrorIs(t, err, ErrZeroRangeLength)
			assertCommon(cl)
			collected := *cl
			require.Len(t, *cl, 1)
			require.Equal(t, stat.MethodObjectRange, collected[0].mtd)
			require.Equal(t, err, collected[0].err)
		})
		t.Run("missing signer", func(t *testing.T) {
			_, c, cl := bind()
			_, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
			assertCommon(cl)
			collected := *cl
			require.Len(t, *cl, 1)
			require.Equal(t, stat.MethodObjectRange, collected[0].mtd)
			require.NoError(t, collected[0].err)
		})
		t.Run("sign request", func(t *testing.T) {
			_, c, cl := bind()
			_, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, usertest.FailSigner(anyValidSigner), anyValidOpts)
			assertSignRequestErr(t, err)
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 1)
			require.Equal(t, stat.MethodObjectRange, collected[0].mtd)
			require.Equal(t, err, collected[0].err)
		})
		t.Run("transport failure", func(t *testing.T) {
			srv, c, cl := bind()
			transportErr := errors.New("any transport failure")
			srv.abortHandlerAfterResponse(2, transportErr)

			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, math.MaxInt, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			for err == nil {
				_, err = r.Read([]byte{1})
			}
			assertObjectStreamTransportErr(t, transportErr, err)
			assertCommon(cl)
			collected := *cl
			require.Equal(t, stat.MethodObjectRange, collected[0].mtd)
			require.NoError(t, collected[0].err)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/656")
			require.Len(t, collected, 2) // move upper
			require.Equal(t, stat.MethodObjectRangeStream, collected[1].mtd)
			require.Equal(t, err, collected[1].err)
		})
		t.Run("OK", func(t *testing.T) {
			srv, c, cl := bind()
			const sleepDur = 100 * time.Millisecond
			// duration is pretty short overall, but most likely larger than the exec time w/o sleep
			srv.setSleepDuration(sleepDur)

			r, err := c.ObjectRangeInit(ctx, anyCID, anyOID, anyValidOff, anyValidLn, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, r)
			require.NoError(t, err)
			assertCommon(cl)
			collected := *cl
			require.Equal(t, stat.MethodObjectRange, collected[0].mtd)
			require.NoError(t, collected[0].err)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/656")
			require.Len(t, collected, 2) // move upper
			require.Equal(t, stat.MethodObjectRangeStream, collected[1].mtd)
			require.NoError(t, collected[1].err)
			require.Greater(t, collected[1].dur, sleepDur)
		})
	})
}
