package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/tzhash/tz"
)

// Checksum represents checksum of some digital data.
//
// Checksum is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Checksum
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances must be created using one of the constructors.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
//
//	_ = Checksum(refs.Checksum{}) // not recommended
type Checksum refs.Checksum

// Type represents the enumeration
// of checksum types.
type Type uint32

const (
	// Unknown is an undefined checksum type.
	// Deprecated: use 0 instead.
	Unknown Type = iota

	// SHA256 is a SHA256 checksum type.
	SHA256

	// TZ is a Tillich-Zémor checksum type.
	TZ
)

func typeToProto(t Type) refs.ChecksumType {
	switch t {
	default:
		return refs.ChecksumType(t)
	case SHA256:
		return refs.SHA256
	case TZ:
		return refs.TillichZemor
	}
}

// New constructs new Checksum instance. It is the caller's responsibility to
// ensure that the hash matches the type.
func New(typ Type, hsh []byte) Checksum {
	var res refs.Checksum
	res.SetType(typeToProto(typ))
	res.SetSum(hsh)
	return Checksum(res)
}

// NewSHA256 constructs new Checksum from SHA-256 hash.
func NewSHA256(h [sha256.Size]byte) Checksum {
	return New(SHA256, h[:])
}

// NewTillichZemor constructs new Checksum from Tillich-Zémor homomorphic hash.
func NewTillichZemor(h [tz.Size]byte) Checksum {
	return New(TZ, h[:])
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
	case TZ:
		return NewTillichZemor(tz.Sum(data)), nil
	}
}

// ReadFromV2 reads Checksum from the refs.Checksum message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (c *Checksum) ReadFromV2(m refs.Checksum) error {
	if len(m.GetSum()) == 0 {
		return errors.New("missing value")
	}

	*c = Checksum(m)

	return nil
}

// WriteToV2 writes Checksum to the refs.Checksum message.
// The message must not be nil.
//
// See also ReadFromV2.
func (c Checksum) WriteToV2(m *refs.Checksum) {
	*m = (refs.Checksum)(c)
}

// Type returns checksum type.
//
// Zero Checksum has Unknown checksum type.
//
// See also [NewTillichZemor], [NewSHA256].
func (c Checksum) Type() Type {
	v2 := (refs.Checksum)(c)
	switch typ := v2.GetType(); typ {
	case refs.SHA256:
		return SHA256
	case refs.TillichZemor:
		return TZ
	default:
		return Type(typ)
	}
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
	v2 := (refs.Checksum)(c)
	return v2.GetSum()
}

// SetSHA256 sets checksum to SHA256 hash.
//
// See also Calculate.
// Deprecated: use [NewSHA256] instead.
func (c *Checksum) SetSHA256(v [sha256.Size]byte) { *c = NewSHA256(v) }

// Calculate calculates checksum and sets it
// to the passed checksum. Checksum must not be nil.
//
// Does nothing if the passed type is not one of the:
//   - SHA256;
//   - TZ.
//
// Does not mutate the passed value.
//
// See also SetSHA256, SetTillichZemor.
// Deprecated: use [NewFromData] instead.
func Calculate(c *Checksum, t Type, v []byte) {
	if cs, err := NewFromData(t, v); err == nil {
		*c = cs
	}
}

// SetTillichZemor sets checksum to Tillich-Zémor hash.
//
// See also Calculate.
// Deprecated: use [NewTillichZemor] instead.
func (c *Checksum) SetTillichZemor(v [tz.Size]byte) { *c = NewTillichZemor(v) }

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (c Checksum) String() string {
	v2 := (refs.Checksum)(c)
	return fmt.Sprintf("%s:%s", c.Type(), hex.EncodeToString(v2.GetSum()))
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (m Type) String() string {
	return typeToProto(m).String()
}
