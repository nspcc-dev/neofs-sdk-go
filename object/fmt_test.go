package object_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

// corresponds to validObject.
var validSignedObject = []byte{10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
	18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60}

func TestObject_CalculateAndSetPayloadChecksum(t *testing.T) {
	var obj object.Object
	_, ok := obj.PayloadChecksum()
	require.False(t, ok)

	obj.CalculateAndSetPayloadChecksum()
	cs, ok := obj.PayloadChecksum()
	require.True(t, ok)
	require.Equal(t, checksum.SHA256, cs.Type())
	require.Equal(t, emptySHA256Hash[:], cs.Value())
	require.NoError(t, obj.VerifyPayloadChecksum())

	obj.SetPayload(anyValidRegularPayload)
	obj.CalculateAndSetPayloadChecksum()
	cs, ok = obj.PayloadChecksum()
	require.True(t, ok)
	require.Equal(t, checksum.SHA256, cs.Type())
	require.Equal(t, anyValidPayloadChecksum[:], cs.Value())
	require.NoError(t, obj.VerifyPayloadChecksum())
}

func TestObject_VerifyPayloadChecksum(t *testing.T) {
	var obj object.Object
	require.EqualError(t, obj.VerifyPayloadChecksum(), "payload checksum is not set")

	obj.SetPayload(anyValidRegularPayload)
	obj.SetPayloadChecksum(checksum.NewSHA256(anyValidPayloadChecksum))
	require.NoError(t, obj.VerifyPayloadChecksum())

	// mutate payload
	for i := range anyValidRegularPayload {
		b := bytes.Clone(anyValidRegularPayload)
		b[i]++
		obj.SetPayload(b)
		require.EqualError(t, obj.VerifyPayloadChecksum(), "payload checksum mismatch")
	}

	obj.SetPayload(anyValidRegularPayload)

	// mutate checksum
	for i := range anyValidPayloadChecksum {
		b := bytes.Clone(anyValidPayloadChecksum[:])
		b[i]++
		obj.SetPayloadChecksum(checksum.NewSHA256([sha256.Size]byte(b)))
		require.EqualError(t, obj.VerifyPayloadChecksum(), "payload checksum mismatch")
	}
}

func TestObject_CalculateAndSetID(t *testing.T) {
	var obj object.Object
	id, err := obj.CalculateID()
	require.NoError(t, err)
	require.EqualValues(t, emptySHA256Hash, id)

	require.NoError(t, obj.CalculateAndSetID())
	require.EqualValues(t, emptySHA256Hash, obj.GetID())
	require.NoError(t, obj.VerifyID())

	id, err = validObject.CalculateID()
	require.NoError(t, err)
	require.Equal(t, validObjectID, id)

	// any header field affects ID which is essentially a checksum. Therefore, we
	// set each field individually - it must be taken into account when calculating.
	for _, tc := range []struct {
		name   string
		id     oid.ID
		setHdr func(*object.Object)
	}{
		{name: "version", id: oid.ID{149, 211, 24, 216, 96, 10, 95, 80, 221, 0, 122, 130, 195, 255, 143, 90, 65, 138, 234, 225,
			187, 5, 124, 176, 203, 167, 171, 64, 28, 105, 36, 19}, setHdr: func(obj *object.Object) {
			v := version.New(1817223862, 2425735897)
			obj.SetVersion(&v)
		}},
		{name: "container", id: oid.ID{189, 201, 250, 161, 127, 136, 1, 171, 38, 55, 162, 94, 244, 236, 25, 194, 249, 76, 108, 157,
			106, 207, 230, 103, 205, 172, 108, 95, 245, 216, 87, 233}, setHdr: func(obj *object.Object) {
			obj.SetContainerID(anyValidContainers[0])
		}},
		{name: "owner", id: oid.ID{121, 153, 185, 254, 28, 165, 125, 149, 195, 153, 23, 170, 202, 220, 131, 60, 228, 232, 183,
			213, 78, 101, 149, 77, 60, 167, 17, 217, 224, 176, 209, 117}, setHdr: func(obj *object.Object) {
			obj.SetOwner(anyValidUsers[0])
		}},
		{name: "creation epoch", id: oid.ID{233, 231, 109, 191, 11, 224, 100, 8, 141, 85, 178, 69, 64, 34, 122, 227, 68, 39, 55,
			146, 77, 191, 134, 0, 208, 181, 136, 128, 253, 69, 244, 209}, setHdr: func(obj *object.Object) {
			obj.SetCreationEpoch(11023562854130131584)
		}},
		{name: "payload size", id: oid.ID{43, 209, 94, 92, 255, 43, 251, 30, 12, 184, 202, 9, 81, 120, 168, 105, 118, 131, 49,
			205, 145, 52, 122, 54, 223, 189, 165, 14, 156, 143, 206, 42}, setHdr: func(obj *object.Object) {
			obj.SetPayloadSize(14301110394027098694)
		}},
		{name: "payload checksum", id: oid.ID{222, 91, 43, 92, 236, 51, 230, 199, 141, 85, 161, 148, 4, 137, 203, 159, 207, 146,
			34, 106, 193, 97, 113, 133, 170, 215, 207, 130, 160, 197, 119, 41}, setHdr: func(obj *object.Object) {
			obj.SetPayloadChecksum(checksum.NewSHA256(anyValidPayloadChecksum))
		}},
		{name: "type", id: oid.ID{88, 19, 117, 205, 15, 152, 61, 10, 161, 139, 239, 17, 173, 95, 115, 67, 244, 161, 72, 65, 192, 31,
			214, 201, 193, 121, 18, 157, 137, 131, 232, 101}, setHdr: func(obj *object.Object) {
			obj.SetType(273597346)
		}},
		{name: "homomorphic checksum", id: oid.ID{169, 103, 36, 160, 165, 242, 215, 67, 43, 37, 115, 178, 199, 253, 211, 68, 204,
			76, 225, 188, 194, 180, 249, 109, 92, 82, 173, 253, 9, 131, 86, 55}, setHdr: func(obj *object.Object) {
			cs, err := checksum.NewFromData(checksum.TillichZemor, anyValidRegularPayload)
			require.NoError(t, err)
			obj.SetPayloadHomomorphicHash(cs)
		}},
		{name: "session token", id: oid.ID{198, 180, 102, 30, 21, 150, 211, 56, 197, 91, 91, 223, 10, 18, 156, 171, 238, 183, 219,
			184, 181, 198, 152, 220, 242, 212, 20, 196, 32, 183, 246, 91}, setHdr: func(obj *object.Object) {
			obj.SetSessionToken(&anyValidObjectToken)
		}},
		{name: "attributes", id: oid.ID{111, 113, 39, 89, 222, 193, 249, 95, 9, 92, 207, 177, 208, 184, 181, 55, 122, 93, 42,
			237, 171, 27, 5, 85, 61, 78, 14, 57, 139, 11, 0, 113}, setHdr: func(obj *object.Object) {
			obj.SetAttributes(object.NewAttribute("k1", "v1"), object.NewAttribute("k2", "v2"))
		}},
		{name: "parent ID", id: oid.ID{251, 8, 196, 121, 53, 217, 66, 125, 74, 220, 6, 136, 236, 147, 196, 32, 129, 176, 8, 252,
			205, 111, 44, 93, 229, 164, 15, 195, 239, 148, 174, 17}, setHdr: func(obj *object.Object) {
			obj.SetParentID(anyValidIDs[0])
		}},
		{name: "previous ID", id: oid.ID{41, 248, 197, 11, 1, 21, 230, 178, 41, 223, 168, 32, 126, 251, 218, 203, 247, 20, 170,
			115, 225, 81, 196, 171, 228, 226, 110, 237, 114, 117, 86, 8}, setHdr: func(obj *object.Object) {
			obj.SetPreviousID(anyValidIDs[0])
		}},
		{name: "parent signature", id: oid.ID{49, 71, 84, 191, 129, 210, 146, 193, 211, 42, 45, 53, 68, 91, 99, 151, 85, 118, 89, 159,
			255, 4, 8, 201, 15, 245, 0, 248, 162, 93, 0, 222}, setHdr: func(obj *object.Object) {
			var par object.Object
			par.SetSignature(&anyValidSignatures[0])
			obj.SetParent(&par)
		}},
		{name: "parent header", id: oid.ID{51, 4, 113, 101, 21, 249, 152, 60, 58, 17, 23, 29, 114, 211, 60, 101, 150, 13, 1, 159,
			134, 91, 182, 175, 70, 115, 15, 139, 189, 99, 120, 130}, setHdr: func(obj *object.Object) {
			var par object.Object
			par.SetContainerID(anyValidContainers[0])
			par.SetOwner(anyValidUsers[0])
			obj.SetParent(&par)
		}},
		{name: "children", id: oid.ID{101, 182, 100, 255, 18, 42, 237, 120, 194, 67, 161, 195, 147, 203, 57, 106, 232, 132, 127,
			151, 1, 232, 4, 231, 71, 112, 124, 55, 97, 224, 119, 11}, setHdr: func(obj *object.Object) {
			obj.SetChildren(anyValidIDs[0], anyValidIDs[1])
		}},
		{name: "split ID", id: oid.ID{229, 194, 131, 2, 215, 171, 31, 69, 103, 30, 215, 27, 18, 139, 167, 113, 183, 37, 24, 48, 17,
			209, 49, 62, 179, 240, 146, 244, 21, 132, 61, 102}, setHdr: func(obj *object.Object) {
			obj.SetSplitID(anyValidSplitID)
		}},
		{name: "first ID", id: oid.ID{55, 115, 112, 34, 102, 135, 159, 161, 3, 119, 234, 68, 54, 181, 49, 40, 151, 106, 147, 223, 35,
			60, 204, 90, 255, 113, 199, 189, 11, 120, 132, 75}, setHdr: func(obj *object.Object) {
			obj.SetFirstID(anyValidIDs[0])
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var obj object.Object
			tc.setHdr(&obj)

			id, err := obj.CalculateID()
			require.NoError(t, err)
			require.Equal(t, tc.id, id)

			require.NoError(t, obj.CalculateAndSetID())
			require.Equal(t, tc.id, obj.GetID())
			require.NoError(t, obj.VerifyID())
		})
	}
}

func TestObject_VerifyID(t *testing.T) {
	var obj object.Object
	require.ErrorIs(t, obj.VerifyID(), oid.ErrZero)

	validObject.CopyTo(&obj)

	id, err := obj.CalculateID()
	require.NoError(t, err)
	obj.SetID(id)
	require.NoError(t, obj.VerifyID())

	obj.SetID(oidtest.OtherID(id))
	require.EqualError(t, obj.VerifyID(), "incorrect object identifier")
}

func TestContainer_Sign(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		anySigner := neofscryptotest.Signer()
		require.ErrorIs(t, new(object.Object).Sign(anySigner), oid.ErrZero)

		var objWithID object.Object
		objWithID.SetID(oidtest.ID())
		require.Error(t, objWithID.Sign(neofscryptotest.FailSigner(anySigner)))
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

	var obj object.Object
	validObject.CopyTo(&obj)

	/* non-deterministic schemes */
	assertECDSACommon := func(signer neofscrypto.Signer) []byte {
		require.NoError(t, obj.Sign(signer))
		sig := obj.Signature()
		require.NotNil(t, sig)
		require.Equal(t, signer.Scheme(), sig.Scheme())
		x, y := elliptic.UnmarshalCompressed(elliptic.P256(), sig.PublicKeyBytes())
		require.NotNil(t, x)
		require.Equal(t, ecdsaPriv.X, x)
		require.Equal(t, ecdsaPriv.Y, y)
		return sig.Value()
	}

	// SHA-512
	sig := assertECDSACommon(neofsecdsa.Signer(ecdsaPriv))
	require.Len(t, sig, 65)
	require.EqualValues(t, 0x4, sig[0])
	h512 := sha512.Sum512(validSignedObject)
	r, s := new(big.Int).SetBytes(sig[1:33]), new(big.Int).SetBytes(sig[33:])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h512[:], r, s))

	// WalletConnect
	sig = assertECDSACommon(neofsecdsa.SignerWalletConnect(ecdsaPriv))
	require.Len(t, sig, 80)
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(validSignedObject)))
	base64.StdEncoding.Encode(b64, validSignedObject)
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
	// deterministic ECDSA with SHA-256 hashing (RFC 6979)
	sig = assertECDSACommon(neofsecdsa.SignerRFC6979(ecdsaPriv))
	require.Equal(t, []byte{
		117, 254, 164, 56, 148, 113, 8, 171, 216, 251, 102, 211, 37, 52, 181, 63, 206, 226, 22, 24, 36, 90, 249, 59, 247, 3, 213,
		46, 70, 16, 128, 211, 34, 165, 212, 22, 129, 61, 103, 36, 0, 132, 171, 27, 209, 184, 243, 123, 105, 96, 152, 47, 12, 93,
		33, 155, 177, 252, 161, 219, 83, 95, 102, 213,
	}, sig)
	h256 = sha256.Sum256(validSignedObject)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))
}

func TestObject_SignedData(t *testing.T) {
	require.Equal(t, validSignedObject, validObject.SignedData())
}

func TestObject_VerifySignature(t *testing.T) {
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

	var obj object.Object
	require.False(t, obj.VerifySignature())
	obj.SetSignature(new(neofscrypto.Signature))
	require.False(t, obj.VerifySignature())
	obj.SetID(validObjectID)
	require.False(t, obj.VerifySignature())

	var sig neofscrypto.Signature
	for i, tc := range []struct {
		scheme neofscrypto.Scheme
		sig    []byte // of validObject
	}{
		{scheme: neofscrypto.ECDSA_SHA512, sig: []byte{4, 247, 16, 200, 217, 135, 233, 25, 35, 155, 171, 67, 37, 235, 20, 201,
			232, 229, 31, 37, 156, 184, 254, 21, 75, 78, 40, 240, 70, 194, 203, 146, 119, 142, 225, 254, 26, 40, 105, 10, 236,
			240, 82, 53, 142, 181, 175, 246, 25, 212, 216, 55, 129, 246, 62, 151, 67, 200, 147, 64, 95, 92, 172, 94, 138}},
		{scheme: neofscrypto.ECDSA_DETERMINISTIC_SHA256, sig: []byte{169, 18, 104, 126, 238, 153, 176, 13, 101, 3, 76, 31, 39,
			104, 133, 97, 209, 144, 19, 199, 114, 72, 229, 217, 140, 151, 162, 39, 132, 63, 40, 124, 20, 229, 95, 227, 66, 77,
			93, 50, 115, 176, 216, 232, 123, 252, 61, 101, 215, 161, 70, 178, 231, 125, 99, 62, 71, 34, 154, 129, 230, 166, 60, 172}},
		{scheme: neofscrypto.ECDSA_WALLETCONNECT, sig: []byte{100, 244, 204, 67, 168, 37, 153, 29, 55, 45, 121, 64, 162, 193,
			67, 59, 5, 119, 191, 112, 44, 195, 5, 48, 241, 218, 193, 162, 214, 85, 159, 61, 29, 235, 213, 241, 248, 6, 119,
			154, 152, 163, 174, 80, 226, 208, 238, 87, 21, 117, 248, 72, 2, 253, 90, 112, 168, 46, 90, 243, 136, 115, 160, 138,
			145, 255, 142, 198, 175, 123, 38, 24, 131, 233, 193, 48, 1, 218, 192, 222}},
	} {
		sig.SetScheme(tc.scheme)
		sig.SetPublicKeyBytes(pub)
		sig.SetValue(tc.sig)
		obj.SetSignature(&sig)
		require.True(t, obj.VerifySignature(), i)
		// corrupt public key
		for k := range pub {
			pubCp := bytes.Clone(pub)
			pubCp[k]++
			sig.SetPublicKeyBytes(pubCp)
			obj.SetSignature(&sig)
			require.False(t, obj.VerifySignature(), i)
		}
		// corrupt signature
		for k := range tc.sig {
			sigBytesCp := bytes.Clone(tc.sig)
			sigBytesCp[k]++
			sig.SetValue(sigBytesCp)
			obj.SetSignature(&sig)
			require.False(t, obj.VerifySignature(), i)
		}
	}
}

func TestObject_SetIDWithSignature(t *testing.T) {
	anySigner := neofscryptotest.Signer()
	var obj object.Object
	validObject.CopyTo(&obj)

	require.Error(t, obj.SetIDWithSignature(neofscryptotest.FailSigner(anySigner)))

	require.NoError(t, obj.SetIDWithSignature(anySigner))
	require.Equal(t, validObjectID, obj.GetID())
	require.NoError(t, obj.VerifyID())
	require.True(t, obj.VerifySignature())
}

func TestObject_SetVerificationFields(t *testing.T) {
	anySigner := neofscryptotest.Signer()

	var obj object.Object
	validObject.CopyTo(&obj)
	require.Error(t, obj.SetVerificationFields(neofscryptotest.FailSigner(anySigner)))
	require.NoError(t, obj.SetVerificationFields(anySigner))

	require.NoError(t, obj.VerifyID())
	require.NoError(t, obj.VerifyPayloadChecksum())
	require.True(t, obj.VerifySignature())
	require.NoError(t, obj.CheckHeaderVerificationFields())
	require.NoError(t, obj.CheckVerificationFields())
}
