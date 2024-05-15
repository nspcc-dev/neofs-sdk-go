package cidtest_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	id := cidtest.ID()
	require.NotEqual(t, id, cidtest.ID())

	var m refs.ContainerID
	id.WriteToV2(&m)
	var id2 cid.ID
	require.NoError(t, id2.ReadFromV2(&m))
}

func TestChangeID(t *testing.T) {
	id := cidtest.ID()
	require.NotEqual(t, id, cidtest.ChangeID(id))
}

func TestNIDs(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, cidtest.NIDs(n), n)
}
