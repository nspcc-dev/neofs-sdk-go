package checksum

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/tzhash/tz"
)

// Checksum represents checksum of some digital data.
//
// Checksum is mutually compatible with [refs.Checksum] message. See
// [Checksum.ReadFromV2] / [Checksum.WriteToV2] methods.
//
// Instances should be created using one of the constructors.
type Checksum struct {
	typ Type
	val []byte
}

// Type represents the enumeration of checksum types.
type Type uint8

// Supported Type values.
const (
	_      Type = iota
	SHA256      // SHA256
	TZ          // Tillich-Zémor (homomorphic)
)

// NewSHA256 constructs SHA256 checksum.
func NewSHA256(h [sha256.Size]byte) Checksum {
	return Checksum{typ: SHA256, val: h[:]}
}

// NewTZ constructs Tillich-Zémor homomorphic checksum.
func NewTZ(h [tz.Size]byte) Checksum {
	return Checksum{typ: TZ, val: h[:]}
}

// NewFromHash allows to create Checksum instance from accumulated hash.Hash. It
// is the caller's responsibility to ensure that the hash matches the specified
// type.
func NewFromHash(t Type, h hash.Hash) Checksum {
	return Checksum{typ: t, val: h.Sum(nil)}
}

// CopyTo writes deep copy of the Checksum to dst.
func (c Checksum) CopyTo(dst *Checksum) {
	dst.typ = c.typ
	dst.val = bytes.Clone(c.val)
}

// ReadFromV2 reads Checksum from the refs.Checksum message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Checksum.WriteToV2].
func (c *Checksum) ReadFromV2(m *refs.Checksum) error {
	if len(m.Sum) == 0 {
		return errors.New("missing value")
	}

	switch m.Type {
	default:
		c.typ = Type(m.Type)
	case refs.ChecksumType_SHA256:
		c.typ = SHA256
	case refs.ChecksumType_TZ:
		c.typ = TZ
	}

	c.val = m.Sum

	return nil
}

// WriteToV2 writes Checksum to the refs.Checksum message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Checksum.ReadFromV2].
func (c Checksum) WriteToV2(m *refs.Checksum) {
	switch c.typ {
	default:
		m.Type = refs.ChecksumType(c.typ)
	case SHA256:
		m.Type = refs.ChecksumType_SHA256
	case TZ:
		m.Type = refs.ChecksumType_TZ
	}
	m.Sum = c.val
}

// Type returns checksum type.
//
// Zero Checksum is of zero type.
func (c Checksum) Type() Type {
	return c.typ
}

// Value returns checksum bytes.
//
// Zero Checksum has nil value.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (c Checksum) Value() []byte {
	return c.val
}

// Calculate calculates checksum of given type for passed data. Calculate panics
// on any unsupported type, use constants defined in these package only.
//
// Does not mutate the passed value.
//
// See also [NewSHA256], [NewTZ], [NewFromHash].
func Calculate(typ Type, data []byte) Checksum {
	switch typ {
	case SHA256:
		return NewSHA256(sha256.Sum256(data))
	case TZ:
		return NewTZ(tz.Sum(data))
	default:
		panic(fmt.Errorf("unsupported checksum type %v", typ))
	}
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (c Checksum) String() string {
	return fmt.Sprintf("%s:%s", c.typ, hex.EncodeToString(c.val))
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (m Type) String() string {
	var m2 refs.ChecksumType

	switch m {
	default:
		m2 = refs.ChecksumType(m)
	case TZ:
		m2 = refs.ChecksumType_TZ
	case SHA256:
		m2 = refs.ChecksumType_SHA256
	}

	return m2.String()
}
