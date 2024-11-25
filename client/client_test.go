package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	apigrpc "github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
File contains common functionality used for client package testing.
*/

var statusErr apistatus.ServerInternal

func init() {
	statusErr.SetMessage("test status error")
}

func newInvalidRequestErr(cause error) error {
	return fmt.Errorf("invalid request: %w", cause)
}

func newInvalidRequestMetaHeaderErr(cause error) error {
	return newInvalidRequestErr(fmt.Errorf("invalid meta header: %w", cause))
}

func newInvalidRequestVerificationHeaderErr(cause error) error {
	return newInvalidRequestErr(fmt.Errorf("invalid verification header: %w", cause))
}

func newInvalidRequestBodyErr(cause error) error {
	return newInvalidRequestErr(fmt.Errorf("invalid body: %w", cause))
}

func newErrMissingRequestBodyField(name string) error {
	return newInvalidRequestBodyErr(fmt.Errorf("missing %s field", name))
}

func newErrInvalidRequestField(name string, err error) error {
	return newInvalidRequestBodyErr(fmt.Errorf("invalid %s field: %w", name, err))
}

// pairs service spec and implementation to-be-registered in some [grpc.Server].
type testService struct {
	desc *grpc.ServiceDesc
	impl any
}

// the most generic alternative of newClient. Both endpoint and parameter setter
// are optional.
func newCustomClient(t testing.TB, endpoint string, setPrm func(*PrmInit), svcs ...testService) *Client {
	var prm PrmInit
	if setPrm != nil {
		setPrm(&prm)
	}

	c, err := New(prm)
	require.NoError(t, err)

	srv := grpc.NewServer()
	for _, svc := range svcs {
		srv.RegisterService(svc.desc, svc.impl)
	}

	lis := bufconn.Listen(10 << 10)
	go func() { _ = srv.Serve(lis) }()

	var dialPrm PrmDial
	if endpoint == "" {
		endpoint = "grpc://localhost:8080"
	}
	dialPrm.SetServerURI(endpoint) // any valid
	dialPrm.setDialFunc(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) })
	err = c.Dial(dialPrm)
	if err != nil {
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Unimplemented, st.Code())
	}

	return c
}

// extends newClient with response meta info callback.
func newClientWithResponseCallback(t testing.TB, cb func(ResponseMetaInfo) error, svcs ...testService) *Client {
	return newCustomClient(t, "", func(prm *PrmInit) { prm.SetResponseInfoCallback(cb) }, svcs...)
}

// returns ready-to-go [Client] of provided optional services. By default, any
// other service is unsupported.
//
// If caller registers stat callback (like [PrmInit.SetStatisticCallback] does)
// processing nodeKey, it must include NetmapService with implemented
// LocalNodeInfo method.
func newClient(t testing.TB, svcs ...testService) *Client {
	return newCustomClient(t, "", nil, svcs...)
}

func TestClient_Dial(t *testing.T) {
	var prmInit PrmInit

	c, err := New(prmInit)
	require.NoError(t, err)

	t.Run("failure", func(t *testing.T) {
		t.Run("endpoint", func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				s      string
				assert func(t testing.TB, err error)
			}{
				{name: "missing", s: "", assert: func(t testing.TB, err error) {
					require.ErrorIs(t, c.Dial(PrmDial{}), ErrMissingServer)
				}},
				{name: "contains control char", s: "grpc://st1.storage.fs.neo.org:8080" + string(rune(0x7f)), assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "net/url: invalid control character in URL")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "missing port", s: "grpc://st1.storage.fs.neo.org", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "invalid port", s: "grpc://st1.storage.fs.neo.org:foo", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, `invalid port ":foo" after host`)
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "unsupported scheme", s: "unknown://st1.storage.fs.neo.org:8080", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "unsupported scheme: unknown")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "multiaddr", s: "/ip4/st1.storage.fs.neo.org/tcp/8080", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "host only", s: "st1.storage.fs.neo.org", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "invalid port without scheme", s: "st1.storage.fs.neo.org:foo", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var p PrmDial
					p.SetServerURI(tc.s)
					tc.assert(t, c.Dial(p))
				})
			}
		})
		t.Run("dial timeout", func(t *testing.T) {
			var p PrmDial
			p.SetServerURI("grpc://localhost:8080")
			p.SetTimeout(0)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
			p.SetTimeout(-1)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
		})
		t.Run("stream timeout", func(t *testing.T) {
			var p PrmDial
			p.SetServerURI("grpc://localhost:8080")
			p.SetStreamTimeout(0)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
			p.SetStreamTimeout(-1)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
		})
		t.Run("context", func(t *testing.T) {
			var anyValidPrm PrmDial
			anyValidPrm.SetServerURI("localhost:8080")
			t.Run("cancelled", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				p := anyValidPrm
				p.SetContext(ctx)
				err := c.Dial(p)
				require.ErrorIs(t, err, context.Canceled)
			})
			t.Run("deadline", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 0)
				cancel()

				p := anyValidPrm
				p.SetContext(ctx)
				err := c.Dial(p)
				require.ErrorIs(t, err, context.DeadlineExceeded)
			})
		})
	})
}

type nopPublicKey struct{}

func (x nopPublicKey) MaxEncodedSize() int     { return 10 }
func (x nopPublicKey) Encode(buf []byte) int   { return copy(buf, "public_key") }
func (x nopPublicKey) Decode([]byte) error     { return nil }
func (x nopPublicKey) Verify(_, _ []byte) bool { return true }

type nopSigner struct{}

func (nopSigner) Scheme() neofscrypto.Scheme      { return neofscrypto.ECDSA_SHA512 }
func (nopSigner) Sign([]byte) ([]byte, error)     { return []byte("signature"), nil }
func (x nopSigner) Public() neofscrypto.PublicKey { return nopPublicKey{} }

// provides generic server code for various NeoFS API RPC servers.
type testCommonServerSettings[
	REQUEST interface {
		GetMetaHeader() *protosession.RequestMetaHeader
	},
	REQUESTV2 any,
	REQUESTV2PTR interface {
		*REQUESTV2
		FromGRPCMessage(apigrpc.Message) error
	},
	RESPBODY any,
	RESP any,
	RESPV2 any,
	RESPV2PTR interface {
		*RESPV2
		ToGRPCMessage() apigrpc.Message
		FromGRPCMessage(apigrpc.Message) error
	},
] struct {
	handlerErr error

	reqXHdrs []string

	respSleep      time.Duration
	respUnsigned   bool
	respSigner     neofscrypto.Signer
	respMeta       *protosession.ResponseMetaHeader
	respBody       *RESPBODY
	respBodyForced bool // if respBody = nil is explicitly set
}

// makes the server to return given error as a gRPC status from the handler. By
// default, and if nil, some response message is returned.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) setHandlerError(err error) {
	x.handlerErr = err
}

// makes the server to assert that any request has given X-headers. By default,
// and if empty, no headers are expected.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) checkRequestXHeaders(xhdrs []string) {
	if len(xhdrs)%2 != 0 {
		panic("odd number of elements")
	}
	x.reqXHdrs = xhdrs
}

// makes the server to sleep specified time before any request processing. By
// default, and if non-positive, request is handled instantly.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) setSleepDuration(dur time.Duration) {
	x.respSleep = dur
}

// tells the server whether to sign all the responses or not. By default, any
// response is signed.
//
// Overrides signResponsesBy.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) respondWithoutSigning() {
	x.respUnsigned = true
}

// makes the server to always sign responses using given signer. By default, and
// if nil, random signer is used.
//
// No-op if signing is disabled using respondWithoutSigning.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) signResponsesBy(signer neofscrypto.Signer) {
	x.respSigner = signer
}

// makes the server to always respond with the given meta header. By default,
// and if nil, no header is attached.
//
// Overrides respondWithStatus.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) respondWithMeta(meta *protosession.ResponseMetaHeader) {
	x.respMeta = meta
}

// makes the server to always respond with the given status. By default, status
// OK is returned.
//
// Overrides respondWithMeta.
func (x *testCommonServerSettings[_, _, _, _, _, _, _]) respondWithStatus(st *protostatus.Status) {
	x.respondWithMeta(&protosession.ResponseMetaHeader{Status: st})
}

// makes the server to always respond with the given body. By default, any valid
// body is returned.
func (x *testCommonServerSettings[_, _, _, RESPBODY, _, _, _]) respondWithBody(body *RESPBODY) {
	x.respBody = body
	x.respBodyForced = true
}

func (x testCommonServerSettings[REQUEST, REQUESTV2, REQUESTV2PTR, _, _, _, _]) verifyRequest(req REQUEST) error {
	time.Sleep(x.respSleep)

	// signatures
	var reqV2 REQUESTV2
	if err := REQUESTV2PTR(&reqV2).FromGRPCMessage(req); err != nil {
		panic(err)
	}
	if err := verifyServiceMessage(&reqV2); err != nil {
		return newInvalidRequestVerificationHeaderErr(err)
	}
	// meta header
	metaHdr := req.GetMetaHeader()
	curVersion := version.Current()
	switch {
	case metaHdr == nil:
		return newInvalidRequestErr(errors.New("missing meta header"))
	case metaHdr.Version == nil:
		return newInvalidRequestMetaHeaderErr(errors.New("missing protocol version"))
	case metaHdr.Version.Major != curVersion.Major() || metaHdr.Version.Minor != curVersion.Minor():
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong protocol version v%d.%d, expected %s",
			metaHdr.Version.Major, metaHdr.Version.Minor, curVersion))
	case metaHdr.Epoch != 0:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("non-zero epoch #%d", metaHdr.Epoch))
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Epoch))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	case metaHdr.MagicNumber != 0:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("non-zero network magic #%d", metaHdr.MagicNumber))
	case metaHdr.Origin != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("origin header is presented while should not be"))
	case len(metaHdr.XHeaders) != len(x.reqXHdrs)/2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("number of x-headers %d differs parameterized %d",
			len(metaHdr.XHeaders), len(x.reqXHdrs)/2))
	}
	for i := range metaHdr.XHeaders {
		if metaHdr.XHeaders[i].Key != x.reqXHdrs[2*i] {
			return newInvalidRequestMetaHeaderErr(fmt.Errorf("x-header #%d key %q does not equal parameterized %q",
				i, metaHdr.XHeaders[i].Key, x.reqXHdrs[2*i]))
		}
		if metaHdr.XHeaders[i].Value != x.reqXHdrs[2*i+1] {
			return newInvalidRequestMetaHeaderErr(fmt.Errorf("x-header #%d value %q does not equal parameterized %q",
				i, metaHdr.XHeaders[i].Value, x.reqXHdrs[2*i+1]))
		}
	}
	return x.handlerErr
}

func (x testCommonServerSettings[_, _, _, _, RESP, RESPV2, RESPV2PTR]) signResponse(resp *RESP) (*RESP, error) {
	if x.respUnsigned {
		return resp, nil
	}
	var r RESPV2
	respV2 := RESPV2PTR(&r)
	if err := respV2.FromGRPCMessage(resp); err != nil {
		panic(err)
	}
	signer := x.respSigner
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	if err := signServiceMessage(signer, respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}
	return respV2.ToGRPCMessage().(*RESP), nil
}

// func signature shortener.
type testedClientOp = func(*Client) error

// asserts that built test server expecting particular X-headers receives them
// from the connected [Client] through on specified op execution. The op must be
// executed with all the correct parameters to return no error.
func testRequestXHeaders[SRV interface {
	checkRequestXHeaders([]string)
}](
	t *testing.T,
	newSrv func() SRV,
	connect func(testing.TB, any /* SRV */) *Client,
	op func(*Client, []string) error,
) {
	xhdrs := []string{
		"x-key1", "x-val1",
		"x-key2", "x-val2",
	}

	srv := newSrv()
	srv.checkRequestXHeaders(xhdrs)
	c := connect(t, srv)

	err := op(c, xhdrs)
	require.NoError(t, err)
}

func assertSignRequestErr(t testing.TB, err error) { require.ErrorContains(t, err, "sign request") }

// asserts that given op returns an error when the [Client]'s underlying signer
// fails to sign the request. The op must be executed with all the correct
// parameters.
func testSignRequestFailure(t testing.TB, op testedClientOp) {
	c := newClient(t)
	c.prm.signer = neofscryptotest.FailSigner(neofscryptotest.Signer())
	assertSignRequestErr(t, op(c))
}

func assertTransportErr(t testing.TB, transport, err error) {
	require.ErrorContains(t, err, "rpc failure")
	require.ErrorContains(t, err, "write request")
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unknown, st.Code())
	require.Contains(t, st.Message(), transport.Error())
}

// asserts that given [Client] op returns an expected error when built test
// server always responds with gRPC status error. The op must be executed with
// all the correct parameters.
func testTransportFailure[SRV interface {
	setHandlerError(error)
}](
	t testing.TB,
	newSrv func() SRV,
	connect func(t testing.TB, srv any) *Client,
	op testedClientOp,
) {
	transportErr := errors.New("any transport failure")
	srv := newSrv()
	srv.setHandlerError(transportErr)
	c := connect(t, srv)

	err := op(c)
	// note: errors returned from gRPC handlers are gRPC statuses, therefore,
	// strictly speaking, they are not transport errors (like connection refusal for
	// example). At the same time, according to the NeoFS protocol, all its statuses
	// are transmitted in the message. So, returning an error from gRPC handler
	// instead of a status field in the response is a protocol violation and can be
	// equated to a transport error.
	assertTransportErr(t, transportErr, err)
}

// asserts that given [Client] op returns an expected error when built test
// server responds with incorrect verification header. The op must be executed
// with all the correct parameters.
func testInvalidResponseSignatures[SRV interface {
	respondWithoutSigning()
}](
	t testing.TB,
	newSrv func() SRV,
	connect func(t testing.TB, srv any) *Client,
	op testedClientOp,
) {
	srv := newSrv()
	srv.respondWithoutSigning()
	// TODO: add cases with less radical corruption such as replacing one byte or
	//  dropping only one of the signatures
	c := connect(t, srv)
	require.ErrorContains(t, op(c), "invalid response signature")
}

type invalidResponseBodyTestcase[BODY any] struct {
	name      string
	body      *BODY
	assertErr func(testing.TB, error)
}

// asserts that given [Client] op returns expected errors when built test server
// responds with various invalid bodies. The op must be executed with all the
// correct parameters.
func testInvalidResponseBodies[BODY any, SRV interface {
	respondWithBody(*BODY)
}](
	t *testing.T,
	newSrv func() SRV,
	connect func(t testing.TB, srv any) *Client,
	tcs []invalidResponseBodyTestcase[BODY],
	op testedClientOp,
) {
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			srv := newSrv()
			srv.respondWithBody(tc.body)
			c := connect(t, srv)
			err := op(c)
			tc.assertErr(t, err)
		})
	}
}

// asserts that given [Client] op returns expected errors when server responds
// with various NeoFS statuses. The op must be executed with all the correct
// parameters.
func testStatusResponses[SRV interface {
	respondWithStatus(*protostatus.Status)
}](
	t *testing.T,
	newSrv func() SRV,
	connect func(t testing.TB, srv any) *Client,
	op testedClientOp,
) {
	execWithStatus := func(code uint32, msg string, details []*protostatus.Status_Detail) error {
		srv := newSrv()
		st := &protostatus.Status{Code: code, Message: msg, Details: details}
		srv.respondWithStatus(st)
		c := connect(t, srv)
		return op(c)
	}

	t.Run("OK", func(t *testing.T) {
		err := execWithStatus(0, "", make([]*protostatus.Status_Detail, 2))
		require.NoError(t, err)
	})
	t.Run("unrecognized", func(t *testing.T) {
		for _, code := range []uint32{
			1,
			1023,
			1028,
			2054,
			3074,
			4098,
		} {
			t.Run("unrecognized_"+strconv.FormatUint(uint64(code), 10), func(t *testing.T) {
				err := execWithStatus(code, "any message", make([]*protostatus.Status_Detail, 2))
				require.EqualError(t, err, "status: code = unrecognized message = any message")
				require.ErrorIs(t, err, apistatus.ErrUnrecognizedStatusV2)
				require.ErrorAs(t, err, new(*apistatus.UnrecognizedStatusV2))
			})
		}
	})

	type testcase struct {
		name          string
		code          uint32
		details       []*protostatus.Status_Detail
		defaultErrMsg string
		err, constErr error
		extraAssert   func(t testing.TB, msg string, err error)
	}
	tcs := []testcase{
		{name: "internal server error",
			// TODO: use const codes after transition to current module's proto lib
			code: 1024, details: make([]*protostatus.Status_Detail, 2),
			err: new(apistatus.ServerInternal), constErr: apistatus.ErrServerInternal,
			extraAssert: func(t testing.TB, msg string, err error) {
				var e *apistatus.ServerInternal
				require.ErrorAs(t, err, &e)
				require.Equal(t, msg, e.Message())
			},
		},
		{name: "invalid response signature",
			code: 1026, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "signature verification failed",
			err:           new(apistatus.SignatureVerification), constErr: apistatus.ErrSignatureVerification,
			extraAssert: func(t testing.TB, msg string, err error) {
				var e *apistatus.SignatureVerification
				require.ErrorAs(t, err, &e)
				require.Equal(t, msg, e.Message())
			},
		},
		{name: "node maintenance",
			code: 1027, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "node is under maintenance",
			err:           new(apistatus.NodeUnderMaintenance), constErr: apistatus.ErrNodeUnderMaintenance,
			extraAssert: func(t testing.TB, msg string, err error) {
				var e *apistatus.NodeUnderMaintenance
				require.ErrorAs(t, err, &e)
				require.Equal(t, msg, e.Message())
			},
		},
		{name: "missing object",
			code: 2049, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "object not found",
			err:           new(apistatus.ObjectNotFound), constErr: apistatus.ErrObjectNotFound,
		},
		{name: "locked object",
			code: 2050, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "object is locked",
			err:           new(apistatus.ObjectLocked), constErr: apistatus.ErrObjectLocked,
		},
		{name: "lock irregular object",
			code: 2051, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "locking non-regular object is forbidden",
			err:           new(apistatus.LockNonRegularObject), constErr: apistatus.ErrLockNonRegularObject,
		},
		{name: "already removed object",
			code: 2052, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "object already removed",
			err:           new(apistatus.ObjectAlreadyRemoved), constErr: apistatus.ErrObjectAlreadyRemoved,
		},
		{name: "out of object range",
			code: 2053, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "out of range",
			err:           new(apistatus.ObjectOutOfRange), constErr: apistatus.ErrObjectOutOfRange,
		},
		{name: "missing container",
			code: 3072, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "container not found",
			err:           new(apistatus.ContainerNotFound), constErr: apistatus.ErrContainerNotFound,
		},
		{name: "missing eACL",
			code: 3073, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "eACL not found",
			err:           new(apistatus.EACLNotFound), constErr: apistatus.ErrEACLNotFound,
		},
		{name: "missing session token",
			code: 4096, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "session token not found",
			err:           new(apistatus.SessionTokenNotFound), constErr: apistatus.ErrSessionTokenNotFound,
		},
		{name: "expired session token",
			code: 4097, details: make([]*protostatus.Status_Detail, 2),
			defaultErrMsg: "expired session token",
			err:           new(apistatus.SessionTokenExpired), constErr: apistatus.ErrSessionTokenExpired,
		},
	}
	for _, tc := range []struct {
		name              string
		correctMagicBytes []byte
		assert            func(testing.TB, *apistatus.WrongMagicNumber)
	}{
		{ // default
			assert: func(tb testing.TB, e *apistatus.WrongMagicNumber) {
				_, ok := e.CorrectMagic()
				require.Zero(t, ok)
			}},
		{name: "undersize",
			correctMagicBytes: make([]byte, 7),
			assert: func(tb testing.TB, e *apistatus.WrongMagicNumber) {
				_, ok := e.CorrectMagic()
				require.EqualValues(t, -1, ok)
			}},
		{name: "oversize",
			correctMagicBytes: make([]byte, 9),
			assert: func(tb testing.TB, e *apistatus.WrongMagicNumber) {
				_, ok := e.CorrectMagic()
				require.EqualValues(t, -1, ok)
			}},
		{name: "valid",
			correctMagicBytes: []byte{140, 15, 162, 245, 219, 236, 37, 191},
			assert: func(tb testing.TB, e *apistatus.WrongMagicNumber) {
				magic, ok := e.CorrectMagic()
				require.EqualValues(t, 1, ok)
				require.EqualValues(t, uint64(10092464466800944575), magic)
			}},
	} {
		name := "wrong magic number"
		var details []*protostatus.Status_Detail
		if tc.correctMagicBytes != nil {
			details = []*protostatus.Status_Detail{{Id: 0, Value: tc.correctMagicBytes}}
			name += "/with correct magic/" + tc.name
		} else {
			name += "/default"
		}
		tcs = append(tcs, testcase{name: name,
			code: 1025, details: details,
			err: new(apistatus.WrongMagicNumber), constErr: apistatus.ErrWrongMagicNumber,
			extraAssert: func(t testing.TB, _ string, err error) {
				var e *apistatus.WrongMagicNumber
				require.ErrorAs(t, err, &e)
				tc.assert(t, e)
			},
		})
	}
	for _, tc := range []struct {
		name   string
		reason string
		assert func(testing.TB, *apistatus.ObjectAccessDenied)
	}{
		{ // default
			assert: func(tb testing.TB, e *apistatus.ObjectAccessDenied) { require.Zero(t, e.Reason()) }},
		{name: "with reason",
			reason: "Hello, world!",
			assert: func(tb testing.TB, e *apistatus.ObjectAccessDenied) { require.Equal(t, "Hello, world!", e.Reason()) }},
	} {
		name := "object access denial"
		var details []*protostatus.Status_Detail
		if tc.reason != "" {
			details = []*protostatus.Status_Detail{{Id: 0, Value: []byte(tc.reason)}}
			name += "/with reason/" + tc.name
		} else {
			name += "/default"
		}
		tcs = append(tcs, testcase{name: name,
			code: 2048, details: details,
			err: new(apistatus.ObjectAccessDenied), constErr: apistatus.ErrObjectAccessDenied,
			defaultErrMsg: "access to object operation denied",
			extraAssert: func(t testing.TB, _ string, err error) {
				var e *apistatus.ObjectAccessDenied
				require.ErrorAs(t, err, &e)
				tc.assert(t, e)
			},
		})
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			checkWithMsg := func(msg string) {
				err := execWithStatus(tc.code, msg, tc.details)
				require.ErrorAs(t, err, &tc.err)
				require.ErrorIs(t, err, tc.constErr)
				var expectedErrMsg string
				if msg != "" {
					expectedErrMsg = fmt.Sprintf("status: code = %d message = %s", tc.code, msg)
				} else {
					if tc.defaultErrMsg != "" {
						expectedErrMsg = fmt.Sprintf("status: code = %d message = %s", tc.code, tc.defaultErrMsg)
					} else {
						expectedErrMsg = fmt.Sprintf("status: code = %d", tc.code)
					}
				}
				require.EqualError(t, err, expectedErrMsg)
				if tc.extraAssert != nil {
					tc.extraAssert(t, msg, tc.err)
				}
			}
			checkWithMsg("")
			checkWithMsg("Hello, world!")
		})
	}
}

// asserts that given [Client] op returns an expected error when some server
// responds with the incorrect message format. The op must be executed with all
// the correct parameters.
func testUnaryRPCResponseTypeMismatch(t testing.TB, svcName, method string, op testedClientOp) {
	svc := testService{
		desc: &grpc.ServiceDesc{ServiceName: "neo.fs.v2." + svcName, Methods: []grpc.MethodDesc{
			{
				MethodName: method,
				Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
					return timestamppb.Now(), nil // any completely different message
				},
			},
		}},
		impl: nil, // disables interface assert
	}
	c := newClient(t, svc)
	require.ErrorContains(t, op(c), "invalid response signature")
	// TODO: Although the client will not accept such a response, current error
	//  does not make it clear what exactly the problem is. It is worth reacting to
	//  the incorrect structure if possible.
}

// asserts that given [Client] op correctly reports meta information received
// from built test server when consuming the specified service. The op must be
// executed with all the correct parameters.
func testResponseCallback[SRV interface {
	respondWithMeta(*protosession.ResponseMetaHeader)
	signResponsesBy(neofscrypto.Signer)
}](
	t testing.TB,
	newSrv func() SRV,
	newSvc func(t testing.TB, srv any) testService,
	op testedClientOp,
) {
	// NetmapService.LocalNodeInfo is called on dial, so it should also be
	// initialized. The handler is called for it too.
	nodeInfoSrvSigner := neofscryptotest.Signer()
	nodeInfoSrvEpoch := rand.Uint64()
	nodeInfoSrv := newTestGetNodeInfoServer()
	nodeInfoSrv.respondWithMeta(&protosession.ResponseMetaHeader{Epoch: nodeInfoSrvEpoch})
	nodeInfoSrv.signResponsesBy(nodeInfoSrvSigner)

	srvSigner := neofscryptotest.Signer()
	srvEpoch := nodeInfoSrvEpoch + 1
	srv := newSrv()
	srv.respondWithMeta(&protosession.ResponseMetaHeader{Epoch: srvEpoch})
	srv.signResponsesBy(srvSigner)

	var collected []ResponseMetaInfo
	var cbErr error
	c := newClientWithResponseCallback(t, func(meta ResponseMetaInfo) error {
		collected = append(collected, meta)
		return cbErr
	},
		newDefaultNetmapServiceDesc(nodeInfoSrv),
		newSvc(t, srv),
	)

	err := op(c)
	require.NoError(t, err)
	require.Equal(t, []ResponseMetaInfo{
		{key: nodeInfoSrvSigner.PublicKeyBytes, epoch: nodeInfoSrvEpoch},
		{key: srvSigner.PublicKeyBytes, epoch: srvEpoch},
	}, collected)

	cbErr = errors.New("any response meta handler failure")
	err = op(c)
	require.ErrorContains(t, err, "response callback error")
	require.ErrorIs(t, err, cbErr)
	require.Len(t, collected, 3)
	require.Equal(t, collected[2], collected[1])
}

// checks that the [Client] correctly keeps exec statistics of specified ops
// performing communication with built test server. All operations must comply
// with the tested service.
func testStatistic[SRV interface {
	setSleepDuration(time.Duration)
	setHandlerError(error)
}](
	t testing.TB,
	newSrv func() SRV,
	newSvc func(t testing.TB, srv any) testService,
	mtd stat.Method,
	customNonStatFailures []testedClientOp,
	customStatFailures []testedClientOp,
	validInputCall testedClientOp,
) {
	// NetmapService.LocalNodeInfo is called on dial, so it should also be
	// initialized. Statistics are tracked for it too.
	nodeEndpoint := "grpc://localhost:8082" // any valid
	nodePub := []byte("any public key")

	nodeInfoSrv := newTestGetNodeInfoServer()
	nodeInfoSrv.respondWithNodePublicKey(nodePub)

	type statItem struct {
		mtd stat.Method
		dur time.Duration
		err error
	}
	var lastItem *statItem
	cb := func(pub []byte, endpoint string, mtd stat.Method, dur time.Duration, err error) {
		if lastItem == nil {
			require.Nil(t, pub)
		} else {
			require.Equal(t, nodePub, pub)
		}
		require.Equal(t, nodeEndpoint, endpoint)
		require.Positive(t, dur)
		lastItem = &statItem{mtd, dur, err}
	}

	srv := newSrv()
	c := newCustomClient(t, nodeEndpoint, func(prm *PrmInit) { prm.SetStatisticCallback(cb) },
		newDefaultNetmapServiceDesc(nodeInfoSrv),
		newSvc(t, srv),
	)
	// dial
	require.NotNil(t, lastItem)
	require.Equal(t, stat.MethodEndpointInfo, lastItem.mtd)
	require.Positive(t, lastItem.dur)
	require.NoError(t, lastItem.err)

	// custom non-stat failures
	for _, getNonStatErr := range customNonStatFailures {
		err := getNonStatErr(c)
		require.Error(t, err)
		require.Equal(t, mtd, lastItem.mtd)
		require.Positive(t, lastItem.dur)
		// TODO: strange that for some errors statistics are similar to OK
		require.NoError(t, lastItem.err)
	}

	// custom stat failures
	for _, getStatErr := range customStatFailures {
		err := getStatErr(c)
		require.Error(t, err)
		require.Equal(t, mtd, lastItem.mtd)
		require.Positive(t, lastItem.dur)
		require.Equal(t, err, lastItem.err)
	}

	// sign request failure
	signerCp := c.prm.signer
	c.prm.signer = neofscryptotest.FailSigner(c.prm.signer)

	err := validInputCall(c)
	assertSignRequestErr(t, err)
	require.Equal(t, mtd, lastItem.mtd)
	require.Positive(t, lastItem.dur)
	require.Equal(t, err, lastItem.err)

	c.prm.signer = signerCp

	// transport
	transportErr := errors.New("any transport failure")
	srv.setHandlerError(transportErr)

	err = validInputCall(c)
	assertTransportErr(t, transportErr, err)
	require.Equal(t, mtd, lastItem.mtd)
	require.Positive(t, lastItem.dur)
	require.Equal(t, err, lastItem.err)

	srv.setHandlerError(nil)

	// OK
	sleepDur := 100 * time.Millisecond
	// duration is pretty short overall, but most likely larger than the exec time w/o sleep
	srv.setSleepDuration(sleepDur)
	err = validInputCall(c)
	require.NoError(t, err)
	require.Equal(t, mtd, lastItem.mtd)
	require.Greater(t, lastItem.dur, sleepDur)
	require.NoError(t, lastItem.err)
}
