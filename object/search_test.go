package object_test

import (
	"crypto/sha256"
	"math/rand"
	"testing"

	objectv2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

var eqV2Matches = map[object.SearchMatchType]objectv2.MatchType{
	object.MatchUnknown:        objectv2.MatchUnknown,
	object.MatchStringEqual:    objectv2.MatchStringEqual,
	object.MatchStringNotEqual: objectv2.MatchStringNotEqual,
	object.MatchNotPresent:     objectv2.MatchNotPresent,
	object.MatchCommonPrefix:   objectv2.MatchCommonPrefix,
}

func TestMatch(t *testing.T) {
	t.Run("known matches", func(t *testing.T) {
		var (
			x  object.SearchMatchType
			v2 objectv2.MatchType
		)

		for matchType, matchTypeV2 := range eqV2Matches {
			matchType.WriteToV2(&v2)
			x.ReadFromV2(matchTypeV2)

			require.Equal(t, matchTypeV2, v2)
			require.Equal(t, x, matchType)
		}
	})

	t.Run("unknown matches", func(t *testing.T) {
		var (
			unknownMatchType object.SearchMatchType
			v2               objectv2.MatchType
		)

		for matchType := range eqV2Matches {
			unknownMatchType += matchType
		}

		unknownMatchType++
		unknownMatchType.WriteToV2(&v2)

		require.Equal(t, v2, objectv2.MatchUnknown)

		var unknownMatchTypeV2 objectv2.MatchType

		for _, matchTypeV2 := range eqV2Matches {
			unknownMatchTypeV2 += matchTypeV2
		}

		unknownMatchTypeV2++
		unknownMatchType.ReadFromV2(unknownMatchTypeV2)

		require.Equal(t, unknownMatchType, object.MatchUnknown)
	})
}

func TestFilter(t *testing.T) {
	inputs := [][]string{
		{"user-header", "user-value"},
	}

	var filters object.SearchFilters
	for i := range inputs {
		filters.AddFilter(inputs[i][0], inputs[i][1], object.MatchStringEqual)
	}

	require.Len(t, filters, len(inputs))
	for i := range inputs {
		require.Equal(t, inputs[i][0], filters[i].Header())
		require.Equal(t, inputs[i][1], filters[i].Value())
		require.Equal(t, object.MatchStringEqual, filters[i].Operation())
	}

	v2 := make([]objectv2.SearchFilter, 0)
	filters.WriteToV2(&v2)

	var newFilters object.SearchFilters
	newFilters.ReadFromV2(v2)

	require.Equal(t, filters, newFilters)
}

func TestSearchFilters_AddRootFilter(t *testing.T) {
	fs := new(object.SearchFilters)

	fs.AddRootFilter()

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.MatchUnknown, f.Operation())
	require.Equal(t, objectv2.FilterPropertyRoot, f.Header())
	require.Equal(t, "", f.Value())
}

func TestSearchFilters_AddPhyFilter(t *testing.T) {
	fs := new(object.SearchFilters)

	fs.AddPhyFilter()

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.MatchUnknown, f.Operation())
	require.Equal(t, objectv2.FilterPropertyPhy, f.Header())
	require.Equal(t, "", f.Value())
}

func testOID() *oid.ID {
	cs := [sha256.Size]byte{}

	rand.Read(cs[:])

	var id oid.ID
	id.SetSHA256(cs)

	return &id
}

func TestSearchFilters_AddParentIDFilter(t *testing.T) {
	par := testOID()

	fs := object.SearchFilters{}
	fs.AddParentIDFilter(object.MatchStringEqual, par)

	fsV2 := make([]objectv2.SearchFilter, 0)
	fs.WriteToV2(&fsV2)

	require.Len(t, fsV2, 1)

	require.Equal(t, objectv2.FilterHeaderParent, fsV2[0].GetKey())
	require.Equal(t, par.String(), fsV2[0].GetValue())
	require.Equal(t, objectv2.MatchStringEqual, fsV2[0].GetMatchType())
}

func TestSearchFilters_AddObjectIDFilter(t *testing.T) {
	id := testOID()

	fs := new(object.SearchFilters)
	fs.AddObjectIDFilter(object.MatchStringEqual, id)

	t.Run("v2", func(t *testing.T) {
		fsV2 := make([]objectv2.SearchFilter, 0)
		fs.WriteToV2(&fsV2)

		require.Len(t, fsV2, 1)

		require.Equal(t, objectv2.FilterHeaderObjectID, fsV2[0].GetKey())
		require.Equal(t, id.String(), fsV2[0].GetValue())
		require.Equal(t, objectv2.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddSplitIDFilter(t *testing.T) {
	id := object.NewSplitID()

	fs := new(object.SearchFilters)
	fs.AddSplitIDFilter(object.MatchStringEqual, id)

	t.Run("v2", func(t *testing.T) {
		fsV2 := make([]objectv2.SearchFilter, 0)
		fs.WriteToV2(&fsV2)

		require.Len(t, fsV2, 1)

		require.Equal(t, objectv2.FilterHeaderSplitID, fsV2[0].GetKey())
		require.Equal(t, id.String(), fsV2[0].GetValue())
		require.Equal(t, objectv2.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddTypeFilter(t *testing.T) {
	typ := object.TypeTombstone

	fs := new(object.SearchFilters)
	fs.AddTypeFilter(object.MatchStringEqual, typ)

	t.Run("v2", func(t *testing.T) {
		fsV2 := make([]objectv2.SearchFilter, 0)
		fs.WriteToV2(&fsV2)

		require.Len(t, fsV2, 1)

		require.Equal(t, objectv2.FilterHeaderObjectType, fsV2[0].GetKey())
		require.Equal(t, typ.String(), fsV2[0].GetValue())
		require.Equal(t, objectv2.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchMatchType_String(t *testing.T) {
	toPtr := func(v object.SearchMatchType) *object.SearchMatchType {
		return &v
	}

	testEnumStrings(t, new(object.SearchMatchType), []enumStringItem{
		{val: toPtr(object.MatchCommonPrefix), str: "COMMON_PREFIX"},
		{val: toPtr(object.MatchStringEqual), str: "STRING_EQUAL"},
		{val: toPtr(object.MatchStringNotEqual), str: "STRING_NOT_EQUAL"},
		{val: toPtr(object.MatchNotPresent), str: "NOT_PRESENT"},
		{val: toPtr(object.MatchUnknown), str: "MATCH_TYPE_UNSPECIFIED"},
	})
}
