package usertest_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	id := usertest.ID()
	require.NotEqual(t, id, usertest.ID())

	var m refs.OwnerID
	id.WriteToV2(&m)
	var id2 user.ID
	require.NoError(t, id2.ReadFromV2(m))
}

func TestNIDs(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, cidtest.IDs(n), n)
}

func TestOtherID(t *testing.T) {
	ids := usertest.IDs(100)
	require.NotContains(t, ids, usertest.OtherID(ids...))
}
