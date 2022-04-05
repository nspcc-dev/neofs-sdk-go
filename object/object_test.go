package object_test

import (
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	"github.com/stretchr/testify/require"
)

func TestInitCreation(t *testing.T) {
	var o object.Object
	cnr := cidtest.ID()
	own := *ownertest.ID()

	object.InitCreation(&o, object.RequiredFields{
		Container: cnr,
		Owner:     own,
	})

	require.Equal(t, cnr, o.ContainerID())
	require.Equal(t, own, o.OwnerID())
}

func TestEncoding(t *testing.T) {
	o := *objecttest.Object()

	t.Run("binary", func(t *testing.T) {
		data, err := o.Marshal()
		require.NoError(t, err)

		var o2 object.Object
		require.NoError(t, o2.Unmarshal(data))

		require.Equal(t, o, o2)
	})

	t.Run("binary", func(t *testing.T) {
		data, err := o.Marshal()
		require.NoError(t, err)

		var o2 object.Object
		require.NoError(t, o2.Unmarshal(data))

		require.Equal(t, o, o2)
	})
}
