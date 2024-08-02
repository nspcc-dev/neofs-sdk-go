package session_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"
	"math/rand"
	"strconv"
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
	anyValidSignatureBytes  = []byte("any_signature") // valid structurally, not logically
	anyValidSignature       = neofscrypto.NewSignatureFromRawKey(anyValidSignatureScheme, anyValidIssuerPublicKeyBytes, anyValidSignatureBytes)
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
	corrupt   func(*apisession.SessionToken)
}

var invalidProtoTokenCommonTestcases = []invalidProtoTokenTestcase{
	{name: "missing body", err: "missing token body", corrupt: func(st *apisession.SessionToken) {
		st.SetBody(nil)
	}},
	{name: "missing body signature", err: "missing body signature", corrupt: func(st *apisession.SessionToken) {
		st.SetSignature(nil)
	}},
	{name: "body/ID/nil", err: "missing session ID", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetID(nil)
	}},
	{name: "body/ID/empty", err: "missing session ID", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetID([]byte{})
	}},
	{name: "body/ID/undersize", err: "invalid session ID: invalid UUID (got 15 bytes)", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetID(make([]byte, 15))
	}},
	{name: "body/ID/wrong UUID version", err: "invalid session UUID version 3", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().GetID()[6] = 3 << 4
	}},
	{name: "body/ID/oversize", err: "invalid session ID: invalid UUID (got 17 bytes)", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetID(make([]byte, 17))
	}},
	{name: "body/issuer", err: "missing session issuer", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetOwnerID(nil)
	}},
	{name: "body/issuer/value/nil", err: "invalid session issuer: invalid length 0, expected 25", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().GetOwnerID().SetValue(nil)
	}},
	{name: "body/issuer/value/empty", err: "invalid session issuer: invalid length 0, expected 25", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().GetOwnerID().SetValue([]byte{})
	}},
	{name: "body/issuer/value/undersize", err: "invalid session issuer: invalid length 24, expected 25", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().GetOwnerID().SetValue(make([]byte, 24))
	}},
	{name: "body/issuer/value/oversize", err: "invalid session issuer: invalid length 26, expected 25", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().GetOwnerID().SetValue(make([]byte, 26))
	}},
	{name: "body/issuer/value/wrong prefix", err: "invalid session issuer: invalid prefix byte 0x42, expected 0x35", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().GetOwnerID().GetValue()[0] = 0x42
	}},
	{name: "body/issuer/value/checksum mismatch", err: "invalid session issuer: checksum mismatch", corrupt: func(st *apisession.SessionToken) {
		v := st.GetBody().GetOwnerID().GetValue()
		v[len(v)-1]++
	}},
	{name: "body/lifetime", err: "missing token lifetime", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetLifetime(nil)
	}},
	{name: "body/session key/nil", err: "missing session public key", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetSessionKey(nil)
	}},
	{name: "body/session key/empty", err: "missing session public key", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetSessionKey([]byte{})
	}},
	{name: "body/context/nil", err: "missing session context", corrupt: func(st *apisession.SessionToken) {
		st.GetBody().SetContext(nil)
	}},
	{name: "signature/invalid scheme", err: "invalid body signature: scheme 2147483648 overflows int32", corrupt: func(st *apisession.SessionToken) {
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

type lifetime interface {
	Exp() uint64
	SetExp(uint64)
	Iat() uint64
	SetIat(uint64)
	Nbf() uint64
	SetNbf(uint64)
	InvalidAt(uint64) bool
}

func testInvalidAt(t testing.TB, x lifetime) {
	require.False(t, session.InvalidAt(x, 0))
	require.False(t, x.InvalidAt(0))

	nbf := rand.Uint64()
	if nbf == math.MaxUint64 {
		nbf--
	}

	iat := nbf
	exp := iat + 1

	x.SetNbf(nbf)
	x.SetIat(iat)
	x.SetExp(exp)

	require.True(t, session.InvalidAt(x, nbf-1))
	require.True(t, x.InvalidAt(nbf-1))
	require.True(t, session.InvalidAt(x, iat-1))
	require.True(t, x.InvalidAt(iat-1))
	require.False(t, session.InvalidAt(x, iat))
	require.False(t, x.InvalidAt(iat))
	require.False(t, session.InvalidAt(x, exp))
	require.False(t, x.InvalidAt(exp))
	require.True(t, session.InvalidAt(x, exp+1))
	require.True(t, x.InvalidAt(exp+1))
}

func testExpiredAt(t testing.TB, x lifetime) {
	require.False(t, session.ExpiredAt(x, 0))
	const curEpoch = 42
	require.True(t, session.ExpiredAt(x, curEpoch))
	x.SetExp(curEpoch)
	require.False(t, session.ExpiredAt(x, curEpoch-1))
	require.False(t, session.ExpiredAt(x, curEpoch))
	require.True(t, session.ExpiredAt(x, curEpoch+1))
}

func testLifetimeClaim[T session.Container | session.Object](t testing.TB, get func(T) uint64, set func(*T, uint64)) {
	var x T
	require.Zero(t, get(x))
	set(&x, 12094032)
	require.EqualValues(t, 12094032, get(x))
	set(&x, 5469830342)
	require.EqualValues(t, 5469830342, get(x))
}

func testAuthPublicKeyField[T session.Container | session.Object](t testing.TB, get func(T) []byte, set func(*T, []byte)) {
	var x T
	require.Zero(t, get(x))
	set(&x, []byte("any"))
	require.EqualValues(t, "any", get(x))
	set(&x, []byte("other"))
	require.EqualValues(t, "other", get(x))
}

type tokenWithAuthPublicKey interface {
	SetAuthPublicKey([]byte)
	AuthPublicKey() []byte
}

func testSetAuthPublicKey(t testing.TB, x tokenWithAuthPublicKey) {
	require.Zero(t, x.AuthPublicKey())
	usr := usertest.User()
	session.SetAuthPublicKey(x, usr.Public())
	require.Equal(t, usr.PublicKeyBytes, x.AuthPublicKey())
}

func testAssertAuthPublicKey(t testing.TB, x tokenWithAuthPublicKey) {
	usr := usertest.User()
	pub := usr.Public()
	require.False(t, session.AssertAuthPublicKey(x, pub))
	x.SetAuthPublicKey(usr.PublicKeyBytes)
	require.True(t, session.AssertAuthPublicKey(x, pub))
}

func TestInvalidAt(t *testing.T) {
	testInvalidAt(t, new(session.Container))
	testInvalidAt(t, new(session.Object))
}

func TestExpiredAt(t *testing.T) {
	testExpiredAt(t, new(session.Container))
	testExpiredAt(t, new(session.Object))
}

func TestSetAuthPublicKey(t *testing.T) {
	testSetAuthPublicKey(t, new(session.Container))
	testSetAuthPublicKey(t, new(session.Object))
}

func TestAssertAuthPublicKey(t *testing.T) {
	testAssertAuthPublicKey(t, new(session.Container))
	testAssertAuthPublicKey(t, new(session.Object))
}

func testAttachSignature[T interface {
	session.Container | session.Object
	Signature() neofscrypto.Signature
}, PTR interface {
	*T
	AttachSignature(neofscrypto.Signature)
}](t testing.TB, _ T) {
	var x T
	require.Negative(t, x.Signature().Scheme())
	p := PTR(&x)
	p.AttachSignature(anyValidSignature)
	require.Equal(t, anyValidSignature, x.Signature())
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

func testSignECDSA[T interface {
	*session.Container | *session.Object
	SignedData() []byte
	Signature() neofscrypto.Signature
	AttachSignature(neofscrypto.Signature)
}](t testing.TB, ecdsaPriv ecdsa.PrivateKey, token T, signed []byte, rfc6979Sig []byte) {
	/* non-deterministic schemes */
	assertECDSACommon := func(signer neofscrypto.Signer) []byte {
		scheme := signer.Scheme()
		require.NoError(t, session.Sign(token, signer), scheme)
		sig := token.Signature()
		require.True(t, sig.Scheme() >= 0)
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

	/* determenistic schemes */
	sig = assertECDSACommon(neofsecdsa.SignerRFC6979(ecdsaPriv))
	require.Equal(t, rfc6979Sig, sig)
	h256 = sha256.Sum256(signed)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))
}

// TODO: copy-paste from crypto/ecdsa package, share
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

func TestSign(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		require.Error(t, session.Sign(new(session.Container), neofscryptotest.FailSigner(usertest.User())))
		require.Error(t, session.Sign(new(session.Object), neofscryptotest.FailSigner(usertest.User())))
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
			testSignECDSA(t, ecdsaPriv, &c, validSignedContainerTokens[i], rfc6979Sig)
		})
	}
}

func TestHasValidSignature(t *testing.T) {
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
			require.True(t, session.HasValidSignature(c), [2]int{i, j})
			for k := range pub {
				pubCp := bytes.Clone(pub)
				pubCp[k]++
				sig.SetPublicKeyBytes(pubCp)
				c.AttachSignature(sig)
				require.False(t, session.HasValidSignature(c), [2]int{i, j})
			}
			for k := range sigBytes {
				sigBytesCp := bytes.Clone(sigBytes)
				sigBytesCp[k]++
				sig.SetValue(sigBytesCp)
				c.AttachSignature(sig)
				require.False(t, session.HasValidSignature(c), [2]int{i, j})
			}
		}
	}
}

func testIssueECDSA[T interface {
	*session.Container | *session.Object
	SignedData() []byte
	Signature() neofscrypto.Signature
	AttachSignature(neofscrypto.Signature)
	Issuer() user.ID
	SetIssuer(user.ID)
}](t testing.TB, ecdsaPriv ecdsa.PrivateKey, usr user.ID, token T, signed []byte, rfc6979Sig []byte) {
	/* non-deterministic schemes */
	assertECDSACommon := func(signer neofscrypto.Signer) []byte {
		scheme := signer.Scheme()
		require.NoError(t, session.Issue(token, user.NewSigner(signer, usr)), scheme)
		require.Equal(t, usr, token.Issuer())
		sig := token.Signature()
		require.True(t, sig.Scheme() >= 0)
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

	/* determenistic schemes */
	sig = assertECDSACommon(neofsecdsa.SignerRFC6979(ecdsaPriv))
	require.Equal(t, rfc6979Sig, sig)
	h256 = sha256.Sum256(signed)
	r, s = new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:][:32])
	require.True(t, ecdsa.Verify(&ecdsaPriv.PublicKey, h256[:], r, s))
}

func TestIssue(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		var c session.Container
		require.Error(t, session.Issue(&c, usertest.FailSigner(usertest.User())))
		require.ErrorIs(t, session.Issue(&c, user.NewSigner(neofscryptotest.Signer(), user.ID{})),
			user.ErrZeroID)
	})

	ecdsaPriv := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(),
			X: new(big.Int).SetBytes([]byte{171, 217, 50, 167, 24, 120, 67, 137, 0, 64, 134, 74, 236, 242, 73, 4, 28, 177,
				215, 115, 187, 37, 16, 81, 230, 188, 92, 251, 91, 13, 13, 99}),
			Y: new(big.Int).SetBytes([]byte{231, 32, 164, 250, 82, 22, 203, 118, 78, 93, 171, 155, 127, 14, 50, 59, 164,
				134, 138, 73, 198, 174, 245, 20, 67, 156, 41, 172, 66, 75, 169, 90}),
		},
		D: new(big.Int).SetBytes([]byte{210, 88, 65, 214, 139, 200, 102, 205, 2, 133, 138, 60, 243, 34, 112, 162, 179,
			118, 144, 83, 184, 240, 73, 116, 22, 204, 212, 88, 130, 215, 191, 55}),
	}

	var c session.Container
	for i, rfc6979Sig := range [][]byte{
		{52, 207, 3, 101, 183, 3, 244, 108, 214, 63, 93, 21, 42, 80, 231, 143, 194, 134, 36, 13, 164, 127, 240, 90,
			85, 149, 10, 243, 204, 60, 215, 178, 37, 64, 86, 243, 103, 206, 0, 144, 27, 155, 92, 69, 169, 232, 135,
			31, 243, 104, 95, 38, 250, 216, 51, 97, 82, 6, 44, 22, 139, 206, 39, 174},
		{231, 138, 236, 28, 121, 35, 76, 56, 230, 40, 18, 17, 162, 79, 244, 199, 134, 115, 181, 128, 93, 31, 146, 170,
			107, 230, 244, 94, 116, 86, 139, 35, 4, 108, 216, 60, 122, 221, 133, 157, 87, 131, 188, 37, 127, 232,
			12, 203, 108, 128, 72, 75, 121, 10, 35, 134, 1, 212, 112, 217, 98, 25, 218, 217},
	} {
		validContainerTokens[i].CopyTo(&c)
		t.Run("container#"+strconv.Itoa(i), func(t *testing.T) {
			testIssueECDSA(t, ecdsaPriv, anyValidUserID, &c, validSignedContainerTokens[i], rfc6979Sig)
		})
	}
}
