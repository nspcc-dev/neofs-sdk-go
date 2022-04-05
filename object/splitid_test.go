package object_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/stretchr/testify/require"
)

func TestSplitID(t *testing.T) {
	id := object.NewSplitID()

	t.Run("toV2/fromV2", func(t *testing.T) {
		data := id.ToBytes()

		newID := object.NewSplitIDFromBytes(data)
		require.NotNil(t, newID)

		require.Equal(t, id, newID)
	})

	t.Run("string/parse", func(t *testing.T) {
		idStr := id.String()

		newID := object.NewSplitID()
		require.NoError(t, newID.Parse(idStr))

		require.Equal(t, id, newID)
	})

	t.Run("set UUID", func(t *testing.T) {
		newUUID := uuid.New()
		id.SetUUID(newUUID)

		require.Equal(t, newUUID.String(), id.String())
	})

	t.Run("zero value", func(t *testing.T) {
		var newID object.SplitID

		require.NotPanics(t, func() {
			require.Nil(t, newID.ToBytes())
			require.Equal(t, "", newID.String())
		})
	})
}

func TestSplitID_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x object.SplitID

		require.Nil(t, x.ToBytes())
	})
}

func TestNewIDFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x []byte

		require.True(t, object.NewSplitIDFromBytes(x).Empty())
	})
}
