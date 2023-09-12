package eacl_test

import (
	"fmt"
	"testing"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/stretchr/testify/require"
)

func TestNewTarget(t *testing.T) {
	validRoles := []eacl.Role{eacl.RoleContainerOwner, eacl.RoleOthers}
	validKeys := []neofscrypto.PublicKey{test.RandomSigner(t).Public(), test.RandomSignerRFC6979(t).Public()}

	for i, tc := range []struct {
		rs []eacl.Role
		ks []neofscrypto.PublicKey
	}{
		{rs: nil, ks: nil},
		{rs: nil, ks: []neofscrypto.PublicKey{}},
		{rs: []eacl.Role{}, ks: nil},
		{rs: []eacl.Role{}, ks: []neofscrypto.PublicKey{}},
		{rs: validRoles, ks: []neofscrypto.PublicKey{nil}},
		{rs: []eacl.Role{0}, ks: validKeys},
	} {
		t.Run("invalid input", func(t *testing.T) {
			require.Panics(t, func() { eacl.NewTarget(tc.rs, tc.ks) }, i)
		})
	}

	for i, role := range []eacl.Role{
		eacl.RoleContainerOwner,
		eacl.RoleOthers,
	} {
		t.Run(fmt.Sprintf("support role=%v", role), func(t *testing.T) {
			require.NotPanics(t, func() { eacl.NewTarget([]eacl.Role{role}, validKeys) }, i)
		})
	}

	t.Run("system role", func(t *testing.T) {
		require.Panics(t, func() { eacl.NewTarget([]eacl.Role{eacl.RoleSystem}, validKeys) })
	})
}
