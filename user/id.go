package user

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
)

// IDSize is a size of NeoFS user ID in bytes.
const IDSize = 25

// ID identifies users of the NeoFS system.
//
// ID implements built-in comparable interface.
//
// ID is mutually compatible with [refs.OwnerID] message. See [ID.ReadFromV2] /
// [ID.WriteToV2] methods.
//
// Instances can be created using built-in var declaration. Zero ID is not
// valid, so it MUST be initialized using [NewID] or
// [ResolveFromECDSAPublicKey].
type ID [IDSize]byte

// NewID returns the user ID for his wallet address scripthash.
func NewID(scriptHash util.Uint160) ID {
	var id ID
	id[0] = address.Prefix
	copy(id[1:], scriptHash.BytesBE())
	copy(id[21:], hash.Checksum(id[:21]))
	return id
}

func (x *ID) decodeBinary(b []byte) error {
	if len(b) != IDSize {
		return fmt.Errorf("invalid value length %d", len(b))
	}

	if b[0] != address.NEO3Prefix {
		return fmt.Errorf("invalid prefix byte 0x%X, expected 0x%X", b[0], address.NEO3Prefix)
	}

	if !bytes.Equal(b[21:], hash.Checksum(b[:21])) {
		return errors.New("value checksum mismatch")
	}

	copy(x[:], b)

	return nil
}

// ReadFromV2 reads ID from the [refs.OwnerID] message. Returns an error if the
// message is malformed according to the NeoFS API V2 protocol. The message must
// not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [ID.WriteToV2].
func (x *ID) ReadFromV2(m *refs.OwnerID) error {
	if len(m.Value) == 0 {
		return errors.New("missing value field")
	}
	return x.decodeBinary(m.Value)
}

// WriteToV2 writes ID to the [refs.OwnerID] message of the NeoFS API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [ID.ReadFromV2].
func (x ID) WriteToV2(m *refs.OwnerID) {
	m.Value = x[:]
}

// EncodeToString encodes ID into NeoFS API V2 protocol string.
//
// Zero ID is base58 encoding of [IDSize] zeros.
//
// See also [ID.DecodeString].
func (x ID) EncodeToString() string {
	return base58.Encode(x[:])
}

// DecodeString decodes string into ID according to NeoFS API protocol. Returns
// an error if s is malformed.
//
// See also [ID.EncodeToString].
func (x *ID) DecodeString(s string) error {
	var b []byte
	if s != "" {
		var err error
		b, err = base58.Decode(s)
		if err != nil {
			return fmt.Errorf("decode base58: %w", err)
		}
	}
	return x.decodeBinary(b)
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as [ID.EncodeToString]. String
// MUST NOT be used to encode ID into NeoFS protocol string.
func (x ID) String() string {
	return x.EncodeToString()
}

// IsZero checks whether ID is zero.
func (x ID) IsZero() bool {
	for i := range x {
		if x[i] != 0 {
			return false
		}
	}
	return true
}
