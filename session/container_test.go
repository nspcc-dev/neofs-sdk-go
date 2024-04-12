package session_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestContainerProtocolV2(t *testing.T) {
	var validV2 v2session.Token

	var body v2session.TokenBody
	validV2.SetBody(&body)

	// ID
	id := uuid.New()
	binID, err := id.MarshalBinary()
	require.NoError(t, err)
	restoreID := func() {
		body.SetID(binID)
	}
	restoreID()

	// Owner
	usr := usertest.ID(t)
	var usrV2 refs.OwnerID
	usr.WriteToV2(&usrV2)
	restoreUser := func() {
		body.SetOwnerID(&usrV2)
	}
	restoreUser()

	// Lifetime
	var lifetime v2session.TokenLifetime
	lifetime.SetIat(1)
	lifetime.SetNbf(2)
	lifetime.SetExp(3)
	restoreLifetime := func() {
		body.SetLifetime(&lifetime)
	}
	restoreLifetime()

	// Session key
	signer := test.RandomSignerRFC6979(t)
	authKey := signer.Public()
	binAuthKey := neofscrypto.PublicKeyBytes(authKey)
	restoreAuthKey := func() {
		body.SetSessionKey(binAuthKey)
	}
	restoreAuthKey()

	// Context
	cnr := cidtest.ID()
	var cnrV2 refs.ContainerID
	cnr.WriteToV2(&cnrV2)
	var cCnr v2session.ContainerSessionContext
	restoreCtx := func() {
		cCnr.SetContainerID(&cnrV2)
		cCnr.SetWildcard(false)
		body.SetContext(&cCnr)
	}
	restoreCtx()

	// Signature
	var sig refs.Signature
	restoreSig := func() {
		validV2.SetSignature(&sig)
	}
	restoreSig()

	// TODO(@cthulhu-rider): #260 use functionality for message corruption

	for _, testcase := range []struct {
		name      string
		corrupt   []func()
		restore   func()
		assert    func(session.Container)
		breakSign func(*v2session.Token)
	}{
		{
			name: "Signature",
			corrupt: []func(){
				func() {
					validV2.SetSignature(nil)
				},
			},
			restore: restoreSig,
		},
		{
			name: "ID",
			corrupt: []func(){
				func() {
					body.SetID([]byte{1, 2, 3})
				},
				func() {
					id, err := uuid.NewDCEPerson()
					require.NoError(t, err)
					bindID, err := id.MarshalBinary()
					require.NoError(t, err)
					body.SetID(bindID)
				},
			},
			restore: restoreID,
			assert: func(val session.Container) {
				require.Equal(t, id, val.ID())
			},
			breakSign: func(m *v2session.Token) {
				id := m.GetBody().GetID()
				id[len(id)-1]++
			},
		},
		{
			name: "User",
			corrupt: []func(){
				func() {
					var brokenUsrV2 refs.OwnerID
					brokenUsrV2.SetValue(append(usrV2.GetValue(), 1))
					body.SetOwnerID(&brokenUsrV2)
				},
			},
			restore: restoreUser,
			assert: func(val session.Container) {
				require.Equal(t, usr, val.Issuer())
			},
			breakSign: func(m *v2session.Token) {
				id := m.GetBody().GetOwnerID().GetValue()
				copy(id, usertest.ID(t).WalletBytes())
			},
		},
		{
			name: "Lifetime",
			corrupt: []func(){
				func() {
					body.SetLifetime(nil)
				},
			},
			restore: restoreLifetime,
			assert: func(val session.Container) {
				require.True(t, val.InvalidAt(1))
				require.False(t, val.InvalidAt(2))
				require.False(t, val.InvalidAt(3))
				require.True(t, val.InvalidAt(4))
			},
			breakSign: func(m *v2session.Token) {
				lt := m.GetBody().GetLifetime()
				lt.SetIat(lt.GetIat() + 1)
			},
		},
		{
			name: "Auth key",
			corrupt: []func(){
				func() {
					body.SetSessionKey(nil)
				},
				func() {
					body.SetSessionKey([]byte{})
				},
			},
			restore: restoreAuthKey,
			assert: func(val session.Container) {
				require.True(t, val.AssertAuthKey(authKey))
			},
			breakSign: func(m *v2session.Token) {
				body := m.GetBody()
				key := body.GetSessionKey()
				cp := bytes.Clone(key)
				cp[len(cp)-1]++
				body.SetSessionKey(cp)
			},
		},
		{
			name: "Context",
			corrupt: []func(){
				func() {
					body.SetContext(nil)
				},
				func() {
					cCnr.SetWildcard(true)
				},
				func() {
					cCnr.SetContainerID(nil)
				},
				func() {
					var brokenCnr refs.ContainerID
					brokenCnr.SetValue(append(cnrV2.GetValue(), 1))
					cCnr.SetContainerID(&brokenCnr)
				},
			},
			restore: restoreCtx,
			assert: func(val session.Container) {
				require.True(t, val.AppliedTo(cnr))
				require.False(t, val.AppliedTo(cidtest.ID()))
			},
			breakSign: func(m *v2session.Token) {
				cnr := m.GetBody().GetContext().(*v2session.ContainerSessionContext).ContainerID().GetValue()
				cnr[len(cnr)-1]++
			},
		},
	} {
		var val session.Container

		for i, corrupt := range testcase.corrupt {
			corrupt()
			require.Error(t, val.ReadFromV2(validV2), testcase.name, fmt.Sprintf("corrupt #%d", i))

			testcase.restore()
			require.NoError(t, val.ReadFromV2(validV2), testcase.name)

			if testcase.assert != nil {
				testcase.assert(val)
			}

			if testcase.breakSign != nil {
				require.NoError(t, val.Sign(signer), testcase.name)
				require.True(t, val.VerifySignature(), testcase.name)

				var signedV2 v2session.Token
				val.WriteToV2(&signedV2)

				var restored session.Container
				require.NoError(t, restored.ReadFromV2(signedV2), testcase.name)
				require.True(t, restored.VerifySignature(), testcase.name)

				testcase.breakSign(&signedV2)

				require.NoError(t, restored.ReadFromV2(signedV2), testcase.name)
				require.False(t, restored.VerifySignature(), testcase.name)
			}
		}
	}
}

func TestContainer_WriteToV2(t *testing.T) {
	var val session.Container

	assert := func(baseAssert func(v2session.Token)) {
		var m v2session.Token
		val.WriteToV2(&m)
		baseAssert(m)
	}

	// ID
	id := uuid.New()

	binID, err := id.MarshalBinary()
	require.NoError(t, err)

	val.SetID(id)
	assert(func(m v2session.Token) {
		require.Equal(t, binID, m.GetBody().GetID())
	})

	// Owner/Signature
	signer := test.RandomSignerRFC6979(t)

	require.NoError(t, val.Sign(signer))

	usr := signer.UserID()

	var usrV2 refs.OwnerID
	usr.WriteToV2(&usrV2)

	assert(func(m v2session.Token) {
		require.Equal(t, &usrV2, m.GetBody().GetOwnerID())

		sig := m.GetSignature()
		require.NotZero(t, sig.GetKey())
		require.NotZero(t, sig.GetSign())
	})

	// Lifetime
	const iat, nbf, exp = 1, 2, 3
	val.SetIat(iat)
	val.SetNbf(nbf)
	val.SetExp(exp)

	assert(func(m v2session.Token) {
		lt := m.GetBody().GetLifetime()
		require.EqualValues(t, iat, lt.GetIat())
		require.EqualValues(t, nbf, lt.GetNbf())
		require.EqualValues(t, exp, lt.GetExp())
	})

	// Context
	assert(func(m v2session.Token) {
		cCnr, ok := m.GetBody().GetContext().(*v2session.ContainerSessionContext)
		require.True(t, ok)
		require.True(t, cCnr.Wildcard())
		require.Zero(t, cCnr.ContainerID())
	})

	cnr := cidtest.ID()

	var cnrV2 refs.ContainerID
	cnr.WriteToV2(&cnrV2)

	val.ApplyOnlyTo(cnr)

	assert(func(m v2session.Token) {
		cCnr, ok := m.GetBody().GetContext().(*v2session.ContainerSessionContext)
		require.True(t, ok)
		require.False(t, cCnr.Wildcard())
		require.Equal(t, &cnrV2, cCnr.ContainerID())
	})
}

func TestContainer_ApplyOnlyTo(t *testing.T) {
	var val session.Container
	var m v2session.Token
	filled := sessiontest.Container()

	assertDefaults := func() {
		cCnr, ok := m.GetBody().GetContext().(*v2session.ContainerSessionContext)
		require.True(t, ok)
		require.True(t, cCnr.Wildcard())
		require.Zero(t, cCnr.ContainerID())
	}

	assertBinary := func(baseAssert func()) {
		val2 := filled

		require.NoError(t, val2.Unmarshal(val.Marshal()))
		baseAssert()
	}

	assertJSON := func(baseAssert func()) {
		val2 := filled

		jd, err := val.MarshalJSON()
		require.NoError(t, err)

		require.NoError(t, val2.UnmarshalJSON(jd))
		baseAssert()
	}

	val.WriteToV2(&m)

	assertDefaults()
	assertBinary(assertDefaults)
	assertJSON(assertDefaults)

	// set value

	cnr := cidtest.ID()

	var cnrV2 refs.ContainerID
	cnr.WriteToV2(&cnrV2)

	val.ApplyOnlyTo(cnr)

	val.WriteToV2(&m)

	assertCnr := func() {
		cCnr, ok := m.GetBody().GetContext().(*v2session.ContainerSessionContext)
		require.True(t, ok)
		require.False(t, cCnr.Wildcard())
		require.Equal(t, &cnrV2, cCnr.ContainerID())
	}

	assertCnr()
	assertBinary(assertCnr)
	assertJSON(assertCnr)
}

func TestContainer_AppliedTo(t *testing.T) {
	var x session.Container

	cnr1 := cidtest.ID()
	cnr2 := cidtest.ID()

	require.True(t, x.AppliedTo(cnr1))
	require.True(t, x.AppliedTo(cnr2))

	x.ApplyOnlyTo(cnr1)

	require.True(t, x.AppliedTo(cnr1))
	require.False(t, x.AppliedTo(cnr2))
}

func TestContainer_InvalidAt(t *testing.T) {
	var x session.Container

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

func TestContainer_ID(t *testing.T) {
	var x session.Container

	require.Zero(t, x.ID())

	id := uuid.New()

	x.SetID(id)

	require.Equal(t, id, x.ID())
}

func TestContainer_AssertAuthKey(t *testing.T) {
	var x session.Container

	key := test.RandomSignerRFC6979(t).Public()

	require.False(t, x.AssertAuthKey(key))

	x.SetAuthKey(key)

	require.True(t, x.AssertAuthKey(key))
}

func TestContainer_ForVerb(t *testing.T) {
	var val session.Container
	var m v2session.Token
	filled := sessiontest.Container()

	assertDefaults := func() {
		cCnr, ok := m.GetBody().GetContext().(*v2session.ContainerSessionContext)
		require.True(t, ok)
		require.Zero(t, cCnr.Verb())
	}

	assertBinary := func(baseAssert func()) {
		val2 := filled

		require.NoError(t, val2.Unmarshal(val.Marshal()))
		baseAssert()
	}

	assertJSON := func(baseAssert func()) {
		val2 := filled

		jd, err := val.MarshalJSON()
		require.NoError(t, err)

		require.NoError(t, val2.UnmarshalJSON(jd))
		baseAssert()
	}

	val.WriteToV2(&m)

	assertDefaults()
	assertBinary(assertDefaults)
	assertJSON(assertDefaults)

	// set value

	assertVerb := func(verb v2session.ContainerSessionVerb) {
		cCnr, ok := m.GetBody().GetContext().(*v2session.ContainerSessionContext)
		require.True(t, ok)
		require.Equal(t, verb, cCnr.Verb())
	}

	for from, to := range map[session.ContainerVerb]v2session.ContainerSessionVerb{
		session.VerbContainerPut:     v2session.ContainerVerbPut,
		session.VerbContainerDelete:  v2session.ContainerVerbDelete,
		session.VerbContainerSetEACL: v2session.ContainerVerbSetEACL,
	} {
		val.ForVerb(from)

		val.WriteToV2(&m)

		assertVerb(to)
		assertBinary(func() { assertVerb(to) })
		assertJSON(func() { assertVerb(to) })
	}
}

func TestContainer_AssertVerb(t *testing.T) {
	var x session.Container

	const v1, v2 = session.VerbContainerPut, session.VerbContainerDelete

	require.False(t, x.AssertVerb(v1))
	require.False(t, x.AssertVerb(v2))

	x.ForVerb(v1)
	require.True(t, x.AssertVerb(v1))
	require.False(t, x.AssertVerb(v2))
}

func TestIssuedBy(t *testing.T) {
	var (
		token  session.Container
		issuer user.ID
		signer = test.RandomSignerRFC6979(t)
	)

	issuer = signer.UserID()

	require.False(t, session.IssuedBy(token, issuer))

	require.NoError(t, token.Sign(signer))
	require.True(t, session.IssuedBy(token, issuer))
}

func TestContainer_Issuer(t *testing.T) {
	t.Run("signer", func(t *testing.T) {
		var token session.Container

		signer := test.RandomSignerRFC6979(t)

		require.Zero(t, token.Issuer())
		require.NoError(t, token.Sign(signer))

		issuer := signer.UserID()
		require.True(t, token.Issuer().Equals(issuer))
	})

	t.Run("external", func(t *testing.T) {
		var token session.Container

		signer := test.RandomSignerRFC6979(t)
		issuer := signer.UserID()

		token.SetIssuer(issuer)
		require.True(t, token.Issuer().Equals(issuer))
	})

	t.Run("public key", func(t *testing.T) {
		var token session.Container

		signer := test.RandomSignerRFC6979(t)

		require.Nil(t, token.IssuerPublicKeyBytes())
		require.NoError(t, token.Sign(signer))

		require.Equal(t, neofscrypto.PublicKeyBytes(signer.Public()), token.IssuerPublicKeyBytes())
	})
}

func TestContainer_Sign(t *testing.T) {
	val := sessiontest.Container()

	require.NoError(t, val.SetSignature(test.RandomSignerRFC6979(t)))
	require.Zero(t, val.Issuer())
	require.True(t, val.VerifySignature())

	require.NoError(t, val.Sign(test.RandomSignerRFC6979(t)))

	require.True(t, val.VerifySignature())

	t.Run("issue#546", func(t *testing.T) {
		signer1 := test.RandomSignerRFC6979(t)
		signer2 := test.RandomSignerRFC6979(t)
		require.False(t, signer1.UserID().Equals(signer2.UserID()))

		token1 := sessiontest.Container()
		require.NoError(t, token1.Sign(signer1))
		require.Equal(t, signer1.UserID(), token1.Issuer())

		// copy token and re-sign
		var token2 session.Container
		token1.CopyTo(&token2)
		require.NoError(t, token2.Sign(signer2))
		require.Equal(t, signer2.UserID(), token2.Issuer())
	})
}

func TestContainer_SignedData(t *testing.T) {
	id := usertest.ID(t)

	val := sessiontest.Container()
	val.SetIssuer(id)

	signedData := val.SignedData()
	var dec session.Container
	require.NoError(t, dec.UnmarshalSignedData(signedData))
	require.Equal(t, val, dec)

	signer := user.NewSigner(test.RandomSigner(t), id)
	test.SignedDataComponentUser(t, signer, &val)
}

func TestContainer_VerifyDataSignature(t *testing.T) {
	signer := test.RandomSignerRFC6979(t)

	var tok session.Container

	data := make([]byte, 100)
	//nolint:staticcheck
	rand.Read(data)

	var sig neofscrypto.Signature
	require.NoError(t, sig.Calculate(signer, data))

	var sigV2 refs.Signature
	sig.WriteToV2(&sigV2)

	require.False(t, tok.VerifySessionDataSignature(data, sigV2.GetSign()))

	tok.SetAuthKey(signer.Public())
	require.True(t, tok.VerifySessionDataSignature(data, sigV2.GetSign()))
	require.False(t, tok.VerifySessionDataSignature(append(data, 1), sigV2.GetSign()))
	require.False(t, tok.VerifySessionDataSignature(data, append(sigV2.GetSign(), 1)))
}
