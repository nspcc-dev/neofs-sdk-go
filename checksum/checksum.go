package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"

	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/tzhash/tz"
)

// Checksum represents checksum of some digital data.
//
// Checksum is mutually compatible with [refs.Checksum] message. See
// [Checksum.FromProtoMessage] / [Checksum.ProtoMessage] methods.
//
// Instances must be created using one of the constructors.
type Checksum struct {
	typ Type
	val []byte
}

// Type represents the enumeration
// of checksum types.
type Type int32

const (
	_            Type = iota
	SHA256            // SHA-256 hash
	TillichZemor      // Tillich-Zémor homomorphic hash
)

func typeToProto(t Type) refs.ChecksumType {
	switch t {
	default:
		return refs.ChecksumType(t)
	case SHA256:
		return refs.ChecksumType_SHA256
	case TillichZemor:
		return refs.ChecksumType_TZ
	}
}

func typeFromProto(t refs.ChecksumType) Type {
	switch t {
	default:
		return Type(t)
	case refs.ChecksumType_SHA256:
		return SHA256
	case refs.ChecksumType_TZ:
		return TillichZemor
	}
}

// New constructs new Checksum instance. It is the caller's responsibility to
// ensure that the hash matches the type.
func New(typ Type, hsh []byte) Checksum {
	return Checksum{typ: typ, val: hsh}
}

// NewSHA256 constructs new Checksum from SHA-256 hash.
func NewSHA256(h [sha256.Size]byte) Checksum {
	return New(SHA256, h[:])
}

// NewTillichZemor constructs new Checksum from Tillich-Zémor homomorphic hash.
func NewTillichZemor(h [tz.Size]byte) Checksum {
	return New(TillichZemor, h[:])
}

// NewFromHash constructs new Checksum of specified type from accumulated
// hash.Hash. It is the caller's responsibility to ensure that the hash matches
// the type.
func NewFromHash(t Type, h hash.Hash) Checksum {
	return New(t, h.Sum(nil))
}

// NewFromData calculates Checksum of given type for specified data. The typ
// must be an enum value declared in current package.
func NewFromData(typ Type, data []byte) (Checksum, error) {
	switch typ {
	default:
		return Checksum{}, fmt.Errorf("unsupported checksum type %d", typ)
	case SHA256:
		return NewSHA256(sha256.Sum256(data)), nil
	case TillichZemor:
		return NewTillichZemor(tz.Sum(data)), nil
	}
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// c from it.
//
// See also [Checksum.ProtoMessage].
func (c *Checksum) FromProtoMessage(m *refs.Checksum) error {
	if m.Type < 0 {
		return fmt.Errorf("negative type %d", m.Type)
	}
	if len(m.Sum) == 0 {
		return errors.New("missing value")
	}

	c.typ = typeFromProto(m.Type)
	c.val = m.Sum

	return nil
}

// ProtoMessage converts c into message to transmit using the NeoFS API
// protocol.
//
// See also [Checksum.FromProtoMessage].
func (c Checksum) ProtoMessage() *refs.Checksum {
	return &refs.Checksum{
		Type: typeToProto(c.typ),
		Sum:  c.val,
	}
}

// Type returns checksum type.
//
// Zero Checksum has Unknown checksum type.
//
// See also [NewTillichZemor], [NewSHA256].
func (c Checksum) Type() Type {
	return c.typ
}

// Value returns checksum bytes. Return value
// MUST NOT be mutated.
//
// Zero Checksum has nil sum.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// See also [NewTillichZemor], [NewSHA256].
func (c Checksum) Value() []byte {
	return c.val
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (c Checksum) String() string {
	return fmt.Sprintf("%s:%s", c.Type(), hex.EncodeToString(c.Value()))
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (m Type) String() string {
	return typeToProto(m).String()
}
