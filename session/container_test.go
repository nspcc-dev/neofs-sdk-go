package session_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"math/big"
	"strconv"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

const (
	anyValidContainerVerb = 837285395
)

// set by init.
var validContainerTokens = make([]session.Container, 2)

func init() {
	for i := range validContainerTokens {
		validContainerTokens[i].SetID(anyValidSessionID)
		validContainerTokens[i].SetIssuer(anyValidUserID)
		validContainerTokens[i].SetExp(anyValidExp)
		validContainerTokens[i].SetIat(anyValidIat)
		validContainerTokens[i].SetNbf(anyValidNbf)
		validContainerTokens[i].SetAuthKey(anyValidSessionKey)
		validContainerTokens[i].ForVerb(anyValidContainerVerb)
		validContainerTokens[i].AttachSignature(anyValidSignature)
	}
	validContainerTokens[1].ApplyOnlyTo(anyValidContainerID)
}

// corresponds to validContainerTokens.
var validSignedContainerTokens = [][]byte{
	{10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195,
		17, 92, 26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4,
		34, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179,
		8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 50, 8, 8, 147, 236, 159, 143, 3, 16, 1},
	{10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195,
		17, 92, 26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4,
		34, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179,
		8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 50, 42, 8, 147, 236, 159, 143, 3, 26, 34, 10, 32, 243, 245, 75, 198, 48, 107,
		141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240,
		79, 114, 252, 227},
}

// corresponds to validContainerTokens.
var validBinContainerTokens = [][]byte{
	{10, 112, 10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92,
		26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 34, 33, 2,
		149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2,
		1, 95, 85, 78, 45, 197, 136, 50, 8, 8, 147, 236, 159, 143, 3, 16, 1, 18, 56, 10, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123,
		174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249,
		34, 218, 18, 13, 97, 110, 121, 95, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 236, 236, 175, 192, 4},
	{10, 146, 1, 10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92,
		26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 34, 33, 2,
		149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2,
		1, 95, 85, 78, 45, 197, 136, 50, 42, 8, 147, 236, 159, 143, 3, 26, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
		21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 56, 10, 33,
		3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223,
		219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218, 18, 13, 97, 110, 121, 95, 115, 105, 103, 110, 97, 116,
		117, 114, 101, 24, 236, 236, 175, 192, 4},
}

// corresponds to validContainerTokens.
var validJSONContainerTokens = []string{`
{
 "body": {
  "id": "YxhvRhasSBSLu69iCv/nvA==",
  "ownerID": {
   "value": "NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="
  },
  "lifetime": {
   "exp": "32058350",
   "nbf": "93843742391",
   "iat": "1209418302"
  },
  "sessionKey": "ApUrMsRbsT6D6X7xsQ1OYF53RzezCDXxTwIBX1VOLcWI",
  "container": {
   "verb": 837285395,
   "wildcard": true,
   "containerID": null
  }
 },
 "signature": {
  "key": "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
  "signature": "YW55X3NpZ25hdHVyZQ==",
  "scheme": 1208743532
 }
}
`, `
{
 "body": {
  "id": "YxhvRhasSBSLu69iCv/nvA==",
  "ownerID": {
   "value": "NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="
  },
  "lifetime": {
   "exp": "32058350",
   "nbf": "93843742391",
   "iat": "1209418302"
  },
  "sessionKey": "ApUrMsRbsT6D6X7xsQ1OYF53RzezCDXxTwIBX1VOLcWI",
  "container": {
   "verb": 837285395,
   "wildcard": false,
   "containerID": {
    "value": "8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="
   }
  }
 },
 "signature": {
  "key": "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
  "signature": "YW55X3NpZ25hdHVyZQ==",
  "scheme": 1208743532
 }
}
`}

func TestContainer_FromProtoMessage(t *testing.T) {
	m := &protosession.SessionToken{
		Body: &protosession.SessionToken_Body{
			Id:         anyValidSessionID[:],
			OwnerId:    &refs.OwnerID{Value: anyValidUserID[:]},
			Lifetime:   &protosession.SessionToken_Body_TokenLifetime{Exp: anyValidExp, Nbf: anyValidNbf, Iat: anyValidIat},
			SessionKey: anyValidSessionKeyBytes,
			Context: &protosession.SessionToken_Body_Container{Container: &protosession.ContainerSessionContext{
				Verb:        anyValidContainerVerb,
				ContainerId: &refs.ContainerID{Value: anyValidContainerID[:]},
			}},
		},
		Signature: &refs.Signature{
			Key:    anyValidIssuerPublicKeyBytes,
			Sign:   anyValidSignatureBytes,
			Scheme: anyValidSignatureScheme,
		},
	}

	var val session.Container
	require.NoError(t, val.FromProtoMessage(m))
	require.Equal(t, val.ID(), anyValidSessionID)
	require.Equal(t, val.Issuer(), anyValidUserID)
	require.EqualValues(t, anyValidExp, val.Exp())
	require.EqualValues(t, anyValidIat, val.Iat())
	require.EqualValues(t, anyValidNbf, val.Nbf())
	authUser, err := val.AuthUser()
	require.NoError(t, err)
	require.Equal(t, user.NewFromECDSAPublicKey(*(*ecdsa.PublicKey)(anyValidSessionKey)), authUser)
	require.True(t, val.AssertAuthKey(anyValidSessionKey))
	require.True(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	sig, ok := val.Signature()
	require.True(t, ok)
	require.EqualValues(t, anyValidSignatureScheme, sig.Scheme())
	require.Equal(t, anyValidIssuerPublicKeyBytes, sig.PublicKeyBytes())
	require.Equal(t, anyValidSignatureBytes, sig.Value())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range append(invalidProtoTokenCommonTestcases, invalidProtoTokenTestcase{
			name: "context/missing", err: "missing session context",
			corrupt: func(m *protosession.SessionToken) { m.Body.Context = nil },
		}, invalidProtoTokenTestcase{
			name: "context/wrong", err: "invalid context: invalid context *session.SessionToken_Body_Object",
			corrupt: func(m *protosession.SessionToken) { m.Body.Context = new(protosession.SessionToken_Body_Object) },
		}, invalidProtoTokenTestcase{
			name: "context/invalid verb", err: "invalid context: negative verb -1",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.Verb = -1
			},
		}, invalidProtoTokenTestcase{
			name: "context/neither container nor wildcard", err: "invalid context: missing container or wildcard flag",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.Reset()
			},
		}, invalidProtoTokenTestcase{
			name: "context/both container and wildcard", err: "invalid context: container conflicts with wildcard flag",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.Wildcard = true
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/nil value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.ContainerId.Value = nil
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.ContainerId.Value = []byte{}
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.ContainerId.Value = make([]byte, 31)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Container).Container.ContainerId.Value = make([]byte, 33)
			},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(session.Container).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestContainer_ProtoMessage(t *testing.T) {
	var val session.Container

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetSignature())
	body := m.GetBody()
	require.NotNil(t, body)
	require.Zero(t, body.GetId())
	require.Zero(t, body.GetOwnerId())
	require.Zero(t, body.GetLifetime())
	require.Zero(t, body.GetSessionKey())
	c := body.GetContext()
	require.IsType(t, new(protosession.SessionToken_Body_Container), c)
	cc := c.(*protosession.SessionToken_Body_Container).Container
	require.Zero(t, cc.GetVerb())
	require.Zero(t, cc.GetContainerId())
	require.True(t, cc.GetWildcard())

	// filled
	val.SetID(anyValidSessionID)
	val.SetIssuer(anyValidUserID)
	val.SetExp(anyValidExp)
	val.SetIat(anyValidIat)
	val.SetNbf(anyValidNbf)
	val.SetAuthKey(anyValidSessionKey)
	val.ForVerb(anyValidContainerVerb)
	val.ApplyOnlyTo(anyValidContainerID)
	val.AttachSignature(anyValidSignature)

	m = val.ProtoMessage()
	body = m.GetBody()
	require.NotNil(t, body)
	require.Equal(t, anyValidSessionID[:], body.GetId())
	require.Equal(t, anyValidUserID[:], body.GetOwnerId().GetValue())
	lt := body.GetLifetime()
	require.EqualValues(t, anyValidExp, lt.GetExp())
	require.EqualValues(t, anyValidIat, lt.GetIat())
	require.EqualValues(t, anyValidNbf, lt.GetNbf())
	require.Equal(t, anyValidSessionKeyBytes, body.GetSessionKey())
	sig := m.GetSignature()
	require.NotNil(t, sig)
	require.EqualValues(t, anyValidSignatureScheme, sig.GetScheme())
	require.Equal(t, anyValidIssuerPublicKeyBytes, sig.GetKey())
	require.Equal(t, anyValidSignatureBytes, sig.GetSign())
	c = body.GetContext()
	require.IsType(t, new(protosession.SessionToken_Body_Container), c)
	cc = c.(*protosession.SessionToken_Body_Container).Container
	require.EqualValues(t, anyValidContainerVerb, cc.GetVerb())
	require.Equal(t, anyValidContainerID[:], cc.GetContainerId().GetValue())
	require.Zero(t, cc.GetWildcard())
}

func TestContainer_Marshal(t *testing.T) {
	for i := range validContainerTokens {
		require.Equal(t, validBinContainerTokens[i], validContainerTokens[i].Marshal(), i)
	}
}

func TestContainer_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(session.Container).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range append(invalidBinTokenCommonTestcases, invalidBinTokenTestcase{
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.SessionToken_Body_Object",
			b: []byte{10, 4, 42, 2, 18, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/both container and wildcard", err: "invalid context: container conflicts with wildcard flag",
			b: []byte{10, 6, 50, 4, 16, 1, 26, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			b: []byte{10, 4, 50, 2, 26, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			b: []byte{10, 37, 50, 35, 26, 33, 10, 31, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252},
		}, invalidBinTokenTestcase{
			name: "body/context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			b: []byte{10, 39, 50, 37, 26, 35, 10, 33, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 1},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Container).Unmarshal(tc.b), tc.err)
			})
		}
	})
	t.Run("no container", func(t *testing.T) {
		var val session.Container
		cnr := cidtest.ID()
		cnrOther := cidtest.OtherID(cnr)
		val.ApplyOnlyTo(cnr)
		require.True(t, val.AppliedTo(cnr))
		require.False(t, val.AppliedTo(cnrOther))

		b := []byte{10, 2, 50, 0}
		require.NoError(t, val.Unmarshal(b))
		require.True(t, val.AppliedTo(cnr))
		require.True(t, val.AppliedTo(cnrOther))
	})

	var val session.Container
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Zero(t, val.ID())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	authUser, err := val.AuthUser()
	require.Error(t, err)
	require.Zero(t, authUser)
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	for i := range validContainerTokens {
		err := val.Unmarshal(validBinContainerTokens[i])
		require.NoError(t, err)
		require.Equal(t, validContainerTokens[i], val)
	}
}

func TestContainer_MarshalJSON(t *testing.T) {
	for i := range validContainerTokens {
		b, err := json.MarshalIndent(validContainerTokens[i], "", " ")
		require.NoError(t, err, i)
		require.JSONEq(t, validJSONContainerTokens[i], string(b))
	}
}

func TestContainer_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(session.Container).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range append(invalidJSONTokenCommonTestcases, invalidJSONTokenTestcase{
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.SessionToken_Body_Object", j: `
{"body":{"object":{}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/both container and wildcard", err: "invalid context: container conflicts with wildcard flag", j: `
{"body":{"container":{"wildcard":true,"containerID":{}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/container/empty value", err: "invalid context: invalid container ID: invalid length 0", j: `
{"body":{"container":{"containerID":{}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/container/undersize", err: "invalid context: invalid container ID: invalid length 31", j: `
{"body":{"container":{"containerID":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/A=="}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/container/oversize", err: "invalid context: invalid container ID: invalid length 33", j: `
{"body":{"container":{"containerID":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OMB"}}}}
`}) {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Container).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var val session.Container
	// zero
	require.NoError(t, val.UnmarshalJSON([]byte("{}")))
	require.Zero(t, val.ID())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	authUser, err := val.AuthUser()
	require.Error(t, err)
	require.Zero(t, authUser)
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	for i := range validContainerTokens {
		require.NoError(t, val.UnmarshalJSON([]byte(validJSONContainerTokens[i])), i)
		require.Equal(t, validContainerTokens[i], val, i)
	}
}

func TestContainer_AttachSignature(t *testing.T) {
	var val session.Container
	_, ok := val.Signature()
	require.False(t, ok)
	val.AttachSignature(anyValidSignature)
	sig, ok := val.Signature()
	require.True(t, ok)
	require.Equal(t, anyValidSignature, sig)
}

func TestContainer_ApplyOnlyTo(t *testing.T) {
	var val session.Container
	cnr1 := cidtest.ID()
	cnr2 := cidtest.OtherID(cnr1)

	require.True(t, val.AppliedTo(cnr1))
	require.True(t, val.AppliedTo(cnr2))

	val.ApplyOnlyTo(cnr1)
	require.True(t, val.AppliedTo(cnr1))
	require.False(t, val.AppliedTo(cnr2))

	val.ApplyOnlyTo(cnr2)
	require.False(t, val.AppliedTo(cnr1))
	require.True(t, val.AppliedTo(cnr2))

	val.ApplyOnlyTo(cid.ID{})
	require.True(t, val.AppliedTo(cnr1))
	require.True(t, val.AppliedTo(cnr2))
}

func TestContainer_InvalidAt(t *testing.T) {
	testValidAt(t, new(session.Container))
}

func TestContainer_ID(t *testing.T) {
	testTokenID(t, session.Container{})
}

func TestContainer_SetAuthKey(t *testing.T) {
	testSetAuthKey(t, (*session.Container).SetAuthKey, session.Container.AssertAuthKey)
}

func TestContainer_ForVerb(t *testing.T) {
	var val session.Container
	const verb1 = anyValidContainerVerb
	const verb2 = verb1 + 1

	require.True(t, val.AssertVerb(0))
	require.False(t, val.AssertVerb(verb1))
	require.False(t, val.AssertVerb(verb2))

	val.ForVerb(verb1)
	require.True(t, val.AssertVerb(verb1))
	require.False(t, val.AssertVerb(verb2))

	val.ForVerb(verb2)
	require.False(t, val.AssertVerb(verb1))
	require.True(t, val.AssertVerb(verb2))

	val.ForVerb(0)
	require.False(t, val.AssertVerb(verb1))
	require.False(t, val.AssertVerb(verb2))
}

func TestIssuedBy(t *testing.T) {
	var (
		token  session.Container
		issuer user.ID
		signer = usertest.User()
	)

	issuer = signer.UserID()

	require.False(t, session.IssuedBy(token, issuer))

	require.NoError(t, token.Sign(signer))
	require.True(t, session.IssuedBy(token, issuer))
}

func TestContainer_Issuer(t *testing.T) {
	testTokenIssuer(t, session.Container{})
}

func TestContainer_Sign(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		require.Error(t, new(session.Container).Sign(usertest.FailSigner(usertest.User())))
		require.ErrorIs(t, new(session.Container).Sign(user.NewSigner(neofscryptotest.Signer(), user.ID{})), user.ErrZeroID)
	})

	ecdsaPriv := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(),
			X: new(big.Int).SetBytes([]byte{244, 235, 150, 254, 16, 223, 121, 92, 82, 95, 93, 0, 218, 75, 97,
				182, 224, 29, 29, 126, 136, 127, 95, 227, 148, 120, 101, 174, 116, 191, 113, 56}),
			Y: new(big.Int).SetBytes([]byte{162, 142, 254, 167, 43, 228, 23, 134, 112, 148, 125, 252, 40, 205,
				120, 74, 50, 155, 194, 180, 37, 229, 18, 105, 143, 250, 110, 254, 3, 20, 159, 152}),
		},
		D: new(big.Int).SetBytes([]byte{37, 38, 152, 197, 254, 145, 122, 170, 199, 181, 85, 225, 135, 215,
			58, 94, 65, 111, 216, 11, 91, 240, 13, 191, 233, 192, 59, 95, 242, 32, 142, 145}),
	}

	var c session.Container
	for i, rfc6979Sig := range [][]byte{
		{190, 18, 239, 30, 103, 101, 136, 235, 201, 103, 161, 14, 141, 211, 187, 115, 174, 185, 216, 240, 250, 20,
			104, 255, 159, 187, 123, 153, 69, 51, 114, 161, 38, 249, 83, 23, 227, 242, 14, 169, 163, 96, 174,
			153, 174, 130, 142, 199, 157, 243, 8, 254, 0, 177, 165, 9, 148, 18, 72, 211, 199, 188, 220, 44},
		{5, 89, 114, 211, 237, 183, 201, 129, 24, 221, 131, 188, 255, 135, 221, 55, 49, 206, 184, 128, 44, 66,
			244, 148, 51, 41, 242, 41, 97, 153, 7, 70, 72, 63, 192, 149, 12, 37, 63, 4, 192, 125, 161, 27, 123,
			242, 76, 178, 148, 202, 241, 54, 4, 108, 34, 182, 217, 246, 125, 107, 132, 53, 91, 188},
	} {
		validContainerTokens[i].CopyTo(&c)
		t.Run("container#"+strconv.Itoa(i), func(t *testing.T) {
			testSignCDSA(t, ecdsaPriv, anyValidUserID, &c, validSignedContainerTokens[i], rfc6979Sig)
			testSetSignatureECDSA(t, ecdsaPriv, &c, validSignedContainerTokens[i], rfc6979Sig)
		})
	}
}

func TestContainer_VerifySignature(t *testing.T) {
	// keys used for this test
	// ecdsa.PrivateKey{
	// 	PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(),
	// 		X: new(big.Int).SetBytes([]byte{34, 204, 96, 183, 108, 209, 95, 61, 67, 216, 229, 8, 26, 112, 174, 164, 239,
	// 			94, 128, 115, 198, 56, 227, 185, 129, 205, 101, 244, 163, 157, 172, 116}),
	// 		Y: new(big.Int).SetBytes([]byte{187, 225, 7, 20, 148, 140, 234, 98, 202, 109, 145, 126, 126, 62, 188, 15, 56,
	// 			195, 237, 150, 247, 93, 101, 231, 140, 240, 19, 72, 16, 99, 6, 99}),
	// 	},
	// 	D: new(big.Int).SetBytes([]byte{125, 34, 9, 148, 39, 247, 116, 124, 27, 11, 166, 201, 232, 182, 153, 32, 117, 126,
	// 		24, 47, 85, 107, 215, 199, 26, 166, 96, 87, 234, 110, 151, 114}),
	// }
	pub := []byte{3, 34, 204, 96, 183, 108, 209, 95, 61, 67, 216, 229, 8, 26, 112, 174, 164, 239, 94, 128, 115,
		198, 56, 227, 185, 129, 205, 101, 244, 163, 157, 172, 116}
	var sig neofscrypto.Signature

	var c session.Container
	for i, tc := range []struct {
		scheme neofscrypto.Scheme
		sigs   [][]byte // of validContainerTokens
	}{
		{scheme: neofscrypto.ECDSA_SHA512, sigs: [][]byte{
			{4, 42, 31, 236, 138, 99, 174, 186, 104, 85, 109, 115, 31, 152, 42, 84, 148, 73, 12, 21, 206, 199, 211, 246, 191,
				185, 143, 181, 125, 99, 149, 43, 26, 49, 26, 152, 186, 161, 95, 12, 157, 144, 212, 203, 158, 233, 148, 226,
				165, 55, 67, 155, 84, 84, 129, 65, 10, 137, 254, 20, 157, 139, 229, 46, 218},
			{4, 9, 87, 55, 182, 1, 68, 11, 29, 1, 146, 125, 72, 110, 146, 231, 62, 138, 245, 54, 16, 161, 248, 28, 7, 201,
				26, 25, 158, 27, 144, 224, 99, 226, 173, 191, 116, 60, 207, 247, 101, 233, 87, 205, 55, 162, 129, 182,
				211, 149, 194, 23, 242, 124, 238, 56, 80, 109, 45, 165, 15, 129, 91, 7, 180},
		}},
		{scheme: neofscrypto.ECDSA_DETERMINISTIC_SHA256, sigs: [][]byte{
			{211, 170, 29, 36, 209, 239, 196, 159, 185, 21, 248, 226, 171, 179, 107, 14, 171, 214, 250, 240, 188, 188, 95, 8,
				217, 230, 5, 85, 176, 231, 159, 77, 23, 181, 10, 140, 183, 169, 166, 218, 181, 21, 216, 53, 5, 39, 29, 89, 189,
				7, 79, 67, 114, 72, 62, 136, 144, 73, 91, 76, 151, 52, 1, 205},
			{129, 98, 32, 222, 16, 112, 71, 181, 155, 28, 175, 176, 189, 243, 132, 130, 112, 157, 244, 105, 218, 22, 28, 27,
				105, 109, 49, 184, 52, 180, 37, 151, 104, 161, 105, 108, 247, 104, 201, 72, 75, 5, 233, 94, 152, 136, 202, 63,
				121, 77, 193, 129, 137, 248, 215, 211, 77, 6, 147, 100, 201, 79, 18, 125},
		}},
		{scheme: neofscrypto.ECDSA_WALLETCONNECT, sigs: [][]byte{
			{105, 91, 121, 219, 8, 156, 202, 92, 24, 217, 154, 168, 237, 8, 93, 138, 226, 111, 165, 72, 22, 245, 197, 64, 14,
				26, 207, 40, 110, 182, 182, 190, 53, 107, 12, 43, 115, 20, 250, 194, 251, 194, 160, 151, 48, 244, 126, 10, 185,
				226, 201, 137, 35, 122, 186, 69, 8, 239, 68, 66, 87, 126, 116, 12, 150, 15, 108, 163, 129, 197, 192, 140, 15, 96,
				16, 38, 160, 81, 110, 250},
		}},
	} {
		sig.SetScheme(tc.scheme)
		for j, sigBytes := range tc.sigs {
			validContainerTokens[j].CopyTo(&c)
			sig.SetPublicKeyBytes(pub)
			sig.SetValue(sigBytes)
			c.AttachSignature(sig)
			require.True(t, c.VerifySignature(), [2]int{i, j})
			for k := range pub {
				pubCp := bytes.Clone(pub)
				pubCp[k]++
				sig.SetPublicKeyBytes(pubCp)
				c.AttachSignature(sig)
				require.False(t, c.VerifySignature(), [2]int{i, j})
			}
			for k := range sigBytes {
				sigBytesCp := bytes.Clone(sigBytes)
				sigBytesCp[k]++
				sig.SetValue(sigBytesCp)
				c.AttachSignature(sig)
				require.False(t, c.VerifySignature(), [2]int{i, j})
			}
		}
	}
}

func TestContainer_SignedData(t *testing.T) {
	for i := range validSignedContainerTokens {
		require.Equal(t, validSignedContainerTokens[i], validContainerTokens[i].SignedData(), i)
	}
}

func TestContainer_UnmarshalSignedData(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(session.Container).UnmarshalSignedData([]byte("Hello, world!"))
			require.ErrorContains(t, err, "decode body")
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range append(invalidSignedTokenCommonTestcases, invalidBinTokenTestcase{
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.SessionToken_Body_Object",
			b: []byte{42, 2, 18, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/both container and wildcard", err: "invalid context: container conflicts with wildcard flag",
			b: []byte{50, 4, 16, 1, 26, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			b: []byte{50, 2, 26, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			b: []byte{50, 35, 26, 33, 10, 31, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252},
		}, invalidBinTokenTestcase{
			name: "body/context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			b: []byte{50, 37, 26, 35, 10, 33, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 1},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Container).UnmarshalSignedData(tc.b), tc.err)
			})
		}
	})

	var val session.Container
	// zero
	require.NoError(t, val.UnmarshalSignedData(nil))
	require.Zero(t, val.ID())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	authUser, err := val.AuthUser()
	require.Error(t, err)
	require.Zero(t, authUser)
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))

	// filled
	for i := range validContainerTokens {
		err := val.UnmarshalSignedData(validSignedContainerTokens[i])
		require.NoError(t, err)
		require.Equal(t, validSignedContainerTokens[i], val.SignedData())
	}
}

func TestContainer_VerifyDataSignature(t *testing.T) {
	usr := neofscryptotest.Signer()
	data := []byte("Hello, world!")
	validSig, err := usr.RFC6979.Sign(data)
	require.NoError(t, err)

	var tok session.Container
	require.False(t, tok.VerifySessionDataSignature(data, validSig))

	tok.SetAuthKey(usr.Public())
	require.True(t, tok.VerifySessionDataSignature(data, validSig))

	otherPub := neofscryptotest.Signer().Public()
	require.NotEqual(t, usr.Public(), otherPub)
	tokCp := tok
	tokCp.SetAuthKey(otherPub)
	require.False(t, tokCp.VerifySessionDataSignature(data, validSig))

	for i := range data {
		otherData := bytes.Clone(data)
		otherData[i]++
		require.False(t, tok.VerifySessionDataSignature(otherData, validSig))
	}
	for i := range validSig {
		otherSig := bytes.Clone(validSig)
		otherSig[i]++
		require.False(t, tok.VerifySessionDataSignature(data, otherSig))
	}
}

func TestContainer_SetExp(t *testing.T) {
	testLifetimeClaim(t, session.Container.Exp, (*session.Container).SetExp)
}

func TestContainer_SetIat(t *testing.T) {
	testLifetimeClaim(t, session.Container.Iat, (*session.Container).SetIat)
}

func TestContainer_SetNbf(t *testing.T) {
	testLifetimeClaim(t, session.Container.Nbf, (*session.Container).SetNbf)
}
