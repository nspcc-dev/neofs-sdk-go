package eacl

import (
	"crypto/sha256"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util/slice"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func TestTable_CopyTo(t *testing.T) {
	const srcRecordAction = ActionAllow
	const srcRecordOp = acl.OpObjectGet
	srcCnr := cidtest.ID()

	src := NewForContainer(srcCnr, []Record{
		NewRecord(srcRecordAction, srcRecordOp, NewTarget([]Role{
			RoleContainerOwner,
			RoleOthers,
		}, []neofscrypto.PublicKey{
			test.RandomPublicKey(),
			test.RandomPublicKey(),
		}),
			NewFilter(HeaderFromRequest, "any_key", MatchStringEqual, "any_val"),
		),
	})

	srcTarget := Target{
		roles: make([]Role, len(src.records[0].target.roles)),
		keys:  make([][]byte, len(src.records[0].target.keys)),
	}

	copy(srcTarget.roles, src.records[0].target.roles)
	for i := range src.records[0].target.keys {
		srcTarget.keys[i] = slice.Copy(src.records[0].target.keys[i])
	}

	srcFilters := make([]Filter, len(src.records[0].filters))
	copy(srcFilters, src.records[0].filters)

	var dst Table
	src.CopyTo(&dst)

	require.Equal(t, src, dst)

	changeAllTableFields(&dst)

	require.Equal(t, version.Current(), *src.version)
	require.Equal(t, srcCnr, *src.cid)
	require.Equal(t, srcRecordAction, src.records[0].action)
	require.Equal(t, srcRecordOp, src.records[0].operation)
	require.Equal(t, srcTarget, src.records[0].target)
	require.Equal(t, srcFilters, src.records[0].filters)

	t.Run("full-to-full", func(t *testing.T) {
		src.CopyTo(&dst)
		require.Equal(t, src, dst)
	})

	t.Run("zero-to-full", func(t *testing.T) {
		var zero Table
		zero.CopyTo(&src)
		require.Zero(t, src)
	})
}

func changeAllTableFields(t *Table) {
	bCnr := make([]byte, sha256.Size)
	t.cid.Encode(bCnr)

	bCnr[0]++

	err := t.cid.Decode(bCnr)
	if err != nil {
		panic(err)
	}

	t.version.SetMajor(t.version.Major() + 1)
	t.version.SetMinor(t.version.Minor() + 1)

	for i := range t.records {
		changeAllRecordFields(&t.records[i])
	}
}

func changeAllRecordFields(r *Record) {
	r.action++
	r.operation++
	changeAllTargetFields(&r.target)
	for i := range r.filters {
		changeAllFilterFields(&r.filters[i])
	}
}

func changeAllFilterFields(f *Filter) {
	f.hdrType++
	f.matcher++
	f.key += "1"
	f.value += "1"
}

func changeAllTargetFields(t *Target) {
	for i := range t.roles {
		t.roles[i]++
	}

	for i := range t.keys {
		t.keys[i][0]++
	}
}
