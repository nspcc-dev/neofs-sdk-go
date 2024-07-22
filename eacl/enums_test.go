package eacl_test

import (
	"testing"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

var (
	eqV2Actions = map[eacl.Action]v2acl.Action{
		eacl.ActionUnspecified: v2acl.ActionUnknown,
		eacl.ActionAllow:       v2acl.ActionAllow,
		eacl.ActionDeny:        v2acl.ActionDeny,
	}

	eqV2Operations = map[eacl.Operation]v2acl.Operation{
		eacl.OperationUnspecified: v2acl.OperationUnknown,
		eacl.OperationGet:         v2acl.OperationGet,
		eacl.OperationHead:        v2acl.OperationHead,
		eacl.OperationPut:         v2acl.OperationPut,
		eacl.OperationDelete:      v2acl.OperationDelete,
		eacl.OperationSearch:      v2acl.OperationSearch,
		eacl.OperationRange:       v2acl.OperationRange,
		eacl.OperationRangeHash:   v2acl.OperationRangeHash,
	}

	eqV2Roles = map[eacl.Role]v2acl.Role{
		eacl.RoleUnspecified: v2acl.RoleUnknown,
		eacl.RoleUser:        v2acl.RoleUser,
		eacl.RoleSystem:      v2acl.RoleSystem,
		eacl.RoleOthers:      v2acl.RoleOthers,
	}

	eqV2Matches = map[eacl.Match]v2acl.MatchType{
		eacl.MatchUnspecified:    v2acl.MatchTypeUnknown,
		eacl.MatchStringEqual:    v2acl.MatchTypeStringEqual,
		eacl.MatchStringNotEqual: v2acl.MatchTypeStringNotEqual,
		eacl.MatchNotPresent:     v2acl.MatchTypeNotPresent,
		eacl.MatchNumGT:          v2acl.MatchTypeNumGT,
		eacl.MatchNumGE:          v2acl.MatchTypeNumGE,
		eacl.MatchNumLT:          v2acl.MatchTypeNumLT,
		eacl.MatchNumLE:          v2acl.MatchTypeNumLE,
	}

	eqV2HeaderTypes = map[eacl.FilterHeaderType]v2acl.HeaderType{
		eacl.HeaderTypeUnspecified: v2acl.HeaderTypeUnknown,
		eacl.HeaderFromRequest:     v2acl.HeaderTypeRequest,
		eacl.HeaderFromObject:      v2acl.HeaderTypeObject,
		eacl.HeaderFromService:     v2acl.HeaderTypeService,
	}

	actionStrings = map[eacl.Action]string{
		0:                "ACTION_UNSPECIFIED",
		eacl.ActionAllow: "ALLOW",
		eacl.ActionDeny:  "DENY",
		3:                "3",
	}
	roleStrings = map[eacl.Role]string{
		0:               "ROLE_UNSPECIFIED",
		eacl.RoleUser:   "USER",
		eacl.RoleSystem: "SYSTEM",
		eacl.RoleOthers: "OTHERS",
		4:               "4",
	}
	opStrings = map[eacl.Operation]string{
		0:                       "OPERATION_UNSPECIFIED",
		eacl.OperationGet:       "GET",
		eacl.OperationHead:      "HEAD",
		eacl.OperationPut:       "PUT",
		eacl.OperationDelete:    "DELETE",
		eacl.OperationSearch:    "SEARCH",
		eacl.OperationRange:     "GETRANGE",
		eacl.OperationRangeHash: "GETRANGEHASH",
		8:                       "8",
	}
	matcherStrings = map[eacl.Match]string{
		0:                        "MATCH_TYPE_UNSPECIFIED",
		eacl.MatchStringEqual:    "STRING_EQUAL",
		eacl.MatchStringNotEqual: "STRING_NOT_EQUAL",
		eacl.MatchNotPresent:     "NOT_PRESENT",
		eacl.MatchNumGT:          "NUM_GT",
		eacl.MatchNumGE:          "NUM_GE",
		eacl.MatchNumLT:          "NUM_LT",
		eacl.MatchNumLE:          "NUM_LE",
		8:                        "8",
	}
	headerTypeStrings = map[eacl.FilterHeaderType]string{
		0:                      "HEADER_UNSPECIFIED",
		eacl.HeaderFromRequest: "REQUEST",
		eacl.HeaderFromObject:  "OBJECT",
		eacl.HeaderFromService: "SERVICE",
		4:                      "4",
	}
)

func TestAction(t *testing.T) {
	require.Equal(t, eacl.ActionUnspecified, eacl.ActionUnknown)
	t.Run("known actions", func(t *testing.T) {
		for i := eacl.ActionUnspecified; i <= eacl.ActionDeny; i++ {
			require.Equal(t, eqV2Actions[i], i.ToV2())
			require.Equal(t, eacl.ActionFromV2(i.ToV2()), i)
		}
	})

	t.Run("unknown actions", func(t *testing.T) {
		require.EqualValues(t, 1000, eacl.Action(1000).ToV2())
		require.EqualValues(t, 1000, eacl.ActionFromV2(1000))
	})
}

func TestOperation(t *testing.T) {
	require.Equal(t, eacl.OperationUnspecified, eacl.OperationUnknown)
	t.Run("known operations", func(t *testing.T) {
		for i := eacl.OperationUnspecified; i <= eacl.OperationRangeHash; i++ {
			require.Equal(t, eqV2Operations[i], i.ToV2())
			require.Equal(t, eacl.OperationFromV2(i.ToV2()), i)
		}
	})

	t.Run("unknown operations", func(t *testing.T) {
		require.EqualValues(t, 1000, eacl.Operation(1000).ToV2())
		require.EqualValues(t, 1000, eacl.OperationFromV2(1000))
	})
}

func TestRole(t *testing.T) {
	t.Run("known roles", func(t *testing.T) {
		for i := eacl.RoleUnspecified; i <= eacl.RoleOthers; i++ {
			require.Equal(t, eqV2Roles[i], i.ToV2())
			require.Equal(t, eacl.RoleFromV2(i.ToV2()), i)
		}
	})

	t.Run("unknown roles", func(t *testing.T) {
		require.EqualValues(t, 1000, eacl.Operation(1000).ToV2())
		require.EqualValues(t, 1000, eacl.RoleFromV2(1000))
	})
}

func TestMatch(t *testing.T) {
	require.Equal(t, eacl.MatchUnspecified, eacl.MatchUnknown)
	t.Run("known matches", func(t *testing.T) {
		for i := eacl.MatchUnspecified; i <= eacl.MatchStringNotEqual; i++ {
			require.Equal(t, eqV2Matches[i], i.ToV2())
			require.Equal(t, eacl.MatchFromV2(i.ToV2()), i)
		}
	})

	t.Run("unknown matches", func(t *testing.T) {
		require.EqualValues(t, 1000, eacl.Match(1000).ToV2())
		require.EqualValues(t, 1000, eacl.MatchFromV2(1000))
	})
}

func TestFilterHeaderType(t *testing.T) {
	require.Equal(t, eacl.HeaderTypeUnspecified, eacl.HeaderTypeUnknown)
	t.Run("known header types", func(t *testing.T) {
		for i := eacl.HeaderTypeUnspecified; i <= eacl.HeaderFromService; i++ {
			require.Equal(t, eqV2HeaderTypes[i], i.ToV2())
			require.Equal(t, eacl.FilterHeaderTypeFromV2(i.ToV2()), i)
		}
	})

	t.Run("unknown header types", func(t *testing.T) {
		require.EqualValues(t, 1000, eacl.FilterHeaderType(1000).ToV2())
		require.EqualValues(t, 1000, eacl.FilterHeaderTypeFromV2(1000))
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
		"UNSPECIFIED",
	} {
		require.False(t, e.DecodeString(str))
	}
}

func TestActionProto(t *testing.T) {
	for x, y := range map[v2acl.Action]eacl.Action{
		v2acl.ActionUnknown: eacl.ActionUnspecified,
		v2acl.ActionAllow:   eacl.ActionAllow,
		v2acl.ActionDeny:    eacl.ActionDeny,
	} {
		require.EqualValues(t, x, y)
	}
}

func TestAction_String(t *testing.T) {
	for a, s := range actionStrings {
		require.Equal(t, s, a.String())
	}

	toPtr := func(v eacl.Action) *eacl.Action {
		return &v
	}

	testEnumStrings(t, new(eacl.Action), []enumStringItem{
		{val: toPtr(eacl.ActionAllow), str: "ALLOW"},
		{val: toPtr(eacl.ActionDeny), str: "DENY"},
		{val: toPtr(eacl.ActionUnspecified), str: "ACTION_UNSPECIFIED"},
	})
}

type enum interface{ ~uint32 }

func testEnumToString[T enum](t testing.TB, m map[T]string, f func(T) string) {
	for n, s := range m {
		require.Equal(t, s, f(n))
	}
}

func testEnumFromString[T enum](t *testing.T, m map[T]string, f func(string) (T, bool)) {
	t.Run("invalid", func(t *testing.T) {
		for _, s := range []string{"", "foo", "1.2"} {
			_, ok := f(s)
			require.False(t, ok)
		}
	})
	for n, s := range m {
		v, ok := f(s)
		require.True(t, ok)
		require.Equal(t, n, v)
	}
}

func TestActionToString(t *testing.T) {
	testEnumToString(t, actionStrings, eacl.ActionToString)
}

func TestActionFromString(t *testing.T) {
	testEnumFromString(t, actionStrings, eacl.ActionFromString)
}

func TestRoleProto(t *testing.T) {
	for x, y := range map[v2acl.Role]eacl.Role{
		v2acl.RoleUnknown: eacl.RoleUnspecified,
		v2acl.RoleUser:    eacl.RoleUser,
		v2acl.RoleSystem:  eacl.RoleSystem,
		v2acl.RoleOthers:  eacl.RoleOthers,
	} {
		require.EqualValues(t, x, y)
	}
}

func TestRole_String(t *testing.T) {
	for r, s := range roleStrings {
		require.Equal(t, s, r.String())
	}

	toPtr := func(v eacl.Role) *eacl.Role {
		return &v
	}

	testEnumStrings(t, new(eacl.Role), []enumStringItem{
		{val: toPtr(eacl.RoleUser), str: "USER"},
		{val: toPtr(eacl.RoleSystem), str: "SYSTEM"},
		{val: toPtr(eacl.RoleOthers), str: "OTHERS"},
		{val: toPtr(eacl.RoleUnspecified), str: "ROLE_UNSPECIFIED"},
	})
}

func TestRoleToString(t *testing.T) {
	testEnumToString(t, roleStrings, eacl.RoleToString)
}

func TestRoleFromString(t *testing.T) {
	testEnumFromString(t, roleStrings, eacl.RoleFromString)
}

func TestOperationProto(t *testing.T) {
	for x, y := range map[v2acl.Operation]eacl.Operation{
		v2acl.OperationUnknown:   eacl.OperationUnspecified,
		v2acl.OperationGet:       eacl.OperationGet,
		v2acl.OperationHead:      eacl.OperationHead,
		v2acl.OperationPut:       eacl.OperationPut,
		v2acl.OperationDelete:    eacl.OperationDelete,
		v2acl.OperationSearch:    eacl.OperationSearch,
		v2acl.OperationRange:     eacl.OperationRange,
		v2acl.OperationRangeHash: eacl.OperationRangeHash,
	} {
		require.EqualValues(t, x, y)
	}
}

func TestOperation_String(t *testing.T) {
	for op, s := range opStrings {
		require.Equal(t, s, op.String())
	}

	toPtr := func(v eacl.Operation) *eacl.Operation {
		return &v
	}

	testEnumStrings(t, new(eacl.Operation), []enumStringItem{
		{val: toPtr(eacl.OperationGet), str: "GET"},
		{val: toPtr(eacl.OperationPut), str: "PUT"},
		{val: toPtr(eacl.OperationHead), str: "HEAD"},
		{val: toPtr(eacl.OperationDelete), str: "DELETE"},
		{val: toPtr(eacl.OperationSearch), str: "SEARCH"},
		{val: toPtr(eacl.OperationRange), str: "GETRANGE"},
		{val: toPtr(eacl.OperationRangeHash), str: "GETRANGEHASH"},
		{val: toPtr(eacl.OperationUnspecified), str: "OPERATION_UNSPECIFIED"},
	})
}

func TestOperationToString(t *testing.T) {
	testEnumToString(t, opStrings, eacl.OperationToString)
}

func TestOperationFromString(t *testing.T) {
	testEnumFromString(t, opStrings, eacl.OperationFromString)
}

func TestMatchProto(t *testing.T) {
	for x, y := range map[v2acl.MatchType]eacl.Match{
		v2acl.MatchTypeUnknown:        eacl.MatchUnspecified,
		v2acl.MatchTypeStringEqual:    eacl.MatchStringEqual,
		v2acl.MatchTypeStringNotEqual: eacl.MatchStringNotEqual,
		v2acl.MatchTypeNotPresent:     eacl.MatchNotPresent,
		v2acl.MatchTypeNumGT:          eacl.MatchNumGT,
		v2acl.MatchTypeNumGE:          eacl.MatchNumGE,
		v2acl.MatchTypeNumLT:          eacl.MatchNumLT,
		v2acl.MatchTypeNumLE:          eacl.MatchNumLE,
	} {
		require.EqualValues(t, x, y)
	}
}

func TestMatch_String(t *testing.T) {
	for m, s := range matcherStrings {
		require.Equal(t, s, m.String())
	}

	toPtr := func(v eacl.Match) *eacl.Match {
		return &v
	}

	testEnumStrings(t, new(eacl.Match), []enumStringItem{
		{val: toPtr(eacl.MatchStringEqual), str: "STRING_EQUAL"},
		{val: toPtr(eacl.MatchStringNotEqual), str: "STRING_NOT_EQUAL"},
		{val: toPtr(eacl.MatchUnspecified), str: "MATCH_TYPE_UNSPECIFIED"},
		{val: toPtr(eacl.MatchNotPresent), str: "NOT_PRESENT"},
		{val: toPtr(eacl.MatchNumGT), str: "NUM_GT"},
		{val: toPtr(eacl.MatchNumGE), str: "NUM_GE"},
		{val: toPtr(eacl.MatchNumLT), str: "NUM_LT"},
		{val: toPtr(eacl.MatchNumLE), str: "NUM_LE"},
	})
}

func TestMatcherToString(t *testing.T) {
	testEnumToString(t, matcherStrings, eacl.MatcherToString)
}

func TestMatcherFromString(t *testing.T) {
	testEnumFromString(t, matcherStrings, eacl.MatcherFromString)
}

func TestFilterHeaderTypeProto(t *testing.T) {
	for x, y := range map[v2acl.HeaderType]eacl.FilterHeaderType{
		v2acl.HeaderTypeUnknown: eacl.FilterHeaderType(0),
		v2acl.HeaderTypeRequest: eacl.HeaderFromRequest,
		v2acl.HeaderTypeObject:  eacl.HeaderFromObject,
		v2acl.HeaderTypeService: eacl.HeaderFromService,
	} {
		require.EqualValues(t, x, y)
	}
}

func TestFilterHeaderType_String(t *testing.T) {
	for h, s := range headerTypeStrings {
		require.Equal(t, s, h.String())
	}

	toPtr := func(v eacl.FilterHeaderType) *eacl.FilterHeaderType {
		return &v
	}

	testEnumStrings(t, new(eacl.FilterHeaderType), []enumStringItem{
		{val: toPtr(eacl.HeaderFromRequest), str: "REQUEST"},
		{val: toPtr(eacl.HeaderFromObject), str: "OBJECT"},
		{val: toPtr(eacl.HeaderFromService), str: "SERVICE"},
		{val: toPtr(eacl.HeaderTypeUnspecified), str: "HEADER_UNSPECIFIED"},
	})
}

func TestHeaderTypeToString(t *testing.T) {
	testEnumToString(t, headerTypeStrings, eacl.HeaderTypeToString)
}

func TestHeaderTypeFromString(t *testing.T) {
	testEnumFromString(t, headerTypeStrings, eacl.HeaderTypeFromString)
}
