package object

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Object represents in-memory structure of the NeoFS object.
// Type is compatible with NeoFS API V2 protocol.
//
// Instance can be created depending on scenario:
//   - [Object.InitCreation] (an object to be placed in container);
//   - New (blank instance, usually needed for decoding);
//   - [Object.ReadFromV2] (when working under NeoFS API V2 protocol).
type Object object.Object

// RequiredFields contains the minimum set of object data that must be set
// by the NeoFS user at the stage of creation.
type RequiredFields struct {
	// Identifier of the NeoFS container associated with the object.
	Container cid.ID

	// Object owner's user ID in the NeoFS system.
	Owner user.ID
}

// InitCreation initializes the object instance with minimum set of required fields.
func (o *Object) InitCreation(rf RequiredFields) {
	o.SetContainerID(rf.Container)
	o.SetOwner(rf.Owner)
}

// NewFromV2 wraps v2 [object.Object] message to [Object].
//
// Deprecated: BUG: fields' format is not checked. Use [Object.ReadFromV2]
// instead.
func NewFromV2(oV2 *object.Object) *Object {
	return (*Object)(oV2)
}

// New creates and initializes blank [Object].
func New() *Object {
	return new(Object)
}

func verifySplitHeaderV2(m object.SplitHeader) error {
	// parent ID
	if mID := m.GetParent(); mID != nil {
		if err := new(oid.ID).ReadFromV2(*mID); err != nil {
			return fmt.Errorf("invalid parent split member ID: %w", err)
		}
	}
	// previous
	if mID := m.GetPrevious(); mID != nil {
		if err := new(oid.ID).ReadFromV2(*mID); err != nil {
			return fmt.Errorf("invalid previous split member ID: %w", err)
		}
	}
	// first
	if mID := m.GetFirst(); mID != nil {
		if err := new(oid.ID).ReadFromV2(*mID); err != nil {
			return fmt.Errorf("invalid first split member ID: %w", err)
		}
	}
	// split ID
	if b := m.GetSplitID(); len(b) > 0 {
		var uid uuid.UUID
		if err := uid.UnmarshalBinary(b); err != nil {
			return fmt.Errorf("invalid split UUID: %w", err)
		} else if ver := uid.Version(); ver != 4 {
			return fmt.Errorf("invalid split UUID version %d", ver)
		}
	}
	// children
	if mc := m.GetChildren(); len(mc) > 0 {
		for i := range mc {
			if err := new(oid.ID).ReadFromV2(mc[i]); err != nil {
				return fmt.Errorf("invalid child split member ID #%d: %w", i, err)
			}
		}
	}
	// parent signature
	if ms := m.GetParentSignature(); ms != nil {
		if err := new(neofscrypto.Signature).ReadFromV2(*ms); err != nil {
			return fmt.Errorf("invalid parent signature: %w", err)
		}
	}
	// parent header
	if mh := m.GetParentHeader(); mh != nil {
		if err := verifyHeaderV2(*mh); err != nil {
			return fmt.Errorf("invalid parent header: %w", err)
		}
	}
	return nil
}

func verifyHeaderV2(m object.Header) error {
	// version
	if mv := m.GetVersion(); mv != nil {
		if err := new(version.Version).ReadFromV2(*mv); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}
	// owner
	if mu := m.GetOwnerID(); mu != nil {
		if err := new(user.ID).ReadFromV2(*mu); err != nil {
			return fmt.Errorf("invalid owner: %w", err)
		}
	}
	// container
	if mc := m.GetContainerID(); mc != nil {
		if err := new(cid.ID).ReadFromV2(*mc); err != nil {
			return fmt.Errorf("invalid container: %w", err)
		}
	}
	// payload checksum
	if mc := m.GetPayloadHash(); mc != nil {
		if err := new(checksum.Checksum).ReadFromV2(*mc); err != nil {
			return fmt.Errorf("invalid payload checksum: %w", err)
		}
	}
	// payload homomorphic checksum
	if mc := m.GetHomomorphicHash(); mc != nil {
		if err := new(checksum.Checksum).ReadFromV2(*mc); err != nil {
			return fmt.Errorf("invalid payload homomorphic checksum: %w", err)
		}
	}
	// session token
	if ms := m.GetSessionToken(); ms != nil {
		if err := new(session.Object).ReadFromV2(*ms); err != nil {
			return fmt.Errorf("invalid session token: %w", err)
		}
	}
	// split header
	if ms := m.GetSplit(); ms != nil {
		if err := verifySplitHeaderV2(*ms); err != nil {
			return fmt.Errorf("invalid split header: %w", err)
		}
	}
	// attributes
	if ma := m.GetAttributes(); len(ma) > 0 {
		done := make(map[string]struct{}, len(ma))
		for i := range ma {
			key := ma[i].GetKey()
			if key == "" {
				return fmt.Errorf("empty key of the attribute #%d", i)
			}
			if _, ok := done[key]; ok {
				return fmt.Errorf("duplicated attribute %s", key)
			}
			val := ma[i].GetValue()
			if val == "" {
				return fmt.Errorf("empty value of the attribute #%d (%s)", i, key)
			}
			switch key {
			case AttributeExpirationEpoch:
				if _, err := strconv.ParseUint(val, 10, 64); err != nil {
					return fmt.Errorf("invalid expiration attribute (must be a uint): %w", err)
				}
			}
			done[key] = struct{}{}
		}
	}
	return nil
}

// ReadFromV2 reads Object from the [object.Object] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
func (o *Object) ReadFromV2(m object.Object) error {
	// ID
	if mID := m.GetObjectID(); mID != nil {
		if err := new(oid.ID).ReadFromV2(*mID); err != nil {
			return fmt.Errorf("invalid ID: %w", err)
		}
	}
	// signature
	if ms := m.GetSignature(); ms != nil {
		if err := new(neofscrypto.Signature).ReadFromV2(*ms); err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}
	// header
	if mh := m.GetHeader(); mh != nil {
		if err := verifyHeaderV2(*mh); err != nil {
			return fmt.Errorf("invalid header: %w", err)
		}
	}
	*o = Object(m)
	return nil
}

// ToV2 converts [Object] to v2 [object.Object] message.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (o Object) ToV2() *object.Object {
	return (*object.Object)(&o)
}

// CopyTo writes deep copy of the [Object] to dst.
func (o Object) CopyTo(dst *Object) {
	id := (*object.Object)(&o).GetObjectID()
	(*object.Object)(dst).SetObjectID(copyObjectID(id))

	sig := (*object.Object)(&o).GetSignature()
	(*object.Object)(dst).SetSignature(copySignature(sig))

	header := (*object.Object)(&o).GetHeader()
	(*object.Object)(dst).SetHeader(copyHeader(header))

	dst.SetPayload(bytes.Clone(o.Payload()))
}

// MarshalHeaderJSON marshals object's header into JSON format.
func (o Object) MarshalHeaderJSON() ([]byte, error) {
	return (*object.Object)(&o).GetHeader().MarshalJSON()
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
// See also [Object.SetID].
// Deprecated: use [Object.GetID] instead.
func (o *Object) ID() (oid.ID, bool) {
	id := o.GetID()
	return id, !id.IsZero()
}

// GetID returns identifier of the object. Zero return means unset ID.
//
// See also [Object.SetID].
func (o Object) GetID() oid.ID {
	var res oid.ID
	m := (*object.Object)(&o)
	if id := m.GetObjectID(); id != nil {
		if err := res.ReadFromV2(*id); err != nil {
			panic(fmt.Errorf("unexpected ID decoding failure: %w", err))
		}
	}

	return res
}

// SetID sets object identifier.
//
// See also [Object.GetID].
func (o *Object) SetID(v oid.ID) {
	var v2 refs.ObjectID
	v.WriteToV2(&v2)

	(*object.Object)(o).
		SetObjectID(&v2)
}

// ResetID removes object identifier.
//
// See also [Object.SetID].
func (o *Object) ResetID() {
	(*object.Object)(o).
		SetObjectID(nil)
}

// Signature returns signature of the object identifier.
//
// See also [Object.SetSignature].
func (o Object) Signature() *neofscrypto.Signature {
	sigv2 := (*object.Object)(&o).GetSignature()
	if sigv2 == nil {
		return nil
	}

	var sig neofscrypto.Signature
	if err := sig.ReadFromV2(*sigv2); err != nil {
		return nil
	}

	return &sig
}

// SetSignature sets signature of the object identifier.
//
// See also [Object.Signature].
func (o *Object) SetSignature(v *neofscrypto.Signature) {
	var sigv2 *refs.Signature

	if v != nil {
		sigv2 = new(refs.Signature)

		v.WriteToV2(sigv2)
	}

	(*object.Object)(o).SetSignature(sigv2)
}

// Payload returns payload bytes.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetPayload].
func (o Object) Payload() []byte {
	return (*object.Object)(&o).GetPayload()
}

// SetPayload sets payload bytes.
//
// See also [Object.Payload].
func (o *Object) SetPayload(v []byte) {
	(*object.Object)(o).SetPayload(v)
}

// Version returns version of the object. Returns nil if the version is unset.
//
// See also [Object.SetVersion].
func (o Object) Version() *version.Version {
	verV2 := (*object.Object)(&o).GetHeader().GetVersion()
	if verV2 == nil {
		return nil
	}
	var ver version.Version
	if err := ver.ReadFromV2(*verV2); err != nil {
		return nil
	}
	return &ver
}

// SetVersion sets version of the object.
//
// See also [Object.Version].
func (o *Object) SetVersion(v *version.Version) {
	var verV2 refs.Version
	v.WriteToV2(&verV2)

	o.setHeaderField(func(h *object.Header) {
		h.SetVersion(&verV2)
	})
}

// PayloadSize returns payload length of the object.
//
// See also [Object.SetPayloadSize].
func (o Object) PayloadSize() uint64 {
	return (*object.Object)(&o).
		GetHeader().
		GetPayloadLength()
}

// SetPayloadSize sets payload length of the object.
//
// See also [Object.PayloadSize].
func (o *Object) SetPayloadSize(v uint64) {
	o.setHeaderField(func(h *object.Header) {
		h.SetPayloadLength(v)
	})
}

// ContainerID returns identifier of the related container.
//
// See also [Object.SetContainerID].
// Deprecated: use [Object.GetContainerID] instead.
func (o *Object) ContainerID() (v cid.ID, isSet bool) {
	cnr := o.GetContainerID()
	return cnr, !cnr.IsZero()
}

// SetContainerID sets identifier of the related container.
//
// See also [Object.GetContainerID].
func (o *Object) SetContainerID(v cid.ID) {
	var cidV2 refs.ContainerID
	v.WriteToV2(&cidV2)

	o.setHeaderField(func(h *object.Header) {
		h.SetContainerID(&cidV2)
	})
}

// GetContainerID returns identifier of the related container. Zero means unset
// binding.
//
// See also [Object.SetContainerID].
func (o Object) GetContainerID() cid.ID {
	var cnr cid.ID
	if m := (*object.Object)(&o).GetHeader().GetContainerID(); m != nil {
		if err := cnr.ReadFromV2(*m); err != nil {
			panic(fmt.Errorf("unexpected container ID decoding failure: %w", err))
		}
	}
	return cnr
}

// OwnerID returns identifier of the object owner.
//
// See also [Object.SetOwnerID].
// Deprecated: use [Object.Owner] instead.
func (o Object) OwnerID() *user.ID { res := o.Owner(); return &res }

// Owner returns user ID of the object owner. Zero return means unset ID.
//
// See also [Object.SetOwner].
func (o Object) Owner() user.ID {
	var id user.ID

	m := (*object.Object)(&o).GetHeader().GetOwnerID()
	if m != nil {
		// unlike other IDs, user.ID.ReadFromV2 also expects correct prefix and checksum
		// suffix. So, we cannot call it and panic on error here because nothing
		// prevents user from setting incorrect owner ID (Object.SetOwnerID accepts it).
		// At the same time, we always expect correct length.
		b := m.GetValue()
		if len(b) != user.IDSize {
			panic(fmt.Errorf("unexpected user ID decoding failure: invalid length %d, expected %d", len(b), user.IDSize))
		}
		copy(id[:], b)
	}

	return id
}

// SetOwnerID sets identifier of the object owner.
//
// See also [Object.OwnerID].
// Deprecated: use [Object.SetOwner] accepting value instead.
func (o *Object) SetOwnerID(v *user.ID) { o.SetOwner(*v) }

// SetOwner sets identifier of the object owner.
//
// See also [Object.GetOwner].
func (o *Object) SetOwner(v user.ID) {
	o.setHeaderField(func(h *object.Header) {
		var m refs.OwnerID
		v.WriteToV2(&m)

		h.SetOwnerID(&m)
	})
}

// CreationEpoch returns epoch number in which object was created.
//
// See also [Object.SetCreationEpoch].
func (o Object) CreationEpoch() uint64 {
	return (*object.Object)(&o).
		GetHeader().
		GetCreationEpoch()
}

// SetCreationEpoch sets epoch number in which object was created.
//
// See also [Object.CreationEpoch].
func (o *Object) SetCreationEpoch(v uint64) {
	o.setHeaderField(func(h *object.Header) {
		h.SetCreationEpoch(v)
	})
}

// PayloadChecksum returns checksum of the object payload and
// bool that indicates checksum presence in the object.
//
// Zero [Object] does not have payload checksum.
//
// See also [Object.SetPayloadChecksum].
func (o Object) PayloadChecksum() (checksum.Checksum, bool) {
	var v checksum.Checksum
	v2 := (*object.Object)(&o)

	if hash := v2.GetHeader().GetPayloadHash(); hash != nil {
		err := v.ReadFromV2(*hash)
		return v, (err == nil)
	}

	return v, false
}

// SetPayloadChecksum sets checksum of the object payload.
//
// See also [Object.PayloadChecksum].
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
// Zero [Object] does not have payload homomorphic checksum.
//
// See also [Object.SetPayloadHomomorphicHash].
func (o Object) PayloadHomomorphicHash() (checksum.Checksum, bool) {
	var v checksum.Checksum
	v2 := (*object.Object)(&o)

	if hash := v2.GetHeader().GetHomomorphicHash(); hash != nil {
		err := v.ReadFromV2(*hash)
		return v, (err == nil)
	}

	return v, false
}

// SetPayloadHomomorphicHash sets homomorphic hash of the object payload.
//
// See also [Object.PayloadHomomorphicHash].
func (o *Object) SetPayloadHomomorphicHash(v checksum.Checksum) {
	var v2 refs.Checksum
	v.WriteToV2(&v2)

	o.setHeaderField(func(h *object.Header) {
		h.SetHomomorphicHash(&v2)
	})
}

// Attributes returns all object attributes.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetAttributes], [Object.UserAttributes].
func (o Object) Attributes() []Attribute {
	attrs := (*object.Object)(&o).
		GetHeader().
		GetAttributes()

	res := make([]Attribute, len(attrs))

	for i := range attrs {
		res[i] = *NewAttributeFromV2(&attrs[i])
	}

	return res
}

// UserAttributes returns user attributes of the Object.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetAttributes], [Object.Attributes].
func (o Object) UserAttributes() []Attribute {
	attrs := (*object.Object)(&o).
		GetHeader().
		GetAttributes()

	res := make([]Attribute, 0, len(attrs))

	for i := range attrs {
		if !strings.HasPrefix(attrs[i].GetKey(), object.SysAttributePrefix) {
			res = append(res, *NewAttributeFromV2(&attrs[i]))
		}
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
//
// See also [Object.SetPreviousID].
// Deprecated: use [Object.GetPreviousID] instead.
func (o *Object) PreviousID() (oid.ID, bool) {
	id := o.GetPreviousID()
	return id, !id.IsZero()
}

// GetPreviousID returns identifier of the previous sibling object. Zero return
// means unset ID.
//
// See also [Object.SetPreviousID].
func (o Object) GetPreviousID() oid.ID {
	var id oid.ID
	if m := (*object.Object)(&o).GetHeader().GetSplit().GetPrevious(); m != nil {
		if err := id.ReadFromV2(*m); err != nil {
			panic(fmt.Errorf("unexpected ID decoding failure: %w", err))
		}
	}
	return id
}

// ResetPreviousID resets identifier of the previous sibling object.
//
// See also [Object.SetPreviousID], [Object.GetPreviousID].
func (o *Object) ResetPreviousID() {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetPrevious(nil)
	})
}

// SetPreviousID sets identifier of the previous sibling object.
//
// See also [Object.GetPreviousID].
func (o *Object) SetPreviousID(v oid.ID) {
	var v2 refs.ObjectID
	v.WriteToV2(&v2)

	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetPrevious(&v2)
	})
}

// Children return list of the identifiers of the child objects.
//
// See also [Object.SetChildren].
func (o Object) Children() []oid.ID {
	v2 := (*object.Object)(&o)
	ids := v2.GetHeader().GetSplit().GetChildren()

	var (
		id  oid.ID
		res = make([]oid.ID, len(ids))
	)

	for i := range ids {
		if err := id.ReadFromV2(ids[i]); err != nil {
			panic(fmt.Errorf("unexpected child#%d decoding failure: %w", i, err))
		}
		res[i] = id
	}

	return res
}

// SetChildren sets list of the identifiers of the child objects.
//
// See also [Object.Children].
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

// SetFirstID sets the first part's ID of the object's
// split chain.
//
// See also [Object.GetFirstID].
func (o *Object) SetFirstID(id oid.ID) {
	var v2 refs.ObjectID
	id.WriteToV2(&v2)

	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetFirst(&v2)
	})
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
	var id oid.ID
	if m := (*object.Object)(&o).GetHeader().GetSplit().GetFirst(); m != nil {
		if err := id.ReadFromV2(*m); err != nil {
			panic(fmt.Errorf("unexpected ID decoding failure: %w", err))
		}
	}
	return id
}

// SplitID return split identity of split object. If object is not split returns nil.
//
// See also [Object.SetSplitID].
func (o Object) SplitID() *SplitID {
	return NewSplitIDFromV2(
		(*object.Object)(&o).
			GetHeader().
			GetSplit().
			GetSplitID(),
	)
}

// SetSplitID sets split identifier for the split object.
//
// See also [Object.SplitID].
func (o *Object) SetSplitID(id *SplitID) {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetSplitID(id.ToV2())
	})
}

// ParentID returns identifier of the parent object.
//
// See also [Object.SetParentID].
// Deprecated: use [Object.GetParentID] instead.
func (o *Object) ParentID() (oid.ID, bool) {
	id := o.GetParentID()
	return id, !id.IsZero()
}

// GetParentID returns identifier of the parent object. Zero return means unset
// ID.
//
// See also [Object.SetParentID].
func (o Object) GetParentID() oid.ID {
	var id oid.ID
	if m := (*object.Object)(&o).GetHeader().GetSplit().GetParent(); m != nil {
		if err := id.ReadFromV2(*m); err != nil {
			panic(fmt.Errorf("unexpected ID decoding failure: %w", err))
		}
	}
	return id
}

// SetParentID sets identifier of the parent object.
//
// See also [Object.GetParentID].
func (o *Object) SetParentID(v oid.ID) {
	var v2 refs.ObjectID
	v.WriteToV2(&v2)

	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetParent(&v2)
	})
}

// ResetParentID removes identifier of the parent object.
//
// See also [Object.SetParentID].
func (o *Object) ResetParentID() {
	o.setSplitFields(func(split *object.SplitHeader) {
		split.SetParent(nil)
	})
}

// Parent returns parent object w/o payload.
//
// See also [Object.SetParent].
func (o Object) Parent() *Object {
	h := (*object.Object)(&o).
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
//
// See also [Object.Parent].
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

// SessionToken returns token of the session within which object was created.
//
// See also [Object.SetSessionToken].
func (o Object) SessionToken() *session.Object {
	tokv2 := (*object.Object)(&o).GetHeader().GetSessionToken()
	if tokv2 == nil {
		return nil
	}

	var res session.Object

	_ = res.ReadFromV2(*tokv2)

	return &res
}

// SetSessionToken sets token of the session within which object was created.
//
// See also [Object.SessionToken].
func (o *Object) SetSessionToken(v *session.Object) {
	o.setHeaderField(func(h *object.Header) {
		var tokv2 *v2session.Token

		if v != nil {
			tokv2 = new(v2session.Token)
			v.WriteToV2(tokv2)
		}

		h.SetSessionToken(tokv2)
	})
}

// Type returns type of the object.
//
// See also [Object.SetType].
func (o Object) Type() Type {
	return Type((*object.Object)(&o).
		GetHeader().
		GetObjectType())
}

// SetType sets type of the object.
//
// See also [Object.Type].
func (o *Object) SetType(v Type) {
	o.setHeaderField(func(h *object.Header) {
		h.SetObjectType(object.Type(v))
	})
}

// CutPayload returns [Object] w/ empty payload.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (o *Object) CutPayload() *Object {
	ov2 := new(object.Object)
	*ov2 = *(*object.Object)(o)
	ov2.SetPayload(nil)

	return (*Object)(ov2)
}

// HasParent checks if parent (split ID) is present.
func (o Object) HasParent() bool {
	return (*object.Object)(&o).
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
//
// See also [Object.Unmarshal].
func (o Object) Marshal() []byte {
	return (*object.Object)(&o).StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of object.
//
// See also [Object.Marshal].
func (o *Object) Unmarshal(data []byte) error {
	var m object.Object
	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return o.ReadFromV2(m)
}

// MarshalJSON encodes object to protobuf JSON format.
//
// See also [Object.UnmarshalJSON].
func (o Object) MarshalJSON() ([]byte, error) {
	return (*object.Object)(&o).MarshalJSON()
}

// UnmarshalJSON decodes object from protobuf JSON format.
//
// See also [Object.MarshalJSON].
func (o *Object) UnmarshalJSON(data []byte) error {
	var m object.Object
	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return o.ReadFromV2(m)
}

// HeaderLen returns length of the binary header.
func (o Object) HeaderLen() int {
	return (*object.Object)(&o).GetHeader().StableSize()
}
