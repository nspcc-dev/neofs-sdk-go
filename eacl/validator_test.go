package eacl

import (
	"bytes"
	"slices"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func checkIgnoredAction(t *testing.T, expected Action, v *Validator, vu *ValidationUnit) {
	action, ok, err := v.CalculateAction(vu)
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, expected, action)
}

func checkAction(t *testing.T, expected Action, v *Validator, vu *ValidationUnit) {
	action, ok, err := v.CalculateAction(vu)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, expected, action)
}

func checkDefaultAction(t *testing.T, v *Validator, vu *ValidationUnit, msgAndArgs ...any) {
	action, ok, err := v.CalculateAction(vu)
	require.NoError(t, err)
	require.False(t, ok, msgAndArgs)
	require.Equal(t, ActionAllow, action, msgAndArgs...)
}

func TestFilterMatch(t *testing.T) {
	tgt := Target{}
	tgt.SetRole(RoleOthers)

	t.Run("simple header match", func(t *testing.T) {
		tb := &Table{}

		r1 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt},
			NewObjectPropertyFilter("a", MatchStringEqual, "xxx"))
		r2 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt},
			NewRequestHeaderFilter("b", MatchStringNotEqual, "yyy"))
		r3 := ConstructRecord(ActionAllow, OperationUnspecified, []Target{tgt})

		tb.SetRecords([]Record{r1, r2, r3})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		checkAction(t, ActionAllow, v, vu)

		hs.obj = makeHeaders("b", "yyy")
		checkAction(t, ActionAllow, v, vu)

		hs.obj = makeHeaders("a", "xxx")
		checkAction(t, ActionDeny, v, vu)

		hs.obj = nil
		hs.req = makeHeaders("b", "yyy")
		checkAction(t, ActionAllow, v, vu)

		hs.req = makeHeaders("b", "abc")
		checkAction(t, ActionDeny, v, vu)
	})

	t.Run("all filters must match", func(t *testing.T) {
		tb := &Table{}
		r1 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt},
			NewObjectPropertyFilter("a", MatchStringEqual, "xxx"),
			NewRequestHeaderFilter("b", MatchStringEqual, "yyy"))
		r2 := ConstructRecord(ActionAllow, OperationUnspecified, []Target{tgt})

		tb.SetRecords([]Record{r1, r2})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		hs.obj = makeHeaders("a", "xxx")
		checkAction(t, ActionAllow, v, vu)

		hs.req = makeHeaders("b", "yyy")
		checkAction(t, ActionDeny, v, vu)

		hs.obj = nil
		checkAction(t, ActionAllow, v, vu)
	})

	t.Run("filters with unknown type are skipped", func(t *testing.T) {
		tb := &Table{}
		r1 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt},
			ConstructFilter(HeaderTypeUnspecified, "a", MatchStringEqual, "xxx"))
		r2 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt},
			ConstructFilter(0xFF, "b", MatchStringEqual, "yyy"))
		r3 := ConstructRecord(ActionAllow, OperationUnspecified, []Target{tgt})

		tb.SetRecords([]Record{r1, r2, r3})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		checkDefaultAction(t, v, vu)

		hs.obj = makeHeaders("a", "xxx")
		checkDefaultAction(t, v, vu)

		hs.obj = nil
		hs.req = makeHeaders("b", "yyy")
		checkDefaultAction(t, v, vu)
	})

	t.Run("filters with match function are skipped", func(t *testing.T) {
		tb := &Table{}
		r1 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt},
			NewObjectPropertyFilter("a", 0xFF, "xxx"))
		r2 := ConstructRecord(ActionDeny, OperationUnspecified, []Target{tgt})

		tb.SetRecords([]Record{r1, r2})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		checkAction(t, ActionDeny, v, vu)

		hs.obj = makeHeaders("a", "xxx")
		checkAction(t, ActionDeny, v, vu)
	})
}

func TestOperationMatch(t *testing.T) {
	tgt := Target{}
	tgt.SetRole(RoleOthers)

	t.Run("single operation", func(t *testing.T) {
		tb := &Table{}
		tb.SetRecords([]Record{
			newRecord(ActionDeny, OperationPut, tgt),
			newRecord(ActionAllow, OperationGet, tgt),
		})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)

		vu.op = OperationPut
		checkAction(t, ActionDeny, v, vu)

		vu.op = OperationGet
		checkAction(t, ActionAllow, v, vu)
	})

	t.Run("unknown operation", func(t *testing.T) {
		tb := &Table{}
		tb.SetRecords([]Record{
			newRecord(ActionDeny, OperationUnspecified, tgt),
			newRecord(ActionAllow, OperationGet, tgt),
		})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)

		// TODO discuss if both next tests should result in DENY
		vu.op = OperationPut
		checkDefaultAction(t, v, vu)

		vu.op = OperationGet
		checkAction(t, ActionAllow, v, vu)
	})
}

func TestTargetMatches(t *testing.T) {
	pubs := makeKeys(t, 3)
	accs := usertest.IDs(3)

	t.Run("keys", func(t *testing.T) {
		tgt1 := Target{}
		tgt1.SetRawSubjects(pubs[0:2])
		tgt1.SetRole(RoleUser)

		tgt2 := NewTargetByRole(RoleOthers)

		r := &Record{}
		r.SetTargets(tgt1, tgt2)

		u := newValidationUnit(RoleUser, pubs[0], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnit(RoleUser, pubs[2], nil)
		require.False(t, targetMatches(u, r))

		u = newValidationUnit(RoleUnspecified, pubs[1], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnit(RoleOthers, pubs[2], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnit(RoleSystem, pubs[2], nil)
		require.False(t, targetMatches(u, r))
	})

	t.Run("accounts", func(t *testing.T) {
		tgt1 := Target{}
		tgt1.SetAccounts(accs[0:2])
		tgt1.SetRole(RoleUser)

		tgt2 := Target{}
		tgt2.SetRole(RoleOthers)

		r := &Record{}
		r.SetTargets(tgt1, tgt2)

		u := newValidationUnitWithScriptHash(RoleUser, accs[0], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleUser, accs[2], nil)
		require.False(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleUnspecified, accs[1], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleOthers, accs[2], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleSystem, accs[2], nil)
		require.False(t, targetMatches(u, r))
	})

	t.Run("mix", func(t *testing.T) {
		tgt1 := Target{}
		accList := make([][]byte, 0, len(accs))
		for _, acc := range accs {
			accList = append(accList, bytes.Clone(acc[:]))
		}

		tgt1.SetRawSubjects(slices.Concat(pubs[0:2], accList[0:2]))
		tgt1.SetRole(RoleUser)

		tgt2 := Target{}
		tgt2.SetRole(RoleOthers)

		r := &Record{}
		r.SetTargets(tgt1, tgt2)

		t.Run("user role", func(t *testing.T) {
			u := newValidationUnitWithScriptHash(RoleUser, accs[0], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnitWithScriptHash(RoleUser, accs[2], nil)
			require.False(t, targetMatches(u, r))

			u = newValidationUnit(RoleUser, pubs[0], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnit(RoleUser, pubs[2], nil)
			require.False(t, targetMatches(u, r))
		})

		t.Run("others role", func(t *testing.T) {
			u := newValidationUnitWithScriptHash(RoleUnspecified, accs[1], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnitWithScriptHash(RoleOthers, accs[2], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnitWithScriptHash(RoleSystem, accs[2], nil)
			require.False(t, targetMatches(u, r))

			u = newValidationUnit(RoleUnspecified, pubs[1], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnit(RoleOthers, pubs[2], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnit(RoleSystem, pubs[2], nil)
			require.False(t, targetMatches(u, r))
		})
	})
}

func TestSystemRoleModificationIgnored(t *testing.T) {
	tgt := Target{}
	tgt.SetRole(RoleSystem)

	operations := []Operation{
		OperationPut,
		OperationGet,
		OperationDelete,
		OperationHead,
		OperationRange,
		OperationRangeHash,
	}

	tb := &Table{}
	rrs := []Record{}
	for _, operation := range operations {
		rrs = append(rrs, newRecord(ActionDeny, operation, tgt))
	}
	tb.SetRecords(rrs)

	v := NewValidator()
	vu := newValidationUnit(RoleSystem, nil, tb)

	for _, operation := range operations {
		vu.op = operation
		checkIgnoredAction(t, ActionAllow, v, vu)
	}
}

func makeKeys(t *testing.T, n int) [][]byte {
	pubs := make([][]byte, n)
	for i := range pubs {
		pubs[i] = testutil.RandByteSlice(33)
		pubs[i][0] = 0x02
	}
	return pubs
}

type (
	hdr struct {
		key, value string
	}

	headers struct {
		obj []Header
		req []Header
	}
)

func (h hdr) Key() string   { return h.key }
func (h hdr) Value() string { return h.value }

func makeHeaders(kv ...string) []Header {
	hs := make([]Header, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		hs[i/2] = hdr{kv[i], kv[i+1]}
	}
	return hs
}

func (h headers) HeadersOfType(ht FilterHeaderType) ([]Header, bool, error) {
	switch ht {
	case HeaderFromRequest:
		return h.req, true, nil
	case HeaderFromObject:
		return h.obj, true, nil
	default:
		return nil, false, nil
	}
}

func newRecord(a Action, op Operation, tgt ...Target) Record {
	r := Record{}
	r.SetAction(a)
	r.SetOperation(op)
	r.SetTargets(tgt...)
	return r
}

func newValidationUnit(role Role, key []byte, table *Table) *ValidationUnit {
	return new(ValidationUnit).
		WithRole(role).
		WithSenderKey(key).
		WithEACLTable(table)
}

func newValidationUnitWithScriptHash(role Role, account user.ID, table *Table) *ValidationUnit {
	return new(ValidationUnit).
		WithRole(role).
		WithAccount(account).
		WithEACLTable(table)
}

func TestNumericRules(t *testing.T) {
	for _, tc := range []struct {
		m   Match
		h   string
		f   string
		exp bool
	}{
		// >
		{MatchNumGT, "non-decimal", "0", false},
		{MatchNumGT, "0", "non-decimal", false},
		{MatchNumGT, "-2", "-1", false},
		{MatchNumGT, "0", "0", false},
		{MatchNumGT, "-1", "0", false},
		{MatchNumGT, "0", "1", false},
		{MatchNumGT, "111111111111111111111111111110", "111111111111111111111111111111", false}, // more than 64-bit
		{MatchNumGT, "111111111111111111111111111111", "111111111111111111111111111111", false},
		{MatchNumGT, "-111111111111111111111111111111", "-111111111111111111111111111110", false},
		{MatchNumGT, "-1", "-2", true},
		{MatchNumGT, "0", "-1", true},
		{MatchNumGT, "1", "0", true},
		{MatchNumGT, "111111111111111111111111111111", "111111111111111111111111111110", true},
		{MatchNumGT, "-111111111111111111111111111110", "-111111111111111111111111111111", true},
		// >=
		{MatchNumGE, "0", "non-decimal", false},
		{MatchNumGE, "non-decimal", "0", false},
		{MatchNumGE, "-2", "-1", false},
		{MatchNumGE, "0", "0", true},
		{MatchNumGE, "-1", "0", false},
		{MatchNumGE, "0", "1", false},
		{MatchNumGE, "111111111111111111111111111110", "111111111111111111111111111111", false},
		{MatchNumGE, "111111111111111111111111111111", "111111111111111111111111111111", true},
		{MatchNumGE, "-111111111111111111111111111111", "-111111111111111111111111111110", false},
		{MatchNumGE, "-1", "-2", true},
		{MatchNumGE, "0", "-1", true},
		{MatchNumGE, "1", "0", true},
		{MatchNumGE, "111111111111111111111111111111", "111111111111111111111111111110", true},
		{MatchNumGE, "-111111111111111111111111111110", "-111111111111111111111111111111", true},
		// <
		{MatchNumLT, "non-decimal", "0", false},
		{MatchNumLT, "0", "non-decimal", false},
		{MatchNumLT, "-2", "-1", true},
		{MatchNumLT, "0", "0", false},
		{MatchNumLT, "-1", "0", true},
		{MatchNumLT, "0", "1", true},
		{MatchNumLT, "111111111111111111111111111110", "111111111111111111111111111111", true},
		{MatchNumLT, "111111111111111111111111111111", "111111111111111111111111111111", false},
		{MatchNumLT, "-111111111111111111111111111111", "-111111111111111111111111111110", true},
		{MatchNumLT, "-1", "-2", false},
		{MatchNumLT, "0", "-1", false},
		{MatchNumLT, "1", "0", false},
		{MatchNumLT, "111111111111111111111111111111", "111111111111111111111111111110", false},
		{MatchNumLT, "-111111111111111111111111111110", "-111111111111111111111111111111", false},
		// <=
		{MatchNumLE, "non-decimal", "0", false},
		{MatchNumLE, "0", "non-decimal", false},
		{MatchNumLE, "-2", "-1", true},
		{MatchNumLE, "0", "0", true},
		{MatchNumLE, "-1", "0", true},
		{MatchNumLE, "0", "1", true},
		{MatchNumLE, "111111111111111111111111111110", "111111111111111111111111111111", true},
		{MatchNumLE, "111111111111111111111111111111", "111111111111111111111111111111", true},
		{MatchNumLE, "-111111111111111111111111111111", "-111111111111111111111111111110", true},
		{MatchNumLE, "-1", "-2", false},
		{MatchNumLE, "0", "-1", false},
		{MatchNumLE, "1", "0", false},
		{MatchNumLE, "111111111111111111111111111111", "111111111111111111111111111110", false},
		{MatchNumLE, "-111111111111111111111111111110", "-111111111111111111111111111111", false},
	} {
		var rec Record
		rec.SetFilters([]Filter{NewObjectPropertyFilter("any_key", tc.m, tc.f)})
		hs := headers{obj: makeHeaders("any_key", tc.h)}

		v, err := matchFilters(hs, rec.filters)
		require.NoError(t, err)
		if tc.exp {
			require.Zero(t, v, tc)
		} else {
			require.Positive(t, v, tc)
		}
	}
}

func TestAbsenceRules(t *testing.T) {
	hs := headers{obj: makeHeaders(
		"key1", "val1",
		"key2", "val2",
	)}

	var r Record

	r.SetFilters([]Filter{
		NewObjectPropertyFilter("key2", MatchStringEqual, "val2"),
		NewObjectPropertyFilter("key1", MatchNotPresent, ""),
	})
	v, err := matchFilters(hs, r.filters)
	require.NoError(t, err)
	require.Positive(t, v)

	r.SetFilters([]Filter{
		NewObjectPropertyFilter("key1", MatchStringEqual, "val1"),
		NewObjectPropertyFilter("key2", MatchNotPresent, ""),
	})
	v, err = matchFilters(hs, r.filters)
	require.NoError(t, err)
	require.Positive(t, v)

	r.SetFilters([]Filter{
		NewObjectPropertyFilter("key1", MatchStringEqual, "val1"),
		NewObjectPropertyFilter("key2", MatchStringEqual, "val2"),
		NewObjectPropertyFilter("key3", MatchNotPresent, ""),
	})
	v, err = matchFilters(hs, r.filters)
	require.NoError(t, err)
	require.Zero(t, v)
}
