package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

type SplitInfo object.SplitInfo

// NewSplitInfoFromV2 wraps v2 SplitInfo message to SplitInfo.
//
// Nil object.SplitInfo converts to nil.
func NewSplitInfoFromV2(v2 *object.SplitInfo) *SplitInfo {
	return (*SplitInfo)(v2)
}

// NewSplitInfo creates and initializes blank SplitInfo.
//
// Defaults:
//  - splitID: nil;
//  - lastPart nil;
//  - link: nil.
func NewSplitInfo() *SplitInfo {
	return NewSplitInfoFromV2(new(object.SplitInfo))
}

// ToV2 converts SplitInfo to v2 SplitInfo message.
//
// Nil SplitInfo converts to nil.
func (s *SplitInfo) ToV2() *object.SplitInfo {
	return (*object.SplitInfo)(s)
}

func (s *SplitInfo) SplitID() *SplitID {
	return NewSplitIDFromV2(
		(*object.SplitInfo)(s).GetSplitID())
}

func (s *SplitInfo) SetSplitID(v *SplitID) {
	(*object.SplitInfo)(s).SetSplitID(v.ToV2())
}

func (s *SplitInfo) LastPart() *oid.ID {
	return oid.NewIDFromV2(
		(*object.SplitInfo)(s).GetLastPart())
}

func (s *SplitInfo) SetLastPart(v *oid.ID) {
	(*object.SplitInfo)(s).SetLastPart(v.ToV2())
}

func (s *SplitInfo) Link() *oid.ID {
	return oid.NewIDFromV2(
		(*object.SplitInfo)(s).GetLink())
}

func (s *SplitInfo) SetLink(v *oid.ID) {
	(*object.SplitInfo)(s).SetLink(v.ToV2())
}

func (s *SplitInfo) Marshal() ([]byte, error) {
	return (*object.SplitInfo)(s).StableMarshal(nil), nil
}

func (s *SplitInfo) Unmarshal(data []byte) error {
	return (*object.SplitInfo)(s).Unmarshal(data)
}

// MarshalJSON implements json.Marshaler.
func (s *SplitInfo) MarshalJSON() ([]byte, error) {
	return (*object.SplitInfo)(s).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *SplitInfo) UnmarshalJSON(data []byte) error {
	return (*object.SplitInfo)(s).UnmarshalJSON(data)
}
