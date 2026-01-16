package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

type split struct {
	parID    oid.ID
	prev     oid.ID
	parSig   *neofscrypto.Signature
	parHdr   *header
	children []oid.ID
	id       []byte
	first    oid.ID
}

type header struct {
	version     *version.Version
	owner       user.ID
	cnr         cid.ID
	created     uint64
	payloadLn   uint64
	pldHash     *checksum.Checksum
	typ         Type
	pldHomoHash *checksum.Checksum
	session     *session.Object
	sessionV2   *sessionv2.Token
	attrs       []Attribute
	split       split
}

// Object represents in-memory structure of the NeoFS object.
// Type is compatible with NeoFS API V2 protocol.
//
// Instance can be created depending on scenario:
//   - new (blank instance, usually needed for decoding);
//   - New (minimally initialized, usually when assembling new object to put);
//   - [Object.FromProtoMessage] (when working under NeoFS API V2 protocol).
type Object struct {
	id     oid.ID
	sig    *neofscrypto.Signature
	header header
	pld    []byte
}

// New creates and initializes new [Object] with the current version from SDK
// and given mandatory fields.
func New(cnr cid.ID, owner user.ID) *Object {
	var (
		o = new(Object)
		v = version.Current()
	)

	o.SetVersion(&v)
	o.SetContainerID(cnr)
	o.SetOwner(owner)

	return o
}

func (x split) isZero() bool {
	return x.parID.IsZero() && x.prev.IsZero() && x.parSig == nil && x.parHdr == nil && len(x.children) == 0 &&
		len(x.id) == 0 && x.first.IsZero()
}

func (x *split) fromProtoMessage(m *protoobject.Header_Split) error {
	// parent ID
	if m.Parent != nil {
		if err := x.parID.FromProtoMessage(m.Parent); err != nil {
			return fmt.Errorf("invalid parent split member ID: %w", err)
		}
	} else {
		x.parID = oid.ID{}
	}
	// previous
	if m.Previous != nil {
		if err := x.prev.FromProtoMessage(m.Previous); err != nil {
			return fmt.Errorf("invalid previous split member ID: %w", err)
		}
	} else {
		x.prev = oid.ID{}
	}
	// first
	if m.First != nil {
		if err := x.first.FromProtoMessage(m.First); err != nil {
			return fmt.Errorf("invalid first split member ID: %w", err)
		}
	} else {
		x.first = oid.ID{}
	}
	// split ID
	if x.id = m.SplitId; len(m.SplitId) > 0 {
		var uid uuid.UUID
		if err := uid.UnmarshalBinary(m.SplitId); err != nil {
			return fmt.Errorf("invalid split ID: %w", err)
		} else if ver := uid.Version(); ver != 4 {
			return fmt.Errorf("invalid split ID: wrong UUID version %d, expected 4", ver)
		}
	}
	// children
	if len(m.Children) > 0 {
		x.children = make([]oid.ID, len(m.Children))
		for i := range m.Children {
			if m.Children[i] == nil {
				return fmt.Errorf("nil child split member #%d", i)
			}
			if err := x.children[i].FromProtoMessage(m.Children[i]); err != nil {
				return fmt.Errorf("invalid child split member ID #%d: %w", i, err)
			}
		}
	} else {
		x.children = nil
	}
	// parent signature
	if m.ParentSignature != nil {
		if x.parSig == nil {
			x.parSig = new(neofscrypto.Signature)
		}
		if err := x.parSig.FromProtoMessage(m.ParentSignature); err != nil {
			return fmt.Errorf("invalid parent signature: %w", err)
		}
	} else {
		x.parSig = nil
	}
	// parent header
	if m.ParentHeader != nil {
		if x.parHdr == nil {
			x.parHdr = new(header)
		}
		if err := x.parHdr.fromProtoMessage(m.ParentHeader); err != nil {
			return fmt.Errorf("invalid parent header: %w", err)
		}
	} else {
		x.parHdr = nil
	}
	return nil
}

func (x split) protoMessage() *protoobject.Header_Split {
	m := &protoobject.Header_Split{
		SplitId: x.id,
	}
	if !x.parID.IsZero() {
		m.Parent = x.parID.ProtoMessage()
	}
	if !x.prev.IsZero() {
		m.Previous = x.prev.ProtoMessage()
	}
	if x.parSig != nil {
		m.ParentSignature = x.parSig.ProtoMessage()
	}
	if x.parHdr != nil {
		m.ParentHeader = x.parHdr.protoMessage()
	}
	if len(x.children) > 0 {
		m.Children = make([]*refs.ObjectID, len(x.children))
		for i := range x.children {
			m.Children[i] = x.children[i].ProtoMessage()
		}
	}
	if !x.first.IsZero() {
		m.First = x.first.ProtoMessage()
	}
	return m
}

func (x header) isZero() bool {
	return x.version == nil && x.owner.IsZero() && x.cnr.IsZero() && x.created == 0 && x.payloadLn == 0 &&
		x.pldHash == nil && x.typ == 0 && x.pldHomoHash == nil && x.session == nil && x.sessionV2 == nil &&
		len(x.attrs) == 0 && x.split.isZero()
}

func (x *header) fromProtoMessage(m *protoobject.Header) error {
	// version
	if m.Version != nil {
		if x.version == nil {
			x.version = new(version.Version)
		}
		if err := x.version.FromProtoMessage(m.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	} else {
		x.version = nil
	}
	// owner
	if m.OwnerId != nil {
		if err := x.owner.FromProtoMessage(m.OwnerId); err != nil {
			return fmt.Errorf("invalid owner: %w", err)
		}
	} else {
		x.owner = user.ID{}
	}
	// container
	if m.ContainerId != nil {
		if err := x.cnr.FromProtoMessage(m.ContainerId); err != nil {
			return fmt.Errorf("invalid container: %w", err)
		}
	} else {
		x.cnr = cid.ID{}
	}
	// payload checksum
	if m.PayloadHash != nil {
		if x.pldHash == nil {
			x.pldHash = new(checksum.Checksum)
		}
		if err := x.pldHash.FromProtoMessage(m.PayloadHash); err != nil {
			return fmt.Errorf("invalid payload checksum: %w", err)
		}
	} else {
		x.pldHash = nil
	}
	// type
	if m.ObjectType < 0 {
		return fmt.Errorf("negative type %d", m.ObjectType)
	}
	// payload homomorphic checksum
	if m.HomomorphicHash != nil {
		if x.pldHomoHash == nil {
			x.pldHomoHash = new(checksum.Checksum)
		}
		if err := x.pldHomoHash.FromProtoMessage(m.HomomorphicHash); err != nil {
			return fmt.Errorf("invalid payload homomorphic checksum: %w", err)
		}
	} else {
		x.pldHomoHash = nil
	}
	// session token
	if m.SessionToken != nil {
		if x.session == nil {
			x.session = new(session.Object)
		}
		if err := x.session.FromProtoMessageWithVersion(m.SessionToken, x.version); err != nil {
			return fmt.Errorf("invalid session token: %w", err)
		}
	} else {
		x.session = nil
	}
	// session token v2
	if m.SessionTokenV2 != nil {
		if x.sessionV2 == nil {
			x.sessionV2 = new(sessionv2.Token)
		}
		if err := x.sessionV2.FromProtoMessage(m.SessionTokenV2); err != nil {
			return fmt.Errorf("invalid session token v2: %w", err)
		}
	} else {
		x.sessionV2 = nil
	}
	// split header
	if m.Split != nil {
		if err := x.split.fromProtoMessage(m.Split); err != nil {
			return fmt.Errorf("invalid split header: %w", err)
		}
	} else {
		x.split = split{}
	}
	// attributes
	if ma := m.GetAttributes(); len(ma) > 0 {
		x.attrs = make([]Attribute, len(ma))
		done := make(map[string]struct{}, len(ma))
		for i := range ma {
			if ma[i] == nil {
				return fmt.Errorf("nil attribute #%d", i)
			}
			if _, ok := done[ma[i].Key]; ok {
				return fmt.Errorf("duplicated attribute %s", ma[i].Key)
			}
			if err := x.attrs[i].fromProtoMessage(ma[i], true); err != nil {
				return fmt.Errorf("invalid attribute #%d: %w", i, err)
			}
			done[ma[i].Key] = struct{}{}
		}
	} else {
		x.attrs = nil
	}
	x.created = m.CreationEpoch
	x.payloadLn = m.PayloadLength
	x.typ = Type(m.ObjectType)
	return nil
}

func (x header) protoMessage() *protoobject.Header {
	m := &protoobject.Header{
		CreationEpoch: x.created,
		PayloadLength: x.payloadLn,
		ObjectType:    protoobject.ObjectType(x.typ),
	}
	if x.version != nil {
		m.Version = x.version.ProtoMessage()
	}
	if !x.cnr.IsZero() {
		m.ContainerId = x.cnr.ProtoMessage()
	}
	if !x.owner.IsZero() {
		m.OwnerId = x.owner.ProtoMessage()
	}
	if x.pldHash != nil {
		m.PayloadHash = x.pldHash.ProtoMessage()
	}
	if x.pldHomoHash != nil {
		m.HomomorphicHash = x.pldHomoHash.ProtoMessage()
	}
	if x.session != nil {
		m.SessionToken = x.session.ProtoMessage()
	}
	if x.sessionV2 != nil {
		m.SessionTokenV2 = x.sessionV2.ProtoMessage()
	}
	if len(x.attrs) > 0 {
		m.Attributes = make([]*protoobject.Header_Attribute, len(x.attrs))
		for i := range x.attrs {
			m.Attributes[i] = &protoobject.Header_Attribute{Key: x.attrs[i].Key(), Value: x.attrs[i].Value()}
		}
	}
	if !x.split.isZero() {
		m.Split = x.split.protoMessage()
	}
	return m
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// o from it.
//
// See also [Table.ProtoMessage].
func (o *Object) FromProtoMessage(m *protoobject.Object) error {
	// ID
	if m.ObjectId != nil {
		if err := o.id.FromProtoMessage(m.ObjectId); err != nil {
			return fmt.Errorf("invalid ID: %w", err)
		}
	} else {
		o.ResetID()
	}
	// signature
	if m.Signature != nil {
		if o.sig == nil {
			o.sig = new(neofscrypto.Signature)
		}
		if err := o.sig.FromProtoMessage(m.Signature); err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	} else {
		o.sig = nil
	}
	// header
	if m.Header != nil {
		if err := o.header.fromProtoMessage(m.Header); err != nil {
			return fmt.Errorf("invalid header: %w", err)
		}
	} else {
		o.header = header{}
	}
	o.pld = m.Payload
	return nil
}

// ProtoMessage converts o into message to transmit using the NeoFS API
// protocol.
//
// See also [Object.FromProtoMessage].
func (o Object) ProtoMessage() *protoobject.Object {
	m := &protoobject.Object{
		Payload: o.pld,
	}
	if !o.id.IsZero() {
		m.ObjectId = o.id.ProtoMessage()
	}
	if o.sig != nil {
		m.Signature = o.sig.ProtoMessage()
	}
	if !o.header.isZero() {
		m.Header = o.header.protoMessage()
	}
	return m
}

// CopyTo writes deep copy of the [Object] to dst.
func (o Object) CopyTo(dst *Object) {
	o.header.copyTo(&dst.header)
	dst.id = o.id
	dst.sig = cloneSignature(o.sig)
	dst.SetPayload(bytes.Clone(o.Payload()))
}

// MarshalHeaderJSON marshals object's header into JSON format.
func (o Object) MarshalHeaderJSON() ([]byte, error) {
	return neofsproto.MarshalMessageJSON(o.header.protoMessage())
}

// Address returns current object address made of its container and object IDs.
func (o Object) Address() oid.Address {
	return oid.NewAddress(o.GetContainerID(), o.GetID())
}

// GetID returns identifier of the object. Zero return means unset ID.
//
// See also [Object.SetID].
func (o Object) GetID() oid.ID {
	return o.id
}

// SetID sets object identifier.
//
// See also [Object.GetID].
func (o *Object) SetID(v oid.ID) {
	o.id = v
}

// ResetID removes object identifier.
//
// See also [Object.SetID].
func (o *Object) ResetID() {
	o.id = oid.ID{}
}

// Signature returns signature of the object identifier.
//
// See also [Object.SetSignature].
func (o Object) Signature() *neofscrypto.Signature {
	return o.sig
}

// SetSignature sets signature of the object identifier.
//
// See also [Object.Signature].
func (o *Object) SetSignature(v *neofscrypto.Signature) {
	o.sig = v
}

// Payload returns payload bytes.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetPayload].
func (o Object) Payload() []byte {
	return o.pld
}

// SetPayload sets payload bytes.
//
// See also [Object.Payload].
func (o *Object) SetPayload(v []byte) {
	o.pld = v
}

// Version returns version of the object. Returns nil if the version is unset.
//
// See also [Object.SetVersion].
func (o Object) Version() *version.Version {
	return o.header.version
}

// SetVersion sets version of the object.
//
// See also [Object.Version].
func (o *Object) SetVersion(v *version.Version) {
	o.header.version = v
}

// PayloadSize returns payload length of the object.
//
// See also [Object.SetPayloadSize].
func (o Object) PayloadSize() uint64 {
	return o.header.payloadLn
}

// SetPayloadSize sets payload length of the object.
//
// See also [Object.PayloadSize].
func (o *Object) SetPayloadSize(v uint64) {
	o.header.payloadLn = v
}

// SetContainerID sets identifier of the related container.
//
// See also [Object.GetContainerID].
func (o *Object) SetContainerID(v cid.ID) {
	o.header.cnr = v
}

// GetContainerID returns identifier of the related container. Zero means unset
// binding.
//
// See also [Object.SetContainerID].
func (o Object) GetContainerID() cid.ID {
	return o.header.cnr
}

// Owner returns user ID of the object owner. Zero return means unset ID.
//
// See also [Object.SetOwner].
func (o Object) Owner() user.ID {
	return o.header.owner
}

// SetOwner sets identifier of the object owner.
//
// See also [Object.GetOwner].
func (o *Object) SetOwner(v user.ID) {
	o.header.owner = v
}

// CreationEpoch returns epoch number in which object was created.
//
// See also [Object.SetCreationEpoch].
func (o Object) CreationEpoch() uint64 {
	return o.header.created
}

// SetCreationEpoch sets epoch number in which object was created.
//
// See also [Object.CreationEpoch].
func (o *Object) SetCreationEpoch(v uint64) {
	o.header.created = v
}

// PayloadChecksum returns checksum of the object payload and
// bool that indicates checksum presence in the object.
//
// Zero [Object] does not have payload checksum.
//
// See also [Object.SetPayloadChecksum].
func (o Object) PayloadChecksum() (checksum.Checksum, bool) {
	if o.header.pldHash != nil {
		return *o.header.pldHash, true
	}
	return checksum.Checksum{}, false
}

// SetPayloadChecksum sets checksum of the object payload.
//
// See also [Object.PayloadChecksum].
func (o *Object) SetPayloadChecksum(v checksum.Checksum) {
	o.header.pldHash = &v
}

// PayloadHomomorphicHash returns homomorphic hash of the object
// payload and bool that indicates checksum presence in the object.
//
// Zero [Object] does not have payload homomorphic checksum.
//
// See also [Object.SetPayloadHomomorphicHash].
func (o Object) PayloadHomomorphicHash() (checksum.Checksum, bool) {
	if o.header.pldHomoHash != nil {
		return *o.header.pldHomoHash, true
	}
	return checksum.Checksum{}, false
}

// SetPayloadHomomorphicHash sets homomorphic hash of the object payload.
//
// See also [Object.PayloadHomomorphicHash].
func (o *Object) SetPayloadHomomorphicHash(v checksum.Checksum) {
	o.header.pldHomoHash = &v
}

// Attributes returns all object attributes.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetAttributes], [Object.UserAttributes].
func (o Object) Attributes() []Attribute {
	return o.header.attrs
}

// UserAttributes returns user attributes of the Object.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetAttributes], [Object.Attributes].
func (o Object) UserAttributes() []Attribute {
	res := make([]Attribute, 0, len(o.header.attrs))

	for i := range o.header.attrs {
		if !strings.HasPrefix(o.header.attrs[i].Key(), sysAttrPrefix) {
			res = append(res, o.header.attrs[i])
		}
	}

	return res
}

// SetAttributes sets object attributes.
func (o *Object) SetAttributes(v ...Attribute) {
	o.header.attrs = v
}

// GetPreviousID returns identifier of the previous sibling object. Zero return
// means unset ID.
//
// See also [Object.SetPreviousID].
func (o Object) GetPreviousID() oid.ID {
	return o.header.split.prev
}

// ResetPreviousID resets identifier of the previous sibling object.
//
// See also [Object.SetPreviousID], [Object.GetPreviousID].
func (o *Object) ResetPreviousID() {
	o.header.split.prev = oid.ID{}
}

// SetPreviousID sets identifier of the previous sibling object.
//
// See also [Object.GetPreviousID].
func (o *Object) SetPreviousID(v oid.ID) {
	o.header.split.prev = v
}

// Children return list of the identifiers of the child objects.
//
// See also [Object.SetChildren].
func (o Object) Children() []oid.ID {
	return o.header.split.children
}

// SetChildren sets list of the identifiers of the child objects.
//
// See also [Object.Children].
func (o *Object) SetChildren(v ...oid.ID) {
	o.header.split.children = v
}

// SetFirstID sets the first part's ID of the object's
// split chain.
//
// See also [Object.GetFirstID].
func (o *Object) SetFirstID(id oid.ID) {
	o.header.split.first = id
}

// FirstID returns the first part of the object's split chain.
//
// See also [Object.SetFirstID].
func (o Object) FirstID() (oid.ID, bool) {
	id := o.GetFirstID()
	return id, !id.IsZero()
}

// GetFirstID returns the first part of the object's split chain. Zero return means unset ID.
//
// See also [Object.SetFirstID].
func (o Object) GetFirstID() oid.ID {
	return o.header.split.first
}

// SplitID return split identity of split object. If object is not split returns nil.
//
// See also [Object.SetSplitID].
func (o Object) SplitID() *SplitID {
	return NewSplitIDFromV2(o.header.split.id)
}

// SetSplitID sets split identifier for the split object.
//
// See also [Object.SplitID].
func (o *Object) SetSplitID(id *SplitID) {
	o.header.split.id = id.ToV2()
}

// GetParentID returns identifier of the parent object. Zero return means unset
// ID.
//
// See also [Object.SetParentID].
func (o Object) GetParentID() oid.ID {
	return o.header.split.parID
}

// SetParentID sets identifier of the parent object.
//
// See also [Object.GetParentID].
func (o *Object) SetParentID(v oid.ID) {
	o.header.split.parID = v
}

// ResetParentID removes identifier of the parent object.
//
// See also [Object.SetParentID].
func (o *Object) ResetParentID() {
	o.header.split.parID = oid.ID{}
}

// Parent returns parent object w/o payload.
//
// See also [Object.SetParent].
func (o Object) Parent() *Object {
	if o.header.split.parSig == nil && o.header.split.parHdr == nil {
		return nil
	}

	return &Object{
		header: *o.header.split.parHdr,
		id:     o.header.split.parID,
		sig:    o.header.split.parSig,
	}
}

// SetParent sets parent object w/o payload.
//
// See also [Object.Parent].
func (o *Object) SetParent(v *Object) {
	if v != nil {
		o.header.split.parHdr = &v.header
		o.header.split.parID = v.id
		o.header.split.parSig = v.sig
		return
	}
	o.header.split.parHdr = nil
	o.header.split.parID = oid.ID{}
	o.header.split.parSig = nil
}

// SessionToken returns token of the session within which object was created.
//
// See also [Object.SetSessionToken].
func (o Object) SessionToken() *session.Object {
	return o.header.session
}

// SetSessionToken sets token of the session within which object was created.
//
// See also [Object.SessionToken].
func (o *Object) SetSessionToken(v *session.Object) {
	o.header.session = v
}

// SessionTokenV2 returns V2 token of the session within which object was created.
//
// See also [Object.SetSessionTokenV2].
func (o Object) SessionTokenV2() *sessionv2.Token {
	return o.header.sessionV2
}

// SetSessionTokenV2 sets V2 token of the session within which object was created.
//
// See also [Object.SessionTokenV2].
func (o *Object) SetSessionTokenV2(v *sessionv2.Token) {
	o.header.sessionV2 = v
}

// Type returns type of the object.
//
// See also [Object.SetType].
func (o Object) Type() Type {
	return o.header.typ
}

// SetType sets type of the object.
//
// See also [Object.Type].
func (o *Object) SetType(v Type) {
	o.header.typ = v
}

// AssociateDeleted makes this object to delete another object.
// See also [object.AttributeAssociatedObject].
func (o *Object) AssociateDeleted(id oid.ID) {
	o.header.attrs = append(o.header.attrs, Attribute{k: AttributeAssociatedObject, v: id.EncodeToString()})
	o.header.typ = TypeTombstone
	o.pld = nil
}

// AssociateLocked makes this object to lock another object from deletion.
// See also [object.AttributeAssociatedObject].
func (o *Object) AssociateLocked(id oid.ID) {
	o.header.attrs = append(o.header.attrs, Attribute{k: AttributeAssociatedObject, v: id.EncodeToString()})
	o.header.typ = TypeLock
	o.pld = nil
}

// AssociateObject associates object with the current one. See also
// [object.AttributeAssociatedObject] attribute.
func (o *Object) AssociateObject(id oid.ID) {
	o.header.attrs = append(o.header.attrs, Attribute{k: AttributeAssociatedObject, v: id.EncodeToString()})
}

// AssociatedObject returns associated object with the current one. Returns
// zero object ID of no association has been made or if associated object
// ID is incorrect. If more than one association has been made, retuned
// result is undefined. See also [object.AttributeAssociatedObject].
func (o *Object) AssociatedObject() oid.ID {
	var id oid.ID
	for _, attr := range o.header.attrs {
		if attr.k == AttributeAssociatedObject {
			_ = id.DecodeString(attr.v)
			break
		}
	}

	return id
}

// CutPayload returns [Object] w/ empty payload.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (o *Object) CutPayload() *Object {
	if o == nil {
		return nil
	}
	return &Object{
		header: o.header,
		id:     o.id,
		sig:    o.sig,
	}
}

// HasParent checks if parent (split ID) is present.
func (o Object) HasParent() bool {
	return !o.header.split.isZero()
}

// ResetRelations removes all fields of links with other objects.
func (o *Object) ResetRelations() {
	o.header.split = split{}
}

// Marshal marshals object into a protobuf binary form.
//
// See also [Object.Unmarshal].
func (o Object) Marshal() []byte {
	return neofsproto.Marshal(o)
}

// Unmarshal unmarshals protobuf binary representation of object.
//
// See also [Object.Marshal].
func (o *Object) Unmarshal(data []byte) error {
	return neofsproto.Unmarshal(data, o)
}

// MarshalJSON encodes object to protobuf JSON format.
//
// See also [Object.UnmarshalJSON].
func (o Object) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(o)
}

// UnmarshalJSON decodes object from protobuf JSON format.
//
// See also [Object.MarshalJSON].
func (o *Object) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSON(data, o)
}

// HeaderLen returns length of the binary header.
func (o Object) HeaderLen() int {
	return o.header.protoMessage().MarshaledSize()
}
