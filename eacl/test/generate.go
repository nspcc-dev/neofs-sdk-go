package eacltest

import (
	"math/rand/v2"
	"strconv"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// Target returns random eacl.Target.
func Target() eacl.Target {
	if rand.Int()%2 == 0 {
		return eacl.NewTargetByRole(eacl.Role(rand.Uint32()))
	}
	return eacl.NewTargetByAccounts(usertest.IDs(1 + rand.N(10)))
}

// Targets returns n random eacl.Target instances.
func Targets(n int) []eacl.Target {
	res := make([]eacl.Target, n)
	for i := range res {
		res[i] = Target()
	}
	return res
}

// Filter returns random eacl.Filter.
func Filter() eacl.Filter {
	si := strconv.Itoa(rand.Int())
	return eacl.ConstructFilter(eacl.FilterHeaderType(rand.Uint32()), "key_"+si, eacl.Match(rand.Uint32()), "val_"+si)
}

// Filters returns n random eacl.Filter instances.
func Filters(n int) []eacl.Filter {
	res := make([]eacl.Filter, n)
	for i := range res {
		res[i] = Filter()
	}
	return res
}

// Record returns random eacl.Record.
func Record() eacl.Record {
	tn := 1 + rand.N(10)
	fn := 1 + rand.N(10)
	return eacl.ConstructRecord(eacl.Action(rand.Uint32()), eacl.Operation(rand.Uint32()), Targets(tn), Filters(fn)...)
}

// Records returns n random eacl.Record instances.
func Records(n int) []eacl.Record {
	res := make([]eacl.Record, n)
	for i := range res {
		res[i] = Record()
	}
	return res
}

// Table returns random eacl.Table.
func Table() eacl.Table {
	n := 1 + rand.N(10)
	return eacl.NewTableForContainer(cidtest.ID(), Records(n))
}
