package eacl

import (
	"math/rand"
	"testing"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestOperationMatch(t *testing.T) {
	tgt := NewTarget()
	tgt.SetRole(RoleOthers)

	t.Run("single operation", func(t *testing.T) {
		tb := NewTable()
		tb.AddRecord(newRecord(ActionDeny, OperationPut, tgt))
		tb.AddRecord(newRecord(ActionAllow, OperationGet, tgt))

		v := newValidator(t, tb)
		vu := newValidationUnit(RoleOthers, nil)

		vu.op = OperationPut
		require.Equal(t, ActionDeny, v.CalculateAction(vu))

		vu.op = OperationGet
		require.Equal(t, ActionAllow, v.CalculateAction(vu))
	})

	t.Run("unknown operation", func(t *testing.T) {
		tb := NewTable()
		tb.AddRecord(newRecord(ActionDeny, OperationUnknown, tgt))
		tb.AddRecord(newRecord(ActionAllow, OperationGet, tgt))

		v := newValidator(t, tb)
		vu := newValidationUnit(RoleOthers, nil)

		// TODO discuss if both next tests should result in DENY
		vu.op = OperationPut
		require.Equal(t, ActionAllow, v.CalculateAction(vu))

		vu.op = OperationGet
		require.Equal(t, ActionAllow, v.CalculateAction(vu))
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
	r.SetTargets(tgt1, tgt2)

	u := newValidationUnit(RoleUser, pubs[0])
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(RoleUser, pubs[2])
	require.False(t, targetMatches(u, r))

	u = newValidationUnit(RoleUnknown, pubs[1])
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(RoleOthers, pubs[2])
	require.True(t, targetMatches(u, r))

	u = newValidationUnit(RoleSystem, pubs[2])
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

func newRecord(a Action, op Operation, tgt ...*Target) *Record {
	r := NewRecord()
	r.SetAction(a)
	r.SetOperation(op)
	r.SetTargets(tgt...)
	return r
}

type dummySource struct {
	tb *Table
}

func (d dummySource) GetEACL(*cid.ID) (*Table, error) {
	return d.tb, nil
}

func newValidator(t *testing.T, tb *Table) *Validator {
	return NewValidator(
		WithLogger(zaptest.NewLogger(t)),
		WithEACLSource(dummySource{tb}))
}

func newValidationUnit(role Role, key []byte) *ValidationUnit {
	return &ValidationUnit{
		role: role,
		key:  key,
	}
}
