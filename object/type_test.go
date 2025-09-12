package object_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
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

func TestTypeProto(t *testing.T) {
	for x, y := range map[protoobject.ObjectType]object.Type{
		protoobject.ObjectType_REGULAR:       object.TypeRegular,
		protoobject.ObjectType_TOMBSTONE:     object.TypeTombstone,
		protoobject.ObjectType_STORAGE_GROUP: object.TypeStorageGroup,
		protoobject.ObjectType_LOCK:          object.TypeLock,
		protoobject.ObjectType_LINK:          object.TypeLink,
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
