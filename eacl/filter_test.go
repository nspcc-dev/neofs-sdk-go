package eacl_test

import (
	"encoding/json"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func TestFilter_Marshal(t *testing.T) {
	for i := range anyValidFilters {
		require.Equal(t, anyValidBinFilters[i], anyValidFilters[i].Marshal(), i)
	}
}

func TestFilter_Unmarshal(t *testing.T) {
	t.Run("invalid protobuf", func(t *testing.T) {
		err := new(eacl.Filter).Unmarshal([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "cannot parse invalid wire-format data")
	})

	var f eacl.Filter
	for i := range anyValidBinFilters {
		require.NoError(t, f.Unmarshal(anyValidBinFilters[i]), i)
		require.EqualValues(t, anyValidFilters[i], f, i)
	}
}

func TestFilter_MarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := new(eacl.Filter).UnmarshalJSON([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var f1, f2 eacl.Filter
	for i := range anyValidFilters {
		b, err := anyValidFilters[i].MarshalJSON()
		require.NoError(t, err, i)
		require.NoError(t, f1.UnmarshalJSON(b), i)
		require.Equal(t, anyValidFilters[i], f1, i)

		b, err = json.Marshal(anyValidFilters[i])
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &f2), i)
		require.Equal(t, anyValidFilters[i], f2, i)
	}
}

func TestFilter_UnmarshalJSON(t *testing.T) {
	var f1, f2 eacl.Filter
	for i := range anyValidJSONFilters {
		require.NoError(t, f1.UnmarshalJSON([]byte(anyValidJSONFilters[i])), i)
		require.Equal(t, anyValidFilters[i], f1, i)

		require.NoError(t, json.Unmarshal([]byte(anyValidJSONFilters[i]), &f2), i)
		require.Equal(t, anyValidFilters[i], f2, i)
	}
}

func TestConstructFilter(t *testing.T) {
	k := "Hello"
	v := "World"
	f := eacl.ConstructFilter(anyValidHeaderType, k, anyValidMatcher, v)
	require.Equal(t, anyValidHeaderType, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, anyValidMatcher, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewObjectPropertyFilter(t *testing.T) {
	k := "Hello"
	v := "World"
	f := eacl.NewObjectPropertyFilter(k, anyValidMatcher, v)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, anyValidMatcher, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewRequestHeaderFilter(t *testing.T) {
	k := "Hello"
	v := "World"
	f := eacl.NewRequestHeaderFilter(k, anyValidMatcher, v)
	require.Equal(t, eacl.HeaderFromRequest, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, anyValidMatcher, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewCustomServiceFilter(t *testing.T) {
	k := "Hello"
	v := "World"
	f := eacl.NewCustomServiceFilter(k, anyValidMatcher, v)
	require.Equal(t, eacl.HeaderFromService, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, anyValidMatcher, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestFilterSingleObject(t *testing.T) {
	obj := oid.ID{231, 189, 121, 7, 173, 134, 254, 165, 63, 186, 60, 89, 33, 95, 46, 103,
		217, 57, 164, 87, 82, 204, 251, 226, 1, 100, 32, 72, 251, 0, 7, 172}
	f := eacl.NewFilterObjectWithID(obj)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:objectID", f.Key())
	require.Equal(t, eacl.MatchStringEqual, f.Matcher())
	require.Equal(t, "GbckSBPEdM2P41Gkb9cVapFYb5HmRPDTZZp9JExGnsCF", f.Value())
}

func TestFilterObjectsFromContainer(t *testing.T) {
	cnr := cid.ID{231, 189, 121, 7, 173, 134, 254, 165, 63, 186, 60, 89, 33, 95, 46, 103,
		217, 57, 164, 87, 82, 204, 251, 226, 1, 100, 32, 72, 251, 0, 7, 172}
	f := eacl.NewFilterObjectsFromContainer(cnr)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:containerID", f.Key())
	require.Equal(t, eacl.MatchStringEqual, f.Matcher())
	require.Equal(t, "GbckSBPEdM2P41Gkb9cVapFYb5HmRPDTZZp9JExGnsCF", f.Value())
}

func TestFilterObjectOwnerEquals(t *testing.T) {
	owner := user.ID{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57,
		160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
	f := eacl.NewFilterObjectOwnerEquals(owner)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:ownerID", f.Key())
	require.Equal(t, eacl.MatchStringEqual, f.Matcher())
	require.Equal(t, "NQZkR7mG74rJsGAHnpkiFeU9c4f5VLN54f", f.Value())
}

func TestFilterObjectCreationEpochIs(t *testing.T) {
	const epoch = 657984300249
	f := eacl.NewFilterObjectCreationEpochIs(anyValidMatcher, epoch)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:creationEpoch", f.Key())
	require.Equal(t, anyValidMatcher, f.Matcher())
	require.Equal(t, "657984300249", f.Value())
}

func TestFilterObjectPayloadSizeIs(t *testing.T) {
	const sz = 4326750843582
	f := eacl.NewFilterObjectPayloadSizeIs(anyValidMatcher, sz)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:payloadLength", f.Key())
	require.Equal(t, anyValidMatcher, f.Matcher())
	require.Equal(t, "4326750843582", f.Value())
}
