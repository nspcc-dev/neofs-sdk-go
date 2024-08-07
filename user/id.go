package user

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// IDSize is the size of an [ID] in bytes.
const IDSize = 25

// ID identifies users of the NeoFS system and represents Neo3 account address.
// Zero ID is usually prohibited, see docs for details.
//
// ID implements built-in comparable interface.
//
// ID is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.OwnerID
// message. See ReadFromV2 / WriteToV2 methods.
//
// Zero ID is not valid.
type ID [IDSize]byte

// ErrZeroID is an error returned on zero [ID] encounter.
var ErrZeroID = errors.New("zero user ID")

// NewFromScriptHash creates new ID and makes [ID.SetScriptHash].
func NewFromScriptHash(scriptHash util.Uint160) ID {
	var x ID
	x[0] = address.Prefix
	copy(x[1:], scriptHash.BytesBE())
	copy(x[21:], hash.Checksum(x[:21]))
	return x
}

// NewFromECDSAPublicKey creates new ID corresponding to Neo3 verification
// script hash of the given ECDSA public key. The point must be on the
// [elliptic.P256] curve.
func NewFromECDSAPublicKey(pub ecdsa.PublicKey) ID {
	return NewFromScriptHash((*keys.PublicKey)(&pub).GetScriptHash())
}

// DecodeString creates new ID and makes [ID.DecodeString].
func DecodeString(s string) (ID, error) {
	var id ID
	return id, id.DecodeString(s)
}

func (x *ID) decodeBytes(b []byte) error {
	switch {
	case len(b) != IDSize:
		return fmt.Errorf("invalid length %d, expected %d", len(b), IDSize)
	case b[0] != address.NEO3Prefix:
		return fmt.Errorf("invalid prefix byte 0x%X, expected 0x%X", b[0], address.NEO3Prefix)
	case !bytes.Equal(b[21:], hash.Checksum(b[:21])):
		return errors.New("checksum mismatch")
	}
	*x = ID(b)
	return nil
}

// ReadFromV2 reads ID from the refs.OwnerID message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *ID) ReadFromV2(m refs.OwnerID) error {
	err := x.decodeBytes(m.GetValue())
	if err == nil && x.IsZero() {
		err = ErrZeroID
	}
	return err
}

// WriteToV2 writes ID to the refs.OwnerID message.
// The message must not be nil.
//
// See also ReadFromV2.
func (x ID) WriteToV2(m *refs.OwnerID) {
	m.SetValue(x[:])
}

// SetScriptHash forms user ID from wallet address scripthash.
// Deprecated: use [NewFromScriptHash] instead.
func (x *ID) SetScriptHash(scriptHash util.Uint160) { *x = NewFromScriptHash(scriptHash) }

// ScriptHash gets scripthash from user ID.
func (x ID) ScriptHash() util.Uint160 {
	return util.Uint160(x[1:21])
}

// WalletBytes returns NeoFS user ID as Neo3 wallet address in a binary format.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also Neo3 wallet docs.
// Deprecated: use x[:] instead.
func (x ID) WalletBytes() []byte {
	return x[:]
}

// EncodeToString encodes ID into NeoFS API V2 protocol string.
//
// See also DecodeString.
func (x ID) EncodeToString() string {
	return base58.Encode(x[:])
}

// DecodeString decodes NeoFS API V2 protocol string. Returns an error
// if s is malformed. Use [DecodeString] to decode s into a new ID.
//
// DecodeString always changes the ID.
//
// See also EncodeToString.
func (x *ID) DecodeString(s string) error {
	b, err := base58.Decode(s)
	if err != nil {
		return fmt.Errorf("decode base58: %w", err)
	}

	return x.decodeBytes(b)
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
// Deprecated: ID is comparable.
func (x ID) Equals(x2 ID) bool {
	return x == x2
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
