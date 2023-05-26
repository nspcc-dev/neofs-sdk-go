package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func randOID(t *testing.T) oid.ID {
	var id oid.ID
	id.SetSHA256(randSHA256Checksum(t))

	return id
}

func randCID(t *testing.T) cid.ID {
	var id cid.ID
	id.SetSHA256(randSHA256Checksum(t))

	return id
}

func randSHA256Checksum(t *testing.T) (cs [sha256.Size]byte) {
	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	return
}

func TestPrmObjectDelete_ByAddress(t *testing.T) {
	var prm PrmObjectDelete

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

func TestClient_ObjectDelete(t *testing.T) {
	t.Run("missing signer", func(t *testing.T) {
		c := newClient(t, nil, nil)

		var nonilAddr v2refs.Address
		nonilAddr.SetObjectID(new(v2refs.ObjectID))
		nonilAddr.SetContainerID(new(v2refs.ContainerID))

		_, err := c.ObjectDelete(context.Background(), PrmObjectDelete{addr: nonilAddr})
		require.ErrorIs(t, err, ErrMissingSigner)
	})
}
