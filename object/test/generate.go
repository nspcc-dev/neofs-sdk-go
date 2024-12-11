package objecttest

import (
	"math/rand"
	"strconv"

	"github.com/google/uuid"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Range returns random object.Range.
func Range() object.Range {
	x := object.NewRange()

	x.SetOffset(1024)
	x.SetLength(2048)

	return *x
}

// Attribute returns random object.Attribute.
func Attribute() object.Attribute {
	x := object.NewAttribute("key", "value")

	return *x
}

// SplitID returns random object.SplitID.
func SplitID() object.SplitID {
	x := object.NewSplitID()

	x.SetUUID(uuid.New())

	return *x
}

func generate(withParent bool) object.Object {
	x := object.New()
	ver := version.Current()

	x.SetID(oidtest.ID())
	tok := sessiontest.ObjectSigned(usertest.User())
	x.SetSessionToken(&tok)
	x.SetPayload([]byte{1, 2, 3})
	x.SetOwner(usertest.ID())
	x.SetContainerID(cidtest.ID())
	x.SetType(object.TypeTombstone)
	x.SetVersion(&ver)
	x.SetPayloadSize(111)
	x.SetCreationEpoch(222)
	x.SetPreviousID(oidtest.ID())
	x.SetParentID(oidtest.ID())
	x.SetChildren(oidtest.ID(), oidtest.ID())
	as := make([]object.Attribute, 1+rand.Int()%4)
	for i := range as {
		si := strconv.Itoa(i)
		as[i].SetKey("key_" + si)
		as[i].SetValue("val_" + si)
	}
	x.SetAttributes(as...)
	splitID := SplitID()
	x.SetSplitID(&splitID)
	x.SetPayloadChecksum(checksumtest.Checksum())
	x.SetPayloadHomomorphicHash(checksumtest.Checksum())

	if withParent {
		par := generate(false)
		x.SetParent(&par)
	}

	return *x
}

// Raw returns random object.Object.
// Deprecated: (v1.0.0) use Object instead.
func Raw() object.Object {
	return Object()
}

// Object returns random object.Object.
func Object() object.Object {
	return generate(true)
}

// Tombstone returns random object.Tombstone.
func Tombstone() object.Tombstone {
	x := object.NewTombstone()

	splitID := SplitID()
	x.SetSplitID(&splitID)
	x.SetExpirationEpoch(13)
	x.SetMembers([]oid.ID{oidtest.ID(), oidtest.ID()})

	return *x
}

// SplitInfo returns random object.SplitInfo.
func SplitInfo() object.SplitInfo {
	x := object.NewSplitInfo()

	splitID := SplitID()
	x.SetSplitID(&splitID)
	x.SetLink(oidtest.ID())
	x.SetLastPart(oidtest.ID())

	return *x
}

// SearchFilters returns random object.SearchFilters.
func SearchFilters() object.SearchFilters {
	x := object.NewSearchFilters()

	x.AddObjectIDFilter(object.MatchStringEqual, oidtest.ID())
	x.AddObjectContainerIDFilter(object.MatchStringNotEqual, cidtest.ID())

	return x
}

// Lock returns random object.Lock.
func Lock() *object.Lock {
	var l object.Lock
	l.WriteMembers([]oid.ID{oidtest.ID(), oidtest.ID()})

	return &l
}

// Link returns random object.Link.
func Link() *object.Link {
	ms := make([]object.MeasuredObject, rand.Int()%10)
	for i := range ms {
		ms[i].SetObjectID(oidtest.ID())
		ms[i].SetObjectSize(rand.Uint32())
	}
	var l object.Link
	l.SetObjects(ms)
	return &l
}
