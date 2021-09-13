package eacl

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func newValidationUnit(role Role, key []byte) *ValidationUnit {
	return &ValidationUnit{
		role: role,
		key:  key,
	}
}
