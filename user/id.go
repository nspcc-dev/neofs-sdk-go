package user

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// ID identifies users of the NeoFS system.
//
// ID is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.OwnerID
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration. Zero ID is not valid,
// so it MUST be initialized using some modifying function (e.g. SetScriptHash, etc.).
type ID struct {
	w []byte
}

// ReadFromV2 reads ID from the refs.OwnerID message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *ID) ReadFromV2(m refs.OwnerID) error {
	w := m.GetValue()
	if len(w) != 25 {
		return fmt.Errorf("invalid length %d, expected 25", len(w))
	}

	if w[0] != address.NEO3Prefix {
		return fmt.Errorf("invalid prefix byte 0x%X, expected 0x%X", w[0], address.NEO3Prefix)
	}

	if !bytes.Equal(w[21:], hash.Checksum(w[:21])) {
		return errors.New("checksum mismatch")
	}

	x.w = w

	return nil
}

// WriteToV2 writes ID to the refs.OwnerID message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x ID) WriteToV2(m *refs.OwnerID) {
	m.SetValue(x.w)
}

// SetScriptHash forms user ID from wallet address scripthash.
func (x *ID) SetScriptHash(scriptHash util.Uint160) {
	if cap(x.w) < 25 {
		x.w = make([]byte, 25)
	} else if len(x.w) < 25 {
		x.w = x.w[:25]
	}

	x.w[0] = address.Prefix
	copy(x.w[1:], scriptHash.BytesBE())
	copy(x.w[21:], hash.Checksum(x.w[:21]))
}

// WalletBytes returns NeoFS user ID as Neo3 wallet address in a binary format.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also Neo3 wallet docs.
func (x ID) WalletBytes() []byte {
	return x.w
}

// EncodeToString encodes ID into NeoFS API V2 protocol string.
//
// See also DecodeString.
func (x ID) EncodeToString() string {
	return base58.Encode(x.w)
}

// DecodeString decodes NeoFS API V2 protocol string. Returns an error
// if s is malformed.
//
// DecodeString always changes the ID.
//
// See also EncodeToString.
func (x *ID) DecodeString(s string) error {
	var err error

	x.w, err = base58.Decode(s)
	if err != nil {
		return fmt.Errorf("decode base58: %w", err)
	}

	return nil
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. String MAY return same result as EncodeToString. String MUST NOT
// be used to encode ID into NeoFS protocol string.
func (x ID) String() string {
	return x.EncodeToString()
}

// Equals defines a comparison relation between two ID instances.
func (x ID) Equals(x2 ID) bool {
	return bytes.Equal(x.w, x2.w)
}
