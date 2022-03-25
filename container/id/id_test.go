package cid_test

import (
	"crypto/sha256"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/stretchr/testify/require"
)

func randSHA256Checksum() (cs [sha256.Size]byte) {
	rand.Read(cs[:])
	return
}

func TestID_ToV2(t *testing.T) {
	t.Run("non-zero", func(t *testing.T) {
		checksum := randSHA256Checksum()

		id := cidtest.IDWithChecksum(checksum)

		var idV2 refs.ContainerID
		id.WriteToV2(&idV2)

		var newID cid.ID
		newID.ReadFromV2(idV2)

		require.Equal(t, *id, newID)
		require.Equal(t, checksum[:], idV2.GetValue())
	})

	t.Run("zero", func(t *testing.T) {
		var (
			x  cid.ID
			v2 refs.ContainerID
		)

		x.WriteToV2(&v2)

		require.Nil(t, v2.GetValue())
	})

	t.Run("default values", func(t *testing.T) {
		var (
			id    cid.ID
			cidV2 refs.ContainerID
		)

		// convert to v2 message
		id.WriteToV2(&cidV2)
		require.Nil(t, cidV2.GetValue())
	})
}

func TestID_Equal(t *testing.T) {
	cs := randSHA256Checksum()

	id1 := cidtest.IDWithChecksum(cs)
	id2 := cidtest.IDWithChecksum(cs)

	require.True(t, id1.Equals(id2))

	id3 := cidtest.ID()

	require.False(t, id1.Equals(id3))
}

func TestID_String(t *testing.T) {
	t.Run("Parse/String", func(t *testing.T) {
		id := cidtest.ID()
		var id2 cid.ID

		require.NoError(t, id2.Parse(id.String()))
		require.Equal(t, *id, id2)
	})

	t.Run("zero", func(t *testing.T) {
		var id cid.ID

		require.Empty(t, id.String())
	})
}

func TestNewFromV2(t *testing.T) {
	t.Run("from zero", func(t *testing.T) {
		var (
			x  cid.ID
			v2 refs.ContainerID
		)

		x.ReadFromV2(v2)

		require.Empty(t, x.String())
	})
}
