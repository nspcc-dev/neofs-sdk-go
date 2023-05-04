package client

import (
	"bytes"
	"testing"

	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestPrmObjectRead_ByAddress(t *testing.T) {
	var prm PrmObjectHead

	var (
		objID  oid.ID
		contID cid.ID
		oidV2  v2refs.ObjectID
		cidV2  v2refs.ContainerID
	)

	t.Run("ByID", func(t *testing.T) {
		objID = randOID(t)
		prm.ByID(objID)

		objID.WriteToV2(&oidV2)

		require.True(t, bytes.Equal(oidV2.GetValue(), prm.addr.GetObjectID().GetValue()))
	})

	t.Run("FromContainer", func(t *testing.T) {
		contID = randCID(t)
		prm.FromContainer(contID)

		contID.WriteToV2(&cidV2)

		require.True(t, bytes.Equal(cidV2.GetValue(), prm.addr.GetContainerID().GetValue()))
	})

	t.Run("ByAddress", func(t *testing.T) {
		var addr oid.Address
		addr.SetObject(objID)
		addr.SetContainer(contID)

		prm.ByAddress(addr)
		require.True(t, bytes.Equal(oidV2.GetValue(), prm.addr.GetObjectID().GetValue()))
		require.True(t, bytes.Equal(cidV2.GetValue(), prm.addr.GetContainerID().GetValue()))
	})
}
