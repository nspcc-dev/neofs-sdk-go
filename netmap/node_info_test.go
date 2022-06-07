package netmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeInfo_SetAttribute(t *testing.T) {
	var n NodeInfo

	const key = "some key"
	val := "some value"

	require.Zero(t, n.Attribute(val))

	n.SetAttribute(key, val)
	require.Equal(t, val, n.Attribute(key))

	val = "some other value"

	n.SetAttribute(key, val)
	require.Equal(t, val, n.Attribute(key))
}
