package object

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// common interface of [Object] and [Header] allowing to use any of them.
type objectOrHeader interface {
	Attribute(string) string
}

// common interface of [*Object] and [*Header] allowing to use any of them.
type objectOrHeaderPtr interface {
	objectOrHeader
	SetAttribute(string, string)
}

// Object represents in-memory structure of the NeoFS object.
//
// Object is mutually compatible with [object.Object] message. See
// [Object.ReadFromV2] / [Object.WriteToV2] methods.
type Object struct {
	Header

	id      oid.ID
	sig     neofscrypto.Signature
	payload []byte

	idSet, sigSet bool
}

// New constructs new Object owned by the specified user and associated with
// particular container.
func New(cnr cid.ID, owner user.ID) Object {
	var obj Object
	obj.version, obj.verSet = version.Current, true
	obj.SetContainerID(cnr)
	obj.SetOwnerID(owner)
	return obj
}

// CopyTo writes deep copy of the [Object] to dst.
func (o Object) CopyTo(dst *Object) {
	o.Header.CopyTo(&dst.Header)
	if dst.idSet = o.idSet; dst.idSet {
		dst.id = o.id
	}
	if dst.sigSet = o.sigSet; dst.sigSet {
		o.sig.CopyTo(&dst.sig)
	}
	dst.payload = bytes.Clone(o.Payload())
}

func (o *Object) readFromV2(m *object.Object, checkFieldPresence bool) error {
	if m.Header != nil {
		if err := o.Header.readFromV2(m.Header, checkFieldPresence); err != nil {
			return fmt.Errorf("invalid header: %w", err)
		}
	} else {
		o.Header = Header{}
	}
	if o.idSet = m.ObjectId != nil; o.idSet {
		if err := o.id.ReadFromV2(m.ObjectId); err != nil {
			return fmt.Errorf("invalid ID: %w", err)
		}
	}
	if o.sigSet = m.Signature != nil; o.sigSet {
		if err := o.sig.ReadFromV2(m.Signature); err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}
	o.payload = m.Payload
	return nil
}

// ReadFromV2 reads Object from the object.Object message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Object.WriteToV2].
func (o *Object) ReadFromV2(m *object.Object) error {
	return o.readFromV2(m, true)
}

// WriteToV2 writes Object to the object.Object message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Object.ReadFromV2].
func (o Object) WriteToV2(m *object.Object) {
	if o.idSet {
		m.ObjectId = new(refs.ObjectID)
		o.id.WriteToV2(m.ObjectId)
	} else {
		m.ObjectId = nil
	}
	if o.sigSet {
		m.Signature = new(refs.Signature)
		o.sig.WriteToV2(m.Signature)
	}
	// the header may be zero, but this is cumbersome to check, and the pointer to
	// an empty structure does not differ from nil when transmitted
	m.Header = new(object.Header)
	o.Header.WriteToV2(m.Header)
	m.Payload = o.payload
}

// MarshaledSize returns length of the Object encoded into the binary format of
// the NeoFS API protocol (Protocol Buffers V3 with direct field order).
//
// See also [Object.Marshal].
func (o Object) MarshaledSize() int {
	var m object.Object
	o.WriteToV2(&m)
	return m.MarshaledSize()
}

// Marshal encodes Object into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also [Object.Unmarshal].
func (o Object) Marshal() []byte {
	var m object.Object
	o.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Object. Returns an
// error describing a format violation of the specified fields. Unmarshal does
// not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Object.Marshal].
func (o *Object) Unmarshal(data []byte) error {
	var m object.Object
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return o.readFromV2(&m, false)
}

// MarshalJSON encodes Object into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Object.UnmarshalJSON].
func (o Object) MarshalJSON() ([]byte, error) {
	var m object.Object
	o.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Object (Protocol
// Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Object.MarshalJSON].
func (o *Object) UnmarshalJSON(data []byte) error {
	var m object.Object
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}
	return o.readFromV2(&m, false)
}

// ID returns object identifier. Zero return indicates ID absence.
//
// See also [Object.SetID].
func (o Object) ID() oid.ID {
	if o.idSet {
		return o.id
	}
	return oid.ID{}
}

// SetID sets object identifier.
//
// See also [Object.ID], [Object.ResetID].
func (o *Object) SetID(v oid.ID) {
	o.id, o.idSet = v, true
}

// Signature returns object signature. Zero-scheme return indicates signature
// absence.
//
// See also [Object.SetSignature].
func (o Object) Signature() neofscrypto.Signature {
	if o.sigSet {
		return o.sig
	}
	return neofscrypto.Signature{}
}

// SetSignature sets object signature.
//
// See also [Object.Signature], [Sign], [VerifySignature].
func (o *Object) SetSignature(sig neofscrypto.Signature) {
	o.sig, o.sigSet = sig, true
}

// SignedData returns signed data of the given Object.
//
// See also [Sign].
func SignedData(obj Object) []byte {
	var m refs.ObjectID
	id := obj.ID()
	if id != [sha256.Size]byte{} {
		id.WriteToV2(&m)
	}
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Sign calculates and sets signature of the given object using provided signer.
//
// See also [VerifySignature], [Object.SetSignature].
func Sign(obj *Object, signer neofscrypto.Signer) error {
	var sig neofscrypto.Signature
	err := sig.Calculate(signer, SignedData(*obj))
	if err == nil {
		obj.SetSignature(sig)
	}
	return err
}

// VerifySignature checks whether signature of the given object is presented and
// valid.
//
// See also [Sign], [Object.Signature].
func VerifySignature(obj Object) bool {
	sig := obj.Signature()
	return sig.Scheme() != 0 && sig.Verify(SignedData(obj))
}

// Payload returns payload bytes.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [Object.SetPayload].
func (o Object) Payload() []byte {
	return o.payload
}

// SetPayload sets payload bytes.
//
// See also [Object.Payload].
func (o *Object) SetPayload(v []byte) {
	o.payload = v
}
