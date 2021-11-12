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
	t.Run("non-nil", func(t *testing.T) {
		checksum := randSHA256Checksum()

		id := cidtest.IDWithChecksum(checksum)

		idV2 := id.ToV2()

		require.Equal(t, id, cid.NewFromV2(idV2))
		require.Equal(t, checksum[:], idV2.GetValue())
	})

	t.Run("nil", func(t *testing.T) {
		var x *cid.ID

		require.Nil(t, x.ToV2())
	})

	t.Run("default values", func(t *testing.T) {
		id := cid.New()

		// convert to v2 message
		cidV2 := id.ToV2()
		require.Nil(t, cidV2.GetValue())
	})
}

func TestID_Equal(t *testing.T) {
	cs := randSHA256Checksum()

	id1 := cidtest.IDWithChecksum(cs)
	id2 := cidtest.IDWithChecksum(cs)

	require.True(t, id1.Equal(id2))

	id3 := cidtest.ID()

	require.False(t, id1.Equal(id3))
}

func TestID_String(t *testing.T) {
	t.Run("Parse/String", func(t *testing.T) {
		id := cidtest.ID()
		id2 := cid.New()

		require.NoError(t, id2.Parse(id.String()))
		require.Equal(t, id, id2)
	})

	t.Run("nil", func(t *testing.T) {
		id := cid.New()

		require.Empty(t, id.String())
	})
}

func TestContainerIDEncoding(t *testing.T) {
	id := cidtest.ID()

	t.Run("binary", func(t *testing.T) {
		data, err := id.Marshal()
		require.NoError(t, err)

		id2 := cid.New()
		require.NoError(t, id2.Unmarshal(data))

		require.Equal(t, id, id2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := id.MarshalJSON()
		require.NoError(t, err)

		a2 := cid.New()
		require.NoError(t, a2.UnmarshalJSON(data))

		require.Equal(t, id, a2)
	})
}

func TestNewFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *refs.ContainerID

		require.Nil(t, cid.NewFromV2(x))
	})
}
