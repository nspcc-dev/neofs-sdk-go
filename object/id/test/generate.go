package oidtest

import (
	"math/rand"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// ID returns random oid.ID.
func ID() oid.ID {
	var res oid.ID
	//nolint:staticcheck
	rand.Read(res[:])
	return res
}

// ChangeID returns object ID other than the given one.
func ChangeID(id oid.ID) oid.ID {
	id[0]++
	return id
}

// NIDs returns n random oid.ID instances.
func NIDs(n int) []oid.ID {
	res := make([]oid.ID, n)
	for i := range res {
		res[i] = ID()
	}
	return res
}

// Address returns random oid.Address.
func Address() oid.Address {
	var x oid.Address

	x.SetContainer(cidtest.ID())
	x.SetObject(ID())

	return x
}

// ChangeAddress returns object address other than the given one.
func ChangeAddress(addr oid.Address) oid.Address {
	var res oid.Address
	res.SetObject(ChangeID(addr.Object()))
	res.SetContainer(cidtest.ChangeID(addr.Container()))
	return res
}

// NAddresses returns n random oid.Address instances.
func NAddresses(n int) []oid.Address {
	res := make([]oid.Address, n)
	for i := range res {
		res[i] = Address()
	}
	return res
}
