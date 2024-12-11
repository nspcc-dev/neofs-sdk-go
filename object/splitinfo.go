package object

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
)

// SplitInfo is an SDK representation of [object.SplitInfo].
type SplitInfo struct {
	splitID []byte
	last    oid.ID
	link    oid.ID
	first   oid.ID
}

// NewSplitInfo creates and initializes blank [SplitInfo].
//
// Defaults:
//   - splitID: nil;
//   - lastPart nil;
//   - link: nil.
func NewSplitInfo() *SplitInfo {
	return new(SplitInfo)
}

// ProtoMessage converts s into message to transmit using the NeoFS API
// protocol.
//
// See also [SplitInfo.FromProtoMessage].
func (s SplitInfo) ProtoMessage() *protoobject.SplitInfo {
	m := &protoobject.SplitInfo{
		SplitId: s.splitID,
	}
	if !s.last.IsZero() {
		m.LastPart = s.last.ProtoMessage()
	}
	if !s.first.IsZero() {
		m.FirstPart = s.first.ProtoMessage()
	}
	if !s.link.IsZero() {
		m.Link = s.link.ProtoMessage()
	}
	return m
}

// SplitID returns [SplitID] if it has been set. New objects may miss it,
// use [SplitInfo.GetFirstPart] as a split chain identifier.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [SplitInfo.SetSplitID].
func (s SplitInfo) SplitID() *SplitID {
	return NewSplitIDFromV2(s.splitID)
}

// SetSplitID sets split ID in object ID. It resets split ID if nil passed.
//
// See also [SplitInfo.SplitID].
//
// DEPRECATED.[SplitInfo.SetFirstPart] usage is required for the _new_ split
// objects, it serves as chain identification.
func (s *SplitInfo) SetSplitID(v *SplitID) {
	s.splitID = v.ToV2()
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
	return s.last
}

// SetLastPart sets the last object ID.
//
// See also [SplitInfo.GetLastPart].
func (s *SplitInfo) SetLastPart(v oid.ID) {
	s.last = v
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
	return s.link
}

// SetLink sets linker object ID.
//
// See also [SplitInfo.GetLink].
func (s *SplitInfo) SetLink(v oid.ID) {
	s.link = v
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
	return s.first
}

// SetFirstPart sets the first part of the split chain.
//
// See also [SplitInfo.GetFirstPart].
func (s *SplitInfo) SetFirstPart(v oid.ID) {
	s.first = v
}

// Marshal marshals [SplitInfo] into a protobuf binary form.
//
// See also [SplitInfo.Unmarshal].
func (s SplitInfo) Marshal() []byte {
	return neofsproto.Marshal(s)
}

// Unmarshal unmarshals protobuf binary representation of [SplitInfo].
//
// See also [SplitInfo.Marshal].
func (s *SplitInfo) Unmarshal(data []byte) error {
	return neofsproto.Unmarshal(data, s)
}

// MarshalJSON implements json.Marshaler.
//
// See also [SplitInfo.UnmarshalJSON].
func (s SplitInfo) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(s)
}

// UnmarshalJSON implements json.Unmarshaler.
//
// See also [SplitInfo.MarshalJSON].
func (s *SplitInfo) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSON(data, s)
}

var errSplitInfoMissingFields = errors.New("neither link object ID nor last part object ID is set")

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// s from it.
//
// See also [SplitInfo.ProtoMessage].
func (s *SplitInfo) FromProtoMessage(m *protoobject.SplitInfo) error {
	if s.splitID = m.SplitId; len(m.SplitId) > 0 {
		var uid uuid.UUID
		if err := uid.UnmarshalBinary(m.SplitId); err != nil {
			return fmt.Errorf("invalid split ID: %w", err)
		} else if v := uid.Version(); v != 4 {
			return fmt.Errorf("invalid split ID: wrong UUID version %d, expected 4", v)
		}
	}

	if m.Link == nil && m.LastPart == nil {
		return errSplitInfoMissingFields
	}

	if m.Link != nil {
		err := s.link.FromProtoMessage(m.Link)
		if err != nil {
			return fmt.Errorf("could not convert link object ID: %w", err)
		}
	} else {
		s.link = oid.ID{}
	}

	if m.LastPart != nil {
		err := s.last.FromProtoMessage(m.LastPart)
		if err != nil {
			return fmt.Errorf("could not convert last part object ID: %w", err)
		}
	} else {
		s.last = oid.ID{}
	}

	if m.FirstPart != nil { // can be missing for old objects
		err := s.first.FromProtoMessage(m.FirstPart)
		if err != nil {
			return fmt.Errorf("could not convert first part object ID: %w", err)
		}
	} else {
		s.first = oid.ID{}
	}

	return nil
}
