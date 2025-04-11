package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
File contains common functionality used for client package testing.
*/

var statusErr apistatus.ServerInternal

func init() {
	statusErr.SetMessage("test status error")
}

// flattens all slices into one.
// TODO: propose to [slices] package.
func join[SS ~[]S, S ~[]E, E any](ss SS) S {
	var res S
	for i := range ss {
		res = append(res, ss[i]...)
	}
	return res
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

// static server settings used for [Client] testing.
var (
	testServerEndpoint     = "localhost:8080"
	testServerSignerOnDial = neofscryptotest.Signer()
	testServerStateOnDial  = struct {
		pub   []byte
		epoch uint64
	}{
		pub:   neofscrypto.PublicKeyBytes(testServerSignerOnDial.Public()),
		epoch: rand.Uint64(),
	}
)

// pairs service spec and implementation to-be-registered in some [grpc.Server].
type testService struct {
	desc *grpc.ServiceDesc
	impl any
}

// the most generic alternative of newClient. Both endpoint and parameter setter
// are optional.
func newCustomClient(t testing.TB, setPrm func(*PrmInit), svcs ...testService) *Client {
	var prm PrmInit
	if setPrm != nil {
		setPrm(&prm)
	}

	c, err := New(prm)
	require.NoError(t, err)

	// serve dial RPC
	const netmapSvcName = "neo.fs.v2.netmap.NetmapService"
	const nodeInfoMtdName = "LocalNodeInfo"
	netmapSvcInd := -1
	nodeInfoMtdInd := -1
loop:
	for i := range svcs {
		if svcs[i].desc.ServiceName == netmapSvcName {
			netmapSvcInd = i
			for j := range svcs[i].desc.Methods {
				if svcs[i].desc.Methods[j].MethodName == nodeInfoMtdName {
					nodeInfoMtdInd = j
					break loop
				}
			}
		}
	}

	type nodeInfoServer interface {
		LocalNodeInfo(context.Context, *protonetmap.LocalNodeInfoRequest) (*protonetmap.LocalNodeInfoResponse, error)
	}
	dialSrv := newTestGetNodeInfoServer()
	dialSrv.signResponsesBy(testServerSignerOnDial.ECDSAPrivateKey)
	dialSrv.respondWithNodePublicKey(testServerStateOnDial.pub)
	dialSrv.respondWithMeta(&protosession.ResponseMetaHeader{Epoch: testServerStateOnDial.epoch})
	handleDial := func(_ any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
		var req protonetmap.LocalNodeInfoRequest
		if err := dec(&req); err != nil {
			return nil, err
		}
		return dialSrv.LocalNodeInfo(ctx, &req)
	}

	if netmapSvcInd < 0 {
		svcs = append(svcs, testService{
			desc: &grpc.ServiceDesc{
				ServiceName: netmapSvcName,
				HandlerType: (*nodeInfoServer)(nil),
				Methods:     []grpc.MethodDesc{{MethodName: nodeInfoMtdName, Handler: handleDial}},
			},
		})
	} else {
		dcp := *svcs[netmapSvcInd].desc // safe copy prevents mutation
		dcp.Methods = slices.Clone(dcp.Methods)
		if nodeInfoMtdInd < 0 {
			dcp.Methods = append(dcp.Methods, grpc.MethodDesc{MethodName: nodeInfoMtdName, Handler: handleDial})
		} else {
			originalHandler := dcp.Methods[nodeInfoMtdInd].Handler
			called := false
			dcp.Methods[nodeInfoMtdInd].Handler = func(srv any, ctx context.Context, dec func(any) error, in grpc.UnaryServerInterceptor) (any, error) {
				if !called {
					called = true
					return handleDial(srv, ctx, dec, in)
				}
				return originalHandler(srv, ctx, dec, in)
			}
		}
		svcs[netmapSvcInd].desc = &dcp
	}

	srv := grpc.NewServer()
	for _, svc := range svcs {
		srv.RegisterService(svc.desc, svc.impl)
	}

	lis := bufconn.Listen(10 << 10)
	go func() { _ = srv.Serve(lis) }()

	var dialPrm PrmDial
	dialPrm.SetServerURI(testServerEndpoint)
	dialPrm.setDialFunc(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) })
	err = c.Dial(dialPrm)
	require.NoError(t, err)

	return c
}

// extends newClient with response meta info callback.

// returns ready-to-go [Client] of provided optional services. By default, any
// other service is unsupported.
//
// Note: [Client] uses NetmapService.LocalNodeInfo RPC to dial the server. Test
// [Client] always receives testServerStateOnDial. Take this into account if the
// test keeps track of all ops like stat test.
func newClient(t testing.TB, svcs ...testService) *Client {
	return newCustomClient(t, nil, svcs...)
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

func TestClientClose(t *testing.T) {
	var prmInit PrmInit

	c, err := New(prmInit) // No Dial called.
	require.NoError(t, err)

	require.NoError(t, c.Close())

	c = newClient(t) // Dial called.

	require.NoError(t, c.Close())
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

// various cross-service protocol messages. Any message (incl. set elements)
// must be cloned via [proto.Clone] before passing anywhere.
var (
	// correct NeoFS protocol version with required fields only.
	validMinProtoVersion = &protorefs.Version{}
	// correct NeoFS protocol version with all fields.
	validFullProtoVersion = &protorefs.Version{Major: 538919038, Minor: 3957317479}
	// set of correct container IDs.
	validProtoContainerIDs = []*protorefs.ContainerID{
		{Value: []byte{198, 137, 143, 192, 231, 50, 106, 89, 225, 118, 7, 42, 40, 225, 197, 183, 9, 205, 71, 140, 233, 30, 63, 73, 224, 244, 235, 18, 205, 45, 155, 236}},
		{Value: []byte{26, 71, 99, 242, 146, 121, 0, 142, 95, 50, 78, 190, 222, 104, 252, 72, 48, 219, 67, 226, 30, 90, 103, 51, 1, 234, 136, 143, 200, 240, 75, 250}},
		{Value: []byte{51, 124, 45, 83, 227, 119, 66, 76, 220, 196, 118, 197, 116, 44, 138, 83, 103, 102, 134, 191, 108, 124, 162, 255, 184, 137, 193, 242, 178, 10, 23, 29}},
	}
)

var anyValidEACL = eacltest.Table()

type (
	invalidSessionTokenProtoTestcase = struct {
		name, msg string
		corrupt   func(*protosession.SessionToken)
	}
)

// various sets of cross-service testcases.
var (
	invalidUUIDProtoTestcases = []struct {
		name, msg string
		corrupt   func(valid []byte) []byte
	}{
		{name: "undersize", msg: "invalid UUID (got 15 bytes)", corrupt: func(valid []byte) []byte {
			return valid[:15]
		}},
		{name: "oversize", msg: "invalid UUID (got 17 bytes)", corrupt: func(valid []byte) []byte {
			return append(valid, 1)
		}},
		{name: "wrong version", msg: "wrong UUID version 3, expected 4", corrupt: func(valid []byte) []byte {
			valid[6] = 3 << 4
			return valid
		}},
	}
	invalidContainerIDProtoTestcases = []struct {
		name, msg string
		corrupt   func(valid *protorefs.ContainerID)
	}{
		{name: "nil", msg: "invalid length 0", corrupt: func(valid *protorefs.ContainerID) {
			valid.Value = nil
		}},
		{name: "empty", msg: "invalid length 0", corrupt: func(valid *protorefs.ContainerID) {
			valid.Value = []byte{}
		}},
		{name: "undersize", msg: "invalid length 31", corrupt: func(valid *protorefs.ContainerID) {
			valid.Value = valid.Value[:31]
		}},
		{name: "oversize", msg: "invalid length 33", corrupt: func(valid *protorefs.ContainerID) {
			valid.Value = append(valid.Value, 1)
		}},
		{name: "zero", msg: "zero container ID", corrupt: func(valid *protorefs.ContainerID) {
			for i := range valid.Value {
				valid.Value[i] = 0
			}
		}},
	}
	invalidUserIDProtoTestcases = []struct {
		name, msg string
		corrupt   func(valid *protorefs.OwnerID)
	}{
		{name: "nil", msg: "invalid length 0, expected 25", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value = nil
		}},
		{name: "empty", msg: "invalid length 0, expected 25", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value = []byte{}
		}},
		{name: "owner/undersize", msg: "invalid length 24, expected 25", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value = valid.Value[:24]
		}},
		{name: "owner/oversize", msg: "invalid length 26, expected 25", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value = append(valid.Value, 1)
		}},
		{name: "owner/wrong prefix", msg: "invalid prefix byte 0x42, expected 0x35", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value[0] = 0x42
			h := sha256.Sum256(valid.Value[:21])
			hh := sha256.Sum256(h[:])
			copy(valid.Value[21:], hh[:])
		}},
		{name: "owner/wrong checksum", msg: "checksum mismatch", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value[24]++
		}},
		{name: "owner/zero", msg: "invalid prefix byte 0x0, expected 0x35", corrupt: func(valid *protorefs.OwnerID) {
			valid.Value = make([]byte, 25)
		}},
	}
	invalidObjectIDProtoTestcases = []struct {
		name, msg string
		corrupt   func(valid *protorefs.ObjectID)
	}{
		{name: "nil", msg: "invalid length 0", corrupt: func(valid *protorefs.ObjectID) {
			valid.Value = nil
		}},
		{name: "empty", msg: "invalid length 0", corrupt: func(valid *protorefs.ObjectID) {
			valid.Value = []byte{}
		}},
		{name: "undersize", msg: "invalid length 31", corrupt: func(valid *protorefs.ObjectID) {
			valid.Value = valid.Value[:31]
		}},
		{name: "oversize", msg: "invalid length 33", corrupt: func(valid *protorefs.ObjectID) {
			valid.Value = append(valid.Value, 1)
		}},
		{name: "zero", msg: "zero object ID", corrupt: func(valid *protorefs.ObjectID) {
			for i := range valid.Value {
				valid.Value[i] = 0
			}
		}},
	}
	invalidChecksumTestcases = []struct {
		name, msg string
		corrupt   func(valid *protorefs.Checksum)
	}{
		{name: "negative scheme", msg: "negative type -1", corrupt: func(valid *protorefs.Checksum) {
			valid.Type = -1
		}},
		{name: "value/nil", msg: "missing value", corrupt: func(valid *protorefs.Checksum) {
			valid.Sum = nil
		}},
		{name: "value/empty", msg: "missing value", corrupt: func(valid *protorefs.Checksum) {
			valid.Sum = []byte{}
		}},
	}
	invalidSignatureProtoTestcases = []struct {
		name, msg string
		corrupt   func(valid *protorefs.Signature)
	}{
		{name: "negative scheme", msg: "negative scheme -1", corrupt: func(valid *protorefs.Signature) {
			valid.Scheme = -1
		}},
	}
	invalidCommonSessionTokenProtoTestcases = []invalidSessionTokenProtoTestcase{
		{name: "body/nil", msg: "missing token body", corrupt: func(valid *protosession.SessionToken) {
			valid.Body = nil
		}},
		{name: "body/ID/nil", msg: "missing session ID", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.Id = nil
		}},
		{name: "body/ID/empty", msg: "missing session ID", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.Id = []byte{}
		}},
		// + other ID cases in init
		{name: "body/issuer/nil", msg: "missing session issuer", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.OwnerId = nil
		}},
		// + other issuer cases in init
		{name: "body/lifetime", msg: "missing token lifetime", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.Lifetime = nil
		}},
		{name: "body/session key/nil", msg: "missing session public key", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.SessionKey = nil
		}},
		{name: "body/session key/empty", msg: "missing session public key", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.SessionKey = []byte{}
		}},
		{name: "body/context/nil", msg: "missing session context", corrupt: func(valid *protosession.SessionToken) {
			valid.Body.Context = nil
		}},
		{name: "signature/nil", msg: "missing body signature", corrupt: func(valid *protosession.SessionToken) {
			valid.Signature = nil
		}},
		// + other signature cases in init
	}
)

func init() {
	for _, tc := range invalidUUIDProtoTestcases {
		invalidCommonSessionTokenProtoTestcases = append(invalidCommonSessionTokenProtoTestcases, invalidSessionTokenProtoTestcase{
			name: "body/ID/" + tc.name, msg: "invalid session ID: " + tc.msg,
			corrupt: func(valid *protosession.SessionToken) { valid.Body.Id = tc.corrupt(valid.Body.Id) },
		})
	}
	for _, tc := range invalidUserIDProtoTestcases {
		invalidCommonSessionTokenProtoTestcases = append(invalidCommonSessionTokenProtoTestcases, invalidSessionTokenProtoTestcase{
			name: "body/issuer/" + tc.name, msg: "invalid session issuer: " + tc.msg,
			corrupt: func(valid *protosession.SessionToken) { tc.corrupt(valid.Body.OwnerId) },
		})
	}
	for _, tc := range invalidSignatureProtoTestcases {
		invalidCommonSessionTokenProtoTestcases = append(invalidCommonSessionTokenProtoTestcases, invalidSessionTokenProtoTestcase{
			name: "signature/" + tc.name, msg: "invalid body signature: " + tc.msg,
			corrupt: func(valid *protosession.SessionToken) { tc.corrupt(valid.Signature) },
		})
	}
}

// for sharing between servers of requests with required container ID.
type testRequiredContainerIDServerSettings struct {
	expectedReqCnrID *cid.ID
}

// makes the server to assert that any request carries given container ID. By
// default, any ID is accepted.
func (x *testRequiredContainerIDServerSettings) checkRequestContainerID(id cid.ID) {
	x.expectedReqCnrID = &id
}

func (x testRequiredContainerIDServerSettings) verifyRequestContainerID(m *protorefs.ContainerID) error {
	if m == nil {
		return newErrMissingRequestBodyField("container ID")
	}
	if x.expectedReqCnrID != nil {
		if err := checkContainerIDTransport(*x.expectedReqCnrID, m); err != nil {
			return newErrInvalidRequestField("container ID", err)
		}
	}
	return nil
}

// provides generic server code for various NeoFS API RPC servers.
type testCommonServerSettings struct {
	handlerSleepDur  time.Duration
	handlerErrForced bool
	handlerErr       error
}

// makes the server to return given error as a gRPC status from the handler. By
// default, and if nil, some response message is returned.
func (x *testCommonServerSettings) setHandlerError(err error) {
	x.handlerErrForced, x.handlerErr = true, err
}

// makes the server to sleep specified time before any request processing. By
// default, and if non-positive, request is handled instantly.
func (x *testCommonServerSettings) setSleepDuration(dur time.Duration) { x.handlerSleepDur = dur }

// provides generic server code for various NeoFS API unary RPC servers.
type testCommonUnaryServerSettings[
	REQBODY neofsproto.Message,
	REQ interface {
		GetBody() REQBODY
		GetMetaHeader() *protosession.RequestMetaHeader
		GetVerifyHeader() *protosession.RequestVerificationHeader
	},
	RESPBODY interface {
		proto.Message
		neofsproto.Message
	},
	RESP interface {
		GetBody() RESPBODY
		GetMetaHeader() *protosession.ResponseMetaHeader
	},
] struct {
	testCommonServerSettings
	testCommonRequestServerSettings[REQBODY, REQ]
	testCommonResponseServerSettings[RESPBODY, RESP]
}

// provides generic server code for various NeoFS API server-side stream RPC
// servers.
type testCommonServerStreamServerSettings[
	REQBODY neofsproto.Message,
	REQ interface {
		GetBody() REQBODY
		GetMetaHeader() *protosession.RequestMetaHeader
		GetVerifyHeader() *protosession.RequestVerificationHeader
	},
	RESPBODY interface {
		proto.Message
		neofsproto.Message
	},
	RESP interface {
		GetBody() RESPBODY
		GetMetaHeader() *protosession.ResponseMetaHeader
	},
] struct {
	testCommonServerSettings
	testCommonRequestServerSettings[REQBODY, REQ]
	resps    map[uint]testCommonResponseServerSettings[RESPBODY, RESP]
	respErrN uint
	respErr  error
}

// tunes processing of N-th response starting from 0.
func (x *testCommonServerStreamServerSettings[_, _, RESPBODY, RESP]) tuneNResp(n uint,
	tune func(*testCommonResponseServerSettings[RESPBODY, RESP])) {
	type t = testCommonResponseServerSettings[RESPBODY, RESP]
	if x.resps == nil {
		x.resps = make(map[uint]t, 1)
	}
	s := x.resps[n]
	tune(&s)
	x.resps[n] = s
}

// tells the server whether to sign the n-th response or not. By default, any
// response is signed.
//
// Overrides signResponsesBy.
func (x *testCommonServerStreamServerSettings[_, _, RESPBODY, RESP]) respondWithoutSigning(n uint) {
	x.tuneNResp(n, func(s *testCommonResponseServerSettings[RESPBODY, RESP]) {
		s.respondWithoutSigning()
	})
}

// makes the server to sign n-th response using given signer. By default, and
// if nil, random signer is used.
//
// No-op if signing is disabled using respondWithoutSigning.
// nolint:unused // will be needed for https://github.com/nspcc-dev/neofs-sdk-go/issues/653
func (x *testCommonServerStreamServerSettings[_, _, RESPBODY, RESP]) signResponsesBy(n uint, signer ecdsa.PrivateKey) {
	x.tuneNResp(n, func(s *testCommonResponseServerSettings[RESPBODY, RESP]) {
		s.signResponsesBy(signer)
	})
}

// makes the server to return n-th response with given meta header. By default,
// and if nil, no header is attached.
//
// Overrides respondWithStatus.
// nolint:unused // will be needed for https://github.com/nspcc-dev/neofs-sdk-go/issues/653
func (x *testCommonServerStreamServerSettings[_, _, RESPBODY, RESP]) respondWithMeta(n uint, meta *protosession.ResponseMetaHeader) {
	x.tuneNResp(n, func(s *testCommonResponseServerSettings[RESPBODY, RESP]) {
		s.respondWithMeta(meta)
	})
}

// makes the server to return given status in the n-th response. By default,
// status OK is returned.
//
// Overrides respondWithMeta.
func (x *testCommonServerStreamServerSettings[_, _, RESPBODY, RESP]) respondWithStatus(n uint, st *protostatus.Status) {
	x.tuneNResp(n, func(s *testCommonResponseServerSettings[RESPBODY, RESP]) {
		s.respondWithStatus(st)
	})
}

// makes the server to return n-th request with the given body. By default, any
// valid body is returned.
func (x *testCommonServerStreamServerSettings[_, _, RESPBODY, RESP]) respondWithBody(n uint, body RESPBODY) {
	x.tuneNResp(n, func(s *testCommonResponseServerSettings[RESPBODY, RESP]) {
		s.respondWithBody(body)
	})
}

// makes the server to return given error as a gRPC status from the handler
// after the n-th response transmission. If n is zero, handler returns
// immediately. By default, all responses are sent. Note that nil error is also
// returned since it leads to a particular gRPC status.
//
// Overrides respondWithStatus.
func (x *testCommonServerStreamServerSettings[_, _, _, _]) abortHandlerAfterResponse(n uint, err error) {
	if n == 0 {
		x.setHandlerError(err)
	} else {
		x.respErrN, x.respErr = n, err
	}
}

// provides generic server code for various NeoFS API client-side stream RPC
// servers.
type testCommonClientStreamServerSettings[
	REQBODY neofsproto.Message,
	REQ interface {
		GetBody() REQBODY
		GetMetaHeader() *protosession.RequestMetaHeader
		GetVerifyHeader() *protosession.RequestVerificationHeader
	},
	RESPBODY interface {
		proto.Message
		neofsproto.Message
	},
	RESP interface {
		GetBody() RESPBODY
		GetMetaHeader() *protosession.ResponseMetaHeader
	},
] struct {
	testCommonServerSettings
	testCommonRequestServerSettings[REQBODY, REQ]
	testCommonResponseServerSettings[RESPBODY, RESP]
	reqCounter uint
	reqErrN    uint
	reqErr     error
	respN      uint
}

// makes the server to return given error as a gRPC status from the handler
// after the n-th request receipt. If n is zero, handler returns immediately. By
// default, all requests are processed and response message is returned. Note
// that nil error is also returned since it leads to a particular gRPC status.
//
// Overrides respondWithStatusOnRequest.
func (x *testCommonClientStreamServerSettings[_, _, _, _]) abortHandlerAfterRequest(n uint, err error) {
	if n == 0 {
		x.setHandlerError(err)
	} else {
		x.reqErrN, x.reqErr = n, err
	}
}

// makes the server to immediately respond right after the n-th request
// received.
func (x *testCommonClientStreamServerSettings[_, _, _, _]) respondAfterRequest(n uint) {
	x.respN = n
}

type testCommonRequestServerSettings[
	REQBODY neofsproto.Message,
	REQ interface {
		GetBody() REQBODY
		GetMetaHeader() *protosession.RequestMetaHeader
		GetVerifyHeader() *protosession.RequestVerificationHeader
	},
] struct {
	reqCreds *authCredentials
	reqXHdrs []string
}

// makes the server to assert that any request has given X-headers. By default,
// and if empty, no headers are expected.
func (x *testCommonRequestServerSettings[_, _]) checkRequestXHeaders(xhdrs []string) {
	if len(xhdrs)%2 != 0 {
		panic("odd number of elements")
	}
	x.reqXHdrs = xhdrs
}

// makes the server to assert that any request is signed by s. By default, any
// signer is accepted.
//
// Has no effect with checkRequestDataSignature.
func (x *testCommonRequestServerSettings[_, _]) authenticateRequest(s neofscrypto.Signer) {
	c := authCredentialsFromSigner(s)
	x.reqCreds = &c
}

func (x testCommonRequestServerSettings[REQBODY, REQ]) verifyRequest(req REQ) error {
	body := req.GetBody()
	metaHdr := req.GetMetaHeader()
	verifyHdr := req.GetVerifyHeader()

	// signatures
	if verifyHdr == nil {
		return newInvalidRequestErr(errors.New("missing verification header"))
	}
	if verifyHdr.Origin != nil {
		return newInvalidRequestVerificationHeaderErr(errors.New("origin field is set while should not be"))
	}
	if err := verifyDataSignature(
		neofsproto.MarshalMessage(body), verifyHdr.BodySignature, x.reqCreds); err != nil {
		return newInvalidRequestVerificationHeaderErr(fmt.Errorf("body signature: %w", err))
	}
	if err := verifyDataSignature(
		neofsproto.MarshalMessage(metaHdr), verifyHdr.MetaSignature, x.reqCreds); err != nil {
		return newInvalidRequestVerificationHeaderErr(fmt.Errorf("meta signature: %w", err))
	}
	if err := verifyDataSignature(
		neofsproto.MarshalMessage(verifyHdr.Origin), verifyHdr.OriginSignature, x.reqCreds); err != nil {
		return newInvalidRequestVerificationHeaderErr(fmt.Errorf("verification header's origin signature: %w", err))
	}
	// meta header
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
	return nil
}

type testCommonResponseServerSettings[
	RESPBODY interface {
		proto.Message
		neofsproto.Message
	},
	RESP interface {
		GetBody() RESPBODY
		GetMetaHeader() *protosession.ResponseMetaHeader
	},
] struct {
	respUnsigned   bool
	respSigner     *ecdsa.PrivateKey
	respMeta       *protosession.ResponseMetaHeader
	respBody       RESPBODY
	respBodyForced bool // if respBody = nil is explicitly set
}

// tells the server whether to sign all the responses or not. By default, any
// response is signed.
//
// Overrides signResponsesBy.
func (x *testCommonResponseServerSettings[_, _]) respondWithoutSigning() {
	x.respUnsigned = true
}

// makes the server to always sign responses using given signer. By default, and
// if nil, random signer is used.
//
// No-op if signing is disabled using respondWithoutSigning.
func (x *testCommonResponseServerSettings[_, _]) signResponsesBy(key ecdsa.PrivateKey) {
	x.respSigner = &key
}

// makes the server to always respond with the given meta header. By default,
// and if nil, no header is attached.
//
// Overrides respondWithStatus.
func (x *testCommonResponseServerSettings[_, _]) respondWithMeta(meta *protosession.ResponseMetaHeader) {
	x.respMeta = meta
}

// makes the server to always respond with the given status. By default, status
// OK is returned.
//
// Overrides respondWithMeta.
func (x *testCommonResponseServerSettings[_, _]) respondWithStatus(st *protostatus.Status) {
	x.respondWithMeta(&protosession.ResponseMetaHeader{Status: st})
}

// makes the server to always respond with the given body. By default, any valid
// body is returned.
func (x *testCommonResponseServerSettings[RESPBODY, _]) respondWithBody(body RESPBODY) {
	x.respBody = proto.Clone(body).(RESPBODY)
	x.respBodyForced = true
}

func (x testCommonResponseServerSettings[_, RESP]) signResponse(resp RESP) (*protosession.ResponseVerificationHeader, error) {
	if x.respUnsigned {
		return nil, nil
	}
	var signer ecdsa.PrivateKey
	if x.respSigner != nil {
		signer = *x.respSigner
	} else {
		signer = neofscryptotest.ECDSAPrivateKey()
	}
	// body
	bs, err := signMessage(signer, resp.GetBody())
	if err != nil {
		return nil, fmt.Errorf("sign body: %w", err)
	}
	// meta
	ms, err := signMessage(signer, resp.GetMetaHeader())
	if err != nil {
		return nil, fmt.Errorf("sign meta: %w", err)
	}
	// origin
	ors, err := signMessage(signer, (*protosession.ResponseVerificationHeader)(nil))
	if err != nil {
		return nil, fmt.Errorf("sign verification header's origin: %w", err)
	}
	return &protosession.ResponseVerificationHeader{
		BodySignature:   bs,
		MetaSignature:   ms,
		OriginSignature: ors,
	}, nil
}

// func signature shortener.
type testedClientOp = func(*Client) error

// asserts that built test server expecting particular X-headers receives them
// from the connected [Client] through on specified op execution. The op must be
// executed with all the correct parameters to return no error.
func testRequestXHeaders[SRV interface{ checkRequestXHeaders([]string) }](
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
	c := connect(t, srv)

	srv.checkRequestXHeaders(xhdrs)
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
	st, ok := status.FromError(err)
	require.True(t, ok, err)
	require.Equal(t, codes.Unknown, st.Code())
	require.Contains(t, st.Message(), transport.Error())
}

// asserts that given [Client] op returns an expected error when built test
// server always responds with gRPC status error. The op must be executed with
// all the correct parameters.
func testTransportFailure[SRV interface{ setHandlerError(error) }](
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
func testInvalidResponseVerificationHeader[SRV interface{ respondWithoutSigning() }](
	t testing.TB,
	newSrv func() SRV,
	connect func(t testing.TB, srv any) *Client,
	op testedClientOp,
) {
	srv := newSrv()
	srv.respondWithoutSigning()
	// TODO: add cases with less radical corruption such as replacing one byte or
	//  dropping only one of the signatures.
	// Note: TBD during transition to proto/* packages in current repository.
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
func testInvalidResponseBodies[BODY any, SRV interface{ respondWithBody(*BODY) }](
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

// asserts that given [Client] op returns expected context errors when user
// passes done context. The op must be executed with the provided context and
// correct other parameters.
func testContextErrors[SRV any](
	t *testing.T,
	newSrv func() SRV,
	connect func(t testing.TB, srv any) *Client,
	op func(context.Context, *Client) error,
) {
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/624")
	srv := newSrv()
	c := connect(t, srv)
	require.NoError(t, op(context.Background(), c))
	t.Run("cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := op(ctx, c)
		require.ErrorIs(t, err, context.Canceled)
	})
	t.Run("timed out", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now())
		t.Cleanup(cancel)
		err := op(ctx, c)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
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
				require.ErrorContains(t, err, "status: code = unrecognized message = any message")
				require.ErrorIs(t, err, apistatus.ErrUnrecognizedStatusV2)
				require.ErrorAs(t, err, new(*apistatus.UnrecognizedStatusV2))
				t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/648")
				require.ErrorIs(t, err, apistatus.Error)
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
				require.ErrorIs(t, err, apistatus.Error)
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
				require.ErrorContains(t, err, expectedErrMsg)
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
func testIncorrectUnaryRPCResponseFormat(t testing.TB, svcName, method string, op testedClientOp) {
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
	// TODO(https://github.com/nspcc-dev/neofs-sdk-go/issues/661): Although the
	//  client will not accept such a response, current error does not make it clear
	//  what exactly the problem is. It is worth reacting to the incorrect structure
	//  if possible.
}

// asserts that given [Client] op correctly reports meta information received
// from built test server when consuming the specified service. The op must be
// executed with all the correct parameters.
func testUnaryResponseCallback[SRV interface {
	respondWithMeta(*protosession.ResponseMetaHeader)
	signResponsesBy(ecdsa.PrivateKey)
}](
	t testing.TB,
	newSrv func() SRV,
	newSvc func(t testing.TB, srv any) testService,
	op testedClientOp,
) {
	srv := newSrv()
	srvSigner := neofscryptotest.Signer()
	srvPub := neofscrypto.PublicKeyBytes(srvSigner.Public())
	srv.signResponsesBy(srvSigner.ECDSAPrivateKey)
	srvEpoch := rand.Uint64()
	srv.respondWithMeta(&protosession.ResponseMetaHeader{Epoch: srvEpoch})

	var collected []ResponseMetaInfo
	var handlerErr error
	handler := func(meta ResponseMetaInfo) error {
		collected = append(collected, meta)
		return handlerErr
	}
	assert := func(expEpoch uint64, expPub []byte) {
		require.Len(t, collected, 1)
		require.Equal(t, expEpoch, collected[0].Epoch())
		require.Equal(t, expPub, collected[0].ResponderKey())
		collected = nil
	}

	c := newCustomClient(t, func(prm *PrmInit) { prm.SetResponseInfoCallback(handler) }, newSvc(t, srv))
	// [Client.EndpointInfo] is always called to dial the server: this is also submitted
	assert(testServerStateOnDial.epoch, testServerStateOnDial.pub)

	err := op(c)
	require.NoError(t, err)
	assert(srvEpoch, srvPub)

	handlerErr = errors.New("any response meta handler failure")
	err = op(c)
	require.ErrorContains(t, err, "response callback error")
	require.ErrorIs(t, err, handlerErr)
	assert(srvEpoch, srvPub)
}

// checks that the [Client] correctly keeps exec statistics of specified ops
// performing communication with built test server. All operations must comply
// with the tested service.
//
// If non-stat failure cases are specified, they must include request signature
// failure caused by the op signer parameter.
func testStatistic[SRV interface {
	setSleepDuration(time.Duration)
	setHandlerError(error)
}](
	t testing.TB,
	newSrv func() SRV,
	newSvc func(t testing.TB, srv any) testService,
	expMtd stat.Method,
	customNonStatFailures []testedClientOp,
	customStatFailures []testedClientOp,
	validInputCall testedClientOp,
) {
	srv := newSrv()
	svc := newSvc(t, srv)

	type collectedItem struct {
		pub      []byte
		endpoint string
		mtd      stat.Method
		dur      time.Duration
		err      error
	}
	var collected []collectedItem
	handler := func(pub []byte, endpoint string, mtd stat.Method, dur time.Duration, err error) {
		collected = append(collected, collectedItem{pub: pub, endpoint: endpoint, mtd: mtd, dur: dur, err: err})
	}
	assertCommon := func(mtd stat.Method, pub []byte, err error) {
		require.Len(t, collected, 1)
		require.Equal(t, pub, collected[0].pub)
		require.Equal(t, testServerEndpoint, collected[0].endpoint)
		require.Equal(t, mtd, collected[0].mtd)
		require.Positive(t, collected[0].dur)
		require.Equal(t, err, collected[0].err)
	}

	c := newCustomClient(t, func(prm *PrmInit) { prm.SetStatisticCallback(handler) }, svc)
	// [Client.EndpointInfo] is always called to dial the server: this is also submitted
	assertCommon(stat.MethodEndpointInfo, nil, nil) // server key is not yet received
	collected = nil

	assert := func(err error) {
		assertCommon(expMtd, testServerStateOnDial.pub, err)
	}

	// custom non-stat failures
	for _, getNonStatErr := range customNonStatFailures {
		err := getNonStatErr(c)
		require.Error(t, err)
		assert(nil)
		collected = nil
	}

	// custom stat failures
	for _, getStatErr := range customStatFailures {
		err := getStatErr(c)
		require.Error(t, err)
		assert(err)
		collected = nil
	}

	if len(customNonStatFailures) == 0 {
		// sign request failure
		signerCp := c.prm.signer
		c.prm.signer = neofscryptotest.FailSigner(c.prm.signer)

		err := validInputCall(c)
		assertSignRequestErr(t, err)
		assert(err)
		collected = nil

		c.prm.signer = signerCp
	}

	// transport
	transportErr := errors.New("any transport failure")
	srv.setHandlerError(transportErr)

	err := validInputCall(c)
	assertTransportErr(t, transportErr, err)
	assert(err)
	collected = nil

	srv.setHandlerError(nil)

	// OK
	const sleepDur = 100 * time.Millisecond
	// duration is pretty short overall, but most likely larger than the exec time w/o sleep
	srv.setSleepDuration(sleepDur)

	err = validInputCall(c)
	require.NoError(t, err)
	assert(err)
	require.Greater(t, collected[0].dur, sleepDur)
}

func TestNewGRPC(t *testing.T) {
	conn, err := grpc.NewClient("any", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Run("negative stream message timeout", func(t *testing.T) {
		require.PanicsWithValue(t, "negative stream message timeout -1ms", func() {
			_, _ = NewGRPC(conn, nil, -time.Millisecond, nil)
		})
	})
	t.Run("default buffer pool", func(t *testing.T) {
		c, err := NewGRPC(conn, nil, 0, nil)
		require.NoError(t, err)
		require.NotNil(t, c.buffers)
		b := c.buffers.Get()
		require.IsType(t, (*[]byte)(nil), b)
		require.Len(t, *b.(*[]byte), 4<<20)
	})
	t.Run("default stream message timeout", func(t *testing.T) {
		c, err := NewGRPC(conn, nil, 0, nil)
		require.NoError(t, err)
		require.Equal(t, 10*time.Second, c.streamTimeout)
	})
	t.Run("no response interceptor", func(t *testing.T) {
		c, err := NewGRPC(conn, nil, 0, nil)
		require.NoError(t, err)
		require.Nil(t, c.prm.cbRespInfo)
	})

	const anyStreamMsgTimeout = time.Minute
	anySignBufferPool := &sync.Pool{New: func() any { return "Hello, world!" }}
	var caughtPub []byte
	anyInterceptorErr := errors.New("any interceptor error")

	c, err := NewGRPC(conn, anySignBufferPool, anyStreamMsgTimeout, func(pub []byte) error {
		caughtPub = pub
		return anyInterceptorErr
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	require.Equal(t, conn, c.conn)
	require.NotNil(t, c.accounting)
	require.NotNil(t, c.container)
	require.NotNil(t, c.netmap)
	require.NotNil(t, c.object)
	require.NotNil(t, c.reputation)
	require.NotNil(t, c.session)

	require.NotNil(t, c.prm.signer)
	require.IsType(t, neofsecdsa.SignerRFC6979{}, c.prm.signer)
	require.Equal(t, anySignBufferPool, c.buffers)

	require.Equal(t, anyStreamMsgTimeout, c.streamTimeout)

	require.NotNil(t, c.prm.cbRespInfo)
	pub := []byte("any public key")
	err = c.prm.cbRespInfo(ResponseMetaInfo{key: pub})
	require.ErrorIs(t, err, anyInterceptorErr)
	require.Equal(t, pub, caughtPub)
}
