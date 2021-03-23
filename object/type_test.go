package object_test

import (
	"testing"

	objectv2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/stretchr/testify/require"
)

func TestType_ToV2(t *testing.T) {
	typs := []struct {
		t  object.Type
		t2 objectv2.Type
	}{
		{
			t:  object.TypeRegular,
			t2: objectv2.TypeRegular,
		},
		{
			t:  object.TypeTombstone,
			t2: objectv2.TypeTombstone,
		},
		{
			t:  object.TypeStorageGroup,
			t2: objectv2.TypeStorageGroup,
		},
		{
			t:  object.TypeLock,
			t2: objectv2.TypeLock,
		},
	}

	var t2 objectv2.Type

	for _, item := range typs {
		item.t.WriteToV2(&t2)

		require.Equal(t, item.t2, t2)

		var newItem object.Type
		newItem.ReadFromV2(item.t2)

		require.Equal(t, item.t, newItem)
	}
}

func TestType_String(t *testing.T) {
	toPtr := func(v object.Type) *object.Type {
		return &v
	}

	testEnumStrings(t, new(object.Type), []enumStringItem{
		{val: toPtr(object.TypeTombstone), str: "TOMBSTONE"},
		{val: toPtr(object.TypeStorageGroup), str: "STORAGE_GROUP"},
		{val: toPtr(object.TypeRegular), str: "REGULAR"},
		{val: toPtr(object.TypeLock), str: "LOCK"},
	})
}

type enumIface interface {
	Parse(string) bool
	String() string
}

type enumStringItem struct {
	val enumIface
	str string
}

func testEnumStrings(t *testing.T, e enumIface, items []enumStringItem) {
	for _, item := range items {
		require.Equal(t, item.str, item.val.String())

		s := item.val.String()

		require.True(t, e.Parse(s), s)

		require.EqualValues(t, item.val, e, item.val)
	}

	// incorrect strings
	for _, str := range []string{
		"some string",
		"undefined",
	} {
		require.False(t, e.Parse(str))
	}
}
