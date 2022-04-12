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
	a := NewAddress()

	id := cidtest.ID()

	a.SetContainerID(id)

	cID, set := a.ContainerID()
	require.True(t, set)
	require.Equal(t, id, cID)
}

func TestAddress_SetObjectID(t *testing.T) {
	a := NewAddress()

	oid := oidtest.ID()

	a.SetObjectID(oid)

	oID, set := a.ObjectID()
	require.True(t, set)
	require.Equal(t, oid, oID)
}

func TestAddress_Parse(t *testing.T) {
	cid := cidtest.ID()

	oid := oidtest.ID()

	t.Run("should parse successful", func(t *testing.T) {
		s := strings.Join([]string{cid.String(), oid.String()}, addressSeparator)
		a := NewAddress()

		require.NoError(t, a.Parse(s))
		oID, set := a.ObjectID()
		require.True(t, set)
		require.Equal(t, oid, oID)
		cID, set := a.ContainerID()
		require.True(t, set)
		require.Equal(t, cid, cID)
	})

	t.Run("should fail for bad address", func(t *testing.T) {
		s := strings.Join([]string{cid.String()}, addressSeparator)
		require.EqualError(t, NewAddress().Parse(s), errInvalidAddressString.Error())
	})

	t.Run("should fail on container.ID", func(t *testing.T) {
		s := strings.Join([]string{"1", "2"}, addressSeparator)
		require.Error(t, NewAddress().Parse(s))
	})

	t.Run("should fail on object.ID", func(t *testing.T) {
		s := strings.Join([]string{cid.String(), "2"}, addressSeparator)
		require.Error(t, NewAddress().Parse(s))
	})
}

func TestAddressEncoding(t *testing.T) {
	a := NewAddress()
	a.SetObjectID(oidtest.ID())
	a.SetContainerID(cidtest.ID())

	t.Run("binary", func(t *testing.T) {
		data, err := a.Marshal()
		require.NoError(t, err)

		a2 := NewAddress()
		require.NoError(t, a2.Unmarshal(data))

		require.Equal(t, a, a2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := a.MarshalJSON()
		require.NoError(t, err)

		a2 := NewAddress()
		require.NoError(t, a2.UnmarshalJSON(data))

		require.Equal(t, a, a2)
	})
}

func TestNewAddressFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *refs.Address

		require.Nil(t, NewAddressFromV2(x))
	})
}

func TestAddress_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *Address

		require.Nil(t, x.ToV2())
	})
}

func TestNewAddress(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		a := NewAddress()

		// check initial values
		_, set := a.ContainerID()
		require.False(t, set)
		_, set = a.ObjectID()
		require.False(t, set)

		// convert to v2 message
		aV2 := a.ToV2()

		require.Nil(t, aV2.GetContainerID())
		require.Nil(t, aV2.GetObjectID())
	})
}
