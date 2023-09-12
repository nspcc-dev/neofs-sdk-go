package eacltest

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
)

// Target returns random eacl.Target.
func Target(tb testing.TB) eacl.Target {
	roles := []eacl.Role{eacl.RoleContainerOwner, eacl.RoleOthers}
	keys := []neofscrypto.PublicKey{test.RandomSigner(tb).Public(), test.RandomSignerRFC6979(tb).Public()}

	return eacl.NewTarget(roles, keys)
}

// Record returns random eacl.Record.
func Record(tb testing.TB) eacl.Record {
	fNum := rand.Int() % 10
	var fs []eacl.Filter
	if fNum > 0 {
		fs = make([]eacl.Filter, fNum)
		for i := range fs {
			fs[i] = eacl.NewFilter(eacl.HeaderFromObject, "key"+strconv.Itoa(i), eacl.MatchStringEqual, "val"+strconv.Itoa(i))
		}
	}

	return eacl.NewRecord(eacl.ActionAllow, acl.OpObjectHash, eacl.NewTargetWithKey(test.RandomSigner(tb).Public()), fs...)
}

func Table(tb testing.TB) eacl.Table {
	t := eacl.NewForContainer(cidtest.ID(), []eacl.Record{
		Record(tb),
		Record(tb),
	})

	return t
}
