package object_test

import (
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
