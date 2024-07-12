package oid

import (
	"crypto/sha256"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// Size is the size of an [ID] in bytes.
const Size = sha256.Size

// ID represents NeoFS object identifier in a container.
//
// ID implements built-in comparable interface.
//
// ID is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.ObjectID
// message. See ReadFromV2 / WriteToV2 methods.
type ID [Size]byte

// NewFromObjectHeaderBinary returns new ID calculated from the given NeoFS
// object header encoded into Protocol Buffers V3 with ascending order of fields
// by number. It's callers responsibility to ensure the format of b.
func NewFromObjectHeaderBinary(b []byte) ID { return sha256.Sum256(b) }

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

// ReadFromV2 reads ID from the refs.ObjectID message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol.
//
// See also WriteToV2.
func (id *ID) ReadFromV2(m refs.ObjectID) error {
	return id.Decode(m.GetValue())
}

// WriteToV2 writes ID to the refs.ObjectID message.
// The message must not be nil.
//
// See also ReadFromV2.
func (id ID) WriteToV2(m *refs.ObjectID) {
	m.SetValue(id[:])
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
// Decode expects that src has [IDSize] bytes length. If the input is malformed,
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

// SetSHA256 sets object identifier value to SHA256 checksum.
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

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as EncodeToString. String MUST NOT
// be used to encode ID into NeoFS protocol string.
func (id ID) String() string {
	return id.EncodeToString()
}

// CalculateIDSignature signs object id with provided key.
func (id ID) CalculateIDSignature(signer neofscrypto.Signer) (neofscrypto.Signature, error) {
	data, err := id.Marshal()
	if err != nil {
		return neofscrypto.Signature{}, fmt.Errorf("marshal ID: %w", err)
	}

	var sig neofscrypto.Signature

	return sig, sig.Calculate(signer, data)
}

// Marshal marshals ID into a protobuf binary form.
func (id ID) Marshal() ([]byte, error) {
	var v2 refs.ObjectID
	v2.SetValue(id[:])

	return v2.StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of ID.
func (id *ID) Unmarshal(data []byte) error {
	var v2 refs.ObjectID
	if err := v2.Unmarshal(data); err != nil {
		return err
	}

	return id.ReadFromV2(v2)
}

// MarshalJSON encodes ID to protobuf JSON format.
func (id ID) MarshalJSON() ([]byte, error) {
	var v2 refs.ObjectID
	v2.SetValue(id[:])

	return v2.MarshalJSON()
}

// UnmarshalJSON decodes ID from protobuf JSON format.
func (id *ID) UnmarshalJSON(data []byte) error {
	var v2 refs.ObjectID
	if err := v2.UnmarshalJSON(data); err != nil {
		return err
	}

	return id.ReadFromV2(v2)
}
