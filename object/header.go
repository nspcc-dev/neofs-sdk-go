package object

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	"github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	apisession "github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MaxHeaderLen is a maximum allowed length of binary object header to be
// created via NeoFS API protocol. See [Header.MarshaledSize].
const MaxHeaderLen = 16 << 10

// Various system attributes.
const (
	sysAttributePrefix       = "__NEOFS__"
	attributeExpirationEpoch = sysAttributePrefix + "EXPIRATION_EPOCH"
)

// Header groups meta information about particular NeoFS object required for
// system storage and data access. Each Header is associated with exactly one
// Object.
type Header struct {
	version       version.Version
	container     cid.ID
	owner         user.ID
	attrs         []*object.Header_Attribute
	creationEpoch uint64
	session       session.Object
	// payload meta
	typ                 Type
	payloadSize         uint64
	payloadChecksum     checksum.Checksum
	payloadHomoChecksum checksum.Checksum
	// split-chain relations
	splitFirst    oid.ID
	splitPrevious oid.ID
	parentID      oid.ID
	parentSig     neofscrypto.Signature
	parentHdr     *Header
	splitID       []byte   // deprecated
	children      []oid.ID // deprecated

	verSet, cnrSet, ownerSet, sessionSet, csSet, csHomoSet,
	splitFirstSet, splitPreviousSet, parentIDSet, parentSigSet bool
}

// CopyTo writes deep copy of the Header to dst.
func (x Header) CopyTo(dst *Header) {
	if dst.verSet = x.verSet; dst.verSet {
		dst.version = x.version
	}
	if dst.cnrSet = x.cnrSet; dst.cnrSet {
		dst.container = x.container
	}
	if dst.ownerSet = x.ownerSet; dst.ownerSet {
		dst.owner = x.owner
	}
	if dst.sessionSet = x.sessionSet; dst.sessionSet {
		x.session.CopyTo(&dst.session)
	}
	if dst.csSet = x.csSet; dst.csSet {
		x.payloadChecksum.CopyTo(&dst.payloadChecksum)
	}
	if dst.csHomoSet = x.csHomoSet; dst.csHomoSet {
		x.payloadHomoChecksum.CopyTo(&dst.payloadHomoChecksum)
	}
	if dst.splitFirstSet = x.splitFirstSet; dst.splitFirstSet {
		dst.splitFirst = x.splitFirst
	}
	if dst.splitPreviousSet = x.splitPreviousSet; dst.splitPreviousSet {
		dst.splitPrevious = x.splitPrevious
	}
	if dst.parentIDSet = x.parentIDSet; dst.parentIDSet {
		dst.parentID = x.parentID
	}
	if dst.parentSigSet = x.parentSigSet; dst.parentSigSet {
		x.parentSig.CopyTo(&dst.parentSig)
	}
	if x.parentHdr != nil {
		dst.parentHdr = new(Header)
		x.parentHdr.CopyTo(dst.parentHdr)
	} else {
		dst.parentHdr = nil
	}
	if x.attrs != nil {
		dst.attrs = make([]*object.Header_Attribute, len(x.attrs))
		for i := range x.attrs {
			if x.attrs[i] != nil {
				dst.attrs[i] = &object.Header_Attribute{Key: x.attrs[i].Key, Value: x.attrs[i].Value}
			}
		}
	} else {
		dst.attrs = nil
	}
	if x.children != nil {
		dst.children = make([]oid.ID, len(x.children))
		copy(dst.children, x.children)
	} else {
		dst.children = nil
	}
	dst.creationEpoch = x.creationEpoch
	dst.payloadSize = x.payloadSize
	dst.typ = x.typ
	dst.splitID = bytes.Clone(x.splitID)
}

func (x *Header) readFromV2(m *object.Header, checkFieldPresence bool) error {
	if m.ObjectType < 0 {
		return fmt.Errorf("invalid type field %d", m.ObjectType)
	}
	if x.cnrSet = m.ContainerId != nil; x.cnrSet {
		if err := x.container.ReadFromV2(m.ContainerId); err != nil {
			return fmt.Errorf("invalid container: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing container")
	}
	if x.ownerSet = m.OwnerId != nil; x.ownerSet {
		if err := x.owner.ReadFromV2(m.OwnerId); err != nil {
			return fmt.Errorf("invalid owner: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing owner")
	}
	if x.verSet = m.Version != nil; x.verSet {
		if err := x.version.ReadFromV2(m.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}
	if x.csSet = m.PayloadHash != nil; x.csSet {
		if err := x.payloadChecksum.ReadFromV2(m.PayloadHash); err != nil {
			return fmt.Errorf("invalid payload checksum: %w", err)
		}
	}
	if x.csHomoSet = m.HomomorphicHash != nil; x.csHomoSet {
		if err := x.payloadHomoChecksum.ReadFromV2(m.HomomorphicHash); err != nil {
			return fmt.Errorf("invalid payload homomorphic checksum: %w", err)
		}
	}
	if x.sessionSet = m.SessionToken != nil; x.sessionSet {
		if err := x.session.ReadFromV2(m.SessionToken); err != nil {
			return fmt.Errorf("invalid session: %w", err)
		}
	}
	if m.Split != nil {
		if x.splitID = m.Split.SplitId; len(x.splitID) > 0 {
			if len(x.splitID) != 16 {
				return fmt.Errorf("invalid split-chain ID: wrong length %d", len(x.splitID))
			}
			if ver := x.splitID[6] >> 4; ver != 4 {
				return fmt.Errorf("invalid split-chain ID: wrong version #%d", ver)
			}
		}
		if x.parentIDSet = m.Split.Parent != nil; x.parentIDSet {
			if err := x.parentID.ReadFromV2(m.Split.Parent); err != nil {
				return fmt.Errorf("invalid parent ID: %w", err)
			}
		}
		if x.parentSigSet = m.Split.ParentSignature != nil; x.parentSigSet {
			if err := x.parentSig.ReadFromV2(m.Split.ParentSignature); err != nil {
				return fmt.Errorf("invalid parent signature: %w", err)
			}
		}
		if x.splitPreviousSet = m.Split.Previous != nil; x.splitPreviousSet {
			if err := x.splitPrevious.ReadFromV2(m.Split.Previous); err != nil {
				return fmt.Errorf("invalid previous split-chain element: %w", err)
			}
		}
		if x.splitFirstSet = m.Split.First != nil; x.splitFirstSet {
			if err := x.splitFirst.ReadFromV2(m.Split.First); err != nil {
				return fmt.Errorf("invalid first split-chain element: %w", err)
			}
		}
		if len(m.Split.Children) > 0 {
			x.children = make([]oid.ID, len(m.Split.Children))
			for i := range m.Split.Children {
				if m.Split.Children[i] == nil {
					return fmt.Errorf("nil child split-chain element #%d", i)
				}
				if err := x.children[i].ReadFromV2(m.Split.Children[i]); err != nil {
					return fmt.Errorf("invalid child split-chain element #%d: %w", i, err)
				}
			}
		} else {
			x.children = nil
		}
		if m.Split.ParentHeader != nil {
			if x.parentHdr == nil {
				x.parentHdr = new(Header)
			}
			if err := x.parentHdr.readFromV2(m.Split.ParentHeader, checkFieldPresence); err != nil {
				return fmt.Errorf("invalid parent header: %w", err)
			}
		} else {
			x.parentHdr = nil
		}
	}
	for i := range m.Attributes {
		key := m.Attributes[i].GetKey()
		if key == "" {
			return fmt.Errorf("invalid attribute #%d: missing key", i)
		} // also prevents further NPE
		for j := 0; j < i; j++ {
			if m.Attributes[j].Key == key {
				return fmt.Errorf("multiple attributes with key=%s", key)
			}
		}
		if m.Attributes[i].Value == "" {
			return fmt.Errorf("invalid attribute #%d (%s): missing value", i, key)
		}
		switch key {
		case attributeExpirationEpoch:
			if _, err := strconv.ParseUint(m.Attributes[i].Value, 10, 64); err != nil {
				return fmt.Errorf("invalid expiration attribute (#%d): invalid integer (%w)", i, err)
			}
		case attributeTimestamp:
			if _, err := strconv.ParseInt(m.Attributes[i].Value, 10, 64); err != nil {
				return fmt.Errorf("invalid timestamp attribute (#%d): invalid integer (%w)", i, err)
			}
		}
	}
	x.attrs = m.Attributes
	x.typ = Type(m.ObjectType)
	x.creationEpoch = m.CreationEpoch
	x.payloadSize = m.PayloadLength
	return nil
}

// ReadFromV2 reads Header from the [object.Header] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Header.WriteToV2].
func (x *Header) ReadFromV2(m *object.Header) error {
	return x.readFromV2(m, true)
}

// WriteToV2 writes Header to the [object.Header] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Header.ReadFromV2].
func (x Header) WriteToV2(m *object.Header) {
	if x.cnrSet {
		m.ContainerId = new(refs.ContainerID)
		x.container.WriteToV2(m.ContainerId)
	} else {
		m.ContainerId = nil
	}
	if x.ownerSet {
		m.OwnerId = new(refs.OwnerID)
		x.owner.WriteToV2(m.OwnerId)
	} else {
		m.OwnerId = nil
	}
	if x.verSet {
		m.Version = new(refs.Version)
		x.version.WriteToV2(m.Version)
	} else {
		m.Version = nil
	}
	if x.csSet {
		m.PayloadHash = new(refs.Checksum)
		x.payloadChecksum.WriteToV2(m.PayloadHash)
	} else {
		m.PayloadHash = nil
	}
	if x.csHomoSet {
		m.HomomorphicHash = new(refs.Checksum)
		x.payloadHomoChecksum.WriteToV2(m.HomomorphicHash)
	} else {
		m.HomomorphicHash = nil
	}
	if x.sessionSet {
		m.SessionToken = new(apisession.SessionToken)
		x.session.WriteToV2(m.SessionToken)
	} else {
		m.SessionToken = nil
	}
	if x.parentIDSet || x.splitPreviousSet || x.splitFirstSet || x.parentSigSet || x.parentHdr != nil ||
		len(x.splitID) > 0 || len(x.children) > 0 {
		m.Split = &object.Header_Split{
			SplitId: x.splitID,
		}
		if x.parentIDSet {
			m.Split.Parent = new(refs.ObjectID)
			x.parentID.WriteToV2(m.Split.Parent)
		}
		if x.splitPreviousSet {
			m.Split.Previous = new(refs.ObjectID)
			x.splitPrevious.WriteToV2(m.Split.Previous)
		}
		if x.splitFirstSet {
			m.Split.First = new(refs.ObjectID)
			x.splitFirst.WriteToV2(m.Split.First)
		}
		if x.parentSigSet {
			m.Split.ParentSignature = new(refs.Signature)
			x.parentSig.WriteToV2(m.Split.ParentSignature)
		}
		if x.parentHdr != nil {
			m.Split.ParentHeader = new(object.Header)
			x.parentHdr.WriteToV2(m.Split.ParentHeader)
		}
		if len(x.children) > 0 {
			m.Split.Children = make([]*refs.ObjectID, len(x.children))
			for i := range x.children {
				m.Split.Children[i] = new(refs.ObjectID)
				x.children[i].WriteToV2(m.Split.Children[i])
			}
		} else {
			m.Split.Children = nil
		}
	} else {
		m.Split = nil
	}
	m.Attributes = x.attrs
	m.ObjectType = object.ObjectType(x.typ)
	m.CreationEpoch = x.creationEpoch
	m.PayloadLength = x.payloadSize
}

// MarshaledSize returns length of the Header encoded into the binary format of
// the NeoFS API protocol (Protocol Buffers V3 with direct field order).
//
// See also [Header.Marshal].
func (x Header) MarshaledSize() int {
	var m object.Header
	x.WriteToV2(&m)
	return m.MarshaledSize()
}

// Marshal encodes Header into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also [Header.MarshaledSize], [Header.Unmarshal].
func (x Header) Marshal() []byte {
	var m object.Header
	x.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Header. Returns an
// error describing a format violation of the specified fields. Unmarshal does
// not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Header.Marshal].
func (x *Header) Unmarshal(data []byte) error {
	var m object.Header
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return x.readFromV2(&m, false)
}

// MarshalJSON encodes Header into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Header.UnmarshalJSON].
func (x Header) MarshalJSON() ([]byte, error) {
	var m object.Header
	x.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Header (Protocol
// Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Header.MarshalJSON].
func (x *Header) UnmarshalJSON(data []byte) error {
	var m object.Header
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}
	return x.readFromV2(&m, false)
}

// PayloadSize returns payload length of the object in bytes. Note that
// PayloadSize may differ from actual length of the object payload bytes.
//
// See also [Header.SetPayloadSize].
func (x Header) PayloadSize() uint64 {
	return x.payloadSize
}

// SetPayloadSize sets payload length of the object in bytes. Note that
// SetPayloadSize does not affect actual object payload bytes.
//
// See also [Header.PayloadSize].
func (x *Header) SetPayloadSize(v uint64) {
	x.payloadSize = v
}

// ContainerID returns identifier of the container with which the object is
// associated. Zero return indicates no binding.
//
// See also [Header.SetContainerID].
func (x Header) ContainerID() cid.ID {
	if x.cnrSet {
		return x.container
	}
	return cid.ID{}
}

// SetContainerID associates the object with the referenced container.
//
// See also [Header.ContainerID].
func (x *Header) SetContainerID(v cid.ID) {
	x.container, x.cnrSet = v, true
}

// OwnerID returns identifier of the object owner. Zero return indicates no
// binding.
//
// See also [Header.SetOwnerID].
func (x Header) OwnerID() user.ID {
	if x.ownerSet {
		return x.owner
	}
	return user.ID{}
}

// SetOwnerID sets identifier of the object owner.
//
// See also [Header.OwnerID].
func (x *Header) SetOwnerID(v user.ID) {
	x.owner, x.ownerSet = v, true
}

// CreationEpoch returns number of the NeoFS epoch when object was created.
//
// See also [Header.SetCreationEpoch].
func (x Header) CreationEpoch() uint64 {
	return x.creationEpoch
}

// SetCreationEpoch sets number of the NeoFS epoch when object was created.
//
// See also [Header.CreationEpoch].
func (x *Header) SetCreationEpoch(v uint64) {
	x.creationEpoch = v
}

// PayloadChecksum returns checksum of the object payload. Zero-type return
// indicates checksum absence.
//
// See also [Header.SetPayloadChecksum].
func (x Header) PayloadChecksum() checksum.Checksum {
	if x.csSet {
		return x.payloadChecksum
	}
	return checksum.Checksum{}
}

// SetPayloadChecksum sets checksum of the object payload.
//
// See also [Header.PayloadChecksum].
func (x *Header) SetPayloadChecksum(v checksum.Checksum) {
	x.payloadChecksum, x.csSet = v, true
}

// PayloadHomomorphicChecksum returns homomorphic checksum of the object
// payload. Zero-type return indicates checksum absence.
//
// See also [Header.SetPayloadHomomorphicChecksum].
func (x Header) PayloadHomomorphicChecksum() checksum.Checksum {
	if x.csHomoSet {
		return x.payloadHomoChecksum
	}
	return checksum.Checksum{}
}

// SetPayloadHomomorphicChecksum sets homomorphic checksum of the object
// payload.
//
// See also [Header.PayloadHomomorphicChecksum].
func (x *Header) SetPayloadHomomorphicChecksum(v checksum.Checksum) {
	x.payloadHomoChecksum, x.csHomoSet = v, true
}

// SetAttribute sets object attribute value by key. Both key and value MUST NOT
// be empty. Attributes set by the creator (owner) are most commonly ignored by
// the NeoFS system and used for application layer. Some attributes are
// so-called system or well-known attributes: they are reserved for system
// needs. System attributes SHOULD NOT be modified using SetAttribute, use
// corresponding methods/functions. List of the reserved keys is documented in
// the particular protocol version.
//
// SetAttribute overwrites existing attribute value.
//
// See also [Header.Attribute], [Header.IterateAttributes].
func (x *Header) SetAttribute(key, value string) {
	if key == "" {
		panic("empty attribute key")
	} else if value == "" {
		panic("empty attribute value")
	}

	for i := range x.attrs {
		if x.attrs[i].GetKey() == key {
			x.attrs[i].Value = value
			return
		}
	}

	x.attrs = append(x.attrs, &object.Header_Attribute{Key: key, Value: value})
}

// Attribute reads value of the object attribute by key. Empty result means
// attribute absence.
//
// See also [Header.SetAttribute], [Header.IterateAttributes].
func (x Header) Attribute(key string) string {
	for i := range x.attrs {
		if x.attrs[i].GetKey() == key {
			return x.attrs[i].GetValue()
		}
	}
	return ""
}

// NumberOfAttributes returns number of all attributes specified for this
// object.
//
// See also [Header.SetAttribute], [Header.IterateAttributes].
func (x Header) NumberOfAttributes() int {
	return len(x.attrs)
}

// IterateAttributes iterates over all object attributes and passes them into f.
// The handler MUST NOT be nil.
//
// See also [Header.SetAttribute], [Header.Attribute],
// [Header.NumberOfAttributes].
func (x Header) IterateAttributes(f func(key, val string)) {
	for i := range x.attrs {
		f(x.attrs[i].GetKey(), x.attrs[i].GetValue())
	}
}

// SetExpirationEpoch sets NeoFS epoch when the object becomes expired. By
// default, objects never expires.
//
// Reaction of NeoFS system components to the objects' 'expired' property may
// vary. For example, in the basic scenario, expired objects are auto-deleted
// from the storage. Detailed behavior can be found in the NeoFS Specification.
//
// Note that the value determines exactly the last epoch of the object's
// relevance: for example, with the value N, the object is relevant in epoch N
// and expired in any epoch starting from N+1.
func (x *Header) SetExpirationEpoch(epoch uint64) {
	x.SetAttribute(attributeExpirationEpoch, strconv.FormatUint(epoch, 10))
}

// ExpirationEpoch returns last NeoFS epoch of the object lifetime. Zero return
// means the object will never expire. For more details see
// [Header.SetExpirationEpoch].
func (x Header) ExpirationEpoch() uint64 {
	var epoch uint64
	if s := x.Attribute(attributeExpirationEpoch); s != "" {
		var err error
		epoch, err = strconv.ParseUint(s, 10, 64)
		if err != nil {
			// this could happen due to package developer only
			panic(fmt.Errorf("parse expiration epoch attribute: %w", err))
		}
	}
	return epoch
}

// PreviousSplitObject returns identifier of the object that is the previous
// link in the split-chain of the common parent. Zero return indicates no
// relation.
//
// See also [Header.SetPreviousSplitObject].
func (x Header) PreviousSplitObject() oid.ID {
	if x.splitPreviousSet {
		return x.splitPrevious
	}
	return oid.ID{}
}

// SetPreviousSplitObject sets identifier of the object that is the previous
// link in the split-chain of the common parent.
//
// See also [Header.SetPreviousSplitObject].
func (x *Header) SetPreviousSplitObject(v oid.ID) {
	x.splitPrevious, x.splitPreviousSet = v, true
}

// SetFirstSplitObject sets identifier of the object that is the first link in
// the split-chain of the common parent.
//
// See also [Header.FirstSplitObject].
func (x *Header) SetFirstSplitObject(id oid.ID) {
	x.splitFirst, x.splitFirstSet = id, true
}

// FirstSplitObject sets identifier of the object that is the first link in the
// split-chain of the common parent. Zero return indicates no relation.
//
// See also [Header.SetFirstSplitObject].
func (x Header) FirstSplitObject() oid.ID {
	if x.splitFirstSet {
		return x.splitFirst
	}
	return oid.ID{}
}

// ParentID returns identifier of the parent object of which this object is the
// split-chain part. Zero return indicates no relation.
//
// See also [Header.SetParentID].
func (x Header) ParentID() oid.ID {
	if x.parentIDSet {
		return x.parentID
	}
	return oid.ID{}
}

// SetParentID sets identifier of the parent object of which this object is the
// split-chain part.
//
// See also [Header.ParentID].
func (x *Header) SetParentID(v oid.ID) {
	x.parentID, x.parentIDSet = v, true
}

// ParentSignature returns signature of the parent object of which this object
// is the split-chain part. Zero-scheme return indicates signature absence.
//
// See also [Header.SetParentSignature].
func (x Header) ParentSignature() neofscrypto.Signature {
	if x.parentSigSet {
		return x.parentSig
	}
	return neofscrypto.Signature{}
}

// SetParentSignature sets signature of the parent object of which this object
// is the split-chain part.
//
// See also [Header.ParentSignature].
func (x *Header) SetParentSignature(sig neofscrypto.Signature) {
	x.parentSig, x.parentSigSet = sig, true
}

// ParentHeader returns header of the parent object of which this object is the
// split-chain part. Second value indicates parent header's presence.
//
// See also [Header.SetParentHeader].
func (x Header) ParentHeader() (Header, bool) {
	if x.parentHdr != nil {
		return *x.parentHdr, true
	}
	return Header{}, false
}

// SetParentHeader sets header of the parent object of which this object is the
// split-chain part.
//
// See also [Object.Parent].
func (x *Header) SetParentHeader(h Header) {
	x.parentHdr = &h
}

// SessionToken returns token of the NeoFS session within which object was
// created. Second value indicates session token presence.
//
// See also [Header.SetSessionToken].
func (x Header) SessionToken() (session.Object, bool) {
	return x.session, x.sessionSet
}

// SetSessionToken sets token of the NeoFS session within which object was
// created.
//
// See also [Header.SessionToken].
func (x *Header) SetSessionToken(t session.Object) {
	x.session, x.sessionSet = t, true
}

// Type returns object type indicating payload format.
//
// See also [Header.SetType].
func (x Header) Type() Type {
	return x.typ
}

// SetType sets object type indicating payload format.
//
// See also [Header.Type].
func (x *Header) SetType(v Type) {
	x.typ = v
}

// CalculateID calculates and returns CAS ID for the object with specified
// header.
func CalculateID(hdr Header) cid.ID {
	return sha256.Sum256(hdr.Marshal())
}
