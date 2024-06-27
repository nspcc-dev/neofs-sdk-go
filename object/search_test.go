package object_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	versiontest "github.com/nspcc-dev/neofs-sdk-go/version/test"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

var eqV2Matches = map[object.SearchMatchType]v2object.MatchType{
	object.MatchUnknown:        v2object.MatchUnknown,
	object.MatchStringEqual:    v2object.MatchStringEqual,
	object.MatchStringNotEqual: v2object.MatchStringNotEqual,
	object.MatchNotPresent:     v2object.MatchNotPresent,
	object.MatchCommonPrefix:   v2object.MatchCommonPrefix,
	object.MatchNumGT:          v2object.MatchNumGT,
	object.MatchNumGE:          v2object.MatchNumGE,
	object.MatchNumLT:          v2object.MatchNumLT,
	object.MatchNumLE:          v2object.MatchNumLE,
}

func TestMatch(t *testing.T) {
	t.Run("known matches", func(t *testing.T) {
		for matchType, matchTypeV2 := range eqV2Matches {
			require.Equal(t, matchTypeV2, matchType.ToV2())
			require.Equal(t, object.SearchMatchFromV2(matchTypeV2), matchType)
		}
	})

	t.Run("unknown matches", func(t *testing.T) {
		var unknownMatchType object.SearchMatchType

		for matchType := range eqV2Matches {
			unknownMatchType += matchType
		}

		unknownMatchType++

		require.Equal(t, unknownMatchType.ToV2(), v2object.MatchUnknown)

		var unknownMatchTypeV2 v2object.MatchType

		for _, matchTypeV2 := range eqV2Matches {
			unknownMatchTypeV2 += matchTypeV2
		}

		unknownMatchTypeV2++

		require.Equal(t, object.SearchMatchFromV2(unknownMatchTypeV2), object.MatchUnknown)
	})
}

func TestFilter(t *testing.T) {
	inputs := [][]string{
		{"user-header", "user-value"},
	}

	filters := object.NewSearchFilters()
	for i := range inputs {
		filters.AddFilter(inputs[i][0], inputs[i][1], object.MatchStringEqual)
	}

	require.Len(t, filters, len(inputs))
	for i := range inputs {
		require.Equal(t, inputs[i][0], filters[i].Header())
		require.Equal(t, inputs[i][1], filters[i].Value())
		require.Equal(t, object.MatchStringEqual, filters[i].Operation())
	}

	v2 := filters.ToV2()
	newFilters := object.NewSearchFiltersFromV2(v2)
	require.Equal(t, filters, newFilters)
}

func TestSearchFilters_AddRootFilter(t *testing.T) {
	fs := new(object.SearchFilters)

	fs.AddRootFilter()

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.MatchUnknown, f.Operation())
	require.Equal(t, object.FilterRoot, f.Header())
	require.Equal(t, "", f.Value())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterPropertyRoot, fsV2[0].GetKey())
		require.Equal(t, "", fsV2[0].GetValue())
		require.Equal(t, v2object.MatchUnknown, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddPhyFilter(t *testing.T) {
	fs := new(object.SearchFilters)

	fs.AddPhyFilter()

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.MatchUnknown, f.Operation())
	require.Equal(t, object.FilterPhysical, f.Header())
	require.Equal(t, "", f.Value())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterPropertyPhy, fsV2[0].GetKey())
		require.Equal(t, "", fsV2[0].GetValue())
		require.Equal(t, v2object.MatchUnknown, fsV2[0].GetMatchType())
	})
}

func testOID() oid.ID {
	cs := [sha256.Size]byte{}
	//nolint:staticcheck
	rand.Read(cs[:])

	var id oid.ID
	id.SetSHA256(cs)

	return id
}

func TestSearchFilters_AddParentIDFilter(t *testing.T) {
	par := testOID()

	fs := object.SearchFilters{}
	fs.AddParentIDFilter(object.MatchStringEqual, par)

	require.Len(t, fs, 1)

	f := fs[0]

	require.Equal(t, object.FilterParentID, f.Header())
	require.Equal(t, par.EncodeToString(), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderParent, fsV2[0].GetKey())
		require.Equal(t, par.EncodeToString(), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddObjectIDFilter(t *testing.T) {
	id := testOID()

	fs := new(object.SearchFilters)
	fs.AddObjectIDFilter(object.MatchStringEqual, id)

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.FilterID, f.Header())
	require.Equal(t, id.EncodeToString(), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderObjectID, fsV2[0].GetKey())
		require.Equal(t, id.EncodeToString(), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddSplitIDFilter(t *testing.T) {
	id := *object.NewSplitID()

	fs := new(object.SearchFilters)
	fs.AddSplitIDFilter(object.MatchStringEqual, id)

	f := (*fs)[0]

	require.Equal(t, object.FilterSplitID, f.Header())
	require.Equal(t, id.String(), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderSplitID, fsV2[0].GetKey())
		require.Equal(t, id.String(), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddFirstIDFilter(t *testing.T) {
	id := testOID()

	fs := new(object.SearchFilters)
	fs.AddFirstSplitObjectFilter(object.MatchStringEqual, id)

	f := (*fs)[0]

	require.Equal(t, object.FilterFirstSplitObject, f.Header())
	require.Equal(t, id.String(), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.ReservedFilterPrefix+"split.first", fsV2[0].GetKey())
		require.Equal(t, id.String(), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddTypeFilter(t *testing.T) {
	typ := object.TypeTombstone

	fs := new(object.SearchFilters)
	fs.AddTypeFilter(object.MatchStringEqual, typ)

	f := (*fs)[0]

	require.Equal(t, object.FilterType, f.Header())
	require.Equal(t, typ.String(), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderObjectType, fsV2[0].GetKey())
		require.Equal(t, typ.EncodeToString(), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFiltersEncoding(t *testing.T) {
	fs := object.NewSearchFilters()
	fs.AddFilter("key 1", "value 2", object.MatchStringEqual)
	fs.AddFilter("key 2", "value 2", object.MatchStringNotEqual)
	fs.AddFilter("key 2", "value 2", object.MatchCommonPrefix)

	t.Run("json", func(t *testing.T) {
		data, err := fs.MarshalJSON()
		require.NoError(t, err)

		fs2 := object.NewSearchFilters()
		require.NoError(t, fs2.UnmarshalJSON(data))

		require.Equal(t, fs, fs2)
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
		{val: toPtr(object.MatchNumGT), str: "NUM_GT"},
		{val: toPtr(object.MatchNumGE), str: "NUM_GE"},
		{val: toPtr(object.MatchNumLT), str: "NUM_LT"},
		{val: toPtr(object.MatchNumLE), str: "NUM_LE"},
	})
}

func testChecksumSha256() [sha256.Size]byte {
	cs := [sha256.Size]byte{}
	//nolint:staticcheck
	rand.Read(cs[:])

	return cs
}

func testChecksumTZ() [tz.Size]byte {
	cs := [tz.Size]byte{}
	//nolint:staticcheck
	rand.Read(cs[:])

	return cs
}

func TestSearchFilters_AddPayloadHashFilter(t *testing.T) {
	cs := testChecksumSha256()

	fs := new(object.SearchFilters)
	fs.AddPayloadHashFilter(object.MatchStringEqual, cs)

	f := (*fs)[0]

	require.Equal(t, object.FilterPayloadChecksum, f.Header())
	require.Equal(t, hex.EncodeToString(cs[:]), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderPayloadHash, fsV2[0].GetKey())
		require.Equal(t, hex.EncodeToString(cs[:]), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func ExampleSearchFilters_AddPayloadHashFilter() {
	hash, _ := hex.DecodeString("66842cfea090b1d906b52400fae49d86df078c0670f2bdd059ba289ebe24a498")

	var v [sha256.Size]byte
	copy(v[:], hash[:sha256.Size])

	var cs checksum.Checksum
	cs.SetSHA256(v)

	fmt.Println(hex.EncodeToString(cs.Value()))
	// Output: 66842cfea090b1d906b52400fae49d86df078c0670f2bdd059ba289ebe24a498
}

func TestSearchFilters_AddHomomorphicHashFilter(t *testing.T) {
	cs := testChecksumTZ()

	fs := new(object.SearchFilters)
	fs.AddHomomorphicHashFilter(object.MatchStringEqual, cs)

	f := (*fs)[0]

	require.Equal(t, object.FilterPayloadHomomorphicHash, f.Header())
	require.Equal(t, hex.EncodeToString(cs[:]), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderHomomorphicHash, fsV2[0].GetKey())
		require.Equal(t, hex.EncodeToString(cs[:]), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func ExampleSearchFilters_AddHomomorphicHashFilter() {
	hash, _ := hex.DecodeString("7e302ebb3937e810feb501965580c746048db99cebd095c3ce27022407408bf904dde8d9aa8085d2cf7202345341cc947fa9d722c6b6699760d307f653815d0c")

	var v [tz.Size]byte
	copy(v[:], hash[:tz.Size])

	var cs checksum.Checksum
	cs.SetTillichZemor(v)

	fmt.Println(hex.EncodeToString(cs.Value()))
	// Output: 7e302ebb3937e810feb501965580c746048db99cebd095c3ce27022407408bf904dde8d9aa8085d2cf7202345341cc947fa9d722c6b6699760d307f653815d0c
}

func TestSearchFilters_AddCreationEpochFilter(t *testing.T) {
	epoch := rand.Uint64()

	fs := new(object.SearchFilters)
	fs.AddCreationEpochFilter(object.MatchStringEqual, epoch)

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.FilterCreationEpoch, f.Header())
	require.Equal(t, strconv.FormatUint(epoch, 10), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderCreationEpoch, fsV2[0].GetKey())
		require.Equal(t, strconv.FormatUint(epoch, 10), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_AddPayloadSizeFilter(t *testing.T) {
	size := rand.Uint64()

	fs := new(object.SearchFilters)
	fs.AddPayloadSizeFilter(object.MatchStringEqual, size)

	require.Len(t, *fs, 1)

	f := (*fs)[0]

	require.Equal(t, object.FilterPayloadSize, f.Header())
	require.Equal(t, strconv.FormatUint(size, 10), f.Value())
	require.Equal(t, object.MatchStringEqual, f.Operation())

	t.Run("v2", func(t *testing.T) {
		fsV2 := fs.ToV2()

		require.Len(t, fsV2, 1)

		require.Equal(t, v2object.FilterHeaderPayloadLength, fsV2[0].GetKey())
		require.Equal(t, strconv.FormatUint(size, 10), fsV2[0].GetValue())
		require.Equal(t, v2object.MatchStringEqual, fsV2[0].GetMatchType())
	})
}

func TestSearchFilters_HasNonAttributeFilter(t *testing.T) {
	const anyMatcher = object.MatchStringEqual
	var fs object.SearchFilters

	fs.AddFilter("key", "value", anyMatcher)
	require.False(t, fs[0].IsNonAttribute())

	for _, f := range []func(){
		func() { fs.AddFilter("$Object:any", "", anyMatcher) },
		func() { fs.AddObjectVersionFilter(anyMatcher, versiontest.Version()) },
		func() { fs.AddParentIDFilter(anyMatcher, oidtest.ID()) },
		func() { fs.AddObjectContainerIDFilter(anyMatcher, cidtest.ID()) },
		func() { fs.AddObjectOwnerIDFilter(anyMatcher, usertest.ID()) },
		func() { fs.AddCreationEpochFilter(anyMatcher, rand.Uint64()) },
		func() { fs.AddPayloadSizeFilter(anyMatcher, rand.Uint64()) },
		func() { fs.AddPayloadHashFilter(anyMatcher, [sha256.Size]byte{1}) },
		func() { fs.AddTypeFilter(anyMatcher, object.TypeTombstone) },
		func() { fs.AddHomomorphicHashFilter(anyMatcher, [tz.Size]byte{1}) },
		func() { fs.AddParentIDFilter(anyMatcher, oidtest.ID()) },
		func() { fs.AddSplitIDFilter(anyMatcher, objecttest.SplitID()) },
		func() { fs.AddRootFilter() },
		func() { fs.AddPhyFilter() },
	} {
		fs = fs[:0]
		f()
		require.True(t, fs[0].IsNonAttribute())
	}
}
