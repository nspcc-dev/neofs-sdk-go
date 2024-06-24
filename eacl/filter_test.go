package eacl_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	eacltest "github.com/nspcc-dev/neofs-sdk-go/eacl/test"
	"github.com/stretchr/testify/require"
)

func TestFilter_AttributeType(t *testing.T) {
	var f eacl.Filter
	require.Zero(t, f.AttributeType())

	f.SetAttributeType(13)
	require.EqualValues(t, 13, f.AttributeType())
	f.SetAttributeType(42)
	require.EqualValues(t, 42, f.AttributeType())
}

func TestFilter_Matcher(t *testing.T) {
	var f eacl.Filter
	require.Zero(t, f.Matcher())

	f.SetMatcher(13)
	require.EqualValues(t, 13, f.Matcher())
	f.SetMatcher(42)
	require.EqualValues(t, 42, f.Matcher())
}

func TestFilter_Key(t *testing.T) {
	var f eacl.Filter
	require.Zero(t, f.Key())

	for _, tc := range []struct {
		set, exp string
	}{
		{"any_key", "any_key"},
		{eacl.FilterObjectVersion, "$Object:version"},
		{eacl.FilterObjectID, "$Object:objectID"},
		{eacl.FilterObjectContainerID, "$Object:containerID"},
		{eacl.FilterObjectContainerID, "$Object:containerID"},
		{eacl.FilterObjectOwnerID, "$Object:ownerID"},
		{eacl.FilterObjectCreationEpoch, "$Object:creationEpoch"},
		{eacl.FilterObjectPayloadSize, "$Object:payloadLength"},
		{eacl.FilterObjectPayloadChecksum, "$Object:payloadHash"},
		{eacl.FilterObjectType, "$Object:objectType"},
		{eacl.FilterObjectPayloadHomomorphicChecksum, "$Object:homomorphicHash"},
	} {
		f.SetKey(tc.set)
		require.EqualValues(t, tc.exp, f.Key(), tc)
	}
}

func TestFilter_Value(t *testing.T) {
	var f eacl.Filter
	require.Zero(t, f.Value())

	f.SetValue("any_value")
	require.EqualValues(t, "any_value", f.Value())

	f.SetValue("other_value")
	require.EqualValues(t, "other_value", f.Value())
}

func TestFilter_CopyTo(t *testing.T) {
	src := eacltest.Filter()

	var dst eacl.Filter
	src.CopyTo(&dst)
	require.Equal(t, src, dst)

	originKey := src.Key()
	otherKey := originKey + "_extra"
	src.SetKey(otherKey)
	require.EqualValues(t, otherKey, src.Key())
	require.EqualValues(t, originKey, dst.Key())
}
