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

func TestAction_String(t *testing.T) {
	toPtr := func(v eacl.Action) *eacl.Action {
		return &v
	}

	testEnumStrings(t, new(eacl.Action), []enumStringItem{
		{val: toPtr(eacl.ActionAllow), str: "ALLOW"},
		{val: toPtr(eacl.ActionDeny), str: "DENY"},
		{val: toPtr(eacl.ActionUnspecified), str: "ACTION_UNSPECIFIED"},
	})
}

func TestRole_String(t *testing.T) {
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

func TestOperation_String(t *testing.T) {
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

func TestMatch_String(t *testing.T) {
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

func TestFilterHeaderType_String(t *testing.T) {
	toPtr := func(v eacl.FilterHeaderType) *eacl.FilterHeaderType {
		return &v
	}

	testEnumStrings(t, new(eacl.FilterHeaderType), []enumStringItem{
		{val: toPtr(eacl.HeaderFromRequest), str: "REQUEST"},
		{val: toPtr(eacl.HeaderFromObject), str: "OBJECT"},
		{val: toPtr(eacl.HeaderTypeUnspecified), str: "HEADER_UNSPECIFIED"},
	})
}
