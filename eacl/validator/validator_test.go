package validator

import (
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

func TestFilterMatch(t *testing.T) {
	tgt := eacl.NewTarget()
	tgt.SetRole(eacl.RoleOthers)

	t.Run("simple header match", func(t *testing.T) {
		tb := eacl.NewTable()

		r := newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt)
		r.AddFilter(eacl.HeaderFromObject, eacl.MatchStringEqual, "a", "xxx")
		tb.AddRecord(r)

		r = newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt)
		r.AddFilter(eacl.HeaderFromRequest, eacl.MatchStringNotEqual, "b", "yyy")
		tb.AddRecord(r)

		tb.AddRecord(newRecord(eacl.ActionAllow, eacl.OperationUnknown, tgt))

		v := New()
		vu := newValidationUnit(eacl.RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		hs.obj = makeHeaders("b", "yyy")
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		hs.obj = makeHeaders("a", "xxx")
		require.Equal(t, eacl.ActionDeny, v.CalculateAction(vu))

		hs.obj = nil
		hs.req = makeHeaders("b", "yyy")
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		hs.req = makeHeaders("b", "abc")
		require.Equal(t, eacl.ActionDeny, v.CalculateAction(vu))
	})

	t.Run("all filters must match", func(t *testing.T) {
		tb := eacl.NewTable()
		r := newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt)
		r.AddFilter(eacl.HeaderFromObject, eacl.MatchStringEqual, "a", "xxx")
		r.AddFilter(eacl.HeaderFromRequest, eacl.MatchStringEqual, "b", "yyy")
		tb.AddRecord(r)
		tb.AddRecord(newRecord(eacl.ActionAllow, eacl.OperationUnknown, tgt))

		v := New()
		vu := newValidationUnit(eacl.RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		hs.obj = makeHeaders("a", "xxx")
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		hs.req = makeHeaders("b", "yyy")
		require.Equal(t, eacl.ActionDeny, v.CalculateAction(vu))

		hs.obj = nil
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))
	})

	t.Run("filters with unknown type are skipped", func(t *testing.T) {
		tb := eacl.NewTable()
		r := newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt)
		r.AddFilter(eacl.HeaderTypeUnknown, eacl.MatchStringEqual, "a", "xxx")
		tb.AddRecord(r)

		r = newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt)
		r.AddFilter(0xFF, eacl.MatchStringEqual, "b", "yyy")
		tb.AddRecord(r)

		tb.AddRecord(newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt))

		v := New()
		vu := newValidationUnit(eacl.RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		hs.obj = makeHeaders("a", "xxx")
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		hs.obj = nil
		hs.req = makeHeaders("b", "yyy")
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))
	})

	t.Run("filters with match function are skipped", func(t *testing.T) {
		tb := eacl.NewTable()
		r := newRecord(eacl.ActionAllow, eacl.OperationUnknown, tgt)
		r.AddFilter(eacl.HeaderFromObject, 0xFF, "a", "xxx")
		tb.AddRecord(r)
		tb.AddRecord(newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt))

		v := New()
		vu := newValidationUnit(eacl.RoleOthers, nil, tb)
		hs := headers{}
		vu.hdrSrc = &hs

		require.Equal(t, eacl.ActionDeny, v.CalculateAction(vu))

		hs.obj = makeHeaders("a", "xxx")
		require.Equal(t, eacl.ActionDeny, v.CalculateAction(vu))
	})
}

func TestOperationMatch(t *testing.T) {
	tgt := eacl.NewTarget()
	tgt.SetRole(eacl.RoleOthers)

	t.Run("single operation", func(t *testing.T) {
		tb := eacl.NewTable()
		tb.AddRecord(newRecord(eacl.ActionDeny, eacl.OperationPut, tgt))
		tb.AddRecord(newRecord(eacl.ActionAllow, eacl.OperationGet, tgt))

		v := New()
		vu := newValidationUnit(eacl.RoleOthers, nil, tb)

		vu.op = eacl.OperationPut
		require.Equal(t, eacl.ActionDeny, v.CalculateAction(vu))

		vu.op = eacl.OperationGet
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))
	})

	t.Run("unknown operation", func(t *testing.T) {
		tb := eacl.NewTable()
		tb.AddRecord(newRecord(eacl.ActionDeny, eacl.OperationUnknown, tgt))
		tb.AddRecord(newRecord(eacl.ActionAllow, eacl.OperationGet, tgt))

		v := New()
		vu := newValidationUnit(eacl.RoleOthers, nil, tb)

		// TODO discuss if both next tests should result in DENY
		vu.op = eacl.OperationPut
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))

		vu.op = eacl.OperationGet
		require.Equal(t, eacl.ActionAllow, v.CalculateAction(vu))
	})
}

func TestTargetMatches(t *testing.T) {
	pubs := makeKeys(t, 3)

	tgt1 := eacl.NewTarget()
	tgt1.SetBinaryKeys(pubs[0:2])
	tgt1.SetRole(eacl.RoleUser)

	tgt2 := eacl.NewTarget()
	tgt2.SetRole(eacl.RoleOthers)

	r := eacl.NewRecord()
	r.SetTargets(tgt1, tgt2)

	u := newValidationUnit(eacl.RoleUser, pubs[0], nil)
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(eacl.RoleUser, pubs[2], nil)
	require.False(t, targetMatches(u, r))

	u = newValidationUnit(eacl.RoleUnknown, pubs[1], nil)
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(eacl.RoleOthers, pubs[2], nil)
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(eacl.RoleSystem, pubs[2], nil)
	require.False(t, targetMatches(u, r))
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

func (h headers) HeadersOfType(ht eacl.FilterHeaderType) ([]Header, bool) {
	switch ht {
	case eacl.HeaderFromRequest:
		return h.req, true
	case eacl.HeaderFromObject:
		return h.obj, true
	default:
		return nil, false
	}
}

func newRecord(a eacl.Action, op eacl.Operation, tgt ...*eacl.Target) *eacl.Record {
	r := eacl.NewRecord()
	r.SetAction(a)
	r.SetOperation(op)
	r.SetTargets(tgt...)
	return r
}

func newValidationUnit(role eacl.Role, key []byte, table *eacl.Table) *ValidationUnit {
	return new(ValidationUnit).
		WithRole(role).
		WithSenderKey(key).
		WithEACLTable(table)
}
