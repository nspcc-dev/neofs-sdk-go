package object

import (
	"crypto/sha256"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/tombstone"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func generateIDList(sz int) []oid.ID {
	res := make([]oid.ID, sz)
	cs := [sha256.Size]byte{}

	for i := 0; i < sz; i++ {
		res[i] = oid.ID{}
		rand.Read(cs[:])
		res[i].SetSHA256(cs)
	}

	return res
}

func TestTombstone(t *testing.T) {
	var ts Tombstone

	exp := uint64(13)
	ts.SetExpirationEpoch(exp)
	require.Equal(t, exp, ts.ExpirationEpoch())

	splitID := NewSplitID()
	ts.SetSplitID(splitID)
	require.Equal(t, splitID, ts.SplitID())

	members := generateIDList(3)
	ts.SetMembers(members)
	require.Equal(t, members, ts.Members())
}

func TestTombstoneEncoding(t *testing.T) {
	var ts Tombstone
	ts.SetExpirationEpoch(13)
	ts.SetSplitID(NewSplitID())
	ts.SetMembers(generateIDList(5))

	t.Run("binary", func(t *testing.T) {
		data, err := ts.Marshal()
		require.NoError(t, err)

		var ts2 Tombstone
		require.NoError(t, ts2.Unmarshal(data))

		require.Equal(t, ts, ts2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := ts.MarshalJSON()
		require.NoError(t, err)

		var ts2 Tombstone
		require.NoError(t, ts2.UnmarshalJSON(data))

		require.Equal(t, ts, ts2)
	})
}

func TestNewTombstoneFromV2(t *testing.T) {
	t.Run("from zero V2", func(t *testing.T) {
		var (
			v2 tombstone.Tombstone
			x  Tombstone
		)

		x.ReadFromV2(v2)

		require.Nil(t, x.SplitID())
		require.Nil(t, x.Members())
		require.Zero(t, x.ExpirationEpoch())
	})
}

func TestNewTombstone(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var ts Tombstone

		// check initial values
		require.Nil(t, ts.SplitID())
		require.Nil(t, ts.Members())
		require.Zero(t, ts.ExpirationEpoch())
	})
}
