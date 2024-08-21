package eacl

import (
	"math/rand/v2"
	"testing"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/stretchr/testify/require"
)

func TestValidationUnit_WithContainerID(t *testing.T) {
	cnr := cidtest.ID()
	u := new(ValidationUnit).WithContainerID(&cnr)
	require.Equal(t, cnr, *u.cid)
}

func TestValidationUnit_WithRole(t *testing.T) {
	role := Role(rand.Uint32())
	u := new(ValidationUnit).WithRole(role)
	require.Equal(t, role, u.role)
}

func TestValidationUnit_WithOperation(t *testing.T) {
	op := Operation(rand.Uint32())
	u := new(ValidationUnit).WithOperation(op)
	require.Equal(t, op, u.op)
}

func TestValidationUnit_WithHeaderSource(t *testing.T) {
	hdrs := new(headers)
	u := new(ValidationUnit).WithHeaderSource(hdrs)
	require.Equal(t, hdrs, u.hdrSrc)
}

func TestValidationUnit_WithSenderKey(t *testing.T) {
	key := []byte("any_key")
	u := new(ValidationUnit).WithSenderKey(key)
	require.Equal(t, key, u.key)
}

func TestValidationUnit_WithEACLTable(t *testing.T) {
	eACL := NewTableForContainer(cidtest.ID(), []Record{
		ConstructRecord(Action(rand.Uint32()), Operation(rand.Uint32()), []Target{
			NewTargetByRole(Role(rand.Uint32())),
		}),
	})
	u := new(ValidationUnit).WithEACLTable(&eACL)
	require.Equal(t, eACL, *u.table)
}
