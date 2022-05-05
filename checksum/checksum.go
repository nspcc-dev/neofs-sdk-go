package checksum

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// Checksum represents v2-compatible checksum.
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

// NewFromV2 wraps v2 Checksum message to Checksum.
//
// Nil refs.Checksum converts to nil.
func NewFromV2(cV2 *refs.Checksum) *Checksum {
	return (*Checksum)(cV2)
}

// New creates and initializes blank Checksum.
//
// Defaults:
//  - sum: nil;
//  - type: Unknown.
func New() *Checksum {
	return NewFromV2(new(refs.Checksum))
}

// Type returns checksum type.
func (c *Checksum) Type() Type {
	switch (*refs.Checksum)(c).GetType() {
	case refs.SHA256:
		return SHA256
	case refs.TillichZemor:
		return TZ
	default:
		return Unknown
	}
}

// Sum returns checksum bytes.
func (c *Checksum) Sum() []byte {
	return (*refs.Checksum)(c).GetSum()
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

// ToV2 converts Checksum to v2 Checksum message.
//
// Nil Checksum converts to nil.
func (c *Checksum) ToV2() *refs.Checksum {
	return (*refs.Checksum)(c)
}

func Equal(cs1, cs2 *Checksum) bool {
	return cs1.Type() == cs2.Type() && bytes.Equal(cs1.Sum(), cs2.Sum())
}

// Marshal marshals Checksum into a protobuf binary form.
func (c *Checksum) Marshal() ([]byte, error) {
	return (*refs.Checksum)(c).StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of Checksum.
func (c *Checksum) Unmarshal(data []byte) error {
	return (*refs.Checksum)(c).Unmarshal(data)
}

// MarshalJSON encodes Checksum to protobuf JSON format.
func (c *Checksum) MarshalJSON() ([]byte, error) {
	return (*refs.Checksum)(c).MarshalJSON()
}

// UnmarshalJSON decodes Checksum from protobuf JSON format.
func (c *Checksum) UnmarshalJSON(data []byte) error {
	return (*refs.Checksum)(c).UnmarshalJSON(data)
}

func (c *Checksum) String() string {
	return hex.EncodeToString((*refs.Checksum)(c).GetSum())
}

// Parse parses Checksum from its string representation.
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

// String returns string representation of Type.
//
// String mapping:
//  * TZ: TZ;
//  * SHA256: SHA256;
//  * Unknown, default: CHECKSUM_TYPE_UNSPECIFIED.
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

// FromString parses Type from a string representation.
// It is a reverse action to String().
//
// Returns true if s was parsed successfully.
func (m *Type) FromString(s string) bool {
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
