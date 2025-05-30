package object_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

const (
	anyValidSearchMatcher = object.SearchMatchType(1937803447)
)

var protoMatches = map[object.SearchMatchType]protoobject.MatchType{
	object.MatchUnspecified:    protoobject.MatchType_MATCH_TYPE_UNSPECIFIED,
	object.MatchStringEqual:    protoobject.MatchType_STRING_EQUAL,
	object.MatchStringNotEqual: protoobject.MatchType_STRING_NOT_EQUAL,
	object.MatchNotPresent:     protoobject.MatchType_NOT_PRESENT,
	object.MatchCommonPrefix:   protoobject.MatchType_COMMON_PREFIX,
	object.MatchNumGT:          protoobject.MatchType_NUM_GT,
	object.MatchNumGE:          protoobject.MatchType_NUM_GE,
	object.MatchNumLT:          protoobject.MatchType_NUM_LT,
	object.MatchNumLE:          protoobject.MatchType_NUM_LE,
}

var searchMatchTypeStrings = map[object.SearchMatchType]string{
	0:                          "MATCH_TYPE_UNSPECIFIED",
	object.MatchStringEqual:    "STRING_EQUAL",
	object.MatchStringNotEqual: "STRING_NOT_EQUAL",
	object.MatchNotPresent:     "NOT_PRESENT",
	object.MatchCommonPrefix:   "COMMON_PREFIX",
	object.MatchNumGT:          "NUM_GT",
	object.MatchNumGE:          "NUM_GE",
	object.MatchNumLT:          "NUM_LT",
	object.MatchNumLE:          "NUM_LE",
	9:                          "9",
}

var validSearchFilters object.SearchFilters // set by init.

func init() {
	validSearchFilters.AddPhyFilter()
	validSearchFilters.AddRootFilter()
	validSearchFilters.AddFilter("k1", "v1", object.MatchStringEqual)
	validSearchFilters.AddFilter("k2", "v2", object.MatchStringNotEqual)
	validSearchFilters.AddFilter("k3", "v3", object.MatchNotPresent)
	validSearchFilters.AddFilter("k4", "v4", object.MatchCommonPrefix)
	validSearchFilters.AddFilter("k5", "v5", object.MatchNumGT)
	validSearchFilters.AddFilter("k6", "v6", object.MatchNumGE)
	validSearchFilters.AddFilter("k7", "v7", object.MatchNumLT)
	validSearchFilters.AddFilter("k8", "v8", object.MatchNumLE)
	validSearchFilters.AddObjectVersionFilter(100, anyValidVersions[0])
	validSearchFilters.AddObjectIDFilter(101, anyValidIDs[0])
	validSearchFilters.AddObjectContainerIDFilter(102, anyValidContainers[0])
	validSearchFilters.AddObjectOwnerIDFilter(103, anyValidUsers[0])
	validSearchFilters.AddCreationEpochFilter(104, anyValidCreationEpoch)
	validSearchFilters.AddPayloadSizeFilter(105, anyValidPayloadSize)
	validSearchFilters.AddPayloadHashFilter(106, anySHA256Hash)
	validSearchFilters.AddTypeFilter(107, anyValidType)
	validSearchFilters.AddHomomorphicHashFilter(108, anyTillichZemorHash)
	validSearchFilters.AddParentIDFilter(109, anyValidIDs[1])
	validSearchFilters.AddSplitIDFilter(110, *anyValidSplitID)
	validSearchFilters.AddFirstSplitObjectFilter(111, anyValidIDs[2])
}

// corresponds to validSearchFilters.
var validSearchFiltersProto = []struct {
	k string
	m protoobject.MatchType
	v string
}{
	{"$Object:PHY", protoobject.MatchType_MATCH_TYPE_UNSPECIFIED, ""},
	{"$Object:ROOT", protoobject.MatchType_MATCH_TYPE_UNSPECIFIED, ""},
	{"k1", protoobject.MatchType_STRING_EQUAL, "v1"},
	{"k2", protoobject.MatchType_STRING_NOT_EQUAL, "v2"},
	{"k3", protoobject.MatchType_NOT_PRESENT, "v3"},
	{"k4", protoobject.MatchType_COMMON_PREFIX, "v4"},
	{"k5", protoobject.MatchType_NUM_GT, "v5"},
	{"k6", protoobject.MatchType_NUM_GE, "v6"},
	{"k7", protoobject.MatchType_NUM_LT, "v7"},
	{"k8", protoobject.MatchType_NUM_LE, "v8"},
	{"$Object:version", 100, "v88789927.2018985309"},
	{"$Object:objectID", 101, "CzyDjRYWpwLHxqXVFBXKQGP5XM7ebAR9ndTvBdaSxMMV"},
	{"$Object:containerID", 102, "HWpbBkyxCi7nhDnn4W3v5rYt2mDfH2wedknQzRkTwquj"},
	{"$Object:ownerID", 103, "NRJF3hqhZAe4NeWpABW6Q3ajkfhFUkY2ek"},
	{"$Object:creationEpoch", 104, "13233261290750647837"},
	{"$Object:payloadLength", 105, "5544264194415343420"},
	{"$Object:payloadHash", 106, "e9cc25bd0f92d28ab24ad58dc7f95e144985109af19803cd65d2998d8b1ed87d"},
	{"$Object:objectType", 107, "2082391263"},
	{"$Object:homomorphicHash", 108, "a09506a729461d3dbe9a1e75b4965b921810c3d5d86a77cbb29f2501fcd05717a5131660321c91eb7f6b56d833e254f25eba5a51b8ec76413a456ee816f983ad"},
	{"$Object:split.parent", 109, "GS6gbvdEKHZQf3EZRYd3fSdep3HAqEueJ1jz7iZDuoBG"},
	{"$Object:split.splitID", 110, "e0840350-202c-45b8-b920-e2c9cec49329"},
	{"$Object:split.first", 111, "EvdVD3NXuxFK1PWYSsvZwcbeqsv42SQV39TwXgs6jUQH"},
}

// corresponds to validSearchFilters.
var validJSONSearchFilters = `
[
 {
  "matchType": "MATCH_TYPE_UNSPECIFIED",
  "key": "$Object:PHY",
  "value": ""
 },
 {
  "matchType": "MATCH_TYPE_UNSPECIFIED",
  "key": "$Object:ROOT",
  "value": ""
 },
 {
  "matchType": "STRING_EQUAL",
  "key": "k1",
  "value": "v1"
 },
 {
  "matchType": "STRING_NOT_EQUAL",
  "key": "k2",
  "value": "v2"
 },
 {
  "matchType": "NOT_PRESENT",
  "key": "k3",
  "value": "v3"
 },
 {
  "matchType": "COMMON_PREFIX",
  "key": "k4",
  "value": "v4"
 },
 {
  "matchType": "NUM_GT",
  "key": "k5",
  "value": "v5"
 },
 {
  "matchType": "NUM_GE",
  "key": "k6",
  "value": "v6"
 },
 {
  "matchType": "NUM_LT",
  "key": "k7",
  "value": "v7"
 },
 {
  "matchType": "NUM_LE",
  "key": "k8",
  "value": "v8"
 },
 {
  "matchType": 100,
  "key": "$Object:version",
  "value": "v88789927.2018985309"
 },
 {
  "matchType": 101,
  "key": "$Object:objectID",
  "value": "CzyDjRYWpwLHxqXVFBXKQGP5XM7ebAR9ndTvBdaSxMMV"
 },
 {
  "matchType": 102,
  "key": "$Object:containerID",
  "value": "HWpbBkyxCi7nhDnn4W3v5rYt2mDfH2wedknQzRkTwquj"
 },
 {
  "matchType": 103,
  "key": "$Object:ownerID",
  "value": "NRJF3hqhZAe4NeWpABW6Q3ajkfhFUkY2ek"
 },
 {
  "matchType": 104,
  "key": "$Object:creationEpoch",
  "value": "13233261290750647837"
 },
 {
  "matchType": 105,
  "key": "$Object:payloadLength",
  "value": "5544264194415343420"
 },
 {
  "matchType": 106,
  "key": "$Object:payloadHash",
  "value": "e9cc25bd0f92d28ab24ad58dc7f95e144985109af19803cd65d2998d8b1ed87d"
 },
 {
  "matchType": 107,
  "key": "$Object:objectType",
  "value": "2082391263"
 },
 {
  "matchType": 108,
  "key": "$Object:homomorphicHash",
  "value": "a09506a729461d3dbe9a1e75b4965b921810c3d5d86a77cbb29f2501fcd05717a5131660321c91eb7f6b56d833e254f25eba5a51b8ec76413a456ee816f983ad"
 },
 {
  "matchType": 109,
  "key": "$Object:split.parent",
  "value": "GS6gbvdEKHZQf3EZRYd3fSdep3HAqEueJ1jz7iZDuoBG"
 },
 {
  "matchType": 110,
  "key": "$Object:split.splitID",
  "value": "e0840350-202c-45b8-b920-e2c9cec49329"
 },
 {
  "matchType": 111,
  "key": "$Object:split.first",
  "value": "EvdVD3NXuxFK1PWYSsvZwcbeqsv42SQV39TwXgs6jUQH"
 }
]
`

func assertSearchFilter(t testing.TB, fs object.SearchFilters, i int, k string, m object.SearchMatchType, v string, isSys bool) {
	require.Equal(t, k, fs[i].Header())
	require.Equal(t, m, fs[i].Operation())
	require.Equal(t, v, fs[i].Value())
	require.Equal(t, isSys, fs[i].IsNonAttribute())
}

func TestSearchFilters_MarshalJSON(t *testing.T) {
	b, err := validSearchFilters.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, validJSONSearchFilters, string(b))
}

func TestSearchFilters_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(object.SearchFilters).UnmarshalJSON([]byte("Hello, world!"))
			e := new(json.SyntaxError)
			require.ErrorAs(t, err, &e)
		})
	})

	var fs object.SearchFilters
	// empty
	require.NoError(t, fs.UnmarshalJSON([]byte("[]")))
	require.Empty(t, fs)

	// filled
	require.NoError(t, fs.UnmarshalJSON([]byte(validJSONSearchFilters)))
	require.Equal(t, validSearchFilters, fs)
}

func TestSearchFilters_AddFilter(t *testing.T) {
	const k1, m1, v1 = "k1", object.SearchMatchType(584744206), "v1"
	const k2, m2, v2 = "k2", object.SearchMatchType(930572326), "v2"

	var filters object.SearchFilters
	filters.AddFilter(k1, v1, m1)
	filters.AddFilter(k2, v2, m2)

	require.Len(t, filters, 2)
	assertSearchFilter(t, filters, 0, k1, m1, v1, false)
	assertSearchFilter(t, filters, 1, k2, m2, v2, false)
}

func TestSearchFilters_AddObjectVersionFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddObjectVersionFilter(anyValidSearchMatcher, anyValidVersions[0])
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterVersion, anyValidSearchMatcher, "v88789927.2018985309", true)
}

func TestSearchFilters_AddObjectContainerIDFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddObjectContainerIDFilter(anyValidSearchMatcher, anyValidContainers[0])
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterContainerID, anyValidSearchMatcher,
		"HWpbBkyxCi7nhDnn4W3v5rYt2mDfH2wedknQzRkTwquj", true)
}

func TestSearchFilters_AddObjectOwnerIDFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddObjectOwnerIDFilter(anyValidSearchMatcher, anyValidUsers[0])
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterOwnerID, anyValidSearchMatcher,
		"NRJF3hqhZAe4NeWpABW6Q3ajkfhFUkY2ek", true)
}

func TestSearchFilters_AddRootFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddRootFilter()
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterRoot, 0, "", true)
}

func TestSearchFilters_AddPhyFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddPhyFilter()
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterPhysical, 0, "", true)
}

func TestSearchFilters_AddParentIDFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddParentIDFilter(anyValidSearchMatcher, anyValidIDs[0])
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterParentID, anyValidSearchMatcher,
		"CzyDjRYWpwLHxqXVFBXKQGP5XM7ebAR9ndTvBdaSxMMV", true)
}

func TestSearchFilters_AddObjectIDFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddObjectIDFilter(anyValidSearchMatcher, anyValidIDs[0])
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterID, anyValidSearchMatcher,
		"CzyDjRYWpwLHxqXVFBXKQGP5XM7ebAR9ndTvBdaSxMMV", true)
}

func TestSearchFilters_AddSplitIDFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddSplitIDFilter(anyValidSearchMatcher, *anyValidSplitID)
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterSplitID, anyValidSearchMatcher,
		"e0840350-202c-45b8-b920-e2c9cec49329", true)
}

func TestSearchFilters_AddFirstSplitObjectFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddFirstSplitObjectFilter(anyValidSearchMatcher, anyValidIDs[0])
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterFirstSplitObject, anyValidSearchMatcher,
		"CzyDjRYWpwLHxqXVFBXKQGP5XM7ebAR9ndTvBdaSxMMV", true)
}

func TestSearchFilters_AddTypeFilter(t *testing.T) {
	var fs object.SearchFilters
	var i int
	for typ, s := range typeStrings {
		fs.AddTypeFilter(anyValidSearchMatcher, typ)
		require.Len(t, fs, i+1)
		assertSearchFilter(t, fs, i, object.FilterType, anyValidSearchMatcher, s, true)
		i++
	}
}

func TestSearchMatchTypeProto(t *testing.T) {
	for x, y := range protoMatches {
		require.EqualValues(t, x, y)
	}
}

func TestSearchMatchType_String(t *testing.T) {
	for r, s := range searchMatchTypeStrings {
		require.Equal(t, s, r.String())
	}

	toPtr := func(v object.SearchMatchType) *object.SearchMatchType {
		return &v
	}

	testEnumStrings(t, new(object.SearchMatchType), []enumStringItem{
		{val: toPtr(object.MatchCommonPrefix), str: "COMMON_PREFIX"},
		{val: toPtr(object.MatchStringEqual), str: "STRING_EQUAL"},
		{val: toPtr(object.MatchStringNotEqual), str: "STRING_NOT_EQUAL"},
		{val: toPtr(object.MatchNotPresent), str: "NOT_PRESENT"},
		{val: toPtr(object.MatchUnspecified), str: "MATCH_TYPE_UNSPECIFIED"},
		{val: toPtr(object.MatchNumGT), str: "NUM_GT"},
		{val: toPtr(object.MatchNumGE), str: "NUM_GE"},
		{val: toPtr(object.MatchNumLT), str: "NUM_LT"},
		{val: toPtr(object.MatchNumLE), str: "NUM_LE"},
	})
}

func TestSearchMatchTypeToString(t *testing.T) {
	for n, s := range searchMatchTypeStrings {
		require.Equal(t, s, n.String())
	}
}

func TestSearchMatchTypeFromString(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		for _, s := range []string{"", "foo", "1.2"} {
			require.False(t, new(object.SearchMatchType).DecodeString(s))
		}
	})
	var v object.SearchMatchType
	for n, s := range searchMatchTypeStrings {
		require.True(t, v.DecodeString(s))
		require.Equal(t, n, v)
	}
}

func TestSearchFilters_AddPayloadHashFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddPayloadHashFilter(anyValidSearchMatcher, anySHA256Hash)
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterPayloadChecksum, anyValidSearchMatcher,
		"e9cc25bd0f92d28ab24ad58dc7f95e144985109af19803cd65d2998d8b1ed87d", true)
}

func ExampleSearchFilters_AddPayloadHashFilter() {
	hash, _ := hex.DecodeString("66842cfea090b1d906b52400fae49d86df078c0670f2bdd059ba289ebe24a498")
	cs := checksum.NewSHA256([sha256.Size]byte(hash))
	fmt.Println(hex.EncodeToString(cs.Value()))
	// Output: 66842cfea090b1d906b52400fae49d86df078c0670f2bdd059ba289ebe24a498
}

func TestSearchFilters_AddHomomorphicHashFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddHomomorphicHashFilter(anyValidSearchMatcher, anyTillichZemorHash)
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterPayloadHomomorphicHash, anyValidSearchMatcher,
		"a09506a729461d3dbe9a1e75b4965b921810c3d5d86a77cbb29f2501fcd05717a5131660321c91eb7f6b56d833e254f25eba5a51b8ec76413a456ee816f983ad",
		true)
}

func ExampleSearchFilters_AddHomomorphicHashFilter() {
	hash, _ := hex.DecodeString("7e302ebb3937e810feb501965580c746048db99cebd095c3ce27022407408bf904dde8d9aa8085d2cf7202345341cc947fa9d722c6b6699760d307f653815d0c")
	cs := checksum.NewTillichZemor([tz.Size]byte(hash))
	fmt.Println(hex.EncodeToString(cs.Value()))
	// Output: 7e302ebb3937e810feb501965580c746048db99cebd095c3ce27022407408bf904dde8d9aa8085d2cf7202345341cc947fa9d722c6b6699760d307f653815d0c
}

func TestSearchFilters_AddCreationEpochFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddCreationEpochFilter(anyValidSearchMatcher, anyValidCreationEpoch)
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterCreationEpoch, anyValidSearchMatcher, "13233261290750647837", true)
}

func TestSearchFilters_AddPayloadSizeFilter(t *testing.T) {
	var fs object.SearchFilters
	fs.AddPayloadSizeFilter(anyValidSearchMatcher, anyValidPayloadSize)
	require.Len(t, fs, 1)
	assertSearchFilter(t, fs, 0, object.FilterPayloadSize, anyValidSearchMatcher, "5544264194415343420", true)
}

func TestNewSearchFilters(t *testing.T) {
	require.Empty(t, object.NewSearchFilters())
}

func TestSearchFilters_ProtoMessage(t *testing.T) {
	const nFilters = 22
	require.Len(t, validSearchFiltersProto, nFilters, "not all applied filters are asserted")
	var fs object.SearchFilters

	// zero
	m := fs.ProtoMessage()
	require.Empty(t, m)

	// filled
	m = validSearchFilters.ProtoMessage()
	require.Len(t, m, nFilters)

	for i, exp := range validSearchFiltersProto {
		act := m[i]
		require.Equal(t, exp.k, act.GetKey())
		require.Equal(t, exp.m, act.GetMatchType())
		require.Equal(t, exp.v, act.GetValue())
	}
}

func TestSearchFilters_FromProtoMessage(t *testing.T) {
	ms := []*protoobject.SearchFilter{
		{MatchType: protoobject.MatchType(rand.Int32()), Key: "key_1", Value: "val_1"},
		{MatchType: protoobject.MatchType(rand.Int32()), Key: "key_2", Value: "val_2"},
	}

	var fs object.SearchFilters
	require.NoError(t, fs.FromProtoMessage(ms))
	require.Len(t, fs, len(ms))
	for i := range ms {
		require.EqualValues(t, ms[i].MatchType, fs[i].Operation())
		require.Equal(t, ms[i].Key, fs[i].Header())
		require.Equal(t, ms[i].Value, fs[i].Value())
	}

	require.NoError(t, fs.FromProtoMessage(nil))
	require.Empty(t, fs)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func([]*protoobject.SearchFilter)
		}{} {
			cp := slices.Clone(ms)
			for i := range ms {
				cp[i] = proto.Clone(ms[i]).(*protoobject.SearchFilter)
			}
			tc.corrupt(cp)
			require.EqualError(t, fs.FromProtoMessage(ms), tc.err)
		}
	})
}
