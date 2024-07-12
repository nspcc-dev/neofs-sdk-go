package cid_test

import (
	"testing"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/stretchr/testify/require"
)

const emptyID = "11111111111111111111111111111111"

func TestID_ToV2(t *testing.T) {
	t.Run("non-zero", func(t *testing.T) {
		id := cidtest.ID()

		var idV2 refs.ContainerID
		id.WriteToV2(&idV2)

		var newID cid.ID
		require.NoError(t, newID.ReadFromV2(idV2))

		require.Equal(t, id, newID)
		require.Equal(t, id[:], idV2.GetValue())
	})

	t.Run("zero", func(t *testing.T) {
		var (
			x  cid.ID
			v2 refs.ContainerID
		)

		x.WriteToV2(&v2)
		require.Equal(t, emptyID, base58.Encode(v2.GetValue()))
	})
}

func TestID_Equal(t *testing.T) {
	id1 := cidtest.ID()
	require.True(t, id1.Equals(id1))
	id2 := id1
	require.True(t, id1.Equals(id2))
	require.True(t, id2.Equals(id1))
	id3 := cidtest.OtherID(id1)
	require.False(t, id1.Equals(id3))
	require.False(t, id3.Equals(id1))
	require.False(t, id2.Equals(id3))
	require.False(t, id3.Equals(id2))
}

func TestID_String(t *testing.T) {
	t.Run("DecodeString/EncodeToString", func(t *testing.T) {
		id := cidtest.ID()
		var id2 cid.ID

		require.NoError(t, id2.DecodeString(id.EncodeToString()))
		require.Equal(t, id, id2)
	})

	t.Run("zero", func(t *testing.T) {
		var id cid.ID

		require.Equal(t, emptyID, id.EncodeToString())
	})
}

func TestNewFromV2(t *testing.T) {
	t.Run("from zero", func(t *testing.T) {
		var (
			x  cid.ID
			v2 refs.ContainerID
		)

		require.Error(t, x.ReadFromV2(v2))
	})
}

func TestID_Encode(t *testing.T) {
	var id cid.ID

	t.Run("panic", func(t *testing.T) {
		dst := make([]byte, cid.Size-1)

		require.Panics(t, func() {
			id.Encode(dst)
		})
	})

	t.Run("correct", func(t *testing.T) {
		dst := make([]byte, cid.Size)

		require.NotPanics(t, func() {
			id.Encode(dst)
		})
		require.Equal(t, emptyID, id.EncodeToString())
	})
}
