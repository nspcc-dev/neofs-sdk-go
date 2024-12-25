package objecttest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/stretchr/testify/require"
)

func TestObject(t *testing.T) {
	obj := objecttest.Object()
	require.NotEqual(t, obj, objecttest.Object())

	var dst1 object.Object
	require.NoError(t, dst1.Unmarshal(obj.Marshal()))
	require.Equal(t, obj, dst1)

	var dst2 object.Object
	require.NoError(t, dst2.FromProtoMessage(obj.ProtoMessage()))
	require.Equal(t, obj, dst2)

	j, err := obj.MarshalJSON()
	require.NoError(t, err)
	var dst3 object.Object
	require.NoError(t, dst3.UnmarshalJSON(j))
	require.Equal(t, obj, dst3)
}
