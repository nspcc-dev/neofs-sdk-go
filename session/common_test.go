package session_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/session"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type lifetime interface {
	Exp() uint64
	SetExp(uint64)
	Iat() uint64
	SetIat(uint64)
	Nbf() uint64
	SetNbf(uint64)
}

func testInvalidAt(t testing.TB, x lifetime) {
	require.False(t, session.InvalidAt(x, 0))

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
	require.True(t, session.InvalidAt(x, iat-1))
	require.False(t, session.InvalidAt(x, iat))
	require.False(t, session.InvalidAt(x, exp))
	require.True(t, session.InvalidAt(x, exp+1))
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

func TestSetAuthPublicKey(t *testing.T) {
	testSetAuthPublicKey(t, new(session.Container))
	testSetAuthPublicKey(t, new(session.Object))
}

func TestAssertAuthPublicKey(t *testing.T) {
	testAssertAuthPublicKey(t, new(session.Container))
	testAssertAuthPublicKey(t, new(session.Object))
}
