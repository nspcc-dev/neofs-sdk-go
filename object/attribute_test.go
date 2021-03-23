package object

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/stretchr/testify/require"
)

func TestAttribute(t *testing.T) {
	key, val := "some key", "some value"

	var a Attribute
	a.SetKey(key)
	a.SetValue(val)

	require.Equal(t, key, a.Key())
	require.Equal(t, val, a.Value())

	var aV2 object.Attribute

	a.WriteToV2(&aV2)

	require.Equal(t, key, aV2.GetKey())
	require.Equal(t, val, aV2.GetValue())
}

func TestAttributeEncoding(t *testing.T) {
	var a Attribute
	a.SetKey("key")
	a.SetValue("value")

	t.Run("binary", func(t *testing.T) {
		data, err := a.Marshal()
		require.NoError(t, err)

		var a2 Attribute
		require.NoError(t, a2.Unmarshal(data))

		require.Equal(t, a, a2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := a.MarshalJSON()
		require.NoError(t, err)

		var a2 Attribute
		require.NoError(t, a2.UnmarshalJSON(data))

		require.Equal(t, a, a2)
	})
}

func TestNewAttributeFromV2(t *testing.T) {
	t.Run("from zero", func(t *testing.T) {
		var (
			v2 object.Attribute
			x  Attribute
		)

		x.ReadFromV2(v2)

		require.Empty(t, x.Key())
		require.Empty(t, x.Value())
	})
}

func TestAttribute_ToV2(t *testing.T) {
	t.Run("zero to V2", func(t *testing.T) {
		var (
			x  Attribute
			v2 object.Attribute
		)

		x.WriteToV2(&v2)

		require.Empty(t, v2.GetKey())
		require.Empty(t, v2.GetValue())
	})
}

func TestNewAttribute(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var a Attribute

		// check initial values
		require.Empty(t, a.Key())
		require.Empty(t, a.Value())
	})
}

func TestAttributes_Slice(t *testing.T) {
	var (
		aa Attributes

		a1 Attribute
		a2 Attribute
		a3 Attribute
		a4 Attribute
		a5 Attribute
	)

	require.Zero(t, aa.Len())

	a1.SetKey("key1")
	a1.SetValue("value1")
	a2.SetKey("key2")
	a2.SetValue("value2")
	a3.SetKey("key3")
	a3.SetValue("value3")
	a4.SetKey("key4")
	a4.SetValue("value4")
	a5.SetKey("key5")
	a5.SetValue("value5")

	aa.Append(a1)
	require.Equal(t, 1, aa.Len())

	aa.Append(a2, a3, a4, a5)
	require.Equal(t, 5, aa.Len())

	number := 1
	aa.Iterate(func(attribute Attribute) bool {
		require.Equal(t, fmt.Sprintf("key%d", number), attribute.Key())
		require.Equal(t, fmt.Sprintf("value%d", number), attribute.Value())

		number++

		return false
	})
}
