package objecttest

import (
	"testing"

	"github.com/google/uuid"
	objecttest "github.com/nspcc-dev/neofs-api-go/v2/object/test"
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

func generate(t testing.TB, withParent bool) object.Object {
	x := object.New()
	ver := version.Current()

	x.SetID(oidtest.ID())
	tok := sessiontest.Object()
	x.SetSessionToken(&tok)
	x.SetPayload([]byte{1, 2, 3})
	owner := usertest.ID(t)
	x.SetOwnerID(&owner)
	x.SetContainerID(cidtest.ID())
	x.SetType(object.TypeTombstone)
	x.SetVersion(&ver)
	x.SetPayloadSize(111)
	x.SetCreationEpoch(222)
	x.SetPreviousID(oidtest.ID())
	x.SetParentID(oidtest.ID())
	x.SetChildren(oidtest.ID(), oidtest.ID())
	x.SetAttributes(Attribute(), Attribute())
	splitID := SplitID()
	x.SetSplitID(&splitID)
	x.SetPayloadChecksum(checksumtest.Checksum())
	x.SetPayloadHomomorphicHash(checksumtest.Checksum())

	if withParent {
		par := generate(t, false)
		x.SetParent(&par)
	}

	return *x
}

// Raw returns random object.Object.
// Deprecated: (v1.0.0) use Object instead.
func Raw(t testing.TB) object.Object {
	return Object(t)
}

// Object returns random object.Object.
func Object(t testing.TB) object.Object {
	return generate(t, true)
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
	return (*object.Link)(objecttest.GenerateLink(false))
}
