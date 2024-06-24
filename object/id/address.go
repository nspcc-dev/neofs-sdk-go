package oid

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"google.golang.org/protobuf/encoding/protojson"
)

// Address represents global object identifier in NeoFS network. Each object
// belongs to exactly one container and is uniquely addressed within the container.
//
// ID implements built-in comparable interface.
//
// Address is mutually compatible with [refs.Address] message. See
// [Address.ReadFromV2] / [Address.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type Address struct {
	cnr cid.ID

	obj ID
}

// ReadFromV2 reads ID from the [refs.Address] message. Returns an error if the
// message is malformed according to the NeoFS API V2 protocol. The message must
// not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Address.WriteToV2].
func (x *Address) ReadFromV2(m *refs.Address) error {
	if m.ContainerId == nil {
		return errors.New("missing container ID")
	}

	if m.ObjectId == nil {
		return errors.New("missing object ID")
	}

	err := x.cnr.ReadFromV2(m.ContainerId)
	if err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	err = x.obj.ReadFromV2(m.ObjectId)
	if err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}

	return nil
}

// WriteToV2 writes ID to the [refs.Address] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Address.ReadFromV2].
func (x Address) WriteToV2(m *refs.Address) {
	m.ContainerId = new(refs.ContainerID)
	x.cnr.WriteToV2(m.ContainerId)
	m.ObjectId = new(refs.ObjectID)
	x.obj.WriteToV2(m.ObjectId)
}

// MarshalJSON encodes Address into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Address.UnmarshalJSON].
func (x Address) MarshalJSON() ([]byte, error) {
	var m refs.Address
	x.WriteToV2(&m)

	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Address (Protocol
// Buffers V3 JSON). Returns an error describing a format violation.
//
// See also [Address.MarshalJSON].
func (x *Address) UnmarshalJSON(data []byte) error {
	var m refs.Address
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}

	return x.ReadFromV2(&m)
}

// Container returns unique identifier of the NeoFS object container.
//
// Zero Address has zero container ID, which is incorrect according to NeoFS
// API protocol.
//
// See also [Address.SetContainer].
func (x Address) Container() cid.ID {
	return x.cnr
}

// SetContainer sets unique identifier of the NeoFS object container.
//
// See also [Address.Container].
func (x *Address) SetContainer(id cid.ID) {
	x.cnr = id
}

// Object returns unique identifier of the object in the container identified by
// [Address.Container].
//
// Zero Address has zero object ID, which is incorrect according to NeoFS
// API protocol.
//
// See also [Address.SetObject].
func (x Address) Object() ID {
	return x.obj
}

// SetObject sets unique identifier of the object in the container identified by
// [Address.Container].
//
// See also [Address.Object].
func (x *Address) SetObject(id ID) {
	x.obj = id
}

// delimiter of container and object IDs in Address protocol string.
const idDelimiter = "/"

// EncodeToString encodes Address into NeoFS API protocol string: concatenation
// of the string-encoded container and object IDs delimited by a slash.
//
// See also [Address.DecodeString].
func (x Address) EncodeToString() string {
	return x.cnr.EncodeToString() + "/" + x.obj.EncodeToString()
}

// DecodeString decodes string into Address according to NeoFS API protocol. Returns
// an error if s is malformed.
//
// See also [Address.EncodeToString].
func (x *Address) DecodeString(s string) error {
	indDelimiter := strings.Index(s, idDelimiter)
	if indDelimiter < 0 {
		return errors.New("missing delimiter")
	}

	err := x.cnr.DecodeString(s[:indDelimiter])
	if err != nil {
		return fmt.Errorf("decode container string: %w", err)
	}

	err = x.obj.DecodeString(s[indDelimiter+1:])
	if err != nil {
		return fmt.Errorf("decode object string: %w", err)
	}

	return nil
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as [Address.EncodeToString].
// String MUST NOT be used to encode Address into NeoFS protocol string.
func (x Address) String() string {
	return x.EncodeToString()
}
