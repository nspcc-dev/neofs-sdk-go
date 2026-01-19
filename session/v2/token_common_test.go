package session_test

import (
	"testing"
	"time"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

const (
	anyValidVersion         = 2
	anyValidNNS             = "alice.neo"
	anyValidContainerVerb   = 8
	anyValidObjectVerb      = 1
	anyValidNonce           = 12345
	anyValidExp             = 32058350
	anyValidIat             = 1209418302
	anyValidNbf             = 93843742391
	anyValidSignatureScheme = 1208743532
)

var (
	anyValidExpTime = time.Unix(anyValidExp, 0)
	anyValidNbfTime = time.Unix(anyValidNbf, 0)
	anyValidIatTime = time.Unix(anyValidIat, 0)
	anyValidUserID  = user.ID{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57,
		160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
	anyValidContainerID = cid.ID{243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66,
		6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227}
	anyValidIssuerPublicKeyBytes = []byte{3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239,
		245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218}
	anyValidSignatureBytes = []byte("any_signature") // valid structurally, not logically
	anyValidSignature      = neofscrypto.NewSignatureFromRawKey(anyValidSignatureScheme, anyValidIssuerPublicKeyBytes, anyValidSignatureBytes)
)

var validProto = func() *protosession.SessionTokenV2 {
	return &protosession.SessionTokenV2{
		Body: &protosession.SessionTokenV2_Body{
			Version: anyValidVersion,
			Nonce:   anyValidNonce,
			Issuer:  &refs.OwnerID{Value: anyValidUserID[:]},
			Subjects: []*protosession.Target{{
				Identifier: &protosession.Target_NnsName{NnsName: anyValidNNS},
			}},
			Lifetime: &protosession.TokenLifetime{Exp: anyValidExp, Nbf: anyValidNbf, Iat: anyValidIat},
			Contexts: []*protosession.SessionContextV2{
				{
					Container: &refs.ContainerID{Value: anyValidContainerID[:]},
					Verbs:     []protosession.Verb{anyValidContainerVerb},
				},
				{
					Container: &refs.ContainerID{Value: anyValidContainerID[:]},
					Verbs:     []protosession.Verb{anyValidObjectVerb},
				},
			},
			Final: true,
		},
		Origin: &protosession.SessionTokenV2{
			Body: &protosession.SessionTokenV2_Body{
				Issuer: &refs.OwnerID{Value: anyValidUserID[:]},
				Subjects: []*protosession.Target{{
					Identifier: &protosession.Target_NnsName{NnsName: anyValidNNS},
				}},
				Lifetime: &protosession.TokenLifetime{Exp: anyValidExp, Nbf: anyValidNbf, Iat: anyValidIat},
				Contexts: []*protosession.SessionContextV2{
					{
						Container: &refs.ContainerID{Value: anyValidContainerID[:]},
						Verbs:     []protosession.Verb{anyValidObjectVerb},
					},
				},
			},
			Signature: &refs.Signature{
				Key:    anyValidIssuerPublicKeyBytes,
				Sign:   anyValidSignatureBytes,
				Scheme: anyValidSignatureScheme,
			},
		},
		Signature: &refs.Signature{
			Key:    anyValidIssuerPublicKeyBytes,
			Sign:   anyValidSignatureBytes,
			Scheme: anyValidSignatureScheme,
		},
	}
}()

type invalidProtoTokenTestcase struct {
	name, err string
	corrupt   func(*protosession.SessionTokenV2)
}

var invalidProtoTokenCommonTestcases = []invalidProtoTokenTestcase{
	{"missing body", "missing token body", func(m *protosession.SessionTokenV2) {
		m.Body = nil
	}},
	{"body/issuer", "missing issuer", func(m *protosession.SessionTokenV2) {
		m.Body.Issuer = nil
	}},
	{"body/issuer/value/empty", "invalid issuer: invalid length 0, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer = &refs.OwnerID{Value: []byte{}}
	}},
	{"body/issuer/value/undersize", "invalid issuer: invalid length 24, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer = &refs.OwnerID{Value: make([]byte, 24)}
	}},
	{"body/issuer/value/oversize", "invalid issuer: invalid length 26, expected 25", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer = &refs.OwnerID{Value: make([]byte, 26)}
	}},
	{"body/issuer/value/wrong prefix", "invalid issuer: invalid prefix byte 0x42, expected 0x35", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Value[0] = 0x42
	}},
	{"body/issuer/value/checksum mismatch", "invalid issuer: checksum mismatch", func(st *protosession.SessionTokenV2) {
		st.Body.Issuer.Value[24]++
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
	{"body/subjects/entry nil", "nil subject at index 0", func(m *protosession.SessionTokenV2) {
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
	{"body/contexts/entry nil", "nil context at index 0", func(m *protosession.SessionTokenV2) {
		m.Body.Contexts = []*protosession.SessionContextV2{nil}
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
	{"missing signature", "missing body signature", func(m *protosession.SessionTokenV2) {
		m.Signature = nil
	}},
	{"signature/scheme/negative", "invalid body signature: negative scheme -1", func(m *protosession.SessionTokenV2) {
		m.Signature = &refs.Signature{Scheme: -1}
	}},
	{"origin/missing body", "invalid origin token: missing token body", func(m *protosession.SessionTokenV2) {
		m.Origin = &protosession.SessionTokenV2{Body: nil}
	}},
}

// validToken is constructed by setting values to be identical to validProto.
var validToken = func() session.Token {
	var tok session.Token
	tok.SetVersion(anyValidVersion)
	tok.SetNonce(anyValidNonce)
	tok.SetFinal(true)
	tok.SetIssuer(anyValidUserID)
	err := tok.AddSubject(session.NewTargetNamed(anyValidNNS))
	if err != nil {
		panic(err)
	}
	tok.SetIat(anyValidIatTime)
	tok.SetNbf(anyValidNbfTime)
	tok.SetExp(anyValidExpTime)

	ctx1, err := session.NewContext(anyValidContainerID, []session.Verb{anyValidContainerVerb})
	if err != nil {
		panic(err)
	}
	err = tok.AddContext(ctx1)
	if err != nil {
		panic(err)
	}

	ctx2, err := session.NewContext(anyValidContainerID, []session.Verb{anyValidObjectVerb})
	if err != nil {
		panic(err)
	}
	err = tok.AddContext(ctx2)
	if err != nil {
		panic(err)
	}

	// Create origin token for delegation
	origin := session.Token{}
	origin.SetIssuer(anyValidUserID)
	err = origin.AddSubject(session.NewTargetNamed(anyValidNNS))
	if err != nil {
		panic(err)
	}
	origin.SetIat(anyValidIatTime)
	origin.SetNbf(anyValidNbfTime)
	origin.SetExp(anyValidExpTime)

	ctxOrig, err := session.NewContext(anyValidContainerID, []session.Verb{anyValidObjectVerb})
	if err != nil {
		panic(err)
	}
	err = origin.AddContext(ctxOrig)
	if err != nil {
		panic(err)
	}
	origin.AttachSignature(anyValidSignature)
	tok.SetOrigin(&origin)

	// Attach signature to token
	tok.AttachSignature(anyValidSignature)

	return tok
}()

// validBinToken is the binary encoding of validToken.
// This should match the output of validToken.Marshal().
var validBinToken = []byte{10, 151, 1, 8, 2, 16, 185, 96, 26, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254,
	30, 138, 195, 17, 92, 34, 11, 18, 9, 97, 108, 105, 99, 101, 46, 110, 101, 111, 42, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2,
	24, 190, 132, 217, 192, 4, 50, 39, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35,
	99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 1, 8, 50, 39, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51,
	168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 1, 1, 56, 1, 18, 56, 10, 33, 3, 202,
	217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218, 18,
	13, 97, 110, 121, 95, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 236, 236, 175, 192, 4, 26, 163, 1, 10, 103, 26, 27, 10, 25, 53, 51, 5,
	166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92, 34, 11, 18, 9, 97, 108, 105, 99, 101, 46,
	110, 101, 111, 42, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 50, 39, 10, 34, 10, 32, 243, 245, 75,
	198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 1, 1,
	18, 56, 10, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134,
	228, 101, 249, 34, 218, 18, 13, 97, 110, 121, 95, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 236, 236, 175, 192, 4}

// validJSONToken is the JSON encoding of validToken.
// This should match the output of validToken.MarshalJSON().
var validJSONToken = `
{"body":{
"version":2,"nonce":12345,
"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}, 
"subjects":[{"nnsName":"alice.neo"}], 
"lifetime":{"exp":"32058350", "nbf":"93843742391", "iat":"1209418302"}, 
"contexts":[{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, "verbs":["CONTAINER_PUT"]}, {"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, "verbs":["OBJECT_PUT"]}],
"final":true},"signature":{"key":"A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa", "signature":"YW55X3NpZ25hdHVyZQ==", "scheme":1208743532},
"origin":{"body":{"version":0, "nonce":0, "issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}, "subjects":[{"nnsName":"alice.neo"}], 
"lifetime":{"exp":"32058350", "nbf":"93843742391", "iat":"1209418302"}, "contexts":[{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, "verbs":["OBJECT_PUT"]}], "final":false}, 
"signature":{"key":"A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa", "signature":"YW55X3NpZ25hdHVyZQ==", "scheme":1208743532}, "origin":null}}
`

func checkTokenFields(t *testing.T, val session.Token) {
	require.EqualValues(t, anyValidVersion, val.Version())
	require.EqualValues(t, anyValidNonce, val.Nonce())
	require.True(t, val.IsFinal())
	require.Equal(t, val.Issuer(), anyValidUserID)
	require.Len(t, val.Subjects(), 1)
	require.Equal(t, val.Subjects()[0].NNSName(), anyValidNNS)
	require.Equal(t, anyValidExpTime, val.Exp())
	require.Equal(t, anyValidIatTime, val.Iat())
	require.Equal(t, anyValidNbfTime, val.Nbf())
	require.True(t, val.AssertVerb(anyValidContainerVerb, anyValidContainerID))
	sig, ok := val.Signature()
	require.True(t, ok)
	require.EqualValues(t, anyValidSignatureScheme, sig.Scheme())
	require.Equal(t, anyValidIssuerPublicKeyBytes, sig.PublicKeyBytes())
	require.Equal(t, anyValidSignatureBytes, sig.Value())
}

type invalidBinTokenTestcase struct {
	name, err string
	b         []byte
}

var invalidBinCommonTestcases = []invalidBinTokenTestcase{
	{name: "body/issuer/value/empty", err: "invalid issuer: invalid length 0, expected 25",
		b: []byte{10, 2, 26, 0}},
	{name: "body/issuer/value/undersize", err: "invalid issuer: invalid length 24, expected 25",
		b: []byte{10, 28, 26, 26, 10, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "body/issuer/value/oversize", err: "invalid issuer: invalid length 26, expected 25",
		b: []byte{10, 30, 26, 28, 10, 26, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "body/issuer/value/wrong prefix", err: "invalid issuer: invalid prefix byte 0x42, expected 0x35",
		b: []byte{10, 29, 26, 27, 10, 25, 66, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{name: "body/issuer/value/checksum mismatch", err: "invalid issuer: checksum mismatch",
		b: []byte{10, 29, 26, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 93}},
	{name: "signature/invalid scheme", err: "invalid body signature: negative scheme -2147483648",
		b: []byte{18, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
	{name: "context/invalid verb", err: "invalid context at index 0: negative verb -1",
		b: []byte{10, 14, 50, 12, 18, 10, 255, 255, 255, 255, 255, 255, 255, 255, 255, 1}},
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
	{name: "origin/invalid signature", err: "invalid origin token: invalid body signature: negative scheme -1",
		b: []byte{26, 13, 18, 11, 24, 255, 255, 255, 255, 255, 255, 255, 255, 255, 1}},
}

type invalidJSONTokenTestcase struct {
	name, err string
	j         string
}

var invalidJSONCommonTestcases = []invalidJSONTokenTestcase{
	{
		name: "body/issuer/value/empty",
		err:  "invalid issuer: invalid length 0, expected 25",
		j:    `{"body":{"issuer":{"value":""}}}`,
	},
	{
		name: "body/issuer/value/undersize",
		err:  "invalid issuer: invalid length 24, expected 25",
		j:    `{"body":{"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMR"}}}`,
	},
	{
		name: "body/issuer/value/oversize",
		err:  "invalid issuer: invalid length 26, expected 25",
		j:    `{"body":{"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXAE="}}}`,
	},
	{
		name: "body/issuer/value/wrong prefix",
		err:  "invalid issuer: invalid prefix byte 0x42, expected 0x35",
		j:    `{"body":{"issuer":{"value":"QjMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}}}`,
	},
	{
		name: "body/issuer/value/checksum mismatch",
		err:  "invalid issuer: checksum mismatch",
		j:    `{"body":{"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXQ=="}}}`,
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
		name: "origin/invalid signature scheme",
		err:  "invalid origin token: invalid body signature: negative scheme -1",
		j:    `{"origin":{"signature":{"scheme":-1}}}`,
	},
}
