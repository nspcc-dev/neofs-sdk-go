package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// SplitInfo groups meta information of split hierarchy for object assembly.
//
// SplitInfo is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/object.SplitInfo
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = SplitInfo(object.SplitInfo{}) // not recommended
type SplitInfo object.SplitInfo

// ReadFromV2 reads SplitInfo from the object.SplitInfo message.
//
// See also WriteToV2.
func (s *SplitInfo) ReadFromV2(m object.SplitInfo) {
	*s = SplitInfo(m)
}

// WriteToV2 writes SplitInfo to the object.SplitInfo message.
// The message must not be nil.
//
// See also ReadFromV2.
func (s SplitInfo) WriteToV2(m *object.SplitInfo) {
	*m = (object.SplitInfo)(s)
}

// SplitID returns object SplitID.
//
// Zero SplitInfo has nil SplitID.
//
// See also SetSplitID.
func (s SplitInfo) SplitID() *SplitID {
	v2 := (object.SplitInfo)(s)
	return NewSplitIDFromBytes(v2.GetSplitID())
}

// SetSplitID sets SplitID.
//
// See also SplitID.
func (s *SplitInfo) SetSplitID(v *SplitID) {
	(*object.SplitInfo)(s).SetSplitID(v.ToBytes())
}

// LastPart returns object ID of the last
// split part.
//
// Zero SplitInfo has nil last part.
//
// See also SetLastPart.
func (s SplitInfo) LastPart() *oid.ID {
	v2 := (object.SplitInfo)(s)

	lpV2 := v2.GetLastPart()
	if lpV2 == nil {
		return nil
	}

	var id oid.ID
	id.ReadFromV2(*lpV2)

	return &id
}

// SetLastPart sets last object part ID.
// Object ID must not be nil.
//
// See also LastPart.
func (s *SplitInfo) SetLastPart(v *oid.ID) {
	var idV2 refs.ObjectID
	v.WriteToV2(&idV2)

	(*object.SplitInfo)(s).SetLastPart(&idV2)
}

// Link returns object ID of the linking object.
//
// Zero SplitInfo has nil link object ID.
//
// See also SetLink.
func (s SplitInfo) Link() *oid.ID {
	v2 := (object.SplitInfo)(s)

	linkV2 := v2.GetLink()
	if linkV2 == nil {
		return nil
	}

	var id oid.ID
	id.ReadFromV2(*linkV2)

	return &id
}

// SetLink sets object ID of the linking object.
// Object ID must not be nil.
//
// See also Link.
func (s *SplitInfo) SetLink(v *oid.ID) {
	var idV2 refs.ObjectID
	v.WriteToV2(&idV2)

	(*object.SplitInfo)(s).SetLink(&idV2)
}

// Marshal marshals SplitInfo into a protobuf binary form.
func (s SplitInfo) Marshal() ([]byte, error) {
	v2 := (object.SplitInfo)(s)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of SplitInfo.
func (s *SplitInfo) Unmarshal(data []byte) error {
	return (*object.SplitInfo)(s).Unmarshal(data)
}

// MarshalJSON encodes SplitInfo to protobuf JSON format.
func (s SplitInfo) MarshalJSON() ([]byte, error) {
	v2 := (object.SplitInfo)(s)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes SplitInfo from protobuf JSON format.
func (s *SplitInfo) UnmarshalJSON(data []byte) error {
	return (*object.SplitInfo)(s).UnmarshalJSON(data)
}
