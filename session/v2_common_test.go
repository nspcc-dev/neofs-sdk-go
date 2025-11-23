package session_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/stretchr/testify/require"
)

const (
	anyValidV2Version       = 2
	anyValidV2NNS           = "alice.neo"
	anyValidV2ContainerVerb = 8
	anyValidV2ObjectVerb    = 1
)

var validV2Proto = func() *protosession.SessionTokenV2 {
	mobjs := make([]*refs.ObjectID, len(anyValidObjectIDs))
	for i := range anyValidObjectIDs {
		mobjs[i] = &refs.ObjectID{Value: anyValidObjectIDs[i][:]}
	}
	return &protosession.SessionTokenV2{
		Body: &protosession.SessionTokenV2_Body{
			Version: anyValidV2Version,
			Id:      anyValidSessionID[:],
			Issuer: &protosession.Target{
				Identifier: &protosession.Target_OwnerId{OwnerId: &refs.OwnerID{Value: anyValidUserID[:]}},
			},
			Subjects: []*protosession.Target{{
				Identifier: &protosession.Target_NnsName{NnsName: anyValidV2NNS},
			}},
			Lifetime: &protosession.TokenLifetime{Exp: anyValidExp, Nbf: anyValidNbf, Iat: anyValidIat},
			Contexts: []*protosession.SessionContextV2{
				{
					Container: &refs.ContainerID{Value: anyValidContainerID[:]},
					Verbs:     []protosession.Verb{anyValidV2ContainerVerb},
				},
				{
					Container: &refs.ContainerID{Value: anyValidContainerID[:]},
					Verbs:     []protosession.Verb{anyValidV2ObjectVerb},
					Objects:   mobjs,
				},
			},
		},
		DelegationChain: []*protosession.DelegationInfo{
			{
				Issuer: &protosession.Target{
					Identifier: &protosession.Target_OwnerId{OwnerId: &refs.OwnerID{Value: anyValidUserID[:]}},
				},
				Subjects: []*protosession.Target{{
					Identifier: &protosession.Target_NnsName{NnsName: anyValidV2NNS},
				}},
				Lifetime: &protosession.TokenLifetime{Exp: anyValidExp, Nbf: anyValidNbf, Iat: anyValidIat},
				Verbs:    []protosession.Verb{anyValidV2ObjectVerb},
				Signature: &refs.Signature{
					Key:    anyValidIssuerPublicKeyBytes,
					Sign:   anyValidSignatureBytes,
					Scheme: anyValidSignatureScheme,
				},
			},
		},
		Signature: &refs.Signature{
			Key:    anyValidIssuerPublicKeyBytes,
			Sign:   anyValidSignatureBytes,
			Scheme: anyValidSignatureScheme,
		},
	}
}()

type invalidProtoV2TokenTestcase struct {
	name, err string
	corrupt   func(*protosession.SessionTokenV2)
}

var invalidProtoV2TokenCommonTestcases = []invalidProtoV2TokenTestcase{
	{"missing body", "missing token body", func(m *protosession.SessionTokenV2) {
		m.Body = nil
	}},
	{"body/ID/nil", "missing session ID", func(m *protosession.SessionTokenV2) {
		m.Body.Id = nil
	}},
	{"body/ID/empty", "missing session ID", func(st *protosession.SessionTokenV2) {
		st.Body.Id = []byte{}
	}},
	{"body/ID/undersize", "invalid session ID: invalid UUID (got 15 bytes)", func(st *protosession.SessionTokenV2) {
		st.Body.Id = make([]byte, 15)
	}},
	{"body/ID/wrong UUID version", "invalid session ID: wrong UUID version 3, expected 4", func(m *protosession.SessionTokenV2) {
		m.Body.Id[6] = 3 << 4
	}},
	{"body/ID/oversize", "invalid session ID: invalid UUID (got 17 bytes)", func(st *protosession.SessionTokenV2) {
		st.Body.Id = make([]byte, 17)
	}},
	{"body/issuer", "missing issuer", func(m *protosession.SessionTokenV2) {
		m.Body.Issuer = nil
	}},
	{"body/issuer/Identifier/nil", "invalid issuer: unknown target identifier type: <nil>", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier = nil
	}},
	{"body/issuer/Identifier/Target_OwnerId/nil", "invalid issuer: nil owner ID in target", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier = &protosession.Target_OwnerId{}
	}},
	{"body/issuer/Identifier/Target_OwnerId/value/empty", "invalid issuer: invalid length 0, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier = &protosession.Target_OwnerId{OwnerId: &refs.OwnerID{Value: []byte{}}}
	}},
	{"body/issuer/Identifier/Target_OwnerId/value/undersize", "invalid issuer: invalid length 24, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier = &protosession.Target_OwnerId{OwnerId: &refs.OwnerID{Value: make([]byte, 24)}}
	}},
	{"body/issuer/Identifier/Target_OwnerId/value/oversize", "invalid issuer: invalid length 26, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier = &protosession.Target_OwnerId{OwnerId: &refs.OwnerID{Value: make([]byte, 26)}}
	}},
	{"body/issuer/Identifier/Target_OwnerId/value/wrong prefix", "invalid issuer: invalid prefix byte 0x42, expected 0x35", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier.(*protosession.Target_OwnerId).OwnerId.Value[0] = 0x42
	}},
	{"body/issuer/Identifier/Target_OwnerId/value/checksum mismatch", "invalid issuer: checksum mismatch", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Identifier.(*protosession.Target_OwnerId).OwnerId.Value[24]++
	}},
	{"body/issuer/Identifier/Target_NNS/empty", "invalid issuer: empty NNS name in target", func(m *protosession.SessionTokenV2) {
		m.Body.Issuer.Identifier = &protosession.Target_NnsName{NnsName: ""}
	}},
	{"body/missing lifetime", "missing token lifetime", func(m *protosession.SessionTokenV2) {
		m.Body.Lifetime = nil
	}},
	{"body/subjects/nil", "missing subjects", func(m *protosession.SessionTokenV2) {
		m.Body.Subjects = nil
	}},
	{"body/subjects/empty", "missing subjects", func(m *protosession.SessionTokenV2) {
		m.Body.Subjects = []*protosession.Target{}
	}},
	{"body/subjects/entry nil", "invalid subject at index 0: nil target", func(m *protosession.SessionTokenV2) {
		m.Body.Subjects = []*protosession.Target{nil}
	}},
	{"body/subjects/Identifier/Target_OwnerId/value/empty", "invalid subject at index 0: invalid length 0, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Subjects[0].Identifier = &protosession.Target_OwnerId{OwnerId: &refs.OwnerID{Value: []byte{}}}
	}},
	{"body/subjects/Identifier/Target_NNS/empty", "invalid subject at index 0: empty NNS name in target", func(m *protosession.SessionTokenV2) {
		m.Body.Subjects[0].Identifier.(*protosession.Target_NnsName).NnsName = ""
	}},
	{"body/contexts", "missing contexts", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts = nil
	}},
	{"body/contexts/empty", "missing contexts", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts = []*protosession.SessionContextV2{}
	}},
	{"body/contexts/entry nil", "invalid context at index 0: nil context", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts = []*protosession.SessionContextV2{nil}
	}},
	{"body/contexts/object nil", "invalid context at index 0: nil object at index 0", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts = []*protosession.SessionContextV2{{Objects: []*refs.ObjectID{nil}}}
	}},
	{"body/contexts/invalid verb", "invalid context at index 0: negative verb -1", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[0].Verbs[0] = -1
	}},
	{"body/contexts/neither container nor wildcard", "invalid context at index 0: empty context", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[0].Reset()
	}},
	{"body/contexts/container/nil value", "invalid context at index 0: invalid container: invalid length 0", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[0].Container.Value = nil
	}},
	{"body/contexts/container/empty value", "invalid context at index 0: invalid container: invalid length 0", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[0].Container.Value = []byte{}
	}},
	{"body/contexts/container/undersize", "invalid context at index 0: invalid container: invalid length 31", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[0].Container.Value = make([]byte, 31)
	}},
	{"body/contexts/container/oversize", "invalid context at index 0: invalid container: invalid length 33", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[0].Container.Value = make([]byte, 33)
	}},
	{"body/contexts/objects/nil", "invalid context at index 1: invalid object at index 0: invalid length 0", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[1].Objects[0].Value = nil
	}},
	{"body/contexts/objects/entry nil", "invalid context at index 1: invalid object at index 0: invalid length 0", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts[1].Objects[0].Value = []byte{}
	}},
	{"delegation/nil", "invalid delegation at index 0: nil delegation info", func(m *protosession.SessionTokenV2) {
		m.DelegationChain = []*protosession.DelegationInfo{nil}
	}},
	{"delegation/invalid verb", "invalid delegation at index 0: negative verb -1", func(m *protosession.SessionTokenV2) {
		m.DelegationChain[0].Verbs[0] = -1
	}},
	{"delegation/issuer nil", "invalid delegation at index 0: invalid issuer: nil target", func(m *protosession.SessionTokenV2) {
		m.DelegationChain = []*protosession.DelegationInfo{{
			Issuer:   nil,
			Subjects: []*protosession.Target{{Identifier: &protosession.Target_NnsName{NnsName: "sub.neo"}}},
		}}
	}},
	{"delegation/subject nil", "invalid delegation at index 0: invalid subject: nil target", func(m *protosession.SessionTokenV2) {
		m.DelegationChain = []*protosession.DelegationInfo{{
			Issuer:   &protosession.Target{Identifier: &protosession.Target_NnsName{NnsName: "iss.neo"}},
			Subjects: []*protosession.Target{nil},
		}}
	}},
	{"delegation/signature negative", "invalid delegation at index 0: invalid signature: negative scheme -1", func(m *protosession.SessionTokenV2) {
		m.DelegationChain = []*protosession.DelegationInfo{{
			Issuer:    &protosession.Target{Identifier: &protosession.Target_NnsName{NnsName: "iss.neo"}},
			Subjects:  []*protosession.Target{{Identifier: &protosession.Target_NnsName{NnsName: "sub.neo"}}},
			Signature: &refs.Signature{Scheme: -1},
		}}
	}},
	{"missing signature", "missing body signature", func(m *protosession.SessionTokenV2) {
		m.Signature = nil
	}},
	{"signature/scheme/negative", "invalid body signature: negative scheme -1", func(m *protosession.SessionTokenV2) {
		m.Signature = &refs.Signature{Scheme: -1}
	}},
}

// validV2Token is constructed by setting values to be identical to validV2Proto.
var validV2Token = func() session.TokenV2 {
	var tok session.TokenV2
	tok.SetVersion(anyValidV2Version)
	tok.SetID(anyValidSessionID)
	tok.SetIssuer(session.NewTarget(anyValidUserID))
	tok.AddSubject(session.NewTargetFromNNS(anyValidV2NNS))
	tok.SetIat(anyValidIat)
	tok.SetNbf(anyValidNbf)
	tok.SetExp(anyValidExp)

	ctx1 := session.NewContextV2(anyValidContainerID, []session.VerbV2{anyValidV2ContainerVerb})
	tok.AddContext(ctx1)

	ctx2 := session.NewContextV2(anyValidContainerID, []session.VerbV2{anyValidV2ObjectVerb})
	ctx2.SetObjects(anyValidObjectIDs)
	tok.AddContext(ctx2)

	del := session.NewDelegationInfo(
		[]session.Target{session.NewTargetFromNNS(anyValidV2NNS)},
		session.NewLifetime(anyValidIat, anyValidNbf, anyValidExp),
		[]session.VerbV2{anyValidV2ObjectVerb},
	)
	del.SetIssuer(session.NewTarget(anyValidUserID))
	del.AttachSignature(anyValidSignature)
	tok.AddDelegation(del)

	// Attach signature to token
	tok.AttachSignature(anyValidSignature)

	return tok
}()

// validBinV2Token is the binary encoding of validV2Token.
// This should match the output of validV2Token.Marshal().
var validBinV2Token = []byte{0xa, 0x93, 0x2, 0x8, 0x2, 0x12, 0x10, 0x63, 0x18, 0x6f, 0x46, 0x16, 0xac, 0x48,
	0x14, 0x8b, 0xbb, 0xaf, 0x62, 0xa, 0xff, 0xe7, 0xbc, 0x1a, 0x1d, 0xa, 0x1b, 0xa, 0x19, 0x35, 0x33, 0x5,
	0xa6, 0x6f, 0x1d, 0x14, 0x65, 0xc0, 0xa5, 0x1c, 0xa7, 0x39, 0xa0, 0x52, 0x50, 0x29, 0xcb, 0x14, 0xfe,
	0x1e, 0x8a, 0xc3, 0x11, 0x5c, 0x22, 0xb, 0x12, 0x9, 0x61, 0x6c, 0x69, 0x63, 0x65, 0x2e, 0x6e, 0x65, 0x6f,
	0x2a, 0x12, 0x8, 0xee, 0xd7, 0xa4, 0xf, 0x10, 0xb7, 0xbd, 0x97, 0xcc, 0xdd, 0x2, 0x18, 0xbe, 0x84, 0xd9,
	0xc0, 0x4, 0x32, 0x27, 0xa, 0x22, 0xa, 0x20, 0xf3, 0xf5, 0x4b, 0xc6, 0x30, 0x6b, 0x8d, 0x79, 0xff, 0x31,
	0x33, 0xa8, 0x15, 0xfe, 0x3e, 0x42, 0x6, 0x93, 0x2b, 0x23, 0x63, 0xf2, 0xa3, 0x14, 0x1a, 0x1e, 0x93, 0xf0,
	0x4f, 0x72, 0xfc, 0xe3, 0x1a, 0x1, 0x8, 0x32, 0x93, 0x1, 0xa, 0x22, 0xa, 0x20, 0xf3, 0xf5, 0x4b, 0xc6, 0x30,
	0x6b, 0x8d, 0x79, 0xff, 0x31, 0x33, 0xa8, 0x15, 0xfe, 0x3e, 0x42, 0x6, 0x93, 0x2b, 0x23, 0x63, 0xf2, 0xa3,
	0x14, 0x1a, 0x1e, 0x93, 0xf0, 0x4f, 0x72, 0xfc, 0xe3, 0x12, 0x22, 0xa, 0x20, 0xf3, 0xf5, 0x4b, 0xc6, 0x30,
	0x6b, 0x8d, 0x79, 0xff, 0x31, 0x33, 0xa8, 0x15, 0xfe, 0x3e, 0x42, 0x6, 0x93, 0x2b, 0x23, 0x63, 0xf2, 0xa3,
	0x14, 0x1a, 0x1e, 0x93, 0xf0, 0x4f, 0x72, 0xfc, 0xe3, 0x12, 0x22, 0xa, 0x20, 0x2f, 0xf0, 0x5d, 0xd8, 0x9,
	0x40, 0x58, 0xb7, 0xc6, 0x24, 0x1e, 0x53, 0x14, 0xe9, 0x77, 0xfc, 0x60, 0xab, 0x6, 0x7a, 0x73, 0xa8, 0xba,
	0x93, 0xf9, 0x58, 0xb8, 0x45, 0x91, 0xc4, 0x7f, 0x44, 0x12, 0x22, 0xa, 0x20, 0x3b, 0x5, 0x78, 0xbf, 0xfa,
	0x3d, 0xf8, 0x72, 0x89, 0x15, 0xe5, 0x58, 0x39, 0x31, 0x5f, 0x9d, 0xda, 0x4f, 0x50, 0xb1, 0xd9, 0x38, 0x1d,
	0x1d, 0xaf, 0x25, 0x2a, 0xa5, 0x3a, 0x7e, 0xa1, 0xdd, 0x1a, 0x1, 0x1, 0x12, 0x38, 0xa, 0x21, 0x3, 0xca, 0xd9,
	0x8e, 0x62, 0xd1, 0xbe, 0xbc, 0x91, 0x7b, 0xae, 0x15, 0xad, 0xef, 0xef, 0xf5, 0x43, 0x94, 0xcd, 0x77, 0x3a,
	0xdf, 0xdb, 0xd1, 0xdc, 0x71, 0xd7, 0x86, 0xe4, 0x65, 0xf9, 0x22, 0xda, 0x12, 0xd, 0x61, 0x6e, 0x79, 0x5f,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0xec, 0xec, 0xaf, 0xc0, 0x4, 0x1a, 0x7d, 0xa,
	0x1d, 0xa, 0x1b, 0xa, 0x19, 0x35, 0x33, 0x5, 0xa6, 0x6f, 0x1d, 0x14, 0x65, 0xc0, 0xa5, 0x1c, 0xa7, 0x39,
	0xa0, 0x52, 0x50, 0x29, 0xcb, 0x14, 0xfe, 0x1e, 0x8a, 0xc3, 0x11, 0x5c, 0x12, 0xb, 0x12, 0x9, 0x61, 0x6c,
	0x69, 0x63, 0x65, 0x2e, 0x6e, 0x65, 0x6f, 0x1a, 0x12, 0x8, 0xee, 0xd7, 0xa4, 0xf, 0x10, 0xb7, 0xbd, 0x97,
	0xcc, 0xdd, 0x2, 0x18, 0xbe, 0x84, 0xd9, 0xc0, 0x4, 0x22, 0x1, 0x1, 0x2a, 0x38, 0xa, 0x21, 0x3, 0xca, 0xd9,
	0x8e, 0x62, 0xd1, 0xbe, 0xbc, 0x91, 0x7b, 0xae, 0x15, 0xad, 0xef, 0xef, 0xf5, 0x43, 0x94, 0xcd, 0x77, 0x3a,
	0xdf, 0xdb, 0xd1, 0xdc, 0x71, 0xd7, 0x86, 0xe4, 0x65, 0xf9, 0x22, 0xda, 0x12, 0xd, 0x61, 0x6e, 0x79, 0x5f,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0xec, 0xec, 0xaf, 0xc0, 0x4}

// validJSONV2Token is the JSON encoding of validV2Token.
// This should match the output of validV2Token.MarshalJSON().
var validJSONV2Token = `
{"body":{
"version":2, 
"id":"YxhvRhasSBSLu69iCv/nvA==", 
"issuer":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}}, 
"subjects":[{"nnsName":"alice.neo"}], 
"lifetime":{"exp":"32058350", "nbf":"93843742391", "iat":"1209418302"}, 
"contexts":[{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, "objects":[], "verbs":["CONTAINER_PUT"]}, {"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, "objects":[{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, {"value":"L/Bd2AlAWLfGJB5TFOl3/GCrBnpzqLqT+Vi4RZHEf0Q="}, {"value":"OwV4v/o9+HKJFeVYOTFfndpPULHZOB0dryUqpTp+od0="}], "verbs":["OBJECT_PUT"]}]}, 
"delegationChain":[{"issuer":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}}, "subjects":[{"nnsName":"alice.neo"}], "lifetime":{"exp":"32058350", "nbf":"93843742391", "iat":"1209418302"}, "verbs":["OBJECT_PUT"], "signature":{"key":"A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa", "signature":"YW55X3NpZ25hdHVyZQ==", "scheme":1208743532}}], 
"signature":{"key":"A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa", "signature":"YW55X3NpZ25hdHVyZQ==", "scheme":1208743532}}
`

func checkTokenFields(t *testing.T, val session.TokenV2) {
	require.EqualValues(t, anyValidV2Version, val.Version())
	require.Equal(t, val.ID(), anyValidSessionID)
	require.Equal(t, val.Issuer().OwnerID(), anyValidUserID)
	require.Len(t, val.Subjects(), 1)
	require.Equal(t, val.Subjects()[0].NNSName(), anyValidV2NNS)
	require.EqualValues(t, anyValidExp, val.Exp())
	require.EqualValues(t, anyValidIat, val.Iat())
	require.EqualValues(t, anyValidNbf, val.Nbf())
	require.True(t, val.AssertVerb(anyValidV2ContainerVerb, anyValidContainerID))
	for i := range anyValidObjectIDs {
		require.True(t, val.AssertObject(anyValidV2ObjectVerb, anyValidContainerID, anyValidObjectIDs[i]))
	}
	require.Equal(t, val.DelegationChain()[0].Issuer().OwnerID(), anyValidUserID)
	require.Len(t, val.DelegationChain()[0].Subjects(), 1)
	require.Equal(t, val.DelegationChain()[0].Subjects()[0].NNSName(), anyValidV2NNS)
	val.DelegationChain()[0].Verbs()
	sig, ok := val.Signature()
	require.True(t, ok)
	require.EqualValues(t, anyValidSignatureScheme, sig.Scheme())
	require.Equal(t, anyValidIssuerPublicKeyBytes, sig.PublicKeyBytes())
	require.Equal(t, anyValidSignatureBytes, sig.Value())
}

type invalidBinV2TokenTestcase struct {
	name, err string
	b         []byte
}

var invalidBinV2CommonTestcases = []invalidBinV2TokenTestcase{
	{name: "body/ID/undersize", err: "invalid session ID: invalid UUID (got 15 bytes)",
		b: []byte{10, 17, 18, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0}},
	{name: "body/ID/oversize", err: "invalid session ID: invalid UUID (got 17 bytes)",
		b: []byte{10, 19, 18, 17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0}},
	{name: "body/ID/wrong UUID version", err: "invalid session ID: wrong UUID version 3, expected 4",
		b: []byte{10, 18, 18, 16, 0, 0, 0, 0, 0, 0, 48, 0, 0, 0, 0, 0,
			0, 0, 0, 0}},
	{name: "body/issuer/value/empty", err: "invalid issuer: invalid length 0, expected 25",
		b: []byte{10, 4, 26, 2, 10, 0}},
	{name: "body/issuer/value/undersize", err: "invalid issuer: invalid length 24, expected 25",
		b: []byte{10, 30, 26, 28, 10, 26, 10, 24, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "body/issuer/value/oversize", err: "invalid issuer: invalid length 26, expected 25",
		b: []byte{10, 32, 26, 30, 10, 28, 10, 26, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0}},
	{name: "body/issuer/value/wrong prefix", err: "invalid issuer: invalid prefix byte 0x42, expected 0x35",
		b: []byte{10, 31, 26, 29, 10, 27, 10, 25, 66, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0}},
	{name: "body/issuer/value/checksum mismatch", err: "invalid issuer: checksum mismatch",
		b: []byte{10, 31, 26, 29, 10, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101,
			192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17,
			93}},
	{name: "signature/invalid scheme", err: "invalid body signature: negative scheme -2147483648",
		b: []byte{18, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
	{name: "context/invalid verb", err: "invalid context at index 0: negative verb -1",
		b: []byte{10, 14, 50, 12, 26, 10, 255, 255, 255, 255, 255, 255, 255, 255, 255, 1}},
	{name: "context/missing container", err: "invalid context at index 0: empty context",
		b: []byte{10, 2, 50, 0}},
	{name: "context/container/nil value", err: "invalid context at index 0: invalid container: invalid length 0",
		b: []byte{10, 4, 50, 2, 10, 0}},
	{name: "context/container/empty value", err: "invalid context at index 0: invalid container: invalid length 0",
		b: []byte{10, 4, 50, 2, 10, 0}},
	{name: "context/container/undersize", err: "invalid context at index 0: invalid container: invalid length 31",
		b: []byte{10, 37, 50, 35, 10, 33, 10, 31, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0}},
	{name: "context/container/oversize", err: "invalid context at index 0: invalid container: invalid length 33",
		b: []byte{10, 39, 50, 37, 10, 35, 10, 33, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "context/objects/nil element", err: "invalid context at index 0: invalid object at index 0: invalid length 0",
		b: []byte{10, 4, 50, 2, 18, 0}},
	{name: "context/objects/nil value", err: "invalid context at index 0: invalid object at index 0: invalid length 0",
		b: []byte{10, 4, 50, 2, 18, 0}},
	{name: "context/objects/empty value", err: "invalid context at index 0: invalid object at index 0: invalid length 0",
		b: []byte{10, 4, 50, 2, 18, 0}},
	{name: "context/objects/undersize", err: "invalid context at index 0: invalid object at index 0: invalid length 31",
		b: []byte{10, 37, 50, 35, 18, 33, 10, 31, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0}},
	{name: "context/objects/oversize", err: "invalid context at index 0: invalid object at index 0: invalid length 33",
		b: []byte{10, 39, 50, 37, 18, 35, 10, 33, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0}},
}

type invalidJSONV2TokenTestcase struct {
	name, err string
	j         string
}

var invalidJSONV2CommonTestcases = []invalidJSONV2TokenTestcase{
	{
		name: "body/ID/undersize",
		err:  "invalid session ID: invalid UUID (got 15 bytes)",
		j:    `{"body":{"id":"YxhvRhasSBSLu69iCv/n"}}`,
	},
	{
		name: "body/ID/oversize",
		err:  "invalid session ID: invalid UUID (got 17 bytes)",
		j:    `{"body":{"id":"YxhvRhasSBSLu69iCv/nvAE="}}`,
	},
	{
		name: "body/ID/wrong UUID version",
		err:  "invalid session ID: wrong UUID version 3, expected 4",
		j:    `{"body":{"id":"YxhvRhasMBSLu69iCv/nvA=="}}`,
	},
	{
		name: "body/issuer/value/empty",
		err:  "invalid issuer: invalid length 0, expected 25",
		j:    `{"body":{"issuer":{"ownerID":{"value":""}}}}`,
	},
	{
		name: "body/issuer/value/undersize",
		err:  "invalid issuer: invalid length 24, expected 25",
		j:    `{"body":{"issuer":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMR"}}}}`,
	},
	{
		name: "body/issuer/value/oversize",
		err:  "invalid issuer: invalid length 26, expected 25",
		j:    `{"body":{"issuer":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXAE="}}}}`,
	},
	{
		name: "body/issuer/value/wrong prefix",
		err:  "invalid issuer: invalid prefix byte 0x42, expected 0x35",
		j:    `{"body":{"issuer":{"ownerID":{"value":"QjMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}}}}`,
	},
	{
		name: "body/issuer/value/checksum mismatch",
		err:  "invalid issuer: checksum mismatch",
		j:    `{"body":{"issuer":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXQ=="}}}}`,
	},
	{
		name: "signature/invalid scheme",
		err:  "invalid body signature: negative scheme -2147483648",
		j:    `{"signature":{"scheme":-2147483648}}`,
	},
	{
		name: "context/invalid verb",
		err:  "invalid context at index 0: negative verb -1",
		j:    `{"body":{"contexts":[{"verbs":[-1]}]}}`,
	},
	{
		name: "context/missing container",
		err:  "invalid context at index 0: empty context",
		j:    `{"body":{"contexts":[{}]}}`,
	},
	{
		name: "context/container/nil value",
		err:  "invalid context at index 0: invalid container: invalid length 0",
		j:    `{"body":{"contexts":[{"container":{"value":""}}]}}`,
	},
	{
		name: "context/container/empty value",
		err:  "invalid context at index 0: invalid container: invalid length 0",
		j:    `{"body":{"contexts":[{"container":{"value":""}}]}}`,
	},
	{
		name: "context/container/undersize",
		err:  "invalid context at index 0: invalid container: invalid length 31",
		j:    `{"body":{"contexts":[{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/A=="}}]}}`,
	},
	{
		name: "context/container/oversize",
		err:  "invalid context at index 0: invalid container: invalid length 33",
		j:    `{"body":{"contexts":[{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OMB"}}]}}`,
	},
	{
		name: "context/objects/nil element",
		err:  "invalid context at index 0: invalid object at index 0: invalid length 0",
		j:    `{"body":{"contexts":[{"objects":[{}]}]}}`,
	},
	{
		name: "context/objects/nil value",
		err:  "invalid context at index 0: invalid object at index 0: invalid length 0",
		j:    `{"body":{"contexts":[{"objects":[{"value":""}]}]}}`,
	},
	{
		name: "context/objects/empty value",
		err:  "invalid context at index 0: invalid object at index 0: invalid length 0",
		j:    `{"body":{"contexts":[{"objects":[{"value":""}]}]}}`,
	},
	{
		name: "context/objects/undersize",
		err:  "invalid context at index 0: invalid object at index 0: invalid length 31",
		j:    `{"body":{"contexts":[{"objects":[{"value":"L/Bd2AlAWLfGJB5TFOl3/GCrBnpzqLqT+Vi4RZHEfw=="}]}]}}`,
	},
	{
		name: "context/objects/oversize",
		err:  "invalid context at index 0: invalid object at index 0: invalid length 33",
		j:    `{"body":{"contexts":[{"objects":[{"value":"L/Bd2AlAWLfGJB5TFOl3/GCrBnpzqLqT+Vi4RZHEf0QB"}]}]}}`,
	},
}
