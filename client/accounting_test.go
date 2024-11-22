package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	v2accounting "github.com/nspcc-dev/neofs-api-go/v2/accounting"
	protoaccounting "github.com/nspcc-dev/neofs-api-go/v2/accounting/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	protostatus "github.com/nspcc-dev/neofs-api-go/v2/status/grpc"
	accountingtest "github.com/nspcc-dev/neofs-sdk-go/accounting/test"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newDefaultAccountingService(srv protoaccounting.AccountingServiceServer) testService {
	return testService{desc: &protoaccounting.AccountingService_ServiceDesc, impl: srv}
}

// returns Client of Accounting service provided by given server.
func newTestAccountingClient(t testing.TB, srv protoaccounting.AccountingServiceServer) *Client {
	return newClient(t, newDefaultAccountingService(srv))
}

type testGetBalanceServer struct {
	protoaccounting.UnimplementedAccountingServiceServer

	reqXHdrs []string
	reqAcc   []byte

	handlerErr error

	respUnsigned bool
	respSigner   neofscrypto.Signer
	respMeta     *protosession.ResponseMetaHeader
	respBodyCons func() *protoaccounting.BalanceResponse_Body
}

// returns [protoaccounting.AccountingServiceServer] supporting Balance method
// only. Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestGetBalanceServer() *testGetBalanceServer { return new(testGetBalanceServer) }

// makes the server to assert that any request has given X-headers. By
// default, no headers are expected.
func (x *testGetBalanceServer) checkRequestXHeaders(xhdrs []string) {
	if len(xhdrs)%2 != 0 {
		panic("odd number of elements")
	}
	x.reqXHdrs = xhdrs
}

// makes the server to assert that any request is for the given
// account. By default, any account is accepted.
func (x *testGetBalanceServer) checkRequestAccount(acc user.ID) {
	x.reqAcc = acc[:]
}

// makes the server to always respond with the unsigned message. By default, any
// response is signed.
//
// Overrides signResponsesBy.
func (x *testGetBalanceServer) respondWithoutSigning() {
	x.respUnsigned = true
}

// makes the server to always sign responses using given signer. By default,
// random signer is used.
//
// Has no effect with respondWithoutSigning.
func (x *testGetBalanceServer) signResponsesBy(signer neofscrypto.Signer) {
	x.respSigner = signer
}

// makes the server to always respond with the specifically constructed body. By
// default, any valid body is returned.
//
// Conflicts with respondWithBalance.
func (x *testGetBalanceServer) respondWithBody(newBody func() *protoaccounting.BalanceResponse_Body) {
	x.respBodyCons = newBody
}

// makes the server to always respond with the given balance. By default, any
// valid balance is returned.
//
// Conflicts with respondWithBody.
func (x *testGetBalanceServer) respondWithBalance(balance *protoaccounting.Decimal) {
	x.respondWithBody(func() *protoaccounting.BalanceResponse_Body {
		return &protoaccounting.BalanceResponse_Body{Balance: balance}
	})
}

// makes the server to always respond with the given meta header. By default,
// empty header is returned.
//
// Conflicts with respondWithStatus.
func (x *testGetBalanceServer) respondWithMeta(meta *protosession.ResponseMetaHeader) {
	x.respMeta = meta
}

// makes the server to always respond with the given status. By default, status
// OK is returned.
//
// Conflicts with respondWithMeta.
func (x *testGetBalanceServer) respondWithStatus(st *protostatus.Status) {
	x.respondWithMeta(&protosession.ResponseMetaHeader{Status: st})
}

// makes the server to return given error from the handler. By default, some
// response message is returned.
func (x *testGetBalanceServer) setHandlerError(err error) {
	x.handlerErr = err
}

func (x *testGetBalanceServer) verifyBalanceRequest(req *protoaccounting.BalanceRequest) error {
	// signatures
	var reqV2 v2accounting.BalanceRequest
	if err := reqV2.FromGRPCMessage(req); err != nil {
		panic(err)
	}
	if err := verifyServiceMessage(&reqV2); err != nil {
		return newInvalidRequestVerificationHeaderErr(err)
	}
	// meta header
	metaHdr := req.MetaHeader
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
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
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
	// body
	body := req.Body
	switch {
	case body == nil:
		return newInvalidRequestBodyErr(errors.New("missing body"))
	case body.OwnerId == nil:
		return newErrMissingRequestBodyField("account")
	}
	if x.reqAcc != nil && !bytes.Equal(body.OwnerId.Value, x.reqAcc[:]) {
		return newErrInvalidRequestField("account", fmt.Errorf("test input mismatch"))
	}
	return nil
}

func (x *testGetBalanceServer) Balance(_ context.Context, req *protoaccounting.BalanceRequest) (*protoaccounting.BalanceResponse, error) {
	if err := x.verifyBalanceRequest(req); err != nil {
		return nil, err
	}

	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := protoaccounting.BalanceResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyCons != nil {
		resp.Body = x.respBodyCons()
	} else {
		resp.Body = &protoaccounting.BalanceResponse_Body{
			Balance: &protoaccounting.Decimal{
				Value:     rand.Int63(),
				Precision: rand.Uint32(),
			},
		}
	}

	if x.respUnsigned {
		return &resp, nil
	}

	var respV2 v2accounting.BalanceResponse
	if err := respV2.FromGRPCMessage(&resp); err != nil {
		panic(err)
	}
	signer := x.respSigner
	if signer == nil {
		signer = neofscryptotest.Signer()
	}
	if err := signServiceMessage(signer, &respV2, nil); err != nil {
		return nil, fmt.Errorf("sign response message: %w", err)
	}

	return respV2.ToGRPCMessage().(*protoaccounting.BalanceResponse), nil
}

func TestClient_BalanceGet(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	anyUsr := usertest.ID()
	var anyValidPrm PrmBalanceGet
	anyValidPrm.SetAccount(anyUsr)

	t.Run("invalid user input", func(t *testing.T) {
		t.Run("missing account", func(t *testing.T) {
			_, err := c.BalanceGet(ctx, PrmBalanceGet{})
			require.ErrorIs(t, err, ErrMissingAccount)
		})
	})
	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/

		balance := accountingtest.Decimal()
		acc := usertest.ID()
		xhdrs := []string{
			"x-key1", "x-val1",
			"x-key2", "x-val2",
		}

		srv := newTestGetBalanceServer()
		srv.checkRequestAccount(acc)
		srv.checkRequestXHeaders(xhdrs)
		srv.respondWithBalance(&protoaccounting.Decimal{
			Value:     balance.Value(),
			Precision: balance.Precision(),
		})

		c := newTestAccountingClient(t, srv)

		var prm PrmBalanceGet
		prm.SetAccount(acc)
		prm.WithXHeaders(xhdrs...)
		res, err := c.BalanceGet(ctx, prm)
		require.NoError(t, err)
		require.Equal(t, balance, res)

		// statuses
		type customStatusTestcase struct {
			msg    string
			detail *protostatus.Status_Detail
			assert func(testing.TB, error)
		}
		for _, tc := range []struct {
			code     uint32
			err      error
			constErr error
			custom   []customStatusTestcase
		}{
			// TODO: use const codes after transition to current module's proto lib
			{code: 1024, err: new(apistatus.ServerInternal), constErr: apistatus.ErrServerInternal, custom: []customStatusTestcase{
				{msg: "some server failure", assert: func(t testing.TB, err error) {
					var e *apistatus.ServerInternal
					require.ErrorAs(t, err, &e)
					require.Equal(t, "some server failure", e.Message())
				}},
			}},
			{code: 1025, err: new(apistatus.WrongMagicNumber), constErr: apistatus.ErrWrongMagicNumber, custom: []customStatusTestcase{
				{assert: func(t testing.TB, err error) {
					var e *apistatus.WrongMagicNumber
					require.ErrorAs(t, err, &e)
					_, ok := e.CorrectMagic()
					require.Zero(t, ok)
				}},
				{
					detail: &protostatus.Status_Detail{Id: 0, Value: []byte{140, 15, 162, 245, 219, 236, 37, 191}},
					assert: func(t testing.TB, err error) {
						var e *apistatus.WrongMagicNumber
						require.ErrorAs(t, err, &e)
						magic, ok := e.CorrectMagic()
						require.EqualValues(t, 1, ok)
						require.EqualValues(t, uint64(10092464466800944575), magic)
					},
				},
				{
					detail: &protostatus.Status_Detail{Id: 0, Value: []byte{1, 2, 3}},
					assert: func(t testing.TB, err error) {
						var e *apistatus.WrongMagicNumber
						require.ErrorAs(t, err, &e)
						_, ok := e.CorrectMagic()
						require.EqualValues(t, -1, ok)
					},
				},
			}},
			{code: 1026, err: new(apistatus.SignatureVerification), constErr: apistatus.ErrSignatureVerification, custom: []customStatusTestcase{
				{msg: "invalid request signature", assert: func(t testing.TB, err error) {
					var e *apistatus.SignatureVerification
					require.ErrorAs(t, err, &e)
					require.Equal(t, "invalid request signature", e.Message())
				}},
			}},
			{code: 1027, err: new(apistatus.NodeUnderMaintenance), constErr: apistatus.ErrNodeUnderMaintenance, custom: []customStatusTestcase{
				{msg: "node is under maintenance", assert: func(t testing.TB, err error) {
					var e *apistatus.NodeUnderMaintenance
					require.ErrorAs(t, err, &e)
					require.Equal(t, "node is under maintenance", e.Message())
				}},
			}},
		} {
			st := &protostatus.Status{Code: tc.code}
			srv.respondWithStatus(st)

			res, err := c.BalanceGet(ctx, prm)
			require.Zero(t, res)
			require.ErrorAs(t, err, &tc.err)
			require.ErrorIs(t, err, tc.constErr)

			for _, tcCustom := range tc.custom {
				st.Message = tcCustom.msg
				if tcCustom.detail != nil {
					st.Details = []*protostatus.Status_Detail{tcCustom.detail}
				}
				srv.respondWithStatus(st)

				_, err := c.BalanceGet(ctx, prm)
				require.ErrorAs(t, err, &tc.err)
				tcCustom.assert(t, tc.err)
			}
		}
	})
	t.Run("sign request failure", func(t *testing.T) {
		c.prm.signer = neofscryptotest.FailSigner(neofscryptotest.Signer())
		_, err := c.BalanceGet(ctx, anyValidPrm)
		require.ErrorContains(t, err, "sign request")
	})
	t.Run("transport failure", func(t *testing.T) {
		// note: errors returned from gRPC handlers are gRPC statuses, therefore,
		// strictly speaking, they are not transport errors (like connection refusal for
		// example). At the same time, according to the NeoFS protocol, all its statuses
		// are transmitted in the message. So, returning an error from gRPC handler
		// instead of a status field in the response is a protocol violation and can be
		// equated to a transport error.
		transportErr := errors.New("any transport failure")
		srv := newTestGetBalanceServer()
		srv.setHandlerError(transportErr)
		c := newTestAccountingClient(t, srv)

		_, err := c.BalanceGet(ctx, anyValidPrm)
		require.ErrorContains(t, err, "rpc failure")
		require.ErrorContains(t, err, "write request")
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Unknown, st.Code())
		require.Equal(t, err.Error(), st.Message())
	})
	t.Run("response message decoding failure", func(t *testing.T) {
		svc := testService{
			desc: &grpc.ServiceDesc{ServiceName: "neo.fs.v2.accounting.AccountingService", Methods: []grpc.MethodDesc{
				{
					MethodName: "Balance",
					Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
						return timestamppb.Now(), nil // any completely different message
					},
				},
			}},
			impl: nil, // disables interface assert
		}
		c := newClient(t, svc)
		_, err := c.BalanceGet(ctx, anyValidPrm)
		require.ErrorContains(t, err, "invalid response signature")
		// TODO: Although the client will not accept such a response, current error
		//  does not make it clear what exactly the problem is. It is worth reacting to
		//  the incorrect structure if possible.
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		srv := newTestGetBalanceServer()
		srv.respondWithoutSigning()
		// TODO: add cases with less radical corruption such as replacing one byte or
		//  dropping only one of the signatures
		c := newTestAccountingClient(t, srv)

		_, err := c.BalanceGet(ctx, anyValidPrm)
		require.ErrorContains(t, err, "invalid response signature")
	})
	t.Run("invalid response body", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			body      *protoaccounting.BalanceResponse_Body
			assertErr func(testing.TB, error)
		}{
			{name: "missing", body: nil, assertErr: func(t testing.TB, err error) {
				require.ErrorIs(t, err, MissingResponseFieldErr{})
				require.EqualError(t, err, "missing balance field in the response")
				// TODO: worth clarifying that body is completely missing?
			}},
			{name: "missing", body: new(protoaccounting.BalanceResponse_Body), assertErr: func(t testing.TB, err error) {
				require.ErrorIs(t, err, MissingResponseFieldErr{})
				require.EqualError(t, err, "missing balance field in the response")
			}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				srv := newTestGetBalanceServer()
				srv.respondWithBody(func() *protoaccounting.BalanceResponse_Body { return tc.body })
				c := newTestAccountingClient(t, srv)

				_, err := c.BalanceGet(ctx, anyValidPrm)
				tc.assertErr(t, err)
			})
		}
	})
	t.Run("response callback", func(t *testing.T) {
		// NetmapService.LocalNodeInfo is called on dial, so it should also be
		// initialized. The handler is called for it too.
		netmapSrvSigner := neofscryptotest.Signer()
		netmapSrvEpoch := rand.Uint64()
		netmapSrv := newTestGetNodeInfoServer()
		netmapSrv.respondWithMeta(&protosession.ResponseMetaHeader{Epoch: netmapSrvEpoch})
		netmapSrv.signResponsesBy(netmapSrvSigner)

		accountingSrvSigner := neofscryptotest.Signer()
		accountingSrvEpoch := netmapSrvEpoch + 1
		accountingSrv := newTestGetBalanceServer()
		accountingSrv.respondWithMeta(&protosession.ResponseMetaHeader{Epoch: accountingSrvEpoch})
		accountingSrv.signResponsesBy(accountingSrvSigner)

		var collected []ResponseMetaInfo
		var cbErr error
		c := newClientWithResponseCallback(t, func(meta ResponseMetaInfo) error {
			collected = append(collected, meta)
			return cbErr
		},
			newDefaultAccountingService(accountingSrv),
			newDefaultNetmapServiceDesc(netmapSrv),
		)

		_, err := c.BalanceGet(ctx, anyValidPrm)
		require.NoError(t, err)
		require.Equal(t, []ResponseMetaInfo{
			{key: netmapSrvSigner.PublicKeyBytes, epoch: netmapSrvEpoch},
			{key: accountingSrvSigner.PublicKeyBytes, epoch: accountingSrvEpoch},
		}, collected)

		cbErr = errors.New("any response meta handler failure")
		_, err = c.BalanceGet(ctx, anyValidPrm)
		require.ErrorContains(t, err, "response callback error")
		require.ErrorContains(t, err, err.Error())
		require.Len(t, collected, 3)
		require.Equal(t, collected[2], collected[1])
	})
}
