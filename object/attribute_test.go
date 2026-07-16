package object_test

import (
	"strconv"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/stretchr/testify/require"
)

func TestNewAttribute(t *testing.T) {
	key, val := "some key", "some value"

	a := object.NewAttribute(key, val)

	require.Equal(t, key, a.Key())
	require.Equal(t, val, a.Value())
}

func TestAttribute_Marshal(t *testing.T) {
	// TODO
}

func TestAttribute_Unmarshal(t *testing.T) {
	// TODO
}

func TestAttribute_MarshalJSON(t *testing.T) {
	// TODO
}

func TestAttribute_UnmarshalJSON(t *testing.T) {
	// TODO
}

func TestAttribute_SetKey(t *testing.T) {
	var a object.Attribute
	require.Zero(t, a.Key())

	const key = "key"
	a.SetKey(key)
	require.Equal(t, key, a.Key())

	const otherKey = key + "_other"
	a.SetKey(otherKey)
	require.Equal(t, otherKey, a.Key())
}

func TestAttribute_SetValue(t *testing.T) {
	var a object.Attribute
	require.Zero(t, a.Value())

	const val = "key"
	a.SetValue(val)
	require.Equal(t, val, a.Value())

	const otherVal = val + "_other"
	a.SetKey(otherVal)
	require.Equal(t, otherVal, a.Key())
}

func TestSystemAttributes(t *testing.T) {
	for _, tc := range []struct {
		cnst, exp string
	}{
		{cnst: object.AttributeExpirationEpoch, exp: "__NEOFS__EXPIRATION_EPOCH"},
		{cnst: object.AttributeAssociatedObject, exp: "__NEOFS__ASSOCIATE"},
		{cnst: object.AttributeAssociatedObject, exp: "__NEOFS__ASSOCIATE"},
		{cnst: object.AttributeECRuleIndex, exp: "__NEOFS__EC_RULE_IDX"},
		{cnst: object.AttributeECPartIndex, exp: "__NEOFS__EC_PART_IDX"},
	} {
		t.Run(tc.exp, func(t *testing.T) {
			require.Equal(t, tc.exp, tc.cnst)
		})
	}
}

func TestGetIndexAttribute(t *testing.T) {
	var obj object.Object
	const attr = "attr"

	t.Run("missing", func(t *testing.T) {
		i, err := object.GetIndexAttribute(obj, attr)
		require.NoError(t, err)
		require.EqualValues(t, -1, i)
	})

	t.Run("not an integer", func(t *testing.T) {
		obj.SetAttributes(object.NewAttribute(attr, "not_an_int"))

		_, err := object.GetIndexAttribute(obj, attr)
		require.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("negative", func(t *testing.T) {
		obj.SetAttributes(object.NewAttribute(attr, "-123"))

		_, err := object.GetIndexAttribute(obj, attr)
		require.EqualError(t, err, "negative value -123")
	})

	obj.SetAttributes(object.NewAttribute(attr, "1234567890"))

	i, err := object.GetIndexAttribute(obj, attr)
	require.NoError(t, err)
	require.EqualValues(t, 1234567890, i)

	t.Run("multiple", func(t *testing.T) {
		for _, s := range []string{
			"not_an_int",
			"-1",
			"2",
		} {
			obj.SetAttributes(
				object.NewAttribute(attr, "1"),
				object.NewAttribute(attr, s),
			)

			i, err := object.GetIndexAttribute(obj, attr)
			require.NoError(t, err)
			require.EqualValues(t, 1, i)
		}
	})
}

func TestGetIntAttribute(t *testing.T) {
	var obj object.Object
	const attr = "attr"

	t.Run("missing", func(t *testing.T) {
		_, err := object.GetIntAttribute(obj, attr)
		require.ErrorIs(t, err, object.ErrAttributeNotFound)
	})

	t.Run("not an integer", func(t *testing.T) {
		obj.SetAttributes(object.NewAttribute(attr, "not_an_int"))

		_, err := object.GetIntAttribute(obj, attr)
		require.ErrorContains(t, err, "invalid syntax")
	})

	for _, tc := range []struct {
		s string
		i int
	}{
		{s: "1234567890", i: 1234567890},
		{s: "0", i: 0},
		{s: "-1234567890", i: -1234567890},
	} {
		obj.SetAttributes(object.NewAttribute(attr, tc.s))

		i, err := object.GetIntAttribute(obj, attr)
		require.NoError(t, err, tc.s)
		require.EqualValues(t, tc.i, i)
	}

	t.Run("multiple", func(t *testing.T) {
		for _, s := range []string{
			"not_an_int",
			"-1",
			"2",
		} {
			obj.SetAttributes(
				object.NewAttribute(attr, "1"),
				object.NewAttribute(attr, s),
			)

			i, err := object.GetIntAttribute(obj, attr)
			require.NoError(t, err)
			require.EqualValues(t, 1, i)
		}
	})
}

func TestGetAttribute(t *testing.T) {
	var obj object.Object
	const attr = "attr"

	t.Run("missing", func(t *testing.T) {
		require.Empty(t, object.GetAttribute(obj, attr))
	})

	obj.SetAttributes(object.NewAttribute(attr, "val"))
	require.Equal(t, "val", object.GetAttribute(obj, attr))

	t.Run("multiple", func(t *testing.T) {
		obj.SetAttributes(
			object.NewAttribute(attr, "val1"),
			object.NewAttribute(attr, "val2"),
		)

		require.Equal(t, "val1", object.GetAttribute(obj, attr))
	})
}

func TestSetIntAttribute(t *testing.T) {
	var obj object.Object
	const attr = "attr"

	obj.SetAttributes(object.NewAttribute(attr+"_other", "val"))

	check := func(t *testing.T, val int) {
		object.SetIntAttribute(&obj, attr, val)

		attrs := obj.Attributes()
		require.Len(t, attrs, 2)
		require.Equal(t, attr, attrs[1].Key())
		require.Equal(t, strconv.Itoa(val), attrs[1].Value())

		got, err := object.GetIntAttribute(obj, attr)
		require.NoError(t, err, val)
		require.EqualValues(t, val, got)
	}

	check(t, 1234567890)
	check(t, 0)
	check(t, -1234567890)
}

func TestSetAttribute(t *testing.T) {
	var obj object.Object
	const attr = "attr"

	obj.SetAttributes(object.NewAttribute(attr+"_other", "val"))

	check := func(t *testing.T, val string) {
		object.SetAttribute(&obj, attr, val)

		attrs := obj.Attributes()
		require.Len(t, attrs, 2)
		require.Equal(t, attr, attrs[1].Key())
		require.Equal(t, val, attrs[1].Value())

		got := object.GetAttribute(obj, attr)
		require.Equal(t, val, got)
	}

	check(t, "val1")
	check(t, "val2")
}
