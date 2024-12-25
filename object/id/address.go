package oid

import (
	"errors"
	"fmt"
	"strings"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"google.golang.org/protobuf/encoding/protojson"
)

// Address represents global object identifier in NeoFS network. Each object
// belongs to exactly one container and is uniquely addressed within the container.
// Zero Address is usually prohibited, see docs for details.
//
// ID implements built-in comparable interface.
//
// Address is mutually compatible with [refs.Address] message. See
// [Address.FromProtoMessage] / [Address.ProtoMessage] methods.
type Address struct {
	cnr cid.ID

	obj ID
}

// ErrZeroAddress is an error returned on zero [Address] encounter.
var ErrZeroAddress = errors.New("zero object address")

// NewAddress constructs new Address.
func NewAddress(cnr cid.ID, obj ID) Address { return Address{cnr, obj} }

// DecodeAddressString creates new Address and makes [Address.DecodeString].
func DecodeAddressString(s string) (Address, error) {
	var id Address
	return id, id.DecodeString(s)
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [Address.ProtoMessage].
func (x *Address) FromProtoMessage(m *refs.Address) error {
	if m.ContainerId == nil {
		return errors.New("missing container ID")
	}

	if m.ObjectId == nil {
		return errors.New("missing object ID")
	}

	err := x.cnr.FromProtoMessage(m.ContainerId)
	if err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	err = x.obj.FromProtoMessage(m.ObjectId)
	if err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}

	return nil
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [Address.FromProtoMessage].
func (x Address) ProtoMessage() *refs.Address {
	return &refs.Address{
		ContainerId: x.cnr.ProtoMessage(),
		ObjectId:    x.obj.ProtoMessage(),
	}
}

// MarshalJSON encodes Address into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x Address) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(x)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Address
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *Address) UnmarshalJSON(data []byte) error {
	var m refs.Address

	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	return x.FromProtoMessage(&m)
}

// Container returns unique identifier of the NeoFS object container.
//
// Zero Address has zero container ID, which is incorrect according to NeoFS
// API protocol.
//
// See also SetContainer.
func (x Address) Container() cid.ID {
	return x.cnr
}

// SetContainer sets unique identifier of the NeoFS object container.
//
// See also Container.
func (x *Address) SetContainer(id cid.ID) {
	x.cnr = id
}

// Object returns unique identifier of the object in the container
// identified by Container().
//
// Zero Address has zero object ID, which is incorrect according to NeoFS
// API protocol.
//
// See also SetObject.
func (x Address) Object() ID {
	return x.obj
}

// SetObject sets unique identifier of the object in the container
// identified by Container().
//
// See also Object.
func (x *Address) SetObject(id ID) {
	x.obj = id
}

// delimiter of container and object IDs in Address protocol string.
const idDelimiter = "/"

// EncodeToString encodes Address into NeoFS API protocol string: concatenation
// of the string-encoded container and object IDs delimited by a slash.
//
// See also DecodeString.
func (x Address) EncodeToString() string {
	return x.cnr.EncodeToString() + "/" + x.obj.EncodeToString()
}

// DecodeString decodes string into Address according to NeoFS API protocol.
// Returns an error if s is malformed. Use [DecodeAddressString] to decode s
// into a new ID.
//
// See also DecodeString.
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
// SDK versions. String MAY return same result as EncodeToString. String MUST NOT
// be used to encode Address into NeoFS protocol string.
func (x Address) String() string {
	return x.EncodeToString()
}
