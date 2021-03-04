package object

import (
	"errors"

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

// Object represents in-memory descriptor of the NeoFS object.
//
// Object is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/object.Object
// message. See ReadFromV2 / WriteToV2 methods.
//
//
// Instance can be created depending on scenario:
//   * InitCreation (an object to be placed in container);
//   * var declaration (blank instance, usually needed for decoding).
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Object(object.Object{}) // not recommended
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

// ReadFromV2 reads Object from the object.Object message.
//
// See also WriteToV2.
func (o *Object) ReadFromV2(m object.Object) {
	*o = Object(m)
}

// WriteToV2 writes Object to the object.Object message.
// The message must not be nil.
//
// See also ReadFromV2.
func (o Object) WriteToV2(m *object.Object) {
	*m = (object.Object)(o)
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
//
// Zero Object has nil ID.
//
// See also SetID.
func (o Object) ID() *oid.ID {
	var v oid.ID

	v2 := (object.Object)(o)
	v.ReadFromV2(*v2.GetObjectID())

	return &v
}

// SetID sets object identifier. ID must not be nil.
//
// See also ID.
func (o *Object) SetID(v *oid.ID) {
	var v2 refs.ObjectID
	v.WriteToV2(&v2)

	(*object.Object)(o).
		SetObjectID(&v2)
}

// Signature returns signature of the object identifier.
//
// Zero Object has nil Signature.
//
// See also SetSignature.
func (o Object) Signature() *signature.Signature {
	v2 := (object.Object)(o)
	return signature.NewFromV2(
		v2.GetSignature())
}

// SetSignature sets signature of the object identifier.
// Signature must not be nil.
//
// See also Signature.
func (o *Object) SetSignature(v *signature.Signature) {
	(*object.Object)(o).SetSignature(v.ToV2())
}

// Payload returns payload bytes.
//
// Zero Object has nil payload.
//
// See also SetPayload.
func (o Object) Payload() []byte {
	v2 := (object.Object)(o)
	return v2.GetPayload()
}

// SetPayload sets payload bytes.
//
// See also Payload.
func (o *Object) SetPayload(v []byte) {
	(*object.Object)(o).SetPayload(v)
}

// Version returns version of the object.
//
// Zero Object has nil version.
//
// See also SetVersion.
func (o Object) Version() *version.Version {
	v2 := (object.Object)(o)
	return version.NewFromV2(
		v2.
			GetHeader().
			GetVersion(),
	)
}

// SetVersion sets version of the object.
// Version must not be nil.
//
// See also Version.
func (o *Object) SetVersion(v *version.Version) {
	o.setHeaderField(func(h *object.Header) {
		h.SetVersion(v.ToV2())
	})
}

// PayloadSize returns payload length of the object.
//
// Zero Object has 0 payload size.
//
// See also SetPayloadSize.
func (o Object) PayloadSize() uint64 {
	v2 := (object.Object)(o)
	return v2.GetHeader().GetPayloadLength()
}

// SetPayloadSize sets payload length of the object.
//
// See also PayloadSize.
func (o *Object) SetPayloadSize(v uint64) {
	o.setHeaderField(func(h *object.Header) {
		h.SetPayloadLength(v)
	})
}

// ContainerID returns identifier of the related container.
//
// Zero Object has nil ID.
//
// See also SetContainerID.
func (o Object) ContainerID() *cid.ID {
	v2 := (object.Object)(o)

	cidV2 := v2.GetHeader().GetContainerID()
	if cidV2 == nil {
		return nil
	}

	var cID cid.ID
	cID.ReadFromV2(*cidV2)

	return &cID
}

// SetContainerID sets identifier of the related container.
// Container ID must not be nil.
//
// See also ContainerID.
func (o *Object) SetContainerID(v *cid.ID) {
	var cidV2 refs.ContainerID
	v.WriteToV2(&cidV2)

	o.setHeaderField(func(h *object.Header) {
		h.SetContainerID(&cidV2)
	})
}

// OwnerID returns identifier of the object owner.
//
// Zero Object has nil ID.
//
// See also SetOwnerID.
func (o Object) OwnerID() *owner.ID {
	v2 := (object.Object)(o)
	return owner.NewIDFromV2(
		v2.
			GetHeader().
			GetOwnerID(),
	)
}

// SetOwnerID sets identifier of the object owner.
// Owner ID must not be nil.
//
// See also OwnerID.
func (o *Object) SetOwnerID(v *owner.ID) {
	o.setHeaderField(func(h *object.Header) {
		h.SetOwnerID(v.ToV2())
	})
}

// CreationEpoch returns epoch number in which object was created.
//
// Zero Object has 0 creation epoch.
//
// See also SetCreationEpoch.
func (o Object) CreationEpoch() uint64 {
	v2 := (object.Object)(o)
	return v2.
		GetHeader().
		GetCreationEpoch()
}

// SetCreationEpoch sets epoch number in which object was created.
//
// See also CreationEpoch.
func (o *Object) SetCreationEpoch(v uint64) {
	o.setHeaderField(func(h *object.Header) {
		h.SetCreationEpoch(v)
	})
}

// PayloadChecksum returns checksum of the object payload.
//
// Zero Object has zero checksum.
//
// See also SetPayloadChecksum.
func (o Object) PayloadChecksum() checksum.Checksum {
	var v checksum.Checksum
	v2 := (object.Object)(o)

	if hash := v2.GetHeader().GetPayloadHash(); hash != nil {
		v.ReadFromV2(*hash)
	}

	return v
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

// PayloadHomomorphicHash returns homomorphic hash of the object payload.
//
// Zero Object has zero checksum.
//
// See also SetPayloadHomomorphicHash.
func (o Object) PayloadHomomorphicHash() checksum.Checksum {
	var v checksum.Checksum
	v2 := (object.Object)(o)

	if hash := v2.GetHeader().GetHomomorphicHash(); hash != nil {
		v.ReadFromV2(*hash)
	}

	return v
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
//
// Zero Object has empty Attribute slice.
//
// See also SetAttributes.
func (o Object) Attributes() Attributes {
	v2 := (object.Object)(o)
	attrsV2 := v2.GetHeader().GetAttributes()

	var (
		res  = make(Attributes, len(attrsV2))
		attr Attribute
	)

	for i := range attrsV2 {
		attr.ReadFromV2(attrsV2[i])
		res[i] = attr
	}

	return res
}

// SetAttributes sets object attributes.
//
// See also Attributes.
func (o *Object) SetAttributes(v Attributes) {
	attrs := make([]object.Attribute, len(v))

	var attrV2 object.Attribute

	for i := range v {
		v[i].WriteToV2(&attrV2)
		attrs[i] = attrV2
	}

	o.setHeaderField(func(h *object.Header) {
		h.SetAttributes(attrs)
	})
}

// PreviousID returns identifier of the previous sibling object.
//
// Zero Object has nil ID.
//
// See also SetPreviousID.
func (o Object) PreviousID() *oid.ID {
	var v oid.ID
	v2 := (object.Object)(o)

	v2Prev := v2.GetHeader().GetSplit().GetPrevious()
	if v2Prev == nil {
		return nil
	}

	v.ReadFromV2(*v2Prev)

	return &v
}

// SetPreviousID sets identifier of the previous sibling object.
// Object ID must not be nil.
//
// See also PreviousID.
func (o *Object) SetPreviousID(v *oid.ID) {
	var v2 refs.ObjectID
	v.WriteToV2(&v2)

	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetPrevious(&v2)
	})
}

// Children returns list of the identifiers of the child objects.
//
// Zero Object has zero empty children slice.
//
// See also SetChildren.
func (o Object) Children() []oid.ID {
	v2 := (object.Object)(o)
	ids := v2.GetHeader().GetSplit().GetChildren()

	var (
		id  oid.ID
		res = make([]oid.ID, len(ids))
	)

	for i := range ids {
		id.ReadFromV2(ids[i])
		res[i] = id
	}

	return res
}

// SetChildren sets list of the identifiers of the child objects.
//
// See also Children.
func (o *Object) SetChildren(v ...oid.ID) {
	var (
		v2  refs.ObjectID
		ids = make([]refs.ObjectID, len(v))
	)

	for i := range v {
		v[i].WriteToV2(&v2)
		ids[i] = v2
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
//
// Zero NotificationInfo has 0 epoch.
//
// See also SetEpoch.
func (n NotificationInfo) Epoch() uint64 {
	return n.ni.Epoch()
}

// SetEpoch sets object notification tick
// epoch.
//
// See also Epoch.
func (n *NotificationInfo) SetEpoch(epoch uint64) {
	n.ni.SetEpoch(epoch)
}

// Topic returns optional object notification
// topic.
//
// Zero NotificationInfo has empty topic.
//
// See also SetTopic.
func (n NotificationInfo) Topic() string {
	return n.ni.Topic()
}

// SetTopic sets optional object notification
// topic.
//
// See also Topic.
func (n *NotificationInfo) SetTopic(topic string) {
	n.ni.SetTopic(topic)
}

// ErrNotificationNotSet means that object does not have notification.
var ErrNotificationNotSet = object.ErrNotificationNotSet

// NotificationInfo returns notification info
// read from the object structure.
// Returns any error that appeared during notification
// information parsing.
//
// Zero object does not have any notifications set.
//
// Returns ErrNotificationNotSet if no object notification
// has been set.
//
// See also SetNotification.
func (o Object) NotificationInfo() (*NotificationInfo, error) {
	ni, err := object.GetNotificationInfo((*object.Object)(&o))
	if err != nil {
		if errors.Is(err, object.ErrNotificationNotSet) {
			return nil, ErrNotificationNotSet
		}

		return nil, err
	}

	return &NotificationInfo{
		ni: *ni,
	}, nil
}

// SetNotification writes NotificationInfo to the object structure.
//
// See also NotificationInfo.
func (o *Object) SetNotification(ni NotificationInfo) {
	object.WriteNotificationInfo((*object.Object)(o), ni.ni)
}

// SplitID returns split identity of split object. If object is not split
// returns nil.
//
// Zero Object has nil SplitID.
//
// See also SetSplitID.
func (o Object) SplitID() *SplitID {
	v2 := (object.Object)(o)
	return NewSplitIDFromBytes(
		v2.
			GetHeader().
			GetSplit().
			GetSplitID(),
	)
}

// SetSplitID sets split identifier for the split object.
//
// See also SplitID.
func (o *Object) SetSplitID(id *SplitID) {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetSplitID(id.ToBytes())
	})
}

// ParentID returns identifier of the parent object.
//
// Zero Object has nil ParentID.
//
// See also SetParentID.
func (o Object) ParentID() *oid.ID {
	var v oid.ID
	v2 := (object.Object)(o)

	v.ReadFromV2(
		*v2.
			GetHeader().
			GetSplit().
			GetParent(),
	)

	return &v
}

// SetParentID sets identifier of the parent object.
// Parent ID must not be nil.
//
// See also ParentID.
func (o *Object) SetParentID(v *oid.ID) {
	var v2 refs.ObjectID
	v.WriteToV2(&v2)

	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetParent(&v2)
	})
}

// Parent returns parent object w/o payload.
//
// Zero Object has nil parent object.
//
// See also SetParent.
func (o Object) Parent() *Object {
	v2 := (object.Object)(o)
	h := v2.GetHeader().GetSplit()

	parSig := h.GetParentSignature()
	parHdr := h.GetParentHeader()

	if parSig == nil && parHdr == nil {
		return nil
	}

	var oV2 object.Object
	oV2.SetObjectID(h.GetParent())
	oV2.SetSignature(parSig)
	oV2.SetHeader(parHdr)

	var obj Object
	obj.ReadFromV2(oV2)

	return &obj
}

// SetParent sets parent object w/o payload.
// Parent must not be nil.
//
// See also Parent.
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
//
// Zero Object has nil token.
//
// See also SetSessionToken.
func (o Object) SessionToken() *session.Token {
	v2 := (object.Object)(o)
	return session.NewTokenFromV2(v2.
		GetHeader().
		GetSessionToken(),
	)
}

// SetSessionToken sets token of the session
// within which object was created. Session token
// must not be nil.
//
// See also SessionToken.
func (o *Object) SetSessionToken(v *session.Token) {
	o.setHeaderField(func(h *object.Header) {
		h.SetSessionToken(v.ToV2())
	})
}

// Type returns type of the object.
//
// Zero Object has zero Type.
//
// See also SetType.
func (o Object) Type() Type {
	var (
		v2 = (object.Object)(o)
		t  Type
	)

	t.ReadFromV2(
		v2.
			GetHeader().
			GetObjectType(),
	)

	return t
}

// SetType sets type of the object.
//
// See also Type.
func (o *Object) SetType(v Type) {
	var tV2 object.Type
	v.WriteToV2(&tV2)

	o.setHeaderField(func(h *object.Header) {
		h.SetObjectType(tV2)
	})
}

// CutPayload returns Object w/ empty payload.
//
// Changes of non-payload fields affect source object.
func (o Object) CutPayload() Object {
	o.SetPayload(nil)

	return o
}

// HasParent returns true if the object has a parent.
func (o Object) HasParent() bool {
	v2 := (object.Object)(o)
	return v2.GetHeader().GetSplit() != nil
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
func (o Object) Marshal() ([]byte, error) {
	v2 := (object.Object)(o)
	return v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of object.
func (o *Object) Unmarshal(data []byte) error {
	return (*object.Object)(o).Unmarshal(data)
}

// MarshalJSON encodes object to protobuf JSON format.
func (o Object) MarshalJSON() ([]byte, error) {
	v2 := (object.Object)(o)
	return v2.MarshalJSON()
}

// UnmarshalJSON decodes object from protobuf JSON format.
func (o *Object) UnmarshalJSON(data []byte) error {
	return (*object.Object)(o).UnmarshalJSON(data)
}
