package cid

import (
	"crypto/sha256"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
)

// ID represents NeoFS container identifier.
//
// ID implements built-in comparable interface.
//
// ID is mutually compatible with [refs.ContainerID] message. See
// [ID.ReadFromV2] / [ID.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type ID [sha256.Size]byte

func (id *ID) decodeBinary(b []byte) error {
	if len(b) != sha256.Size {
		return fmt.Errorf("invalid value length %d", len(b))
	}
	copy(id[:], b)
	return nil
}

// ReadFromV2 reads ID from the [refs.ContainerID] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [ID.WriteToV2].
func (id *ID) ReadFromV2(m *refs.ContainerID) error {
	if len(m.Value) == 0 {
		return fmt.Errorf("missing value field")
	}
	return id.decodeBinary(m.Value)
}

// WriteToV2 writes ID to the [refs.ContainerID] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [ID.ReadFromV2].
func (id ID) WriteToV2(m *refs.ContainerID) {
	m.Value = id[:]
}

// EncodeToString encodes ID into NeoFS API protocol string.
//
// Zero ID is base58 encoding of 32 zeros.
//
// See also [ID.DecodeString].
func (id ID) EncodeToString() string {
	return base58.Encode(id[:])
}

// DecodeString decodes string into ID according to NeoFS API protocol. Returns
// an error if s is malformed.
//
// See also [ID.EncodeToString].
func (id *ID) DecodeString(s string) error {
	var b []byte
	if s != "" {
		var err error
		b, err = base58.Decode(s)
		if err != nil {
			return fmt.Errorf("decode base58: %w", err)
		}
	}
	return id.decodeBinary(b)
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as [ID.EncodeToString]. String
// MUST NOT be used to encode ID into NeoFS protocol string.
func (id ID) String() string {
	return id.EncodeToString()
}

// IsZero checks whether ID is zero.
func (id ID) IsZero() bool {
	for i := range id {
		if id[i] != 0 {
			return false
		}
	}
	return true
}
