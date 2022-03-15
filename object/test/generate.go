package objecttest

import (
	"github.com/google/uuid"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	sigtest "github.com/nspcc-dev/neofs-sdk-go/signature/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Range returns random object.Range.
func Range() *object.Range {
	x := object.NewRange()

	x.SetOffset(1024)
	x.SetLength(2048)

	return x
}

// Attribute returns random object.Attribute.
func Attribute() *object.Attribute {
	x := object.NewAttribute()

	x.SetKey("key")
	x.SetValue("value")

	return x
}

// SplitID returns random object.SplitID.
func SplitID() *object.SplitID {
	x := object.NewSplitID()

	x.SetUUID(uuid.New())

	return x
}

func generate(withParent bool) *object.Object {
	x := object.New()

	x.SetID(oidtest.ID())
	x.SetSessionToken(sessiontest.Token())
	x.SetPayload([]byte{1, 2, 3})
	x.SetOwnerID(ownertest.ID())
	x.SetContainerID(cidtest.ID())
	x.SetType(object.TypeTombstone)
	x.SetVersion(version.Current())
	x.SetPayloadSize(111)
	x.SetCreationEpoch(222)
	x.SetPreviousID(oidtest.ID())
	x.SetParentID(oidtest.ID())
	x.SetChildren(*oidtest.ID(), *oidtest.ID())
	x.SetAttributes(*Attribute(), *Attribute())
	x.SetSplitID(SplitID())
	x.SetPayloadChecksum(checksumtest.Checksum())
	x.SetPayloadHomomorphicHash(checksumtest.Checksum())
	x.SetSignature(sigtest.Signature())

	if withParent {
		x.SetParent(generate(false))
	}

	return x
}

// Raw returns random object.Object.
// Deprecated: (v1.0.0) use Object instead.
func Raw() *object.Object {
	return Object()
}

// Object returns random object.Object.
func Object() *object.Object {
	return generate(true)
}

// Tombstone returns random object.Tombstone.
func Tombstone() *object.Tombstone {
	x := object.NewTombstone()

	x.SetSplitID(SplitID())
	x.SetExpirationEpoch(13)
	x.SetMembers([]oid.ID{*oidtest.ID(), *oidtest.ID()})

	return x
}

// SplitInfo returns random object.SplitInfo.
func SplitInfo() *object.SplitInfo {
	x := object.NewSplitInfo()

	x.SetSplitID(SplitID())
	x.SetLink(oidtest.ID())
	x.SetLastPart(oidtest.ID())

	return x
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
	l.WriteMembers([]oid.ID{*oidtest.ID(), *oidtest.ID()})

	return &l
}
