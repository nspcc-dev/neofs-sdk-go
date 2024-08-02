package session_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	apisession "github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
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
		validContainerTokens[i].SetAuthPublicKey(anyValidSessionKeyBytes)
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
	require.Equal(t, anyValidSessionKeyBytes, val.AuthPublicKey())
	require.True(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	require.EqualValues(t, anyValidSignatureScheme, val.Signature().Scheme())
	require.Equal(t, anyValidIssuerPublicKeyBytes, val.Signature().PublicKeyBytes())
	require.Equal(t, anyValidSignatureBytes, val.Signature().Value())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range append(invalidProtoTokenCommonTestcases, invalidProtoTokenTestcase{
			name: "context/missing", err: "missing session context",
			corrupt: func(m *apisession.SessionToken) { m.GetBody().SetContext(nil) },
		}, invalidProtoTokenTestcase{
			name: "context/wrong", err: "invalid context: invalid context *session.ObjectSessionContext",
			corrupt: func(m *apisession.SessionToken) { m.GetBody().SetContext(new(apisession.ObjectSessionContext)) },
		}, invalidProtoTokenTestcase{
			name: "context/invalid verb", err: "invalid context: verb 2147483648 overflows int32",
			corrupt: func(m *apisession.SessionToken) {
				var c apisession.ContainerSessionContext
				c.SetWildcard(true)
				c.SetVerb(math.MaxInt32 + 1)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/neither container nor wildcard", err: "invalid context: missing container or wildcard flag",
			corrupt: func(m *apisession.SessionToken) { m.GetBody().SetContext(new(apisession.ContainerSessionContext)) },
		}, invalidProtoTokenTestcase{
			name: "context/both container and wildcard", err: "invalid context: container conflicts with wildcard flag",
			corrupt: func(m *apisession.SessionToken) {
				var c apisession.ContainerSessionContext
				c.SetContainerID(new(refs.ContainerID))
				c.SetWildcard(true)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/nil value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *apisession.SessionToken) {
				var c apisession.ContainerSessionContext
				c.SetContainerID(new(refs.ContainerID))
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *apisession.SessionToken) {
				var id refs.ContainerID
				id.SetValue([]byte{})
				var c apisession.ContainerSessionContext
				c.SetContainerID(&id)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			corrupt: func(m *apisession.SessionToken) {
				var id refs.ContainerID
				id.SetValue(make([]byte, 31))
				var c apisession.ContainerSessionContext
				c.SetContainerID(&id)
				m.GetBody().SetContext(&c)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			corrupt: func(m *apisession.SessionToken) {
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
	val.SetAuthPublicKey(anyValidSessionKeyBytes)
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
	require.Zero(t, val.AuthPublicKey())
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	require.Negative(t, val.Signature().Scheme())

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
		b, err := json.MarshalIndent(validContainerTokens[i], "", " ")
		require.NoError(t, err, i)
		if string(b) != validJSONContainerTokens[i] {
			// protojson is inconsistent https://github.com/golang/protobuf/issues/1121
			var val session.Container
			require.NoError(t, val.UnmarshalJSON(b), i)
			t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
			require.Equal(t, validContainerTokens[i], val, i)
		}
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
	require.Zero(t, val.AuthPublicKey())
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.True(t, val.AppliedTo(anyValidContainerID))
	require.Negative(t, val.Signature().Scheme())

	// filled
	for i := range validContainerTokens {
		require.NoError(t, val.UnmarshalJSON([]byte(validJSONContainerTokens[i])), i)
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
		require.Equal(t, validContainerTokens[i], val, i)
	}
}

func TestContainer_AttachSignature(t *testing.T) {
	testAttachSignature(t, session.Container{})
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

func TestContainer_ID(t *testing.T) {
	testTokenID(t, session.Container{})
}

func TestContainer_AssertAuthKey(t *testing.T) {
	var x session.Container

	key := neofscryptotest.Signer().Public()

	require.False(t, x.AssertAuthKey(key))

	x.SetAuthKey(key)

	require.True(t, x.AssertAuthKey(key))
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

	require.NoError(t, session.Issue(&token, signer))
	require.True(t, session.IssuedBy(token, issuer))
}

func TestContainer_Issuer(t *testing.T) {
	testTokenIssuer(t, session.Container{})
}

func TestContainer_Sign(t *testing.T) {
	val := sessiontest.Container()

	require.NoError(t, val.SetSignature(neofscryptotest.Signer()))
	require.Zero(t, val.Issuer())
	require.True(t, val.VerifySignature())

	require.NoError(t, val.Sign(usertest.User()))

	require.True(t, val.VerifySignature())

	t.Run("issue#546", func(t *testing.T) {
		usr1 := usertest.User()
		usr2 := usertest.User()
		require.False(t, usr1.UserID() == usr2.UserID())

		token1 := sessiontest.Container()
		require.NoError(t, token1.Sign(usr1))
		require.Equal(t, usr1.UserID(), token1.Issuer())

		// copy token and re-sign
		var token2 session.Container
		token1.CopyTo(&token2)
		require.NoError(t, token2.Sign(usr2))
		require.Equal(t, usr2.UserID(), token2.Issuer())
	})
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
	require.Zero(t, val.AuthPublicKey())
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

func TestVerifyContainerSessionDataSignatureRFC6979(t *testing.T) {
	usr := neofscryptotest.Signer()
	data := []byte("Hello, world!")
	validSig, err := usr.RFC6979.Sign(data)
	require.NoError(t, err)

	var tok session.Container
	require.False(t, session.VerifyContainerSessionDataSignatureRFC6979(tok, data, validSig))
	require.False(t, tok.VerifySessionDataSignature(data, validSig))

	tok.SetAuthPublicKey(usr.PublicKeyBytes)
	require.True(t, session.VerifyContainerSessionDataSignatureRFC6979(tok, data, validSig))
	require.True(t, tok.VerifySessionDataSignature(data, validSig))

	for i := range usr.PublicKeyBytes {
		otherPub := bytes.Clone(usr.PublicKeyBytes)
		otherPub[i]++
		tok := tok
		tok.SetAuthPublicKey(otherPub)
		require.False(t, session.VerifyContainerSessionDataSignatureRFC6979(tok, data, validSig))
		require.False(t, tok.VerifySessionDataSignature(data, validSig))
	}
	for i := range data {
		otherData := bytes.Clone(data)
		otherData[i]++
		require.False(t, session.VerifyContainerSessionDataSignatureRFC6979(tok, otherData, validSig))
		require.False(t, tok.VerifySessionDataSignature(otherData, validSig))
	}
	for i := range validSig {
		otherSig := bytes.Clone(validSig)
		otherSig[i]++
		require.False(t, session.VerifyContainerSessionDataSignatureRFC6979(tok, data, otherSig))
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

func TestContainer_SetAuthPublicKey(t *testing.T) {
	testAuthPublicKeyField(t, session.Container.AuthPublicKey, (*session.Container).SetAuthPublicKey)
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
