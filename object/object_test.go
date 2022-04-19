package object_test

import (
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestInitCreation(t *testing.T) {
	var o object.Object
	cnr := cidtest.ID()
	own := *usertest.ID()

	object.InitCreation(&o, object.RequiredFields{
		Container: cnr,
		Owner:     own,
	})

	cID, set := o.ContainerID()
	require.True(t, set)
	require.Equal(t, cnr, cID)
	require.Equal(t, &own, o.OwnerID())
}
