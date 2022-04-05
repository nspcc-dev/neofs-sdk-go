package container_test

import (
	"fmt"
	"testing"

	containerv2 "github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/stretchr/testify/require"
)

func TestAttribute(t *testing.T) {
	t.Run("zero to V2", func(t *testing.T) {
		var (
			x  container.Attribute
			v2 containerv2.Attribute
		)

		x.WriteToV2(&v2)

		require.Empty(t, v2.GetKey())
		require.Empty(t, v2.GetValue())
	})

	t.Run("default values", func(t *testing.T) {
		var attr container.Attribute

		// check initial values
		require.Empty(t, attr.Key())
		require.Empty(t, attr.Value())
	})

	const (
		key   = "key"
		value = "value"
	)

	var attr container.Attribute
	attr.SetKey(key)
	attr.SetValue(value)

	require.Equal(t, key, attr.Key())
	require.Equal(t, value, attr.Value())

	t.Run("test v2", func(t *testing.T) {
		const (
			newKey   = "newKey"
			newValue = "newValue"
		)

		var v2 containerv2.Attribute
		attr.WriteToV2(&v2)

		require.Equal(t, key, v2.GetKey())
		require.Equal(t, value, v2.GetValue())

		v2.SetKey(newKey)
		v2.SetValue(newValue)

		var newAttr container.Attribute
		newAttr.ReadFromV2(v2)

		require.Equal(t, newKey, newAttr.Key())
		require.Equal(t, newValue, newAttr.Value())
	})
}

func TestAttributes(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		var (
			x  container.Attributes
			v2 []containerv2.Attribute
		)

		x.WriteToV2(&v2)

		require.Empty(t, v2)
	})

	var (
		keys = []string{"key1", "key2", "key3"}
		vals = []string{"val1", "val2", "val3"}
	)

	attrs := make(container.Attributes, len(keys))

	for i := range keys {
		attrs[i].SetKey(keys[i])
		attrs[i].SetValue(vals[i])
	}

	t.Run("test v2", func(t *testing.T) {
		const postfix = "x"

		var v2 []containerv2.Attribute
		attrs.WriteToV2(&v2)

		require.Len(t, v2, len(keys))

		for i := range v2 {
			k := v2[i].GetKey()
			v := v2[i].GetValue()

			require.Equal(t, keys[i], k)
			require.Equal(t, vals[i], v)

			v2[i].SetKey(k + postfix)
			v2[i].SetValue(v + postfix)
		}

		var newAttrs container.Attributes
		newAttrs.ReadFromV2(v2)

		require.Len(t, newAttrs, len(keys))

		for i := range newAttrs {
			require.Equal(t, keys[i]+postfix, newAttrs[i].Key())
			require.Equal(t, vals[i]+postfix, newAttrs[i].Value())
		}
	})
}

func TestAttributes_Slice(t *testing.T) {
	var (
		aa container.Attributes

		a1 container.Attribute
		a2 container.Attribute
		a3 container.Attribute
		a4 container.Attribute
		a5 container.Attribute
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
	aa.Iterate(func(attribute container.Attribute) bool {
		require.Equal(t, fmt.Sprintf("key%d", number), attribute.Key())
		require.Equal(t, fmt.Sprintf("value%d", number), attribute.Value())

		number++

		return false
	})
}

func TestNewAttributeFromV2(t *testing.T) {
	t.Run("from zero V2", func(t *testing.T) {
		var (
			x  container.Attribute
			v2 containerv2.Attribute
		)

		x.ReadFromV2(v2)

		require.Empty(t, x.Key())
		require.Empty(t, x.Value())
	})
}

func TestGetNameWithZone(t *testing.T) {
	c := container.InitCreation()

	for _, item := range [...]struct {
		name, zone string
	}{
		{"name1", ""},
		{"name1", "zone1"},
		{"name2", "zone1"},
		{"name2", "zone2"},
		{"", "zone2"},
		{"", ""},
	} {
		container.SetNativeNameWithZone(&c, item.name, item.zone)

		name, zone := container.GetNativeNameWithZone(&c)

		require.Equal(t, item.name, name, item.name)
		require.Equal(t, item.zone, zone, item.zone)
	}
}

func TestSetNativeName(t *testing.T) {
	c := container.InitCreation()

	const nameDefZone = "some name"

	container.SetNativeName(&c, nameDefZone)

	name, zone := container.GetNativeNameWithZone(&c)

	require.Equal(t, nameDefZone, name)
	require.Equal(t, containerv2.SysAttributeZoneDefault, zone)
}
