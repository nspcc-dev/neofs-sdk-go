package session_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/stretchr/testify/require"
)

func testLifetimeClaim[T session.Container | session.Object](t testing.TB, get func(T) uint64, set func(*T, uint64)) {
	var x T
	require.Zero(t, get(x))
	set(&x, 12094032)
	require.EqualValues(t, 12094032, get(x))
	set(&x, 5469830342)
	require.EqualValues(t, 5469830342, get(x))
}

type lifetime interface {
	SetExp(uint64)
	SetIat(uint64)
	SetNbf(uint64)
	InvalidAt(uint64) bool
}

func testInvalidAt(t testing.TB, x lifetime) {
	require.False(t, x.InvalidAt(0))

	const iat = 13
	const nbf = iat + 1
	const exp = nbf + 1

	x.SetIat(iat)
	x.SetNbf(nbf)
	x.SetExp(exp)

	require.True(t, x.InvalidAt(iat-1))
	require.True(t, x.InvalidAt(iat))
	require.False(t, x.InvalidAt(nbf))
	require.False(t, x.InvalidAt(exp))
	require.True(t, x.InvalidAt(exp+1))
}

func TestInvalidAt(t *testing.T) {
	testInvalidAt(t, new(session.Container))
	testInvalidAt(t, new(session.Object))
}
