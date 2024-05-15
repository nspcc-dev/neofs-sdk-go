package objecttest

import (
	"math/rand"
	"strconv"

	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

func header(withParent bool) object.Header {
	h := object.New(cidtest.ID(), usertest.ID()).Header
	h.SetSessionToken(sessiontest.Object())
	h.SetType(object.Type(rand.Uint32()))
	h.SetPayloadSize(rand.Uint64())
	h.SetCreationEpoch(rand.Uint64())
	h.SetPreviousSplitObject(oidtest.ID())
	h.SetFirstSplitObject(oidtest.ID())
	h.SetParentID(oidtest.ID())
	h.SetParentSignature(neofscryptotest.Signature())
	h.SetPayloadChecksum(checksumtest.Checksum())
	h.SetPayloadHomomorphicChecksum(checksumtest.Checksum())

	nAttr := rand.Int() % 4
	for i := 0; i < nAttr; i++ {
		si := strconv.Itoa(rand.Int())
		h.SetAttribute("attr_"+si, "val_"+si)
	}

	if withParent {
		h.SetParentHeader(header(false))
	}

	return h
}

// Header returns random object.Header.
func Header() object.Header {
	return header(true)
}

// Object returns random object.Object.
func Object() object.Object {
	payload := make([]byte, rand.Int()%32)
	rand.Read(payload)

	obj := object.Object{Header: Header()}
	obj.SetID(oidtest.ID())
	obj.SetSignature(neofscryptotest.Signature())
	obj.SetPayload(payload)
	return obj
}

// Tombstone returns random object.Tombstone.
func Tombstone() object.Tombstone {
	var x object.Tombstone
	x.SetMembers(oidtest.NIDs(rand.Int()%3 + 1))
	return x
}

// SplitInfo returns random object.SplitInfo.
func SplitInfo() object.SplitInfo {
	var x object.SplitInfo
	x.SetFirstPart(oidtest.ID())
	x.SetLastPart(oidtest.ID())
	x.SetLinker(oidtest.ID())
	return x
}

// Lock returns random object.Lock.
func Lock() object.Lock {
	var l object.Lock
	l.SetList(oidtest.NIDs(rand.Int()%3 + 1))
	return l
}

// SplitChain returns random object.SplitChain.
func SplitChain() object.SplitChain {
	els := make([]object.SplitChainElement, rand.Int()%3+1)
	for i := range els {
		els[i].SetID(oidtest.ID())
		els[i].SetPayloadSize(rand.Uint32())
	}
	var x object.SplitChain
	x.SetElements(els)
	return x
}
