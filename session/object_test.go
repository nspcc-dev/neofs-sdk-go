package session_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"math/big"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

const anyValidObjectVerb = 32905430

var anyValidObjectIDs = []oid.ID{
	{243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227},
	{47, 240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20, 233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147, 249, 88, 184, 69, 145, 196, 127, 68},
	{59, 5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88, 57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165, 58, 126, 161, 221},
}

// set by init.
var validObjectToken session.Object

func init() {
	validObjectToken.SetID(anyValidSessionID)
	validObjectToken.SetIssuer(anyValidUserID)
	validObjectToken.SetExp(anyValidExp)
	validObjectToken.SetIat(anyValidIat)
	validObjectToken.SetNbf(anyValidNbf)
	validObjectToken.SetAuthKey(anyValidSessionKey)
	validObjectToken.ForVerb(anyValidContainerVerb)
	validObjectToken.BindContainer(anyValidContainerID)
	validObjectToken.LimitByObjects(anyValidObjectIDs...)
	validObjectToken.AttachSignature(anyValidSignature)
}

// corresponds to validObjectToken.
var validSignedObjectToken = []byte{
	10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53, 51, 5, 166, 111, 29,
	20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92, 26, 18, 8, 238, 215, 164, 15, 16,
	183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 34, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241,
	177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 42, 153, 1, 8, 147, 236, 159, 143,
	3, 18, 144, 1, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35,
	99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255,
	49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10, 32,
	47, 240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20, 233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147, 249, 88,
	184, 69, 145, 196, 127, 68, 18, 34, 10, 32, 59, 5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88, 57, 49, 95, 157, 218,
	79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165, 58, 126, 161, 221,
}

// corresponds to validObjectToken.
var validBinObjectToken = []byte{
	10, 130, 2, 10, 16, 99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188, 18, 27, 10, 25, 53, 51, 5,
	166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92, 26, 18, 8, 238, 215,
	164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190, 132, 217, 192, 4, 34, 33, 2, 149, 43, 50, 196, 91, 177, 62, 131, 233,
	126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136, 42, 153, 1, 8, 147, 236,
	159, 143, 3, 18, 144, 1, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147,
	43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121,
	255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10,
	32, 47, 240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20, 233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147, 249, 88,
	184, 69, 145, 196, 127, 68, 18, 34, 10, 32, 59, 5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88, 57, 49, 95, 157, 218,
	79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165, 58, 126, 161, 221, 18, 56, 10, 33, 3, 202, 217, 142, 98, 209, 190, 188,
	145, 123, 174, 21, 173, 239, 239, 245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34,
	218, 18, 13, 97, 110, 121, 95, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 236, 236, 175, 192, 4,
}

// corresponds to validObjectToken.
var validJSONObjectToken = `
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
  "object": {
   "verb": 837285395,
   "target": {
    "container": {
     "value": "8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="
    },
    "objects": [
     {
      "value": "8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="
     },
     {
      "value": "L/Bd2AlAWLfGJB5TFOl3/GCrBnpzqLqT+Vi4RZHEf0Q="
     },
     {
      "value": "OwV4v/o9+HKJFeVYOTFfndpPULHZOB0dryUqpTp+od0="
     }
    ]
   }
  }
 },
 "signature": {
  "key": "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
  "signature": "YW55X3NpZ25hdHVyZQ==",
  "scheme": 1208743532
 }
}
`

func TestObject_FromProtoMessage(t *testing.T) {
	mobjs := make([]*refs.ObjectID, len(anyValidObjectIDs))
	for i := range anyValidObjectIDs {
		mobjs[i] = &refs.ObjectID{Value: anyValidObjectIDs[i][:]}
	}
	m := &protosession.SessionToken{
		Body: &protosession.SessionToken_Body{
			Id:         anyValidSessionID[:],
			OwnerId:    &refs.OwnerID{Value: anyValidUserID[:]},
			Lifetime:   &protosession.SessionToken_Body_TokenLifetime{Exp: anyValidExp, Nbf: anyValidNbf, Iat: anyValidIat},
			SessionKey: anyValidSessionKeyBytes,
			Context: &protosession.SessionToken_Body_Object{Object: &protosession.ObjectSessionContext{
				Verb: anyValidObjectVerb,
				Target: &protosession.ObjectSessionContext_Target{
					Container: &refs.ContainerID{Value: anyValidContainerID[:]},
					Objects:   mobjs,
				},
			}},
		},
		Signature: &refs.Signature{
			Key:    anyValidIssuerPublicKeyBytes,
			Sign:   anyValidSignatureBytes,
			Scheme: anyValidSignatureScheme,
		},
	}

	var val session.Object
	require.NoError(t, val.FromProtoMessage(m))
	require.Equal(t, val.ID(), anyValidSessionID)
	require.Equal(t, val.Issuer(), anyValidUserID)
	require.EqualValues(t, anyValidExp, val.Exp())
	require.EqualValues(t, anyValidIat, val.Iat())
	require.EqualValues(t, anyValidNbf, val.Nbf())
	require.True(t, val.AssertAuthKey(anyValidSessionKey))
	require.True(t, val.AssertVerb(anyValidObjectVerb))
	require.True(t, val.AssertContainer(anyValidContainerID))
	for i := range anyValidObjectIDs {
		require.True(t, val.AssertObject(anyValidObjectIDs[i]))
	}
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
			name: "context/wrong", err: "invalid context: invalid context *session.SessionToken_Body_Container",
			corrupt: func(m *protosession.SessionToken) { m.Body.Context = new(protosession.SessionToken_Body_Container) },
		}, invalidProtoTokenTestcase{
			name: "context/invalid verb", err: "invalid context: negative verb -1",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Verb = -1
			},
		}, invalidProtoTokenTestcase{
			name: "context/invalid verb", err: "invalid context: negative verb -1",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Verb = -1
			},
		}, invalidProtoTokenTestcase{
			name: "context/missing container", err: "invalid context: missing target container",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target = nil
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/nil value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Container = new(refs.ContainerID)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Container.Value = []byte{}
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Container.Value = make([]byte, 31)
			},
		}, invalidProtoTokenTestcase{
			name: "context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Container.Value = make([]byte, 33)
			},
		}, invalidProtoTokenTestcase{
			name: "context/objects/nil element", err: "invalid context: nil target object #1",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Objects[1] = nil
			},
		}, invalidProtoTokenTestcase{
			name: "context/objects/nil value", err: "invalid context: invalid target object: invalid length 0",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Objects[1].Value = nil
			},
		}, invalidProtoTokenTestcase{
			name: "context/objects/empty value", err: "invalid context: invalid target object: invalid length 0",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Objects[1].Value = []byte{}
			},
		}, invalidProtoTokenTestcase{
			name: "context/objects/undersize", err: "invalid context: invalid target object: invalid length 31",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Objects[1].Value = make([]byte, 31)
			},
		}, invalidProtoTokenTestcase{
			name: "context/objects/oversize", err: "invalid context: invalid target object: invalid length 33",
			corrupt: func(m *protosession.SessionToken) {
				m.GetBody().GetContext().(*protosession.SessionToken_Body_Object).Object.Target.Objects[1].Value = make([]byte, 33)
			},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(session.Object).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestObject_ProtoMessage(t *testing.T) {
	var val session.Object

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
	require.IsType(t, new(protosession.SessionToken_Body_Object), c)
	oc := c.(*protosession.SessionToken_Body_Object).Object
	require.Zero(t, oc.GetVerb())
	require.Zero(t, oc.GetTarget())

	// filled
	val.SetID(anyValidSessionID)
	val.SetIssuer(anyValidUserID)
	val.SetExp(anyValidExp)
	val.SetIat(anyValidIat)
	val.SetNbf(anyValidNbf)
	val.SetAuthKey(anyValidSessionKey)
	val.ForVerb(anyValidObjectVerb)
	val.BindContainer(anyValidContainerID)
	val.LimitByObjects(anyValidObjectIDs...)
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
	require.IsType(t, new(protosession.SessionToken_Body_Object), c)
	oc = c.(*protosession.SessionToken_Body_Object).Object
	require.EqualValues(t, anyValidObjectVerb, oc.GetVerb())
	require.NotNil(t, oc.Target)
	require.Equal(t, anyValidContainerID[:], oc.Target.GetContainer().GetValue())
	mo := oc.Target.GetObjects()
	require.Len(t, mo, len(anyValidObjectIDs))
	for i := range anyValidObjectIDs {
		require.Equal(t, anyValidObjectIDs[i][:], mo[i].GetValue())
	}
}

func TestObject_Marshal(t *testing.T) {
	require.Equal(t, validBinObjectToken, validObjectToken.Marshal())
}

func TestObject_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(session.Object).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range append(invalidBinTokenCommonTestcases, invalidBinTokenTestcase{
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.SessionToken_Body_Container",
			b: []byte{10, 2, 50, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			b: []byte{10, 6, 42, 4, 18, 2, 10, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			b: []byte{10, 39, 42, 37, 18, 35, 10, 33, 10, 31, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
				21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252},
		}, invalidBinTokenTestcase{
			name: "body/context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			b: []byte{10, 41, 42, 39, 18, 37, 10, 35, 10, 33, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
				21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 1},
		}, invalidBinTokenTestcase{
			name: "body/context/object/empty value", err: "invalid context: invalid target object: invalid length 0",
			b: []byte{10, 114, 42, 112, 18, 110, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
				21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10,
				32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242,
				163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 0, 18, 34, 10, 32, 59, 5, 120, 191, 250, 61, 248,
				114, 137, 21, 229, 88, 57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165, 58, 126,
				161, 221},
		}, invalidBinTokenTestcase{
			name: "body/context/object/undersize", err: "invalid context: invalid target object: invalid length 31",
			b: []byte{10, 149, 1, 42, 146, 1, 18, 143, 1, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
				21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10, 32,
				243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20,
				26, 30, 147, 240, 79, 114, 252, 227, 18, 33, 10, 31, 47, 240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20,
				233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147, 249, 88, 184, 69, 145, 196, 127, 18, 34, 10, 32, 59, 5, 120,
				191, 250, 61, 248, 114, 137, 21, 229, 88, 57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165,
				58, 126, 161, 221},
		}, invalidBinTokenTestcase{
			name: "body/context/object/oversize", err: "invalid context: invalid target object: invalid length 33",
			b: []byte{10, 151, 1, 42, 148, 1, 18, 145, 1, 10, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168,
				21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 34, 10, 32,
				243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20,
				26, 30, 147, 240, 79, 114, 252, 227, 18, 35, 10, 33, 47, 240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20,
				233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147, 249, 88, 184, 69, 145, 196, 127, 68, 1, 18, 34, 10, 32, 59,
				5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88, 57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29, 175, 37,
				42, 165, 58, 126, 161, 221},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Object).Unmarshal(tc.b), tc.err)
			})
		}
	})
	t.Run("no container", func(t *testing.T) {
		var val session.Object
		objs := oidtest.IDs(3)
		val.LimitByObjects(objs[:2]...)
		require.True(t, val.AssertObject(objs[0]))
		require.True(t, val.AssertObject(objs[1]))
		require.False(t, val.AssertObject(objs[2]))

		b := []byte{10, 4, 42, 2, 18, 0}
		require.NoError(t, val.Unmarshal(b))
		for i := range objs {
			require.True(t, val.AssertObject(objs[i]))
		}
	})

	var val session.Object
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Zero(t, val.ID())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	require.False(t, val.AssertContainer(anyValidContainerID))
	for i := range anyValidObjectIDs {
		require.True(t, val.AssertObject(anyValidObjectIDs[i]))
	}
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	err := val.Unmarshal(validBinObjectToken)
	require.NoError(t, err)
	require.Equal(t, validObjectToken, val)
}

func TestObject_MarshalJSON(t *testing.T) {
	//nolint:staticcheck
	b, err := json.MarshalIndent(validObjectToken, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONObjectToken, string(b))
}

func TestObject_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(session.Object).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range append(invalidJSONTokenCommonTestcases, invalidJSONTokenTestcase{
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.SessionToken_Body_Container", j: `
{"body":{"container":{}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/container/empty value", err: "invalid context: invalid container ID: invalid length 0", j: `
{"body":{"object":{"target":{"container":{}}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/container/undersize", err: "invalid context: invalid container ID: invalid length 31", j: `
{"body":{"object":{"target":{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/A=="}}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/container/oversize", err: "invalid context: invalid container ID: invalid length 33", j: `
{"body":{"object":{"target":{"container":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OMB"}}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/object/empty value", err: "invalid context: invalid target object: invalid length 0", j: `
{"body":{"object":{"target":{"objects":[{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, {"value":""}, {"value":"OwV4v/o9+HKJFeVYOTFfndpPULHZOB0dryUqpTp+od0="}]}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/object/undersize", err: "invalid context: invalid target object: invalid length 31", j: `
{"body":{"object":{"target":{"objects":[{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, {"value":"L/Bd2AlAWLfGJB5TFOl3/GCrBnpzqLqT+Vi4RZHEfw=="}, {"value":"OwV4v/o9+HKJFeVYOTFfndpPULHZOB0dryUqpTp+od0="}]}}}}
`}, invalidJSONTokenTestcase{
			name: "body/context/object/oversize", err: "invalid context: invalid target object: invalid length 33", j: `
{"body":{"object":{"target":{"objects":[{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="}, {"value":"L/Bd2AlAWLfGJB5TFOl3/GCrBnpzqLqT+Vi4RZHEf0QB"}, {"value":"OwV4v/o9+HKJFeVYOTFfndpPULHZOB0dryUqpTp+od0="}]}}}}
`}) {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Object).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var val session.Object
	// zero
	require.NoError(t, val.UnmarshalJSON([]byte("{}")))
	require.Zero(t, val.ID())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidContainerVerb))
	for i := range anyValidObjectIDs {
		require.True(t, val.AssertObject(anyValidObjectIDs[i]))
	}
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	require.NoError(t, val.UnmarshalJSON([]byte(validJSONObjectToken)))
	require.Equal(t, validObjectToken, val)
}

func TestObject_AttachSignature(t *testing.T) {
	var val session.Object
	_, ok := val.Signature()
	require.False(t, ok)
	val.AttachSignature(anyValidSignature)
	sig, ok := val.Signature()
	require.True(t, ok)
	require.Equal(t, anyValidSignature, sig)
}

func TestObject_BindContainer(t *testing.T) {
	var val session.Object
	cnr1 := cidtest.ID()
	cnr2 := cidtest.OtherID(cnr1)

	require.False(t, val.AssertContainer(cnr1))
	require.False(t, val.AssertContainer(cnr2))

	val.BindContainer(cnr1)
	require.True(t, val.AssertContainer(cnr1))
	require.False(t, val.AssertContainer(cnr2))

	val.BindContainer(cnr2)
	require.False(t, val.AssertContainer(cnr1))
	require.True(t, val.AssertContainer(cnr2))

	val.BindContainer(cid.ID{})
	require.False(t, val.AssertContainer(cnr1))
	require.False(t, val.AssertContainer(cnr2))
}

func TestObject_LimitByObjects(t *testing.T) {
	var x session.Object

	obj1 := oidtest.ID()
	obj2 := oidtest.ID()
	objOther := oidtest.ID()

	require.True(t, x.AssertObject(obj1))
	require.True(t, x.AssertObject(obj2))
	require.True(t, x.AssertObject(objOther))

	x.LimitByObjects(obj1, obj2)

	require.True(t, x.AssertObject(obj1))
	require.True(t, x.AssertObject(obj2))
	require.False(t, x.AssertObject(objOther))
}

func TestObject_InvalidAt(t *testing.T) {
	testValidAt(t, new(session.Object))
}

func TestObject_ID(t *testing.T) {
	testTokenID(t, session.Object{})
}

func TestObject_SetAuthKey(t *testing.T) {
	testSetAuthKey(t, (*session.Object).SetAuthKey, session.Object.AssertAuthKey, session.Object.AuthKeyBytes)
}

func TestObject_ForVerb(t *testing.T) {
	var val session.Object
	const verb1 = anyValidObjectVerb
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

func TestObject_Issuer(t *testing.T) {
	testTokenIssuer(t, session.Object{})
}

func TestObject_Sign(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		require.Error(t, new(session.Object).Sign(usertest.FailSigner(usertest.User())))
		require.ErrorIs(t, new(session.Object).Sign(user.NewSigner(neofscryptotest.Signer(), user.ID{})), user.ErrZeroID)
	})

	ecdsaPriv := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(),
			X: new(big.Int).SetBytes([]byte{62, 189, 227, 96, 231, 242, 24, 64, 42, 170, 29, 55, 182, 194,
				249, 108, 30, 148, 108, 174, 30, 231, 53, 68, 115, 29, 241, 13, 51, 25, 155, 43}),
			Y: new(big.Int).SetBytes([]byte{136, 146, 121, 11, 234, 137, 251, 64, 44, 241, 84, 74, 155, 77, 39,
				139, 155, 185, 229, 26, 216, 16, 7, 91, 103, 247, 239, 154, 86, 178, 10, 26}),
		},
		D: new(big.Int).SetBytes([]byte{163, 20, 59, 38, 227, 11, 133, 215, 52, 179, 128, 186, 160, 119, 108,
			250, 126, 175, 247, 137, 208, 141, 168, 209, 28, 64, 224, 13, 96, 178, 158, 181}),
	}

	rfc6979Sig := []byte{120, 242, 0, 205, 206, 237, 230, 3, 144, 134, 159, 250, 242, 153, 33, 45, 166, 75, 215,
		22, 38, 221, 241, 21, 47, 151, 18, 53, 98, 46, 2, 88, 51, 231, 141, 127, 121, 192, 187, 102, 214, 17, 57,
		220, 153, 70, 50, 150, 251, 126, 101, 121, 154, 94, 170, 140, 153, 75, 221, 192, 85, 21, 95, 103}

	var o session.Object
	validObjectToken.CopyTo(&o)

	testSignCDSA(t, ecdsaPriv, anyValidUserID, &o, validSignedObjectToken, rfc6979Sig)
	testSetSignatureECDSA(t, ecdsaPriv, &o, validSignedObjectToken, rfc6979Sig)
}

func TestObject_VerifySignature(t *testing.T) {
	// keys used for this test
	// ecdsa.PrivateKey{
	// 	PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(),
	// 		X: new(big.Int).SetBytes([]byte{207, 151, 62, 248, 240, 176, 177, 121, 222, 235, 70, 179, 253, 248, 9, 5,
	// 			100, 217, 185, 205, 124, 56, 77, 135, 72, 1, 244, 193, 84, 254, 145, 119}),
	// 		Y: new(big.Int).SetBytes([]byte{190, 106, 150, 193, 105, 247, 90, 245, 136, 42, 104, 150, 197, 89, 78, 3, 46,
	// 			26, 211, 8, 173, 235, 182, 244, 154, 221, 218, 202, 181, 222, 125, 106}),
	// 	},
	// 	D: new(big.Int).SetBytes([]byte{1, 119, 135, 48, 159, 121, 104, 170, 177, 137, 6, 102, 120, 73, 198, 228, 111,
	// 		164, 40, 172, 215, 106, 110, 136, 55, 60, 101, 227, 141, 97, 125, 147}),
	// }
	pub := []byte{2, 207, 151, 62, 248, 240, 176, 177, 121, 222, 235, 70, 179, 253, 248, 9, 5, 100, 217, 185, 205,
		124, 56, 77, 135, 72, 1, 244, 193, 84, 254, 145, 119}
	var sig neofscrypto.Signature

	var o session.Object
	for i, tc := range []struct {
		scheme neofscrypto.Scheme
		sig    []byte // of validObjectToken
	}{
		{scheme: neofscrypto.ECDSA_SHA512, sig: []byte{
			4, 33, 70, 101, 200, 184, 171, 87, 235, 229, 195, 179, 29, 179, 93, 46, 128, 73, 53, 190, 109, 133, 103, 147,
			67, 228, 232, 117, 191, 141, 1, 56, 75, 250, 5, 191, 220, 76, 115, 196, 185, 198, 27, 135, 124, 11, 177, 43, 3,
			226, 75, 229, 168, 33, 106, 55, 42, 49, 173, 123, 25, 96, 167, 249, 240,
		}},
		{scheme: neofscrypto.ECDSA_DETERMINISTIC_SHA256, sig: []byte{
			59, 5, 35, 98, 240, 29, 13, 237, 176, 180, 209, 78, 241, 163, 190, 63, 180, 130, 200, 242, 221, 190, 130,
			103, 69, 110, 197, 162, 117, 181, 55, 69, 78, 229, 181, 196, 239, 199, 179, 225, 60, 46, 35, 177, 84, 121, 248,
			32, 167, 81, 97, 227, 72, 8, 197, 59, 9, 16, 92, 5, 229, 170, 58, 76,
		}},
		{scheme: neofscrypto.ECDSA_WALLETCONNECT, sig: []byte{
			108, 83, 164, 199, 35, 37, 175, 2, 218, 108, 86, 147, 196, 234, 34, 76, 65, 172, 55, 101, 217, 75, 144, 2, 77,
			232, 196, 36, 124, 28, 76, 229, 139, 50, 152, 100, 118, 107, 12, 106, 200, 96, 40, 123, 178, 12, 254, 75,
			240, 74, 240, 88, 93, 94, 62, 208, 54, 35, 218, 208, 93, 237, 176, 77, 75, 230, 137, 194, 84, 156, 199, 0,
			167, 132, 120, 106, 110, 80, 67, 58,
		}},
	} {
		sig.SetScheme(tc.scheme)
		validObjectToken.CopyTo(&o)
		sig.SetPublicKeyBytes(pub)
		sig.SetValue(tc.sig)
		o.AttachSignature(sig)
		require.True(t, o.VerifySignature(), i)
		for k := range pub {
			pubCp := bytes.Clone(pub)
			pubCp[k]++
			sig.SetPublicKeyBytes(pubCp)
			o.AttachSignature(sig)
			require.False(t, o.VerifySignature(), i)
		}
		for k := range tc.sig {
			sigBytesCp := bytes.Clone(tc.sig)
			sigBytesCp[k]++
			sig.SetValue(sigBytesCp)
			o.AttachSignature(sig)
			require.False(t, o.VerifySignature(), i)
		}
	}
}

func TestObject_SignedData(t *testing.T) {
	require.Equal(t, validSignedObjectToken, validObjectToken.SignedData())
}

func TestObject_UnmarshalSignedData(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(session.Object).UnmarshalSignedData([]byte("Hello, world!"))
			require.ErrorContains(t, err, "decode body")
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range append(invalidSignedTokenCommonTestcases, invalidBinTokenTestcase{
			name: "body/context/wrong oneof", err: "invalid context: invalid context *session.SessionToken_Body_Container",
			b: []byte{50, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/empty value", err: "invalid context: invalid container ID: invalid length 0",
			b: []byte{42, 4, 18, 2, 10, 0},
		}, invalidBinTokenTestcase{
			name: "body/context/container/undersize", err: "invalid context: invalid container ID: invalid length 31",
			b: []byte{42, 37, 18, 35, 10, 33, 10, 31, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252},
		}, invalidBinTokenTestcase{
			name: "body/context/container/oversize", err: "invalid context: invalid container ID: invalid length 33",
			b: []byte{42, 39, 18, 37, 10, 35, 10, 33, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 1},
		}, invalidBinTokenTestcase{
			name: "body/context/object/empty value", err: "invalid context: invalid target object: invalid length 0",
			b: []byte{42, 76, 18, 74, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 0, 18, 34, 10, 32,
				59, 5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88, 57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29,
				175, 37, 42, 165, 58, 126, 161, 221},
		}, invalidBinTokenTestcase{
			name: "body/context/object/undersize", err: "invalid context: invalid target object: invalid length 31",
			b: []byte{42, 109, 18, 107, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 33, 10, 31, 47,
				240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20, 233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147,
				249, 88, 184, 69, 145, 196, 127, 18, 34, 10, 32, 59, 5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88,
				57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165, 58, 126, 161, 221},
		}, invalidBinTokenTestcase{
			name: "body/context/object/oversize", err: "invalid context: invalid target object: invalid length 33",
			b: []byte{42, 111, 18, 109, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254,
				62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 18, 35, 10, 33, 47,
				240, 93, 216, 9, 64, 88, 183, 198, 36, 30, 83, 20, 233, 119, 252, 96, 171, 6, 122, 115, 168, 186, 147,
				249, 88, 184, 69, 145, 196, 127, 68, 1, 18, 34, 10, 32, 59, 5, 120, 191, 250, 61, 248, 114, 137, 21, 229, 88,
				57, 49, 95, 157, 218, 79, 80, 177, 217, 56, 29, 29, 175, 37, 42, 165, 58, 126, 161, 221},
		}) {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Object).UnmarshalSignedData(tc.b), tc.err)
			})
		}
	})

	var val session.Object
	// zero
	require.NoError(t, val.UnmarshalSignedData(nil))
	require.Zero(t, val.ID())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.False(t, val.AssertVerb(anyValidObjectVerb))
	for i := range anyValidObjectIDs {
		require.True(t, val.AssertObject(anyValidObjectIDs[i]))
	}

	// filled
	err := val.UnmarshalSignedData(validSignedObjectToken)
	require.NoError(t, err)
	require.Equal(t, validSignedObjectToken, val.SignedData())
}

func TestObject_SetExp(t *testing.T) {
	testLifetimeClaim(t, session.Object.Exp, (*session.Object).SetExp)
}

func TestObject_SetIat(t *testing.T) {
	testLifetimeClaim(t, session.Object.Iat, (*session.Object).SetIat)
}

func TestObject_SetNbf(t *testing.T) {
	testLifetimeClaim(t, session.Object.Nbf, (*session.Object).SetNbf)
}

func TestObject_IssuerPublicKeyBytes(t *testing.T) {
	var val session.Object
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

func TestObject_ExpiredAt(t *testing.T) {
	var val session.Object
	const epoch = 13

	require.False(t, val.ExpiredAt(0))
	require.True(t, val.ExpiredAt(1))
	require.True(t, val.ExpiredAt(epoch))

	val.SetExp(epoch)
	require.False(t, val.ExpiredAt(0))
	require.False(t, val.ExpiredAt(epoch))
	require.True(t, val.ExpiredAt(epoch+1))
}

func TestObject_SetAuthKeyBytes(t *testing.T) {
	var val session.Object
	require.False(t, val.AssertAuthKey(anyValidSessionKey))
	require.Zero(t, val.AuthKeyBytes())

	val.SetAuthKeyBytes(anyValidSessionKeyBytes)

	require.True(t, val.AssertAuthKey(anyValidSessionKey))
	require.Equal(t, anyValidSessionKeyBytes, val.AuthKeyBytes())
}
