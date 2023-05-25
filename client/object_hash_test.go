package client

import (
	"bytes"
	"context"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestPrmObjectHash_ByAddress(t *testing.T) {
	var prm PrmObjectHash

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

func TestClient_ObjectHash(t *testing.T) {
	c := newClient(t, nil, nil)

	t.Run("missing signer", func(t *testing.T) {
		var nonilAddr v2refs.Address
		nonilAddr.SetObjectID(new(v2refs.ObjectID))
		nonilAddr.SetContainerID(new(v2refs.ContainerID))

		var reqBody v2object.GetRangeHashRequestBody
		reqBody.SetRanges(make([]v2object.Range, 1))

		_, err := c.ObjectHash(context.Background(), PrmObjectHash{
			addr: nonilAddr,
			body: reqBody,
		})

		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
