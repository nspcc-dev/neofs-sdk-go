package oidtest

import (
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// ID returns random oid.ID.
func ID() oid.ID {
	return oid.ID(testutil.RandByteSlice(oid.Size))
}

// OtherID returns random oid.ID other than any given one.
func OtherID(vs ...oid.ID) oid.ID {
loop:
	for {
		v := ID()
		for i := range vs {
			if v == vs[i] {
				continue loop
			}
		}
		return v
	}
}

// IDs returns n random oid.ID instances.
func IDs(n int) []oid.ID {
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

// OtherAddress returns random oid.Address other than any given one.
func OtherAddress(vs ...oid.Address) oid.Address {
loop:
	for {
		v := Address()
		for i := range vs {
			if v == vs[i] {
				continue loop
			}
		}
		return v
	}
}

// Addresses returns n random oid.Address instances.
func Addresses(n int) []oid.Address {
	res := make([]oid.Address, n)
	for i := range res {
		res[i] = Address()
	}
	return res
}
