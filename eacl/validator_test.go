package eacl

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
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
	var tgt Target
	tgt.SetRole(RoleOthers)

	t.Run("simple header match", func(t *testing.T) {
		var f Filter
		f.SetAttributeType(AttributeObject)
		f.SetKey("a")
		f.SetMatcher(MatchStringEqual)
		f.SetValue("xxx")

		var r1 Record
		r1.SetAction(ActionDeny)
		r1.SetTargets([]Target{tgt})
		r1.SetFilters([]Filter{f})

		f.SetAttributeType(AttributeAPIRequest)
		f.SetKey("b")
		f.SetMatcher(MatchStringNotEqual)
		f.SetValue("yyy")

		var r2 Record
		r1.CopyTo(&r2)
		r2.SetFilters([]Filter{f})

		var r3 Record
		r3.SetAction(ActionAllow)
		r3.SetTargets([]Target{tgt})

		var tb Table
		tb.SetRecords([]Record{r1, r2, r3})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, &tb)
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
		fs := make([]Filter, 2)
		fs[0].SetAttributeType(AttributeObject)
		fs[0].SetKey("a")
		fs[0].SetMatcher(MatchStringEqual)
		fs[0].SetValue("xxx")
		fs[1].SetAttributeType(AttributeAPIRequest)
		fs[1].SetKey("b")
		fs[1].SetMatcher(MatchStringEqual)
		fs[1].SetValue("yyy")

		var r1 Record
		r1.SetAction(ActionDeny)
		r1.SetTargets([]Target{tgt})
		r1.SetFilters(fs)

		var r2 Record
		r2.SetAction(ActionAllow)
		r2.SetTargets([]Target{tgt})

		var tb Table
		tb.SetRecords([]Record{r1, r2})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, &tb)
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
		var f Filter
		f.SetAttributeType(0)
		f.SetKey("a")
		f.SetMatcher(MatchStringEqual)
		f.SetValue("xxx")

		var r1 Record
		r1.SetAction(ActionDeny)
		r1.SetTargets([]Target{tgt})
		r1.SetFilters([]Filter{f})

		f.SetAttributeType(0xFF)
		f.SetKey("b")
		f.SetValue("yyy")

		var r2 Record
		r2.SetAction(ActionDeny)
		r2.SetTargets([]Target{tgt})
		r2.SetFilters([]Filter{f})

		var r3 Record
		r2.SetAction(ActionDeny)
		r2.SetTargets([]Target{tgt})

		var tb Table
		tb.SetRecords([]Record{r1, r2, r3})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, &tb)
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
		var f Filter
		f.SetAttributeType(AttributeObject)
		f.SetKey("a")
		f.SetMatcher(0xFF)
		f.SetValue("xxx")

		var r1 Record
		r1.SetAction(ActionAllow)
		r1.SetTargets([]Target{tgt})
		r1.SetFilters([]Filter{f})

		var r2 Record
		r2.SetAction(ActionDeny)
		r2.SetTargets([]Target{tgt})

		var tb Table
		tb.SetRecords([]Record{r1, r2})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, &tb)
		hs := headers{}
		vu.hdrSrc = &hs

		checkAction(t, ActionDeny, v, vu)

		hs.obj = makeHeaders("a", "xxx")
		checkAction(t, ActionDeny, v, vu)
	})
}

func TestOperationMatch(t *testing.T) {
	var tgt Target
	tgt.SetRole(RoleOthers)

	t.Run("single operation", func(t *testing.T) {
		var r1, r2 Record
		r1.SetAction(ActionDeny)
		r1.SetOperation(acl.OpObjectPut)
		r1.SetTargets([]Target{tgt})
		r2.SetAction(ActionAllow)
		r2.SetOperation(acl.OpObjectGet)
		r2.SetTargets([]Target{tgt})

		var tb Table
		tb.SetRecords([]Record{r1, r2})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, &tb)

		vu.op = acl.OpObjectPut
		checkAction(t, ActionDeny, v, vu)

		vu.op = acl.OpObjectGet
		checkAction(t, ActionAllow, v, vu)
	})

	t.Run("unknown operation", func(t *testing.T) {
		var r1, r2 Record
		r1.SetAction(ActionDeny)
		r1.SetOperation(0)
		r1.SetTargets([]Target{tgt})
		r2.SetAction(ActionAllow)
		r2.SetOperation(acl.OpObjectGet)
		r2.SetTargets([]Target{tgt})

		var tb Table
		tb.SetRecords([]Record{r1, r2})

		v := NewValidator()
		vu := newValidationUnit(RoleOthers, nil, &tb)

		// TODO discuss if both next tests should result in DENY
		vu.op = acl.OpObjectPut
		checkDefaultAction(t, v, vu)

		vu.op = acl.OpObjectGet
		checkAction(t, ActionAllow, v, vu)
	})
}

func TestTargetMatches(t *testing.T) {
	pubs := makeKeys(t, 3)

	var tgt1 Target
	tgt1.SetPublicKeys(pubs[0:2])
	tgt1.SetRole(RoleUser)

	var tgt2 Target
	tgt2.SetRole(RoleOthers)

	var r Record
	r.SetTargets([]Target{tgt1, tgt2})

	u := newValidationUnit(RoleUser, pubs[0], nil)
	require.True(t, targetMatches(u, &r))

	u = newValidationUnit(RoleUser, pubs[2], nil)
	require.False(t, targetMatches(u, &r))

	u = newValidationUnit(0, pubs[1], nil)
	require.True(t, targetMatches(u, &r))

	u = newValidationUnit(RoleOthers, pubs[2], nil)
	require.True(t, targetMatches(u, &r))

	u = newValidationUnit(RoleSystem, pubs[2], nil)
	require.False(t, targetMatches(u, &r))
}

func TestSystemRoleModificationIgnored(t *testing.T) {
	var tgt Target
	tgt.SetRole(RoleSystem)

	operations := []acl.Op{
		acl.OpObjectPut,
		acl.OpObjectGet,
		acl.OpObjectDelete,
		acl.OpObjectHead,
		acl.OpObjectRange,
		acl.OpObjectHash,
	}

	var tb Table
	for _, operation := range operations {
		var r Record
		r.SetAction(ActionDeny)
		r.SetOperation(operation)
		r.SetTargets([]Target{tgt})
	}

	v := NewValidator()
	vu := newValidationUnit(RoleSystem, nil, &tb)

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

func (h headers) HeadersOfType(ht AttributeType) ([]Header, bool) {
	switch ht {
	case AttributeAPIRequest:
		return h.req, true
	case AttributeObject:
		return h.obj, true
	default:
		return nil, false
	}
}

func newValidationUnit(role Role, key []byte, table *Table) *ValidationUnit {
	return new(ValidationUnit).
		WithRole(role).
		WithSenderKey(key).
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
		var f Filter
		f.SetAttributeType(AttributeObject)
		f.SetKey("any_key")
		f.SetMatcher(tc.m)
		f.SetValue(tc.f)
		hs := headers{obj: makeHeaders("any_key", tc.h)}

		v := matchFilters(hs, []Filter{f})
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

	fs := make([]Filter, 2)
	fs[0].SetAttributeType(AttributeObject)
	fs[0].SetKey("key2")
	fs[0].SetMatcher(MatchStringEqual)
	fs[0].SetValue("val2")
	fs[1].SetAttributeType(AttributeObject)
	fs[1].SetKey("key1")
	fs[1].SetMatcher(MatchNotPresent)
	fs[1].SetValue("")

	v := matchFilters(hs, fs)
	require.Positive(t, v)

	fs[0].SetKey("key1")
	fs[0].SetValue("val1")
	fs[1].SetKey("key2")
	v = matchFilters(hs, fs)
	require.Positive(t, v)

	fs = []Filter{fs[0], fs[0], fs[1]}
	fs[1].SetKey("key2")
	fs[1].SetValue("val2")
	fs[2].SetKey("key3")
	v = matchFilters(hs, fs)
	require.Zero(t, v)
}
