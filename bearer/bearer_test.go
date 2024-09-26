package bearer_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/big"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

// Bearer token components.
const (
	anyValidExp = 32058350
	anyValidIat = 1209418302
	anyValidNbf = 93843742391
)

// Crypto.
const (
	anyValidSignatureScheme = 1208743532
)

var (
	anyValidIssuerPublicKeyBytes = []byte{3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239,
		245, 67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218}
	anyValidSignatureBytes = []byte("any_signature") // valid structurally, not logically
	anyValidSignature      = neofscrypto.NewSignatureFromRawKey(anyValidSignatureScheme, anyValidIssuerPublicKeyBytes, anyValidSignatureBytes)
)

// Other NeoFS stuff.
var (
	anyValidContainerID = cid.ID{243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66,
		6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227}
	anyValidIssuer = user.ID{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57,
		160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
	anyValidSubject = user.ID{53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102,
		233, 115, 142, 143, 145, 26, 229, 252, 61, 36, 160, 242, 243}
	anyValidUsers = []user.ID{
		{53, 192, 107, 60, 173, 127, 3, 69, 236, 239, 37, 107, 173, 167, 143, 161, 20, 113, 133, 199, 150, 139, 182, 171, 184},
		{53, 14, 64, 221, 23, 249, 186, 133, 93, 139, 98, 131, 163, 110, 134, 7, 6, 164, 198, 136, 124, 32, 202, 88, 107},
		{53, 85, 27, 45, 129, 140, 18, 120, 223, 85, 88, 64, 164, 131, 242, 184, 191, 237, 33, 233, 187, 82, 171, 117, 111},
	}
)

var (
	anyValidFilters = []eacl.Filter{
		eacl.ConstructFilter(eacl.FilterHeaderType(4509681), "key_54093643", eacl.Match(949385), "val_34811040"),
		eacl.ConstructFilter(eacl.FilterHeaderType(582984), "key_1298432", eacl.Match(7539428), "val_8243258"),
	}
	anyValidTargets = []eacl.Target{
		eacl.NewTargetByRole(eacl.Role(690857412)),
		eacl.NewTargetByAccounts(anyValidUsers),
	}
	anyValidRecords = []eacl.Record{
		eacl.ConstructRecord(eacl.Action(5692342), eacl.Operation(12943052), []eacl.Target{anyValidTargets[0]}, anyValidFilters[0]),
		eacl.ConstructRecord(eacl.Action(43658603), eacl.Operation(94383138), anyValidTargets, anyValidFilters...),
	}
	anyValidEACL     = eacl.NewTableForContainer(anyValidContainerID, anyValidRecords)
	validBearerToken bearer.Token // set by init
)

func init() {
	validBearerToken.SetEACLTable(anyValidEACL)
	validBearerToken.ForUser(anyValidSubject)
	validBearerToken.SetIssuer(anyValidIssuer)
	validBearerToken.SetExp(anyValidExp)
	validBearerToken.SetIat(anyValidIat)
	validBearerToken.SetNbf(anyValidNbf)
	validBearerToken.AttachSignature(anyValidSignature)
}

// corresponds to validBearerToken.
var validSignedBearerToken = []byte{
	10, 153, 2, 10, 4, 8, 2, 16, 16, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62,
	66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 26, 57, 8, 204, 253, 149, 6, 16, 182, 183,
	219, 2, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48, 57, 51, 54, 52, 51, 34,
	12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 34, 6, 8, 196, 203, 182, 201, 2, 26, 177, 1, 8, 162, 216, 128,
	45, 16, 235, 218, 232, 20, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48, 57,
	51, 54, 52, 51, 34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 26, 35, 8, 200, 202, 35, 16, 228, 149, 204,
	3, 26, 11, 107, 101, 121, 95, 49, 50, 57, 56, 52, 51, 50, 34, 11, 118, 97, 108, 95, 56, 50, 52, 51, 50, 53, 56, 34, 6, 8,
	196, 203, 182, 201, 2, 34, 81, 18, 25, 53, 192, 107, 60, 173, 127, 3, 69, 236, 239, 37, 107, 173, 167, 143, 161, 20, 113,
	133, 199, 150, 139, 182, 171, 184, 18, 25, 53, 14, 64, 221, 23, 249, 186, 133, 93, 139, 98, 131, 163, 110, 134, 7, 6, 164,
	198, 136, 124, 32, 202, 88, 107, 18, 25, 53, 85, 27, 45, 129, 140, 18, 120, 223, 85, 88, 64, 164, 131, 242, 184, 191,
	237, 33, 233, 187, 82, 171, 117, 111, 18, 27, 10, 25, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142,
	143, 145, 26, 229, 252, 61, 36, 160, 242, 243, 26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24, 190,
	132, 217, 192, 4, 34, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20, 254,
	30, 138, 195, 17, 92,
}

// corresponds to validBearerToken.
var validBinBearerToken = []byte{
	10, 234, 2, 10, 153, 2, 10, 4, 8, 2, 16, 16, 18, 34, 10, 32, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21,
	254, 62, 66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 26, 57, 8, 204, 253, 149, 6, 16,
	182, 183, 219, 2, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48, 57, 51, 54, 52,
	51, 34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 34, 6, 8, 196, 203, 182, 201, 2, 26, 177, 1, 8, 162, 216,
	128, 45, 16, 235, 218, 232, 20, 26, 37, 8, 241, 159, 147, 2, 16, 137, 249, 57, 26, 12, 107, 101, 121, 95, 53, 52, 48,
	57, 51, 54, 52, 51, 34, 12, 118, 97, 108, 95, 51, 52, 56, 49, 49, 48, 52, 48, 26, 35, 8, 200, 202, 35, 16, 228, 149,
	204, 3, 26, 11, 107, 101, 121, 95, 49, 50, 57, 56, 52, 51, 50, 34, 11, 118, 97, 108, 95, 56, 50, 52, 51, 50, 53, 56, 34, 6,
	8, 196, 203, 182, 201, 2, 34, 81, 18, 25, 53, 192, 107, 60, 173, 127, 3, 69, 236, 239, 37, 107, 173, 167, 143, 161, 20,
	113, 133, 199, 150, 139, 182, 171, 184, 18, 25, 53, 14, 64, 221, 23, 249, 186, 133, 93, 139, 98, 131, 163, 110, 134, 7, 6,
	164, 198, 136, 124, 32, 202, 88, 107, 18, 25, 53, 85, 27, 45, 129, 140, 18, 120, 223, 85, 88, 64, 164, 131, 242, 184,
	191, 237, 33, 233, 187, 82, 171, 117, 111, 18, 27, 10, 25, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115,
	142, 143, 145, 26, 229, 252, 61, 36, 160, 242, 243, 26, 18, 8, 238, 215, 164, 15, 16, 183, 189, 151, 204, 221, 2, 24,
	190, 132, 217, 192, 4, 34, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203, 20,
	254, 30, 138, 195, 17, 92, 18, 56, 10, 33, 3, 202, 217, 142, 98, 209, 190, 188, 145, 123, 174, 21, 173, 239, 239, 245,
	67, 148, 205, 119, 58, 223, 219, 209, 220, 113, 215, 134, 228, 101, 249, 34, 218, 18, 13, 97, 110, 121, 95, 115, 105,
	103, 110, 97, 116, 117, 114, 101, 24, 236, 236, 175, 192, 4,
}

// corresponds to validBearerToken.
var validJSONBearerToken = `
{
 "body": {
  "eaclTable": {
   "version": {
    "major": 2,
    "minor": 16
   },
   "containerID": {
    "value": "8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OM="
   },
   "records": [
    {
     "operation": 12943052,
     "action": 5692342,
     "filters": [
      {
       "headerType": 4509681,
       "matchType": 949385,
       "key": "key_54093643",
       "value": "val_34811040"
      }
     ],
     "targets": [
      {
       "role": 690857412
      }
     ]
    },
    {
     "operation": 94383138,
     "action": 43658603,
     "filters": [
      {
       "headerType": 4509681,
       "matchType": 949385,
       "key": "key_54093643",
       "value": "val_34811040"
      },
      {
       "headerType": 582984,
       "matchType": 7539428,
       "key": "key_1298432",
       "value": "val_8243258"
      }
     ],
     "targets": [
      {
       "role": 690857412
      },
      {
       "keys": [
        "NcBrPK1/A0Xs7yVrraePoRRxhceWi7aruA==",
        "NQ5A3Rf5uoVdi2KDo26GBwakxoh8IMpYaw==",
        "NVUbLYGMEnjfVVhApIPyuL/tIem7Uqt1bw=="
       ]
      }
     ]
    }
   ]
  },
  "ownerID": {
   "value": "NZMOukLD9zMO+ZFm6XOOj5Ea5fw9JKDy8w=="
  },
  "lifetime": {
   "exp": "32058350",
   "nbf": "93843742391",
   "iat": "1209418302"
  },
  "issuer": {
   "value": "NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="
  }
 },
 "signature": {
  "key": "A8rZjmLRvryRe64Vre/v9UOUzXc639vR3HHXhuRl+SLa",
  "signature": "YW55X3NpZ25hdHVyZQ==",
  "scheme": 1208743532
 }
}
`

func TestToken_SetEACLTable(t *testing.T) {
	var val bearer.Token
	require.True(t, val.EACLTable().IsZero())

	val.SetEACLTable(anyValidEACL)
	require.Equal(t, anyValidEACL, val.EACLTable())

	e := eacltest.Table()
	val.SetEACLTable(e)
	require.Equal(t, e, val.EACLTable())
}

func TestToken_ForUser(t *testing.T) {
	var val bearer.Token
	usr1 := usertest.ID()
	usr2 := usertest.OtherID(usr1)

	require.True(t, val.AssertUser(usr1))
	require.True(t, val.AssertUser(usr2))

	val.ForUser(usr1)
	require.True(t, val.AssertUser(usr1))
	require.False(t, val.AssertUser(usr2))

	val.ForUser(usr2)
	require.False(t, val.AssertUser(usr1))
	require.True(t, val.AssertUser(usr2))
}

func testLifetimeClaim(t *testing.T, setter func(*bearer.Token, uint64), getter func(bearer.Token) uint64) {
	var val bearer.Token
	require.Zero(t, getter(val))
	setter(&val, 12094032)
	require.EqualValues(t, 12094032, getter(val))
	setter(&val, 5469830342)
	require.EqualValues(t, 5469830342, getter(val))
}

func TestToken_SetIat(t *testing.T) {
	testLifetimeClaim(t, (*bearer.Token).SetIat, bearer.Token.Iat)
}

func TestToken_SetNbf(t *testing.T) {
	testLifetimeClaim(t, (*bearer.Token).SetNbf, bearer.Token.Nbf)
}

func TestToken_SetExp(t *testing.T) {
	testLifetimeClaim(t, (*bearer.Token).SetExp, bearer.Token.Exp)
}

func TestToken_ValidAt(t *testing.T) {
	var val bearer.Token

	require.True(t, val.ValidAt(0))
	require.False(t, val.InvalidAt(0))
	require.False(t, val.ValidAt(1))
	require.True(t, val.InvalidAt(1))

	val.SetIat(1)
	val.SetNbf(2)
	val.SetExp(4)

	require.False(t, val.ValidAt(0))
	require.True(t, val.InvalidAt(0))
	require.False(t, val.ValidAt(1))
	require.True(t, val.InvalidAt(1))
	require.True(t, val.ValidAt(2))
	require.False(t, val.InvalidAt(2))
	require.True(t, val.ValidAt(3))
	require.False(t, val.InvalidAt(3))
	require.True(t, val.ValidAt(4))
	require.False(t, val.InvalidAt(4))
	require.False(t, val.ValidAt(5))
	require.True(t, val.InvalidAt(5))
}

func TestToken_AssertContainer(t *testing.T) {
	var val bearer.Token
	cnr1 := cidtest.ID()
	cnr2 := cidtest.OtherID(cnr1)

	require.True(t, val.AssertContainer(cnr1))
	require.True(t, val.AssertContainer(cnr2))

	var eaclTable eacl.Table
	val.SetEACLTable(eaclTable)
	require.True(t, val.AssertContainer(cnr1))
	require.True(t, val.AssertContainer(cnr2))

	eaclTable.SetCID(cnr1)
	val.SetEACLTable(eaclTable)
	require.True(t, val.AssertContainer(cnr1))
	require.False(t, val.AssertContainer(cnr2))

	eaclTable.SetCID(cnr2)
	val.SetEACLTable(eaclTable)
	require.False(t, val.AssertContainer(cnr1))
	require.True(t, val.AssertContainer(cnr2))
}

func TestToken_Sign(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		require.Error(t, new(bearer.Token).Sign(usertest.FailSigner(usertest.User())))
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
	tok := validBearerToken

	/* non-deterministic schemes */
	assertECDSACommon := func(signer neofscrypto.Signer) []byte {
		scheme := signer.Scheme()
		require.NoError(t, tok.Sign(user.NewSigner(signer, anyValidIssuer)), scheme)
		require.Equal(t, anyValidIssuer, tok.Issuer())
		sig, ok := tok.Signature()
		require.True(t, ok)
		require.Equal(t, signer.Scheme(), scheme)
		sigBytes := sig.Value()
		x, y := elliptic.UnmarshalCompressed(elliptic.P256(), sig.PublicKeyBytes())
		require.NotNil(t, x, scheme)
		require.Equal(t, ecdsaPriv.X, x)
		require.Equal(t, ecdsaPriv.Y, y)
		return sigBytes
	}

	// SHA-512
	sig := assertECDSACommon(neofsecdsa.Signer(ecdsaPriv))
	require.Len(t, sig, 65)
	require.EqualValues(t, 0x4, sig[0])
	h512 := sha512.Sum512(validSignedBearerToken)
	r, s := new(big.Int).SetBytes(sig[1:33]), new(big.Int).SetBytes(sig[33:])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h512[:], r, s))

	// WalletConnect
	sig = assertECDSACommon(neofsecdsa.SignerWalletConnect(ecdsaPriv))
	require.Len(t, sig, 80)
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(validSignedBearerToken)))
	base64.StdEncoding.Encode(b64, validSignedBearerToken)
	payloadLen := 2*16 + len(b64)
	b := make([]byte, 4+io.GetVarSize(payloadLen)+payloadLen+2)
	n := copy(b, []byte{0x01, 0x00, 0x01, 0xf0})
	n += io.PutVarUint(b[n:], uint64(payloadLen))
	n += hex.Encode(b[n:], sig[64:])
	n += copy(b[n:], b64)
	copy(b[n:], []byte{0x00, 0x00})
	h256 := sha256.Sum256(b)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))

	/* deterministic schemes */
	sig = assertECDSACommon(neofsecdsa.SignerRFC6979(ecdsaPriv))
	require.Equal(t, []byte{
		30, 35, 64, 192, 126, 242, 239, 239, 137, 114, 85, 153, 28, 107, 99, 254, 173, 189, 12, 215, 27, 179, 145, 49, 48,
		73, 106, 85, 70, 99, 69, 163, 18, 127, 187, 34, 173, 126, 195, 35, 133, 106, 59, 193, 82, 39, 208, 58, 99, 162, 250,
		226, 181, 231, 58, 194, 57, 201, 33, 217, 68, 9, 188, 39}, sig)
	h256 = sha256.Sum256(validSignedBearerToken)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))
}

func TestToken_SignedData(t *testing.T) {
	require.Equal(t, validSignedBearerToken, validBearerToken.SignedData())
}

func TestToken_VerifySignature(t *testing.T) {
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
	for i, tc := range []struct {
		scheme neofscrypto.Scheme
		sig    []byte // of validBearerToken
	}{
		{scheme: neofscrypto.ECDSA_SHA512, sig: []byte{
			4, 249, 194, 70, 39, 54, 168, 159, 35, 7, 51, 46, 8, 172, 125, 118, 249, 22, 50, 112, 79, 219, 14, 105, 103,
			184, 180, 11, 65, 11, 87, 255, 91, 112, 14, 26, 5, 60, 155, 86, 250, 41, 37, 160, 132, 58, 63, 139, 213, 248,
			71, 201, 15, 17, 11, 149, 193, 47, 35, 13, 185, 41, 19, 206, 44,
		}},
		{scheme: neofscrypto.ECDSA_DETERMINISTIC_SHA256, sig: []byte{
			234, 98, 185, 176, 113, 150, 200, 224, 62, 229, 134, 202, 187, 123, 222, 213, 233, 31, 205, 120, 85,
			118, 125, 12, 165, 36, 247, 67, 142, 247, 113, 100, 221, 219, 57, 135, 243, 255, 140, 245, 224, 32, 7, 73,
			170, 229, 157, 86, 168, 108, 217, 18, 166, 114, 130, 25, 14, 233, 223, 198, 142, 86, 212, 183,
		}},
		{scheme: neofscrypto.ECDSA_WALLETCONNECT, sig: []byte{
			180, 200, 26, 152, 82, 31, 220, 218, 249, 173, 15, 248, 31, 61, 83, 207, 225, 115, 37, 17, 8, 73, 8, 179,
			76, 160, 16, 77, 65, 110, 197, 103, 169, 191, 99, 241, 64, 54, 172, 178, 253, 79, 254, 123, 115, 112, 232, 64,
			75, 21, 174, 147, 85, 217, 223, 62, 184, 12, 248, 197, 172, 99, 121, 44, 82, 156, 192, 231, 119, 159, 111,
			123, 23, 110, 67, 32, 102, 129, 98, 254,
		}},
	} {
		sig.SetScheme(tc.scheme)
		sig.SetPublicKeyBytes(pub)
		sig.SetValue(tc.sig)
		tok := validBearerToken
		tok.AttachSignature(sig)
		require.True(t, tok.VerifySignature(), i)
		for k := range pub {
			pubCp := bytes.Clone(pub)
			pubCp[k]++
			sig.SetPublicKeyBytes(pubCp)
			tok.AttachSignature(sig)
			require.False(t, tok.VerifySignature(), i)
		}
		for k := range tc.sig {
			sigBytesCp := bytes.Clone(tc.sig)
			sigBytesCp[k]++
			sig.SetValue(sigBytesCp)
			tok.AttachSignature(sig)
			require.False(t, tok.VerifySignature(), i)
		}
	}
}

func TestToken_UnmarshalSignedData(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(bearer.Token).UnmarshalSignedData([]byte("Hello, world!"))
			require.ErrorContains(t, err, "decode body")
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "eacl/invalid container/empty value", err: "invalid eACL: invalid container ID: invalid length 0",
				b: []byte{10, 2, 18, 0}},
			{name: "eacl/invalid container/undersized value", err: "invalid eACL: invalid container ID: invalid length 31",
				b: []byte{10, 35, 18, 33, 10, 31, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62,
					66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252}},
			{name: "eacl/invalid container/oversized value", err: "invalid eACL: invalid container ID: invalid length 33",
				b: []byte{10, 37, 18, 35, 10, 33, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62,
					66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 1}},
			{name: "body/subject/value/empty", err: "invalid target user: invalid length 0, expected 25",
				b: []byte{18, 0}},
			{name: "subject/value/undersize", err: "invalid target user: invalid length 24, expected 25",
				b: []byte{18, 26, 10, 24, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242}},
			{name: "subject/value/oversize", err: "invalid target user: invalid length 26, expected 25",
				b: []byte{18, 28, 10, 26, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243, 1}},
			{name: "subject/value/wrong prefix", err: "invalid target user: invalid prefix byte 0x42, expected 0x35",
				b: []byte{18, 27, 10, 25, 66, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243}},
			{name: "subject/value/checksum mismatch", err: "invalid target user: checksum mismatch",
				b: []byte{18, 27, 10, 25, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 244}},
			{name: "issuer/value/empty", err: "invalid issuer: invalid length 0, expected 25",
				b: []byte{34, 0}},
			{name: "issuer/value/undersize", err: "invalid issuer: invalid length 24, expected 25",
				b: []byte{34, 26, 10, 24, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242}},
			{name: "issuer/value/oversize", err: "invalid issuer: invalid length 26, expected 25",
				b: []byte{34, 28, 10, 26, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243, 1}},
			{name: "issuer/value/wrong prefix", err: "invalid issuer: invalid prefix byte 0x42, expected 0x35",
				b: []byte{34, 27, 10, 25, 66, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243}},
			{name: "issuer/value/checksum mismatch", err: "invalid issuer: checksum mismatch",
				b: []byte{34, 27, 10, 25, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 244}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(bearer.Token).UnmarshalSignedData(tc.b), tc.err)
			})
		}
	})

	var val bearer.Token
	// zero
	require.NoError(t, val.UnmarshalSignedData(nil))
	require.True(t, val.EACLTable().IsZero())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	require.True(t, val.AssertContainer(cidtest.ID()))
	require.True(t, val.AssertUser(usertest.ID()))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	err := val.UnmarshalSignedData(validSignedBearerToken)
	require.NoError(t, err)
	val.AttachSignature(anyValidSignature)
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, validBearerToken, val)
}

func TestToken_AttachSignature(t *testing.T) {
	var val bearer.Token
	_, ok := val.Signature()
	require.False(t, ok)
	val.AttachSignature(anyValidSignature)
	sig, ok := val.Signature()
	require.True(t, ok)
	require.Equal(t, anyValidSignature, sig)
}

func TestToken_ReadFromV2(t *testing.T) {
	var lt acl.TokenLifetime
	lt.SetExp(anyValidExp)
	lt.SetIat(anyValidIat)
	lt.SetNbf(anyValidNbf)

	var ms refs.OwnerID
	ms.SetValue(anyValidSubject[:])

	var mi refs.OwnerID
	mi.SetValue(anyValidIssuer[:])

	var mb acl.BearerTokenBody
	mb.SetEACL(anyValidEACL.ToV2())
	mb.SetOwnerID(&ms)
	mb.SetIssuer(&mi)
	mb.SetLifetime(&lt)

	var msig refs.Signature
	anyValidSignature.WriteToV2(&msig)

	var m acl.BearerToken
	m.SetBody(&mb)
	m.SetSignature(&msig)

	var val bearer.Token
	require.NoError(t, val.ReadFromV2(m))

	require.Equal(t, anyValidEACL, val.EACLTable())
	require.Equal(t, anyValidIssuer, val.Issuer())
	require.True(t, val.AssertUser(anyValidSubject))

	require.EqualValues(t, anyValidExp, val.Exp())
	require.EqualValues(t, anyValidIat, val.Iat())
	require.EqualValues(t, anyValidNbf, val.Nbf())

	sig, ok := val.Signature()
	require.True(t, ok)
	require.Equal(t, anyValidSignature, sig)

	// reset optional fields
	mb.SetIssuer(nil)
	mb.SetOwnerID(nil)
	val2 := val
	require.NoError(t, val2.ReadFromV2(m))
	require.Zero(t, val2.Issuer())
	require.True(t, val2.AssertUser(usertest.ID()))

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*acl.BearerToken)
		}{
			{name: "body/missing", err: "missing token body",
				corrupt: func(m *acl.BearerToken) { m.SetBody(nil) }},
			{name: "body/eacl/missing", err: "missing eACL table",
				corrupt: func(m *acl.BearerToken) { m.GetBody().SetEACL(nil) }},
			{name: "body/eacl/invalid container/nil value", err: "invalid eACL: invalid container ID: invalid length 0",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetEACL().SetContainerID(new(refs.ContainerID)) }},
			{name: "body/eacl/invalid container/empty value", err: "invalid eACL: invalid container ID: invalid length 0", corrupt: func(m *acl.BearerToken) {
				var mc refs.ContainerID
				mc.SetValue([]byte{})
				m.GetBody().GetEACL().SetContainerID(&mc)
			}},
			{name: "body/eacl/invalid container/undersized value", err: "invalid eACL: invalid container ID: invalid length 31", corrupt: func(m *acl.BearerToken) {
				var mc refs.ContainerID
				mc.SetValue(make([]byte, 31))
				m.GetBody().GetEACL().SetContainerID(&mc)
			}},
			{name: "body/eacl/invalid container/oversized value", err: "invalid eACL: invalid container ID: invalid length 33", corrupt: func(m *acl.BearerToken) {
				var mc refs.ContainerID
				mc.SetValue(make([]byte, 33))
				m.GetBody().GetEACL().SetContainerID(&mc)
			}},
			{name: "body/subject/value/nil", err: "invalid target user: invalid length 0, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetOwnerID().SetValue(nil) }},
			{name: "body/subject/value/empty", err: "invalid target user: invalid length 0, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetOwnerID().SetValue([]byte{}) }},
			{name: "body/subject/value/undersize", err: "invalid target user: invalid length 24, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetOwnerID().SetValue(make([]byte, 24)) }},
			{name: "body/subject/value/oversize", err: "invalid target user: invalid length 26, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetOwnerID().SetValue(make([]byte, 26)) }},
			{name: "body/subject/value/wrong prefix", err: "invalid target user: invalid prefix byte 0x42, expected 0x35",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetOwnerID().GetValue()[0] = 0x42 }},
			{name: "body/subject/value/checksum mismatch", err: "invalid target user: checksum mismatch",
				corrupt: func(m *acl.BearerToken) {
					v := m.GetBody().GetOwnerID().GetValue()
					v[len(v)-1]++
				}},
			{name: "body/lifetime/missing", err: "missing token lifetime",
				corrupt: func(m *acl.BearerToken) { m.GetBody().SetLifetime(nil) }},
			{name: "body/issuer/value/nil", err: "invalid issuer: invalid length 0, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetIssuer().SetValue(nil) }},
			{name: "body/issuer/value/empty", err: "invalid issuer: invalid length 0, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetIssuer().SetValue([]byte{}) }},
			{name: "body/issuer/value/undersize", err: "invalid issuer: invalid length 24, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetIssuer().SetValue(make([]byte, 24)) }},
			{name: "body/issuer/value/oversize", err: "invalid issuer: invalid length 26, expected 25",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetIssuer().SetValue(make([]byte, 26)) }},
			{name: "body/issuer/value/wrong prefix", err: "invalid issuer: invalid prefix byte 0x42, expected 0x35",
				corrupt: func(m *acl.BearerToken) { m.GetBody().GetIssuer().GetValue()[0] = 0x42 }},
			{name: "body/issuer/value/checksum mismatch", err: "invalid issuer: checksum mismatch",
				corrupt: func(m *acl.BearerToken) {
					v := m.GetBody().GetIssuer().GetValue()
					v[len(v)-1]++
				}},
			{name: "signature/missing", err: "missing body signature",
				corrupt: func(m *acl.BearerToken) { m.SetSignature(nil) }},
			{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32",
				corrupt: func(m *acl.BearerToken) { m.GetSignature().SetScheme(math.MaxInt32 + 1) }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				var m acl.BearerToken
				st.WriteToV2(&m)
				tc.corrupt(&m)
				require.EqualError(t, new(bearer.Token).ReadFromV2(m), tc.err)
			})
		}
	})
}

func TestToken_WriteToV2(t *testing.T) {
	var val bearer.Token
	var m acl.BearerToken

	// zero
	val.WriteToV2(&m)
	require.Zero(t, m.GetBody())
	require.Zero(t, m.GetSignature())

	// filled
	val.SetEACLTable(anyValidEACL)
	val.ForUser(anyValidSubject)
	val.SetIssuer(anyValidIssuer)
	val.SetExp(anyValidExp)
	val.SetIat(anyValidIat)
	val.SetNbf(anyValidNbf)
	val.AttachSignature(anyValidSignature)

	val.WriteToV2(&m)

	body := m.GetBody()
	require.NotNil(t, body)
	require.Equal(t, anyValidSubject[:], body.GetOwnerID().GetValue())
	require.Equal(t, anyValidIssuer[:], body.GetIssuer().GetValue())

	lt := body.GetLifetime()
	require.EqualValues(t, anyValidExp, lt.GetExp())
	require.EqualValues(t, anyValidIat, lt.GetIat())
	require.EqualValues(t, anyValidNbf, lt.GetNbf())

	sig := m.GetSignature()
	require.NotNil(t, sig)
	require.EqualValues(t, anyValidSignatureScheme, sig.GetScheme())
	require.Equal(t, anyValidIssuerPublicKeyBytes, sig.GetKey())
	require.Equal(t, anyValidSignatureBytes, sig.GetSign())

	e := m.GetBody().GetEACL()
	require.NotNil(t, e)
	require.EqualValues(t, 2, e.GetVersion().GetMajor())
	require.EqualValues(t, 16, e.GetVersion().GetMinor())
	require.Equal(t, anyValidContainerID[:], e.GetContainerID().GetValue())

	rs := e.GetRecords()
	require.Len(t, rs, 2)
	// record#0
	require.EqualValues(t, 5692342, rs[0].GetAction())
	require.EqualValues(t, 12943052, rs[0].GetOperation())

	ts := rs[0].GetTargets()
	require.Len(t, ts, 1)
	require.EqualValues(t, 690857412, ts[0].GetRole())
	require.Zero(t, ts[0].GetKeys())

	fs := rs[0].GetFilters()
	require.Len(t, fs, 1)
	require.EqualValues(t, 4509681, fs[0].GetHeaderType())
	require.EqualValues(t, 949385, fs[0].GetMatchType())
	require.Equal(t, "key_54093643", fs[0].GetKey())
	require.Equal(t, "val_34811040", fs[0].GetValue())
	// record#1
	require.EqualValues(t, 43658603, rs[1].GetAction())
	require.EqualValues(t, 94383138, rs[1].GetOperation())

	ts = rs[1].GetTargets()
	require.Len(t, ts, 2)
	require.EqualValues(t, 690857412, ts[0].GetRole())
	require.Zero(t, ts[0].GetKeys())
	require.Zero(t, ts[1].GetRole())
	ks := ts[1].GetKeys()
	require.Len(t, ks, len(anyValidUsers))
	for i := range ks {
		require.Equal(t, anyValidUsers[i][:], ks[i])
	}

	fs = rs[1].GetFilters()
	require.Len(t, fs, 2)
	require.EqualValues(t, 4509681, fs[0].GetHeaderType())
	require.EqualValues(t, 949385, fs[0].GetMatchType())
	require.Equal(t, "key_54093643", fs[0].GetKey())
	require.Equal(t, "val_34811040", fs[0].GetValue())
	require.EqualValues(t, 582984, fs[1].GetHeaderType())
	require.EqualValues(t, 7539428, fs[1].GetMatchType())
	require.Equal(t, "key_1298432", fs[1].GetKey())
	require.Equal(t, "val_8243258", fs[1].GetValue())
}

func TestToken_Marshal(t *testing.T) {
	require.Equal(t, validBinBearerToken, validBearerToken.Marshal())
}

func TestToken_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(bearer.Token).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name string
			err  string
			b    []byte
		}{
			{name: "body/eacl/invalid container/empty value", err: "invalid eACL: invalid container ID: invalid length 0",
				b: []byte{10, 4, 10, 2, 18, 0}},
			{name: "body/eacl/invalid container/undersized value", err: "invalid eACL: invalid container ID: invalid length 31",
				b: []byte{10, 37, 10, 35, 18, 33, 10, 31, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62,
					66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252}},
			{name: "body/eacl/invalid container/oversized value", err: "invalid eACL: invalid container ID: invalid length 33",
				b: []byte{10, 39, 10, 37, 18, 35, 10, 33, 243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62,
					66, 6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227, 1}},
			{name: "body/subject/value/empty", err: "invalid target user: invalid length 0, expected 25",
				b: []byte{10, 2, 18, 0}},
			{name: "body/subject/value/undersize", err: "invalid target user: invalid length 24, expected 25",
				b: []byte{10, 28, 18, 26, 10, 24, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242}},
			{name: "body/subject/value/oversize", err: "invalid target user: invalid length 26, expected 25",
				b: []byte{10, 30, 18, 28, 10, 26, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243, 1}},
			{name: "body/subject/value/wrong prefix", err: "invalid target user: invalid prefix byte 0x42, expected 0x35",
				b: []byte{10, 29, 18, 27, 10, 25, 66, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243}},
			{name: "body/subject/value/checksum mismatch", err: "invalid target user: checksum mismatch",
				b: []byte{10, 29, 18, 27, 10, 25, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 244}},
			{name: "body/issuer/value/empty", err: "invalid issuer: invalid length 0, expected 25",
				b: []byte{10, 2, 34, 0}},
			{name: "body/issuer/value/undersize", err: "invalid issuer: invalid length 24, expected 25",
				b: []byte{10, 28, 34, 26, 10, 24, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242}},
			{name: "body/issuer/value/oversize", err: "invalid issuer: invalid length 26, expected 25",
				b: []byte{10, 30, 34, 28, 10, 26, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243, 1}},
			{name: "body/issuer/value/wrong prefix", err: "invalid issuer: invalid prefix byte 0x42, expected 0x35",
				b: []byte{10, 29, 34, 27, 10, 25, 66, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 243}},
			{name: "body/issuer/value/checksum mismatch", err: "invalid issuer: checksum mismatch",
				b: []byte{10, 29, 34, 27, 10, 25, 53, 147, 14, 186, 66, 195, 247, 51, 14, 249, 145, 102, 233, 115, 142, 143,
					145, 26, 229, 252, 61, 36, 160, 242, 244}},
			{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32",
				b: []byte{18, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(bearer.Token).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var val bearer.Token
	// zero
	require.NoError(t, val.Unmarshal(nil))

	require.True(t, val.EACLTable().IsZero())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	require.True(t, val.AssertContainer(cidtest.ID()))
	require.True(t, val.AssertUser(usertest.ID()))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	err := val.Unmarshal(validBinBearerToken)
	require.NoError(t, err)
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, validBearerToken, val)
}

func TestToken_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validBearerToken, "", " ")
	require.NoError(t, err)
	if string(b) != validJSONBearerToken {
		// protojson is inconsistent https://github.com/golang/protobuf/issues/1121
		var val bearer.Token
		require.NoError(t, val.UnmarshalJSON(b))
		t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
		require.Equal(t, validBearerToken, val)
	}
}

func TestToken_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(bearer.Token).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range []struct{ name, err, j string }{
			{name: "body/eacl/invalid container/empty value", err: "invalid eACL: invalid container ID: invalid length 0",
				j: `{"body":{"eaclTable":{"containerID":{}}}}`},
			{name: "body/eacl/invalid container/undersized value", err: "invalid eACL: invalid container ID: invalid length 31",
				j: `{"body":{"eaclTable":{"containerID":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/A=="}}}}`},
			{name: "body/eacl/invalid container/oversized value", err: "invalid eACL: invalid container ID: invalid length 33",
				j: `{"body":{"eaclTable":{"containerID":{"value":"8/VLxjBrjXn/MTOoFf4+QgaTKyNj8qMUGh6T8E9y/OMB"}}}}`},
			{name: "body/subject/value/empty", err: "invalid target user: invalid length 0, expected 25",
				j: `{"body":{"ownerID":{}}}`},
			{name: "body/subject/value/undersize", err: "invalid target user: invalid length 24, expected 25",
				j: `{"body":{"ownerID":{"value":"NZMOukLD9zMO+ZFm6XOOj5Ea5fw9JKDy"}}}`},
			{name: "body/subject/value/oversize", err: "invalid target user: invalid length 26, expected 25",
				j: `{"body":{"ownerID":{"value":"NZMOukLD9zMO+ZFm6XOOj5Ea5fw9JKDy8wE="}}}`},
			{name: "body/subject/value/wrong prefix", err: "invalid target user: invalid prefix byte 0x42, expected 0x35",
				j: `{"body":{"ownerID":{"value":"QpMOukLD9zMO+ZFm6XOOj5Ea5fw9JKDy8w=="}}}`},
			{name: "body/subject/value/checksum mismatch", err: "invalid target user: checksum mismatch",
				j: `{"body":{"ownerID":{"value":"NZMOukLD9zMO+ZFm6XOOj5Ea5fw9JKDy9A=="}}}`},
			{name: "body/issuer/value/empty", err: "invalid issuer: invalid length 0, expected 25",
				j: `{"body":{"issuer":{}}}`},
			{name: "body/issuer/value/undersize", err: "invalid issuer: invalid length 24, expected 25",
				j: `{"body":{"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMR"}}}`},
			{name: "body/issuer/value/oversize", err: "invalid issuer: invalid length 26, expected 25",
				j: `{"body":{"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXAE="}}}`},
			{name: "body/issuer/value/wrong prefix", err: "invalid issuer: invalid prefix byte 0x42, expected 0x35",
				j: `{"body":{"issuer":{"value":"QjMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}}}`},
			{name: "body/issuer/value/checksum mismatch", err: "invalid issuer: checksum mismatch",
				j: `{"body":{"issuer":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXQ=="}}}`},
			{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32",
				j: `{"signature":{"scheme":-2147483648}}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(bearer.Token).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var val bearer.Token
	// zero
	require.NoError(t, val.UnmarshalJSON([]byte("{}")))
	require.True(t, val.EACLTable().IsZero())
	require.Zero(t, val.Issuer())
	require.Zero(t, val.Exp())
	require.Zero(t, val.Iat())
	require.Zero(t, val.Nbf())
	require.True(t, val.AssertContainer(cidtest.ID()))
	require.True(t, val.AssertUser(usertest.ID()))
	_, ok := val.Signature()
	require.False(t, ok)

	// filled
	require.NoError(t, val.UnmarshalJSON([]byte(validJSONBearerToken)))
	t.Skip("https://github.com/nspcc-dev/neofs-sdk-go/issues/606")
	require.Equal(t, validBearerToken, val)
}

func TestToken_ResolveIssuer(t *testing.T) {
	var val bearer.Token
	require.True(t, val.Issuer().IsZero())
	require.True(t, val.Issuer().IsZero())

	var sig neofscrypto.Signature
	sig.SetPublicKeyBytes([]byte("not_a_public_key"))
	val.AttachSignature(sig)
	require.True(t, val.ResolveIssuer().IsZero())
	require.True(t, val.Issuer().IsZero())

	pk1 := []byte{2, 151, 13, 13, 100, 79, 94, 79, 149, 226, 182, 159, 160, 48, 20, 197, 220, 219, 177, 30, 251, 43, 2, 226, 189,
		56, 222, 144, 72, 49, 161, 86, 182}
	usr1 := user.ID{53, 9, 197, 19, 207, 204, 207, 163, 25, 35, 54, 153, 124, 116, 113, 223, 237, 82, 9, 128, 85, 165, 139, 101, 153}
	sig.SetPublicKeyBytes(pk1)
	val.AttachSignature(sig)
	require.Equal(t, usr1, val.ResolveIssuer())
	require.True(t, val.Issuer().IsZero())

	pk2 := []byte{3, 249, 82, 207, 175, 111, 206, 80, 104, 165, 65, 162, 174, 87, 84, 94, 87, 96, 13, 125, 139, 80, 70, 229, 46,
		249, 68, 68, 158, 149, 89, 26, 186}
	usr2 := user.ID{53, 43, 179, 214, 121, 57, 226, 21, 143, 17, 146, 143, 147, 213, 51, 78, 81, 41, 40, 183, 252, 45, 241, 70, 252}
	sig.SetPublicKeyBytes(pk2)
	val.AttachSignature(sig)
	require.Equal(t, usr2, val.ResolveIssuer())
	require.True(t, val.Issuer().IsZero())
}

func TestToken_SetIssuer(t *testing.T) {
	var token bearer.Token
	require.True(t, token.Issuer().IsZero())

	usr1 := usertest.ID()
	token.SetIssuer(usr1)
	require.Equal(t, usr1, token.Issuer())
	require.Equal(t, usr1, token.ResolveIssuer())

	usr2 := usertest.OtherID(usr1)
	token.SetIssuer(usr2)
	require.Equal(t, usr2, token.Issuer())
	require.Equal(t, usr2, token.ResolveIssuer())
}

func TestToken_SigningKeyBytes(t *testing.T) {
	var tok bearer.Token
	require.Zero(t, tok.SigningKeyBytes())

	var sig neofscrypto.Signature
	sig.SetPublicKeyBytes(anyValidIssuerPublicKeyBytes)
	tok.AttachSignature(sig)
	require.Equal(t, anyValidIssuerPublicKeyBytes, tok.SigningKeyBytes())
}
