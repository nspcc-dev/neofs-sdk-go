package client

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type (
	invalidObjectSplitInfoProtoTestcase = struct {
		name, msg string
		corrupt   func(valid *protoobject.SplitInfo)
	}
	invalidObjectHeaderProtoTestcase = struct {
		name, msg string
		corrupt   func(valid *protoobject.Header)
	}
)

// various sets of Object service testcases.
var (
	invalidObjectSplitInfoProtoTestcases = []invalidObjectSplitInfoProtoTestcase{
		{name: "neither linker nor last", msg: "neither link object ID nor last part object ID is set", corrupt: func(valid *protoobject.SplitInfo) {
			valid.Reset()
		}},
		// + other cases in init
	}
	invalidObjectSessionTokenProtoTestcases = append(invalidCommonSessionTokenProtoTestcases, invalidSessionTokenProtoTestcase{
		name: "context/wrong", msg: "invalid context: invalid context *session.SessionToken_Body_Container",
		corrupt: func(valid *protosession.SessionToken) {
			valid.Body.Context = new(protosession.SessionToken_Body_Container)
		}},
		invalidSessionTokenProtoTestcase{
			name: "context/verb/negative", msg: "invalid context: negative verb -1",
			corrupt: func(valid *protosession.SessionToken) {
				c := valid.Body.Context.(*protosession.SessionToken_Body_Object).Object
				c.Verb = -1
			},
		},
		invalidSessionTokenProtoTestcase{
			name: "context/container/nil", msg: "invalid context: missing target container",
			corrupt: func(valid *protosession.SessionToken) {
				c := valid.Body.Context.(*protosession.SessionToken_Body_Object).Object
				c.Target.Container = nil
			},
		}) // + other container and object ID cases in init
	invalidObjectHeaderProtoTestcases = []invalidObjectHeaderProtoTestcase{
		// 1. version (any accepted, even absent)
		// 2. container (init)
		// 3. owner (init)
		// 4. creation epoch (any accepted)
		// 5. payload length (any accepted)
		// 6. payload checksum (init)
		{name: "type/negative", msg: "negative type -1", corrupt: func(valid *protoobject.Header) {
			valid.ObjectType = -1
		}},
		// 8. homomorphic payload checksum (init)
		// 9. session token (init)
		{name: "attributes/no key", msg: "invalid attribute #1: missing key",
			corrupt: func(valid *protoobject.Header) {
				valid.Attributes = []*protoobject.Header_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "", Value: "v2"}, {Key: "k3", Value: "v3"},
				}
			}},
		{name: "attributes/no value", msg: "invalid attribute #1: missing value",
			corrupt: func(valid *protoobject.Header) {
				valid.Attributes = []*protoobject.Header_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "k2", Value: ""}, {Key: "k3", Value: "v3"},
				}
			}},
		{name: "attributes/duplicated", msg: "duplicated attribute k1",
			corrupt: func(valid *protoobject.Header) {
				valid.Attributes = []*protoobject.Header_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "k2", Value: "v2"}, {Key: "k1", Value: "v3"},
				}
			}},
		{name: "attributes/expiration", msg: `invalid attribute #1: invalid expiration epoch (must be a uint): strconv.ParseUint: parsing "foo": invalid syntax`,
			corrupt: func(valid *protoobject.Header) {
				valid.Attributes = []*protoobject.Header_Attribute{
					{Key: "k1", Value: "v1"}, {Key: "__NEOFS__EXPIRATION_EPOCH", Value: "foo"}, {Key: "k3", Value: "v3"},
				}
			}},
		// 11. split (init)
	}
)

func init() {
	// session token
	for _, tc := range invalidContainerIDProtoTestcases {
		invalidObjectSessionTokenProtoTestcases = append(invalidObjectSessionTokenProtoTestcases, invalidSessionTokenProtoTestcase{
			name: "context/container/" + tc.name, msg: "invalid context: invalid container ID: " + tc.msg,
			corrupt: func(valid *protosession.SessionToken) {
				c := valid.Body.Context.(*protosession.SessionToken_Body_Object).Object
				tc.corrupt(c.Target.Container)
			},
		})
	}
	for _, tc := range invalidObjectIDProtoTestcases {
		invalidObjectSessionTokenProtoTestcases = append(invalidObjectSessionTokenProtoTestcases, invalidSessionTokenProtoTestcase{
			name: "context/objects/" + tc.name, msg: "invalid context: invalid target object: " + tc.msg,
			corrupt: func(valid *protosession.SessionToken) {
				c := valid.Body.Context.(*protosession.SessionToken_Body_Object).Object
				c.Target.Objects = []*protorefs.ObjectID{
					proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
					proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID),
					proto.Clone(validProtoObjectIDs[2]).(*protorefs.ObjectID),
				}
				tc.corrupt(c.Target.Objects[1])
			},
		})
	}
	// split info
	for _, tc := range invalidUUIDProtoTestcases {
		invalidObjectSplitInfoProtoTestcases = append(invalidObjectSplitInfoProtoTestcases, invalidObjectSplitInfoProtoTestcase{
			name: "split ID/" + tc.name, msg: "invalid split ID: " + tc.msg,
			corrupt: func(valid *protoobject.SplitInfo) { valid.SplitId = tc.corrupt(valid.SplitId) },
		})
	}
	for _, tc := range invalidObjectIDProtoTestcases {
		invalidObjectSplitInfoProtoTestcases = append(invalidObjectSplitInfoProtoTestcases, invalidObjectSplitInfoProtoTestcase{
			name: "last ID/" + tc.name, msg: "could not convert last part object ID: " + tc.msg,
			corrupt: func(valid *protoobject.SplitInfo) { tc.corrupt(valid.LastPart) },
		}, invalidObjectSplitInfoProtoTestcase{
			name: "linker/" + tc.name, msg: "could not convert link object ID: " + tc.msg,
			corrupt: func(valid *protoobject.SplitInfo) { tc.corrupt(valid.Link) },
		}, invalidObjectSplitInfoProtoTestcase{
			name: "first ID/" + tc.name, msg: "could not convert first part object ID: " + tc.msg,
			corrupt: func(valid *protoobject.SplitInfo) { tc.corrupt(valid.FirstPart) },
		})
	}
	// header
	for _, tc := range invalidContainerIDProtoTestcases {
		invalidObjectHeaderProtoTestcases = append(invalidObjectHeaderProtoTestcases, invalidObjectHeaderProtoTestcase{
			name: "container/" + tc.name, msg: "invalid container: " + tc.msg,
			corrupt: func(valid *protoobject.Header) { tc.corrupt(valid.ContainerId) },
		})
	}
	for _, tc := range invalidUserIDProtoTestcases {
		invalidObjectHeaderProtoTestcases = append(invalidObjectHeaderProtoTestcases, invalidObjectHeaderProtoTestcase{
			name: "owner/" + tc.name, msg: "invalid owner: " + tc.msg,
			corrupt: func(valid *protoobject.Header) { tc.corrupt(valid.OwnerId) },
		})
	}
	for _, tc := range invalidChecksumTestcases {
		invalidObjectHeaderProtoTestcases = append(invalidObjectHeaderProtoTestcases, invalidObjectHeaderProtoTestcase{
			name: "payload checksum/" + tc.name, msg: "invalid payload checksum: " + tc.msg,
			corrupt: func(valid *protoobject.Header) { tc.corrupt(valid.PayloadHash) },
		}, invalidObjectHeaderProtoTestcase{
			name: "payload homomorphic checksum/" + tc.name, msg: "invalid payload homomorphic checksum: " + tc.msg,
			corrupt: func(valid *protoobject.Header) { tc.corrupt(valid.HomomorphicHash) },
		})
	}
	type splitTestcase = struct {
		name, msg string
		corrupt   func(split *protoobject.Header_Split)
	}
	var splitTestcases []splitTestcase
	for _, tc := range invalidObjectHeaderProtoTestcases {
		splitTestcases = append(splitTestcases, splitTestcase{
			name: "parent header/" + tc.name, msg: "invalid parent header: " + strings.ReplaceAll(tc.msg, "invalid header: ", ""),
			corrupt: func(valid *protoobject.Header_Split) { tc.corrupt(valid.ParentHeader) },
		})
	}
	for _, tc := range invalidObjectIDProtoTestcases {
		splitTestcases = append(splitTestcases, splitTestcase{
			name: "parent ID/" + tc.name, msg: "invalid parent split member ID: " + tc.msg,
			corrupt: func(valid *protoobject.Header_Split) { tc.corrupt(valid.Parent) },
		}, splitTestcase{
			name: "previous ID/" + tc.name, msg: "invalid previous split member ID: " + tc.msg,
			corrupt: func(valid *protoobject.Header_Split) { tc.corrupt(valid.Previous) },
		}, splitTestcase{
			name: "first ID/" + tc.name, msg: "invalid first split member ID: " + tc.msg,
			corrupt: func(valid *protoobject.Header_Split) { tc.corrupt(valid.First) },
		}, splitTestcase{
			name: "children/" + tc.name, msg: "invalid child split member ID #1: " + tc.msg,
			corrupt: func(valid *protoobject.Header_Split) {
				valid.Children = []*protorefs.ObjectID{
					proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
					proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID),
					proto.Clone(validProtoObjectIDs[2]).(*protorefs.ObjectID),
				}
				tc.corrupt(valid.Children[1])
			},
		})
	}
	for _, tc := range invalidSignatureProtoTestcases {
		splitTestcases = append(splitTestcases, splitTestcase{
			name: "parent signature/" + tc.name, msg: "invalid parent signature: " + tc.msg,
			corrupt: func(valid *protoobject.Header_Split) { tc.corrupt(valid.ParentSignature) },
		})
	}
	for _, tc := range invalidUUIDProtoTestcases {
		splitTestcases = append(splitTestcases, splitTestcase{
			name: "split ID/" + tc.name, msg: "invalid split ID: " + tc.msg,
			corrupt: func(valid *protoobject.Header_Split) { valid.SplitId = tc.corrupt(valid.SplitId) },
		})
	}
	for _, tc := range splitTestcases {
		invalidObjectHeaderProtoTestcases = append(invalidObjectHeaderProtoTestcases, invalidObjectHeaderProtoTestcase{
			name:    "split header/" + tc.name,
			msg:     "invalid split header: " + tc.msg,
			corrupt: func(valid *protoobject.Header) { tc.corrupt(valid.Split) },
		})
	}
	for _, tc := range invalidObjectSessionTokenProtoTestcases {
		invalidObjectHeaderProtoTestcases = append(invalidObjectHeaderProtoTestcases, invalidObjectHeaderProtoTestcase{
			name: "session token/" + tc.name, msg: "invalid session token: " + tc.msg,
			corrupt: func(valid *protoobject.Header) { tc.corrupt(valid.SessionToken) },
		})
	}
}

// returns Client-compatible Object service handled by given server. Provided
// server must implement [protoobject.ObjectServiceServer]: the parameter is not
// of this type to support generics.
func newDefaultObjectService(t testing.TB, srv any) testService {
	require.Implements(t, (*protoobject.ObjectServiceServer)(nil), srv)
	return testService{desc: &protoobject.ObjectService_ServiceDesc, impl: srv}
}

// returns Client of Object service provided by given server. Provided server
// must implement [protoobject.ObjectServiceServer]: the parameter is not of
// this type to support generics.
func newTestObjectClient(t testing.TB, srv any) *Client {
	return newClient(t, newDefaultObjectService(t, srv))
}

func assertObjectStreamTransportErr(t testing.TB, transportErr, err error) {
	require.Error(t, err)
	require.NotContains(t, err.Error(), "open stream") // gRPC client cannot catch this
	st, ok := status.FromError(err)
	require.True(t, ok, err)
	require.Equal(t, codes.Unknown, st.Code())
	require.Contains(t, st.Message(), transportErr.Error())
}

// for sharing between servers of requests that can be for local execution only.
type testLocalRequestServerSettings struct {
	reqLocal bool
}

// makes the server to assert that any request has TTL = 1. By default, TTL must
// be 2.
func (x *testLocalRequestServerSettings) checkRequestLocal() { x.reqLocal = true }

func (x testLocalRequestServerSettings) verifyTTL(m *protosession.RequestMetaHeader) error {
	var exp uint32
	if x.reqLocal {
		exp = 1
	} else {
		exp = 2
	}
	if act := m.GetTtl(); act != exp {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("wrong TTL %d, expected %d", act, exp))
	}
	return nil
}

// for sharing between servers of requests with required object address.
type testObjectAddressServerSettings struct {
	c                testRequiredContainerIDServerSettings
	expectedReqObjID *oid.ID
}

// makes the server to assert that any request carries given object address. By
// default, any address is accepted.
func (x *testObjectAddressServerSettings) checkRequestObjectAddress(c cid.ID, o oid.ID) {
	x.c.checkRequestContainerID(c)
	x.expectedReqObjID = &o
}

func (x testObjectAddressServerSettings) verifyObjectAddress(m *protorefs.Address) error {
	if m == nil {
		return newErrMissingRequestBodyField("object address")
	}
	if err := x.c.verifyRequestContainerID(m.ContainerId); err != nil {
		return err
	}
	if m.ObjectId == nil {
		return newErrMissingRequestBodyField("object ID")
	}
	if x.expectedReqObjID != nil {
		if err := checkObjectIDTransport(*x.expectedReqObjID, m.ObjectId); err != nil {
			return newErrInvalidRequestField("container ID", err)
		}
	}
	return nil
}

// for sharing between servers of requests with an object session token.
type testObjectSessionServerSettings struct {
	expectedToken   *session.Object
	expectedTokenV2 *sessionv2.Token
}

// makes the server to assert that any request carries given session token. By
// default, session token must not be attached.
func (x *testObjectSessionServerSettings) checkRequestSessionToken(st session.Object) {
	x.expectedToken = &st
	x.expectedTokenV2 = nil
}

// makes the server to assert that any request carries given V2 session token. By
// default, session token must not be attached.
func (x *testObjectSessionServerSettings) checkRequestSessionTokenV2(st sessionv2.Token) {
	x.expectedTokenV2 = &st
	x.expectedToken = nil
}

func (x testObjectSessionServerSettings) verifySessionToken(m *protosession.SessionToken) error {
	if m == nil {
		if x.expectedToken != nil {
			return newInvalidRequestMetaHeaderErr(errors.New("session token is missing while should not be"))
		}
		return nil
	}
	if x.expectedToken == nil {
		return newInvalidRequestMetaHeaderErr(errors.New("session token attached while should not be"))
	}
	if err := checkObjectSessionTransport(*x.expectedToken, m); err != nil {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("session token: %w", err))
	}
	return nil
}

func (x testObjectSessionServerSettings) verifySessionTokenV2(m *protosession.SessionTokenV2) error {
	if m == nil {
		if x.expectedTokenV2 != nil {
			return newInvalidRequestMetaHeaderErr(errors.New("session token V2 is missing while should not be"))
		}
		return nil
	}
	if x.expectedTokenV2 == nil {
		return newInvalidRequestMetaHeaderErr(errors.New("session token V2 attached while should not be"))
	}
	if err := checkSessionTokenV2Transport(*x.expectedTokenV2, m); err != nil {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("session token V2: %w", err))
	}
	return nil
}

// for sharing between servers of requests with a bearer token.
type testBearerTokenServerSettings struct {
	expectedToken *bearer.Token
}

// makes the server to assert that any request carries given bearer token. By
// default, bearer token must not be attached.
func (x *testBearerTokenServerSettings) checkRequestBearerToken(bt bearer.Token) {
	x.expectedToken = &bt
}

func (x testBearerTokenServerSettings) verifyBearerToken(m *protoacl.BearerToken) error {
	if m == nil {
		if x.expectedToken != nil {
			return newInvalidRequestMetaHeaderErr(errors.New("bearer token is missing while should not be"))
		}
		return nil
	}
	if x.expectedToken == nil {
		return newInvalidRequestMetaHeaderErr(errors.New("bearer token attached while should not be"))
	}
	if err := checkBearerTokenTransport(*x.expectedToken, m); err != nil {
		return newInvalidRequestMetaHeaderErr(fmt.Errorf("bearer token: %w", err))
	}
	return nil
}
