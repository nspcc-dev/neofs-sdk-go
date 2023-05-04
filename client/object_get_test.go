package client

import (
	"bytes"
	"math/rand"
	"testing"

	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
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

func TestPrmObjectRange_SetRange(t *testing.T) {
	var prm PrmObjectRange

	var (
		ln  = rand.Uint64()
		off = rand.Uint64()
		rng *object.Range
	)

	t.Run("SetLength", func(t *testing.T) {
		prm.SetLength(ln)
		rng = object.NewRangeFromV2(&prm.rng)

		require.Equal(t, ln, rng.GetLength())
	})

	t.Run("SetOffset", func(t *testing.T) {
		prm.SetOffset(off)
		rng = object.NewRangeFromV2(&prm.rng)

		require.Equal(t, off, rng.GetOffset())
	})

	t.Run("SetRange", func(t *testing.T) {
		var tmp object.Range
		tmp.SetLength(ln)
		tmp.SetOffset(off)

		prm.SetRange(tmp)
		require.Equal(t, ln, tmp.ToV2().GetLength())
		require.Equal(t, off, tmp.ToV2().GetOffset())
	})
}
