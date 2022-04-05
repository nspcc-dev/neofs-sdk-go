package object

import (
	"github.com/google/uuid"
)

// SplitID is a UUIDv4 used as attribute in split objects.
type SplitID struct {
	uuid *uuid.UUID
}

// NewSplitID returns UUID representation of splitID attribute.
//
// Defaults:
//  - id: random UUID.
func NewSplitID() SplitID {
	id := uuid.New()
	return SplitID{
		uuid: &id,
	}
}

// NewSplitIDFromBytes returns parsed UUID from bytes.
// If v is invalid UUIDv4 byte sequence, then function returns
// zero SplitID.
//
// Nil converts to zero SplitID.
func NewSplitIDFromBytes(v []byte) SplitID {
	var sid SplitID

	if v == nil {
		return sid
	}

	id := uuid.New()

	err := id.UnmarshalBinary(v)
	if err != nil {
		return sid
	}

	return SplitID{
		uuid: &id,
	}
}

// Parse is a reverse action to String().
func (id *SplitID) Parse(s string) error {
	uid, err := uuid.Parse(s)
	if err != nil {
		return err
	}

	id.uuid = &uid

	return nil
}

// String implements fmt.Stringer interface method.
func (id SplitID) String() string {
	if id.Empty() {
		return ""
	}

	return id.uuid.String()
}

// SetUUID sets pre created UUID structure as SplitID.
func (id *SplitID) SetUUID(v uuid.UUID) {
	if id != nil {
		id.uuid = &v
	}
}

// ToBytes converts SplitID to a representation of SplitID in neofs-api v2.
//
// Zero SplitID converts to nil.
func (id SplitID) ToBytes() []byte {
	if id.Empty() {
		return nil
	}

	data, _ := id.uuid.MarshalBinary() // err is always nil

	return data
}

// Empty returns true if it is called on
// zero SplitID.
func (id SplitID) Empty() bool {
	return id.uuid == nil
}
