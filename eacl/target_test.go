package eacl_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/util"
	protoacl "github.com/nspcc-dev/neofs-api-go/v2/acl"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestTarget_ToV2(t *testing.T) {
	r := eacl.NewTargetByRole(anyValidRole)
	subjs := [][]byte{
		anyValidECDSABinPublicKeys[0],
		anyUserSet[0][:],
		anyValidECDSABinPublicKeys[1],
		anyUserSet[1][:],
		anyUserSet[2][:],
	}
	r.SetRawSubjects(subjs)
	m := r.ToV2()
	require.EqualValues(t, anyValidRole, m.GetRole())
	require.Equal(t, subjs, m.GetKeys())

	t.Run("default values", func(t *testing.T) {
		target := eacl.NewTarget()

		// check initial values
		require.Zero(t, target.Role())
		require.Nil(t, target.BinaryKeys())

		// convert to v2 message
		targetV2 := target.ToV2()

		require.Equal(t, protoacl.RoleUnknown, targetV2.GetRole())
		require.Nil(t, targetV2.GetKeys())
	})
}

func TestNewTargetFromV2(t *testing.T) {
	role := protoacl.Role(rand.Uint32())
	var m protoacl.Target
	m.SetRole(role)
	m.SetKeys(anyValidBinPublicKeys)

	r := eacl.NewTargetFromV2(&m)
	require.EqualValues(t, role, r.Role())
	require.Equal(t, anyValidBinPublicKeys, m.GetKeys())

	t.Run("nil", func(t *testing.T) {
		require.Equal(t, new(eacl.Target), eacl.NewTargetFromV2(nil))
	})
}

func TestTarget_Marshal(t *testing.T) {
	for i := range anyValidTargets {
		require.Equal(t, anyValidBinTargets[i], anyValidTargets[i].Marshal())
	}
}

func TestTarget_Unmarshal(t *testing.T) {
	t.Run("invalid protobuf", func(t *testing.T) {
		err := new(eacl.Target).Unmarshal([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "cannot parse invalid wire-format data")
	})

	var tgt eacl.Target
	for i := range anyValidBinTargets {
		err := tgt.Unmarshal(anyValidBinTargets[i])
		require.NoError(t, err)
		require.Equal(t, anyValidTargets[i], tgt)
	}
}

func TestTarget_MarshalJSON(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		err := new(eacl.Target).UnmarshalJSON([]byte("Hello, world!"))
		require.ErrorContains(t, err, "proto")
		require.ErrorContains(t, err, "syntax error")
	})

	var tgt1, tgt2 eacl.Target
	for i := range anyValidTargets {
		b, err := anyValidTargets[i].MarshalJSON()
		require.NoError(t, err, i)
		require.NoError(t, tgt1.UnmarshalJSON(b), i)
		require.Equal(t, anyValidTargets[i], tgt1, i)

		b, err = json.Marshal(anyValidTargets[i])
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &tgt2), i)
		require.Equal(t, anyValidTargets[i], tgt2, i)
	}
}

func TestTarget_UnmarshalJSON(t *testing.T) {
	var tgt1, tgt2 eacl.Target
	for i := range anyValidJSONTargets {
		require.NoError(t, tgt1.UnmarshalJSON([]byte(anyValidJSONTargets[i])), i)
		require.Equal(t, anyValidTargets[i], tgt1, i)

		require.NoError(t, json.Unmarshal([]byte(anyValidJSONTargets[i]), &tgt2), i)
		require.Equal(t, anyValidTargets[i], tgt2, i)
	}
}

func TestTarget_SetRole(t *testing.T) {
	var tgt eacl.Target
	require.Zero(t, tgt.Role())

	tgt.SetRole(anyValidRole)
	require.Equal(t, anyValidRole, tgt.Role())

	otherRole := anyValidRole + 1
	tgt.SetRole(otherRole)
	require.Equal(t, otherRole, tgt.Role())
}

func TestTarget_SetBinaryKeys(t *testing.T) {
	var tgt eacl.Target
	require.Zero(t, tgt.BinaryKeys())

	ks := make([][]byte, 3)
	for i := range ks {
		ks[i] = make([]byte, 33)
		//nolint:staticcheck
		rand.Read(ks[i])
	}
	tgt.SetBinaryKeys(ks)
	require.Equal(t, ks, tgt.BinaryKeys())

	otherKeys := make([][]byte, 3)
	for i := range otherKeys {
		otherKeys[i] = make([]byte, 33)
		//nolint:staticcheck
		rand.Read(otherKeys[i])
	}
	tgt.SetBinaryKeys(otherKeys)
	require.Equal(t, otherKeys, tgt.BinaryKeys())
}

func TestTargetByRole(t *testing.T) {
	tgt := eacl.NewTargetByRole(anyValidRole)
	require.Equal(t, anyValidRole, tgt.Role())
	require.Zero(t, tgt.Accounts())
}

func TestNewTargetByAccounts(t *testing.T) {
	accs := usertest.IDs(5)
	tgt := eacl.NewTargetByAccounts(accs)
	require.Equal(t, accs, tgt.Accounts())
	require.Zero(t, tgt.Role())
}

func randomScriptHashes(n int) []util.Uint160 {
	hs := make([]util.Uint160, n)
	for i := range hs {
		//nolint:staticcheck
		rand.Read(hs[i][:])
	}
	return hs
}

func assertUsersMatchScriptHashes(t testing.TB, usrs []user.ID, hs []util.Uint160) {
	require.Len(t, usrs, len(hs))
	for i := range usrs {
		require.EqualValues(t, 0x35, usrs[i][0])
		require.Equal(t, hs[i][:], usrs[i][1:21])
		require.Equal(t, hash.Checksum(usrs[i][:21])[:4], usrs[i][21:])
	}
}

func TestNewTargetByScriptHashes(t *testing.T) {
	hs := randomScriptHashes(3)
	tgt := eacl.NewTargetByScriptHashes(hs)
	assertUsersMatchScriptHashes(t, tgt.Accounts(), hs)
}

func TestSetTargetAccounts(t *testing.T) {
	hs := randomScriptHashes(3)
	var tgt eacl.Target
	eacl.SetTargetAccounts(&tgt, hs...)
	assertUsersMatchScriptHashes(t, tgt.Accounts(), hs)
}

func TestSetTargetECDSAKeys(t *testing.T) {
	var tgt eacl.Target
	require.Zero(t, tgt.BinaryKeys())
	eacl.SetTargetECDSAKeys(&tgt)
	require.Zero(t, tgt.BinaryKeys())

	eacl.SetTargetECDSAKeys(&tgt, anyECDSAPublicKeysPtr...)
	require.Equal(t, anyValidECDSABinPublicKeys, tgt.BinaryKeys())
}

func TestTargetECDSAKeys(t *testing.T) {
	var tgt eacl.Target
	require.Empty(t, eacl.TargetECDSAKeys(&tgt))

	tgt.SetBinaryKeys(anyValidECDSABinPublicKeys)
	require.Equal(t, anyECDSAPublicKeysPtr, eacl.TargetECDSAKeys(&tgt))
}

func TestTarget_SetRawSubjects(t *testing.T) {
	var tgt eacl.Target
	require.Zero(t, tgt.RawSubjects())
	require.Zero(t, tgt.Accounts())
	require.Zero(t, tgt.BinaryKeys())

	garbageSubjs := [][]byte{[]byte("foo"), []byte("bar")}
	tgt.SetRawSubjects(garbageSubjs)
	require.Equal(t, garbageSubjs, tgt.RawSubjects())
	require.Zero(t, tgt.Accounts())
	require.Zero(t, tgt.BinaryKeys())

	subjs := [][]byte{
		garbageSubjs[0],
		make([]byte, 33),
		nil,
		garbageSubjs[1],
		nil,
		make([]byte, 33),
	}
	//nolint:staticcheck
	rand.Read(subjs[1])
	//nolint:staticcheck
	rand.Read(subjs[5])
	usrs := usertest.IDs(2)
	subjs[2] = usrs[0][:]
	subjs[4] = usrs[1][:]

	tgt.SetRawSubjects(subjs)
	require.Equal(t, subjs, tgt.RawSubjects())
	require.Equal(t, usrs, tgt.Accounts())
	require.Equal(t, [][]byte{subjs[1], subjs[5]}, tgt.BinaryKeys())
}
