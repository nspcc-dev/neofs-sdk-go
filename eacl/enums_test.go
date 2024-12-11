package eacl_test

import (
	"fmt"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	"github.com/stretchr/testify/require"
)

var (
	protoActions = map[eacl.Action]protoacl.Action{
		eacl.ActionUnspecified: protoacl.Action_ACTION_UNSPECIFIED,
		eacl.ActionAllow:       protoacl.Action_ALLOW,
		eacl.ActionDeny:        protoacl.Action_DENY,
	}

	protoOperations = map[eacl.Operation]protoacl.Operation{
		eacl.OperationUnspecified: protoacl.Operation_OPERATION_UNSPECIFIED,
		eacl.OperationGet:         protoacl.Operation_GET,
		eacl.OperationHead:        protoacl.Operation_HEAD,
		eacl.OperationPut:         protoacl.Operation_PUT,
		eacl.OperationDelete:      protoacl.Operation_DELETE,
		eacl.OperationSearch:      protoacl.Operation_SEARCH,
		eacl.OperationRange:       protoacl.Operation_GETRANGE,
		eacl.OperationRangeHash:   protoacl.Operation_GETRANGEHASH,
	}

	protoRoles = map[eacl.Role]protoacl.Role{
		eacl.RoleUnspecified: protoacl.Role_ROLE_UNSPECIFIED,
		eacl.RoleUser:        protoacl.Role_USER,
		eacl.RoleSystem:      protoacl.Role_SYSTEM,
		eacl.RoleOthers:      protoacl.Role_OTHERS,
	}

	protoMatches = map[eacl.Match]protoacl.MatchType{
		eacl.MatchUnspecified:    protoacl.MatchType_MATCH_TYPE_UNSPECIFIED,
		eacl.MatchStringEqual:    protoacl.MatchType_STRING_EQUAL,
		eacl.MatchStringNotEqual: protoacl.MatchType_STRING_NOT_EQUAL,
		eacl.MatchNotPresent:     protoacl.MatchType_NOT_PRESENT,
		eacl.MatchNumGT:          protoacl.MatchType_NUM_GT,
		eacl.MatchNumGE:          protoacl.MatchType_NUM_GE,
		eacl.MatchNumLT:          protoacl.MatchType_NUM_LT,
		eacl.MatchNumLE:          protoacl.MatchType_NUM_LE,
	}

	protoHeaderTypes = map[eacl.FilterHeaderType]protoacl.HeaderType{
		eacl.HeaderTypeUnspecified: protoacl.HeaderType_HEADER_UNSPECIFIED,
		eacl.HeaderFromRequest:     protoacl.HeaderType_REQUEST,
		eacl.HeaderFromObject:      protoacl.HeaderType_OBJECT,
		eacl.HeaderFromService:     protoacl.HeaderType_SERVICE,
	}

	actionStrings = map[eacl.Action]string{
		-1:               "-1",
		0:                "ACTION_UNSPECIFIED",
		eacl.ActionAllow: "ALLOW",
		eacl.ActionDeny:  "DENY",
		3:                "3",
	}
	roleStrings = map[eacl.Role]string{
		-1:              "-1",
		0:               "ROLE_UNSPECIFIED",
		eacl.RoleUser:   "USER",
		eacl.RoleSystem: "SYSTEM",
		eacl.RoleOthers: "OTHERS",
		4:               "4",
	}
	opStrings = map[eacl.Operation]string{
		-1:                      "-1",
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
		-1:                       "-1",
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
		-1:                     "-1",
		0:                      "HEADER_UNSPECIFIED",
		eacl.HeaderFromRequest: "REQUEST",
		eacl.HeaderFromObject:  "OBJECT",
		eacl.HeaderFromService: "SERVICE",
		4:                      "4",
	}
)

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
	for x, y := range protoActions {
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

type enum interface {
	~int32
	fmt.Stringer
}

func testEnumToString[T enum](t testing.TB, m map[T]string) {
	for n, s := range m {
		require.Equal(t, s, n.String())
	}
}

func testEnumDecodeString[T enum](t *testing.T, m map[T]string, f func(*T, string) bool) {
	t.Run("invalid", func(t *testing.T) {
		for _, s := range []string{"", "foo", "1.2"} {
			require.False(t, f(new(T), s))
		}
	})
	for n, s := range m {
		var res T
		ok := f(&res, s)
		require.True(t, ok)
		require.Equal(t, n, res)
	}
}

func TestActionToString(t *testing.T) {
	testEnumToString(t, actionStrings)
}

func TestActionFromString(t *testing.T) {
	testEnumDecodeString(t, actionStrings, (*eacl.Action).DecodeString)
}

func TestRoleProto(t *testing.T) {
	for x, y := range protoRoles {
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
	testEnumToString(t, roleStrings)
}

func TestRoleFromString(t *testing.T) {
	testEnumDecodeString(t, roleStrings, (*eacl.Role).DecodeString)
}

func TestOperationProto(t *testing.T) {
	for x, y := range protoOperations {
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
	testEnumToString(t, opStrings)
}

func TestOperationFromString(t *testing.T) {
	testEnumDecodeString(t, opStrings, (*eacl.Operation).DecodeString)
}

func TestMatchProto(t *testing.T) {
	for x, y := range protoMatches {
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
	testEnumToString(t, matcherStrings)
}

func TestMatcherFromString(t *testing.T) {
	testEnumDecodeString(t, matcherStrings, (*eacl.Match).DecodeString)
}

func TestFilterHeaderTypeProto(t *testing.T) {
	for x, y := range protoHeaderTypes {
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
	testEnumToString(t, headerTypeStrings)
}

func TestHeaderTypeFromString(t *testing.T) {
	testEnumDecodeString(t, headerTypeStrings, (*eacl.FilterHeaderType).DecodeString)
}
