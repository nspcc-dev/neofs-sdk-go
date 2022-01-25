package objecttest

import (
	"github.com/google/uuid"
	checksumtest "github.com/nspcc-dev/neofs-sdk-go/checksum/test"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/id/test"
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

func generateRaw(withParent bool) *object.RawObject {
	x := object.NewRaw()

	x.SetID(test.ID())
	x.SetSessionToken(sessiontest.Token())
	x.SetPayload([]byte{1, 2, 3})
	x.SetOwnerID(ownertest.ID())
	x.SetContainerID(cidtest.ID())
	x.SetType(object.TypeTombstone)
	x.SetVersion(version.Current())
	x.SetPayloadSize(111)
	x.SetCreationEpoch(222)
	x.SetPreviousID(test.ID())
	x.SetParentID(test.ID())
	x.SetChildren(test.ID(), test.ID())
	x.SetAttributes(Attribute(), Attribute())
	x.SetSplitID(SplitID())
	x.SetPayloadChecksum(checksumtest.Checksum())
	x.SetPayloadHomomorphicHash(checksumtest.Checksum())
	x.SetSignature(sigtest.Signature())

	if withParent {
		x.SetParent(generateRaw(false).Object())
	}

	return x
}

// Raw returns random object.RawObject.
func Raw() *object.RawObject {
	return generateRaw(true)
}

// Object returns random object.Object.
func Object() *object.Object {
	return Raw().Object()
}

// Tombstone returns random object.Tombstone.
func Tombstone() *object.Tombstone {
	x := object.NewTombstone()

	x.SetSplitID(SplitID())
	x.SetExpirationEpoch(13)
	x.SetMembers([]*oid.ID{test.ID(), test.ID()})

	return x
}

// SplitInfo returns random object.SplitInfo.
func SplitInfo() *object.SplitInfo {
	x := object.NewSplitInfo()

	x.SetSplitID(SplitID())
	x.SetLink(test.ID())
	x.SetLastPart(test.ID())

	return x
}

// SearchFilters returns random object.SearchFilters.
func SearchFilters() object.SearchFilters {
	x := object.NewSearchFilters()

	x.AddObjectIDFilter(object.MatchStringEqual, test.ID())
	x.AddObjectContainerIDFilter(object.MatchStringNotEqual, cidtest.ID())

	return x
}
