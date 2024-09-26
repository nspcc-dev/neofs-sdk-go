package session_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	apisession "github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
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
		8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 50, 2, 16, 1},
	{10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195,
		17, 92, 26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4,
		34, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179,
		8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 50, 36, 26, 34, 10, 32, 243, 245, 75, 198, 48, 107,
		141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240,
		79, 114, 252, 227},
}

// corresponds to validContainerTokens.
var validBinContainerTokens = [][]byte{
	{10, 106, 10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92,
		26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 34, 33, 2,
		149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2,
		1, 95, 85, 78, 45, 197, 136, 50, 2, 16, 1, 18, 56, 10, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123,
		174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249,
		34, 218, 18, 13, 97, 110, 121, 95, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 236, 236, 175, 192, 4},
	{10, 140, 1, 10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53,
		51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92,
		26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 34, 33, 2,
		149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2,
		1, 95, 85, 78, 45, 197, 136, 50, 36, 26, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
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

func TestContainer_ReadFromV2(t *testing.T) {
	var lt apisession.TokenLifetime
	lt.SetExp(anyValidExp)
	lt.SetIat(anyValidIat)
	lt.SetNbf(anyValidNbf)
	var mo refs.OwnerID
	mo.SetValue(anyValidUserID[:])
	var mcnr refs.ContainerID
	mcnr.SetValue(anyValidContainerID[:])
	var mc apisession.ContainerSessionContext
	mc.SetContainerID(&mcnr)
	mc.SetVerb(anyValidContainerVerb)
	var mb apisession.TokenBody
	mb.SetID(anyValidSessionID[:])
	mb.SetOwnerID(&mo)
	mb.SetLifetime(&lt)
	mb.SetSessionKey(anyValidSessionKeyBytes)
	mb.SetContext(&mc)
	var msig refs.Signature
	msig.SetKey(anyValidIssuerPublicKeyBytes)
	msig.SetScheme(refs.SignatureScheme(anyValidSignatureScheme))
	msig.SetSign(anyValidSignatureBytes)
	var m apisession.Token
	m.SetBody(&mb)
	m.SetSignature(&msig)

	var val session.Container
	require.NoError(t, val.ReadFromV2(m))
	require.Equal(t, val.ID(), anyValidSessionID)
	require.Equal(t, val.Issuer(), anyValidUserID)
	require.EqualValues(t, anyValidExp, val.Exp())
	require.EqualValues(t, anyValidIat, val.Iat())
	require.EqualValues(t, anyValidNbf, val.Nbf())
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
			corrupt: func(m *apisession.Token) { m.GetBody().SetContext(nil) },
		}, invalidProtoTokenTestcase{
			name: "context/wrong", err: "invalid context: invalid context *session.ObjectSessionContext",
			corrupt: func(m *apisession.Token) { m.GetBody().SetContext(new(apisession.ObjectSessionContext)) },
		}, invalidProtoTokenTestcase{
			name: "context/invalid verb", err: "invalid context: verb 2147483648 overflows int32",
			corrupt: func(m *apisession.Token) {
				var c apisession.ContainerSessionContext
				c.SetWildcard(true)
				c.SetVerb(math.MaxInt32 + 1)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/neither container nor wildcard", err: "invalid context: missing container or wildcard flag",
			corrupt: func(m *apisession.Token) { m.GetBody().SetContext(new(apisession.ContainerSessionContext)) },
		}, invalidProtoTokenTestcase{
			name: "context/both container and wildcard", err: "invalid context: container conflicts with wildcard flag",
			corrupt: func(m *apisession.Token) {
				var c apisession.ContainerSessionContext
				c.SetContainerID(new(refs.ContainerID))
				c.SetWildcard(true)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/nil value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *apisession.Token) {
				var c apisession.ContainerSessionContext
				c.SetContainerID(new(refs.ContainerID))
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *apisession.Token) {
				var id refs.ContainerID
				id.SetValue([]byte{})
				var c apisession.ContainerSessionContext
				c.SetContainerID(&id)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			corrupt: func(m *apisession.Token) {
				var id refs.ContainerID
				id.SetValue(make([]byte, 31))
				var c apisession.ContainerSessionContext
				c.SetContainerID(&id)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			corrupt: func(m *apisession.Token) {
				var id refs.ContainerID
				id.SetValue(make([]byte, 33))
				var c apisession.ContainerSessionContext
				c.SetContainerID(&id)
				m.GetBody().SetContext(&c)
			},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				var m apisession.Token
				st.WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(session.Container).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestContainer_WriteToV2(t *testing.T) {
	var val session.Container
	var m apisession.Token

	// zero
	val.WriteToV2(&m)
	require.Zero(t, m.GetSignature())
	body := m.GetBody()
	require.NotNil(t, body)
	require.Zero(t, body.GetID())
	require.Zero(t, body.GetOwnerID())
	require.Zero(t, body.GetLifetime())
	require.Zero(t, body.GetSessionKey())
	c := body.GetContext()
	require.IsType(t, new(apisession.ContainerSessionContext), c)
	cc := c.(*apisession.ContainerSessionContext)
	require.Zero(t, cc.Verb())
	require.Zero(t, cc.ContainerID())
	require.True(t, cc.Wildcard())

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

	val.WriteToV2(&m)
	body = m.GetBody()
	require.NotNil(t, body)
	require.Equal(t, anyValidSessionID[:], body.GetID())
	require.Equal(t, anyValidUserID[:], body.GetOwnerID().GetValue())
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
	require.IsType(t, new(apisession.ContainerSessionContext), c)
	cc = c.(*apisession.ContainerSessionContext)
	require.EqualValues(t, anyValidContainerVerb, cc.Verb())
	require.Equal(t, anyValidContainerID[:], cc.ContainerID().GetValue())
	require.Zero(t, cc.Wildcard())
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
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.ObjectSessionContext",
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
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	for i := range validContainerTokens {
		err := val.Unmarshal(validBinContainerTokens[i])
		require.NoError(t, err)
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
		require.Equal(t, validContainerTokens[i], val)
	}
}

func TestContainer_MarshalJSON(t *testing.T) {
	for i := range validContainerTokens {
		//nolint:staticcheck
		b, err := json.MarshalIndent(validContainerTokens[i], "", " ")
		require.NoError(t, err, i)
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
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
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.ObjectSessionContext", j: `
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
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	for i := range validContainerTokens {
		require.NoError(t, val.UnmarshalJSON([]byte(validJSONContainerTokens[i])), i)
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
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
		{7, 252, 130, 23, 167, 44, 8, 109, 123, 206, 34, 95, 110, 184, 195, 141, 43, 84, 35, 138, 93, 216, 168,
			230, 242, 242, 159, 103, 133, 142, 141, 104, 77, 166, 42, 74, 3, 150, 102, 137, 185, 116, 51, 101,
			147, 33, 4, 7, 14, 65, 174, 28, 44, 91, 168, 58, 128, 38, 163, 102, 52, 239, 213, 118},
		{239, 43, 34, 239, 180, 70, 238, 28, 100, 254, 33, 4, 89, 177, 32, 18, 215, 175, 8, 126, 126, 104, 102,
			180, 121, 13, 39, 78, 50, 132, 119, 250, 114, 225, 242, 135, 253, 191, 99, 129, 229, 108, 148, 223,
			24, 240, 44, 229, 102, 141, 124, 151, 121, 196, 250, 63, 116, 107, 113, 75, 109, 169, 249, 11},
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
			{4, 31, 195, 240, 176, 85, 91, 249, 98, 82, 96, 126, 76, 27, 6, 181, 195, 193, 197, 62, 209, 78, 170, 109, 31, 169,
				249, 24, 211, 167, 110, 165, 200, 49, 194, 72, 123, 151, 121, 63, 29, 111, 22, 71, 220, 145, 58, 135, 95,
				244, 202, 224, 70, 162, 136, 39, 30, 58, 151, 240, 9, 65, 144, 32, 184},
			{4, 43, 119, 4, 66, 20, 131, 214, 29, 233, 25, 125, 222, 56, 184, 15, 153, 70, 48, 112, 211, 193, 79, 49, 233,
				36, 188, 130, 244, 42, 19, 134, 179, 5, 32, 143, 63, 35, 52, 228, 149, 202, 170, 174, 150, 246, 116, 182,
				44, 89, 25, 91, 172, 56, 163, 22, 33, 103, 8, 245, 245, 140, 212, 146, 186},
		}},
		{scheme: neofscrypto.ECDSA_DETERMINISTIC_SHA256, sigs: [][]byte{
			{184, 33, 118, 69, 74, 185, 216, 122, 57, 209, 165, 34, 215, 252, 81, 171, 91, 211, 169, 223, 107, 78, 246, 20,
				87, 15, 37, 126, 255, 170, 43, 89, 138, 25, 255, 54, 243, 205, 122, 120, 184, 22, 43, 72, 252, 254, 109,
				91, 176, 30, 116, 54, 181, 75, 172, 137, 245, 155, 232, 0, 96, 102, 15, 228},
			{221, 183, 51, 58, 146, 202, 120, 39, 156, 110, 158, 90, 11, 13, 7, 216, 227, 69, 190, 152, 110, 159, 17, 64, 251,
				35, 96, 40, 106, 69, 211, 112, 139, 127, 24, 179, 13, 199, 161, 102, 117, 217, 61, 25, 144, 222, 171, 203, 240,
				247, 50, 152, 151, 244, 69, 69, 69, 21, 221, 232, 12, 131, 163, 87},
		}},
		{scheme: neofscrypto.ECDSA_WALLETCONNECT, sigs: [][]byte{
			{30, 49, 223, 33, 75, 83, 235, 194, 92, 37, 74, 128, 38, 58, 215, 178, 79, 130, 40, 59, 77, 83, 126, 46, 68, 1,
				233, 170, 162, 153, 83, 65, 53, 171, 44, 138, 187, 214, 130, 160, 167, 96, 171, 7, 164, 95, 40, 58, 108, 214, 246,
				192, 239, 15, 36, 194, 179, 189, 192, 117, 166, 80, 176, 247, 117, 104, 6, 229, 191, 221, 25, 152, 75, 103, 187,
				125, 152, 193, 180, 204},
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
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.ObjectSessionContext",
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
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))

	// filled
	for i := range validContainerTokens {
		err := val.UnmarshalSignedData(validSignedContainerTokens[i])
		require.NoError(t, err)
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
		require.Equal(t, validContainerTokens[i], val)
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

func TestContainer_IssuerPublicKeyBytes(t *testing.T) {
	var val session.Container
	require.Zero(t, val.IssuerPublicKeyBytes())

	sig := neofscrypto.NewSignatureFromRawKey(anyValidSignatureScheme, anyValidIssuerPublicKeyBytes, anyValidSignatureBytes)
	val.AttachSignature(sig)
	require.Equal(t, anyValidIssuerPublicKeyBytes, val.IssuerPublicKeyBytes())

	otherKey := bytes.Clone(anyValidIssuerPublicKeyBytes)
	otherKey[0]++
	sig.SetPublicKeyBytes(otherKey)
	val.AttachSignature(sig)
	require.Equal(t, otherKey, val.IssuerPublicKeyBytes())
}
