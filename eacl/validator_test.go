package eacl

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func checkIgnoredAction(t *testing.T, expected Action, v *Validator, vu *ValidationUnit) {
	action, ok := v.CalculateAction(vu)
	require.False(t, ok)
	require.Equal(t, expected, action)
}

func checkAction(t *testing.T, expected Action, v *Validator, vu *ValidationUnit) {
	action, ok := v.CalculateAction(vu)
	require.True(t, ok)
	require.Equal(t, expected, action)
}

func checkDefaultAction(t *testing.T, v *Validator, vu *ValidationUnit, msgAndArgs ...any) {
	action, ok := v.CalculateAction(vu)
	require.False(t, ok, msgAndArgs)
	require.Equal(t, ActionAllow, action, msgAndArgs...)
}

func TestFilterMatch(t *testing.T) {
	tgt := *NewTarget()
	tgt.SetRole(RoleOthers)

	t.Run("simple header match", func(t *testing.T) {
		tb := NewTable()

		r := newRecord(ActionDeny, OperationUnspecified, tgt)
		r.AddFilter(HeaderFromObject, MatchStringEqual, "a", "xxx")
		tb.AddRecord(r)

		r = newRecord(ActionDeny, OperationUnspecified, tgt)
		r.AddFilter(HeaderFromRequest, MatchStringNotEqual, "b", "yyy")
		tb.AddRecord(r)

		tb.AddRecord(newRecord(ActionAllow, OperationUnspecified, tgt))

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
		tb := NewTable()
		r := newRecord(ActionDeny, OperationUnspecified, tgt)
		r.AddFilter(HeaderFromObject, MatchStringEqual, "a", "xxx")
		r.AddFilter(HeaderFromRequest, MatchStringEqual, "b", "yyy")
		tb.AddRecord(r)
		tb.AddRecord(newRecord(ActionAllow, OperationUnspecified, tgt))

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
		tb := NewTable()
		r := newRecord(ActionDeny, OperationUnspecified, tgt)
		r.AddFilter(HeaderTypeUnspecified, MatchStringEqual, "a", "xxx")
		tb.AddRecord(r)

		r = newRecord(ActionDeny, OperationUnspecified, tgt)
		r.AddFilter(0xFF, MatchStringEqual, "b", "yyy")
		tb.AddRecord(r)

		tb.AddRecord(newRecord(ActionDeny, OperationUnspecified, tgt))

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
		tb := NewTable()
		r := newRecord(ActionAllow, OperationUnspecified, tgt)
		r.AddFilter(HeaderFromObject, 0xFF, "a", "xxx")
		tb.AddRecord(r)
		tb.AddRecord(newRecord(ActionDeny, OperationUnspecified, tgt))

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
	tgt := *NewTarget()
	tgt.SetRole(RoleOthers)

	t.Run("single operation", func(t *testing.T) {
		tb := NewTable()
		tb.AddRecord(newRecord(ActionDeny, OperationPut, tgt))
		tb.AddRecord(newRecord(ActionAllow, OperationGet, tgt))

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, tb)

		vu.op = OperationPut
		checkAction(t, ActionDeny, v, vu)

		vu.op = OperationGet
		checkAction(t, ActionAllow, v, vu)
	})

	t.Run("unknown operation", func(t *testing.T) {
		tb := NewTable()
		tb.AddRecord(newRecord(ActionDeny, OperationUnspecified, tgt))
		tb.AddRecord(newRecord(ActionAllow, OperationGet, tgt))

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
		tgt1 := NewTarget()
		tgt1.SetBinaryKeys(pubs[0:2])
		tgt1.SetRole(RoleUser)

		tgt2 := NewTarget()
		tgt2.SetRole(RoleOthers)

		r := NewRecord()
		r.SetTargets(*tgt1, *tgt2)

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
		tgt1 := NewTarget()
		tgt1.SetAccounts(accs[0:2])
		tgt1.SetRole(RoleUser)

		tgt2 := NewTarget()
		tgt2.SetRole(RoleOthers)

		r := NewRecord()
		r.SetTargets(*tgt1, *tgt2)

		u := newValidationUnitWithScriptHash(RoleUser, accs[0], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleUser, accs[2], nil)
		require.False(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleUnknown, accs[1], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleOthers, accs[2], nil)
		require.True(t, targetMatches(u, r))

		u = newValidationUnitWithScriptHash(RoleSystem, accs[2], nil)
		require.False(t, targetMatches(u, r))
	})

	t.Run("mix", func(t *testing.T) {
		tgt1 := NewTarget()
		accList := make([][]byte, 0, len(accs))
		for _, acc := range accs {
			accList = append(accList, bytes.Clone(acc[:]))
		}

		tgt1.SetBinaryKeys(append(pubs[0:2], accList[0:2]...))
		tgt1.SetRole(RoleUser)

		tgt2 := NewTarget()
		tgt2.SetRole(RoleOthers)

		r := NewRecord()
		r.SetTargets(*tgt1, *tgt2)

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
			u := newValidationUnitWithScriptHash(RoleUnknown, accs[1], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnitWithScriptHash(RoleOthers, accs[2], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnitWithScriptHash(RoleSystem, accs[2], nil)
			require.False(t, targetMatches(u, r))

			u = newValidationUnit(RoleUnknown, pubs[1], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnit(RoleOthers, pubs[2], nil)
			require.True(t, targetMatches(u, r))

			u = newValidationUnit(RoleSystem, pubs[2], nil)
			require.False(t, targetMatches(u, r))
		})
	})
}

func TestSystemRoleModificationIgnored(t *testing.T) {
	tgt := *NewTarget()
	tgt.SetRole(RoleSystem)

	operations := []Operation{
		OperationPut,
		OperationGet,
		OperationDelete,
		OperationHead,
		OperationRange,
		OperationRangeHash,
	}

	tb := NewTable()
	for _, operation := range operations {
		tb.AddRecord(newRecord(ActionDeny, operation, tgt))
	}

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
		pubs[i] = make([]byte, 33)
		pubs[i][0] = 0x02
		//nolint:staticcheck
		_, err := rand.Read(pubs[i][1:])
		require.NoError(t, err)
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

func (h headers) HeadersOfType(ht FilterHeaderType) ([]Header, bool) {
	switch ht {
	case HeaderFromRequest:
		return h.req, true
	case HeaderFromObject:
		return h.obj, true
	default:
		return nil, false
	}
}

func newRecord(a Action, op Operation, tgt ...Target) *Record {
	r := NewRecord()
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
		rec.AddObjectAttributeFilter(tc.m, "any_key", tc.f)
		hs := headers{obj: makeHeaders("any_key", tc.h)}

		v := matchFilters(hs, rec.filters)
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

	r.AddObjectAttributeFilter(MatchStringEqual, "key2", "val2")
	r.AddObjectAttributeFilter(MatchNotPresent, "key1", "")
	v := matchFilters(hs, r.filters)
	require.Positive(t, v)

	r.filters = r.filters[:0]
	r.AddObjectAttributeFilter(MatchStringEqual, "key1", "val1")
	r.AddObjectAttributeFilter(MatchNotPresent, "key2", "")
	v = matchFilters(hs, r.filters)
	require.Positive(t, v)

	r.filters = r.filters[:0]
	r.AddObjectAttributeFilter(MatchStringEqual, "key1", "val1")
	r.AddObjectAttributeFilter(MatchStringEqual, "key2", "val2")
	r.AddObjectAttributeFilter(MatchNotPresent, "key3", "")
	v = matchFilters(hs, r.filters)
	require.Zero(t, v)
}
