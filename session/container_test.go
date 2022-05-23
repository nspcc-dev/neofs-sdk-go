package session_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func TestContainer_ReadFromV2(t *testing.T) {
	var x session.Container
	var m v2session.Token
	var b v2session.TokenBody
	var c v2session.ContainerSessionContext
	id := uuid.New()

	t.Run("protocol violation", func(t *testing.T) {
		require.Error(t, x.ReadFromV2(m))

		m.SetBody(&b)

		require.Error(t, x.ReadFromV2(m))

		b.SetID(id[:])

		require.Error(t, x.ReadFromV2(m))

		b.SetContext(&c)

		require.Error(t, x.ReadFromV2(m))

		c.SetWildcard(true)

		require.NoError(t, x.ReadFromV2(m))
	})

	m.SetBody(&b)
	b.SetContext(&c)
	b.SetID(id[:])
	c.SetWildcard(true)

	t.Run("container", func(t *testing.T) {
		cnr1 := cidtest.ID()
		cnr2 := cidtest.ID()

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AppliedTo(cnr1))
		require.True(t, x.AppliedTo(cnr2))

		var cnrv2 refs.ContainerID
		cnr1.WriteToV2(&cnrv2)

		c.SetContainerID(&cnrv2)
		c.SetWildcard(false)

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AppliedTo(cnr1))
		require.False(t, x.AppliedTo(cnr2))
	})

	t.Run("verb", func(t *testing.T) {
		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AssertVerb(0))

		verb := v2session.ContainerSessionVerb(rand.Uint32())

		c.SetVerb(verb)

		require.NoError(t, x.ReadFromV2(m))
		require.True(t, x.AssertVerb(session.ContainerVerb(verb)))
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

func TestEncodingContainer(t *testing.T) {
	tok := *sessiontest.ContainerSigned()

	t.Run("binary", func(t *testing.T) {
		data := tok.Marshal()

		var tok2 session.Container
		require.NoError(t, tok2.Unmarshal(data))

		require.Equal(t, tok, tok2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := tok.MarshalJSON()
		require.NoError(t, err)

		var tok2 session.Container
		require.NoError(t, tok2.UnmarshalJSON(data))

		require.Equal(t, tok, tok2)
	})
}

func TestContainerAppliedTo(t *testing.T) {
	var x session.Container

	cnr1 := cidtest.ID()
	cnr2 := cidtest.ID()

	require.True(t, x.AppliedTo(cnr1))
	require.True(t, x.AppliedTo(cnr2))

	x.ApplyOnlyTo(cnr1)

	require.True(t, x.AppliedTo(cnr1))
	require.False(t, x.AppliedTo(cnr2))
}

func TestContainerExp(t *testing.T) {
	var x session.Container

	exp := rand.Uint64()

	require.True(t, x.ExpiredAt(exp))

	x.SetExp(exp)

	require.False(t, x.ExpiredAt(exp-1))
	require.True(t, x.ExpiredAt(exp))
	require.True(t, x.ExpiredAt(exp+1))
}

func TestContainerLifetime(t *testing.T) {
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
	require.True(t, x.InvalidAt(exp))
}

func TestContainerID(t *testing.T) {
	var x session.Container

	require.Zero(t, x.ID())

	id := uuid.New()

	x.SetID(id)

	require.Equal(t, id, x.ID())
}

func TestContainerAuthKey(t *testing.T) {
	var x session.Container

	key := randPublicKey()

	require.False(t, x.AssertAuthKey(key))

	x.SetAuthKey(key)

	require.True(t, x.AssertAuthKey(key))
}

func TestContainerVerb(t *testing.T) {
	var x session.Container

	const v1, v2 = session.VerbContainerPut, session.VerbContainerDelete

	require.False(t, x.AssertVerb(v1))
	require.False(t, x.AssertVerb(v2))

	x.ForVerb(v1)
	require.True(t, x.AssertVerb(v1))
	require.False(t, x.AssertVerb(v2))
}

func TestContainerSignature(t *testing.T) {
	var x session.Container

	const nbf = 11
	const iat = 22
	const exp = 33
	id := uuid.New()
	key := randPublicKey()
	cnr := cidtest.ID()
	verb := session.VerbContainerPut

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

		func() { x.ApplyOnlyTo(cnr) },
		func() { x.ApplyOnlyTo(cidtest.ID()) },

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

func TestIssuedBy(t *testing.T) {
	var (
		token  session.Container
		issuer user.ID
		signer = randSigner()
	)

	user.IDFromKey(&issuer, signer.PublicKey)

	require.False(t, session.IssuedBy(token, issuer))

	require.NoError(t, token.Sign(signer))
	require.True(t, session.IssuedBy(token, issuer))
}

func TestContainer_Issuer(t *testing.T) {
	var token session.Container
	signer := randSigner()

	require.Zero(t, token.Issuer())

	require.NoError(t, token.Sign(signer))

	var issuer user.ID

	user.IDFromKey(&issuer, signer.PublicKey)

	require.True(t, token.Issuer().Equals(issuer))
}
