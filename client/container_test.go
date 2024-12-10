package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	protoacl "github.com/nspcc-dev/neofs-api-go/v2/acl/grpc"
	apicontainer "github.com/nspcc-dev/neofs-api-go/v2/container"
	protocontainer "github.com/nspcc-dev/neofs-api-go/v2/container/grpc"
	protonetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	apigrpc "github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// returns Client-compatible Container service handled by given server. Provided
// server must implement [protocontainer.ContainerServiceServer]: the parameter
// is not of this type to support generics.
func newDefaultContainerService(t testing.TB, srv any) testService {
	require.Implements(t, (*protocontainer.ContainerServiceServer)(nil), srv)
	return testService{desc: &protocontainer.ContainerService_ServiceDesc, impl: srv}
}

// returns Client of Container service provided by given server. Provided server
// must implement [protocontainer.ContainerServiceServer]: the parameter is
// not of this type to support generics.
func newTestContainerClient(t testing.TB, srv any) *Client {
	return newClient(t, newDefaultContainerService(t, srv))
}

// for sharing between servers of requests with RFC 6979 signature of particular
// data.
type testRFC6979DataSignatureServerSettings[
	SIGNED apigrpc.Message,
	SIGNEDV2 any,
	SIGNEDV2PTR interface {
		*SIGNEDV2
		signedMessageV2
	},
] struct {
	reqCreds         *authCredentials
	reqDataSignature *neofscrypto.Signature
}

// makes the server to assert that any request's payload is signed by s. By
// default, any signer is accepted.
//
// Has no effect with checkRequestDataSignature.
func (x *testRFC6979DataSignatureServerSettings[_, _, _]) authenticateRequestPayload(s neofscrypto.Signer) {
	c := authCredentialsFromSigner(s)
	x.reqCreds = &c
}

// makes the server to assert that any request carries given signature without
// verification. By default, any signature matching the data is accepted.
//
// Overrides checkRequestDataSignerKey.
func (x *testRFC6979DataSignatureServerSettings[_, _, _]) checkRequestDataSignature(s neofscrypto.Signature) {
	x.reqDataSignature = &s
}

func (x testRFC6979DataSignatureServerSettings[_, _, _]) verifyDataSignature(signedField string, data []byte, m *protorefs.SignatureRFC6979) error {
	field := signedField + " signature"
	if m == nil {
		return newErrMissingRequestBodyField(field)
	}
	if x.reqDataSignature != nil {
		if err := checkSignatureRFC6979Transport(*x.reqDataSignature, m); err != nil {
			return newErrInvalidRequestField(field, err)
		}
		return nil
	}
	if err := verifyDataSignature(data, &protorefs.Signature{
		Key:    m.Key,
		Sign:   m.Sign,
		Scheme: protorefs.SignatureScheme_ECDSA_RFC6979_SHA256,
	}, x.reqCreds); err != nil {
		return newErrInvalidRequestField(field, err)
	}
	return nil
}

func (x testRFC6979DataSignatureServerSettings[SIGNED, SIGNEDV2, SIGNEDV2PTR]) verifyMessageSignature(signedField string, signed SIGNED, m *protorefs.SignatureRFC6979) error {
	mV2 := SIGNEDV2PTR(new(SIGNEDV2))
	if err := mV2.FromGRPCMessage(signed); err != nil {
		panic(err)
	}
	return x.verifyDataSignature(signedField, mV2.StableMarshal(nil), m)
}

// for sharing between servers of requests with a container session token.
type testContainerSessionServerSettings struct {
	expectedToken *session.Container
}

// makes the server to assert that any request carries given session token. By
// default, session token must not be attached.
func (x *testContainerSessionServerSettings) checkRequestSessionToken(st session.Container) {
	x.expectedToken = &st
}

func (x testContainerSessionServerSettings) verifySessionToken(m *protosession.SessionToken) error {
	if m == nil {
		if x.expectedToken != nil {
			return newInvalidRequestMetaHeaderErr(errors.New("session token is missing while should not be"))
		}
		return nil
	}
	if x.expectedToken == nil {
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	}
	if err := checkContainerSessionTransport(*x.expectedToken, m); err != nil {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("session token: %w", err))
	}
	return nil
}

type testPutContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.PutRequest_Body,
		apicontainer.PutRequestBody,
		*apicontainer.PutRequestBody,
		*protocontainer.PutRequest,
		apicontainer.PutRequest,
		*apicontainer.PutRequest,
		*protocontainer.PutResponse_Body,
		apicontainer.PutResponseBody,
		*apicontainer.PutResponseBody,
		*protocontainer.PutResponse,
		apicontainer.PutResponse,
		*apicontainer.PutResponse,
	]
	testContainerSessionServerSettings
	testRFC6979DataSignatureServerSettings[*protocontainer.Container, apicontainer.Container, *apicontainer.Container]
	reqContainer *container.Container
}

// returns [protocontainer.ContainerServiceServer] supporting Put method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestPutContainerServer() *testPutContainerServer { return new(testPutContainerServer) }

// makes the server to assert that any request carries given container. By
// default, any valid container is accepted.
func (x *testPutContainerServer) checkRequestContainer(cnr container.Container) {
	x.reqContainer = &cnr
}

func (x *testPutContainerServer) verifyRequest(req *protocontainer.PutRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	if err := x.verifySessionToken(req.MetaHeader.SessionToken); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// container
	if body.Container == nil {
		return newErrMissingRequestBodyField("container")
	}
	if x.reqContainer != nil {
		if err := checkContainerTransport(*x.reqContainer, body.Container); err != nil {
			return newErrInvalidRequestField("container", err)
		}
	}
	// signature
	return x.verifyMessageSignature("container", body.Container, body.Signature)
}

func (x *testPutContainerServer) Put(_ context.Context, req *protocontainer.PutRequest) (*protocontainer.PutResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protocontainer.PutResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinPutContainerResponseBody).(*protocontainer.PutResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testGetContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.GetRequest_Body,
		apicontainer.GetRequestBody,
		*apicontainer.GetRequestBody,
		*protocontainer.GetRequest,
		apicontainer.GetRequest,
		*apicontainer.GetRequest,
		*protocontainer.GetResponse_Body,
		apicontainer.GetResponseBody,
		*apicontainer.GetResponseBody,
		*protocontainer.GetResponse,
		apicontainer.GetResponse,
		*apicontainer.GetResponse,
	]
	testRequiredContainerIDServerSettings
}

// returns [protocontainer.ContainerServiceServer] supporting Get method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestGetContainerServer() *testGetContainerServer { return new(testGetContainerServer) }

func (x *testGetContainerServer) verifyRequest(req *protocontainer.GetRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// session token
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// bearer token
	if req.MetaHeader.BearerToken != nil {
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	return x.verifyRequestContainerID(body.ContainerId)
}

func (x *testGetContainerServer) Get(_ context.Context, req *protocontainer.GetRequest) (*protocontainer.GetResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protocontainer.GetResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinGetContainerResponseBody).(*protocontainer.GetResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testListContainersServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.ListRequest_Body,
		apicontainer.ListRequestBody,
		*apicontainer.ListRequestBody,
		*protocontainer.ListRequest,
		apicontainer.ListRequest,
		*apicontainer.ListRequest,
		*protocontainer.ListResponse_Body,
		apicontainer.ListResponseBody,
		*apicontainer.ListResponseBody,
		*protocontainer.ListResponse,
		apicontainer.ListResponse,
		*apicontainer.ListResponse,
	]
	reqOwner *user.ID
}

// returns [protocontainer.ContainerServiceServer] supporting List method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestListContainersServer() *testListContainersServer { return new(testListContainersServer) }

// makes the server to assert that any request carries given owner. By default,
// any user is accepted.
func (x *testListContainersServer) checkOwner(owner user.ID) { x.reqOwner = &owner }

func (x *testListContainersServer) verifyRequest(req *protocontainer.ListRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// owner
	if body.OwnerId == nil {
		return newErrMissingRequestBodyField("owner")
	}
	if x.reqOwner != nil {
		if err := checkUserIDTransport(*x.reqOwner, body.OwnerId); err != nil {
			return newErrInvalidRequestField("owner", err)
		}
	}
	return nil
}

func (x *testListContainersServer) List(_ context.Context, req *protocontainer.ListRequest) (*protocontainer.ListResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protocontainer.ListResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinListContainersResponseBody).(*protocontainer.ListResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testDeleteContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.DeleteRequest_Body,
		apicontainer.DeleteRequestBody,
		*apicontainer.DeleteRequestBody,
		*protocontainer.DeleteRequest,
		apicontainer.DeleteRequest,
		*apicontainer.DeleteRequest,
		*protocontainer.DeleteResponse_Body,
		apicontainer.DeleteResponseBody,
		*apicontainer.DeleteResponseBody,
		*protocontainer.DeleteResponse,
		apicontainer.DeleteResponse,
		*apicontainer.DeleteResponse,
	]
	testContainerSessionServerSettings
	testRequiredContainerIDServerSettings
	testRFC6979DataSignatureServerSettings[*protorefs.ContainerID, refs.ContainerID, *refs.ContainerID]
}

// returns [protocontainer.ContainerServiceServer] supporting Delete method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestDeleteContainerServer() *testDeleteContainerServer { return new(testDeleteContainerServer) }

func (x *testDeleteContainerServer) verifyRequest(req *protocontainer.DeleteRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// session token
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	if err := x.verifySessionToken(req.MetaHeader.SessionToken); err != nil {
		return err
	}
	// bearer token
	if req.MetaHeader.BearerToken != nil {
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// ID
	mc := body.GetContainerId()
	if err := x.verifyRequestContainerID(mc); err != nil {
		return err
	}
	// signature
	return x.verifyDataSignature("container ID", mc.GetValue(), body.Signature)
}

func (x *testDeleteContainerServer) Delete(_ context.Context, req *protocontainer.DeleteRequest) (*protocontainer.DeleteResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protocontainer.DeleteResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinDeleteContainerResponseBody).(*protocontainer.DeleteResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testGetEACLServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.GetExtendedACLRequest_Body,
		apicontainer.GetExtendedACLRequestBody,
		*apicontainer.GetExtendedACLRequestBody,
		*protocontainer.GetExtendedACLRequest,
		apicontainer.GetExtendedACLRequest,
		*apicontainer.GetExtendedACLRequest,
		*protocontainer.GetExtendedACLResponse_Body,
		apicontainer.GetExtendedACLResponseBody,
		*apicontainer.GetExtendedACLResponseBody,
		*protocontainer.GetExtendedACLResponse,
		apicontainer.GetExtendedACLResponse,
		*apicontainer.GetExtendedACLResponse,
	]
	testRequiredContainerIDServerSettings
}

// returns [protocontainer.ContainerServiceServer] supporting GetExtendedACL
// method only. Default implementation performs common verification of any
// request, and responds with any valid message. Some methods allow to tune the
// behavior.
func newTestGetEACLServer() *testGetEACLServer { return new(testGetEACLServer) }

func (x *testGetEACLServer) verifyRequest(req *protocontainer.GetExtendedACLRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// meta header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// ID
	return x.verifyRequestContainerID(body.ContainerId)
}

func (x *testGetEACLServer) GetExtendedACL(_ context.Context, req *protocontainer.GetExtendedACLRequest) (*protocontainer.GetExtendedACLResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protocontainer.GetExtendedACLResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinEACLResponseBody).(*protocontainer.GetExtendedACLResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testSetEACLServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.SetExtendedACLRequest_Body,
		apicontainer.SetExtendedACLRequestBody,
		*apicontainer.SetExtendedACLRequestBody,
		*protocontainer.SetExtendedACLRequest,
		apicontainer.SetExtendedACLRequest,
		*apicontainer.SetExtendedACLRequest,
		*protocontainer.SetExtendedACLResponse_Body,
		apicontainer.SetExtendedACLResponseBody,
		*apicontainer.SetExtendedACLResponseBody,
		*protocontainer.SetExtendedACLResponse,
		apicontainer.SetExtendedACLResponse,
		*apicontainer.SetExtendedACLResponse,
	]
	testContainerSessionServerSettings
	testRFC6979DataSignatureServerSettings[*protoacl.EACLTable, v2acl.Table, *v2acl.Table]
	reqEACL *eacl.Table
}

// returns [protocontainer.ContainerServiceServer] supporting SetExtendedACL
// method only. Default implementation performs common verification of any
// request, and responds with any valid message. Some methods allow to tune the
// behavior.
func newTestSetEACLServer() *testSetEACLServer { return new(testSetEACLServer) }

// makes the server to assert that any request carries given eACL. By
// default, any eACL is accepted.
func (x *testSetEACLServer) checkRequestEACL(eACL eacl.Table) { x.reqEACL = &eACL }

func (x *testSetEACLServer) verifyRequest(req *protocontainer.SetExtendedACLRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// session token
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	if err := x.verifySessionToken(req.MetaHeader.SessionToken); err != nil {
		return err
	}
	// bearer token
	if req.MetaHeader.BearerToken != nil {
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// eACL
	if body.Eacl == nil {
		return newErrMissingRequestBodyField("eACL")
	}
	if x.reqEACL != nil {
		if err := checkEACLTransport(*x.reqEACL, body.Eacl); err != nil {
			return newErrInvalidRequestField("eACL", err)
		}
	}
	// signature
	return x.verifyMessageSignature("eACL", body.Eacl, body.Signature)
}

func (x *testSetEACLServer) SetExtendedACL(_ context.Context, req *protocontainer.SetExtendedACLRequest) (*protocontainer.SetExtendedACLResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}
	resp := &protocontainer.SetExtendedACLResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinSetEACLResponseBody).(*protocontainer.SetExtendedACLResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

type testAnnounceContainerSpaceServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonUnaryServerSettings[
		*protocontainer.AnnounceUsedSpaceRequest_Body,
		apicontainer.AnnounceUsedSpaceRequestBody,
		*apicontainer.AnnounceUsedSpaceRequestBody,
		*protocontainer.AnnounceUsedSpaceRequest,
		apicontainer.AnnounceUsedSpaceRequest,
		*apicontainer.AnnounceUsedSpaceRequest,
		*protocontainer.AnnounceUsedSpaceResponse_Body,
		apicontainer.AnnounceUsedSpaceResponseBody,
		*apicontainer.AnnounceUsedSpaceResponseBody,
		*protocontainer.AnnounceUsedSpaceResponse,
		apicontainer.AnnounceUsedSpaceResponse,
		*apicontainer.AnnounceUsedSpaceResponse,
	]
	reqAnnouncements []container.SizeEstimation
}

// returns [protocontainer.ContainerServiceServer] supporting AnnounceUsedSpace
// method only. Default implementation performs common verification of any
// request, and responds with any valid message. Some methods allow to tune the
// behavior.
func newTestAnnounceContainerSpaceServer() *testAnnounceContainerSpaceServer {
	return new(testAnnounceContainerSpaceServer)
}

// makes the server to assert that any request carries given announcements. By
// default, any valid values are accepted.
func (x *testAnnounceContainerSpaceServer) checkRequestAnnouncements(els []container.SizeEstimation) {
	x.reqAnnouncements = els
}

func (x *testAnnounceContainerSpaceServer) verifyRequest(req *protocontainer.AnnounceUsedSpaceRequest) error {
	if err := x.testCommonUnaryServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// mead header
	switch metaHdr := req.MetaHeader; {
	case metaHdr.Ttl != 2:
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected 2", metaHdr.Ttl))
	case metaHdr.SessionToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	case metaHdr.BearerToken != nil:
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// announcements
	if len(body.Announcements) == 0 {
		return newErrMissingRequestBodyField("announcements")
	}
	if x.reqAnnouncements != nil {
		if v1, v2 := len(x.reqAnnouncements), len(body.Announcements); v1 != v2 {
			return fmt.Errorf("number of records (client: %d, message: %d)", v1, v2)
		}
		for i := range x.reqAnnouncements {
			if err := checkContainerSizeEstimationTransport(x.reqAnnouncements[i], body.Announcements[i]); err != nil {
				return newErrInvalidRequestField("announcements", fmt.Errorf("elements#%d: %w", i, err))
			}
		}
	}
	return nil
}

func (x *testAnnounceContainerSpaceServer) AnnounceUsedSpace(_ context.Context, req *protocontainer.AnnounceUsedSpaceRequest) (*protocontainer.AnnounceUsedSpaceResponse, error) {
	time.Sleep(x.handlerSleepDur)
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}
	if x.handlerErr != nil {
		return nil, x.handlerErr
	}

	resp := &protocontainer.AnnounceUsedSpaceResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = proto.Clone(validMinUsedSpaceResponseBody).(*protocontainer.AnnounceUsedSpaceResponse_Body)
	}

	var err error
	resp.VerifyHeader, err = x.signResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("sign response: %w", err)
	}
	return resp, nil
}

func TestClient_ContainerPut(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmContainerPut
	anyValidContainer := containertest.Container()
	anyValidSigner := neofscryptotest.Signer().RFC6979

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestPutContainerServer()
				c := newTestContainerClient(t, srv)

				srv.checkRequestContainer(anyValidContainer)
				srv.authenticateRequestPayload(anyValidSigner)
				srv.authenticateRequest(c.prm.signer)
				_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, PrmContainerPut{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testStatusResponses(t, newTestPutContainerServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("precalculated container signature", func(t *testing.T) {
					srv := newTestPutContainerServer()
					c := newTestContainerClient(t, srv)

					var sig neofscrypto.Signature
					sig.SetPublicKeyBytes([]byte("any public key"))
					sig.SetValue([]byte("any value"))
					opts := anyValidOpts
					opts.AttachSignature(sig)

					srv.checkRequestDataSignature(sig)
					_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestPutContainerServer()
					c := newTestContainerClient(t, srv)

					st := sessiontest.ContainerSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.PutResponse_Body
					}{
						{name: "min", body: validMinPutContainerResponseBody},
						{name: "full", body: validFullPutContainerResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestPutContainerServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							id, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
							require.NoError(t, checkContainerIDTransport(id, tc.body.GetContainerId()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestPutContainerServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "Put", func(c *Client) error {
						_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestPutContainerServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protocontainer.PutResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil,
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, MissingResponseFieldErr{})
								require.EqualError(t, err, "missing container ID field in the response")
							}},
						{name: "empty", body: new(protocontainer.PutResponse_Body),
							assertErr: func(t testing.TB, err error) {
								require.ErrorIs(t, err, MissingResponseFieldErr{})
								require.EqualError(t, err, "missing container ID field in the response")
							}},
					}
					// 1. container ID
					for _, tc := range invalidContainerIDProtoTestcases {
						id := proto.Clone(validProtoContainerIDs[0]).(*protorefs.ContainerID)
						tc.corrupt(id)
						body := &protocontainer.PutResponse_Body{ContainerId: id}
						tcs = append(tcs, testcase{name: "container ID/" + tc.name, body: body, assertErr: func(tb testing.TB, err error) {
							require.EqualError(t, err, "invalid container ID field in the response: "+tc.msg)
						}})
					}

					testInvalidResponseBodies(t, newTestPutContainerServer, newTestContainerClient, tcs, func(c *Client) error {
						_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ContainerPut(ctx, anyValidContainer, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("sign container failure", func(t *testing.T) {
		c := newClient(t)
		t.Run("wrong scheme", func(t *testing.T) {
			_, err := c.ContainerPut(ctx, anyValidContainer, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), anyValidOpts)
			require.EqualError(t, err, "calculate container signature: incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
		})
		t.Run("signer failure", func(t *testing.T) {
			_, err := c.ContainerPut(ctx, anyValidContainer, neofscryptotest.FailSigner(neofscryptotest.Signer()), anyValidOpts)
			require.ErrorContains(t, err, "calculate container signature")
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestPutContainerServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestPutContainerServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestPutContainerServer, newDefaultContainerService, func(c *Client) error {
			_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestPutContainerServer, newDefaultContainerService, stat.MethodContainerPut,
			[]testedClientOp{func(c *Client) error {
				_, err := c.ContainerPut(ctx, anyValidContainer, nil, anyValidOpts)
				return err
			}},
			[]testedClientOp{func(c *Client) error {
				_, err := c.ContainerPut(ctx, anyValidContainer, neofscryptotest.FailSigner(anyValidSigner), anyValidOpts)
				return err
			}}, func(c *Client) error {
				_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_ContainerGet(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmContainerGet
	anyID := cidtest.ID()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestGetContainerServer()
				c := newTestContainerClient(t, srv)

				srv.checkRequestContainerID(anyID)
				srv.authenticateRequest(c.prm.signer)
				_, err := c.ContainerGet(ctx, anyID, PrmContainerGet{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestGetContainerServer, newTestContainerClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.ContainerGet(ctx, anyID, opts)
						return err
					})
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.GetResponse_Body
					}{
						{name: "min", body: validMinGetContainerResponseBody},
						{name: "full", body: validFullGetContainerResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestGetContainerServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							cnr, err := c.ContainerGet(ctx, anyID, anyValidOpts)
							require.NoError(t, err)
							require.NoError(t, checkContainerTransport(cnr, tc.body.GetContainer()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestGetContainerServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "Get", func(c *Client) error {
						_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestGetContainerServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protocontainer.GetResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "missing container in response")
							}},
						{name: "empty", body: new(protocontainer.GetResponse_Body),
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "missing container in response")
							}},
					}
					// 1. container
					type invalidContainerTestcase = struct {
						name, msg string
						corrupt   func(valid *protocontainer.Container)
					}
					// 1.1 version
					ctcs := []invalidContainerTestcase{{name: "version/missing", msg: "missing version", corrupt: func(valid *protocontainer.Container) {
						valid.Version = nil
					}}}
					// 1.2 owner
					ctcs = append(ctcs, invalidContainerTestcase{name: "owner/missing", msg: "missing owner", corrupt: func(valid *protocontainer.Container) {
						valid.OwnerId = nil
					}})
					for _, tc := range invalidUserIDProtoTestcases {
						ctcs = append(ctcs, invalidContainerTestcase{
							name: "owner/" + tc.name, msg: "invalid owner: " + tc.msg,
							corrupt: func(valid *protocontainer.Container) { tc.corrupt(valid.OwnerId) },
						})
					}
					// 1.3 nonce
					ctcs = append(ctcs, invalidContainerTestcase{name: "nonce/missing", msg: "missing nonce", corrupt: func(valid *protocontainer.Container) {
						valid.Nonce = nil
					}})
					for _, tc := range invalidUUIDProtoTestcases {
						ctcs = append(ctcs, invalidContainerTestcase{
							name: "nonce/" + tc.name, msg: "invalid nonce: " + tc.msg,
							corrupt: func(valid *protocontainer.Container) { valid.Nonce = tc.corrupt(valid.Nonce) },
						})
					}
					// 1.4 basic ACL
					// 1.5  attributes
					for _, tc := range []struct {
						name, msg string
						attrs     []string
					}{
						{name: "attributes/empty key", msg: "empty attribute key",
							attrs: []string{"k1", "v1", "", "v2", "k3", "v3"}},
						{name: "attributes/empty value", msg: `empty "k2" attribute value`,
							attrs: []string{"k1", "v1", "k2", "", "k3", "v3"}},
						{name: "attributes/duplicated", msg: "duplicated attribute k1",
							attrs: []string{"k1", "v1", "k2", "v2", "k1", "v3"}},
						{name: "attributes/timestamp/invalid", msg: `invalid attribute value Timestamp: foo (strconv.ParseInt: parsing "foo": invalid syntax)`,
							attrs: []string{"k1", "v1", "Timestamp", "foo", "k1", "v3"}},
					} {
						require.Zero(t, len(tc.attrs)%2)
						as := make([]*protocontainer.Container_Attribute, 0, len(tc.attrs)/2)
						for i := range len(tc.attrs) / 2 {
							as = append(as, &protocontainer.Container_Attribute{Key: tc.attrs[2*i], Value: tc.attrs[2*i+1]})
						}
						ctcs = append(ctcs, invalidContainerTestcase{
							name: "attributes/" + tc.name, msg: tc.msg,
							corrupt: func(valid *protocontainer.Container) { valid.Attributes = as },
						})
					}
					// 1.6 policy
					ctcs = append(ctcs, invalidContainerTestcase{name: "policy/missing", msg: "missing placement policy", corrupt: func(valid *protocontainer.Container) {
						valid.PlacementPolicy = nil
					}})
					for _, tc := range []struct {
						name, msg string
						corrupt   func(valid *protonetmap.PlacementPolicy)
					}{
						{name: "missing replicas", msg: "missing replicas", corrupt: func(valid *protonetmap.PlacementPolicy) {
							valid.Replicas = nil
						}},
						// TODO: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
						// {name: "selectors/clause/negative", msg: "invalid selector #1: negative clause -1", corrupt: func(valid *protonetmap.PlacementPolicy) {
						// 	valid.Selectors[1].Clause = -1
						// }},
						// {name: "filters/op/negative", msg: "invalid filter #1: negative op -1", corrupt: func(valid *protonetmap.PlacementPolicy) {
						// 	valid.Filters[1].Op = -1
						// }},
					} {
						ctcs = append(ctcs, invalidContainerTestcase{
							name: "policy" + tc.name, msg: "invalid placement policy: " + tc.msg,
							corrupt: func(valid *protocontainer.Container) { tc.corrupt(valid.PlacementPolicy) },
						})
					}

					for _, tc := range ctcs {
						body := proto.Clone(validFullGetContainerResponseBody).(*protocontainer.GetResponse_Body)
						tc.corrupt(body.Container)
						tcs = append(tcs, testcase{
							name: "container/" + tc.name, body: body,
							assertErr: func(t testing.TB, err error) {
								require.EqualError(t, err, "invalid container in response: "+tc.msg)
							},
						})
					}

					testInvalidResponseBodies(t, newTestGetContainerServer, newTestContainerClient, tcs, func(c *Client) error {
						_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestGetContainerServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestGetContainerServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestGetContainerServer, newDefaultContainerService, func(c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestGetContainerServer, newDefaultContainerService, stat.MethodContainerGet,
			nil, nil, func(c *Client) error {
				_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_ContainerList(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmContainerList
	anyUser := usertest.ID()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestListContainersServer()
				c := newTestContainerClient(t, srv)

				srv.checkOwner(anyUser)
				srv.authenticateRequest(c.prm.signer)
				_, err := c.ContainerList(ctx, anyUser, PrmContainerList{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestListContainersServer, newTestContainerClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.ContainerList(ctx, anyUser, opts)
						return err
					})
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.ListResponse_Body
					}{
						{name: "nil", body: nil},
						{name: "min", body: validMinListContainersResponseBody},
						{name: "full", body: validFullListContainersResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestListContainersServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							res, err := c.ContainerList(ctx, anyUser, anyValidOpts)
							require.NoError(t, err)
							mids := tc.body.GetContainerIds()
							require.Len(t, res, len(mids))
							for i := range mids {
								require.NoError(t, checkContainerIDTransport(res[i], mids[i]), i)
							}
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestListContainersServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "List", func(c *Client) error {
						_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestListContainersServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protocontainer.ListResponse_Body]
					var tcs []testcase
					// 1. container IDs
					type invalidIDsTestcase = struct {
						name, msg string
						corrupt   func(valid []*protorefs.ContainerID) // 3 elements
					}
					tcsIDs := []invalidIDsTestcase{
						{
							name:    "nil element",
							msg:     "invalid length 0",
							corrupt: func(valid []*protorefs.ContainerID) { valid[1] = nil },
						},
					}
					for _, tc := range invalidContainerIDProtoTestcases {
						tcsIDs = append(tcsIDs, invalidIDsTestcase{
							name:    "invalid element/" + tc.name,
							msg:     tc.msg,
							corrupt: func(valid []*protorefs.ContainerID) { tc.corrupt(valid[1]) },
						})
					}
					for _, tc := range tcsIDs {
						ids := make([]*protorefs.ContainerID, len(validProtoContainerIDs))
						for i, id := range validProtoContainerIDs {
							ids[i] = proto.Clone(id).(*protorefs.ContainerID)
						}
						tc.corrupt(ids)
						body := &protocontainer.ListResponse_Body{ContainerIds: ids}
						tcs = append(tcs, testcase{name: "container IDs/" + tc.name, body: body, assertErr: func(tb testing.TB, err error) {
							require.EqualError(t, err, "invalid ID in the response: "+tc.msg)
						}})
					}

					testInvalidResponseBodies(t, newTestListContainersServer, newTestContainerClient, tcs, func(c *Client) error {
						_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestListContainersServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestListContainersServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestListContainersServer, newDefaultContainerService, func(c *Client) error {
			_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestListContainersServer, newDefaultContainerService, stat.MethodContainerList,
			nil, nil, func(c *Client) error {
				_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_ContainerDelete(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmContainerDelete
	anyValidSigner := neofscryptotest.Signer().RFC6979
	anyID := cidtest.ID()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestDeleteContainerServer()
				c := newTestContainerClient(t, srv)

				srv.checkRequestContainerID(anyID)
				srv.authenticateRequestPayload(anyValidSigner)
				srv.authenticateRequest(c.prm.signer)
				err := c.ContainerDelete(ctx, anyID, anyValidSigner, PrmContainerDelete{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						return c.ContainerDelete(ctx, anyID, anyValidSigner, opts)
					})
				})
				t.Run("precalculated container signature", func(t *testing.T) {
					srv := newTestDeleteContainerServer()
					c := newTestContainerClient(t, srv)

					var sig neofscrypto.Signature
					sig.SetPublicKeyBytes([]byte("any public key"))
					sig.SetValue([]byte("any value"))
					opts := anyValidOpts
					opts.AttachSignature(sig)

					srv.checkRequestDataSignature(sig)
					err := c.ContainerDelete(ctx, anyID, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestDeleteContainerServer()
					c := newTestContainerClient(t, srv)

					st := sessiontest.ContainerSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					err := c.ContainerDelete(ctx, anyID, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.DeleteResponse_Body
					}{
						{name: "min", body: validMinDeleteContainerResponseBody},
						{name: "full", body: validFullDeleteContainerResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestDeleteContainerServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							err := c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client) error {
						return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "Delete", func(c *Client) error {
						return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client) error {
						return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newClient(t)
		t.Run("missing signer", func(t *testing.T) {
			err := c.ContainerDelete(ctx, anyID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("sign ID failure", func(t *testing.T) {
		c := newTestContainerClient(t, newTestDeleteContainerServer())
		t.Run("wrong scheme", func(t *testing.T) {
			err := c.ContainerDelete(ctx, anyID, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), anyValidOpts)
			require.EqualError(t, err, "incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
			require.ErrorIs(t, err, neofscrypto.ErrIncorrectSigner)
		})
		t.Run("signer failure", func(t *testing.T) {
			err := c.ContainerDelete(ctx, anyID, neofscryptotest.FailSigner(anyValidSigner), anyValidOpts)
			require.ErrorContains(t, err, "calculate signature")
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestDeleteContainerServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client) error {
			return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestDeleteContainerServer, newDefaultContainerService, func(c *Client) error {
			return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestDeleteContainerServer, newDefaultContainerService, stat.MethodContainerDelete,
			[]testedClientOp{func(c *Client) error {
				return c.ContainerDelete(ctx, anyID, nil, anyValidOpts)
			}}, []testedClientOp{func(c *Client) error {
				return c.ContainerDelete(ctx, anyID, neofscryptotest.FailSigner(anyValidSigner), anyValidOpts)
			}}, func(c *Client) error {
				return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
			},
		)
	})
}

func TestClient_ContainerEACL(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmContainerEACL
	anyID := cidtest.ID()

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestGetEACLServer()
				c := newTestContainerClient(t, srv)

				srv.checkRequestContainerID(anyID)
				srv.authenticateRequest(c.prm.signer)
				_, err := c.ContainerEACL(ctx, anyID, PrmContainerEACL{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestGetEACLServer, newTestContainerClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						_, err := c.ContainerEACL(ctx, anyID, opts)
						return err
					})
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.GetExtendedACLResponse_Body
					}{
						{name: "min", body: validMinEACLResponseBody},
						{name: "full", body: validFullEACLResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestGetEACLServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							eACL, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
							require.NoError(t, err)
							require.NoError(t, checkEACLTransport(eACL, tc.body.GetEacl()))
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestGetEACLServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
						return err
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "GetExtendedACL", func(c *Client) error {
						_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
						return err
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestGetEACLServer, newTestContainerClient, func(c *Client) error {
						_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
						return err
					})
				})
				t.Run("payloads", func(t *testing.T) {
					type testcase = invalidResponseBodyTestcase[protocontainer.GetExtendedACLResponse_Body]
					tcs := []testcase{
						{name: "missing", body: nil, assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, MissingResponseFieldErr{})
							require.EqualError(t, err, "missing eACL field in the response")
						}},
						{name: "empty", body: new(protocontainer.GetExtendedACLResponse_Body), assertErr: func(t testing.TB, err error) {
							require.ErrorIs(t, err, MissingResponseFieldErr{})
							require.EqualError(t, err, "missing eACL field in the response")
						}},
					}
					// 1. eACL
					type invalidEACLTestcase = struct {
						name, msg string
						corrupt   func(valid *protoacl.EACLTable)
					}
					var etcs []invalidEACLTestcase
					// 1.2 container ID
					for _, tc := range invalidContainerIDProtoTestcases {
						etcs = append(etcs, invalidEACLTestcase{
							name: "container ID/" + tc.name, msg: "invalid container ID: " + tc.msg,
							corrupt: func(valid *protoacl.EACLTable) { tc.corrupt(valid.ContainerId) },
						})
					}
					// 1.3 records
					for _, tc := range []struct {
						name, msg string
						corrupt   func(valid *protoacl.EACLRecord)
					}{
						// TODO: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
						// {name: "op/negative", msg: "negative op -1", corrupt: func(valid *protoacl.EACLRecord) {
						// 	valid.Operation = -1
						// }},
						// {name: "action/negative", msg: "negative action -1", corrupt: func(valid *protoacl.EACLRecord) {
						// 	valid.Action = -1
						// }},
						// {name: "filters/header type/negative", msg: "invalid filter #1: negative header type -1", corrupt: func(valid *protoacl.EACLRecord) {
						// 	valid.Filters = []*protoacl.EACLRecord_Filter{{}, {HeaderType: -1}}
						// }},
						// {name: "filters/matcher/negative", msg: "invalid filter #1: negative matcher -1", corrupt: func(valid *protoacl.EACLRecord) {
						// 	valid.Filters = []*protoacl.EACLRecord_Filter{{}, {MatchType: -1}}
						// }},
						// {name: "targets/role/negative", msg: "invalid target #1: negative role -1", corrupt: func(valid *protoacl.EACLRecord) {
						// 	valid.Targets = []*protoacl.EACLRecord_Target{{}, {Role: -1}}
						// }},
					} {
						etcs = append(etcs, invalidEACLTestcase{
							name: "records/" + tc.name, msg: "invalid record #1: " + tc.msg,
							corrupt: func(valid *protoacl.EACLTable) { tc.corrupt(valid.Records[1]) },
						})
					}

					for _, tc := range etcs {
						body := proto.Clone(validFullEACLResponseBody).(*protocontainer.GetExtendedACLResponse_Body)
						tc.corrupt(body.Eacl)
						tcs = append(tcs, testcase{name: "eACL/" + tc.name, body: body, assertErr: func(tb testing.TB, err error) {
							require.EqualError(t, err, "invalid eACL field in the response: "+tc.msg)
						}})
					}

					testInvalidResponseBodies(t, newTestGetEACLServer, newTestContainerClient, tcs, func(c *Client) error {
						_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
						return err
					})
				})
			})
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestGetEACLServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestGetEACLServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestGetEACLServer, newDefaultContainerService, func(c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestGetEACLServer, newDefaultContainerService, stat.MethodContainerEACL,
			nil, nil, func(c *Client) error {
				_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
				return err
			},
		)
	})
}

func TestClient_ContainerSetEACL(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmContainerSetEACL
	anyValidSigner := usertest.User().RFC6979

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestSetEACLServer()
				c := newTestContainerClient(t, srv)

				srv.checkRequestEACL(anyValidEACL)
				srv.authenticateRequestPayload(anyValidSigner)
				srv.authenticateRequest(c.prm.signer)
				err := c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, PrmContainerSetEACL{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestSetEACLServer, newTestContainerClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, opts)
					})
				})
				t.Run("precalculated container signature", func(t *testing.T) {
					srv := newTestSetEACLServer()
					c := newTestContainerClient(t, srv)

					var sig neofscrypto.Signature
					sig.SetPublicKeyBytes([]byte("any public key"))
					sig.SetValue([]byte("any value"))
					opts := anyValidOpts
					opts.AttachSignature(sig)

					srv.checkRequestDataSignature(sig)
					err := c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, opts)
					require.NoError(t, err)
				})
				t.Run("session token", func(t *testing.T) {
					srv := newTestSetEACLServer()
					c := newTestContainerClient(t, srv)

					st := sessiontest.ContainerSigned(usertest.User())
					opts := anyValidOpts
					opts.WithinSession(st)

					srv.checkRequestSessionToken(st)
					err := c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, opts)
					require.NoError(t, err)
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.SetExtendedACLResponse_Body
					}{
						{name: "min", body: validMinSetEACLResponseBody},
						{name: "full", body: validFullSetEACLResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestSetEACLServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							err := c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
							require.NoError(t, err)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestSetEACLServer, newTestContainerClient, func(c *Client) error {
						return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "SetExtendedACL", func(c *Client) error {
						return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestSetEACLServer, newTestContainerClient, func(c *Client) error {
						return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		c := newTestContainerClient(t, newTestDeleteContainerServer())
		t.Run("missing signer", func(t *testing.T) {
			err := c.ContainerSetEACL(ctx, anyValidEACL, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
		t.Run("missing container ID in eACL", func(t *testing.T) {
			eACL := anyValidEACL
			eACL.SetCID(cid.ID{})
			err := c.ContainerSetEACL(ctx, eACL, anyValidSigner, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingEACLContainer)
		})
	})
	t.Run("sign container failure", func(t *testing.T) {
		c := newTestContainerClient(t, newTestSetEACLServer())
		t.Run("wrong scheme", func(t *testing.T) {
			err := c.ContainerSetEACL(ctx, anyValidEACL, user.NewAutoIDSigner(neofscryptotest.ECDSAPrivateKey()), anyValidOpts)
			require.EqualError(t, err, "incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
			require.ErrorIs(t, err, neofscrypto.ErrIncorrectSigner)
		})
		t.Run("signer failure", func(t *testing.T) {
			err := c.ContainerSetEACL(ctx, anyValidEACL, usertest.FailSigner(anyValidSigner), anyValidOpts)
			require.ErrorContains(t, err, "calculate signature")
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestSetEACLServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestSetEACLServer, newTestContainerClient, func(c *Client) error {
			return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestSetEACLServer, newDefaultContainerService, func(c *Client) error {
			return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestSetEACLServer, newDefaultContainerService, stat.MethodContainerSetEACL,
			[]testedClientOp{func(c *Client) error {
				return c.ContainerSetEACL(ctx, anyValidEACL, nil, anyValidOpts)
			}}, []testedClientOp{func(c *Client) error {
				return c.ContainerSetEACL(ctx, anyValidEACL, usertest.FailSigner(anyValidSigner), anyValidOpts)
			}}, func(c *Client) error {
				return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
			},
		)
	})
}

func TestClient_ContainerAnnounceUsedSpace(t *testing.T) {
	ctx := context.Background()
	var anyValidOpts PrmAnnounceSpace
	anyValidAnnouncements := []container.SizeEstimation{containertest.SizeEstimation(), containertest.SizeEstimation()}

	t.Run("messages", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("requests", func(t *testing.T) {
			t.Run("required data", func(t *testing.T) {
				srv := newTestAnnounceContainerSpaceServer()
				c := newTestContainerClient(t, srv)

				srv.checkRequestAnnouncements(anyValidAnnouncements)
				srv.authenticateRequest(c.prm.signer)
				err := c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, PrmAnnounceSpace{})
				require.NoError(t, err)
			})
			t.Run("options", func(t *testing.T) {
				t.Run("X-headers", func(t *testing.T) {
					testRequestXHeaders(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(c *Client, xhs []string) error {
						opts := anyValidOpts
						opts.WithXHeaders(xhs...)
						return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, opts)
					})
				})
			})
		})
		t.Run("responses", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				t.Run("payloads", func(t *testing.T) {
					for _, tc := range []struct {
						name string
						body *protocontainer.AnnounceUsedSpaceResponse_Body
					}{
						{name: "min", body: validMinUsedSpaceResponseBody},
						{name: "full", body: validFullUsedSpaceResponseBody},
					} {
						t.Run(tc.name, func(t *testing.T) {
							srv := newTestAnnounceContainerSpaceServer()
							c := newTestContainerClient(t, srv)

							srv.respondWithBody(tc.body)
							err := c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, PrmAnnounceSpace{})
							require.NoError(t, err)
						})
					}
				})
				t.Run("statuses", func(t *testing.T) {
					testStatusResponses(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(c *Client) error {
						return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
					})
				})
			})
			t.Run("invalid", func(t *testing.T) {
				t.Run("format", func(t *testing.T) {
					testIncorrectUnaryRPCResponseFormat(t, "container.ContainerService", "AnnounceUsedSpace", func(c *Client) error {
						return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
					})
				})
				t.Run("verification header", func(t *testing.T) {
					testInvalidResponseVerificationHeader(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(c *Client) error {
						return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
					})
				})
			})
		})
	})
	t.Run("invalid user input", func(t *testing.T) {
		t.Run("missing announcements", func(t *testing.T) {
			c := newClient(t)
			err := c.ContainerAnnounceUsedSpace(ctx, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingAnnouncements)
			err = c.ContainerAnnounceUsedSpace(ctx, []container.SizeEstimation{}, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingAnnouncements)
		})
	})
	t.Run("context", func(t *testing.T) {
		testContextErrors(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(ctx context.Context, c *Client) error {
			return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
		})
	})
	t.Run("sign request failure", func(t *testing.T) {
		testSignRequestFailure(t, func(c *Client) error {
			return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
		})
	})
	t.Run("transport failure", func(t *testing.T) {
		testTransportFailure(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(c *Client) error {
			return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testUnaryResponseCallback(t, newTestAnnounceContainerSpaceServer, newDefaultContainerService, func(c *Client) error {
			return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
		})
	})
	t.Run("exec statistics", func(t *testing.T) {
		testStatistic(t, newTestAnnounceContainerSpaceServer, newDefaultContainerService, stat.MethodContainerAnnounceUsedSpace,
			nil, []testedClientOp{func(c *Client) error {
				return c.ContainerAnnounceUsedSpace(ctx, nil, anyValidOpts)
			}}, func(c *Client) error {
				return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
			},
		)
	})
}
