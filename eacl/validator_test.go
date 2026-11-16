package eacl

import (
	"bytes"
	"errors"
	"fmt"
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
	require.True(t, ok)
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
	require.True(t, ok, msgAndArgs)
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

func TestUnavailableObjectFilters(t *testing.T) {
	v := NewValidator()
	targets := []Target{NewTargetByRole(RoleOthers)}
	hs := headers{obj: makeHeaders(
		FilterObjectContainerID, "container",
		FilterObjectID, "object",
		FilterObjectOwnerID, "owner",
		FilterObjectPayloadSize, "10",
		"classified", "secret",
	)}
	for _, op := range []Operation{
		OperationSearch, OperationDelete,
		OperationGet, OperationHead, OperationPut, OperationRange, OperationRangeHash,
	} {
		t.Run(op.String(), func(t *testing.T) {
			for _, tc := range []struct {
				name        string
				filters     []Filter
				unavailable bool
			}{
				{
					name:        "absent object attribute",
					filters:     []Filter{NewObjectPropertyFilter("missing", MatchNotPresent, "")},
					unavailable: true,
				},
				{
					name:        "supplied object attribute",
					filters:     []Filter{NewObjectPropertyFilter("classified", MatchStringEqual, "secret")},
					unavailable: true,
				},
				{
					name:        "object header field",
					filters:     []Filter{NewObjectPropertyFilter(FilterObjectOwnerID, MatchStringEqual, "owner")},
					unavailable: true,
				},
				{
					name:        "numeric object header field",
					filters:     []Filter{NewObjectPropertyFilter(FilterObjectPayloadSize, MatchNumGT, "9")},
					unavailable: true,
				},
				{
					name: "whole rule is inapplicable",
					filters: []Filter{
						NewObjectPropertyFilter(FilterObjectContainerID, MatchStringEqual, "container"),
						NewObjectPropertyFilter("missing", MatchNotPresent, ""),
					},
					unavailable: true,
				},
				{
					name:    "container ID",
					filters: []Filter{NewObjectPropertyFilter(FilterObjectContainerID, MatchStringEqual, "container")},
				},
				{
					name:    "object ID",
					filters: []Filter{NewObjectPropertyFilter(FilterObjectID, MatchStringEqual, "object")},
				},
				{
					name:    "absent request header",
					filters: []Filter{NewRequestHeaderFilter("missing", MatchNotPresent, "")},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					for _, action := range []Action{ActionAllow, ActionDeny} {
						t.Run(action.String(), func(t *testing.T) {
							fallback := ActionAllow
							if action == ActionAllow {
								fallback = ActionDeny
							}
							table := ConstructTable([]Record{
								ConstructRecord(action, op, targets, tc.filters...),
								ConstructRecord(fallback, op, targets),
							})
							var original Table
							table.CopyTo(&original)
							unit := newValidationUnit(RoleOthers, nil, &table).
								WithOperation(op).WithHeaderSource(hs)
							expected := action
							if tc.unavailable && (op == OperationSearch || op == OperationDelete) {
								expected = fallback
							}
							checkAction(t, expected, v, unit)
							require.Equal(t, original, table, "validation must not mutate the table")
							if tc.unavailable && (op == OperationSearch || op == OperationDelete) {
								// Inapplicable rules do not need a header source.
								checkAction(t, fallback, v, unit.WithHeaderSource(nil))
								singleRule := ConstructTable([]Record{table.Records()[0]})
								checkDefaultAction(t, v, unit.WithEACLTable(&singleRule))
							}
						})
					}
				})
			}
		})
	}
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

func (h headers) HeaderValue(ht FilterHeaderType, key string) ([]byte, error) {
	var hs []Header
	switch ht {
	case HeaderFromRequest:
		hs = h.req
	case HeaderFromObject:
		hs = h.obj
	default:
		return nil, nil
	}
	for _, header := range hs {
		if header != nil && header.Key() == key {
			return []byte(header.Value()), nil
		}
	}
	return nil, nil
}

type testHeaderSource func(FilterHeaderType, string) ([]byte, error)

func (f testHeaderSource) HeaderValue(typ FilterHeaderType, key string) ([]byte, error) {
	return f(typ, key)
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

func TestHeaderValueMatch(t *testing.T) {
	for _, tc := range []struct {
		name     string
		matcher  Match
		expected string
		value    []byte
		matches  bool
	}{
		{name: "equal", matcher: MatchStringEqual, expected: "secret", value: []byte("secret"), matches: true},
		{name: "unequal", matcher: MatchStringEqual, expected: "secret", value: []byte("public")},
		{name: "empty value", matcher: MatchStringEqual, value: []byte{}, matches: true},
		{name: "missing value", matcher: MatchStringEqual},
		{name: "not equal", matcher: MatchStringNotEqual, expected: "secret", value: []byte("public"), matches: true},
		{name: "not equal missing value", matcher: MatchStringNotEqual, expected: "secret"},
		{name: "absent", matcher: MatchNotPresent, matches: true},
		{name: "present empty", matcher: MatchNotPresent, value: []byte{}},
		{name: "present", matcher: MatchNotPresent, value: []byte("secret")},
		{name: "unknown matcher", matcher: Match(255), value: []byte("secret")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := testHeaderSource(func(typ FilterHeaderType, key string) ([]byte, error) {
				require.Equal(t, HeaderFromObject, typ)
				require.Equal(t, "classified", key)
				return tc.value, nil
			})
			table := ConstructTable([]Record{ConstructRecord(ActionDeny, OperationGet,
				[]Target{NewTargetByRole(RoleOthers)}, NewObjectPropertyFilter("classified", tc.matcher, tc.expected))})
			unit := newValidationUnit(RoleOthers, nil, &table).WithOperation(OperationGet).WithHeaderSource(source)
			expected := ActionAllow
			if tc.matches {
				expected = ActionDeny
			}
			checkAction(t, expected, NewValidator(), unit)
		})
	}
}

func TestTypedHeaderSourceValidation(t *testing.T) {
	targets := []Target{NewTargetByRole(RoleOthers)}
	deny := ConstructRecord(ActionDeny, OperationGet, targets,
		NewObjectPropertyFilter("classified", MatchStringEqual, "secret"))
	newUnit := func(records ...Record) *ValidationUnit {
		table := ConstructTable(records)
		return newValidationUnit(RoleOthers, nil, &table).WithOperation(OperationGet)
	}
	validator := NewValidator()
	unreadable := testHeaderSource(func(FilterHeaderType, string) ([]byte, error) {
		return nil, errors.New("unexpected header lookup")
	})

	t.Run("first matching record wins", func(t *testing.T) {
		allow := ConstructRecord(ActionAllow, OperationGet, targets)
		checkAction(t, ActionAllow, validator, newUnit(allow, deny).WithHeaderSource(unreadable))
	})
	t.Run("operation and target checked first", func(t *testing.T) {
		otherOp := deny
		otherOp.SetOperation(OperationPut)
		otherTarget := deny
		otherTarget.SetTargets(NewTargetByRole(RoleUser))
		checkDefaultAction(t, validator, newUnit(otherOp, otherTarget).WithHeaderSource(unreadable))
	})
	t.Run("mismatch does not hide later unavailable headers", func(t *testing.T) {
		rule := ConstructRecord(ActionDeny, OperationGet, targets,
			NewObjectPropertyFilter("classified", MatchStringEqual, "secret"),
			NewRequestHeaderFilter("probe", MatchStringEqual, "yes"))
		source := testHeaderSource(func(typ FilterHeaderType, _ string) ([]byte, error) {
			if typ == HeaderFromRequest {
				return nil, ErrHeaderUnavailable
			}
			return nil, nil
		})
		action, final, err := validator.CalculateAction(newUnit(rule).WithHeaderSource(source))
		require.NoError(t, err)
		require.Equal(t, ActionAllow, action)
		require.False(t, final)
	})
	t.Run("unfiltered rules do not allocate", func(t *testing.T) {
		unit := newUnit(ConstructRecord(ActionAllow, OperationGet, targets))
		allocs := testing.AllocsPerRun(100, func() {
			checkAction(t, ActionAllow, validator, unit)
		})
		require.Zero(t, allocs)
	})
	t.Run("string matching does not allocate", func(t *testing.T) {
		value := []byte("secret")
		source := testHeaderSource(func(FilterHeaderType, string) ([]byte, error) { return value, nil })
		for _, count := range []int{1, 20} {
			filters := make([]Filter, count)
			for i := range filters {
				filters[i] = NewObjectPropertyFilter("classified", MatchStringEqual, "secret")
			}
			unit := newUnit(ConstructRecord(ActionDeny, OperationGet, targets, filters...)).WithHeaderSource(source)
			allocs := testing.AllocsPerRun(100, func() {
				checkAction(t, ActionDeny, validator, unit)
			})
			require.Zero(t, allocs)
		}
	})
}

func TestTypedHeaderSourceErrors(t *testing.T) {
	errSource := errors.New("invalid header")
	for _, tc := range []struct {
		name   string
		err    error
		action Action
		want   error
	}{
		{"unavailable", ErrHeaderUnavailable, ActionAllow, nil},
		{"wrapped unavailable", fmt.Errorf("lookup: %w", ErrHeaderUnavailable), ActionAllow, nil},
		{"invalid", errSource, ActionDeny, errSource},
		{"wrapped invalid", fmt.Errorf("lookup: %w", errSource), ActionDeny, errSource},
	} {
		for _, matcher := range []Match{MatchStringEqual, MatchNotPresent, MatchNumGT, Match(255)} {
			t.Run(fmt.Sprintf("%s/%s", tc.name, matcher), func(t *testing.T) {
				source := testHeaderSource(func(FilterHeaderType, string) ([]byte, error) { return nil, tc.err })
				table := ConstructTable([]Record{ConstructRecord(ActionDeny, OperationGet,
					[]Target{NewTargetByRole(RoleOthers)}, NewObjectPropertyFilter("classified", matcher, "invalid"))})
				unit := newValidationUnit(RoleOthers, nil, &table).WithOperation(OperationGet).WithHeaderSource(source)
				action, final, err := NewValidator().CalculateAction(unit)
				require.ErrorIs(t, err, tc.want)
				require.Equal(t, tc.action, action)
				require.False(t, final)
			})
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
