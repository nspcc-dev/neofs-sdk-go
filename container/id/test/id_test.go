package cidtest_test

import (
	"math/rand/v2"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
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
	require.NoError(t, id2.ReadFromV2(m))
}

func TestNIDs(t *testing.T) {
	n := rand.Int() % 10
	require.Len(t, cidtest.IDs(n), n)
}

func TestOtherID(t *testing.T) {
	ids := cidtest.IDs(100)
	require.NotContains(t, ids, cidtest.OtherID(ids...))
}
