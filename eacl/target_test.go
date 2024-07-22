package eacl_test

import (
	"crypto/ecdsa"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestTarget(t *testing.T) {
	pubs := []*ecdsa.PublicKey{
		randomPublicKey(t),
		randomPublicKey(t),
	}

	target := eacl.NewTarget()
	target.SetRole(eacl.RoleSystem)
	eacl.SetTargetECDSAKeys(target, pubs...)

	v2 := target.ToV2()
	require.NotNil(t, v2)
	require.Equal(t, v2acl.RoleSystem, v2.GetRole())
	require.Len(t, v2.GetKeys(), len(pubs))
	for i, key := range v2.GetKeys() {
		require.Equal(t, key, (*keys.PublicKey)(pubs[i]).Bytes())
	}

	newTarget := eacl.NewTargetFromV2(v2)
	require.Equal(t, target, newTarget)

	t.Run("from nil v2 target", func(t *testing.T) {
		require.Equal(t, new(eacl.Target), eacl.NewTargetFromV2(nil))
	})
}

func TestTargetAccounts(t *testing.T) {
	accs := []util.Uint160{
		(*keys.PublicKey)(randomPublicKey(t)).GetScriptHash(),
		(*keys.PublicKey)(randomPublicKey(t)).GetScriptHash(),
	}

	target := eacl.NewTarget()
	target.SetRole(eacl.RoleSystem)
	eacl.SetTargetAccounts(target, accs...)

	v2 := target.ToV2()
	require.NotNil(t, v2)
	require.Equal(t, v2acl.RoleSystem, v2.GetRole())
	require.Len(t, v2.GetKeys(), len(accs))
	for i, key := range v2.GetKeys() {
		var u = user.NewFromScriptHash(accs[i])
		require.Equal(t, key, u[:])
	}

	newTarget := eacl.NewTargetFromV2(v2)
	require.Equal(t, target, newTarget)

	t.Run("from nil v2 target", func(t *testing.T) {
		require.Equal(t, new(eacl.Target), eacl.NewTargetFromV2(nil))
	})
}

func TestTargetUsers(t *testing.T) {
	accs := usertest.IDs(2)

	target := eacl.NewTarget()
	target.SetRole(eacl.RoleSystem)
	target.SetAccounts(accs)

	v2 := target.ToV2()
	require.NotNil(t, v2)
	require.Equal(t, v2acl.RoleSystem, v2.GetRole())
	require.Len(t, v2.GetKeys(), len(accs))
	for i, key := range v2.GetKeys() {
		require.Equal(t, key, accs[i][:])
	}

	newTarget := eacl.NewTargetFromV2(v2)
	require.Equal(t, target, newTarget)

	t.Run("from nil v2 target", func(t *testing.T) {
		require.Equal(t, new(eacl.Target), eacl.NewTargetFromV2(nil))
	})
}

func TestTargetEncoding(t *testing.T) {
	tar := eacl.NewTarget()
	tar.SetRole(eacl.RoleSystem)
	eacl.SetTargetECDSAKeys(tar, randomPublicKey(t))

	t.Run("binary", func(t *testing.T) {
		tar2 := eacl.NewTarget()
		require.NoError(t, tar2.Unmarshal(tar.Marshal()))

		require.Equal(t, tar, tar2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := tar.MarshalJSON()
		require.NoError(t, err)

		tar2 := eacl.NewTarget()
		require.NoError(t, tar2.UnmarshalJSON(data))

		require.Equal(t, tar, tar2)
	})
}

func TestTarget_ToV2(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		target := eacl.NewTarget()

		// check initial values
		require.Zero(t, target.Role())
		require.Nil(t, target.BinaryKeys())

		// convert to v2 message
		targetV2 := target.ToV2()

		require.Equal(t, acl.RoleUnknown, targetV2.GetRole())
		require.Nil(t, targetV2.GetKeys())
	})
}

func TestTargetByRole(t *testing.T) {
	r := eacl.Role(rand.Uint32())
	tgt := eacl.NewTargetByRole(r)
	require.Equal(t, r, tgt.Role())
	require.Zero(t, tgt.Accounts())
}

func TestNewTargetByAccounts(t *testing.T) {
	accs := usertest.IDs(5)
	tgt := eacl.NewTargetByAccounts(accs)
	require.Equal(t, accs, tgt.Accounts())
	require.Zero(t, tgt.Role())
}

func TestNewTargetByScriptHashes(t *testing.T) {
	hs := make([]util.Uint160, 5)
	for i := range hs {
		//nolint:staticcheck
		rand.Read(hs[i][:])
	}
	tgt := eacl.NewTargetByScriptHashes(hs)
	accs := tgt.Accounts()
	require.Len(t, accs, len(hs))
	for i := range accs {
		require.EqualValues(t, 0x35, accs[i][0])
		require.Equal(t, hs[i][:], accs[i][1:21])
		require.Equal(t, hash.Checksum(accs[i][:21])[:4], accs[i][21:])
	}
}
