package object

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// SplitInfo is an SDK representation of [object.SplitInfo].
type SplitInfo object.SplitInfo

// NewSplitInfoFromV2 wraps v2 [object.SplitInfo] message to [SplitInfo].
//
// Nil object.SplitInfo converts to nil.
func NewSplitInfoFromV2(v2 *object.SplitInfo) *SplitInfo {
	return (*SplitInfo)(v2)
}

// NewSplitInfo creates and initializes blank [SplitInfo].
//
// Defaults:
//   - splitID: nil;
//   - lastPart nil;
//   - link: nil.
func NewSplitInfo() *SplitInfo {
	return NewSplitInfoFromV2(new(object.SplitInfo))
}

// ToV2 converts [SplitInfo] to v2 [object.SplitInfo] message.
//
// Nil SplitInfo converts to nil.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (s *SplitInfo) ToV2() *object.SplitInfo {
	return (*object.SplitInfo)(s)
}

// SplitID returns [SplitID] if it has been set. New objects may miss it,
// use [SplitInfo.GetFirstPart] as a split chain identifier.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [SplitInfo.SetSplitID].
func (s *SplitInfo) SplitID() *SplitID {
	return NewSplitIDFromV2(
		(*object.SplitInfo)(s).GetSplitID())
}

// SetSplitID sets split ID in object ID. It resets split ID if nil passed.
//
// See also [SplitInfo.SplitID].
//
// DEPRECATED.[SplitInfo.SetFirstPart] usage is required for the _new_ split
// objects, it serves as chain identification.
func (s *SplitInfo) SetSplitID(v *SplitID) {
	(*object.SplitInfo)(s).SetSplitID(v.ToV2())
}

// LastPart returns last object ID, can be used to retrieve original object.
// The second return value is a flag, indicating if the last part is present.
//
// See also [SplitInfo.SetLastPart].
// Deprecated: use [SplitInfo.GetLastPart] instead.
func (s SplitInfo) LastPart() (oid.ID, bool) {
	id := s.GetLastPart()
	return id, !id.IsZero()
}

// GetLastPart returns last object ID, can be used to retrieve original object.
// Zero return means unset ID.
//
// See also [SplitInfo.SetLastPart].
func (s SplitInfo) GetLastPart() oid.ID {
	var id oid.ID
	m := (*object.SplitInfo)(&s).GetLastPart()
	if m != nil {
		_ = id.ReadFromV2(*m)
	}
	return id
}

// SetLastPart sets the last object ID.
//
// See also [SplitInfo.GetLastPart].
func (s *SplitInfo) SetLastPart(v oid.ID) {
	var idV2 refs.ObjectID
	v.WriteToV2(&idV2)

	(*object.SplitInfo)(s).SetLastPart(&idV2)
}

// Link returns a linker object ID.
// The second return value is a flag, indicating if the last part is present.
//
// See also [SplitInfo.SetLink].
// Deprecated: use [SplitInfo.GetLink] instead.
func (s SplitInfo) Link() (oid.ID, bool) {
	id := s.GetLink()
	return id, !id.IsZero()
}

// GetLink returns a linker object ID. Zero return means unset ID.
//
// See also [SplitInfo.SetLink].
func (s SplitInfo) GetLink() oid.ID {
	var id oid.ID
	if m := (*object.SplitInfo)(&s).GetLink(); m != nil {
		_ = id.ReadFromV2(*m)
	}
	return id
}

// SetLink sets linker object ID.
//
// See also [SplitInfo.GetLink].
func (s *SplitInfo) SetLink(v oid.ID) {
	var idV2 refs.ObjectID
	v.WriteToV2(&idV2)

	(*object.SplitInfo)(s).SetLink(&idV2)
}

// FirstPart returns the first part of the split chain.
//
// See also [SplitInfo.SetFirstPart].
// Deprecated: use [SplitInfo.GetFirstPart] instead.
func (s SplitInfo) FirstPart() (oid.ID, bool) {
	id := s.GetFirstPart()
	return id, !id.IsZero()
}

// GetFirstPart returns the first part of the split chain. Zero means unset ID.
//
// See also [SplitInfo.SetFirstPart].
func (s SplitInfo) GetFirstPart() oid.ID {
	var id oid.ID
	if m := (*object.SplitInfo)(&s).GetFirstPart(); m != nil {
		_ = id.ReadFromV2(*m)
	}
	return id
}

// SetFirstPart sets the first part of the split chain.
//
// See also [SplitInfo.GetFirstPart].
func (s *SplitInfo) SetFirstPart(v oid.ID) {
	var idV2 refs.ObjectID
	v.WriteToV2(&idV2)

	(*object.SplitInfo)(s).SetFirstPart(&idV2)
}

// Marshal marshals [SplitInfo] into a protobuf binary form.
//
// See also [SplitInfo.Unmarshal].
func (s *SplitInfo) Marshal() []byte {
	return (*object.SplitInfo)(s).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of [SplitInfo].
//
// See also [SplitInfo.Marshal].
func (s *SplitInfo) Unmarshal(data []byte) error {
	err := (*object.SplitInfo)(s).Unmarshal(data)
	if err != nil {
		return err
	}

	return formatCheckSI((*object.SplitInfo)(s))
}

// MarshalJSON implements json.Marshaler.
//
// See also [SplitInfo.UnmarshalJSON].
func (s *SplitInfo) MarshalJSON() ([]byte, error) {
	return (*object.SplitInfo)(s).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
//
// See also [SplitInfo.MarshalJSON].
func (s *SplitInfo) UnmarshalJSON(data []byte) error {
	err := (*object.SplitInfo)(s).UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return formatCheckSI((*object.SplitInfo)(s))
}

var errSplitInfoMissingFields = errors.New("neither link object ID nor last part object ID is set")

func formatCheckSI(v2 *object.SplitInfo) error {
	link := v2.GetLink()
	lastPart := v2.GetLastPart()
	if link == nil && lastPart == nil {
		return errSplitInfoMissingFields
	}

	var oID oid.ID

	if link != nil {
		err := oID.ReadFromV2(*link)
		if err != nil {
			return fmt.Errorf("could not convert link object ID: %w", err)
		}
	}

	if lastPart != nil {
		err := oID.ReadFromV2(*lastPart)
		if err != nil {
			return fmt.Errorf("could not convert last part object ID: %w", err)
		}
	}

	firstPart := v2.GetFirstPart()
	if firstPart != nil { // can be missing for old objects
		err := oID.ReadFromV2(*firstPart)
		if err != nil {
			return fmt.Errorf("could not convert first part object ID: %w", err)
		}
	}

	return nil
}
