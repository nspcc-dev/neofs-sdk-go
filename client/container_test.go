package client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	protoacl "github.com/nspcc-dev/neofs-api-go/v2/acl/grpc"
	apicontainer "github.com/nspcc-dev/neofs-api-go/v2/container"
	protocontainer "github.com/nspcc-dev/neofs-api-go/v2/container/grpc"
	protonetmap "github.com/nspcc-dev/neofs-api-go/v2/netmap/grpc"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	apisession "github.com/nspcc-dev/neofs-api-go/v2/session"
	protosession "github.com/nspcc-dev/neofs-api-go/v2/session/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var anyValidMinProtoContainer = &protocontainer.Container{
	Version: new(protorefs.Version),
	OwnerId: &protorefs.OwnerID{Value: []byte{53, 233, 31, 174, 37, 64, 241, 22, 182, 130, 7, 210, 222, 150, 85, 18, 106, 4,
		253, 122, 191, 90, 168, 187, 245}},
	Nonce: []byte{207, 5, 57, 28, 224, 103, 76, 207, 133, 186, 108, 96, 185, 52, 37, 205},
	PlacementPolicy: &protonetmap.PlacementPolicy{
		Replicas: make([]*protonetmap.Replica, 1),
	},
}

var anyValidFullProtoContainer = &protocontainer.Container{
	Version:  &protorefs.Version{Major: 538919038, Minor: 3957317479},
	OwnerId:  proto.Clone(anyValidMinProtoContainer.OwnerId).(*protorefs.OwnerID),
	Nonce:    bytes.Clone(anyValidMinProtoContainer.Nonce),
	BasicAcl: 1043832770,
	Attributes: []*protocontainer.Container_Attribute{
		{Key: "k1", Value: "v1"},
		{Key: "k2", Value: "v2"},
		{Key: "Name", Value: "any container name"},
		{Key: "Timestamp", Value: "1732577694"},
		{Key: "__NEOFS__NAME", Value: "any domain name"},
		{Key: "__NEOFS__ZONE", Value: "any domain zone"},
		{Key: "__NEOFS__DISABLE_HOMOMORPHIC_HASHING", Value: "true"},
	},
	PlacementPolicy: &protonetmap.PlacementPolicy{
		Replicas: []*protonetmap.Replica{
			{Count: 3060437, Selector: "selector1"},
			{Count: 156936495, Selector: "selector2"},
		},
		ContainerBackupFactor: 920231904,
		Selectors: []*protonetmap.Selector{
			{Name: "selector1", Count: 1663184999, Clause: 1, Attribute: "attribute1", Filter: "filter1"},
			{Name: "selector2", Count: 2649065896, Clause: 2, Attribute: "attribute2", Filter: "filter2"},
			{Name: "selector_max", Count: 2649065896, Clause: math.MaxInt32, Attribute: "attribute_max", Filter: "filter_max"},
		},
		Filters: []*protonetmap.Filter{
			{Name: "filter1", Key: "key1", Op: 0, Value: "value1", Filters: []*protonetmap.Filter{
				{},
				{},
			}},
			{Op: 1},
			{Op: 2},
			{Op: 3},
			{Op: 4},
			{Op: 5},
			{Op: 6},
			{Op: 7},
			{Op: 8},
			{Op: math.MaxInt32},
		},
		SubnetId: &protorefs.SubnetID{Value: 987533317},
	},
}

var anyValidProtoContainerIDs = []*protorefs.ContainerID{
	{Value: []byte{198, 137, 143, 192, 231, 50, 106, 89, 225, 118, 7, 42, 40, 225, 197, 183, 9, 205, 71, 140, 233, 30, 63, 73, 224, 244, 235, 18, 205, 45, 155, 236}},
	{Value: []byte{26, 71, 99, 242, 146, 121, 0, 142, 95, 50, 78, 190, 222, 104, 252, 72, 48, 219, 67, 226, 30, 90, 103, 51, 1, 234, 136, 143, 200, 240, 75, 250}},
	{Value: []byte{51, 124, 45, 83, 227, 119, 66, 76, 220, 196, 118, 197, 116, 44, 138, 83, 103, 102, 134, 191, 108, 124, 162, 255, 184, 137, 193, 242, 178, 10, 23, 29}},
}

var anyValidMinEACL = &protoacl.EACLTable{}

var anyValidFullEACL = &protoacl.EACLTable{
	Version:     &protorefs.Version{Major: 538919038, Minor: 3957317479},
	ContainerId: proto.Clone(anyValidProtoContainerIDs[0]).(*protorefs.ContainerID),
	Records: []*protoacl.EACLRecord{
		{},
		{Operation: 1, Action: 1},
		{Operation: 2, Action: 2},
		{Operation: 3, Action: 3},
		{Operation: 4, Action: math.MaxInt32},
		{Operation: 5},
		{Operation: 6},
		{Operation: 7},
		{Operation: math.MaxInt32},
		{Filters: []*protoacl.EACLRecord_Filter{
			{HeaderType: 0, MatchType: 0, Key: "key1", Value: "val1"},
			{HeaderType: 1, MatchType: 1},
			{HeaderType: 2, MatchType: 2},
			{HeaderType: 3, MatchType: 3},
			{HeaderType: math.MaxInt32, MatchType: 4},
			{MatchType: 5},
			{MatchType: 6},
			{MatchType: 7},
			{MatchType: math.MaxInt32},
		}},
		{Targets: []*protoacl.EACLRecord_Target{
			{Role: 0, Keys: [][]byte{[]byte("key1"), []byte("key2")}},
			{Role: 1},
			{Role: 2},
			{Role: 3},
			{Role: math.MaxInt32},
		}},
	},
}

var invalidProtoContainerIDTestcases = []struct {
	name    string
	msg     string
	corrupt func(valid *protorefs.ContainerID)
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

// returns Client-compatible Container service handler by given srv. Provided
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

// for sharing between servers of requests with container session token.
type testContainerSessionServerSettings struct {
	expectedToken *session.Container
}

// makes the server to assert that any request carries given session token. By
// default, session token must not be attached.
func (x *testContainerSessionServerSettings) checkRequestSessionToken(st session.Container) {
	x.expectedToken = &st
}

func (x *testContainerSessionServerSettings) verifySessionToken(m *protosession.SessionToken) error {
	if m == nil {
		if x.expectedToken != nil {
			return newInvalidRequestMetaHeaderErr(errors.New("session token is missing while should not be"))
		}
		return nil
	}
	if x.expectedToken == nil {
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	}
	var stV2 apisession.Token
	if err := stV2.FromGRPCMessage(m); err != nil {
		panic(err)
	}
	var st session.Container
	if err := st.ReadFromV2(stV2); err != nil {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("invalid session token: %w", err))
	}
	if !assert.ObjectsAreEqual(st, *x.expectedToken) {
		return newInvalidRequestMetaHeaderErr(errors.New("session token differs the parameterized one"))
	}
	return nil
}

// for sharing between servers of requests with RFC 6979 signature of particular
// data.
type testRFC6979DataSignatureServerSettings struct {
	reqPub           *ecdsa.PublicKey
	reqDataSignature *neofscrypto.Signature
}

// makes the server to assert that any request carries signature of the
// particular data calculated using given private key. By default, any key can
// be used.
//
// Has no effect with checkRequestDataSignature.
func (x *testRFC6979DataSignatureServerSettings) checkRequestDataSignerKey(pk ecdsa.PrivateKey) {
	x.reqPub = &pk.PublicKey
}

// makes the server to assert that any request carries given signature without
// verification. By default, any signature matching the data is accepted.
//
// Overrides checkRequestDataSignerKey.
func (x *testRFC6979DataSignatureServerSettings) checkRequestDataSignature(s neofscrypto.Signature) {
	x.reqDataSignature = &s
}

func (x *testRFC6979DataSignatureServerSettings) verifyDataSignature(signedField string, data []byte, m *protorefs.SignatureRFC6979) error {
	field := signedField + " signature"
	if m == nil {
		return newErrMissingRequestBodyField(field)
	}
	if x.reqDataSignature != nil {
		if exp := x.reqDataSignature.PublicKeyBytes(); !bytes.Equal(m.Key, exp) {
			return newErrInvalidRequestField(field, fmt.Errorf("public key %x != parameterized %x", m.Key, exp))
		}
		if exp := x.reqDataSignature.Value(); !bytes.Equal(m.Sign, exp) {
			return newErrInvalidRequestField(field, fmt.Errorf("value %x != parameterized %x", m.Sign, exp))
		}
		return nil
	}

	reqPubX, reqPubY := elliptic.UnmarshalCompressed(elliptic.P256(), m.Key)
	if reqPubX == nil {
		return newErrInvalidRequestField(field, fmt.Errorf("invalid EC point binary %x", m.Key))
	}
	if x.reqPub != nil && (reqPubX.Cmp(x.reqPub.X) != 0 || reqPubY.Cmp(x.reqPub.Y) != 0) {
		return newErrInvalidRequestField(field, fmt.Errorf("EC point != the parameterized one"))
	}
	sig := m.Sign
	if len(sig) != 64 {
		return newErrInvalidRequestField(field, fmt.Errorf("invalid signature length %d", len(sig)))
	}
	h := sha256.Sum256(data)
	if !ecdsa.Verify(&ecdsa.PublicKey{Curve: elliptic.P256(), X: reqPubX, Y: reqPubY}, h[:],
		new(big.Int).SetBytes(sig[0:32]), new(big.Int).SetBytes(sig[32:])) {
		return newErrInvalidRequestField(field, fmt.Errorf("signature mismatches the %s", signedField))
	}
	return nil
}

// for sharing between servers of requests with required container ID.
type testRequestContainerIDServerSettings struct {
	expectedReqCnrID []byte
}

// makes the server to assert that any request carries given container ID. By
// default, any ID is accepted.
func (x *testRequestContainerIDServerSettings) checkRequestContainerID(id cid.ID) {
	x.expectedReqCnrID = id[:]
}

func (x *testRequestContainerIDServerSettings) verifyRequestContainerID(m *protorefs.ContainerID) error {
	if m == nil {
		return newErrMissingRequestBodyField("container ID")
	}
	if x.expectedReqCnrID != nil && !bytes.Equal(m.Value, x.expectedReqCnrID) {
		return newErrInvalidRequestField("container ID", fmt.Errorf("container ID %x != the parameterized %x",
			m.Value, x.expectedReqCnrID))
	}
	return nil
}

type testPutContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.PutRequest,
		apicontainer.PutRequest,
		*apicontainer.PutRequest,
		protocontainer.PutResponse_Body,
		protocontainer.PutResponse,
		apicontainer.PutResponse,
		*apicontainer.PutResponse,
	]
	testContainerSessionServerSettings
	testRFC6979DataSignatureServerSettings
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

// makes the server to always respond with the given ID. By default, any
// valid ID is returned.
//
// Conflicts with respondWithBody.
func (x *testPutContainerServer) respondWithID(id []byte) {
	x.respondWithBody(&protocontainer.PutResponse_Body{
		ContainerId: &protorefs.ContainerID{Value: id},
	})
}

func (x *testPutContainerServer) verifyRequest(req *protocontainer.PutRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// session token
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
	var cnrV2 apicontainer.Container
	if err := cnrV2.FromGRPCMessage(body.Container); err != nil {
		panic(err)
	}
	var cnr container.Container
	if err := cnr.ReadFromV2(cnrV2); err != nil {
		return newErrInvalidRequestField("container", fmt.Errorf("invalid container: %w", err))
	}
	if x.reqContainer != nil && !assert.ObjectsAreEqual(cnr, *x.reqContainer) {
		return newErrInvalidRequestField("container", errors.New("container differs the parameterized one"))
	}
	// signature
	return x.verifyDataSignature("container", cnr.Marshal(), body.Signature)
}

func (x *testPutContainerServer) Put(_ context.Context, req *protocontainer.PutRequest) (*protocontainer.PutResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.PutResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		id := cidtest.ID()
		resp.Body = &protocontainer.PutResponse_Body{
			ContainerId: &protorefs.ContainerID{Value: id[:]},
		}
	}

	return x.signResponse(&resp)
}

type testGetContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.GetRequest,
		apicontainer.GetRequest,
		*apicontainer.GetRequest,
		protocontainer.GetResponse_Body,
		protocontainer.GetResponse,
		apicontainer.GetResponse,
		*apicontainer.GetResponse,
	]
	testRequestContainerIDServerSettings
}

// returns [protocontainer.ContainerServiceServer] supporting Get method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestGetContainerServer() *testGetContainerServer { return new(testGetContainerServer) }

// makes the server to always respond with the given container. By default, any
// valid container is returned.
//
// Conflicts with respondWithBody.
func (x *testGetContainerServer) respondWithContainer(cnr *protocontainer.Container) {
	x.respondWithBody(&protocontainer.GetResponse_Body{
		Container: cnr,
	})
}

func (x *testGetContainerServer) verifyRequest(req *protocontainer.GetRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	return x.verifyRequestContainerID(body.ContainerId)
}

func (x *testGetContainerServer) Get(_ context.Context, req *protocontainer.GetRequest) (*protocontainer.GetResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.GetResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = &protocontainer.GetResponse_Body{
			Container: proto.Clone(anyValidFullProtoContainer).(*protocontainer.Container),
			Signature: &protorefs.SignatureRFC6979{Key: []byte("any_key"), Sign: []byte("any_signature")},
			SessionToken: &protosession.SessionToken{
				Body: &protosession.SessionToken_Body{
					Id:         []byte("any_ID"),
					OwnerId:    &protorefs.OwnerID{Value: []byte("any_user")},
					Lifetime:   &protosession.SessionToken_Body_TokenLifetime{Exp: 1, Nbf: 2, Iat: 3},
					SessionKey: []byte("any_session_key"),
				},
				Signature: &protorefs.Signature{
					Key:    []byte("any_key"),
					Sign:   []byte("any_signature"),
					Scheme: protorefs.SignatureScheme(rand.Int31()),
				},
			},
		}
	}

	return x.signResponse(&resp)
}

type testListContainersServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.ListRequest,
		apicontainer.ListRequest,
		*apicontainer.ListRequest,
		protocontainer.ListResponse_Body,
		protocontainer.ListResponse,
		apicontainer.ListResponse,
		*apicontainer.ListResponse,
	]
	reqOwner []byte
}

// returns [protocontainer.ContainerServiceServer] supporting List method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestListContainersServer() *testListContainersServer { return new(testListContainersServer) }

// makes the server to assert that any request carries given owner. By default,
// any user is accepted.
func (x *testListContainersServer) checkOwner(owner user.ID) { x.reqOwner = owner[:] }

// makes the server to always respond with the given IDs. By default, several
// valid IDs are returned.
//
// Conflicts with respondWithBody.
func (x *testListContainersServer) respondWithIDs(ids []*protorefs.ContainerID) {
	x.respondWithBody(&protocontainer.ListResponse_Body{
		ContainerIds: ids,
	})
}

func (x *testListContainersServer) verifyRequest(req *protocontainer.ListRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
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
	if x.reqOwner != nil && !bytes.Equal(body.OwnerId.Value, x.reqOwner) {
		return newErrInvalidRequestField("owner", fmt.Errorf("owner %x != the parameterized %x",
			body.OwnerId.Value, x.reqOwner))
	}
	return nil
}

func (x *testListContainersServer) List(_ context.Context, req *protocontainer.ListRequest) (*protocontainer.ListResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.ListResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		ids := make([]*protorefs.ContainerID, len(anyValidProtoContainerIDs))
		for i := range anyValidProtoContainerIDs {
			ids[i] = proto.Clone(anyValidProtoContainerIDs[i]).(*protorefs.ContainerID)
		}
		resp.Body = &protocontainer.ListResponse_Body{ContainerIds: ids}
	}

	return x.signResponse(&resp)
}

type testDeleteContainerServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.DeleteRequest,
		apicontainer.DeleteRequest,
		*apicontainer.DeleteRequest,
		protocontainer.DeleteResponse_Body,
		protocontainer.DeleteResponse,
		apicontainer.DeleteResponse,
		*apicontainer.DeleteResponse,
	]
	testContainerSessionServerSettings
	testRequestContainerIDServerSettings
	testRFC6979DataSignatureServerSettings
}

// returns [protocontainer.ContainerServiceServer] supporting Delete method only.
// Default implementation performs common verification of any request, and
// responds with any valid message. Some methods allow to tune the behavior.
func newTestDeleteContainerServer() *testDeleteContainerServer { return new(testDeleteContainerServer) }

func (x *testDeleteContainerServer) verifyRequest(req *protocontainer.DeleteRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// session token
	if err := x.verifySessionToken(req.MetaHeader.SessionToken); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	// ID
	if err := x.verifyRequestContainerID(body.ContainerId); err != nil {
		return err
	}
	// signature
	return x.verifyDataSignature("container ID", body.ContainerId.Value, body.Signature)
}

func (x *testDeleteContainerServer) Delete(_ context.Context, req *protocontainer.DeleteRequest) (*protocontainer.DeleteResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.DeleteResponse{
		MetaHeader: x.respMeta,
	}

	return x.signResponse(&resp)
}

type testGetEACLServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.GetExtendedACLRequest,
		apicontainer.GetExtendedACLRequest,
		*apicontainer.GetExtendedACLRequest,
		protocontainer.GetExtendedACLResponse_Body,
		protocontainer.GetExtendedACLResponse,
		apicontainer.GetExtendedACLResponse,
		*apicontainer.GetExtendedACLResponse,
	]
	testRequestContainerIDServerSettings
}

// returns [protocontainer.ContainerServiceServer] supporting GetExtendedACL
// method only. Default implementation performs common verification of any
// request, and responds with any valid message. Some methods allow to tune the
// behavior.
func newTestGetEACLServer() *testGetEACLServer { return new(testGetEACLServer) }

// makes the server to always respond with the given eACL. By default, any
// valid eACL is returned.
//
// Conflicts with respondWithBody.
func (x *testGetEACLServer) respondWithEACL(eACL *protoacl.EACLTable) {
	x.respondWithBody(&protocontainer.GetExtendedACLResponse_Body{
		Eacl: eACL,
	})
}

func (x *testGetEACLServer) verifyRequest(req *protocontainer.GetExtendedACLRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// body
	body := req.Body
	if body == nil {
		return newInvalidRequestBodyErr(errors.New("missing body"))
	}
	//
	return x.verifyRequestContainerID(body.ContainerId)
}

func (x *testGetEACLServer) GetExtendedACL(_ context.Context, req *protocontainer.GetExtendedACLRequest) (*protocontainer.GetExtendedACLResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.GetExtendedACLResponse{
		MetaHeader: x.respMeta,
	}
	if x.respBodyForced {
		resp.Body = x.respBody
	} else {
		resp.Body = &protocontainer.GetExtendedACLResponse_Body{
			Eacl:      proto.Clone(anyValidFullEACL).(*protoacl.EACLTable),
			Signature: &protorefs.SignatureRFC6979{Key: []byte("any_key"), Sign: []byte("any_signature")},
			SessionToken: &protosession.SessionToken{
				Body: &protosession.SessionToken_Body{
					Id:         []byte("any_ID"),
					OwnerId:    &protorefs.OwnerID{Value: []byte("any_user")},
					Lifetime:   &protosession.SessionToken_Body_TokenLifetime{Exp: 1, Nbf: 2, Iat: 3},
					SessionKey: []byte("any_session_key"),
				},
				Signature: &protorefs.Signature{
					Key:    []byte("any_key"),
					Sign:   []byte("any_signature"),
					Scheme: protorefs.SignatureScheme(rand.Int31()),
				},
			},
		}
	}

	return x.signResponse(&resp)
}

type testSetEACLServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.SetExtendedACLRequest,
		apicontainer.SetExtendedACLRequest,
		*apicontainer.SetExtendedACLRequest,
		protocontainer.SetExtendedACLResponse_Body,
		protocontainer.SetExtendedACLResponse,
		apicontainer.SetExtendedACLResponse,
		*apicontainer.SetExtendedACLResponse,
	]
	testContainerSessionServerSettings
	testRFC6979DataSignatureServerSettings
	reqEACL *eacl.Table
}

// returns [protocontainer.ContainerServiceServer] supporting SetExtendedACL
// method only. Default implementation performs common verification of any
// request, and responds with any valid message. Some methods allow to tune the
// behavior.
func newTestSetEACLServer() *testSetEACLServer { return new(testSetEACLServer) }

// makes the server to assert that any request carries given eACL. By
// default, any valid eACL is accepted.
func (x *testSetEACLServer) checkRequestEACL(eACL eacl.Table) { x.reqEACL = &eACL }

func (x *testSetEACLServer) verifyRequest(req *protocontainer.SetExtendedACLRequest) error {
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
	}
	// session token
	if err := x.verifySessionToken(req.MetaHeader.SessionToken); err != nil {
		return err
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
	var eACLV2 v2acl.Table
	if err := eACLV2.FromGRPCMessage(body.Eacl); err != nil {
		panic(err)
	}
	var eACL eacl.Table
	if err := eACL.ReadFromV2(eACLV2); err != nil {
		return newErrInvalidRequestField("eACL", fmt.Errorf("invalid container: %w", err))
	}
	if x.reqEACL != nil && !bytes.Equal(eACL.Marshal(), x.reqEACL.Marshal()) {
		return newErrInvalidRequestField("eACL", errors.New("eACL differs the parameterized one"))
	}
	// signature
	return x.verifyDataSignature("eACL", eACL.Marshal(), body.Signature)
}

func (x *testSetEACLServer) SetExtendedACL(_ context.Context, req *protocontainer.SetExtendedACLRequest) (*protocontainer.SetExtendedACLResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.SetExtendedACLResponse{
		MetaHeader: x.respMeta,
	}

	return x.signResponse(&resp)
}

type testAnnounceContainerSpaceServer struct {
	protocontainer.UnimplementedContainerServiceServer
	testCommonServerSettings[
		*protocontainer.AnnounceUsedSpaceRequest,
		apicontainer.AnnounceUsedSpaceRequest,
		*apicontainer.AnnounceUsedSpaceRequest,
		protocontainer.AnnounceUsedSpaceRequest,
		protocontainer.AnnounceUsedSpaceResponse,
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
	if err := x.testCommonServerSettings.verifyRequest(req); err != nil {
		return err
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
	es := make([]container.SizeEstimation, len(body.Announcements))
	for i := range body.Announcements {
		var esV2 apicontainer.UsedSpaceAnnouncement
		if err := esV2.FromGRPCMessage(body.Announcements[i]); err != nil {
			panic(err)
		}
		if err := es[i].ReadFromV2(esV2); err != nil {
			return newErrInvalidRequestField("announcements", fmt.Errorf("invalid element #%d: %w", i, err))
		}
	}
	if x.reqAnnouncements != nil && !assert.ObjectsAreEqual(x.reqAnnouncements, es) {
		return newErrInvalidRequestField("announcements", errors.New("elements differ the parameterized ones"))
	}
	return nil
}

func (x *testAnnounceContainerSpaceServer) AnnounceUsedSpace(_ context.Context, req *protocontainer.AnnounceUsedSpaceRequest) (*protocontainer.AnnounceUsedSpaceResponse, error) {
	if err := x.verifyRequest(req); err != nil {
		return nil, err
	}

	resp := protocontainer.AnnounceUsedSpaceResponse{
		MetaHeader: x.respMeta,
	}

	return x.signResponse(&resp)
}

func TestClient_ContainerPut(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()
	var anyValidOpts PrmContainerPut
	anyValidContainer := containertest.Container()
	anyValidSigner := neofscryptotest.Signer().RFC6979
	anyID := cidtest.ID()

	t.Run("invalid user input", func(t *testing.T) {
		t.Run("missing signer", func(t *testing.T) {
			_, err := c.ContainerPut(ctx, anyValidContainer, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("sign container failure", func(t *testing.T) {
		t.Run("wrong scheme", func(t *testing.T) {
			_, err := c.ContainerPut(ctx, anyValidContainer, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), anyValidOpts)
			require.EqualError(t, err, "calculate container signature: incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
		})
		t.Run("signer failure", func(t *testing.T) {
			_, err := c.ContainerPut(ctx, anyValidContainer, neofscryptotest.FailSigner(neofscryptotest.Signer()), anyValidOpts)
			require.ErrorContains(t, err, "calculate container signature")
		})
	})
	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			signer := neofscryptotest.Signer()
			srv := newTestPutContainerServer()
			srv.checkRequestContainer(anyValidContainer)
			srv.checkRequestDataSignerKey(signer.ECDSAPrivateKey)
			srv.respondWithID(anyID[:])
			c := newTestContainerClient(t, srv)

			id, err := c.ContainerPut(ctx, anyValidContainer, signer.RFC6979, PrmContainerPut{})
			require.NoError(t, err)
			require.Equal(t, anyID, id)
		})
		t.Run("options", func(t *testing.T) {
			t.Run("precalculated container signature", func(t *testing.T) {
				var sig neofscrypto.Signature
				sig.SetPublicKeyBytes([]byte("any public key"))
				sig.SetValue([]byte("any value"))
				srv := newTestPutContainerServer()
				srv.checkRequestDataSignature(sig)
				c := newTestContainerClient(t, srv)

				opts := anyValidOpts
				opts.AttachSignature(sig)
				_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, opts)
				require.NoError(t, err)
			})
			t.Run("X-headers", func(t *testing.T) {
				testRequestXHeaders(t, newTestPutContainerServer, newTestContainerClient, func(c *Client, xhs []string) error {
					opts := anyValidOpts
					opts.WithXHeaders(xhs...)
					_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, opts)
					return err
				})
			})
			t.Run("session token", func(t *testing.T) {
				st := sessiontest.ContainerSigned(usertest.User())
				srv := newTestPutContainerServer()
				srv.checkRequestSessionToken(st)
				c := newTestContainerClient(t, srv)

				opts := anyValidOpts
				opts.WithinSession(st)
				_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, opts)
				require.NoError(t, err)
			})
		})
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestPutContainerServer, newTestContainerClient, func(c *Client) error {
				_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
				return err
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "Put", func(c *Client) error {
			_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestPutContainerServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerPut(ctx, anyValidContainer, anyValidSigner, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response body", func(t *testing.T) {
		type testcase = invalidResponseBodyTestcase[protocontainer.PutResponse_Body]
		tcs := []testcase{
			{name: "missing", body: nil,
				assertErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, MissingResponseFieldErr{})
					require.EqualError(t, err, "missing container ID field in the response")
					// TODO: worth clarifying that body is completely missing
				}},
			{name: "empty", body: new(protocontainer.PutResponse_Body),
				assertErr: func(t testing.TB, err error) {
					require.ErrorIs(t, err, MissingResponseFieldErr{})
					require.EqualError(t, err, "missing container ID field in the response")
				}},
		}
		// container ID
		for _, tc := range invalidProtoContainerIDTestcases {
			id := proto.Clone(anyValidProtoContainerIDs[0]).(*protorefs.ContainerID)
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
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestPutContainerServer, newDefaultContainerService, func(c *Client) error {
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

	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			t.Run("full", func(t *testing.T) {
				srv := newTestGetContainerServer()
				srv.checkRequestContainerID(anyID)
				srv.respondWithContainer(proto.Clone(anyValidFullProtoContainer).(*protocontainer.Container))
				c := newTestContainerClient(t, srv)

				cnr, err := c.ContainerGet(ctx, anyID, PrmContainerGet{})
				require.NoError(t, err)
				require.EqualValues(t, anyValidFullProtoContainer.Version.Major, cnr.Version().Major())
				require.EqualValues(t, anyValidFullProtoContainer.Version.Minor, cnr.Version().Minor())
				require.EqualValues(t, anyValidFullProtoContainer.OwnerId.Value, cnr.Owner())
				require.EqualValues(t, anyValidFullProtoContainer.BasicAcl, cnr.BasicACL().Bits())
				require.Equal(t, anyValidFullProtoContainer.Attributes[0].Value, cnr.Attribute(anyValidFullProtoContainer.Attributes[0].Key))
				require.Equal(t, anyValidFullProtoContainer.Attributes[1].Value, cnr.Attribute(anyValidFullProtoContainer.Attributes[1].Key))
				require.Equal(t, "any container name", cnr.Name())
				require.EqualValues(t, 1732577694, cnr.CreatedAt().Unix())
				require.Equal(t, "any domain name", cnr.ReadDomain().Name())
				require.Equal(t, "any domain zone", cnr.ReadDomain().Zone())
				require.True(t, cnr.IsHomomorphicHashingDisabled())
				policy := cnr.PlacementPolicy()
				require.EqualValues(t, anyValidFullProtoContainer.PlacementPolicy.ContainerBackupFactor, policy.ContainerBackupFactor())
				rs := policy.Replicas()
				require.Len(t, rs, len(anyValidFullProtoContainer.PlacementPolicy.Replicas))
				for i := range rs {
					m := anyValidFullProtoContainer.PlacementPolicy.Replicas[i]
					require.EqualValues(t, m.Count, rs[i].NumberOfObjects())
					require.Equal(t, m.Selector, rs[i].SelectorName())
				}
				ss := policy.Selectors()
				require.Len(t, ss, len(anyValidFullProtoContainer.PlacementPolicy.Selectors))
				for i := range ss {
					m := anyValidFullProtoContainer.PlacementPolicy.Selectors[i]
					require.Equal(t, m.Name, ss[i].Name())
					require.EqualValues(t, m.Count, ss[i].NumberOfNodes())
					switch m.Clause {
					default:
						require.False(t, ss[i].IsSame())
						require.False(t, ss[i].IsDistinct())
					case protonetmap.Clause_SAME:
						require.True(t, ss[i].IsSame())
						require.False(t, ss[i].IsDistinct())
					case protonetmap.Clause_DISTINCT:
						require.False(t, ss[i].IsSame())
						require.True(t, ss[i].IsDistinct())
					}
					require.Equal(t, m.Attribute, ss[i].BucketAttribute())
					require.Equal(t, m.Filter, ss[i].FilterName())
				}
				fs := policy.Filters()
				require.Len(t, fs, len(anyValidFullProtoContainer.PlacementPolicy.Filters))
				for i, f := range fs {
					m := anyValidFullProtoContainer.PlacementPolicy.Filters[i]
					require.Equal(t, m.Name, f.Name())
					require.Equal(t, m.Key, f.Key())
					require.EqualValues(t, m.Value, f.Value())
					if i == 0 {
						subs := fs[i].SubFilters()
						require.Len(t, subs, len(m.Filters))
						for j, sub := range subs {
							m := m.Filters[j]
							require.Equal(t, m.Name, sub.Name())
							require.Equal(t, m.Key, sub.Key())
							require.EqualValues(t, m.Op, sub.Op())
							require.EqualValues(t, m.Value, sub.Value())
							require.Empty(t, sub.SubFilters())
						}
					} else {
						require.Empty(t, fs[i].SubFilters())
					}
				}
				for i, op := range []netmap.FilterOp{
					0,
					netmap.FilterOpEQ,
					netmap.FilterOpNE,
					netmap.FilterOpGT,
					netmap.FilterOpGE,
					netmap.FilterOpLT,
					netmap.FilterOpLE,
					netmap.FilterOpOR,
					netmap.FilterOpAND,
					math.MaxInt32,
				} {
					require.Equal(t, op, fs[i].Op())
				}
			})
			t.Run("min", func(t *testing.T) {
				srv := newTestGetContainerServer()
				m := proto.Clone(anyValidMinProtoContainer).(*protocontainer.Container)
				srv.respondWithContainer(m)
				c := newTestContainerClient(t, srv)

				cnr, err := c.ContainerGet(ctx, anyID, PrmContainerGet{})
				require.NoError(t, err)
				require.Zero(t, cnr.Version())
				require.EqualValues(t, anyValidMinProtoContainer.OwnerId.Value, cnr.Owner())
				require.Zero(t, cnr.BasicACL().Bits())
				cnr.IterateAttributes(func(key, val string) { t.Fatalf("unexpected attribute %q", key) })
				policy := cnr.PlacementPolicy()
				require.Len(t, policy.Replicas(), 1)
				require.Zero(t, policy.Replicas()[0])
				require.Zero(t, policy.ContainerBackupFactor())
				require.Empty(t, policy.Selectors())
				require.Empty(t, policy.Filters())
			})
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
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestGetContainerServer, newTestContainerClient, func(c *Client) error {
				_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
				return err
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "Get", func(c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestGetContainerServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response body", func(t *testing.T) {
		type testcase = invalidResponseBodyTestcase[protocontainer.GetResponse_Body]
		tcs := []testcase{
			{name: "missing", body: nil,
				assertErr: func(t testing.TB, err error) {
					require.EqualError(t, err, "missing container in response")
					// TODO: worth clarifying that body is completely missing
				}},
			{name: "empty", body: new(protocontainer.GetResponse_Body),
				assertErr: func(t testing.TB, err error) {
					require.EqualError(t, err, "missing container in response")
				}},
		}
		// container
		for _, tc := range []struct {
			name      string
			corrupt   func(valid *protocontainer.Container)
			assertErr func(testing.TB, error)
		}{
			{name: "missing version", corrupt: func(valid *protocontainer.Container) {
				valid.Version = nil
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: missing version")
			}},
			{name: "missing owner", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId = nil
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: missing owner")
			}},
			{name: "owner/empty", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId = new(protorefs.OwnerID)
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid owner: invalid length 0, expected 25")
			}},
			{name: "owner/undersize", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId.Value = valid.OwnerId.Value[:24]
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid owner: invalid length 24, expected 25")
			}},
			{name: "owner/oversize", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId.Value = append(valid.OwnerId.Value, 1)
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid owner: invalid length 26, expected 25")
			}},
			{name: "owner/wrong prefix", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId.Value[0] = 0x42
				h := sha256.Sum256(valid.OwnerId.Value[:21])
				hh := sha256.Sum256(h[:])
				copy(valid.OwnerId.Value[21:], hh[:])
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid owner: invalid prefix byte 0x42, expected 0x35")
			}},
			{name: "owner/wrong checksum", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId.Value[24]++
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid owner: checksum mismatch")
			}},
			{name: "owner/zero", corrupt: func(valid *protocontainer.Container) {
				valid.OwnerId.Value = make([]byte, 25)
			}, assertErr: func(t testing.TB, err error) {
				// TODO: better to return user.ErrZero in this case
				require.ErrorContains(t, err, "invalid container in response: invalid owner: invalid prefix byte 0x0, expected 0x35")
			}},
			{name: "missing nonce", corrupt: func(valid *protocontainer.Container) {
				valid.Nonce = nil
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: missing nonce")
			}},
			{name: "nonce/undersize", corrupt: func(valid *protocontainer.Container) {
				valid.Nonce = valid.Nonce[:15]
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid nonce: invalid UUID (got 15 bytes)")
			}},
			{name: "nonce/oversize", corrupt: func(valid *protocontainer.Container) {
				valid.Nonce = append(valid.Nonce, 1)
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid nonce: invalid UUID (got 17 bytes)")
			}},
			{name: "nonce/wrong version", corrupt: func(valid *protocontainer.Container) {
				valid.Nonce = make([]byte, 16)
				valid.Nonce[6] = 3 << 4
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: invalid nonce UUID version 3")
			}},
			{name: "attributes/empty key", corrupt: func(valid *protocontainer.Container) {
				valid.Attributes = []*protocontainer.Container_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "", Value: "v1"}, {Key: "k3", Value: "v3"},
				}
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: empty attribute key")
			}},
			{name: "attributes/empty value", corrupt: func(valid *protocontainer.Container) {
				valid.Attributes = []*protocontainer.Container_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "k2", Value: ""}, {Key: "k3", Value: "v3"},
				}
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: empty attribute value k2")
			}},
			{name: "attributes/duplicated", corrupt: func(valid *protocontainer.Container) {
				valid.Attributes = []*protocontainer.Container_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "k2", Value: "v1"}, {Key: "k1", Value: "v3"},
				}
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, "invalid container in response: duplicated attribute k1")
			}},
			{name: "attributes/timestamp", corrupt: func(valid *protocontainer.Container) {
				valid.Attributes = []*protocontainer.Container_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "Timestamp", Value: "foo"},
				}
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, `invalid container in response: invalid attribute value Timestamp: foo (strconv.ParseInt: parsing "foo": invalid syntax)`)
			}},
			{name: "missing policy", corrupt: func(valid *protocontainer.Container) {
				valid.PlacementPolicy = nil
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, `invalid container in response: missing placement policy`)
			}},
			{name: "policy/missing replicas", corrupt: func(valid *protocontainer.Container) {
				valid.PlacementPolicy.Replicas = nil
			}, assertErr: func(t testing.TB, err error) {
				require.EqualError(t, err, `invalid container in response: invalid placement policy: missing replicas`)
			}},
		} {
			cnr := proto.Clone(anyValidMinProtoContainer).(*protocontainer.Container)
			tc.corrupt(cnr)
			body := &protocontainer.GetResponse_Body{Container: cnr}
			tcs = append(tcs, testcase{name: "container/" + tc.name, body: body, assertErr: tc.assertErr})
		}

		testInvalidResponseBodies(t, newTestGetContainerServer, newTestContainerClient, tcs, func(c *Client) error {
			_, err := c.ContainerGet(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestGetContainerServer, newDefaultContainerService, func(c *Client) error {
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

	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			t.Run("full", func(t *testing.T) {
				srv := newTestListContainersServer()
				srv.checkOwner(anyUser)
				ids := make([]*protorefs.ContainerID, len(anyValidProtoContainerIDs))
				for i := range anyValidProtoContainerIDs {
					ids[i] = proto.Clone(anyValidProtoContainerIDs[i]).(*protorefs.ContainerID)
				}
				srv.respondWithIDs(ids)
				c := newTestContainerClient(t, srv)

				res, err := c.ContainerList(ctx, anyUser, PrmContainerList{})
				require.NoError(t, err)
				require.Len(t, res, len(ids))
				for i := range res {
					require.EqualValues(t, ids[i].Value, res[i])
				}
			})
			t.Run("min", func(t *testing.T) {
				srv := newTestListContainersServer()
				for i, fn := range []func(*testListContainersServer){
					func(srv *testListContainersServer) {
						srv.respondWithBody(nil)
					},
					func(srv *testListContainersServer) {
						srv.respondWithBody(new(protocontainer.ListResponse_Body))
					},
					func(srv *testListContainersServer) {
						srv.respondWithIDs(nil)
					},
				} {
					fn(srv)
					c := newTestContainerClient(t, srv)

					res, err := c.ContainerList(ctx, anyUser, PrmContainerList{})
					require.NoError(t, err, i)
					require.Empty(t, res, i)
				}
			})
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
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestListContainersServer, newTestContainerClient, func(c *Client) error {
				_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
				return err
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "List", func(c *Client) error {
			_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestListContainersServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerList(ctx, anyUser, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response body", func(t *testing.T) {
		type testcase = invalidResponseBodyTestcase[protocontainer.ListResponse_Body]
		var tcs []testcase
		// IDs
		type invalidIDsTestcase = struct {
			name    string
			msg     string
			corrupt func(valid []*protorefs.ContainerID) // 3 elements
		}
		tcsIDs := []invalidIDsTestcase{
			{
				name:    "nil element",
				msg:     "invalid length 0",
				corrupt: func(valid []*protorefs.ContainerID) { valid[1] = nil },
			},
		}
		for _, tc := range invalidProtoContainerIDTestcases {
			tcsIDs = append(tcsIDs, invalidIDsTestcase{
				name:    "invalid element/" + tc.name,
				msg:     tc.msg,
				corrupt: func(valid []*protorefs.ContainerID) { tc.corrupt(valid[1]) },
			})
		}
		for _, tc := range tcsIDs {
			ids := make([]*protorefs.ContainerID, len(anyValidProtoContainerIDs))
			for i, id := range anyValidProtoContainerIDs {
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
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestListContainersServer, newDefaultContainerService, func(c *Client) error {
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
	c := newClient(t)
	ctx := context.Background()
	var anyValidOpts PrmContainerDelete
	anyValidSigner := neofscryptotest.Signer().RFC6979
	anyID := cidtest.ID()

	t.Run("invalid user input", func(t *testing.T) {
		t.Run("missing signer", func(t *testing.T) {
			err := c.ContainerDelete(ctx, anyID, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingSigner)
		})
	})
	t.Run("sign ID failure", func(t *testing.T) {
		t.Run("wrong scheme", func(t *testing.T) {
			err := c.ContainerDelete(ctx, anyID, neofsecdsa.Signer(neofscryptotest.ECDSAPrivateKey()), anyValidOpts)
			// TODO: consider returning 'calculate ID signature' to distinguish from the request signatures
			// FIXME: currently unchecked and request attempt is done. Better to pre-check like Put does
			t.Skip()
			require.EqualError(t, err, "calculate signature: incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
		})
		t.Run("signer failure", func(t *testing.T) {
			err := c.ContainerDelete(ctx, anyID, neofscryptotest.FailSigner(neofscryptotest.Signer()), anyValidOpts)
			require.ErrorContains(t, err, "calculate signature")
		})
	})
	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			signer := neofscryptotest.Signer()
			srv := newTestDeleteContainerServer()
			srv.checkRequestContainerID(anyID)
			srv.checkRequestDataSignerKey(signer.ECDSAPrivateKey)
			c := newTestContainerClient(t, srv)

			err := c.ContainerDelete(ctx, anyID, signer.RFC6979, PrmContainerDelete{})
			require.NoError(t, err)
		})
		t.Run("options", func(t *testing.T) {
			t.Run("precalculated container signature", func(t *testing.T) {
				var sig neofscrypto.Signature
				sig.SetPublicKeyBytes([]byte("any public key"))
				sig.SetValue([]byte("any value"))
				srv := newTestDeleteContainerServer()
				srv.checkRequestDataSignature(sig)
				c := newTestContainerClient(t, srv)

				opts := anyValidOpts
				opts.AttachSignature(sig)
				err := c.ContainerDelete(ctx, anyID, anyValidSigner, opts)
				require.NoError(t, err)
			})
			t.Run("X-headers", func(t *testing.T) {
				testRequestXHeaders(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client, xhs []string) error {
					opts := anyValidOpts
					opts.WithXHeaders(xhs...)
					return c.ContainerDelete(ctx, anyID, anyValidSigner, opts)
				})
			})
			t.Run("session token", func(t *testing.T) {
				st := sessiontest.ContainerSigned(usertest.User())
				srv := newTestDeleteContainerServer()
				srv.checkRequestSessionToken(st)
				c := newTestContainerClient(t, srv)

				opts := anyValidOpts
				opts.WithinSession(st)
				err := c.ContainerDelete(ctx, anyID, anyValidSigner, opts)
				require.NoError(t, err)
			})
		})
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client) error {
				return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "Delete", func(c *Client) error {
			return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestDeleteContainerServer, newTestContainerClient, func(c *Client) error {
			return c.ContainerDelete(ctx, anyID, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestDeleteContainerServer, newDefaultContainerService, func(c *Client) error {
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

	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			t.Run("full", func(t *testing.T) {
				srv := newTestGetEACLServer()
				srv.checkRequestContainerID(anyID)
				srv.respondWithEACL(proto.Clone(anyValidFullEACL).(*protoacl.EACLTable))
				c := newTestContainerClient(t, srv)

				eACL, err := c.ContainerEACL(ctx, anyID, PrmContainerEACL{})
				require.NoError(t, err)
				ver := eACL.Version()
				require.EqualValues(t, anyValidFullEACL.Version.Major, ver.Major())
				require.EqualValues(t, anyValidFullEACL.Version.Minor, ver.Minor())
				cnr := eACL.GetCID()
				require.EqualValues(t, anyValidFullEACL.ContainerId.Value, cnr)

				rs := eACL.Records()
				require.Len(t, rs, 11)
				for _, r := range rs[:9] {
					require.Empty(t, r.Filters())
					require.Empty(t, r.Targets())
				}
				require.Empty(t, rs[9].Targets())
				require.Empty(t, rs[10].Filters())
				for i, a := range []eacl.Operation{
					0,
					eacl.OperationGet,
					eacl.OperationHead,
					eacl.OperationPut,
					eacl.OperationDelete,
					eacl.OperationSearch,
					eacl.OperationRange,
					eacl.OperationRangeHash,
					// FIXME: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					// math.MaxInt32,
				} {
					require.Equal(t, a, rs[i].Operation())
				}
				for i, a := range []eacl.Action{
					0,
					eacl.ActionAllow,
					eacl.ActionDeny,
					// FIXME: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					// math.MaxInt32,
				} {
					require.Equal(t, a, rs[i].Action())
				}

				fs := rs[9].Filters()
				require.Len(t, fs, 9)
				for i, f := range fs {
					mf := anyValidFullEACL.Records[9].Filters[i]
					require.Equal(t, mf.Key, f.Key())
					require.Equal(t, mf.Value, f.Value())
				}
				for i, typ := range []eacl.FilterHeaderType{
					0,
					eacl.HeaderFromRequest,
					eacl.HeaderFromObject,
					eacl.HeaderFromService,
					// FIXME: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					// math.MaxInt32,
				} {
					require.Equal(t, typ, fs[i].From())
				}
				for i, m := range []eacl.Match{
					0,
					eacl.MatchStringEqual,
					eacl.MatchStringNotEqual,
					eacl.MatchNotPresent,
					eacl.MatchNumGT,
					eacl.MatchNumGE,
					eacl.MatchNumLT,
					eacl.MatchNumLE,
					// FIXME: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					// math.MaxInt32,
				} {
					require.Equal(t, m, fs[i].Matcher())
				}

				ts := rs[10].Targets()
				require.Len(t, ts, 5)
				for i, tgt := range ts {
					mt := anyValidFullEACL.Records[10].Targets[i]
					require.Equal(t, mt.Keys, tgt.RawSubjects())
				}
				for i, r := range []eacl.Role{
					0,
					eacl.RoleUser,
					eacl.RoleSystem,
					eacl.RoleOthers,
					// FIXME: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
					// math.MaxInt32,
				} {
					require.Equal(t, r, ts[i].Role())
				}
			})
			t.Run("min", func(t *testing.T) {
				srv := newTestGetEACLServer()
				srv.respondWithEACL(proto.Clone(anyValidMinEACL).(*protoacl.EACLTable))
				c := newTestContainerClient(t, srv)

				eACL, err := c.ContainerEACL(ctx, anyID, PrmContainerEACL{})
				require.NoError(t, err)
				require.Zero(t, eACL.Version())
				require.True(t, eACL.GetCID().IsZero())
				require.Empty(t, eACL.Records())
			})
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
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestGetEACLServer, newTestContainerClient, func(c *Client) error {
				_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
				return err
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "GetExtendedACL", func(c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestGetEACLServer, newTestContainerClient, func(c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("invalid response body", func(t *testing.T) {
		type testcase = invalidResponseBodyTestcase[protocontainer.GetExtendedACLResponse_Body]
		tcs := []testcase{
			{name: "missing", body: nil, assertErr: func(t testing.TB, err error) {
				require.ErrorIs(t, err, MissingResponseFieldErr{})
				require.EqualError(t, err, "missing eACL field in the response")
				// TODO: worth clarifying that body is completely missing
			}},
			{name: "empty", body: new(protocontainer.GetExtendedACLResponse_Body), assertErr: func(t testing.TB, err error) {
				require.ErrorIs(t, err, MissingResponseFieldErr{})
				require.EqualError(t, err, "missing eACL field in the response")
			}},
		}
		// eACL
		for _, tc := range invalidProtoContainerIDTestcases {
			eACL := proto.Clone(anyValidMinEACL).(*protoacl.EACLTable)
			eACL.ContainerId = proto.Clone(anyValidProtoContainerIDs[0]).(*protorefs.ContainerID)
			tc.corrupt(eACL.ContainerId)
			body := &protocontainer.GetExtendedACLResponse_Body{Eacl: eACL}
			tcs = append(tcs, testcase{name: "eACL/container ID/" + tc.name, body: body, assertErr: func(tb testing.TB, err error) {
				require.EqualError(t, err, "invalid eACL field in the response: invalid container ID: "+tc.msg)
			}})
		}

		testInvalidResponseBodies(t, newTestGetEACLServer, newTestContainerClient, tcs, func(c *Client) error {
			_, err := c.ContainerEACL(ctx, anyID, anyValidOpts)
			return err
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestGetEACLServer, newDefaultContainerService, func(c *Client) error {
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
	c := newClient(t)
	ctx := context.Background()
	var anyValidOpts PrmContainerSetEACL
	anyValidSigner := usertest.User().RFC6979
	// TODO: use eacltest.Table() after https://github.com/nspcc-dev/neofs-sdk-go/issues/606
	anyValidEACL := eacl.NewTableForContainer(cidtest.ID(), []eacl.Record{
		eacl.ConstructRecord(eacl.ActionDeny, eacl.OperationPut,
			[]eacl.Target{
				eacl.NewTargetByRole(eacl.RoleOthers),
				eacl.NewTargetByAccounts(usertest.IDs(3)),
			},
			eacl.NewFilterObjectOwnerEquals(usertest.ID()),
			eacl.NewObjectPropertyFilter("attr1", eacl.MatchStringEqual, "val1"),
		),
	})

	t.Run("invalid user input", func(t *testing.T) {
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
		t.Run("wrong scheme", func(t *testing.T) {
			err := c.ContainerSetEACL(ctx, anyValidEACL, user.NewAutoIDSigner(neofscryptotest.ECDSAPrivateKey()), anyValidOpts)
			// TODO: consider returning 'calculate eACL signature' to distinguish from the request signatures
			// FIXME: currently unchecked and request attempt is done. Better to pre-check like Put does
			t.Skip()
			require.EqualError(t, err, "calculate signature: incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
		})
		t.Run("signer failure", func(t *testing.T) {
			err := c.ContainerSetEACL(ctx, anyValidEACL, usertest.FailSigner(usertest.User()), anyValidOpts)
			require.ErrorContains(t, err, "calculate signature")
		})
	})
	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			signer := usertest.User()
			srv := newTestSetEACLServer()
			srv.checkRequestEACL(anyValidEACL)
			srv.checkRequestDataSignerKey(signer.ECDSAPrivateKey)
			c := newTestContainerClient(t, srv)

			err := c.ContainerSetEACL(ctx, anyValidEACL, signer.RFC6979, PrmContainerSetEACL{})
			require.NoError(t, err)
		})
		t.Run("options", func(t *testing.T) {
			t.Run("precalculated container signature", func(t *testing.T) {
				var sig neofscrypto.Signature
				sig.SetPublicKeyBytes([]byte("any public key"))
				sig.SetValue([]byte("any value"))
				srv := newTestSetEACLServer()
				srv.checkRequestDataSignature(sig)
				c := newTestContainerClient(t, srv)

				opts := anyValidOpts
				opts.AttachSignature(sig)
				err := c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, opts)
				require.NoError(t, err)
			})
			t.Run("X-headers", func(t *testing.T) {
				testRequestXHeaders(t, newTestSetEACLServer, newTestContainerClient, func(c *Client, xhs []string) error {
					opts := anyValidOpts
					opts.WithXHeaders(xhs...)
					return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, opts)
				})
			})
			t.Run("session token", func(t *testing.T) {
				st := sessiontest.ContainerSigned(usertest.User())
				srv := newTestSetEACLServer()
				srv.checkRequestSessionToken(st)
				c := newTestContainerClient(t, srv)

				opts := anyValidOpts
				opts.WithinSession(st)
				err := c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, opts)
				require.NoError(t, err)
			})
		})
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestSetEACLServer, newTestContainerClient, func(c *Client) error {
				return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "SetExtendedACL", func(c *Client) error {
			return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestSetEACLServer, newTestContainerClient, func(c *Client) error {
			return c.ContainerSetEACL(ctx, anyValidEACL, anyValidSigner, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestSetEACLServer, newDefaultContainerService, func(c *Client) error {
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
	c := newClient(t)
	ctx := context.Background()
	var anyValidOpts PrmAnnounceSpace
	anyValidAnnouncements := []container.SizeEstimation{containertest.SizeEstimation(), containertest.SizeEstimation()}

	t.Run("invalid user input", func(t *testing.T) {
		t.Run("missing announcements", func(t *testing.T) {
			err := c.ContainerAnnounceUsedSpace(ctx, nil, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingAnnouncements)
			err = c.ContainerAnnounceUsedSpace(ctx, []container.SizeEstimation{}, anyValidOpts)
			require.ErrorIs(t, err, ErrMissingAnnouncements)
		})
	})
	t.Run("exact in-out", func(t *testing.T) {
		/*
			This test is dedicated for cases when user input results in sending a certain
			request to the server and receiving a specific response to it. For user input
			errors, transport, client internals, etc. see/add other tests.
		*/
		t.Run("default", func(t *testing.T) {
			srv := newTestAnnounceContainerSpaceServer()
			srv.checkRequestAnnouncements(anyValidAnnouncements)
			c := newTestContainerClient(t, srv)

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
		t.Run("statuses", func(t *testing.T) {
			testStatusResponses(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(c *Client) error {
				return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
			})
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
	t.Run("response message decoding failure", func(t *testing.T) {
		testUnaryRPCResponseTypeMismatch(t, "container.ContainerService", "AnnounceUsedSpace", func(c *Client) error {
			return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
		})
	})
	t.Run("invalid response verification header", func(t *testing.T) {
		testInvalidResponseSignatures(t, newTestAnnounceContainerSpaceServer, newTestContainerClient, func(c *Client) error {
			return c.ContainerAnnounceUsedSpace(ctx, anyValidAnnouncements, anyValidOpts)
		})
	})
	t.Run("response callback", func(t *testing.T) {
		testResponseCallback(t, newTestAnnounceContainerSpaceServer, newDefaultContainerService, func(c *Client) error {
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
