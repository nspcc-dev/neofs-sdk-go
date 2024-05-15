package eacltest

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
)

// Target returns random [eacl.Target].
func Target() eacl.Target {
	var x eacl.Target
	if rand.Int()%2 == 0 {
		ks := make([][]byte, 3)
		for j := range ks {
			ks[j] = make([]byte, 33)
			rand.Read(ks[j])
		}
		x.SetPublicKeys(ks)
	} else {
		x.SetRole(eacl.Role(rand.Int()))
	}
	return x
}

// NTargets returns n random [eacl.Target] instances.
func NTargets(n int) []eacl.Target {
	res := make([]eacl.Target, n)
	for i := range res {
		res[i] = Target()
	}
	return res
}

// Filter returns random [eacl.Filter].
func Filter() eacl.Filter {
	var x eacl.Filter
	x.SetKey("key_" + strconv.Itoa(rand.Int()))
	x.SetValue("val_" + strconv.Itoa(rand.Int()))
	x.SetAttributeType(eacl.AttributeType(rand.Int()))
	x.SetMatcher(eacl.Match(rand.Int()))
	return x
}

// NFilters returns n random [eacl.Filter] instances.
func NFilters(n int) []eacl.Filter {
	res := make([]eacl.Filter, n)
	for i := range res {
		res[i] = Filter()
	}
	return res
}

// Record returns random [eacl.Record].
func Record() eacl.Record {
	var x eacl.Record
	x.SetAction(eacl.Action(rand.Int()))
	x.SetOperation(acl.Op(rand.Int()))
	x.SetTargets(NTargets(1 + rand.Int()%3))
	if n := rand.Int() % 4; n > 0 {
		x.SetFilters(NFilters(n))
	}
	return x
}

// NRecords returns n random [eacl.Record] instances.
func NRecords(n int) []eacl.Record {
	res := make([]eacl.Record, n)
	for i := range res {
		res[i] = Record()
	}
	return res
}

// Table returns random [eacl.Table].
func Table() eacl.Table {
	var x eacl.Table
	x.LimitToContainer(cidtest.ID())
	x.SetRecords(NRecords(1 + rand.Int()%3))

	if err := x.Unmarshal(x.Marshal()); err != nil { // to fill utility fields
		panic(fmt.Errorf("unexpected eACL encode-decode failure: %w", err))
	}

	return x
}
