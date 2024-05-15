package objecttest_test

import (
	"testing"

	apiobject "github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func TestHeader(t *testing.T) {
	v := objecttest.Header()
	require.NotEqual(t, v, objecttest.Header())

	var v2 object.Header
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apiobject.Header
	v.WriteToV2(&m)
	var v3 object.Header
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v4 object.Header
	require.NoError(t, v4.UnmarshalJSON(j))
	require.Equal(t, v, v4)
}

func TestObject(t *testing.T) {
	v := objecttest.Object()
	require.NotEqual(t, v, objecttest.Object())

	var v2 object.Object
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apiobject.Object
	v.WriteToV2(&m)
	var v3 object.Object
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)

	j, err := v.MarshalJSON()
	require.NoError(t, err)
	var v4 object.Object
	require.NoError(t, v4.UnmarshalJSON(j))
	require.Equal(t, v, v4)
}

func TestTombstone(t *testing.T) {
	v := objecttest.Tombstone()
	require.NotEqual(t, v, objecttest.Tombstone())

	var v2 object.Tombstone
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)
}

func TestSplitInfo(t *testing.T) {
	v := objecttest.SplitInfo()
	require.NotEqual(t, v, objecttest.SplitInfo())

	var v2 object.SplitInfo
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)

	var m apiobject.SplitInfo
	v.WriteToV2(&m)
	var v3 object.SplitInfo
	require.NoError(t, v3.ReadFromV2(&m))
	require.Equal(t, v, v3)
}

func TestLock(t *testing.T) {
	v := objecttest.Lock()
	require.NotEqual(t, v, objecttest.Lock())

	var v2 object.Lock
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)
}

func TestSplitChain(t *testing.T) {
	v := objecttest.SplitChain()
	require.NotEqual(t, v, objecttest.SplitChain())

	var v2 object.SplitChain
	require.NoError(t, v2.Unmarshal(v.Marshal()))
	require.Equal(t, v, v2)
}
