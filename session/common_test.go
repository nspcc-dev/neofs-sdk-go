package session_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/google/uuid"
	apisession "github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

// Session components.
const (
	anyValidExp = 32058350
	anyValidIat = 1209418302
	anyValidNbf = 93843742391
)

var (
	anyValidSessionID = uuid.UUID{99, 24, 111, 70, 22, 172, 72, 20, 139, 187, 175, 98, 10, 255, 231, 188}
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
	anyValidSessionKey     = &neofsecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes([]byte{149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119, 71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136}),
		Y:     new(big.Int).SetBytes([]byte{118, 201, 238, 8, 178, 41, 96, 3, 163, 197, 31, 58, 106, 218, 104, 47, 106, 153, 180, 68, 109, 243, 62, 31, 159, 17, 104, 134, 134, 97, 117, 52}),
	}
	// corresponds to anyValidSessionKey.
	anyValidSessionKeyBytes = []byte{2, 149, 43, 50, 196, 91, 177, 62, 131, 233, 126, 241, 177, 13, 78, 96, 94, 119,
		71, 55, 179, 8, 53, 241, 79, 2, 1, 95, 85, 78, 45, 197, 136}
)

// Other NeoFS stuff.
var (
	anyValidUserID = user.ID{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57,
		160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
	anyValidContainerID = cid.ID{243, 245, 75, 198, 48, 107, 141, 121, 255, 49, 51, 168, 21, 254, 62, 66,
		6, 147, 43, 35, 99, 242, 163, 20, 26, 30, 147, 240, 79, 114, 252, 227}
)

type invalidProtoTokenTestcase struct {
	name, err string
	corrupt   func(*apisession.Token)
}

var invalidProtoTokenCommonTestcases = []invalidProtoTokenTestcase{
	{name: "missing body", err: "missing token body", corrupt: func(st *apisession.Token) {
		st.SetBody(nil)
	}},
	{name: "missing body signature", err: "missing body signature", corrupt: func(st *apisession.Token) {
		st.SetSignature(nil)
	}},
	{name: "body/ID/nil", err: "missing session ID", corrupt: func(st *apisession.Token) {
		st.GetBody().SetID(nil)
	}},
	{name: "body/ID/empty", err: "missing session ID", corrupt: func(st *apisession.Token) {
		st.GetBody().SetID([]byte{})
	}},
	{name: "body/ID/undersize", err: "invalid session ID: invalid UUID (got 15 bytes)", corrupt: func(st *apisession.Token) {
		st.GetBody().SetID(make([]byte, 15))
	}},
	{name: "body/ID/wrong UUID version", err: "invalid session UUID version 3", corrupt: func(st *apisession.Token) {
		st.GetBody().GetID()[6] = 3 << 4
	}},
	{name: "body/ID/oversize", err: "invalid session ID: invalid UUID (got 17 bytes)", corrupt: func(st *apisession.Token) {
		st.GetBody().SetID(make([]byte, 17))
	}},
	{name: "body/issuer", err: "missing session issuer", corrupt: func(st *apisession.Token) {
		st.GetBody().SetOwnerID(nil)
	}},
	{name: "body/issuer/value/nil", err: "invalid session issuer: invalid length 0, expected 25", corrupt: func(st *apisession.Token) {
		st.GetBody().GetOwnerID().SetValue(nil)
	}},
	{name: "body/issuer/value/empty", err: "invalid session issuer: invalid length 0, expected 25", corrupt: func(st *apisession.Token) {
		st.GetBody().GetOwnerID().SetValue([]byte{})
	}},
	{name: "body/issuer/value/undersize", err: "invalid session issuer: invalid length 24, expected 25", corrupt: func(st *apisession.Token) {
		st.GetBody().GetOwnerID().SetValue(make([]byte, 24))
	}},
	{name: "body/issuer/value/oversize", err: "invalid session issuer: invalid length 26, expected 25", corrupt: func(st *apisession.Token) {
		st.GetBody().GetOwnerID().SetValue(make([]byte, 26))
	}},
	{name: "body/issuer/value/wrong prefix", err: "invalid session issuer: invalid prefix byte 0x42, expected 0x35", corrupt: func(st *apisession.Token) {
		st.GetBody().GetOwnerID().GetValue()[0] = 0x42
	}},
	{name: "body/issuer/value/checksum mismatch", err: "invalid session issuer: checksum mismatch", corrupt: func(st *apisession.Token) {
		v := st.GetBody().GetOwnerID().GetValue()
		v[len(v)-1]++
	}},
	{name: "body/lifetime", err: "missing token lifetime", corrupt: func(st *apisession.Token) {
		st.GetBody().SetLifetime(nil)
	}},
	{name: "body/session key/nil", err: "missing session public key", corrupt: func(st *apisession.Token) {
		st.GetBody().SetSessionKey(nil)
	}},
	{name: "body/session key/empty", err: "missing session public key", corrupt: func(st *apisession.Token) {
		st.GetBody().SetSessionKey([]byte{})
	}},
	{name: "body/context/nil", err: "missing session context", corrupt: func(st *apisession.Token) {
		st.GetBody().SetContext(nil)
	}},
	{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32", corrupt: func(st *apisession.Token) {
		st.GetSignature().SetScheme(math.MaxInt32 + 1)
	}},
}

type invalidBinTokenTestcase struct {
	name, err string
	b         []byte
}

var invalidBinTokenCommonTestcases = []invalidBinTokenTestcase{
	{name: "body/ID/undersize", err: "invalid session ID: invalid UUID (got 15 bytes)",
		b: []byte{10, 17, 10, 15, 188, 255, 42, 107, 236, 249, 78, 152, 169, 7, 2, 87, 36, 139, 31}},
	{name: "body/ID/oversize", err: "invalid session ID: invalid UUID (got 17 bytes)",
		b: []byte{10, 19, 10, 17, 109, 141, 40, 16, 21, 245, 76, 128, 150, 236, 154, 53, 157, 172, 12, 195, 1}},
	{name: "body/ID/wrong UUID version", err: "invalid session UUID version 3",
		b: []byte{10, 18, 10, 16, 97, 47, 243, 131, 222, 201, 48, 64, 135, 195, 177, 240, 107, 12, 2, 42}},
	{name: "body/issuer/value/empty", err: "invalid session issuer: invalid length 0, expected 25",
		b: []byte{10, 2, 18, 0}},
	{name: "body/issuer/value/undersize", err: "invalid session issuer: invalid length 24, expected 25",
		b: []byte{10, 28, 18, 26, 10, 24, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17}},
	{name: "body/issuer/value/oversize", err: "invalid session issuer: invalid length 26, expected 25",
		b: []byte{10, 30, 18, 28, 10, 26, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17, 92, 1}},
	{name: "body/issuer/value/wrong prefix", err: "invalid session issuer: invalid prefix byte 0x42, expected 0x35",
		b: []byte{10, 29, 18, 27, 10, 25, 66, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17, 92}},
	{name: "body/issuer/value/checksum mismatch", err: "invalid session issuer: checksum mismatch",
		b: []byte{10, 29, 18, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17, 93}},
	{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32",
		b: []byte{18, 11, 24, 128, 128, 128, 128, 248, 255, 255, 255, 255, 1}},
}

var invalidSignedTokenCommonTestcases = []invalidBinTokenTestcase{
	{name: "ID/undersize", err: "invalid session ID: invalid UUID (got 15 bytes)",
		b: []byte{10, 15, 188, 255, 42, 107, 236, 249, 78, 152, 169, 7, 2, 87, 36, 139, 31}},
	{name: "ID/oversize", err: "invalid session ID: invalid UUID (got 17 bytes)",
		b: []byte{10, 17, 109, 141, 40, 16, 21, 245, 76, 128, 150, 236, 154, 53, 157, 172, 12, 195, 1}},
	{name: "ID/wrong UUID version", err: "invalid session UUID version 3",
		b: []byte{10, 16, 97, 47, 243, 131, 222, 201, 48, 64, 135, 195, 177, 240, 107, 12, 2, 42}},
	{name: "issuer/value/empty", err: "invalid session issuer: invalid length 0, expected 25",
		b: []byte{18, 0}},
	{name: "issuer/value/undersize", err: "invalid session issuer: invalid length 24, expected 25",
		b: []byte{18, 26, 10, 24, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17}},
	{name: "issuer/value/oversize", err: "invalid session issuer: invalid length 26, expected 25",
		b: []byte{18, 28, 10, 26, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17, 92, 1}},
	{name: "issuer/value/wrong prefix", err: "invalid session issuer: invalid prefix byte 0x42, expected 0x35",
		b: []byte{18, 27, 10, 25, 66, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17, 92}},
	{name: "issuer/value/checksum mismatch", err: "invalid session issuer: checksum mismatch",
		b: []byte{18, 27, 10, 25, 53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57, 160, 82, 80, 41, 203,
			20, 254, 30, 138, 195, 17, 93}},
}

type invalidJSONTokenTestcase struct {
	name, err string
	j         string
}

var invalidJSONTokenCommonTestcases = []invalidJSONTokenTestcase{
	{name: "body/ID/undersize", err: "invalid session ID: invalid UUID (got 15 bytes)", j: `
{"body":{"id":"YxhvRhasSBSLu69iCv/n"}}
`},
	{name: "body/ID/oversize", err: "invalid session ID: invalid UUID (got 17 bytes)", j: `
{"body":{"id":"YxhvRhasSBSLu69iCv/nvAE="}}
`},
	{name: "body/ID/wrong UUID version", err: "invalid session UUID version 3", j: `
{"body":{"id":"YxhvRhasMBSLu69iCv/nvA=="}}
`},
	{name: "body/issuer/value/empty", err: "invalid session issuer: invalid length 0, expected 25", j: `
{"body":{"ownerID":{}}}
`},
	{name: "body/issuer/value/undersize", err: "invalid session issuer: invalid length 24, expected 25", j: `
{"body":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMR"}}}
`},
	{name: "body/issuer/value/oversize", err: "invalid session issuer: invalid length 26, expected 25", j: `
{"body":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXAE="}}}
`},
	{name: "body/issuer/value/wrong prefix", err: "invalid session issuer: invalid prefix byte 0x42, expected 0x35", j: `
{"body":{"ownerID":{"value":"QjMFpm8dFGXApRynOaBSUCnLFP4eisMRXA=="}}}
`},
	{name: "body/issuer/value/checksum mismatch", err: "invalid session issuer: checksum mismatch", j: `
{"body":{"ownerID":{"value":"NTMFpm8dFGXApRynOaBSUCnLFP4eisMRXQ=="}}}
`},
	{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32", j: `
{"signature":{"scheme":-2147483648}}
`},
}

func testLifetimeClaim[T session.Container | session.Object](t testing.TB, get func(T) uint64, set func(*T, uint64)) {
	var x T
	require.Zero(t, get(x))
	set(&x, 12094032)
	require.EqualValues(t, 12094032, get(x))
	set(&x, 5469830342)
	require.EqualValues(t, 5469830342, get(x))
}

func testValidAt[T interface {
	*session.Container | *session.Object
	SetExp(uint64)
	SetIat(uint64)
	SetNbf(uint64)
	ValidAt(uint64) bool
	InvalidAt(uint64) bool
}](t testing.TB, x T) {
	require.True(t, x.ValidAt(0))
	require.False(t, x.InvalidAt(0))

	const iat = 13
	const nbf = iat + 1
	const exp = nbf + 1

	x.SetIat(iat)
	x.SetNbf(nbf)
	x.SetExp(exp)

	require.False(t, x.ValidAt(iat-1))
	require.True(t, x.InvalidAt(iat-1))
	require.False(t, x.ValidAt(iat))
	require.True(t, x.InvalidAt(iat))
	require.True(t, x.ValidAt(nbf))
	require.False(t, x.InvalidAt(nbf))
	require.True(t, x.ValidAt(exp))
	require.False(t, x.InvalidAt(exp))
	require.False(t, x.ValidAt(exp+1))
	require.True(t, x.InvalidAt(exp+1))
}

func TestInvalidAt(t *testing.T) {
	testValidAt(t, new(session.Container))
	testValidAt(t, new(session.Object))
}

func testSetAuthKey[T session.Container | session.Object](t testing.TB, set func(*T, neofscrypto.PublicKey), assert func(T, neofscrypto.PublicKey) bool) {
	k1 := neofscryptotest.Signer().Public()
	k2 := neofscryptotest.Signer().Public()
	var x T
	require.False(t, assert(x, k1))
	require.False(t, assert(x, k2))

	set(&x, k1)
	require.True(t, assert(x, k1))
	require.False(t, assert(x, k2))

	set(&x, k2)
	require.False(t, assert(x, k1))
	require.True(t, assert(x, k2))
}

func testTokenIssuer[T interface {
	session.Container | session.Object
	Issuer() user.ID
}, PTR interface {
	*T
	SetIssuer(user.ID)
}](t testing.TB, _ T) {
	var x T
	require.Zero(t, x.Issuer())
	p := PTR(&x)
	p.SetIssuer(anyValidUserID)
	require.Equal(t, anyValidUserID, x.Issuer())
	otherIssuer := usertest.OtherID(anyValidUserID)
	p.SetIssuer(otherIssuer)
	require.Equal(t, otherIssuer, x.Issuer())
}

func testTokenID[T interface {
	session.Container | session.Object
	ID() uuid.UUID
}, PTR interface {
	*T
	SetID(uuid.UUID)
}](t testing.TB, _ T) {
	var x T
	require.Zero(t, x.ID())
	p := PTR(&x)
	uid := uuid.New()
	p.SetID(uid)
	require.Equal(t, uid, x.ID())
	otherID := uid
	otherID[0]++
	p.SetID(otherID)
	require.Equal(t, otherID, x.ID())
}

func testSetSignatureECDSA[T interface {
	*session.Container | *session.Object
	SetSignature(neofscrypto.Signer) error
	Signature() (neofscrypto.Signature, bool)
}](t testing.TB, ecdsaPriv ecdsa.PrivateKey, token T, signed []byte, rfc6979Sig []byte) {
	/* non-deterministic schemes */
	assertECDSACommon := func(signer neofscrypto.Signer) []byte {
		scheme := signer.Scheme()
		require.NoError(t, token.SetSignature(signer), scheme)
		sig, ok := token.Signature()
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
	sig := assertECDSACommon(user.NewAutoIDSigner(ecdsaPriv))
	require.Len(t, sig, 65)
	require.EqualValues(t, 0x4, sig[0])
	h512 := sha512.Sum512(signed)
	r, s := new(big.Int).SetBytes(sig[1:33]), new(big.Int).SetBytes(sig[33:])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h512[:], r, s))

	// WalletConnect
	sig = assertECDSACommon(user.NewSigner(neofsecdsa.SignerWalletConnect(ecdsaPriv), user.NewFromECDSAPublicKey(ecdsaPriv.PublicKey)))
	require.Len(t, sig, 80)
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(signed)))
	base64.StdEncoding.Encode(b64, signed)
	payloadLen := 2*16 + len(b64)
	b := make([]byte, 4+getVarIntSize(payloadLen)+payloadLen+2)
	n := copy(b, []byte{0x01, 0x00, 0x01, 0xf0})
	n += putVarUint(b[n:], uint64(payloadLen))
	n += hex.Encode(b[n:], sig[64:])
	n += copy(b[n:], b64)
	copy(b[n:], []byte{0x00, 0x00})
	h256 := sha256.Sum256(b)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))

	/* deterministic schemes */
	sig = assertECDSACommon(user.NewAutoIDSignerRFC6979(ecdsaPriv))
	require.Equal(t, rfc6979Sig, sig)
	h256 = sha256.Sum256(signed)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))
}

// copy-paste from crypto/ecdsa package.
func getVarIntSize(value int) int {
	var size uintptr

	if value < 0xFD {
		size = 1 // unit8
	} else if value <= 0xFFFF {
		size = 3 // byte + uint16
	} else {
		size = 5 // byte + uint32
	}
	return int(size)
}

func putVarUint(data []byte, val uint64) int {
	if val < 0xfd {
		data[0] = byte(val)
		return 1
	}
	if val <= 0xFFFF {
		data[0] = byte(0xfd)
		binary.LittleEndian.PutUint16(data[1:], uint16(val))
		return 3
	}

	data[0] = byte(0xfe)
	binary.LittleEndian.PutUint32(data[1:], uint32(val))
	return 5
}

func testSignCDSA[T interface {
	*session.Container | *session.Object
	Signature() (neofscrypto.Signature, bool)
	Sign(user.Signer) error
	Issuer() user.ID
}](t testing.TB, ecdsaPriv ecdsa.PrivateKey, usr user.ID, token T, signed []byte, rfc6979Sig []byte) {
	/* non-deterministic schemes */
	assertECDSACommon := func(signer neofscrypto.Signer) []byte {
		scheme := signer.Scheme()
		require.NoError(t, token.Sign(user.NewSigner(signer, usr)), scheme)
		require.Equal(t, usr, token.Issuer())
		sig, ok := token.Signature()
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
	h512 := sha512.Sum512(signed)
	r, s := new(big.Int).SetBytes(sig[1:33]), new(big.Int).SetBytes(sig[33:])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h512[:], r, s))

	// WalletConnect
	sig = assertECDSACommon(neofsecdsa.SignerWalletConnect(ecdsaPriv))
	require.Len(t, sig, 80)
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(signed)))
	base64.StdEncoding.Encode(b64, signed)
	payloadLen := 2*16 + len(b64)
	b := make([]byte, 4+getVarIntSize(payloadLen)+payloadLen+2)
	n := copy(b, []byte{0x01, 0x00, 0x01, 0xf0})
	n += putVarUint(b[n:], uint64(payloadLen))
	n += hex.Encode(b[n:], sig[64:])
	n += copy(b[n:], b64)
	copy(b[n:], []byte{0x00, 0x00})
	h256 := sha256.Sum256(b)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))

	/* deterministic schemes */
	sig = assertECDSACommon(neofsecdsa.SignerRFC6979(ecdsaPriv))
	require.Equal(t, rfc6979Sig, sig)
	h256 = sha256.Sum256(signed)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))
}
