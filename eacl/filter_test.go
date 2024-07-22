package eacl_test

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func TestFilter_ToV2(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		filter := eacl.NewFilter()

		// check initial values
		require.Empty(t, filter.Key())
		require.Empty(t, filter.Value())
		require.Zero(t, filter.From())
		require.Zero(t, filter.Matcher())

		// convert to v2 message
		filterV2 := filter.ToV2()

		require.Empty(t, filterV2.GetKey())
		require.Empty(t, filterV2.GetValue())
		require.Equal(t, acl.HeaderTypeUnknown, filterV2.GetHeaderType())
		require.Equal(t, acl.MatchTypeUnknown, filterV2.GetMatchType())
	})
}

func TestConstructFilter(t *testing.T) {
	h := eacl.FilterHeaderType(rand.Uint32())
	k := "Hello"
	m := eacl.Match(rand.Uint32())
	v := "World"
	f := eacl.ConstructFilter(h, k, m, v)
	require.Equal(t, h, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewObjectPropertyFilter(t *testing.T) {
	k := "Hello"
	m := eacl.Match(rand.Uint32())
	v := "World"
	f := eacl.NewObjectPropertyFilter(k, m, v)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewRequestHeaderFilter(t *testing.T) {
	k := "Hello"
	m := eacl.Match(rand.Uint32())
	v := "World"
	f := eacl.NewRequestHeaderFilter(k, m, v)
	require.Equal(t, eacl.HeaderFromRequest, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewCustomServiceFilter(t *testing.T) {
	k := "Hello"
	m := eacl.Match(rand.Uint32())
	v := "World"
	f := eacl.NewCustomServiceFilter(k, m, v)
	require.Equal(t, eacl.HeaderFromService, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
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
	m := eacl.Match(rand.Uint32())
	f := eacl.NewFilterObjectCreationEpochIs(m, epoch)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:creationEpoch", f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, "657984300249", f.Value())
}

func TestFilterObjectPayloadSizeIs(t *testing.T) {
	const sz = 4326750843582
	m := eacl.Match(rand.Uint32())
	f := eacl.NewFilterObjectPayloadSizeIs(m, sz)
	require.Equal(t, eacl.HeaderFromObject, f.From())
	require.Equal(t, "$Object:payloadLength", f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, "4326750843582", f.Value())
}
