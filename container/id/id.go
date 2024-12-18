package cid

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
)

// Size is the size of an [ID] in bytes.
const Size = sha256.Size

// ID represents NeoFS container identifier. Zero ID is usually prohibited, see
// docs for details.
//
// ID implements built-in comparable interface.
//
// ID is mutually compatible with [refs.ContainerID] message. See
// [ID.FromProtoMessage] / [ID.ProtoMessage] methods.
type ID [Size]byte

// ErrZero is an error returned on zero [ID] encounter.
var ErrZero = errors.New("zero container ID")

// NewFromMarshalledContainer returns new ID calculated from the given NeoFS
// container encoded into Protocol Buffers V3 with ascending order of fields by
// number. It's callers responsibility to ensure the format of b. See
// [container.Container.Marshal].
func NewFromMarshalledContainer(b []byte) ID { return sha256.Sum256(b) }

// DecodeBytes creates new ID and makes [ID.Decode].
func DecodeBytes(b []byte) (ID, error) {
	var id ID
	return id, id.Decode(b)
}

// DecodeString creates new ID and makes [ID.DecodeString].
func DecodeString(s string) (ID, error) {
	var id ID
	return id, id.DecodeString(s)
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// id from it.
//
// See also [ID.ProtoMessage].
func (id *ID) FromProtoMessage(m *refs.ContainerID) error {
	err := id.Decode(m.Value)
	if err == nil && id.IsZero() {
		err = ErrZero
	}
	return err
}

// ProtoMessage converts id into message to transmit using the NeoFS API
// protocol.
//
// See also [ID.FromProtoMessage].
func (id ID) ProtoMessage() *refs.ContainerID {
	return &refs.ContainerID{Value: id[:]}
}

// Encode encodes ID into [Size] bytes of dst. Panics if
// dst length is less than [Size].
//
// Zero ID is all zeros.
//
// See also Decode.
// Deprecated: use id[:] instead.
func (id ID) Encode(dst []byte) {
	if l := len(dst); l < Size {
		panic(fmt.Sprintf("destination length is less than %d bytes: %d", Size, l))
	}

	copy(dst, id[:])
}

// Decode decodes src bytes into ID. Use [DecodeBytes] to decode src into a new
// ID.
//
// Decode expects that src has [Size] bytes length. If the input is malformed,
// Decode returns an error describing format violation. In this case ID
// remains unchanged.
//
// Decode doesn't mutate src.
func (id *ID) Decode(src []byte) error {
	if len(src) != Size {
		return fmt.Errorf("invalid length %d", len(src))
	}

	*id = ID(src)

	return nil
}

// SetSHA256 sets container identifier value to SHA256 checksum of container structure.
// Deprecated: use direct assignment instead.
func (id *ID) SetSHA256(v [sha256.Size]byte) {
	*id = v
}

// Equals defines a comparison relation between two ID instances.
//
// Note that comparison using '==' operator is not recommended since it MAY result
// in loss of compatibility.
// Deprecated: ID is comparable.
func (id ID) Equals(id2 ID) bool {
	return id == id2
}

// EncodeToString encodes ID into NeoFS API protocol string.
//
// Zero ID is base58 encoding of [Size] zeros.
//
// See also DecodeString.
func (id ID) EncodeToString() string {
	return base58.Encode(id[:])
}

// DecodeString decodes string into ID according to NeoFS API protocol. Returns
// an error if s is malformed. Use [DecodeString] to decode s into a new ID.
//
// See also DecodeString.
func (id *ID) DecodeString(s string) error {
	data, err := base58.Decode(s)
	if err != nil {
		return fmt.Errorf("decode base58: %w", err)
	}

	return id.Decode(data)
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as EncodeToString. String MUST NOT
// be used to encode ID into NeoFS protocol string.
func (id ID) String() string {
	return id.EncodeToString()
}

// FromBinary calculates identifier of the binary-encoded container
// in CAS of the NeoFS containers and writes it into id.
//
// See also [container.Container.CalculateID], [container.Container.AssertID].
// Deprecated: use [NewFromContainerBinary].
func (id *ID) FromBinary(cnr []byte) { *id = NewFromMarshalledContainer(cnr) }

// IsZero checks whether ID is zero.
func (id ID) IsZero() bool {
	for i := range id {
		if id[i] != 0 {
			return false
		}
	}
	return true
}
