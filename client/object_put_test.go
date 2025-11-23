package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const signOneReqCalls = 3 // body+headers

type nFailedSigner struct {
	user.Signer
	n, count uint
}

// returns [user.Signer] failing all Sign calls starting from the n-th one.
func newNFailedSigner(base user.Signer, n uint) user.Signer {
	return &nFailedSigner{Signer: base, n: n}
}

func (x *nFailedSigner) Sign(data []byte) ([]byte, error) {
	x.count++
	if x.count < x.n {
		return x.Signer.Sign(data)
	}
	return nil, errors.New("test signer forcefully fails")
}

type testPutObjectServer struct {
	protoobject.UnimplementedObjectServiceServer
	testCommonClientStreamServerSettings[
		*protoobject.PutRequest_Body,
		*protoobject.PutRequest,
		*protoobject.PutResponse_Body,
		*protoobject.PutResponse,
	]
	testObjectSessionServerSettings
	testBearerTokenServerSettings
	testLocalRequestServerSettings

	reqHdr     *object.Object
	reqPayload []byte
	reqCopies  uint32

	reqPayloadLenCounter int
}

// returns [protoobject.ObjectServiceServer] supporting Put method only. Default
// implementation performs common verification of any request, and responds with
// any valid message. The message flow is also strictly controlled. Some methods
// allow to tune the behavior.
func newPutObjectServer() *testPutObjectServer { return new(testPutObjectServer) }

// makes the server to assert that any heading request caries given value in
// copy num field. By default, the field must be zero.
func (x *testPutObjectServer) checkRequestCopiesNumber(n uint32) { x.reqCopies = n }

// makes the server to assert that any heading request carries given object
// header. By default, any header is accepted.
func (x *testPutObjectServer) checkRequestHeader(hdr object.Object) { x.reqHdr = &hdr }

// makes the server to assert that any given data is streamed as an object
// payload. By default, and if nil, any payload is accepted.
func (x *testPutObjectServer) checkRequestPayload(data []byte) { x.reqPayload = data }

func (x *testPutObjectServer) verifyHeadingMessage(m *protoobject.PutRequest_Body_Init) error {
	if m.Header == nil {
		return errors.New("missing header field")
	}
	// 4. copies number
	if x.reqCopies != m.CopiesNumber {
		return fmt.Errorf("copies number field (client: %d, message: %d)", x.reqCopies, m.CopiesNumber)
	}
	if x.reqHdr == nil {
		return nil
	}
	// 1. ID
	id := x.reqHdr.GetID()
	mid := m.GetObjectId()
	if id.IsZero() {
		if mid != nil {
			return errors.New("object ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing object ID field")
		}
		if err := checkObjectIDTransport(id, mid); err != nil {
			return fmt.Errorf("object ID field: %w", err)
		}
	}
	// 2. signature
	// 3. header
	if err := checkObjectHeaderWithSignatureTransport(*x.reqHdr, &protoobject.HeaderWithSignature{
		Header: m.Header, Signature: m.Signature,
	}); err != nil {
		return fmt.Errorf("header with signature fields: %w", err)
	}
	return nil
}

func (x *testPutObjectServer) verifyPayloadChunkMessage(chunk []byte) error {
	ln := len(chunk)
	if ln == 0 {
		return errors.New("empty payload chunk")
	}
	const maxChunkLen = 3 << 20
	if ln > maxChunkLen {
		return fmt.Errorf("intermediate chunk exceeds the expected size limit: %dB > %dB", ln, maxChunkLen)
	}
	if x.reqPayload == nil {
		return nil
	}
	if exp := x.reqPayload[x.reqPayloadLenCounter:]; !bytes.HasPrefix(exp, chunk) {
		return fmt.Errorf("wrong payload chunk (remains: %dB, message: %dB)", len(exp), len(chunk))
	}
	x.reqPayloadLenCounter += ln
	return nil
}

func (x *testPutObjectServer) verifyRequest(req *protoobject.PutRequest) error {
	// TODO(https://github.com/nspcc-dev/neofs-sdk-go/issues/662): why meta is
	//  transmitted in all stream messages when heading parts is enough?
	if err := x.testCommonClientStreamServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	metaHdr := req.MetaHeader
	// TTL
	if err := x.verifyTTL(metaHdr); err != nil {
		return err
	}
	// session token
	if err := x.verifySessionToken(metaHdr.GetSessionToken()); err != nil {
		return err
	}
	if err := x.verifySessionTokenV2(metaHdr.GetSessionTokenV2()); err != nil {
		return err
	}
	// bearer token
	if err := x.verifyBearerToken(metaHdr.GetBearerToken()); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	switch v := body.ObjectPart.(type) {
	default:
		return newErrInvalidRequestField("object part", fmt.Errorf("unsupported oneof type %T", v))
	case nil:
		return newErrMissingRequestBodyField("object part")
	case *protoobject.PutRequest_Body_Init_:
		if x.reqCounter > 1 {
			return newErrInvalidRequestField("object part", fmt.Errorf("heading part must be a 1st stream message only, "+
				"but received in #%d one", x.reqCounter))
		}
		if v.Init == nil {
			panic("nil oneof field container")
		}
		if err := x.verifyHeadingMessage(v.Init); err != nil {
			return newErrInvalidRequestField("heading part", err)
		}
	case *protoobject.PutRequest_Body_Chunk:
		if x.reqCounter <= 1 {
			return newErrInvalidRequestField("object part", errors.New("payload chunk must not be a 1st stream message"))
		}
		if err := x.verifyPayloadChunkMessage(v.Chunk); err != nil {
			return newErrInvalidRequestField("chunk part", err)
		}
	}
	return nil
}

func (x *testPutObjectServer) sendResponse(stream protoobject.ObjectService_PutServer) error {
	resp := &protoobject.PutResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinPutObjectResponseBody).(*protoobject.PutResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return fmt.Errorf("sign response: %w", err)
	}

	return stream.SendAndClose(resp)
}

func (x *testPutObjectServer) reset() {
	x.reqCounter, x.reqPayloadLenCounter = 0, 0
}

func (x *testPutObjectServer) Put(stream protoobject.ObjectService_PutServer) error {
	defer x.reset()
	time.Sleep(x.handlerSleepDur)
	if x.handlerErrForced {
		return x.handlerErr
	}
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if x.reqCounter == 0 {
					return errors.New("stream finished without messages")
				}
				if x.reqPayload != nil && x.reqPayloadLenCounter != len(x.reqPayload) {
					return fmt.Errorf("unfinished payload (expected: %dB, received: %dB)", len(x.reqPayload), x.reqPayloadLenCounter)
				}
				break
			}
			return err
		}
		x.reqCounter++
		if err := x.verifyRequest(req); err != nil {
			return err
		}
		if x.reqErrN > 0 && x.reqCounter >= x.reqErrN {
			return x.reqErr
		}
		if x.respN > 0 && x.reqCounter >= x.respN {
			break
		}
	}
	return x.sendResponse(stream)
}

func TestClient_ObjectPut(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmObjectPutInit
	anyValidHdr := objecttest.Object()
	anyValidSigner := usertest.User()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				for _, tc := range []struct {
					name       string
					payloadLen uint
				}{
					{name: "no payload", payloadLen: 0},
					{name: "one byte", payloadLen: 1},
					{name: "3MB-1", payloadLen: 3<<20 - 1},
					{name: "3MB", payloadLen: 3 << 20},
					{name: "3MB+1", payloadLen: 3<<20 + 1},
					{name: "6MB-1", payloadLen: 6<<20 - 1},
					{name: "6MB", payloadLen: 6 << 20},
					{name: "6MB+1", payloadLen: 6<<20 + 1},
					{name: "10MB", payloadLen: 10 << 20},
				} {
					t.Run(tc.name, func(t *testing.T) {
						srv := newPutObjectServer()
						c := newTestObjectClient(t, srv)

						payload := testutil.RandByteSlice(tc.payloadLen)

						srv.checkRequestHeader(anyValidHdr)
						srv.checkRequestPayload(payload)
						srv.authenticateRequest(anyValidSigner)
						w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, PrmObjectPutInit{})
						require.NoError(t, err)

						chunkLen := len(payload)/10 + 1
						for len(payload) > 0 {
							ln := min(chunkLen, len(payload))
							n, err := w.Write(payload[:ln])
							require.NoError(t, err)
							require.EqualValues(t, ln, n)
							payload = payload[ln:]
						}
						require.NoError(t, w.Close())
					})
				}
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newPutObjectServer, newTestObjectClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, opts)
						if err == nil {
							_, err = w.Write([]byte{1})
							if err == nil {
								err = w.Close()
							}
						}
						return err
					})
				})
				t.Run("local", func(t *testing.T) {
					srv := newPutObjectServer()
					c := newTestObjectClient(t, srv)

					opts := anyValidOpts
					opts.MarkLocal()

					srv.checkRequestLocal()
					w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = w.Write([]byte{1})
					require.NoError(t, err)
					require.NoError(t, w.Close())
				})
				t.Run("session token", func(t *testing.T) {
					srv := newPutObjectServer()
					c := newTestObjectClient(t, srv)

					st := sessiontest.ObjectSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = w.Write([]byte{1})
					require.NoError(t, err)
					require.NoError(t, w.Close())
				})
				t.Run("bearer token", func(t *testing.T) {
					srv := newPutObjectServer()
					c := newTestObjectClient(t, srv)

					bt := bearertest.Token()
					require.NoError(t, bt.Sign(usertest.User()))
					opts := anyValidOpts
					opts.WithBearerToken(bt)

					srv.checkRequestBearerToken(bt)
					w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = w.Write([]byte{1})
					require.NoError(t, err)
					require.NoError(t, w.Close())
				})
				t.Run("copies number", func(t *testing.T) {
					srv := newPutObjectServer()
					c := newTestObjectClient(t, srv)

					n := rand.Uint32()
					opts := anyValidOpts
					opts.SetCopiesNumber(n)

					srv.checkRequestCopiesNumber(n)
					w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, opts)
					require.NoError(t, err)
					_, err = w.Write([]byte{1})
					require.NoError(t, err)
					require.NoError(t, w.Close())
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for statusName, status := range map[string]struct {
						status *protostatus.Status
						err    error
					}{
						"ok": {nil, nil},
						"incomplete": {&protostatus.Status{
							Code:    1,
							Message: "incomplete",
							Details: make([]*protostatus.Status_Detail, 2),
						}, apistatus.ErrIncomplete},
					} {
						t.Run(statusName, func(t *testing.T) {
							for _, tc := range []struct {
								name string
								body *protoobject.PutResponse_Body
							}{
								{name: "min", body: validMinPutObjectResponseBody},
								{name: "full", body: validFullPutObjectResponseBody},
							} {
								t.Run(tc.name, func(t *testing.T) {
									srv := newPutObjectServer()
									c := newTestObjectClient(t, srv)

									srv.respondWithBodyAndStatus(tc.body, status.status)
									w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
									require.NoError(t, err)
									_, err = w.Write([]byte{1})
									require.NoError(t, err)
									require.ErrorIs(t, w.Close(), status.err)
									require.NoError(t, checkObjectIDTransport(w.GetResult().StoredObjectID(), tc.body.GetObjectId()))
								})
							}
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						n    uint
					}{
						{name: "on stream init", n: 1},
						{name: "after heading request", n: 2},
						{name: "on payload transmission", n: 10},
					} {
						t.Run("interrupting/"+tc.name, func(t *testing.T) {
							test := func(ok bool) {
								t.Run(fmt.Sprintf("ok=%t", ok), func(t *testing.T) {
									srv := newPutObjectServer()
									c := newTestObjectClient(t, srv)

									var code uint32
									if !ok {
										for code == 0 {
											code = rand.Uint32()
										}
									}
									srv.respondAfterRequest(tc.n)
									srv.respondWithStatus(&protostatus.Status{Code: code})
									ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // prevent hanging
									t.Cleanup(cancel)
									w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
									require.NoError(t, err)
									for err == nil {
										_, err = w.Write([]byte{1})
										time.Sleep(50 * time.Millisecond) // give the response time to come
									}
									if ok {
										t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/649")
										require.EqualError(t, err, "server unexpectedly interrupted the stream with a response")
									} else {
										t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/648")
										require.ErrorIs(t, err, apistatus.Error)
									}
								})
							}
							test(true)
							test(false)
						})
					}
					t.Run("after stream finish", func(t *testing.T) {
						testStatusResponses(t, newPutObjectServer, newTestObjectClient, func(c *Client) error {
							w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
							if err == nil {
								_, err = w.Write([]byte{1})
								if err == nil {
									err = w.Close()
								}
							}
							return err
						})
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				exec := func(c *Client) error {
					w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
					for err == nil {
						return w.Close()
					}
					return err
				}
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "object.ObjectService", "Put", exec)
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newPutObjectServer, newTestObjectClient, exec)
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protoobject.PutResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil, assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, ErrMissingResponseField)
							require.EqualError(t, err, "missing ID field in the response")
						}},
						{name: "empty", body: new(protoobject.PutResponse_Body), assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, ErrMissingResponseField)
							require.EqualError(t, err, "missing ID field in the response")
						}},
					}
					for _, tc := range invalidObjectIDProtoTestcases {
						body := proto.Clone(validFullPutObjectResponseBody).(*protoobject.PutResponse_Body)
						tc.corrupt(body.ObjectId)
						tcs = append(tcs, testcase{name: "ID/" + tc.name, body: body, assertErr: func(t testing.TB, err error) {
							require.EqualError(t, err, "invalid ID field in the response: "+tc.msg)
						},
						})
					}

					testInvalidResponseBodies(t, newPutObjectServer, newTestObjectClient, tcs, exec)
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ObjectPutInit(ctx, anyValidHdr, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newPutObjectServer, newTestObjectClient, func(ctx context.Context, c *Client) error {
			_, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		t.Run("heading", func(t *testing.T) {
			_, err := newClient(t).ObjectPutInit(ctx, anyValidHdr, usertest.FailSigner(anyValidSigner), anyValidOpts)
			require.ErrorContains(t, err, "header write")
			require.ErrorContains(t, err, "sign message")
		})
		t.Run("payload chunks", func(t *testing.T) {
			for _, n := range []int{0, 1, 10} {
				t.Run(fmt.Sprintf("after %d successes", n), func(t *testing.T) {
					srv := newPutObjectServer()
					c := newTestObjectClient(t, srv)

					okSignings := signOneReqCalls * (n + 1) // +1 for header one
					signer := newNFailedSigner(anyValidSigner, uint(okSignings+1))
					w, err := c.ObjectPutInit(ctx, anyValidHdr, signer, anyValidOpts)
					require.NoError(t, err)

					for range n {
						_, err = w.Write([]byte{1})
						require.NoError(t, err)
					}
					_, err = w.Write([]byte{1})
					require.ErrorContains(t, err, "sign message")
				})
			}
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		test := func(t testing.TB, n uint, handleInit func(testing.TB, io.WriteCloser, error) error) {
			srv := newPutObjectServer()
			c := newTestObjectClient(t, srv)

			transportErr := errors.New("any transport failure")

			srv.abortHandlerAfterRequest(n, transportErr)
			w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
			if handleInit != nil {
				err = handleInit(t, w, err)
			}
			assertObjectStreamTransportErr(t, transportErr, err)
		}
		t.Run("on stream init", func(t *testing.T) {
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/649")
			test(t, 0, func(t testing.TB, w io.WriteCloser, err error) error {
				for err == nil {
					_, err = w.Write([]byte{1})
					time.Sleep(50 * time.Millisecond) // give the response time to come
				}
				require.ErrorContains(t, err, "header write")
				return err
			})
		})
		t.Run("after heading request", func(t *testing.T) {
			test := func(t testing.TB, withPayload bool) {
				test(t, 1, func(t testing.TB, w io.WriteCloser, err error) error {
					require.NoError(t, err)
					if withPayload {
						_, err = w.Write([]byte{1}) // gRPC client stream does not ACK each request
						if err == nil {
							// wait for the response
							err = w.Close()
						} // else it has already come and reflected in err
					} else {
						err = w.Close()
					}
					return err
				})
			}
			t.Run("with payload", func(t *testing.T) { test(t, true) })
			t.Run("without payload", func(t *testing.T) { test(t, false) })
		})
		t.Run("on payload transmission", func(t *testing.T) {
			for _, n := range []uint{0, 2, 10} {
				t.Run(fmt.Sprintf("after %d successes", n), func(t *testing.T) {
					test(t, 2+n, func(t testing.TB, w io.WriteCloser, err error) error {
						require.NoError(t, err)
						for range n {
							_, err = w.Write([]byte{1})
							require.NoError(t, err)
						}
						_, err = w.Write([]byte{1}) // gRPC client stream does not ACK each request
						if err == nil {
							// wait for the response
							err = w.Close()
						} // else it has already come and reflected in err
						return err
					})
				})
			}
		})
	})
	t.Run("no response message", func(t *testing.T) {
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/649")
		assertNoResponseErr := func(t testing.TB, err error) {
			_, ok := status.FromError(err)
			require.False(t, ok)
			require.EqualError(t, err, "server finished stream without response")
		}
		test := func(t testing.TB, n uint, assertStream func(testing.TB, io.WriteCloser, error)) {
			srv := newPutObjectServer()
			c := newTestObjectClient(t, srv)

			srv.abortHandlerAfterRequest(n, nil)
			w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
			assertStream(t, w, err)
		}

		t.Run("on stream init", func(t *testing.T) {
			test(t, 0, func(t testing.TB, _ io.WriteCloser, err error) { assertNoResponseErr(t, err) })
		})
		t.Run("after heading request", func(t *testing.T) {
			test := func(t testing.TB, withPayload bool) {
				test(t, 1, func(t testing.TB, w io.WriteCloser, err error) {
					require.NoError(t, err)
					if withPayload {
						_, err = w.Write([]byte{1}) // gRPC client stream does not ACK each request
						if err == nil {
							// wait for the response
							err = w.Close()
						} // else it has already come and reflected in err
					} else {
						err = w.Close()
					}
					assertNoResponseErr(t, err)
				})
			}
			t.Run("with payload", func(t *testing.T) { test(t, true) })
			t.Run("without payload", func(t *testing.T) { test(t, false) })
		})
		t.Run("on chunk requests", func(t *testing.T) {
			for _, n := range []uint{0, 2, 10} {
				t.Run(fmt.Sprintf("after %d successes", n), func(t *testing.T) {
					test(t, 2+n, func(t testing.TB, w io.WriteCloser, err error) {
						require.NoError(t, err)
						for range n {
							_, err = w.Write([]byte{1})
							require.NoError(t, err)
						}
						_, err = w.Write([]byte{1}) // gRPC client stream does not ACK each request
						if err == nil {
							// wait for the response
							err = w.Close()
						} // else it has already come and reflected in err
						assertNoResponseErr(t, err)
					})
				})
			}
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
		bind := func() (*testPutObjectServer, *Client, *[]collectedItem) {
			srv := newPutObjectServer()
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
			_, err := c.ObjectPutInit(ctx, anyValidHdr, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 1)
			require.Equal(t, stat.MethodObjectPut, collected[0].mtd)
			require.NoError(t, collected[0].err)
		})
		t.Run("sign heading request failure", func(t *testing.T) {
			_, c, cl := bind()
			_, err := c.ObjectPutInit(ctx, anyValidHdr, usertest.FailSigner(anyValidSigner), anyValidOpts)
			require.ErrorContains(t, err, "header write")
			require.ErrorContains(t, err, "sign message")
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 2)
			require.Equal(t, stat.MethodObjectPutStream, collected[0].mtd)
			require.ErrorContains(t, collected[0].err, "sign message")
			require.Equal(t, stat.MethodObjectPut, collected[1].mtd)
			require.Equal(t, err, collected[1].err)
		})
		t.Run("sign chunk request failure", func(t *testing.T) {
			_, c, cl := bind()
			w, err := c.ObjectPutInit(ctx, anyValidHdr, newNFailedSigner(anyValidSigner, signOneReqCalls*2+1), anyValidOpts)
			require.NoError(t, err)
			_, err = w.Write([]byte{1})
			require.NoError(t, err)
			_, err = w.Write([]byte{1})
			require.ErrorContains(t, err, "sign message")
			err = w.Close()
			require.ErrorContains(t, err, "sign message")
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 2)
			require.Equal(t, stat.MethodObjectPut, collected[0].mtd)
			require.NoError(t, collected[0].err)
			require.Equal(t, stat.MethodObjectPutStream, collected[1].mtd)
			require.Equal(t, err, collected[1].err)
		})
		t.Run("transport failure", func(t *testing.T) {
			srv, c, cl := bind()
			transportErr := errors.New("any transport failure")
			srv.abortHandlerAfterRequest(3, transportErr)

			w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			for err == nil {
				_, err = w.Write([]byte{1})
				time.Sleep(50 * time.Millisecond) // give the response time to come
			}
			err = w.Close()
			assertObjectStreamTransportErr(t, transportErr, err)
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 2)
			require.Equal(t, stat.MethodObjectPut, collected[0].mtd)
			require.NoError(t, collected[0].err)
			require.Equal(t, stat.MethodObjectPutStream, collected[1].mtd)
			require.Equal(t, err, collected[1].err)
		})
		t.Run("OK", func(t *testing.T) {
			srv, c, cl := bind()
			const sleepDur = 100 * time.Millisecond
			// duration is pretty short overall, but most likely larger than the exec time w/o sleep
			srv.setSleepDuration(sleepDur)

			w, err := c.ObjectPutInit(ctx, anyValidHdr, anyValidSigner, anyValidOpts)
			require.NoError(t, err)
			_, err = w.Write([]byte{1})
			require.NoError(t, err)
			err = w.Close()
			require.NoError(t, err)
			assertCommon(cl)
			collected := *cl
			require.Len(t, collected, 2)
			require.Equal(t, stat.MethodObjectPut, collected[0].mtd)
			require.NoError(t, collected[0].err)
			require.Equal(t, stat.MethodObjectPutStream, collected[1].mtd)
			require.NoError(t, err, collected[1].err)
			require.Greater(t, collected[1].dur, sleepDur)
		})
	})
}
