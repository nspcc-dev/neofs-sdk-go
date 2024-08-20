package oidtest_test

import (
	"math/rand/v2"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	id := oidtest.ID()
	require.NotEqual(t, id, oidtest.ID())

	var m refs.ObjectID
	id.WriteToV2(&m)
	var id2 oid.ID
	require.NoError(t, id2.ReadFromV2(m))
}

func TestNIDs(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, oidtest.IDs(n), n)
}

func TestOtherID(t *testing.T) {
	ids := oidtest.IDs(100)
	require.NotContains(t, ids, oidtest.OtherID(ids...))
}

func TestAddress(t *testing.T) {
	a := oidtest.Address()
	require.NotEqual(t, a, oidtest.Address())

	var m refs.Address
	a.WriteToV2(&m)
	var id2 oid.Address
	require.NoError(t, id2.ReadFromV2(m))
}

func TestChangeAddress(t *testing.T) {
	addrs := oidtest.Addresses(100)
	require.NotContains(t, addrs, oidtest.OtherAddress(addrs...))
}

func TestNAddresses(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, oidtest.Addresses(n), n)
}
