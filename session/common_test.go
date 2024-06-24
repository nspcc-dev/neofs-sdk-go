package session_test

import (
	"bytes"
	"encoding/json"
	"math"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type token interface {
	Marshal() []byte
	SignedData() []byte
	json.Marshaler
	WriteToV2(*apisession.SessionToken)

	Issuer() user.ID
	IssuerPublicKeyBytes() []byte
	ID() uuid.UUID
	AssertAuthKey(neofscrypto.PublicKey) bool
	InvalidAt(uint64) bool

	VerifySignature() bool
}

type tokenptr interface {
	token
	Unmarshal([]byte) error
	UnmarshalSignedData([]byte) error
	json.Unmarshaler
	ReadFromV2(*apisession.SessionToken) error

	SetIssuer(user.ID)
	SetID(uuid.UUID)
	SetAuthKey(neofscrypto.PublicKey)
	SetExp(uint64)
	SetNbf(uint64)
	SetIat(uint64)

	Sign(user.Signer) error
	SetSignature(neofscrypto.Signer) error
}

func setRequiredTokenAPIFields(tok tokenptr) {
	usr, _ := usertest.TwoUsers()
	tok.SetIssuer(usertest.ID())
	tok.SetID(uuid.New())
	tok.SetAuthKey(usr.Public())
	tok.SetExp(1)
}

type invalidAPITestCase struct {
	name, err string
	corrupt   func(*apisession.SessionToken)
}

func testDecoding[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, full func() T, customCases []invalidAPITestCase) {
	t.Run("missing fields", func(t *testing.T) {
		for _, testCase := range []invalidAPITestCase{
			{name: "body", err: "missing token body", corrupt: func(st *apisession.SessionToken) {
				st.Body = nil
			}},
			{name: "body/ID/nil", err: "missing session ID", corrupt: func(st *apisession.SessionToken) {
				st.Body.Id = nil
			}},
			{name: "body/ID/empty", err: "missing session ID", corrupt: func(st *apisession.SessionToken) {
				st.Body.Id = []byte{}
			}},
			{name: "body/issuer", err: "missing session issuer", corrupt: func(st *apisession.SessionToken) {
				st.Body.OwnerId = nil
			}},
			{name: "body/lifetime", err: "missing token lifetime", corrupt: func(st *apisession.SessionToken) {
				st.Body.Lifetime = nil
			}},
			{name: "body/session key/nil", err: "missing session public key", corrupt: func(st *apisession.SessionToken) {
				st.Body.SessionKey = nil
			}},
			{name: "body/session key/empty", err: "missing session public key", corrupt: func(st *apisession.SessionToken) {
				st.Body.SessionKey = []byte{}
			}},
			{name: "body/context/nil", err: "missing session context", corrupt: func(st *apisession.SessionToken) {
				st.Body.Context = nil
			}},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				src := full()
				var dst PTR = new(T)
				var m apisession.SessionToken

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)

				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.NoError(t, dst.Unmarshal(b))

				j, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.NoError(t, dst.UnmarshalJSON(j))
			})
		}
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, testCase := range append(customCases, []invalidAPITestCase{
			{name: "signature/key/nil", err: "invalid body signature: missing public key", corrupt: func(st *apisession.SessionToken) {
				st.Signature.Key = nil
			}},
			{name: "signature/key/empty", err: "invalid body signature: missing public key", corrupt: func(st *apisession.SessionToken) {
				st.Signature.Key = []byte{}
			}},
			{name: "signature/signature/nil", err: "invalid body signature: missing signature", corrupt: func(st *apisession.SessionToken) {
				st.Signature.Sign = nil
			}},
			{name: "signature/signature/empty", err: "invalid body signature: missing signature", corrupt: func(st *apisession.SessionToken) {
				st.Signature.Sign = []byte{}
			}},
			{name: "signature/unsupported scheme", err: "invalid body signature: unsupported scheme 2147483647", corrupt: func(st *apisession.SessionToken) {
				st.Signature.Scheme = math.MaxInt32
			}},
			{name: "body/ID/wrong length", err: "invalid session ID: invalid UUID (got 15 bytes)", corrupt: func(st *apisession.SessionToken) {
				st.Body.Id = make([]byte, 15)
			}},
			{name: "body/ID/wrong prefix", err: "invalid session UUID version 3", corrupt: func(st *apisession.SessionToken) {
				st.Body.Id[6] = 3 << 4
			}},
			{name: "body/issuer/value/nil", err: "invalid session issuer: missing value field", corrupt: func(st *apisession.SessionToken) {
				st.Body.OwnerId.Value = nil
			}},
			{name: "body/issuer/value/empty", err: "invalid session issuer: missing value field", corrupt: func(st *apisession.SessionToken) {
				st.Body.OwnerId.Value = []byte{}
			}},
			{name: "body/issuer/value/wrong length", err: "invalid session issuer: invalid value length 24", corrupt: func(st *apisession.SessionToken) {
				st.Body.OwnerId.Value = make([]byte, 24)
			}},
			{name: "body/issuer/value/wrong prefix", err: "invalid session issuer: invalid prefix byte 0x42, expected 0x35", corrupt: func(st *apisession.SessionToken) {
				st.Body.OwnerId.Value[0] = 0x42
			}},
			{name: "body/issuer/value/checksum mismatch", err: "invalid session issuer: value checksum mismatch", corrupt: func(st *apisession.SessionToken) {
				st.Body.OwnerId.Value[len(st.Body.OwnerId.Value)-1]++
			}},
		}...) {
			t.Run(testCase.name, func(t *testing.T) {
				src := full()
				var dst PTR = new(T)
				var m apisession.SessionToken

				src.WriteToV2(&m)
				testCase.corrupt(&m)
				require.ErrorContains(t, dst.ReadFromV2(&m), testCase.err)

				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, dst.Unmarshal(b), testCase.err)

				j, err := protojson.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, dst.UnmarshalJSON(j), testCase.err)
			})
		}
	})
}

func testCopyTo[T interface {
	token
	CopyTo(*T)
}](t *testing.T, full T) {
	shallow := full
	var deep T
	full.CopyTo(&deep)
	require.Equal(t, full, deep)
	require.Equal(t, full, shallow)

	originIssuerKey := bytes.Clone(full.IssuerPublicKeyBytes())
	issuerKey := full.IssuerPublicKeyBytes()
	issuerKey[0]++
	require.Equal(t, issuerKey, full.IssuerPublicKeyBytes())
	require.Equal(t, issuerKey, shallow.IssuerPublicKeyBytes())
	require.Equal(t, originIssuerKey, deep.IssuerPublicKeyBytes())
}

func testSignedData[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, unsigned T) {
	t.Run("invalid binary", func(t *testing.T) {
		var c PTR = new(T)
		msg := []byte("definitely_not_protobuf")
		err := c.UnmarshalSignedData(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})

	var x2 PTR = new(T)
	var m apisession.SessionToken_Body

	b := unsigned.SignedData()
	require.NoError(t, x2.UnmarshalSignedData(b))
	require.Equal(t, unsigned, *x2)

	require.NoError(t, proto.Unmarshal(b, &m))
	b, err := proto.Marshal(&m)
	require.NoError(t, err)
	var x3 PTR = new(T)
	require.NoError(t, x3.UnmarshalSignedData(b))
	require.Equal(t, unsigned, *x3)
}

func testAuthKey[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, setRequiredAPIFields func(PTR)) {
	var c PTR = new(T)
	usr, otherUsr := usertest.TwoUsers()

	require.False(t, c.AssertAuthKey(usr.Public()))
	require.False(t, c.AssertAuthKey(otherUsr.Public()))

	c.SetAuthKey(usr.Public())
	require.True(t, c.AssertAuthKey(usr.Public()))
	require.False(t, c.AssertAuthKey(otherUsr.Public()))

	c.SetAuthKey(otherUsr.Public())
	require.False(t, c.AssertAuthKey(usr.Public()))
	require.True(t, c.AssertAuthKey(otherUsr.Public()))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			dst.SetAuthKey(usr.Public())
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.AssertAuthKey(usr.Public()))
			require.False(t, dst.AssertAuthKey(otherUsr.Public()))

			dst.SetAuthKey(otherUsr.Public())
			src.SetAuthKey(usr.Public())
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.AssertAuthKey(usr.Public()))
			require.False(t, dst.AssertAuthKey(otherUsr.Public()))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredAPIFields(src)

			dst.SetAuthKey(otherUsr.Public())
			src.SetAuthKey(usr.Public())
			src.WriteToV2(&msg)
			require.Equal(t, usr.PublicKeyBytes, msg.Body.SessionKey)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.AssertAuthKey(usr.Public()))
			require.False(t, dst.AssertAuthKey(otherUsr.Public()))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			dst.SetAuthKey(usr.Public())
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.AssertAuthKey(usr.Public()))
			require.False(t, dst.AssertAuthKey(otherUsr.Public()))

			dst.SetAuthKey(otherUsr.Public())
			src.SetAuthKey(usr.Public())
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.AssertAuthKey(usr.Public()))
			require.False(t, dst.AssertAuthKey(otherUsr.Public()))
		})
	})
}

func testSign[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, setRequiredAPIFields func(PTR)) {
	var c PTR = new(T)

	require.False(t, c.VerifySignature())

	usr, otherUsr := usertest.TwoUsers()
	usrSigner := user.NewAutoIDSignerRFC6979(usr.PrivateKey)
	otherUsrSigner := user.NewAutoIDSignerRFC6979(otherUsr.PrivateKey)

	require.Error(t, c.Sign(usertest.FailSigner(usr)))
	require.False(t, c.VerifySignature())
	require.Error(t, c.Sign(usertest.FailSigner(otherUsr)))
	require.False(t, c.VerifySignature())

	require.NoError(t, c.Sign(usrSigner))
	require.True(t, c.VerifySignature())
	require.Equal(t, usr.ID, c.Issuer())
	require.Equal(t, usr.PublicKeyBytes, c.IssuerPublicKeyBytes())

	require.NoError(t, c.Sign(otherUsrSigner))
	require.True(t, c.VerifySignature())
	require.Equal(t, otherUsr.ID, c.Issuer())
	require.Equal(t, otherUsr.PublicKeyBytes, c.IssuerPublicKeyBytes())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			require.NoError(t, dst.Sign(usrSigner))
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Zero(t, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.Sign(otherUsrSigner))
			require.NoError(t, src.Sign(usrSigner))
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Equal(t, usr.ID, dst.Issuer())
			require.Equal(t, usr.PublicKeyBytes, dst.IssuerPublicKeyBytes())

			require.NoError(t, src.Sign(otherUsrSigner))
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Equal(t, otherUsr.ID, dst.Issuer())
			require.Equal(t, otherUsr.PublicKeyBytes, dst.IssuerPublicKeyBytes())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredAPIFields(src)

			require.NoError(t, dst.Sign(usrSigner))
			src.WriteToV2(&msg)
			require.Zero(t, msg.Signature)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.False(t, dst.VerifySignature())
			require.Equal(t, src.Issuer(), dst.Issuer())
			require.Zero(t, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.Sign(otherUsrSigner))
			require.NoError(t, src.Sign(usrSigner))
			src.WriteToV2(&msg)
			require.Equal(t, usr.PublicKeyBytes, msg.Signature.Key)
			require.NotEmpty(t, msg.Signature.Sign)
			require.EqualValues(t, usrSigner.Scheme(), msg.Signature.Scheme)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.VerifySignature())
			require.Equal(t, usr.ID, dst.Issuer())
			require.Equal(t, usr.PublicKeyBytes, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.Sign(usrSigner))
			require.NoError(t, src.Sign(otherUsrSigner))
			src.WriteToV2(&msg)
			require.Equal(t, otherUsr.PublicKeyBytes, msg.Signature.Key)
			require.NotEmpty(t, msg.Signature.Sign)
			require.EqualValues(t, otherUsrSigner.Scheme(), msg.Signature.Scheme)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.VerifySignature())
			require.Equal(t, otherUsr.ID, dst.Issuer())
			require.Equal(t, otherUsr.PublicKeyBytes, dst.IssuerPublicKeyBytes())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			require.NoError(t, dst.Sign(usrSigner))
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Zero(t, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.Sign(otherUsrSigner))
			require.NoError(t, src.Sign(usrSigner))
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Equal(t, usr.ID, dst.Issuer())
			require.Equal(t, usr.PublicKeyBytes, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.Sign(usrSigner))
			require.NoError(t, src.Sign(otherUsrSigner))
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Equal(t, otherUsr.ID, dst.Issuer())
			require.Equal(t, otherUsr.PublicKeyBytes, dst.IssuerPublicKeyBytes())
		})
	})
}

func testSetSignature[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, setRequiredAPIFields func(PTR)) {
	var c PTR = new(T)

	require.False(t, c.VerifySignature())

	usr, otherUsr := usertest.TwoUsers()
	usrSigner := neofsecdsa.SignerRFC6979(usr.PrivateKey)
	otherUsrSigner := neofsecdsa.SignerRFC6979(otherUsr.PrivateKey)

	require.Error(t, c.SetSignature(usertest.FailSigner(usr)))
	require.False(t, c.VerifySignature())
	require.Error(t, c.SetSignature(usertest.FailSigner(otherUsr)))
	require.False(t, c.VerifySignature())

	require.NoError(t, c.SetSignature(usrSigner))
	require.True(t, c.VerifySignature())
	require.Zero(t, c.Issuer())
	require.Equal(t, usr.PublicKeyBytes, c.IssuerPublicKeyBytes())

	require.NoError(t, c.SetSignature(otherUsrSigner))
	require.True(t, c.VerifySignature())
	require.Zero(t, c.Issuer())
	require.Equal(t, otherUsr.PublicKeyBytes, c.IssuerPublicKeyBytes())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			require.NoError(t, dst.SetSignature(usrSigner))
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.False(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Zero(t, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.SetSignature(otherUsrSigner))
			require.NoError(t, src.SetSignature(usrSigner))
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Equal(t, usr.PublicKeyBytes, dst.IssuerPublicKeyBytes())

			require.NoError(t, src.SetSignature(otherUsrSigner))
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Equal(t, otherUsr.PublicKeyBytes, dst.IssuerPublicKeyBytes())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredAPIFields(src)

			require.NoError(t, dst.SetSignature(usrSigner))
			src.WriteToV2(&msg)
			require.Zero(t, msg.Signature)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.False(t, dst.VerifySignature())
			require.Equal(t, src.Issuer(), dst.Issuer())
			require.Zero(t, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.SetSignature(otherUsrSigner))
			require.NoError(t, src.SetSignature(usrSigner))
			src.WriteToV2(&msg)
			require.Equal(t, usr.PublicKeyBytes, msg.Signature.Key)
			require.NotEmpty(t, msg.Signature.Sign)
			require.EqualValues(t, usrSigner.Scheme(), msg.Signature.Scheme)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.VerifySignature())
			require.Equal(t, src.Issuer(), dst.Issuer())
			require.Equal(t, usr.PublicKeyBytes, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.SetSignature(usrSigner))
			require.NoError(t, src.SetSignature(otherUsrSigner))
			src.WriteToV2(&msg)
			require.Equal(t, otherUsr.PublicKeyBytes, msg.Signature.Key)
			require.NotEmpty(t, msg.Signature.Sign)
			require.EqualValues(t, otherUsrSigner.Scheme(), msg.Signature.Scheme)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.True(t, dst.VerifySignature())
			require.Equal(t, src.Issuer(), dst.Issuer())
			require.Equal(t, otherUsr.PublicKeyBytes, dst.IssuerPublicKeyBytes())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			require.NoError(t, dst.SetSignature(usrSigner))
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.False(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Zero(t, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.SetSignature(otherUsrSigner))
			require.NoError(t, src.SetSignature(usrSigner))
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Equal(t, usr.PublicKeyBytes, dst.IssuerPublicKeyBytes())

			require.NoError(t, dst.SetSignature(usrSigner))
			require.NoError(t, src.SetSignature(otherUsrSigner))
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.True(t, dst.VerifySignature())
			require.Zero(t, dst.Issuer())
			require.Equal(t, otherUsr.PublicKeyBytes, dst.IssuerPublicKeyBytes())
		})
	})
}

func testIssuer[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, setRequiredAPIFields func(PTR)) {
	var c PTR = new(T)
	usr, otherUsr := usertest.TwoUsers()

	require.Zero(t, c.Issuer())

	c.SetIssuer(usr.ID)
	require.Equal(t, usr.ID, c.Issuer())

	c.SetIssuer(otherUsr.ID)
	require.Equal(t, otherUsr.ID, c.Issuer())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			dst.SetIssuer(usr.ID)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.Issuer())

			src.SetIssuer(usr.ID)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, usr.ID, dst.Issuer())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredAPIFields(src)

			src.SetIssuer(usr.ID)
			src.WriteToV2(&msg)
			require.Equal(t, &refs.OwnerID{Value: usr.ID[:]}, msg.Body.OwnerId)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, usr.ID, dst.Issuer())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			dst.SetIssuer(usr.ID)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.Issuer())

			src.SetIssuer(usr.ID)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, usr.ID, dst.Issuer())
		})
	})
}

func testID[T token, PTR interface {
	*T
	tokenptr
}](t *testing.T, setRequiredAPIFields func(PTR)) {
	var c PTR = new(T)

	require.Zero(t, c.ID())

	id := uuid.New()
	c.SetID(id)
	require.Equal(t, id, c.ID())

	otherID := id
	otherID[0]++
	c.SetID(otherID)
	require.Equal(t, otherID, c.ID())

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			dst.SetID(id)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, dst.ID())

			src.SetID(id)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Equal(t, id, dst.ID())
		})
		t.Run("api", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredAPIFields(src)

			src.SetID(id)
			src.WriteToV2(&msg)
			require.Equal(t, id[:], msg.Body.Id)
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Equal(t, id, dst.ID())
		})
		t.Run("json", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			dst.SetID(id)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, dst.ID())

			src.SetID(id)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dst.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Equal(t, id, dst.ID())
		})
	})
}

func testInvalidAt[T token, PTR interface {
	*T
	tokenptr
}](t testing.TB, _ T) {
	var x PTR = new(T)

	nbf := rand.Uint64()
	if nbf == math.MaxUint64 {
		nbf--
	}

	iat := nbf
	exp := iat + 1

	x.SetNbf(nbf)
	x.SetIat(iat)
	x.SetExp(exp)

	require.True(t, x.InvalidAt(nbf-1))
	require.True(t, x.InvalidAt(iat-1))
	require.False(t, x.InvalidAt(iat))
	require.False(t, x.InvalidAt(exp))
	require.True(t, x.InvalidAt(exp+1))
}

func testLifetimeField[T token, PTR interface {
	*T
	tokenptr
}](
	t *testing.T,
	get func(T) uint64,
	set func(PTR, uint64),
	getAPI func(token *apisession.SessionToken_Body_TokenLifetime) uint64,
	setRequiredAPIFields func(PTR),
) {
	var c T

	require.Zero(t, get(c))

	val := rand.Uint64()
	set(&c, val)
	require.EqualValues(t, val, get(c))

	otherVal := val * 100
	set(&c, otherVal)
	require.EqualValues(t, otherVal, get(c))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst PTR = new(T), new(T)

			set(dst, val)
			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, get(*dst))

			set(src, val)
			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.EqualValues(t, val, get(*dst))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst T
			var srcPtr, dstPtr PTR = &src, &dst
			var msg apisession.SessionToken

			// set required data just to satisfy decoder
			setRequiredAPIFields(srcPtr)

			set(&src, val)
			src.WriteToV2(&msg)
			require.EqualValues(t, val, getAPI(msg.Body.Lifetime))
			require.NoError(t, dstPtr.ReadFromV2(&msg))
			require.EqualValues(t, val, get(dst))
		})
		t.Run("json", func(t *testing.T) {
			var src, dst T
			var dstPtr PTR = &dst

			set(&dst, val)
			j, err := src.MarshalJSON()
			require.NoError(t, err)
			err = dstPtr.UnmarshalJSON(j)
			require.NoError(t, err)
			require.Zero(t, get(dst))

			set(&src, val)
			j, err = src.MarshalJSON()
			require.NoError(t, err)
			err = dstPtr.UnmarshalJSON(j)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
	})
}
