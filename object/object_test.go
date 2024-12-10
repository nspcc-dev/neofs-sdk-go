package object_test

import (
	"bytes"
	"crypto/elliptic"
	"encoding/json"
	"math"
	"math/big"
	"testing"

	"github.com/google/uuid"
	apiobject "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	apisession "github.com/nspcc-dev/neofs-api-go/v2/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

var (
	anySessionIssuerPubKey = &neofsecdsa.PublicKey{
		Curve: elliptic.P256(),
		X: new(big.Int).SetBytes([]byte{154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26,
			192, 33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217}),
		Y: new(big.Int).SetBytes([]byte{94, 32, 107, 98, 243, 3, 170, 187, 6, 229, 38, 125, 17, 208, 247, 106, 3, 209, 115,
			174, 180, 192, 102, 190, 151, 10, 215, 157, 164, 219, 74, 40}),
	}
	// corresponds to anySessionIssuerPubKey.
	anySessionIssuerPubKeyBytes = []byte{2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26,
		192, 33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217}
	anyValidSessionID   = uuid.UUID{118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142, 52, 17, 144}
	anyValidObjectToken session.Object // set by init.
)

var validObject object.Object // set by init.

func init() {
	anyValidObjectToken.SetID(anyValidSessionID)
	anyValidObjectToken.SetIssuer(anyValidUsers[2])
	anyValidObjectToken.SetExp(16429376563136800338)
	anyValidObjectToken.SetIat(7956510363313998522)
	anyValidObjectToken.SetNbf(17237208928641773338)
	anyValidObjectToken.SetAuthKey(anySessionIssuerPubKey)
	anyValidObjectToken.ForVerb(1047242055)
	anyValidObjectToken.BindContainer(anyValidContainers[2])
	anyValidObjectToken.LimitByObjects(anyValidIDs[8], anyValidIDs[9])
	anyValidObjectToken.AttachSignature(neofscrypto.NewSignatureFromRawKey(1134494890, []byte("session_signer"), []byte("session_signature")))

	var par object.Object
	par.SetID(anyValidIDs[1])
	par.SetSignature(&anyValidSignatures[0])
	par.SetVersion(&anyValidVersions[0])
	par.SetContainerID(anyValidContainers[0])
	par.SetOwner(anyValidUsers[0])
	par.SetCreationEpoch(anyValidCreationEpoch)
	par.SetPayloadSize(anyValidPayloadSize)
	par.SetPayloadChecksum(anyValidChecksums[0])
	par.SetType(anyValidType)
	par.SetPayloadHomomorphicHash(anyValidChecksums[1])
	par.SetAttributes(
		*object.NewAttribute("par_attr_key1", "par_attr_val1"),
		*object.NewAttribute("__NEOFS__EXPIRATION_EPOCH", "14208497712700580130"),
		*object.NewAttribute("par_attr_key2", "par_attr_val2"),
	)

	validObject.SetID(anyValidIDs[0])
	validObject.SetSignature(&anyValidSignatures[1])
	validObject.SetPayload(anyValidRegularPayload)
	validObject.SetVersion(&anyValidVersions[1])
	validObject.SetContainerID(anyValidContainers[1])
	validObject.SetOwner(anyValidUsers[1])
	validObject.SetCreationEpoch(anyValidCreationEpoch + 1)
	validObject.SetPayloadSize(anyValidPayloadSize + 1)
	validObject.SetPayloadChecksum(anyValidChecksums[2])
	validObject.SetType(anyValidType + 1)
	validObject.SetPayloadHomomorphicHash(anyValidChecksums[3])
	validObject.SetSessionToken(&anyValidObjectToken)
	validObject.SetAttributes(
		*object.NewAttribute("attr_key1", "attr_val1"),
		*object.NewAttribute("__NEOFS__EXPIRATION_EPOCH", "8516691293958955670"),
		*object.NewAttribute("attr_key2", "attr_val2"),
	)
	validObject.SetPreviousID(anyValidIDs[2])
	validObject.SetParent(&par)
	validObject.SetChildren(anyValidIDs[3], anyValidIDs[4], anyValidIDs[5])
	validObject.SetSplitID(anyValidSplitID)
	validObject.SetFirstID(anyValidIDs[6])
}

// corresponds to validObject.
var validObjectID = oid.ID{104, 79, 12, 22, 52, 73, 44, 178, 246, 55, 17, 33, 163, 133, 252, 128, 121, 232, 127, 223, 91, 227,
	7, 83, 223, 223, 186, 153, 69, 172, 80, 230}

// corresponds to validObject.
var validBinObject = []byte{
	10, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70,
	246, 8, 139, 247, 174, 53, 60, 18, 20, 10, 5, 112, 117, 98, 95, 50, 18, 5, 115, 105, 103, 95, 50, 24, 171, 178, 212, 208, 4, 26,
	151, 8, 10, 11, 8, 209, 134, 217, 250, 1, 16, 202, 208, 129, 82, 18, 34, 10, 32, 217, 213, 19, 152, 91, 248, 2, 180, 17, 177,
	248, 226, 163, 200, 56, 31, 123, 24, 182, 144, 148, 180, 248, 192, 155, 253, 104, 220, 69, 102, 174, 5, 26, 27, 10, 25, 53,
	214, 113, 220, 69, 70, 98, 242, 115, 99, 188, 86, 53, 223, 243, 238, 11, 245, 251, 169, 115, 202, 247, 184, 221, 32, 158,
	188, 250, 184, 255, 160, 255, 210, 183, 1, 40, 189, 238, 172, 200, 143, 221, 203, 248, 76, 50, 17, 8, 193, 243, 161, 60, 18,
	10, 99, 104, 101, 99, 107, 115, 117, 109, 95, 51, 56, 224, 137, 251, 224, 7, 66, 18, 8, 229, 198, 224, 221, 3, 18, 10, 99, 104,
	101, 99, 107, 115, 117, 109, 95, 52, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
	52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151, 159, 221, 73, 224,
	229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155,
	239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109,
	3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18,
	108, 10, 34, 10, 32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
	29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40,
	32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68, 233,
	22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14,
	115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 97, 116,
	117, 114, 101, 24, 170, 137, 252, 156, 4, 82, 22, 10, 9, 97, 116, 116, 114, 95, 107, 101, 121, 49, 18, 9, 97, 116, 116, 114, 95, 118,
	97, 108, 49, 82, 48, 10, 25, 95, 95, 78, 69, 79, 70, 83, 95, 95, 69, 88, 80, 73, 82, 65, 84, 73, 79, 78, 95, 69, 80, 79, 67, 72,
	18, 19, 56, 53, 49, 54, 54, 57, 49, 50, 57, 51, 57, 53, 56, 57, 53, 53, 54, 55, 48, 82, 22, 10, 9, 97, 116, 116, 114, 95, 107, 101,
	121, 50, 18, 9, 97, 116, 116, 114, 95, 118, 97, 108, 50, 90, 135, 4, 10, 34, 10, 32, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47,
	65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 18, 34, 10, 32, 206, 228, 247,
	217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161,
	202, 197, 242, 26, 20, 10, 5, 112, 117, 98, 95, 49, 18, 5, 115, 105, 103, 95, 49, 24, 184, 132, 246, 224, 4, 34, 132, 2, 10,
	11, 8, 167, 167, 171, 42, 16, 221, 138, 221, 194, 7, 18, 34, 10, 32, 245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20,
	135, 96, 204, 179, 93, 183, 250, 180, 255, 162, 182, 222, 220, 99, 125, 136, 117, 206, 34, 26, 27, 10, 25, 53, 59, 15, 5, 52,
	131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 153, 32, 157, 188, 250, 184, 255, 160,
	255, 210, 183, 1, 40, 188, 238, 172, 200, 143, 221, 203, 248, 76, 50, 18, 8, 222, 213, 182, 173, 7, 18, 10, 99, 104, 101, 99,
	107, 115, 117, 109, 95, 49, 56, 223, 137, 251, 224, 7, 66, 18, 8, 240, 184, 222, 148, 7, 18, 10, 99, 104, 101, 99, 107, 115, 117,
	109, 95, 50, 82, 30, 10, 13, 112, 97, 114, 95, 97, 116, 116, 114, 95, 107, 101, 121, 49, 18, 13, 112, 97, 114, 95, 97, 116, 116, 114, 95,
	118, 97, 108, 49, 82, 49, 10, 25, 95, 95, 78, 69, 79, 70, 83, 95, 95, 69, 88, 80, 73, 82, 65, 84, 73, 79, 78, 95, 69, 80, 79, 67,
	72, 18, 20, 49, 52, 50, 48, 56, 52, 57, 55, 55, 49, 50, 55, 48, 48, 53, 56, 48, 49, 51, 48, 82, 30, 10, 13, 112, 97, 114, 95, 97,
	116, 116, 114, 95, 107, 101, 121, 50, 18, 13, 112, 97, 114, 95, 97, 116, 116, 114, 95, 118, 97, 108, 50, 42, 34, 10, 32, 173, 160, 45,
	58, 200, 168, 116, 142, 235, 209, 231, 80, 235, 186, 6, 132, 99, 95, 14, 39, 237, 139, 87, 66, 244, 72, 96, 69, 13, 83, 81, 172,
	42, 34, 10, 32, 238, 167, 85, 68, 91, 254, 165, 81, 182, 145, 16, 91, 35, 224, 17, 46, 164, 138, 86, 50, 196, 148, 215, 210, 247,
	29, 44, 153, 203, 20, 137, 169, 42, 34, 10, 32, 226, 165, 123, 249, 146, 166, 187, 202, 244, 12, 156, 43, 207, 204, 40, 230,
	145, 34, 212, 152, 148, 112, 44, 21, 195, 207, 249, 112, 34, 81, 145, 194, 50, 16, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32,
	226, 201, 206, 196, 147, 41, 58, 34, 10, 32, 119, 231, 221, 167, 7, 141, 50, 77, 49, 23, 194, 169, 82, 56, 150, 162, 103, 20,
	124, 174, 16, 64, 169, 172, 79, 238, 242, 146, 87, 88, 5, 147, 34, 13, 72, 101, 108, 108, 111, 44, 32, 119, 111, 114, 108, 100, 33,
}

// corresponds to validObject.
var validJSONObject = `
{
 "objectID": {
  "value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="
 },
 "signature": {
  "key": "cHViXzI=",
  "signature": "c2lnXzI=",
  "scheme": 1242896683
 },
 "header": {
  "version": {
   "major": 525747025,
   "minor": 171993162
  },
  "containerID": {
   "value": "2dUTmFv4ArQRsfjio8g4H3sYtpCUtPjAm/1o3EVmrgU="
  },
  "ownerID": {
   "value": "NdZx3EVGYvJzY7xWNd/z7gv1+6lzyve43Q=="
  },
  "creationEpoch": "13233261290750647838",
  "payloadLength": "5544264194415343421",
  "payloadHash": {
   "type": 126384577,
   "sum": "Y2hlY2tzdW1fMw=="
  },
  "objectType": 2082391264,
  "homomorphicHash": {
   "type": 1001923429,
   "sum": "Y2hlY2tzdW1fNA=="
  },
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "verb": "VERB_UNSPECIFIED",
     "target": {
      "container": {
       "value": "h1mV27nR6Yng041Gwc34/uIecrH1qx1a1A8zVo5lm40="
      },
      "objects": [
       {
        "value": "OattKWkZkuCkWT2ys55DKCCaeq513Yqoh5XuPUQ6Ir0="
       },
       {
        "value": "bulm6IhE6RaeZDEUtV/bjzX67XFAGTALNs84YmOIzxU="
       }
      ]
     }
    }
   },
   "signature": {
    "key": "c2Vzc2lvbl9zaWduZXI=",
    "signature": "c2Vzc2lvbl9zaWduYXR1cmU=",
    "scheme": 1134494890
   }
  },
  "attributes": [
   {
    "key": "attr_key1",
    "value": "attr_val1"
   },
   {
    "key": "__NEOFS__EXPIRATION_EPOCH",
    "value": "8516691293958955670"
   },
   {
    "key": "attr_key2",
    "value": "attr_val2"
   }
  ],
  "split": {
   "parent": {
    "value": "5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+s="
   },
   "previous": {
    "value": "zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="
   },
   "parentSignature": {
    "key": "cHViXzE=",
    "signature": "c2lnXzE=",
    "scheme": 1277002296
   },
   "parentHeader": {
    "version": {
     "major": 88789927,
     "minor": 2018985309
    },
    "containerID": {
     "value": "9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1ziI="
    },
    "ownerID": {
     "value": "NTsPBTSD/8YIYim45e2M1zSB09Zake2JmQ=="
    },
    "creationEpoch": "13233261290750647837",
    "payloadLength": "5544264194415343420",
    "payloadHash": {
     "type": 1974315742,
     "sum": "Y2hlY2tzdW1fMQ=="
    },
    "objectType": 2082391263,
    "homomorphicHash": {
     "type": 1922538608,
     "sum": "Y2hlY2tzdW1fMg=="
    },
    "sessionToken": null,
    "attributes": [
     {
      "key": "par_attr_key1",
      "value": "par_attr_val1"
     },
     {
      "key": "__NEOFS__EXPIRATION_EPOCH",
      "value": "14208497712700580130"
     },
     {
      "key": "par_attr_key2",
      "value": "par_attr_val2"
     }
    ],
    "split": null
   },
   "children": [
    {
     "value": "raAtOsiodI7r0edQ67oGhGNfDifti1dC9EhgRQ1TUaw="
    },
    {
     "value": "7qdVRFv+pVG2kRBbI+ARLqSKVjLElNfS9x0smcsUiak="
    },
    {
     "value": "4qV7+ZKmu8r0DJwrz8wo5pEi1JiUcCwVw8/5cCJRkcI="
    }
   ],
   "splitID": "4IQDUCAsRbi5IOLJzsSTKQ==",
   "first": {
    "value": "d+fdpweNMk0xF8KpUjiWomcUfK4QQKmsT+7ykldYBZM="
   }
  }
 },
 "payload": "SGVsbG8sIHdvcmxkIQ=="
}
`

func TestInitCreation(t *testing.T) {
	var o object.Object
	cnr := cidtest.ID()
	own := usertest.ID()

	o.InitCreation(object.RequiredFields{
		Container: cnr,
		Owner:     own,
	})

	cID := o.GetContainerID()
	require.Equal(t, cnr, cID)
	require.Equal(t, own, o.Owner())
}

func TestObject_SetID(t *testing.T) {
	var obj object.Object
	require.True(t, obj.GetID().IsZero())
	_, ok := obj.ID()
	require.False(t, ok)

	id1 := oidtest.ID()
	obj.SetID(id1)
	require.Equal(t, id1, obj.GetID())
	res, ok := obj.ID()
	require.True(t, ok)
	require.Equal(t, id1, res)

	id2 := oidtest.OtherID(id1)
	obj.SetID(id2)
	require.Equal(t, id2, obj.GetID())
	res, ok = obj.ID()
	require.True(t, ok)
	require.Equal(t, id2, res)

	// reset
	require.False(t, obj.GetID().IsZero())
	obj.ResetID()
	require.True(t, obj.GetID().IsZero())
	_, ok = obj.ID()
	require.False(t, ok)
}

func TestObject_SetSignature(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.Signature())

	obj.SetSignature(&anyValidSignatures[0])
	sig := obj.Signature()
	require.NotNil(t, sig)
	require.Equal(t, anyValidSignatures[0], *sig)

	obj.SetSignature(&anyValidSignatures[1])
	require.Equal(t, anyValidSignatures[1], *obj.Signature())
}

func TestObject_SetPayload(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.Payload())

	obj.SetPayload(anyValidRegularPayload)
	require.Equal(t, anyValidRegularPayload, obj.Payload())

	otherPayload := append(bytes.Clone(anyValidRegularPayload), "_other"...)
	obj.SetPayload(otherPayload)
	require.Equal(t, otherPayload, obj.Payload())
}

func TestObject_SetVersion(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.Version())

	obj.SetVersion(&anyValidVersions[0])
	require.Equal(t, anyValidVersions[0], *obj.Version())

	obj.SetVersion(&anyValidVersions[1])
	require.Equal(t, anyValidVersions[1], *obj.Version())
}

func TestObject_SetPayloadSize(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.PayloadSize())

	obj.SetPayloadSize(anyValidPayloadSize)
	require.EqualValues(t, anyValidPayloadSize, obj.PayloadSize())

	obj.SetPayloadSize(anyValidPayloadSize + 1)
	require.EqualValues(t, anyValidPayloadSize+1, obj.PayloadSize())
}

func TestObject_SetContainerID(t *testing.T) {
	var obj object.Object
	require.True(t, obj.GetContainerID().IsZero())
	_, ok := obj.ContainerID()
	require.False(t, ok)

	cnr1 := cidtest.ID()
	obj.SetContainerID(cnr1)
	require.Equal(t, cnr1, obj.GetContainerID())
	res, ok := obj.ContainerID()
	require.True(t, ok)
	require.Equal(t, cnr1, res)

	cnr2 := cidtest.OtherID(cnr1)
	obj.SetContainerID(cnr2)
	require.Equal(t, cnr2, obj.GetContainerID())
	res, ok = obj.ContainerID()
	require.True(t, ok)
	require.Equal(t, cnr2, res)
}

func TestObject_SetOwner(t *testing.T) {
	var obj object.Object
	require.True(t, obj.Owner().IsZero())

	usr1 := usertest.ID()
	obj.SetOwner(usr1)
	require.Equal(t, usr1, obj.Owner())

	usr2 := usertest.OtherID(usr1)
	obj.SetOwner(usr2)
	require.Equal(t, usr2, obj.Owner())
}

func TestObject_SetOwnerID(t *testing.T) {
	var obj object.Object
	require.True(t, obj.OwnerID().IsZero())

	usr1 := usertest.ID()
	obj.SetOwnerID(&usr1)
	require.Equal(t, usr1, *obj.OwnerID())

	usr2 := usertest.OtherID(usr1)
	obj.SetOwnerID(&usr2)
	require.Equal(t, usr2, *obj.OwnerID())
}

func TestObject_SetCreationEpoch(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.CreationEpoch())

	obj.SetCreationEpoch(anyValidCreationEpoch)
	require.EqualValues(t, anyValidCreationEpoch, obj.CreationEpoch())

	obj.SetCreationEpoch(anyValidCreationEpoch + 1)
	require.EqualValues(t, anyValidCreationEpoch+1, obj.CreationEpoch())
}

func TestObject_SetPayloadChecksum(t *testing.T) {
	var obj object.Object
	_, ok := obj.PayloadChecksum()
	require.False(t, ok)

	obj.SetPayloadChecksum(anyValidChecksums[0])
	res, ok := obj.PayloadChecksum()
	require.True(t, ok)
	require.Equal(t, anyValidChecksums[0], res)

	obj.SetPayloadChecksum(anyValidChecksums[1])
	res, ok = obj.PayloadChecksum()
	require.True(t, ok)
	require.Equal(t, anyValidChecksums[1], res)
}

func TestObject_SetPayloadHomomorphicHash(t *testing.T) {
	var obj object.Object
	_, ok := obj.PayloadHomomorphicHash()
	require.False(t, ok)

	obj.SetPayloadHomomorphicHash(anyValidChecksums[0])
	res, ok := obj.PayloadHomomorphicHash()
	require.True(t, ok)
	require.Equal(t, anyValidChecksums[0], res)

	obj.SetPayloadHomomorphicHash(anyValidChecksums[1])
	res, ok = obj.PayloadHomomorphicHash()
	require.True(t, ok)
	require.Equal(t, anyValidChecksums[1], res)
}

func TestObject_SetAttributes(t *testing.T) {
	var obj object.Object
	require.Empty(t, obj.Attributes())
	require.Empty(t, obj.UserAttributes())

	a1 := *object.NewAttribute("k1", "v1")
	sa1 := *object.NewAttribute("__NEOFS__sk1", "sv1")
	a2 := *object.NewAttribute("k2", "v2")
	sa2 := *object.NewAttribute("__NEOFS__sk2", "sv2")

	obj.SetAttributes(a1, sa1, a2, sa2)
	require.Equal(t, []object.Attribute{a1, sa1, a2, sa2}, obj.Attributes())
	require.Equal(t, []object.Attribute{a1, a2}, obj.UserAttributes())
}

func TestObject_SetPreviousID(t *testing.T) {
	var obj object.Object
	require.True(t, obj.GetPreviousID().IsZero())
	_, ok := obj.PreviousID()
	require.False(t, ok)

	id1 := oidtest.ID()
	obj.SetPreviousID(id1)
	require.Equal(t, id1, obj.GetPreviousID())
	res, ok := obj.PreviousID()
	require.True(t, ok)
	require.Equal(t, id1, res)

	id2 := oidtest.OtherID(id1)
	obj.SetPreviousID(id2)
	require.Equal(t, id2, obj.GetPreviousID())
	res, ok = obj.PreviousID()
	require.True(t, ok)
	require.Equal(t, id2, res)

	// reset
	require.False(t, obj.GetPreviousID().IsZero())
	obj.ResetPreviousID()
	require.True(t, obj.GetPreviousID().IsZero())
	_, ok = obj.PreviousID()
	require.False(t, ok)
}

func TestObject_SetChildren(t *testing.T) {
	var obj object.Object
	require.Empty(t, obj.Children())

	ids := oidtest.IDs(5)
	obj.SetChildren(ids...)
	require.Equal(t, ids, obj.Children())

	idsOther := oidtest.IDs(7)
	obj.SetChildren(idsOther...)
	require.Equal(t, idsOther, obj.Children())
}

func TestObject_SetFirstID(t *testing.T) {
	var obj object.Object
	require.True(t, obj.GetFirstID().IsZero())
	_, ok := obj.FirstID()
	require.False(t, ok)

	id1 := oidtest.ID()
	obj.SetFirstID(id1)
	require.Equal(t, id1, obj.GetFirstID())
	res, ok := obj.FirstID()
	require.True(t, ok)
	require.Equal(t, id1, res)

	id2 := oidtest.OtherID(id1)
	obj.SetFirstID(id2)
	require.Equal(t, id2, obj.GetFirstID())
	res, ok = obj.FirstID()
	require.True(t, ok)
	require.Equal(t, id2, res)
}

func TestObject_SetSplitID(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.SplitID())

	id1 := objecttest.SplitID()
	obj.SetSplitID(&id1)
	require.Equal(t, id1, *obj.SplitID())

	id2 := objecttest.SplitID()
	obj.SetSplitID(&id2)
	require.Equal(t, id2, *obj.SplitID())
}

func TestObject_SetParentID(t *testing.T) {
	var obj object.Object
	require.True(t, obj.GetParentID().IsZero())
	_, ok := obj.ParentID()
	require.False(t, ok)

	id1 := oidtest.ID()
	obj.SetParentID(id1)
	require.Equal(t, id1, obj.GetParentID())
	res, ok := obj.ParentID()
	require.True(t, ok)
	require.Equal(t, id1, res)

	id2 := oidtest.OtherID(id1)
	obj.SetParentID(id2)
	require.Equal(t, id2, obj.GetParentID())
	res, ok = obj.ParentID()
	require.True(t, ok)
	require.Equal(t, id2, res)

	// reset
	require.False(t, obj.GetParentID().IsZero())
	obj.ResetParentID()
	require.True(t, obj.GetParentID().IsZero())
	_, ok = obj.ParentID()
	require.False(t, ok)
}

func TestObject_SetParent(t *testing.T) {
	var obj object.Object
	require.Nil(t, obj.Parent())

	par := objecttest.Object()
	parHdr := *par.CutPayload()
	obj.SetParent(&parHdr)
	require.Equal(t, parHdr, *obj.Parent())

	parOther := objecttest.Object()
	parHdrOther := *parOther.CutPayload()
	obj.SetParent(&parOther)
	require.Equal(t, parHdrOther, *obj.Parent())
}

func TestObject_SetSessionToken(t *testing.T) {
	var obj object.Object
	require.Nil(t, obj.SessionToken())

	s := sessiontest.ObjectSigned(usertest.User())
	obj.SetSessionToken(&s)
	require.Equal(t, s, *obj.SessionToken())

	sOther := sessiontest.ObjectSigned(usertest.User())
	obj.SetSessionToken(&sOther)
	require.Equal(t, sOther, *obj.SessionToken())
}

func TestObject_Attributes(t *testing.T) {
	var obj object.Object
	require.Zero(t, obj.Type())

	obj.SetType(anyValidType)
	require.Equal(t, anyValidType, obj.Type())

	obj.SetType(anyValidType + 1)
	require.Equal(t, anyValidType+1, obj.Type())
}

func TestObject_CutPayload(t *testing.T) {
	obj := objecttest.Object()
	payload := obj.Payload()

	cut := obj.CutPayload()
	require.Zero(t, cut.Payload())
	cut.SetPayload(payload)
	require.Equal(t, obj, *cut)
}

func TestObject_HasParent(t *testing.T) {
	for i, tc := range []func(*object.Object){
		func(obj *object.Object) { obj.SetParentID(oidtest.ID()) },
		func(obj *object.Object) { obj.SetPreviousID(oidtest.ID()) },
		func(obj *object.Object) { obj.SetChildren(oidtest.IDs(3)...) },
		func(obj *object.Object) { obj.SetSplitID(anyValidSplitID) },
		func(obj *object.Object) { obj.SetFirstID(oidtest.ID()) },
		func(obj *object.Object) { par := objecttest.Object(); obj.SetParent(&par) },
	} {
		var obj object.Object
		require.False(t, obj.HasParent())
		tc(&obj)
		require.True(t, obj.HasParent(), i)
	}
}

func assertNoSplitFields(t testing.TB, obj object.Object) {
	require.True(t, obj.GetParentID().IsZero())
	require.True(t, obj.GetPreviousID().IsZero())
	require.Zero(t, obj.Parent())
	require.Empty(t, obj.Children())
	require.Zero(t, obj.SplitID())
	require.Zero(t, obj.SplitID())
	require.True(t, obj.GetFirstID().IsZero())
}

func TestObject_ResetRelations(t *testing.T) {
	var obj object.Object
	obj.SetParentID(oidtest.ID())
	obj.SetPreviousID(oidtest.ID())
	par := objecttest.Object()
	obj.SetParent(&par)
	obj.SetChildren(oidtest.IDs(3)...)
	obj.SetSplitID(anyValidSplitID)
	obj.SetFirstID(oidtest.ID())

	require.False(t, obj.GetParentID().IsZero())
	require.False(t, obj.GetPreviousID().IsZero())
	require.NotZero(t, obj.Parent())
	require.NotEmpty(t, obj.Children())
	require.NotZero(t, obj.SplitID())
	require.NotZero(t, obj.SplitID())
	require.False(t, obj.GetFirstID().IsZero())

	obj.ResetRelations()
	assertNoSplitFields(t, obj)
}

func TestObject_InitRelations(t *testing.T) {
	var obj object.Object
	assertNoSplitFields(t, obj)
	require.False(t, obj.HasParent())

	obj.InitRelations()
	assertNoSplitFields(t, obj)
	require.True(t, obj.HasParent())
}

func TestObject_ReadFromV2(t *testing.T) {
	var pv refs.Version
	pv.SetMajor(88789927)
	pv.SetMinor(2018985309)
	var pcs refs.Checksum
	pcs.SetType(1974315742)
	pcs.SetSum([]byte("checksum_1"))
	var phcs refs.Checksum
	phcs.SetType(1922538608)
	phcs.SetSum([]byte("checksum_2"))
	pas := make([]apiobject.Attribute, 3)
	pas[0].SetKey("par_attr_key1")
	pas[0].SetValue("par_attr_val1")
	pas[1].SetKey("__NEOFS__EXPIRATION_EPOCH")
	pas[1].SetValue("14208497712700580130")
	pas[2].SetKey("par_attr_key2")
	pas[2].SetValue("par_attr_val2")
	var ph apiobject.Header
	ph.SetVersion(&pv)
	ph.SetContainerID(protoContainerIDFromBytes(anyValidContainers[0][:]))
	ph.SetOwnerID(protoUserIDFromBytes(anyValidUsers[0][:]))
	ph.SetCreationEpoch(anyValidCreationEpoch)
	ph.SetPayloadLength(anyValidPayloadSize)
	ph.SetPayloadHash(&pcs)
	ph.SetObjectType(apiobject.Type(anyValidType))
	ph.SetHomomorphicHash(&phcs)
	ph.SetAttributes(pas)

	var mtl apisession.TokenLifetime
	mtl.SetExp(16429376563136800338)
	mtl.SetIat(7956510363313998522)
	mtl.SetNbf(17237208928641773338)
	var mtc apisession.ObjectSessionContext
	mtc.SetVerb(1047242055)
	mtc.SetTarget(protoContainerIDFromBytes(anyValidContainers[2][:]),
		*protoIDFromBytes(anyValidIDs[8][:]), *protoIDFromBytes(anyValidIDs[9][:]))
	var mtb apisession.TokenBody
	mtb.SetID(anyValidSessionID[:])
	mtb.SetOwnerID(protoUserIDFromBytes(anyValidUsers[2][:]))
	mtb.SetLifetime(&mtl)
	mtb.SetSessionKey(anySessionIssuerPubKeyBytes)
	mtb.SetContext(&mtc)
	var mts refs.Signature
	mts.SetScheme(1134494890)
	mts.SetKey([]byte("session_signer"))
	mts.SetSign([]byte("session_signature"))
	var mt apisession.Token
	mt.SetBody(&mtb)
	mt.SetSignature(&mts)

	var mv refs.Version
	mv.SetMajor(525747025)
	mv.SetMinor(171993162)
	var cs refs.Checksum
	cs.SetType(126384577)
	cs.SetSum([]byte("checksum_3"))
	var hcs refs.Checksum
	hcs.SetType(1001923429)
	hcs.SetSum([]byte("checksum_4"))
	as := make([]apiobject.Attribute, 3)
	as[0].SetKey("attr_key1")
	as[0].SetValue("attr_val1")
	as[1].SetKey("__NEOFS__EXPIRATION_EPOCH")
	as[1].SetValue("8516691293958955670")
	as[2].SetKey("attr_key2")
	as[2].SetValue("attr_val2")
	var ps refs.Signature
	ps.SetKey([]byte("pub_1"))
	ps.SetSign([]byte("sig_1"))
	ps.SetScheme(1277002296)
	var sh apiobject.SplitHeader
	sh.SetParent(protoIDFromBytes(anyValidIDs[1][:]))
	sh.SetPrevious(protoIDFromBytes(anyValidIDs[2][:]))
	sh.SetParentSignature(&ps)
	sh.SetParentHeader(&ph)
	sh.SetChildren([]refs.ObjectID{
		*protoIDFromBytes(anyValidIDs[3][:]),
		*protoIDFromBytes(anyValidIDs[4][:]),
		*protoIDFromBytes(anyValidIDs[5][:]),
	})
	sh.SetSplitID(anyValidSplitIDBytes)
	sh.SetFirst(protoIDFromBytes(anyValidIDs[6][:]))

	var mh apiobject.Header
	mh.SetVersion(&mv)
	mh.SetContainerID(protoContainerIDFromBytes(anyValidContainers[1][:]))
	mh.SetOwnerID(protoUserIDFromBytes(anyValidUsers[1][:]))
	mh.SetCreationEpoch(anyValidCreationEpoch + 1)
	mh.SetPayloadLength(anyValidPayloadSize + 1)
	mh.SetPayloadHash(&cs)
	mh.SetObjectType(apiobject.Type(anyValidType) + 1)
	mh.SetHomomorphicHash(&hcs)
	mh.SetSessionToken(&mt)
	mh.SetAttributes(as)
	mh.SetSplit(&sh)

	var ms refs.Signature
	ms.SetKey([]byte("pub_2"))
	ms.SetSign([]byte("sig_2"))
	ms.SetScheme(1242896683)

	var m apiobject.Object
	m.SetObjectID(protoIDFromBytes(anyValidIDs[0][:]))
	m.SetSignature(&ms)
	m.SetHeader(&mh)
	m.SetPayload(anyValidRegularPayload)

	var obj object.Object
	require.NoError(t, obj.ReadFromV2(m))
	require.Equal(t, validObject, obj)

	// reset optional fields
	m.SetObjectID(nil)
	m.SetSignature(nil)
	m.SetHeader(nil)
	m.SetPayload(nil)
	obj2 := obj
	require.NoError(t, obj2.ReadFromV2(m))
	require.Zero(t, obj2)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*apiobject.Object)
		}{
			{name: "id/nil value", err: "invalid ID: invalid length 0",
				corrupt: func(m *apiobject.Object) { m.SetObjectID(protoIDFromBytes(nil)) }},
			{name: "id/empty value", err: "invalid ID: invalid length 0",
				corrupt: func(m *apiobject.Object) { m.SetObjectID(protoIDFromBytes([]byte{})) }},
			{name: "id/undersize", err: "invalid ID: invalid length 31",
				corrupt: func(m *apiobject.Object) { m.SetObjectID(protoIDFromBytes(anyValidIDs[0][:31])) }},
			{name: "id/oversize", err: "invalid ID: invalid length 33",
				corrupt: func(m *apiobject.Object) { m.SetObjectID(protoIDFromBytes(append(anyValidIDs[0][:], 1))) }},
			{name: "id/zero", err: "invalid ID: zero object ID",
				corrupt: func(m *apiobject.Object) { m.SetObjectID(protoIDFromBytes(make([]byte, 32))) }},
			{name: "signature/scheme/overflow", err: "invalid signature: scheme 2147483648 overflows int32",
				corrupt: func(m *apiobject.Object) {
					var s refs.Signature
					s.SetScheme(math.MaxInt32 + 1)
					m.SetSignature(&s)
				}},
			{name: "header/owner/value/nil", err: "invalid header: invalid owner: invalid length 0, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetOwnerID(protoUserIDFromBytes(nil))
					m.SetHeader(&h)
				}},
			{name: "header/owner/value/empty", err: "invalid header: invalid owner: invalid length 0, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetOwnerID(protoUserIDFromBytes([]byte{}))
					m.SetHeader(&h)
				}},
			{name: "header/owner/value/undersize", err: "invalid header: invalid owner: invalid length 24, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetOwnerID(protoUserIDFromBytes(anyValidUsers[0][:24]))
					m.SetHeader(&h)
				}},
			{name: "header/owner/value/oversize", err: "invalid header: invalid owner: invalid length 26, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetOwnerID(protoUserIDFromBytes(append(anyValidUsers[0][:], 1)))
					m.SetHeader(&h)
				}},
			{name: "header/owner/value/wrong prefix", err: "invalid header: invalid owner: invalid prefix byte 0x42, expected 0x35",
				corrupt: func(m *apiobject.Object) {
					id := anyValidUsers[0]
					id[0] = 0x42
					h := *m.GetHeader()
					h.SetOwnerID(protoUserIDFromBytes(id[:]))
					m.SetHeader(&h)
				}},
			{name: "header/owner/value/checksum mismatch", err: "invalid header: invalid owner: checksum mismatch",
				corrupt: func(m *apiobject.Object) {
					id := anyValidUsers[0]
					id[len(id)-1]++
					h := *m.GetHeader()
					h.SetOwnerID(protoUserIDFromBytes(id[:]))
					m.SetHeader(&h)
				}},
			{name: "header/container/nil value", err: "invalid header: invalid container: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetContainerID(protoContainerIDFromBytes(nil))
					m.SetHeader(&h)
				}},
			{name: "header/container/empty value", err: "invalid header: invalid container: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetContainerID(protoContainerIDFromBytes([]byte{}))
					m.SetHeader(&h)
				}},
			{name: "header/container/undersize", err: "invalid header: invalid container: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetContainerID(protoContainerIDFromBytes(anyValidContainers[0][:31]))
					m.SetHeader(&h)
				}},
			{name: "header/container/oversize", err: "invalid header: invalid container: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetContainerID(protoContainerIDFromBytes(append(anyValidContainers[0][:], 1)))
					m.SetHeader(&h)
				}},
			{name: "header/container/zero", err: "invalid header: invalid container: zero container ID",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetContainerID(protoContainerIDFromBytes(make([]byte, 32)))
					m.SetHeader(&h)
				}},
			{name: "header/payload checksum/missing value", err: "invalid header: invalid payload checksum: missing value",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetPayloadHash(new(refs.Checksum))
					m.SetHeader(&h)
				}},
			{name: "header/payload homomorphic checksum/missing value", err: "invalid header: invalid payload homomorphic checksum: missing value",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					h.SetHomomorphicHash(new(refs.Checksum))
					m.SetHeader(&h)
				}},
			{name: "header/session/body/ID/undersize", err: "invalid header: invalid session token: invalid session ID: invalid UUID (got 15 bytes)",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtb.SetID(anyValidSessionID[:15])
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/ID/oversize", err: "invalid header: invalid session token: invalid session ID: invalid UUID (got 17 bytes)",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtb.SetID(append(anyValidSessionID[:], 1))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/ID/wrong UUID version", err: "invalid header: invalid session token: invalid session ID: wrong UUID version 3, expected 4",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					b := bytes.Clone(anyValidSessionID[:])
					b[6] = 3 << 4
					mtb.SetID(b)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/issuer/value/empty", err: "invalid header: invalid session token: invalid session issuer: invalid length 0, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtb.SetOwnerID(protoUserIDFromBytes(nil))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/issuer/value/undersize", err: "invalid header: invalid session token: invalid session issuer: invalid length 24, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtb.SetOwnerID(protoUserIDFromBytes(anyValidUsers[0][:24]))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/issuer/value/oversize", err: "invalid header: invalid session token: invalid session issuer: invalid length 26, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtb.SetOwnerID(protoUserIDFromBytes(append(anyValidUsers[0][:], 1)))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/issuer/value/wrong prefix", err: "invalid header: invalid session token: invalid session issuer: invalid prefix byte 0x42, expected 0x35",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					b := bytes.Clone(anyValidUsers[0][:])
					b[0] = 0x42
					mtb.SetOwnerID(protoUserIDFromBytes(b))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/issuer/value/checksum mismatch", err: "invalid header: invalid session token: invalid session issuer: checksum mismatch",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					b := bytes.Clone(anyValidUsers[0][:])
					b[len(b)-1]++
					mtb.SetOwnerID(protoUserIDFromBytes(b))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/wrong oneof", err: "invalid header: invalid session token: invalid context: invalid context *session.ContainerSessionContext",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtb.SetContext(new(apisession.ContainerSessionContext))
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/container/empty value", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtc := *mtb.GetContext().(*apisession.ObjectSessionContext)
					mtc.SetTarget(protoContainerIDFromBytes(nil), mtc.GetObjects()...)
					mtb.SetContext(&mtc)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/container/undersize", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtc := *mtb.GetContext().(*apisession.ObjectSessionContext)
					mtc.SetTarget(protoContainerIDFromBytes(anyValidContainers[0][:31]), mtc.GetObjects()...)
					mtb.SetContext(&mtc)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/container/oversize", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtc := *mtb.GetContext().(*apisession.ObjectSessionContext)
					mtc.SetTarget(protoContainerIDFromBytes(append(anyValidContainers[0][:], 1)), mtc.GetObjects()...)
					mtb.SetContext(&mtc)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/object/empty value", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtc := *mtb.GetContext().(*apisession.ObjectSessionContext)
					objs := make([]refs.ObjectID, 2)
					anyValidIDs[0].WriteToV2(&objs[0])
					mtc.SetTarget(protoContainerIDFromBytes(anyValidContainers[0][:]), objs...)
					mtb.SetContext(&mtc)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/object/undersize", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtc := *mtb.GetContext().(*apisession.ObjectSessionContext)
					objs := make([]refs.ObjectID, 2)
					anyValidIDs[0].WriteToV2(&objs[0])
					objs[1].SetValue(anyValidIDs[1][:31])
					mtc.SetTarget(protoContainerIDFromBytes(anyValidContainers[0][:]), objs...)
					mtb.SetContext(&mtc)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/body/context/object/oversize", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mtb := *mt.GetBody()
					mtc := *mtb.GetContext().(*apisession.ObjectSessionContext)
					objs := make([]refs.ObjectID, 2)
					anyValidIDs[0].WriteToV2(&objs[0])
					objs[1].SetValue(append(anyValidIDs[1][:], 1))
					mtc.SetTarget(protoContainerIDFromBytes(anyValidContainers[0][:]), objs...)
					mtb.SetContext(&mtc)
					mt.SetBody(&mtb)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "header/session/signature/scheme/overflow", err: "invalid header: invalid session token: invalid body signature: scheme 2147483648 overflows int32",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					mt := *h.GetSessionToken()
					mts := *mt.GetSignature()
					mts.SetScheme(math.MaxInt32 + 1)
					mt.SetSignature(&mts)
					h.SetSessionToken(&mt)
					m.SetHeader(&h)
				}},
			{name: "attributes/no key", err: "invalid header: empty key of the attribute #1",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					as := make([]apiobject.Attribute, 3)
					as[0].SetKey("k1")
					as[0].SetValue("v1")
					as[1].SetValue("v2")
					as[2].SetKey("k3")
					as[2].SetValue("v3")
					h.SetAttributes(as)
					m.SetHeader(&h)
				}},
			{name: "attributes/no value", err: "invalid header: empty value of the attribute #1 (k2)",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					as := make([]apiobject.Attribute, 3)
					as[0].SetKey("k1")
					as[0].SetValue("v1")
					as[1].SetKey("k2")
					as[2].SetKey("k3")
					as[2].SetValue("v3")
					h.SetAttributes(as)
					m.SetHeader(&h)
				}},
			{name: "attributes/duplicated", err: "invalid header: duplicated attribute k1",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					as := make([]apiobject.Attribute, 3)
					as[0].SetKey("k1")
					as[0].SetValue("v1")
					as[1].SetKey("k2")
					as[1].SetValue("v2")
					as[2].SetKey("k1")
					as[2].SetValue("v3")
					h.SetAttributes(as)
					m.SetHeader(&h)
				}},
			{name: "attributes/expiration", err: "invalid header: invalid expiration attribute (must be a uint): strconv.ParseUint: parsing \"foo\": invalid syntax",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					as := make([]apiobject.Attribute, 3)
					as[0].SetKey("k1")
					as[0].SetValue("v1")
					as[1].SetKey("__NEOFS__EXPIRATION_EPOCH")
					as[1].SetValue("foo")
					as[2].SetKey("k3")
					as[2].SetValue("v3")
					h.SetAttributes(as)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent ID/empty value", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetParent(protoIDFromBytes(nil))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent ID/undersize", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetParent(protoIDFromBytes(anyValidIDs[0][:31]))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent ID/oversize", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetParent(protoIDFromBytes(append(anyValidIDs[0][:], 1)))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent ID/zero", err: "invalid header: invalid split header: invalid parent split member ID: zero object ID",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetParent(protoIDFromBytes(make([]byte, 32)))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/previous/empty value", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetPrevious(protoIDFromBytes(nil))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/previous/undersize", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetPrevious(protoIDFromBytes(anyValidIDs[0][:31]))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/previous/oversize", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetPrevious(protoIDFromBytes(append(anyValidIDs[0][:], 1)))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/previous/zero", err: "invalid header: invalid split header: invalid previous split member ID: zero object ID",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetPrevious(protoIDFromBytes(make([]byte, 32)))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/first/empty value", err: "invalid header: invalid split header: invalid first split member ID: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetFirst(protoIDFromBytes(nil))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/first/undersize", err: "invalid header: invalid split header: invalid first split member ID: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetFirst(protoIDFromBytes(anyValidIDs[0][:31]))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/first/oversize", err: "invalid header: invalid split header: invalid first split member ID: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetFirst(protoIDFromBytes(append(anyValidIDs[0][:], 1)))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/first/zero", err: "invalid header: invalid split header: invalid first split member ID: zero object ID",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetFirst(protoIDFromBytes(make([]byte, 32)))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/ID/undersize", err: "invalid header: invalid split header: invalid split ID: invalid UUID (got 15 bytes)",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetSplitID(anyValidSplitIDBytes[:15])
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/ID/oversize", err: "invalid header: invalid split header: invalid split ID: invalid UUID (got 17 bytes)",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					sh.SetSplitID(append(anyValidSplitIDBytes, 1))
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/ID/wrong UUID version", err: "invalid header: invalid split header: invalid split ID: wrong UUID version 3, expected 4",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					b := bytes.Clone(anyValidSplitIDBytes)
					b[6] = 3 << 4
					sh.SetSplitID(b)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/children/empty value", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					c := make([]refs.ObjectID, 3)
					anyValidIDs[0].WriteToV2(&c[0])
					anyValidIDs[2].WriteToV2(&c[2])
					sh.SetChildren(c)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/children/undersize", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					c := make([]refs.ObjectID, 3)
					anyValidIDs[0].WriteToV2(&c[0])
					c[1].SetValue(anyValidIDs[1][:31])
					anyValidIDs[2].WriteToV2(&c[2])
					sh.SetChildren(c)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/children/oversize", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					c := make([]refs.ObjectID, 3)
					anyValidIDs[0].WriteToV2(&c[0])
					c[1].SetValue(append(anyValidIDs[1][:], 1))
					anyValidIDs[2].WriteToV2(&c[2])
					sh.SetChildren(c)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/children/zero", err: "invalid header: invalid split header: invalid child split member ID #1: zero object ID",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					c := make([]refs.ObjectID, 3)
					anyValidIDs[0].WriteToV2(&c[0])
					c[1].SetValue(make([]byte, 32))
					anyValidIDs[2].WriteToV2(&c[2])
					sh.SetChildren(c)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent signature/scheme/overflow", err: "invalid header: invalid split header: invalid parent signature: scheme 2147483648 overflows int32",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					var s refs.Signature
					anyValidSignatures[0].WriteToV2(&s)
					s.SetScheme(math.MaxInt32 + 1)
					sh.SetParentSignature(&s)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/owner/value/empty", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 0, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetOwnerID(protoUserIDFromBytes(nil))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/owner/value/undersize", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 24, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetOwnerID(protoUserIDFromBytes(anyValidUsers[0][:24]))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/owner/value/oversize", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 26, expected 25",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetOwnerID(protoUserIDFromBytes(append(anyValidUsers[0][:], 1)))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/owner/value/wrong prefix", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid prefix byte 0x42, expected 0x35",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					b := bytes.Clone(anyValidUsers[0][:])
					b[0] = 0x42
					ph.SetOwnerID(protoUserIDFromBytes(b))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/owner/value/checksum mismatch", err: "invalid header: invalid split header: invalid parent header: invalid owner: checksum mismatch",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					b := bytes.Clone(anyValidUsers[0][:])
					b[len(b)-1]++
					ph.SetOwnerID(protoUserIDFromBytes(b))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/container/empty value", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 0",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetContainerID(protoContainerIDFromBytes(nil))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/container/undersize", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 31",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetContainerID(protoContainerIDFromBytes(anyValidContainers[0][:31]))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/container/oversize", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 33",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetContainerID(protoContainerIDFromBytes(append(anyValidContainers[0][:], 1)))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/container/zero", err: "invalid header: invalid split header: invalid parent header: invalid container: zero container ID",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetContainerID(protoContainerIDFromBytes(make([]byte, 32)))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/payload checksum/missing value", err: "invalid header: invalid split header: invalid parent header: invalid payload checksum: missing value",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetPayloadHash(new(refs.Checksum))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
			{name: "header/split/parent/payload homomorphic checksum/missing value", err: "invalid header: invalid split header: invalid parent header: invalid payload homomorphic checksum: missing value",
				corrupt: func(m *apiobject.Object) {
					h := *m.GetHeader()
					sh := *h.GetSplit()
					ph := *sh.GetParentHeader()
					ph.SetHomomorphicHash(new(refs.Checksum))
					sh.SetParentHeader(&ph)
					h.SetSplit(&sh)
					m.SetHeader(&h)
				}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				m := obj.ToV2()
				tc.corrupt(m)
				require.EqualError(t, new(object.Object).ReadFromV2(*m), tc.err)
			})
		}
	})
}

func TestObject_ToV2(t *testing.T) {
	// zero
	m := object.Object{}.ToV2()
	require.Zero(t, m.GetObjectID())
	require.Zero(t, m.GetSignature())
	require.Zero(t, m.GetHeader())
	require.Zero(t, m.GetPayload())

	// filled
	m = validObject.ToV2()
	require.Equal(t, anyValidIDs[0][:], m.GetObjectID().GetValue())
	require.Equal(t, anyValidRegularPayload, m.GetPayload())
	msig := m.GetSignature()
	require.Equal(t, anyValidSignatures[1].PublicKeyBytes(), msig.GetKey())
	require.Equal(t, anyValidSignatures[1].Value(), msig.GetSign())
	require.EqualValues(t, 1242896683, msig.GetScheme())

	mh := m.GetHeader()
	require.EqualValues(t, 525747025, mh.GetVersion().GetMajor())
	require.EqualValues(t, 171993162, mh.GetVersion().GetMinor())
	require.Equal(t, anyValidContainers[1][:], mh.GetContainerID().GetValue())
	require.Equal(t, anyValidUsers[1][:], mh.GetOwnerID().GetValue())
	require.EqualValues(t, anyValidCreationEpoch+1, mh.GetCreationEpoch())
	require.EqualValues(t, anyValidPayloadSize+1, mh.GetPayloadLength())
	require.EqualValues(t, 126384577, mh.GetPayloadHash().GetType())
	require.EqualValues(t, "checksum_3", mh.GetPayloadHash().GetSum())
	require.EqualValues(t, anyValidType+1, mh.GetObjectType())
	require.EqualValues(t, 1001923429, mh.GetHomomorphicHash().GetType())
	require.EqualValues(t, "checksum_4", mh.GetHomomorphicHash().GetSum())

	mt := mh.GetSessionToken()
	mb := mt.GetBody()
	require.Equal(t, anyValidSessionID[:], mb.GetID())
	require.Equal(t, anyValidUsers[2][:], mb.GetOwnerID().GetValue())
	require.Equal(t, anySessionIssuerPubKeyBytes, mb.GetSessionKey())
	require.EqualValues(t, uint64(16429376563136800338), mb.GetLifetime().GetExp())
	require.EqualValues(t, 7956510363313998522, mb.GetLifetime().GetIat())
	require.EqualValues(t, uint64(17237208928641773338), mb.GetLifetime().GetNbf())
	c := mb.GetContext()
	require.IsType(t, new(apisession.ObjectSessionContext), c)
	co := c.(*apisession.ObjectSessionContext)
	require.EqualValues(t, 1047242055, co.GetVerb())
	require.Equal(t, anyValidContainers[2][:], co.GetContainer().GetValue())
	objs := co.GetObjects()
	require.Len(t, objs, 2)
	require.Equal(t, anyValidIDs[8][:], objs[0].GetValue())
	require.Equal(t, anyValidIDs[9][:], objs[1].GetValue())
	ms := mt.GetSignature()
	require.EqualValues(t, 1134494890, ms.GetScheme())
	require.EqualValues(t, "session_signer", ms.GetKey())
	require.EqualValues(t, "session_signature", ms.GetSign())

	as := mh.GetAttributes()
	require.Len(t, as, 3)
	require.Equal(t, "attr_key1", as[0].GetKey())
	require.Equal(t, "attr_val1", as[0].GetValue())
	require.Equal(t, "__NEOFS__EXPIRATION_EPOCH", as[1].GetKey())
	require.Equal(t, "8516691293958955670", as[1].GetValue())
	require.Equal(t, "attr_key2", as[2].GetKey())
	require.Equal(t, "attr_val2", as[2].GetValue())

	sh := mh.GetSplit()
	require.Equal(t, anyValidIDs[1][:], sh.GetParent().GetValue())
	require.Equal(t, anyValidIDs[2][:], sh.GetPrevious().GetValue())
	require.Equal(t, anyValidSplitIDBytes, sh.GetSplitID())
	require.Equal(t, anyValidIDs[6][:], sh.GetFirst().GetValue())
	ch := sh.GetChildren()
	require.Len(t, ch, 3)
	require.Equal(t, anyValidIDs[3][:], ch[0].GetValue())
	require.Equal(t, anyValidIDs[4][:], ch[1].GetValue())
	require.Equal(t, anyValidIDs[5][:], ch[2].GetValue())
	ms = sh.GetParentSignature()
	require.EqualValues(t, 1277002296, ms.GetScheme())
	require.EqualValues(t, "pub_1", ms.GetKey())
	require.EqualValues(t, "sig_1", ms.GetSign())

	ph := sh.GetParentHeader()
	require.NotNil(t, ph)
	require.Zero(t, ph.GetSessionToken())
	require.Zero(t, ph.GetSplit())
	require.EqualValues(t, 88789927, ph.GetVersion().GetMajor())
	require.EqualValues(t, 2018985309, ph.GetVersion().GetMinor())
	require.Equal(t, anyValidContainers[0][:], ph.GetContainerID().GetValue())
	require.Equal(t, anyValidUsers[0][:], ph.GetOwnerID().GetValue())
	require.EqualValues(t, anyValidCreationEpoch, ph.GetCreationEpoch())
	require.EqualValues(t, anyValidPayloadSize, ph.GetPayloadLength())
	require.EqualValues(t, 1974315742, ph.GetPayloadHash().GetType())
	require.EqualValues(t, "checksum_1", ph.GetPayloadHash().GetSum())
	require.EqualValues(t, anyValidType, ph.GetObjectType())
	require.EqualValues(t, 1922538608, ph.GetHomomorphicHash().GetType())
	require.EqualValues(t, "checksum_2", ph.GetHomomorphicHash().GetSum())

	as = ph.GetAttributes()
	require.Len(t, as, 3)
	require.Equal(t, "par_attr_key1", as[0].GetKey())
	require.Equal(t, "par_attr_val1", as[0].GetValue())
	require.Equal(t, "__NEOFS__EXPIRATION_EPOCH", as[1].GetKey())
	require.Equal(t, "14208497712700580130", as[1].GetValue())
	require.Equal(t, "par_attr_key2", as[2].GetKey())
	require.Equal(t, "par_attr_val2", as[2].GetValue())
}

func TestObject_Marshal(t *testing.T) {
	require.Empty(t, object.Object{}.Marshal())
	require.Equal(t, validBinObject, validObject.Marshal())
}

func TestObject_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(object.Object).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "id/empty value", err: "invalid ID: invalid length 0",
				b: []byte{10, 0}},
			{name: "id/undersize", err: "invalid ID: invalid length 31",
				b: []byte{10, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "id/oversize", err: "invalid ID: invalid length 33",
				b: []byte{10, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "id/zero", err: "invalid ID: zero object ID",
				b: []byte{10, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "signature/scheme/overflow", err: "invalid signature: scheme 2147483648 overflows int32",
				b: []byte{18, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
			{name: "header/owner/value/empty", err: "invalid header: invalid owner: invalid length 0, expected 25",
				b: []byte{26, 2, 26, 0}},
			{name: "header/owner/value/undersize", err: "invalid header: invalid owner: invalid length 24, expected 25",
				b: []byte{26, 28, 26, 26, 10, 24, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137}},
			{name: "header/owner/value/oversize", err: "invalid header: invalid owner: invalid length 26, expected 25",
				b: []byte{26, 30, 26, 28, 10, 26, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 153, 1}},
			{name: "header/owner/value/wrong prefix", err: "invalid header: invalid owner: invalid prefix byte 0x42, expected 0x35",
				b: []byte{26, 29, 26, 27, 10, 25, 66, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 153}},
			{name: "header/owner/value/checksum mismatch", err: "invalid header: invalid owner: checksum mismatch",
				b: []byte{26, 29, 26, 27, 10, 25, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 154}},
			{name: "header/container/empty value", err: "invalid header: invalid container: invalid length 0",
				b: []byte{26, 2, 18, 0}},
			{name: "header/container/undersize", err: "invalid header: invalid container: invalid length 31",
				b: []byte{26, 35, 18, 33, 10, 31, 245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179,
					93, 183, 250, 180, 255, 162, 182, 222, 220, 99, 125, 136, 117, 206}},
			{name: "header/container/oversize", err: "invalid header: invalid container: invalid length 33",
				b: []byte{26, 37, 18, 35, 10, 33, 245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179,
					93, 183, 250, 180, 255, 162, 182, 222, 220, 99, 125, 136, 117, 206, 34, 1}},
			{name: "header/container/zero", err: "invalid header: invalid container: zero container ID",
				b: []byte{26, 36, 18, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/payload checksum/missing value", err: "invalid header: invalid payload checksum: missing value",
				b: []byte{26, 2, 50, 0}},
			{name: "header/payload homomorphic checksum/missing value", err: "invalid header: invalid payload homomorphic checksum: missing value",
				b: []byte{26, 2, 66, 0}},
			{name: "header/session/body/ID/undersize", err: "invalid header: invalid session token: invalid session ID: invalid UUID (got 15 bytes)",
				b: []byte{26, 154, 2, 74, 151, 2, 10, 233, 1, 10, 15, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151, 159,
					221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154,
					222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149,
					219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212,
					15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158,
					67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233,
					102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114,
					18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/ID/oversize", err: "invalid header: invalid session token: invalid session ID: invalid UUID (got 17 bytes)",
				b: []byte{26, 156, 2, 74, 153, 2, 10, 235, 1, 10, 17, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 1, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150,
					151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1,
					16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2,
					154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240,
					183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89,
					149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90,
					212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179,
					158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11,
					54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101,
					114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/ID/wrong UUID version", err: "invalid header: invalid session token: invalid session ID: wrong UUID version 3, expected 4",
				b: []byte{26, 155, 2, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 48, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2,
					154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240,
					183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89,
					149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90,
					212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179,
					158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18,
					17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/issuer/value/empty", err: "invalid header: invalid session token: invalid session issuer: invalid length 0, expected 25",
				b: []byte{26, 128, 2, 74, 253, 1, 10, 207, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 0, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154, 222, 137, 183, 154,
					198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154,
					163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183, 253, 76, 187, 136, 215,
					164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149, 219, 185, 209, 233, 137,
					224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51, 86, 142, 101, 155,
					141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32, 154, 122, 174,
					117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68, 233, 22,
					158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18,
					41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110,
					32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/issuer/value/undersize", err: "invalid header: invalid session token: invalid session issuer: invalid length 24, expected 25",
				b: []byte{26, 154, 2, 74, 151, 2, 10, 233, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 26, 10, 24, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154, 222, 137,
					183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154, 144, 131, 197,
					214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183, 253, 76, 187,
					136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149, 219, 185, 209,
					233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51, 86, 142,
					101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32, 154,
					122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68,
					233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136,
					207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115,
					105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/issuer/value/oversize", err: "invalid header: invalid session token: invalid session issuer: invalid length 26, expected 25",
				b: []byte{26, 156, 2, 74, 153, 2, 10, 235, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 28, 10, 26, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 153, 1, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154,
					222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149,
					219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212,
					15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158,
					67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233,
					102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114,
					18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/issuer/value/wrong prefix", err: "invalid header: invalid session token: invalid session issuer: invalid prefix byte 0x42, expected 0x35",
				b: []byte{26, 155, 2, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 66, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 153, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154, 222,
					137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154, 144, 131,
					197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183, 253, 76,
					187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149, 219, 185,
					209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51, 86,
					142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32,
					154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232,
					136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98,
					99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101,
					115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/issuer/value/checksum mismatch", err: "invalid header: invalid session token: invalid session issuer: checksum mismatch",
				b: []byte{26, 155, 2, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129,
					211, 214, 90, 145, 237, 137, 154, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154, 222,
					137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154, 144, 131,
					197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183, 253, 76,
					187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149, 219, 185,
					209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51, 86,
					142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32,
					154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136,
					68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136,
					207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115,
					105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/wrong oneof", err: "invalid header: invalid session token: invalid context: invalid context *session.ContainerSessionContext",
				b: []byte{26, 166, 1, 74, 163, 1, 10, 118, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142, 52,
					17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151, 159,
					221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16, 154,
					222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154, 144,
					131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183, 253,
					76, 187, 136, 215, 164, 217, 50, 0, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114,
					18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/container/empty value", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 0",
				b: []byte{26, 249, 1, 74, 246, 1, 10, 200, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 82, 8, 199, 202, 174, 243, 3, 18, 74, 10, 0, 18, 34, 10, 32, 57, 171, 109,
					41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238,
					61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219,
					143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105,
					111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117,
					114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/container/undersize", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 31",
				b: []byte{26, 154, 2, 74, 151, 2, 10, 233, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 115, 8, 199, 202, 174, 243, 3, 18, 107, 10, 33, 10, 31, 245, 94, 164, 207,
					217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179, 93, 183, 250, 180, 255, 162, 182, 222, 220, 99,
					125, 136, 117, 206, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32,
					154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136,
					68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136,
					207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115,
					105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/container/oversize", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 33",
				b: []byte{26, 156, 2, 74, 153, 2, 10, 235, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 117, 8, 199, 202, 174, 243, 3, 18, 109, 10, 35, 10, 33, 245, 94, 164, 207,
					217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179, 93, 183, 250, 180, 255, 162, 182, 222, 220, 99,
					125, 136, 117, 206, 34, 1, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40,
					32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232,
					136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99,
					136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115,
					115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/object/empty value", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 0",
				b: []byte{26, 249, 1, 74, 246, 1, 10, 200, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 82, 8, 199, 202, 174, 243, 3, 18, 74, 10, 34, 10, 32, 135, 89, 149, 219,
					185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51,
					86, 142, 101, 155, 141, 18, 0, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181,
					95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115,
					115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110,
					97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/object/undersize", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 31",
				b: []byte{26, 154, 2, 74, 151, 2, 10, 233, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 115, 8, 199, 202, 174, 243, 3, 18, 107, 10, 34, 10, 32, 135, 89, 149, 219,
					185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51,
					86, 142, 101, 155, 141, 18, 33, 10, 31, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45,
					225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68,
					233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136,
					207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115,
					105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/body/context/object/oversize", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 33",
				b: []byte{26, 156, 2, 74, 153, 2, 10, 235, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 117, 8, 199, 202, 174, 243, 3, 18, 109, 10, 34, 10, 32, 135, 89, 149, 219,
					185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51,
					86, 142, 101, 155, 141, 18, 35, 10, 33, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45,
					225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 1, 18, 34, 10, 32, 110, 233, 102, 232,
					136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99,
					136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115,
					115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/session/signature/scheme/overflow", err: "invalid header: invalid session token: invalid body signature: scheme 2147483648 overflows int32",
				b: []byte{26, 253, 1, 74, 250, 1, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229, 102, 253, 142,
					52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151,
					159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1, 16,
					154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2, 154,
					144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149, 219,
					185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51,
					86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40,
					32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232,
					136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99,
					136, 207, 21, 18, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
			{name: "attributes/no key", err: "invalid header: empty key of the attribute #1",
				b: []byte{26, 26, 82, 8, 10, 2, 107, 49, 18, 2, 118, 49, 82, 4, 18, 2, 118, 49, 82, 8, 10, 2, 107, 51, 18, 2, 118, 51}},
			{name: "attributes/no value", err: "invalid header: empty value of the attribute #1 (k2)",
				b: []byte{26, 26, 82, 8, 10, 2, 107, 49, 18, 2, 118, 49, 82, 4, 10, 2, 107, 50, 82, 8, 10, 2, 107, 51, 18, 2, 118, 51}},
			{name: "attributes/duplicated", err: "invalid header: duplicated attribute k1",
				b: []byte{26, 30, 82, 8, 10, 2, 107, 49, 18, 2, 118, 49, 82, 8, 10, 2, 107, 50, 18, 2, 118, 50, 82, 8, 10, 2, 107,
					49, 18, 2, 118, 51}},
			{name: "attributes/expiration", err: "invalid header: invalid expiration attribute (must be a uint): strconv.ParseUint: parsing \"foo\": invalid syntax",
				b: []byte{26, 54, 82, 8, 10, 2, 107, 49, 18, 2, 118, 49, 82, 32, 10, 25, 95, 95, 78, 69, 79, 70, 83, 95, 95, 69,
					88, 80, 73, 82, 65, 84, 73, 79, 78, 95, 69, 80, 79, 67, 72, 18, 3, 102, 111, 111, 82, 8, 10, 2, 107, 51, 18, 2, 118, 51}},
			{name: "header/split/parent ID/empty value", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 0",
				b: []byte{26, 4, 90, 2, 10, 0}},
			{name: "header/split/parent ID/undersize", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 31",
				b: []byte{26, 37, 90, 35, 10, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "header/split/parent ID/oversize", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 33",
				b: []byte{26, 39, 90, 37, 10, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "header/split/parent ID/zero", err: "invalid header: invalid split header: invalid parent split member ID: zero object ID",
				b: []byte{26, 38, 90, 36, 10, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/previous/empty value", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 0",
				b: []byte{26, 4, 90, 2, 18, 0}},
			{name: "header/split/previous/undersize", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 31",
				b: []byte{26, 37, 90, 35, 18, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "header/split/previous/oversize", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 33",
				b: []byte{26, 39, 90, 37, 18, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "header/split/previous/zero", err: "invalid header: invalid split header: invalid previous split member ID: zero object ID",
				b: []byte{26, 38, 90, 36, 18, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/first/empty value", err: "invalid header: invalid split header: invalid first split member ID: invalid length 0",
				b: []byte{26, 4, 90, 2, 58, 0}},
			{name: "header/split/first/undersize", err: "invalid header: invalid split header: invalid first split member ID: invalid length 31",
				b: []byte{26, 37, 90, 35, 58, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "header/split/first/oversize", err: "invalid header: invalid split header: invalid first split member ID: invalid length 33",
				b: []byte{26, 39, 90, 37, 58, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "header/split/first/zero", err: "invalid header: invalid split header: invalid first split member ID: zero object ID",
				b: []byte{26, 38, 90, 36, 58, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/ID/undersize", err: "invalid header: invalid split header: invalid split ID: invalid UUID (got 15 bytes)",
				b: []byte{26, 19, 90, 17, 50, 15, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147}},
			{name: "header/split/ID/oversize", err: "invalid header: invalid split header: invalid split ID: invalid UUID (got 17 bytes)",
				b: []byte{26, 21, 90, 19, 50, 17, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41, 1}},
			{name: "header/split/ID/wrong UUID version", err: "invalid header: invalid split header: invalid split ID: wrong UUID version 3, expected 4",
				b: []byte{26, 20, 90, 18, 50, 16, 224, 132, 3, 80, 32, 44, 48, 184, 185, 32, 226, 201, 206, 196, 147, 41}},
			{name: "header/split/children/empty value", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 0",
				b: []byte{26, 40, 90, 38, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 0}},
			{name: "header/split/children/undersize", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 31",
				b: []byte{26, 73, 90, 71, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 33, 10, 31, 229, 77, 63, 235, 2,
					9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253,
					127, 179}},
			{name: "header/split/children/oversize", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 33",
				b: []byte{26, 75, 90, 73, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 35, 10, 33, 229, 77, 63, 235, 2,
					9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243, 57, 253,
					127, 179, 235, 1}},
			{name: "header/split/children/zero", err: "invalid header: invalid split header: invalid child split member ID #1: zero object ID",
				b: []byte{26, 74, 90, 72, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193,
					190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/parent signature/scheme/overflow", err: "invalid header: invalid split header: invalid parent signature: scheme 2147483648 overflows int32",
				b: []byte{26, 15, 90, 13, 26, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
			{name: "header/split/parent/owner/value/empty", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 0, expected 25",
				b: []byte{26, 6, 90, 4, 34, 2, 26, 0}},
			{name: "header/split/parent/owner/value/undersize", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 24, expected 25",
				b: []byte{26, 32, 90, 30, 34, 28, 26, 26, 10, 24, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237,
					140, 215, 52, 129, 211, 214, 90, 145, 237, 137}},
			{name: "header/split/parent/owner/value/oversize", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 26, expected 25",
				b: []byte{26, 34, 90, 32, 34, 30, 26, 28, 10, 26, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237,
					140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 153, 1}},
			{name: "header/split/parent/owner/value/wrong prefix", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid prefix byte 0x42, expected 0x35",
				b: []byte{26, 33, 90, 31, 34, 29, 26, 27, 10, 25, 66, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237,
					140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 153}},
			{name: "header/split/parent/owner/value/checksum mismatch", err: "invalid header: invalid split header: invalid parent header: invalid owner: checksum mismatch",
				b: []byte{26, 33, 90, 31, 34, 29, 26, 27, 10, 25, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237,
					140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 154}},
			{name: "header/split/parent/container/empty value", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 0",
				b: []byte{26, 6, 90, 4, 34, 2, 18, 0}},
			{name: "header/split/parent/container/undersize", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 31",
				b: []byte{26, 39, 90, 37, 34, 35, 18, 33, 10, 31, 245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135,
					96, 204, 179, 93, 183, 250, 180, 255, 162, 182, 222, 220, 99, 125, 136, 117, 206}},
			{name: "header/split/parent/container/oversize", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 33",
				b: []byte{26, 41, 90, 39, 34, 37, 18, 35, 10, 33, 245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135,
					96, 204, 179, 93, 183, 250, 180, 255, 162, 182, 222, 220, 99, 125, 136, 117, 206, 34, 1}},
			{name: "header/split/parent/container/zero", err: "invalid header: invalid split header: invalid parent header: invalid container: zero container ID",
				b: []byte{26, 40, 90, 38, 34, 36, 18, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/parent/payload checksum/missing value", err: "invalid header: invalid split header: invalid parent header: invalid payload checksum: missing value",
				b: []byte{26, 6, 90, 4, 34, 2, 50, 0}},
			{name: "header/split/parent/payload homomorphic checksum/missing value", err: "invalid header: invalid split header: invalid parent header: invalid payload homomorphic checksum: missing value",
				b: []byte{26, 6, 90, 4, 34, 2, 66, 0}},
			{name: "header/split/parent/session/body/ID/undersize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session ID: invalid UUID (got 15 bytes)",
				b: []byte{26, 160, 2, 90, 157, 2, 34, 154, 2, 74, 151, 2, 10, 233, 1, 10, 15, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219,
					229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183,
					128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181,
					110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33,
					236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245,
					171, 29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89,
					61, 178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10,
					32, 110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25,
					48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110,
					101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/ID/oversize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session ID: invalid UUID (got 17 bytes)",
				b: []byte{26, 162, 2, 90, 159, 2, 34, 156, 2, 74, 153, 2, 10, 235, 1, 10, 17, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 1, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208,
					15, 219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128,
					222, 183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128,
					205, 181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26,
					192, 33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10,
					34, 10, 32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177,
					245, 171, 29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164,
					89, 61, 178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34,
					10, 32, 110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64,
					25, 48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103,
					110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137,
					252, 156, 4}},
			{name: "header/split/parent/session/body/ID/wrong UUID version", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session ID: wrong UUID version 3, expected 4",
				b: []byte{26, 161, 2, 90, 158, 2, 34, 155, 2, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 48, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
					29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18,
					17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/issuer/value/empty", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session issuer: invalid length 0, expected 25",
				b: []byte{26, 134, 2, 90, 131, 2, 34, 128, 2, 74, 253, 1, 10, 207, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 0, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1,
					16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2,
					154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240,
					183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149,
					219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15,
					51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40,
					32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232,
					136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99,
					136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115,
					115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/issuer/value/undersize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session issuer: invalid length 24, expected 25",
				b: []byte{26, 160, 2, 90, 157, 2, 34, 154, 2, 74, 151, 2, 10, 233, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157,
					229, 102, 253, 142, 52, 17, 144, 18, 26, 10, 24, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237,
					140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128, 228, 1,
					16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34, 33, 2,
					154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0, 240, 183,
					253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10, 32, 135, 89, 149, 219,
					185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212, 15, 51,
					86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32,
					154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136,
					68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136,
					207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115,
					105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/issuer/value/oversize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session issuer: invalid length 26, expected 25",
				b: []byte{26, 162, 2, 90, 159, 2, 34, 156, 2, 74, 153, 2, 10, 235, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 28, 10, 26, 53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229,
					237, 140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 153, 1, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183,
					128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181,
					110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33,
					236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245,
					171, 29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89,
					61, 178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10,
					32, 110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25,
					48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110,
					101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252,
					156, 4}},
			{name: "header/split/parent/session/body/issuer/value/wrong prefix", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session issuer: invalid prefix byte 0x42, expected 0x35",
				b: []byte{26, 161, 2, 90, 158, 2, 34, 155, 2, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 66, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
					29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18,
					17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/issuer/value/checksum mismatch", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid session issuer: checksum mismatch",
				b: []byte{26, 161, 2, 90, 158, 2, 34, 155, 2, 74, 152, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 223, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
					29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18,
					17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/wrong oneof", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid context *session.ContainerSessionContext",
				b: []byte{26, 172, 1, 90, 169, 1, 34, 166, 1, 74, 163, 1, 10, 118, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157, 229,
					102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229,
					62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183, 128,
					228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110, 34,
					33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236, 0,
					240, 183, 253, 76, 187, 136, 215, 164, 217, 50, 0, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105,
					103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170,
					137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/container/empty value", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid container ID: invalid length 0",
				b: []byte{26, 255, 1, 90, 252, 1, 34, 249, 1, 74, 246, 1, 10, 200, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 82, 8, 199, 202, 174, 243, 3, 18, 74, 10, 0, 18,
					34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221,
					138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110, 233, 102, 232, 136, 68, 233, 22, 158, 100,
					49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14,
					115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105,
					103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/container/undersize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid container ID: invalid length 31",
				b: []byte{26, 160, 2, 90, 157, 2, 34, 154, 2, 74, 151, 2, 10, 233, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157,
					229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219,
					229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183,
					128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181,
					110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33,
					236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 115, 8, 199, 202, 174, 243, 3, 18, 107, 10, 33, 10, 31,
					245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179, 93, 183, 250, 180, 255, 162,
					182, 222, 220, 99, 125, 136, 117, 206, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18,
					17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/container/oversize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid container ID: invalid length 33",
				b: []byte{26, 162, 2, 90, 159, 2, 34, 156, 2, 74, 153, 2, 10, 235, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 117, 8, 199, 202, 174, 243, 3, 18, 109, 10, 35, 10,
					33, 245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179, 93, 183, 250, 180, 255, 162,
					182, 222, 220, 99, 125, 136, 117, 206, 34, 1, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61,
					178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32,
					110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11,
					54, 207, 56, 98, 99, 136, 207, 21, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114,
					18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/object/empty value", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid target object: invalid length 0",
				b: []byte{26, 255, 1, 90, 252, 1, 34, 249, 1, 74, 246, 1, 10, 200, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 82, 8, 199, 202, 174, 243, 3, 18, 74, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
					29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 0, 18, 41, 10, 14,
					115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105,
					103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/object/undersize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid target object: invalid length 31",
				b: []byte{26, 160, 2, 90, 157, 2, 34, 154, 2, 74, 151, 2, 10, 233, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33, 157,
					229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219,
					229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222, 183,
					128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205, 181, 110,
					34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192, 33, 236,
					0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 115, 8, 199, 202, 174, 243, 3, 18, 107, 10, 34, 10, 32, 135,
					89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90,
					212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179,
					158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 33, 10, 31, 178, 74,
					58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8,
					139, 247, 174, 53, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17, 115, 101,
					115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/body/context/object/oversize", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid context: invalid target object: invalid length 33",
				b: []byte{26, 162, 2, 90, 159, 2, 34, 156, 2, 74, 153, 2, 10, 235, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 117, 8, 199, 202, 174, 243, 3, 18, 109, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
					29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 35, 10, 33, 178,
					74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246,
					8, 139, 247, 174, 53, 60, 1, 18, 41, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18, 17,
					115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 170, 137, 252, 156, 4}},
			{name: "header/split/parent/session/signature/scheme/overflow", err: "invalid header: invalid split header: invalid parent header: invalid session token: invalid body signature: scheme 2147483648 overflows int32",
				b: []byte{26, 166, 2, 90, 163, 2, 34, 160, 2, 74, 157, 2, 10, 234, 1, 10, 16, 118, 23, 219, 249, 117, 70, 64, 33,
					157, 229, 102, 253, 142, 52, 17, 144, 18, 27, 10, 25, 53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15,
					219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222, 26, 32, 8, 210, 204, 150, 183, 128, 222,
					183, 128, 228, 1, 16, 154, 222, 137, 183, 154, 198, 183, 155, 239, 1, 24, 186, 149, 174, 136, 210, 128, 205,
					181, 110, 34, 33, 2, 154, 144, 131, 197, 214, 154, 163, 35, 93, 133, 107, 35, 109, 3, 218, 20, 0, 255, 26, 192,
					33, 236, 0, 240, 183, 253, 76, 187, 136, 215, 164, 217, 42, 116, 8, 199, 202, 174, 243, 3, 18, 108, 10, 34, 10,
					32, 135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171,
					29, 90, 212, 15, 51, 86, 142, 101, 155, 141, 18, 34, 10, 32, 57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178,
					179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238, 61, 68, 58, 34, 189, 18, 34, 10, 32, 110,
					233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54,
					207, 56, 98, 99, 136, 207, 21, 18, 46, 10, 14, 115, 101, 115, 115, 105, 111, 110, 95, 115, 105, 103, 110, 101, 114, 18,
					17, 115, 101, 115, 115, 105, 111, 110, 32, 115, 105, 103, 110, 97, 116, 117, 114, 101, 24, 128, 128, 128, 128, 248,
					255, 255, 255, 255, 1}},
			{name: "header/split/parent/split/parent ID/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: invalid length 0",
				b: []byte{26, 8, 90, 6, 34, 4, 90, 2, 10, 0}},
			{name: "header/split/parent/split/parent ID/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: invalid length 31",
				b: []byte{26, 41, 90, 39, 34, 37, 90, 35, 10, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "header/split/parent/split/parent ID/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: invalid length 33",
				b: []byte{26, 43, 90, 41, 34, 39, 90, 37, 10, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "header/split/parent/split/parent ID/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: zero object ID",
				b: []byte{26, 42, 90, 40, 34, 38, 90, 36, 10, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/parent/split/previous/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: invalid length 0",
				b: []byte{26, 8, 90, 6, 34, 4, 90, 2, 18, 0}},
			{name: "header/split/parent/split/previous/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: invalid length 31",
				b: []byte{26, 41, 90, 39, 34, 37, 90, 35, 18, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "header/split/parent/split/previous/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: invalid length 33",
				b: []byte{26, 43, 90, 41, 34, 39, 90, 37, 18, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "header/split/parent/split/previous/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: zero object ID",
				b: []byte{26, 42, 90, 40, 34, 38, 90, 36, 18, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/parent/split/first/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: invalid length 0",
				b: []byte{26, 8, 90, 6, 34, 4, 90, 2, 58, 0}},
			{name: "header/split/parent/split/first/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: invalid length 31",
				b: []byte{26, 41, 90, 39, 34, 37, 90, 35, 58, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "header/split/parent/split/first/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: invalid length 33",
				b: []byte{26, 43, 90, 41, 34, 39, 90, 37, 58, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "header/split/parent/split/first/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: zero object ID",
				b: []byte{26, 42, 90, 40, 34, 38, 90, 36, 58, 34, 10, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "header/split/parent/split/ID/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid split ID: invalid UUID (got 15 bytes)",
				b: []byte{26, 23, 90, 21, 34, 19, 90, 17, 50, 15, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147}},
			{name: "header/split/parent/split/ID/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid split ID: invalid UUID (got 17 bytes)",
				b: []byte{26, 25, 90, 23, 34, 21, 90, 19, 50, 17, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41, 1}},
			{name: "header/split/parent/split/ID/wrong UUID version", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid split ID: wrong UUID version 3, expected 4",
				b: []byte{26, 24, 90, 22, 34, 20, 90, 18, 50, 16, 224, 132, 3, 80, 32, 44, 48, 184, 185, 32, 226, 201, 206, 196, 147, 41}},
			{name: "header/split/parent/split/children/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: invalid length 0",
				b: []byte{26, 80, 90, 78, 34, 76, 90, 74, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27,
					6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 0, 42, 34, 10, 32,
					206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101,
					24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "header/split/parent/split/children/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: invalid length 31",
				b: []byte{26, 113, 90, 111, 34, 109, 90, 107, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35,
					27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 33, 10, 31, 229,
					77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213,
					243, 57, 253, 127, 179, 42, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16,
					102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "header/split/parent/split/children/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: invalid length 33",
				b: []byte{26, 115, 90, 113, 34, 111, 90, 109, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35,
					27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 35, 10, 33, 229,
					77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213,
					243, 57, 253, 127, 179, 235, 1, 42, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153,
					133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "header/split/parent/split/children/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: zero object ID",
				b: []byte{26, 114, 90, 112, 34, 110, 90, 108, 42, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35,
					27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 42, 34, 10, 32, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 34, 10, 32,
					206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101,
					24, 235, 126, 173, 229, 161, 202, 197, 242}},
			{name: "header/split/parent/split/parent signature/scheme/overflow", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent signature: scheme 2147483648 overflows int32",
				b: []byte{26, 19, 90, 17, 34, 15, 90, 13, 26, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.Object).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var obj object.Object
	// zero
	require.NoError(t, obj.Unmarshal(nil))
	require.Zero(t, obj)

	// filled
	require.NoError(t, obj.Unmarshal(validBinObject))
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, validObject, obj)
}

func TestObject_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validObject, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONObject, string(b))
}

func TestObject_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(object.Object).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range []struct{ name, err, j string }{
			{name: "id/empty value", err: "invalid ID: invalid length 0",
				j: `{"objectID":{}}`},
			{name: "id/undersize", err: "invalid ID: invalid length 31",
				j: `{"objectID":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}`},
			{name: "id/oversize", err: "invalid ID: invalid length 33",
				j: `{"objectID":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}`},
			{name: "id/zero", err: "invalid ID: zero object ID",
				j: `{"objectID":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}`},
			{name: "signature/scheme/overflow", err: "invalid signature: scheme 2147483648 overflows int32",
				j: `{"signature":{"scheme":-2147483648}}`},
			{name: "header/owner/value/empty", err: "invalid header: invalid owner: invalid length 0, expected 25",
				j: `{"header":{"ownerID":{}}}`},
			{name: "header/owner/value/undersize", err: "invalid header: invalid owner: invalid length 24, expected 25",
				j: `{"header":{"ownerID":{"value":"NTsPBTSD/8YIYim45e2M1zSB09Zake2J"}}}`},
			{name: "header/owner/value/oversize", err: "invalid header: invalid owner: invalid length 26, expected 25",
				j: `{"header":{"ownerID":{"value":"NTsPBTSD/8YIYim45e2M1zSB09Zake2JmQE="}}}`},
			{name: "header/owner/value/wrong prefix", err: "invalid header: invalid owner: invalid prefix byte 0x42, expected 0x35",
				j: `{"header":{"ownerID":{"value":"QjsPBTSD/8YIYim45e2M1zSB09Zake2JmQ=="}}}`},
			{name: "header/owner/value/checksum mismatch", err: "invalid header: invalid owner: checksum mismatch",
				j: `{"header":{"ownerID":{"value":"NTsPBTSD/8YIYim45e2M1zSB09Zake2Jmg=="}}}`},
			{name: "header/container/empty value", err: "invalid header: invalid container: invalid length 0",
				j: `{"header":{"containerID":{}}}`},
			{name: "header/container/undersize", err: "invalid header: invalid container: invalid length 31",
				j: `{"header":{"containerID":{"value":"9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1zg=="}}}`},
			{name: "header/container/oversize", err: "invalid header: invalid container: invalid length 33",
				j: `{"header":{"containerID":{"value":"9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1ziIB"}}}`},
			{name: "header/container/zero", err: "invalid header: invalid container: zero container ID",
				j: `{"header":{"containerID":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}`},
			{name: "header/payload checksum/missing value", err: "invalid header: invalid payload checksum: missing value",
				j: `{"header":{"payloadHash":{}}}`},
			{name: "header/payload homomorphic checksum/missing value", err: "invalid header: invalid payload homomorphic checksum: missing value",
				j: `{"header":{"homomorphicHash":{}}}`},
			{name: "header/session/body/ID/undersize", err: "invalid header: invalid session token: invalid session ID: invalid UUID (got 15 bytes)",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQR"}}}}`},
			{name: "header/session/body/ID/oversize", err: "invalid header: invalid session token: invalid session ID: invalid UUID (got 17 bytes)",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQRkAE="}}}}`},
			{name: "header/session/body/ID/wrong UUID version", err: "invalid header: invalid session token: invalid session ID: wrong UUID version 3, expected 4",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGMCGd5Wb9jjQRkA=="}}}}`},
			{name: "header/session/body/issuer/value/empty", err: "invalid header: invalid session token: invalid session issuer: invalid length 0, expected 25",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQRkA==", "ownerID":{}}}}}`},
			{name: "header/session/body/issuer/value/undersize", err: "invalid header: invalid session token: invalid session issuer: invalid length 24, expected 25",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQRkA==", "ownerID":{"value":"NTsPBTSD/8YIYim45e2M1zSB09Zake2J"}}}}}`},
			{name: "header/session/body/issuer/value/oversize", err: "invalid header: invalid session token: invalid session issuer: invalid length 26, expected 25",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQRkA==", "ownerID":{"value":"NTsPBTSD/8YIYim45e2M1zSB09Zake2JmQE="}}}}}`},
			{name: "header/session/body/issuer/value/wrong prefix", err: "invalid header: invalid session token: invalid session issuer: invalid prefix byte 0x42, expected 0x35",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQRkA==", "ownerID":{"value":"QjsPBTSD/8YIYim45e2M1zSB09Zake2JmQ=="}}}}}`},
			{name: "header/session/body/issuer/value/checksum mismatch", err: "invalid header: invalid session token: invalid session issuer: checksum mismatch",
				j: `{"header":{"sessionToken":{"body":{"id":"dhfb+XVGQCGd5Wb9jjQRkA==", "ownerID":{"value":"NTsPBTSD/8YIYim45e2M1zSB09Zake2Jmg=="}}}}}`},
			{name: "header/session/body/context/wrong oneof", err: "invalid header: invalid session token: invalid context: invalid context *session.ContainerSessionContext",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "container": {}
   }
  }
 }
}
`},
			{name: "header/session/body/context/container/empty value", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 0",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {}
     }
    }
   }
  }
 }
}
`},
			{name: "header/session/body/context/container/undersize", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 31",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {
       "value": "9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1zg=="
      }
     }
    }
   }
  }
 }
}
`},
			{name: "header/session/body/context/container/oversize", err: "invalid header: invalid session token: invalid context: invalid container ID: invalid length 33",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {
       "value": "9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1ziIB"
      }
     }
    }
   }
  }
 }
}
`},
			{name: "header/session/body/context/object/empty value", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 0",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {
       "value": "h1mV27nR6Yng041Gwc34/uIecrH1qx1a1A8zVo5lm40="
      },
      "objects": [
       {"value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},
       {},
       {"value": "zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}
      ]
     }
    }
   }
  }
 }
}
`},
			{name: "header/session/body/context/object/undersize", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 31",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {
       "value": "h1mV27nR6Yng041Gwc34/uIecrH1qx1a1A8zVo5lm40="
      },
      "objects": [
       {"value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},
       {"value": "5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/sw=="},
       {"value": "zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}
      ]
     }
    }
   }
  }
 }
}
`},
			{name: "header/session/body/context/object/oversize", err: "invalid header: invalid session token: invalid context: invalid target object: invalid length 33",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {
       "value": "h1mV27nR6Yng041Gwc34/uIecrH1qx1a1A8zVo5lm40="
      },
      "objects": [
       {"value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},
       {"value": "5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+sB"},
       {"value": "zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}
      ]
     }
    }
   }
  }
 }
}
`},
			{name: "header/session/signature/scheme/overflow", err: "invalid header: invalid session token: invalid body signature: scheme 2147483648 overflows int32",
				j: `
{
 "header": {
  "sessionToken": {
   "body": {
    "id": "dhfb+XVGQCGd5Wb9jjQRkA==",
    "ownerID": {
     "value": "NfjDD8T+fBepxtAP2+U+lpef3Ung5Woq3g=="
    },
    "lifetime": {
     "exp": "16429376563136800338",
     "nbf": "17237208928641773338",
     "iat": "7956510363313998522"
    },
    "sessionKey": "ApqQg8XWmqMjXYVrI20D2hQA/xrAIewA8Lf9TLuI16TZ",
    "object": {
     "target": {
      "container": {
       "value": "h1mV27nR6Yng041Gwc34/uIecrH1qx1a1A8zVo5lm40="
      },
      "objects": [
       {
        "value": "OattKWkZkuCkWT2ys55DKCCaeq513Yqoh5XuPUQ6Ir0="
       },
       {
        "value": "bulm6IhE6RaeZDEUtV/bjzX67XFAGTALNs84YmOIzxU="
       }
      ]
     }
    }
   },
   "signature": {
    "key": "c2Vzc2lvbl9zaWduZXI=",
    "signature": "c2Vzc2lvbl9zaWduYXR1cmU=",
    "scheme": -2147483648
   }
  }
 }
}
`},
			{name: "attributes/no key", err: "invalid header: empty key of the attribute #1",
				j: `{"header": {"attributes": [{"key": "k1","value": "v1"},{"value": "v2"},{"key": "k3","value": "v3"}]}}`},
			{name: "attributes/no value", err: "invalid header: empty value of the attribute #1 (k2)",
				j: `{"header": {"attributes": [{"key": "k1","value": "v1"},{"key": "k2"},{"key": "k3","value": "v3"}]}}`},
			{name: "attributes/duplicated", err: "invalid header: duplicated attribute k1",
				j: `{"header": {"attributes": [{"key": "k1","value": "v1"},{"key": "k2", "value": "v2"},{"key": "k1","value": "v3"}]}}`},
			{name: "attributes/expiration", err: "invalid header: invalid expiration attribute (must be a uint): strconv.ParseUint: parsing \"foo\": invalid syntax",
				j: `{"header": {"attributes": [{"key": "k1","value": "v1"},{"key": "__NEOFS__EXPIRATION_EPOCH", "value": "foo"}]}}`},
			{name: "header/split/parent ID/empty value", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 0",
				j: `{"header": {"split":{"parent":{}}}}`},
			{name: "header/split/parent ID/undersize", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 31",
				j: `{"header": {"split":{"parent":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}}}`},
			{name: "header/split/parent ID/oversize", err: "invalid header: invalid split header: invalid parent split member ID: invalid length 33",
				j: `{"header": {"split":{"parent":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}}}`},
			{name: "header/split/parent ID/zero", err: "invalid header: invalid split header: invalid parent split member ID: zero object ID",
				j: `{"header": {"split":{"parent":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}`},
			{name: "header/split/previous/empty value", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 0",
				j: `{"header": {"split":{"previous":{}}}}`},
			{name: "header/split/previous/undersize", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 31",
				j: `{"header": {"split":{"previous":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}}}`},
			{name: "header/split/previous/oversize", err: "invalid header: invalid split header: invalid previous split member ID: invalid length 33",
				j: `{"header": {"split":{"previous":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}}}`},
			{name: "header/split/previous/zero", err: "invalid header: invalid split header: invalid previous split member ID: zero object ID",
				j: `{"header": {"split":{"previous":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}`},
			{name: "header/split/first/empty value", err: "invalid header: invalid split header: invalid first split member ID: invalid length 0",
				j: `{"header": {"split":{"first":{}}}}`},
			{name: "header/split/first/undersize", err: "invalid header: invalid split header: invalid first split member ID: invalid length 31",
				j: `{"header": {"split":{"first":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}}}`},
			{name: "header/split/first/oversize", err: "invalid header: invalid split header: invalid first split member ID: invalid length 33",
				j: `{"header": {"split":{"first":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}}}`},
			{name: "header/split/first/zero", err: "invalid header: invalid split header: invalid first split member ID: zero object ID",
				j: `{"header": {"split":{"first":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}`},
			{name: "header/split/ID/undersize", err: "invalid header: invalid split header: invalid split ID: invalid UUID (got 15 bytes)",
				j: `{"header": {"split":{"splitID":"4IQDUCAsRbi5IOLJzsST"}}}`},
			{name: "header/split/ID/oversize", err: "invalid header: invalid split header: invalid split ID: invalid UUID (got 17 bytes)",
				j: `{"header": {"split":{"splitID":"4IQDUCAsRbi5IOLJzsSTKQE="}}}`},
			{name: "header/split/ID/wrong UUID version", err: "invalid header: invalid split header: invalid split ID: wrong UUID version 3, expected 4",
				j: `{"header": {"split":{"splitID":"4IQDUCAsMLi5IOLJzsSTKQ=="}}}`},
			{name: "header/split/children/empty value", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 0",
				j: `{"header": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, {}, {"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}`},
			{name: "header/split/children/undersize", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 31",
				j: `{"header": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, {"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/sw=="}, {"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}`},
			{name: "header/split/children/oversize", err: "invalid header: invalid split header: invalid child split member ID #1: invalid length 33",
				j: `{"header": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, {"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+sB"}, {"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}`},
			{name: "header/split/children/zero", err: "invalid header: invalid split header: invalid child split member ID #1: zero object ID",
				j: `{"header": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, {"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}, {"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}`},
			{name: "header/split/parent signature/scheme/overflow", err: "invalid header: invalid split header: invalid parent signature: scheme 2147483648 overflows int32",
				j: `{"header": {"split":{"parentSignature":{"key":"cHViXzE=","signature":"c2lnXzE=","scheme":-2147483648}}}}`},
			{name: "header/split/parent/owner/value/empty", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 0, expected 25",
				j: `{"header": {"split": {"parentHeader": {"ownerID": {}}}}}`},
			{name: "header/split/parent/owner/value/undersize", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 24, expected 25",
				j: `{"header": {"split": {"parentHeader": {"ownerID": {"value": "NTsPBTSD/8YIYim45e2M1zSB09Zake2J"}}}}}`},
			{name: "header/split/parent/owner/value/oversize", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid length 26, expected 25",
				j: `{"header": {"split": {"parentHeader": {"ownerID": {"value": "NTsPBTSD/8YIYim45e2M1zSB09Zake2JmQE="}}}}}`},
			{name: "header/split/parent/owner/value/wrong prefix", err: "invalid header: invalid split header: invalid parent header: invalid owner: invalid prefix byte 0x42, expected 0x35",
				j: `{"header": {"split": {"parentHeader": {"ownerID": {"value": "QjsPBTSD/8YIYim45e2M1zSB09Zake2JmQ=="}}}}}`},
			{name: "header/split/parent/owner/value/checksum mismatch", err: "invalid header: invalid split header: invalid parent header: invalid owner: checksum mismatch",
				j: `{"header": {"split": {"parentHeader": {"ownerID": {"value": "NTsPBTSD/8YIYim45e2M1zSB09Zake2Jmg=="}}}}}`},
			{name: "header/split/parent/container/empty value", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 0",
				j: `{"header": {"split": {"parentHeader": {"containerID": {}}}}}`},
			{name: "header/split/parent/container/undersize", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 31",
				j: `{"header": {"split": {"parentHeader": {"containerID": {"value": "9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1zg=="}}}}}`},
			{name: "header/split/parent/container/oversize", err: "invalid header: invalid split header: invalid parent header: invalid container: invalid length 33",
				j: `{"header": {"split": {"parentHeader": {"containerID": {"value": "9V6kz9npr0t7ma4IFIdgzLNdt/q0/6K23txjfYh1ziIB"}}}}}`},
			{name: "header/split/parent/container/zero", err: "invalid header: invalid split header: invalid parent header: invalid container: zero container ID",
				j: `{"header": {"split": {"parentHeader": {"containerID": {"value": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}}`},
			{name: "header/split/parent/payload checksum/missing value", err: "invalid header: invalid split header: invalid parent header: invalid payload checksum: missing value",
				j: `{"header": {"split": {"parentHeader": {"payloadHash":{"type":1974315742}}}}}`},
			{name: "header/split/parent/payload homomorphic checksum/missing value", err: "invalid header: invalid split header: invalid parent header: invalid payload homomorphic checksum: missing value",
				j: `{"header": {"split": {"parentHeader": {"homomorphicHash":{"type":1974315742}}}}}`},
			{name: "header/split/parent/split/parent ID/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: invalid length 0",
				j: `{"header": {"split": {"parentHeader": {"split":{"parent":{}}}}}}`},
			{name: "header/split/parent/split/parent ID/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: invalid length 31",
				j: `{"header": {"split": {"parentHeader": {"split":{"parent":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}}}}}`},
			{name: "header/split/parent/split/parent ID/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: invalid length 33",
				j: `{"header": {"split": {"parentHeader": {"split":{"parent":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}}}}}`},
			{name: "header/split/parent/split/parent ID/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent split member ID: zero object ID",
				j: `{"header": {"split": {"parentHeader": {"split":{"parent":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}}}`},
			{name: "header/split/parent/split/previous/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: invalid length 0",
				j: `{"header": {"split": {"parentHeader": {"split":{"previous":{}}}}}}`},
			{name: "header/split/parent/split/previous/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: invalid length 31",
				j: `{"header": {"split": {"parentHeader": {"split":{"previous":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}}}}}`},
			{name: "header/split/parent/split/previous/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: invalid length 33",
				j: `{"header": {"split": {"parentHeader": {"split":{"previous":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}}}}}`},
			{name: "header/split/parent/split/previous/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid previous split member ID: zero object ID",
				j: `{"header": {"split": {"parentHeader": {"split":{"previous":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}}}`},
			{name: "header/split/parent/split/first/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: invalid length 0",
				j: `{"header": {"split": {"parentHeader": {"split":{"first":{}}}}}}`},
			{name: "header/split/parent/split/first/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: invalid length 31",
				j: `{"header": {"split": {"parentHeader": {"split":{"first":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}}}}}`},
			{name: "header/split/parent/split/first/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: invalid length 33",
				j: `{"header": {"split": {"parentHeader": {"split":{"first":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}}}}}`},
			{name: "header/split/parent/split/first/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid first split member ID: zero object ID",
				j: `{"header": {"split": {"parentHeader": {"split":{"first":{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}}}}}`},
			{name: "header/split/parent/split/ID/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid split ID: invalid UUID (got 15 bytes)",
				j: `{"header": {"split": {"parentHeader": {"split":{"splitID":"4IQDUCAsRbi5IOLJzsST"}}}}}`},
			{name: "header/split/parent/split/ID/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid split ID: invalid UUID (got 17 bytes)",
				j: `{"header": {"split": {"parentHeader": {"split":{"splitID":"4IQDUCAsRbi5IOLJzsSTKQE="}}}}}`},
			{name: "header/split/parent/split/ID/wrong UUID version", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid split ID: wrong UUID version 3, expected 4",
				j: `{"header": {"split": {"parentHeader": {"split":{"splitID":"4IQDUCAsMLi5IOLJzsSTKQ=="}}}}}`},
			{name: "header/split/parent/split/children/empty value", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: invalid length 0",
				j: `{"header": {"split": {"parentHeader": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},{},{"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}}}`},
			{name: "header/split/parent/split/children/undersize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: invalid length 31",
				j: `{"header": {"split": {"parentHeader": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},{"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/sw=="},{"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}}}`},
			{name: "header/split/parent/split/children/oversize", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: invalid length 33",
				j: `{"header": {"split": {"parentHeader": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},{"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+sB"},{"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}}}`},
			{name: "header/split/parent/split/children/zero", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid child split member ID #1: zero object ID",
				j: `{"header": {"split": {"parentHeader": {"split":{"children":[{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},{"value":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},{"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="}]}}}}}`},
			{name: "header/split/parent/split/parent signature/scheme/overflow", err: "invalid header: invalid split header: invalid parent header: invalid split header: invalid parent signature: scheme 2147483648 overflows int32",
				j: `{"header": {"split": {"parentHeader": {"split":{"parentSignature":{"key":"cHViXzE=","signature":"c2lnXzE=","scheme":-2147483648}}}}}}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.Object).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var obj object.Object
	// zero
	require.NoError(t, obj.UnmarshalJSON([]byte("{}")))
	require.Zero(t, obj)

	// filled
	require.NoError(t, obj.UnmarshalJSON([]byte(validJSONObject)))
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, validObject, obj)
}

func TestObject_HeaderLen(t *testing.T) {
	require.EqualValues(t, 0, object.Object{}.HeaderLen())
	require.EqualValues(t, 1047, validObject.HeaderLen())
}
