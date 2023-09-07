package eacl

import (
	"math/rand"
	"testing"

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

func checkDefaultAction(t *testing.T, v *Validator, vu *ValidationUnit) {
	action, ok := v.CalculateAction(vu)
	require.False(t, ok)
	require.Equal(t, ActionAllow, action)
}

func TestFilterMatch(t *testing.T) {
	tgt := *NewTarget()
	tgt.SetRole(RoleOthers)

	t.Run("simple header match", func(t *testing.T) {
		tb := NewTable()

		r := newRecord(ActionDeny, OperationUnknown, tgt)
		r.AddFilter(HeaderFromObject, MatchStringEqual, "a", "xxx")
		tb.AddRecord(r)

		r = newRecord(ActionDeny, OperationUnknown, tgt)
		r.AddFilter(HeaderFromRequest, MatchStringNotEqual, "b", "yyy")
		tb.AddRecord(r)

		tb.AddRecord(newRecord(ActionAllow, OperationUnknown, tgt))

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
		r := newRecord(ActionDeny, OperationUnknown, tgt)
		r.AddFilter(HeaderFromObject, MatchStringEqual, "a", "xxx")
		r.AddFilter(HeaderFromRequest, MatchStringEqual, "b", "yyy")
		tb.AddRecord(r)
		tb.AddRecord(newRecord(ActionAllow, OperationUnknown, tgt))

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
		r := newRecord(ActionDeny, OperationUnknown, tgt)
		r.AddFilter(HeaderTypeUnknown, MatchStringEqual, "a", "xxx")
		tb.AddRecord(r)

		r = newRecord(ActionDeny, OperationUnknown, tgt)
		r.AddFilter(0xFF, MatchStringEqual, "b", "yyy")
		tb.AddRecord(r)

		tb.AddRecord(newRecord(ActionDeny, OperationUnknown, tgt))

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
		r := newRecord(ActionAllow, OperationUnknown, tgt)
		r.AddFilter(HeaderFromObject, 0xFF, "a", "xxx")
		tb.AddRecord(r)
		tb.AddRecord(newRecord(ActionDeny, OperationUnknown, tgt))

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
		tb.AddRecord(newRecord(ActionDeny, OperationUnknown, tgt))
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

	u = newValidationUnit(RoleUnknown, pubs[1], nil)
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(RoleOthers, pubs[2], nil)
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(RoleSystem, pubs[2], nil)
	require.False(t, targetMatches(u, r))
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
