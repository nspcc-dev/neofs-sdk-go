package session_test

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	addresstest "github.com/nspcc-dev/neofs-sdk-go/object/address/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func randSigner() ecdsa.PrivateKey {
	k, err := keys.NewPrivateKey()
	if err != nil {
		panic(fmt.Sprintf("generate private key: %v", err))
	}

	return k.PrivateKey
}

func randPublicKey() neofscrypto.PublicKey {
	k := randSigner().PublicKey
	return (*neofsecdsa.PublicKey)(&k)
}

func TestObject_ReadFromV2(t *testing.T) {
	var x session.Object
	var m v2session.Token
	var b v2session.TokenBody
	var c v2session.ObjectSessionContext
	id := uuid.New()

	t.Run("protocol violation", func(t *testing.T) {
		require.Error(t, x.ReadFromV2(m))

		m.SetBody(&b)

		require.Error(t, x.ReadFromV2(m))

		b.SetID(id[:])

		require.Error(t, x.ReadFromV2(m))

		b.SetContext(&c)

		require.NoError(t, x.ReadFromV2(m))
	})

	m.SetBody(&b)
	b.SetContext(&c)
	b.SetID(id[:])

	t.Run("object", func(t *testing.T) {
		var obj address.Address

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AppliedTo(obj))

		obj = *addresstest.Address()

		objv2 := *obj.ToV2()

		c.SetAddress(&objv2)

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AppliedTo(obj))
	})

	t.Run("verb", func(t *testing.T) {
		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AssertVerb(0))

		verb := v2session.ObjectSessionVerb(rand.Uint32())

		c.SetVerb(verb)

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AssertVerb(session.ObjectVerb(verb)))
	})

	t.Run("id", func(t *testing.T) {
		id := uuid.New()
		bID := id[:]

		b.SetID(bID)

		require.NoError(t, x.ReadFromV2(m))
		require.Equal(t, id, x.ID())
	})

	t.Run("lifetime", func(t *testing.T) {
		const nbf, iat, exp = 11, 22, 33

		var lt v2session.TokenLifetime
		lt.SetNbf(nbf)
		lt.SetIat(iat)
		lt.SetExp(exp)

		b.SetLifetime(&lt)

		require.NoError(t, x.ReadFromV2(m))
		require.False(t, x.ExpiredAt(exp-1))
		require.True(t, x.ExpiredAt(exp))
		require.True(t, x.ExpiredAt(exp+1))
		require.True(t, x.InvalidAt(nbf-1))
		require.True(t, x.InvalidAt(iat-1))
		require.False(t, x.InvalidAt(iat))
		require.False(t, x.InvalidAt(exp-1))
		require.True(t, x.InvalidAt(exp))
		require.True(t, x.InvalidAt(exp+1))
	})

	t.Run("session key", func(t *testing.T) {
		key := randPublicKey()

		bKey := make([]byte, key.MaxEncodedSize())
		bKey = bKey[:key.Encode(bKey)]

		b.SetSessionKey(bKey)

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AssertAuthKey(key))
	})
}

func TestEncodingObject(t *testing.T) {
	tok := *sessiontest.ObjectSigned()

	t.Run("binary", func(t *testing.T) {
		data := tok.Marshal()

		var tok2 session.Object
		require.NoError(t, tok2.Unmarshal(data))

		require.Equal(t, tok, tok2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := tok.MarshalJSON()
		require.NoError(t, err)

		var tok2 session.Object
		require.NoError(t, tok2.UnmarshalJSON(data))

		require.Equal(t, tok, tok2)
	})
}

func TestObjectAppliedTo(t *testing.T) {
	var x session.Object

	a := *addresstest.Address()

	require.False(t, x.AppliedTo(a))

	x.ApplyTo(a)

	require.True(t, x.AppliedTo(a))
}

func TestObjectExp(t *testing.T) {
	var x session.Object

	exp := rand.Uint64()

	require.True(t, x.ExpiredAt(exp))

	x.SetExp(exp)

	require.False(t, x.ExpiredAt(exp-1))
	require.True(t, x.ExpiredAt(exp))
	require.True(t, x.ExpiredAt(exp+1))
}

func TestObjectLifetime(t *testing.T) {
	var x session.Object

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
	require.True(t, x.InvalidAt(exp))
}

func TestObjectID(t *testing.T) {
	var x session.Object

	require.Zero(t, x.ID())

	id := uuid.New()

	x.SetID(id)

	require.Equal(t, id, x.ID())
}

func TestObjectAuthKey(t *testing.T) {
	var x session.Object

	key := randPublicKey()

	require.False(t, x.AssertAuthKey(key))

	x.SetAuthKey(key)

	require.True(t, x.AssertAuthKey(key))
}

func TestObjectVerb(t *testing.T) {
	var x session.Object

	const v1, v2 = session.VerbObjectGet, session.VerbObjectPut

	require.False(t, x.AssertVerb(v1, v2))

	x.ForVerb(v1)
	require.True(t, x.AssertVerb(v1))
	require.False(t, x.AssertVerb(v2))
	require.True(t, x.AssertVerb(v1, v2))
	require.True(t, x.AssertVerb(v2, v1))
}

func TestObjectSignature(t *testing.T) {
	var x session.Object

	const nbf = 11
	const iat = 22
	const exp = 33
	id := uuid.New()
	key := randPublicKey()
	obj := *addresstest.Address()
	verb := session.VerbObjectDelete

	signer := randSigner()

	fs := []func(){
		func() { x.SetNbf(nbf) },
		func() { x.SetNbf(nbf + 1) },

		func() { x.SetIat(iat) },
		func() { x.SetIat(iat + 1) },

		func() { x.SetExp(exp) },
		func() { x.SetExp(exp + 1) },

		func() { x.SetID(id) },
		func() {
			idcp := id
			idcp[0]++
			x.SetID(idcp)
		},

		func() { x.SetAuthKey(key) },
		func() { x.SetAuthKey(randPublicKey()) },

		func() { x.ApplyTo(obj) },
		func() { x.ApplyTo(*addresstest.Address()) },

		func() { x.ForVerb(verb) },
		func() { x.ForVerb(verb + 1) },
	}

	for i := 0; i < len(fs); i += 2 {
		fs[i]()

		require.NoError(t, x.Sign(signer))
		require.True(t, x.VerifySignature())

		fs[i+1]()
		require.False(t, x.VerifySignature())

		fs[i]()
		require.True(t, x.VerifySignature())
	}
}

func TestObject_Issuer(t *testing.T) {
	var token session.Object
	signer := randSigner()

	require.Zero(t, token.Issuer())

	require.NoError(t, token.Sign(signer))

	var issuer user.ID

	user.IDFromKey(&issuer, signer.PublicKey)

	require.True(t, token.Issuer().Equals(issuer))
}
