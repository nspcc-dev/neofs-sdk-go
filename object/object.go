package object

import (
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Object represents in-memory structure of the NeoFS object.
// Type is compatible with NeoFS API V2 protocol.
//
// Instance can be created depending on scenario:
//   * InitCreation (an object to be placed in container);
//   * New (blank instance, usually needed for decoding);
//   * NewFromV2 (when working under NeoFS API V2 protocol).
type Object object.Object

// RequiredFields contains the minimum set of object data that must be set
// by the NeoFS user at the stage of creation.
type RequiredFields struct {
	// Identifier of the NeoFS container associated with the object.
	Container cid.ID

	// Object owner ID in the NeoFS system.
	Owner owner.ID
}

// InitCreation initializes the object instance with minimum set of required fields.
// Object is expected (but not required) to be blank. Object must not be nil.
func InitCreation(dst *Object, rf RequiredFields) {
	dst.SetContainerID(&rf.Container)
	dst.SetOwnerID(&rf.Owner)
}

// NewFromV2 wraps v2 Object message to Object.
func NewFromV2(oV2 *object.Object) *Object {
	return (*Object)(oV2)
}

// New creates and initializes blank Object.
//
// Works similar as NewFromV2(new(Object)).
func New() *Object {
	return NewFromV2(new(object.Object))
}

// ToV2 converts Object to v2 Object message.
func (o *Object) ToV2() *object.Object {
	return (*object.Object)(o)
}

// MarshalHeaderJSON marshals object's header
// into JSON format.
func (o *Object) MarshalHeaderJSON() ([]byte, error) {
	return (*object.Object)(o).GetHeader().MarshalJSON()
}

func (o *Object) setHeaderField(setter func(*object.Header)) {
	obj := (*object.Object)(o)
	h := obj.GetHeader()

	if h == nil {
		h = new(object.Header)
		obj.SetHeader(h)
	}

	setter(h)
}

func (o *Object) setSplitFields(setter func(*object.SplitHeader)) {
	o.setHeaderField(func(h *object.Header) {
		split := h.GetSplit()
		if split == nil {
			split = new(object.SplitHeader)
			h.SetSplit(split)
		}

		setter(split)
	})
}

// ID returns object identifier.
func (o *Object) ID() *oid.ID {
	return oid.NewIDFromV2(
		(*object.Object)(o).
			GetObjectID(),
	)
}

// SetID sets object identifier.
func (o *Object) SetID(v *oid.ID) {
	(*object.Object)(o).
		SetObjectID(v.ToV2())
}

// Signature returns signature of the object identifier.
func (o *Object) Signature() *signature.Signature {
	return signature.NewFromV2(
		(*object.Object)(o).GetSignature())
}

// SetSignature sets signature of the object identifier.
func (o *Object) SetSignature(v *signature.Signature) {
	(*object.Object)(o).SetSignature(v.ToV2())
}

// Payload returns payload bytes.
func (o *Object) Payload() []byte {
	return (*object.Object)(o).GetPayload()
}

// SetPayload sets payload bytes.
func (o *Object) SetPayload(v []byte) {
	(*object.Object)(o).SetPayload(v)
}

// Version returns version of the object.
func (o *Object) Version() *version.Version {
	return version.NewFromV2(
		(*object.Object)(o).
			GetHeader().
			GetVersion(),
	)
}

// SetVersion sets version of the object.
func (o *Object) SetVersion(v *version.Version) {
	o.setHeaderField(func(h *object.Header) {
		h.SetVersion(v.ToV2())
	})
}

// PayloadSize returns payload length of the object.
func (o *Object) PayloadSize() uint64 {
	return (*object.Object)(o).
		GetHeader().
		GetPayloadLength()
}

// SetPayloadSize sets payload length of the object.
func (o *Object) SetPayloadSize(v uint64) {
	o.setHeaderField(func(h *object.Header) {
		h.SetPayloadLength(v)
	})
}

// ContainerID returns identifier of the related container.
func (o *Object) ContainerID() *cid.ID {
	return cid.NewFromV2(
		(*object.Object)(o).
			GetHeader().
			GetContainerID(),
	)
}

// SetContainerID sets identifier of the related container.
func (o *Object) SetContainerID(v *cid.ID) {
	o.setHeaderField(func(h *object.Header) {
		h.SetContainerID(v.ToV2())
	})
}

// OwnerID returns identifier of the object owner.
func (o *Object) OwnerID() *owner.ID {
	return owner.NewIDFromV2(
		(*object.Object)(o).
			GetHeader().
			GetOwnerID(),
	)
}

// SetOwnerID sets identifier of the object owner.
func (o *Object) SetOwnerID(v *owner.ID) {
	o.setHeaderField(func(h *object.Header) {
		h.SetOwnerID(v.ToV2())
	})
}

// CreationEpoch returns epoch number in which object was created.
func (o *Object) CreationEpoch() uint64 {
	return (*object.Object)(o).
		GetHeader().
		GetCreationEpoch()
}

// SetCreationEpoch sets epoch number in which object was created.
func (o *Object) SetCreationEpoch(v uint64) {
	o.setHeaderField(func(h *object.Header) {
		h.SetCreationEpoch(v)
	})
}

// PayloadChecksum returns checksum of the object payload and
// bool that indicates checksum presence in the object.
//
// Zero Object does not have payload checksum.
//
// See also SetPayloadChecksum.
func (o *Object) PayloadChecksum() (checksum.Checksum, bool) {
	var v checksum.Checksum
	v2 := (*object.Object)(o)

	if hash := v2.GetHeader().GetPayloadHash(); hash != nil {
		v.ReadFromV2(*hash)
		return v, true
	}

	return v, false
}

// SetPayloadChecksum sets checksum of the object payload.
// Checksum must not be nil.
//
// See also PayloadChecksum.
func (o *Object) SetPayloadChecksum(v checksum.Checksum) {
	var v2 refs.Checksum
	v.WriteToV2(&v2)

	o.setHeaderField(func(h *object.Header) {
		h.SetPayloadHash(&v2)
	})
}

// PayloadHomomorphicHash returns homomorphic hash of the object
// payload and bool that indicates checksum presence in the object.
//
// Zero Object does not have payload homomorphic checksum.
//
// See also SetPayloadHomomorphicHash.
func (o *Object) PayloadHomomorphicHash() (checksum.Checksum, bool) {
	var v checksum.Checksum
	v2 := (*object.Object)(o)

	if hash := v2.GetHeader().GetHomomorphicHash(); hash != nil {
		v.ReadFromV2(*hash)
		return v, true
	}

	return v, false
}

// SetPayloadHomomorphicHash sets homomorphic hash of the object payload.
// Checksum must not be nil.
//
// See also PayloadHomomorphicHash.
func (o *Object) SetPayloadHomomorphicHash(v checksum.Checksum) {
	var v2 refs.Checksum
	v.WriteToV2(&v2)

	o.setHeaderField(func(h *object.Header) {
		h.SetHomomorphicHash(&v2)
	})
}

// Attributes returns object attributes.
func (o *Object) Attributes() []Attribute {
	attrs := (*object.Object)(o).
		GetHeader().
		GetAttributes()

	res := make([]Attribute, len(attrs))

	for i := range attrs {
		res[i] = *NewAttributeFromV2(&attrs[i])
	}

	return res
}

// SetAttributes sets object attributes.
func (o *Object) SetAttributes(v ...Attribute) {
	attrs := make([]object.Attribute, len(v))

	for i := range v {
		attrs[i] = *v[i].ToV2()
	}

	o.setHeaderField(func(h *object.Header) {
		h.SetAttributes(attrs)
	})
}

// PreviousID returns identifier of the previous sibling object.
func (o *Object) PreviousID() *oid.ID {
	return oid.NewIDFromV2(
		(*object.Object)(o).
			GetHeader().
			GetSplit().
			GetPrevious(),
	)
}

// SetPreviousID sets identifier of the previous sibling object.
func (o *Object) SetPreviousID(v *oid.ID) {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetPrevious(v.ToV2())
	})
}

// Children return list of the identifiers of the child objects.
func (o *Object) Children() []oid.ID {
	ids := (*object.Object)(o).
		GetHeader().
		GetSplit().
		GetChildren()

	res := make([]oid.ID, len(ids))

	for i := range ids {
		res[i] = *oid.NewIDFromV2(&ids[i])
	}

	return res
}

// SetChildren sets list of the identifiers of the child objects.
func (o *Object) SetChildren(v ...oid.ID) {
	ids := make([]refs.ObjectID, len(v))

	for i := range v {
		ids[i] = *v[i].ToV2()
	}

	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetChildren(ids)
	})
}

// NotificationInfo groups information about object notification
// that can be written to object.
//
// Topic is an optional field.
type NotificationInfo struct {
	ni object.NotificationInfo
}

// Epoch returns object notification tick
// epoch.
func (n NotificationInfo) Epoch() uint64 {
	return n.ni.Epoch()
}

// SetEpoch sets object notification tick
// epoch.
func (n *NotificationInfo) SetEpoch(epoch uint64) {
	n.ni.SetEpoch(epoch)
}

// Topic return optional object notification
// topic.
func (n NotificationInfo) Topic() string {
	return n.ni.Topic()
}

// SetTopic sets optional object notification
// topic.
func (n *NotificationInfo) SetTopic(topic string) {
	n.ni.SetTopic(topic)
}

// NotificationInfo returns notification info
// read from the object structure.
// Returns any error that appeared during notification
// information parsing.
func (o *Object) NotificationInfo() (*NotificationInfo, error) {
	ni, err := object.GetNotificationInfo((*object.Object)(o))
	if err != nil {
		return nil, err
	}

	return &NotificationInfo{
		ni: *ni,
	}, nil
}

// SetNotification writes NotificationInfo to the object structure.
func (o *Object) SetNotification(ni NotificationInfo) {
	object.WriteNotificationInfo((*object.Object)(o), ni.ni)
}

// SplitID return split identity of split object. If object is not split
// returns nil.
func (o *Object) SplitID() *SplitID {
	return NewSplitIDFromV2(
		(*object.Object)(o).
			GetHeader().
			GetSplit().
			GetSplitID(),
	)
}

// SetSplitID sets split identifier for the split object.
func (o *Object) SetSplitID(id *SplitID) {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetSplitID(id.ToV2())
	})
}

// ParentID returns identifier of the parent object.
func (o *Object) ParentID() *oid.ID {
	return oid.NewIDFromV2(
		(*object.Object)(o).
			GetHeader().
			GetSplit().
			GetParent(),
	)
}

// SetParentID sets identifier of the parent object.
func (o *Object) SetParentID(v *oid.ID) {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetParent(v.ToV2())
	})
}

// Parent returns parent object w/o payload.
func (o *Object) Parent() *Object {
	h := (*object.Object)(o).
		GetHeader().
		GetSplit()

	parSig := h.GetParentSignature()
	parHdr := h.GetParentHeader()

	if parSig == nil && parHdr == nil {
		return nil
	}

	oV2 := new(object.Object)
	oV2.SetObjectID(h.GetParent())
	oV2.SetSignature(parSig)
	oV2.SetHeader(parHdr)

	return NewFromV2(oV2)
}

// SetParent sets parent object w/o payload.
func (o *Object) SetParent(v *Object) {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetParent((*object.Object)(v).GetObjectID())
		split.SetParentSignature((*object.Object)(v).GetSignature())
		split.SetParentHeader((*object.Object)(v).GetHeader())
	})
}

func (o *Object) initRelations() {
	o.setHeaderField(func(h *object.Header) {
		h.SetSplit(new(object.SplitHeader))
	})
}

func (o *Object) resetRelations() {
	o.setHeaderField(func(h *object.Header) {
		h.SetSplit(nil)
	})
}

// SessionToken returns token of the session
// within which object was created.
func (o *Object) SessionToken() *session.Token {
	return session.NewTokenFromV2(
		(*object.Object)(o).
			GetHeader().
			GetSessionToken(),
	)
}

// SetSessionToken sets token of the session
// within which object was created.
func (o *Object) SetSessionToken(v *session.Token) {
	o.setHeaderField(func(h *object.Header) {
		h.SetSessionToken(v.ToV2())
	})
}

// Type returns type of the object.
func (o *Object) Type() Type {
	return TypeFromV2(
		(*object.Object)(o).
			GetHeader().
			GetObjectType(),
	)
}

// SetType sets type of the object.
func (o *Object) SetType(v Type) {
	o.setHeaderField(func(h *object.Header) {
		h.SetObjectType(v.ToV2())
	})
}

// CutPayload returns Object w/ empty payload.
//
// Changes of non-payload fields affect source object.
func (o *Object) CutPayload() *Object {
	ov2 := new(object.Object)
	*ov2 = *(*object.Object)(o)
	ov2.SetPayload(nil)

	return (*Object)(ov2)
}

func (o *Object) HasParent() bool {
	return (*object.Object)(o).
		GetHeader().
		GetSplit() != nil
}

// ResetRelations removes all fields of links with other objects.
func (o *Object) ResetRelations() {
	o.resetRelations()
}

// InitRelations initializes relation field.
func (o *Object) InitRelations() {
	o.initRelations()
}

// Marshal marshals object into a protobuf binary form.
func (o *Object) Marshal() ([]byte, error) {
	return (*object.Object)(o).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of object.
func (o *Object) Unmarshal(data []byte) error {
	return (*object.Object)(o).Unmarshal(data)
}

// MarshalJSON encodes object to protobuf JSON format.
func (o *Object) MarshalJSON() ([]byte, error) {
	return (*object.Object)(o).MarshalJSON()
}

// UnmarshalJSON decodes object from protobuf JSON format.
func (o *Object) UnmarshalJSON(data []byte) error {
	return (*object.Object)(o).UnmarshalJSON(data)
}
