package object_test

import (
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/stretchr/testify/require"
)

var typeStrings = map[object.Type]string{
	object.TypeRegular:      "REGULAR",
	object.TypeTombstone:    "TOMBSTONE",
	object.TypeStorageGroup: "STORAGE_GROUP",
	object.TypeLock:         "LOCK",
	object.TypeLink:         "LINK",
	5:                       "5",
}

func TestType_ToV2(t *testing.T) {
	typs := []struct {
		t  object.Type
		t2 v2object.Type
	}{
		{
			t:  object.TypeRegular,
			t2: v2object.TypeRegular,
		},
		{
			t:  object.TypeTombstone,
			t2: v2object.TypeTombstone,
		},
		{
			t:  object.TypeStorageGroup,
			t2: v2object.TypeStorageGroup,
		},
		{
			t:  object.TypeLock,
			t2: v2object.TypeLock,
		},
		{
			t:  object.TypeLink,
			t2: v2object.TypeLink,
		},
	}

	for _, item := range typs {
		t2 := item.t.ToV2()

		require.Equal(t, item.t2, t2)

		require.Equal(t, item.t, object.TypeFromV2(item.t2))
	}
}

func TestTypeProto(t *testing.T) {
	for x, y := range map[v2object.Type]object.Type{
		v2object.TypeRegular:      object.TypeRegular,
		v2object.TypeTombstone:    object.TypeTombstone,
		v2object.TypeStorageGroup: object.TypeStorageGroup,
		v2object.TypeLock:         object.TypeLock,
		v2object.TypeLink:         object.TypeLink,
	} {
		require.EqualValues(t, x, y)
	}
}

func TestType_String(t *testing.T) {
	for r, s := range typeStrings {
		require.Equal(t, s, r.String())
	}

	toPtr := func(v object.Type) *object.Type {
		return &v
	}

	testEnumStrings(t, new(object.Type), []enumStringItem{
		{val: toPtr(object.TypeTombstone), str: "TOMBSTONE"},
		{val: toPtr(object.TypeStorageGroup), str: "STORAGE_GROUP"},
		{val: toPtr(object.TypeRegular), str: "REGULAR"},
		{val: toPtr(object.TypeLock), str: "LOCK"},
		{val: toPtr(object.TypeLink), str: "LINK"},
	})
}

type enumIface interface {
	DecodeString(string) bool
	EncodeToString() string
}

type enumStringItem struct {
	val enumIface
	str string
}

func testEnumStrings(t *testing.T, e enumIface, items []enumStringItem) {
	for _, item := range items {
		require.Equal(t, item.str, item.val.EncodeToString())

		s := item.val.EncodeToString()

		require.True(t, e.DecodeString(s), s)

		require.EqualValues(t, item.val, e, item.val)
	}

	// incorrect strings
	for _, str := range []string{
		"some string",
		"undefined",
	} {
		require.False(t, e.DecodeString(str))
	}
}

func TestTypeToString(t *testing.T) {
	for n, s := range typeStrings {
		require.Equal(t, s, n.String())
	}
}

func TestTypeFromString(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		for _, s := range []string{"", "foo", "1.2"} {
			require.False(t, new(object.Type).DecodeString(s))
		}
	})
	var v object.Type
	for n, s := range typeStrings {
		require.True(t, v.DecodeString(s))
		require.Equal(t, n, v)
	}
}
