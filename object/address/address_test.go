package address

import (
	"strings"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestAddress_SetContainerID(t *testing.T) {
	var a Address

	id := cidtest.ID()

	a.SetContainerID(id)

	require.Equal(t, id, a.ContainerID())
}

func TestAddress_SetObjectID(t *testing.T) {
	var a Address

	oid := oidtest.ID()

	a.SetObjectID(oid)

	require.Equal(t, oid, a.ObjectID())
}

func TestAddress_Parse(t *testing.T) {
	cid := cidtest.ID()

	oid := oidtest.ID()

	t.Run("should parse successful", func(t *testing.T) {
		s := strings.Join([]string{cid.String(), oid.String()}, addressSeparator)
		var a Address

		require.NoError(t, a.Parse(s))
		require.Equal(t, oid, a.ObjectID())
		require.Equal(t, cid, a.ContainerID())
	})

	t.Run("should fail for bad address", func(t *testing.T) {
		s := strings.Join([]string{cid.String()}, addressSeparator)
		require.EqualError(t, (&Address{}).Parse(s), errInvalidAddressString.Error())
	})

	t.Run("should fail on container.ID", func(t *testing.T) {
		s := strings.Join([]string{"1", "2"}, addressSeparator)
		require.Error(t, (&Address{}).Parse(s))
	})

	t.Run("should fail on object.ID", func(t *testing.T) {
		s := strings.Join([]string{cid.String(), "2"}, addressSeparator)
		require.Error(t, (&Address{}).Parse(s))
	})
}

func TestAddressEncoding(t *testing.T) {
	var a Address
	a.SetObjectID(oidtest.ID())
	a.SetContainerID(cidtest.ID())

	t.Run("binary", func(t *testing.T) {
		data, err := a.Marshal()
		require.NoError(t, err)

		var a2 Address
		require.NoError(t, a2.Unmarshal(data))

		require.Equal(t, a, a2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := a.MarshalJSON()
		require.NoError(t, err)

		var a2 Address
		require.NoError(t, a2.UnmarshalJSON(data))

		require.Equal(t, a, a2)
	})
}

func TestNewAddressFromV2(t *testing.T) {
	t.Run("from zero V2", func(t *testing.T) {
		var (
			x  Address
			v2 refs.Address
		)

		x.ReadFromV2(v2)

		require.True(t, x.ObjectID().Empty())
		require.True(t, x.ContainerID().Empty())
		require.Equal(t, "/", x.String())
	})
}

func TestAddress_ToV2(t *testing.T) {
	t.Run("zero to V2", func(t *testing.T) {
		var (
			x  Address
			v2 refs.Address
		)

		x.WriteToV2(&v2)

		require.Nil(t, v2.GetObjectID())
		require.Nil(t, v2.GetContainerID())
	})
}

func TestNewAddress(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var a Address

		// check initial values
		require.True(t, a.ContainerID().Empty())
		require.True(t, a.ObjectID().Empty())
		require.Equal(t, "/", a.String())
	})
}
