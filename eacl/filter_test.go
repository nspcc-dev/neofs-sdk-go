package eacl

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func newObjectFilter(match Match, key, val string) *Filter {
	return &Filter{
		from:    HeaderFromObject,
		key:     key,
		matcher: match,
		value:   staticStringer(val),
	}
}

func TestFilter(t *testing.T) {
	filter := newObjectFilter(MatchStringEqual, "some name", "200")

	v2 := filter.ToV2()
	require.NotNil(t, v2)
	require.Equal(t, v2acl.HeaderTypeObject, v2.GetHeaderType())
	require.EqualValues(t, v2acl.MatchTypeStringEqual, v2.GetMatchType())
	require.Equal(t, filter.Key(), v2.GetKey())
	require.Equal(t, filter.Value(), v2.GetValue())

	newFilter := NewFilterFromV2(v2)
	require.Equal(t, filter, newFilter)

	t.Run("from nil v2 filter", func(t *testing.T) {
		require.Equal(t, new(Filter), NewFilterFromV2(nil))
	})
}

func TestFilterEncoding(t *testing.T) {
	f := newObjectFilter(MatchStringEqual, "key", "value")

	t.Run("binary", func(t *testing.T) {
		f2 := NewFilter()
		require.NoError(t, f2.Unmarshal(f.Marshal()))

		require.Equal(t, f, f2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := f.MarshalJSON()
		require.NoError(t, err)

		d2 := NewFilter()
		require.NoError(t, d2.UnmarshalJSON(data))

		require.Equal(t, f, d2)
	})
}

func TestFilter_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *Filter

		require.Nil(t, x.ToV2())
	})

	t.Run("default values", func(t *testing.T) {
		filter := NewFilter()

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

func TestFilter_CopyTo(t *testing.T) {
	var filter Filter
	filter.value = staticStringer("value")
	filter.from = 1
	filter.matcher = 1
	filter.key = "1"

	var dst Filter
	t.Run("copy", func(t *testing.T) {
		filter.CopyTo(&dst)

		require.Equal(t, filter, dst)
		require.True(t, bytes.Equal(filter.Marshal(), dst.Marshal()))
	})

	t.Run("change", func(t *testing.T) {
		require.Equal(t, filter.value, dst.value)
		require.Equal(t, filter.from, dst.from)
		require.Equal(t, filter.matcher, dst.matcher)
		require.Equal(t, filter.key, dst.key)

		dst.value = staticStringer("value2")
		dst.from = 2
		dst.matcher = 2
		dst.key = "2"

		require.NotEqual(t, filter.value, dst.value)
		require.NotEqual(t, filter.from, dst.from)
		require.NotEqual(t, filter.matcher, dst.matcher)
		require.NotEqual(t, filter.key, dst.key)
	})
}

func TestConstructFilter(t *testing.T) {
	h := FilterHeaderType(rand.Uint32())
	k := "Hello"
	m := Match(rand.Uint32())
	v := "World"
	f := ConstructFilter(h, k, m, v)
	require.Equal(t, h, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewObjectPropertyFilter(t *testing.T) {
	k := "Hello"
	m := Match(rand.Uint32())
	v := "World"
	f := NewObjectPropertyFilter(k, m, v)
	require.Equal(t, HeaderFromObject, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewRequestHeaderFilter(t *testing.T) {
	k := "Hello"
	m := Match(rand.Uint32())
	v := "World"
	f := NewRequestHeaderFilter(k, m, v)
	require.Equal(t, HeaderFromRequest, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestNewCustomServiceFilter(t *testing.T) {
	k := "Hello"
	m := Match(rand.Uint32())
	v := "World"
	f := NewCustomServiceFilter(k, m, v)
	require.Equal(t, HeaderFromService, f.From())
	require.Equal(t, k, f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, v, f.Value())
}

func TestFilterSingleObject(t *testing.T) {
	obj := oid.ID{231, 189, 121, 7, 173, 134, 254, 165, 63, 186, 60, 89, 33, 95, 46, 103,
		217, 57, 164, 87, 82, 204, 251, 226, 1, 100, 32, 72, 251, 0, 7, 172}
	f := NewFilterObjectWithID(obj)
	require.Equal(t, HeaderFromObject, f.From())
	require.Equal(t, "$Object:objectID", f.Key())
	require.Equal(t, MatchStringEqual, f.Matcher())
	require.Equal(t, "GbckSBPEdM2P41Gkb9cVapFYb5HmRPDTZZp9JExGnsCF", f.Value())
}

func TestFilterObjectsFromContainer(t *testing.T) {
	cnr := cid.ID{231, 189, 121, 7, 173, 134, 254, 165, 63, 186, 60, 89, 33, 95, 46, 103,
		217, 57, 164, 87, 82, 204, 251, 226, 1, 100, 32, 72, 251, 0, 7, 172}
	f := NewFilterObjectsFromContainer(cnr)
	require.Equal(t, HeaderFromObject, f.From())
	require.Equal(t, "$Object:containerID", f.Key())
	require.Equal(t, MatchStringEqual, f.Matcher())
	require.Equal(t, "GbckSBPEdM2P41Gkb9cVapFYb5HmRPDTZZp9JExGnsCF", f.Value())
}

func TestFilterObjectOwnerEquals(t *testing.T) {
	owner := user.ID{53, 51, 5, 166, 111, 29, 20, 101, 192, 165, 28, 167, 57,
		160, 82, 80, 41, 203, 20, 254, 30, 138, 195, 17, 92}
	f := NewFilterObjectOwnerEquals(owner)
	require.Equal(t, HeaderFromObject, f.From())
	require.Equal(t, "$Object:ownerID", f.Key())
	require.Equal(t, MatchStringEqual, f.Matcher())
	require.Equal(t, "NQZkR7mG74rJsGAHnpkiFeU9c4f5VLN54f", f.Value())
}

func TestFilterObjectCreationEpochIs(t *testing.T) {
	const epoch = 657984300249
	m := Match(rand.Uint32())
	f := NewFilterObjectCreationEpochIs(m, epoch)
	require.Equal(t, HeaderFromObject, f.From())
	require.Equal(t, "$Object:creationEpoch", f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, "657984300249", f.Value())
}

func TestFilterObjectPayloadSizeIs(t *testing.T) {
	const sz = 4326750843582
	m := Match(rand.Uint32())
	f := NewFilterObjectPayloadSizeIs(m, sz)
	require.Equal(t, HeaderFromObject, f.From())
	require.Equal(t, "$Object:payloadLength", f.Key())
	require.Equal(t, m, f.Matcher())
	require.Equal(t, "4326750843582", f.Value())
}
