package checksum

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// Checksum represents checksum of the NeoFS primitives.
//
// Checksum is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Checksum
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
//
// Note that direct typecast is not safe and may result in loss of compatibility:
// 	_ = Checksum(refs.Checksum{}) // not recommended
type Checksum refs.Checksum

// Type represents the enumeration
// of checksum types.
type Type uint8

const (
	// Unknown is an undefined checksum type.
	Unknown Type = iota

	// SHA256 is a SHA256 checksum type.
	SHA256

	// TZ is a Tillich-Zemor checksum type.
	TZ
)

// ReadFromV2 reads Checksum from the refs.Checksum message.
//
// See also WriteToV2.
func (c *Checksum) ReadFromV2(m refs.Checksum) {
	*c = Checksum(m)
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
// See also SetTillichZemor and SetSHA256.
func (c Checksum) Type() Type {
	v2 := (refs.Checksum)(c)
	switch v2.GetType() {
	case refs.SHA256:
		return SHA256
	case refs.TillichZemor:
		return TZ
	default:
		return Unknown
	}
}

// Sum returns checksum bytes.
//
// Zero Checksum has nil sum.
//
// See also SetTillichZemor and SetSHA256.
func (c Checksum) Sum() []byte {
	v2 := (refs.Checksum)(c)
	return v2.GetSum()
}

// SetSHA256 sets checksum to SHA256 hash.
func (c *Checksum) SetSHA256(v [sha256.Size]byte) {
	checksum := (*refs.Checksum)(c)

	checksum.SetType(refs.SHA256)
	checksum.SetSum(v[:])
}

// SetTillichZemor sets checksum to Tillich-Zemor hash.
func (c *Checksum) SetTillichZemor(v [64]byte) {
	checksum := (*refs.Checksum)(c)

	checksum.SetType(refs.TillichZemor)
	checksum.SetSum(v[:])
}

// Equal returns boolean value that means
// the equality of the passed Checksums.
func Equal(cs1, cs2 Checksum) bool {
	return cs1.Type() == cs2.Type() && bytes.Equal(cs1.Sum(), cs2.Sum())
}

// String implements fmt.Stringer interface method.
func (c Checksum) String() string {
	v2 := (refs.Checksum)(c)
	return hex.EncodeToString(v2.GetSum())
}

// Parse is a reverse action to String().
func (c *Checksum) Parse(s string) error {
	data, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	var typ refs.ChecksumType

	switch ln := len(data); ln {
	default:
		return fmt.Errorf("unsupported checksum length %d", ln)
	case sha256.Size:
		typ = refs.SHA256
	case 64:
		typ = refs.TillichZemor
	}

	cV2 := (*refs.Checksum)(c)
	cV2.SetType(typ)
	cV2.SetSum(data)

	return nil
}

// Empty returns true if it is called on
// zero checksum.
func (c Checksum) Empty() bool {
	v2 := (refs.Checksum)(c)
	return v2.GetSum() == nil
}

// String implements fmt.Stringer interface method.
func (m Type) String() string {
	var m2 refs.ChecksumType

	switch m {
	default:
		m2 = refs.UnknownChecksum
	case TZ:
		m2 = refs.TillichZemor
	case SHA256:
		m2 = refs.SHA256
	}

	return m2.String()
}

// Parse is a reverse action to String().
//
// Returns true if s was parsed successfully.
func (m *Type) Parse(s string) bool {
	var g refs.ChecksumType

	ok := g.FromString(s)

	if ok {
		switch g {
		default:
			*m = Unknown
		case refs.TillichZemor:
			*m = TZ
		case refs.SHA256:
			*m = SHA256
		}
	}

	return ok
}
