package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"google.golang.org/protobuf/encoding/protojson"
)

// Version represents revision number in SemVer scheme.
//
// ID implements built-in comparable interface.
//
// Version is mutually compatible with [refs.Version] message. See
// [Version.FromProtoMessage] / [Version.ProtoMessage] methods.
//
// Instances should be created using one of the constructors.
type Version struct{ mjr, mnr uint32 }

// New constructs new Version instance.
func New(mjr, mnr uint32) Version {
	var res Version
	res.SetMajor(mjr)
	res.SetMinor(mnr)
	return res
}

// UnmarshalJSON creates new Version and makes [Version.UnmarshalJSON].
func UnmarshalJSON(b []byte) (Version, error) {
	var res Version
	return res, res.UnmarshalJSON(b)
}

const sdkMjr, sdkMnr = 2, 22

// Current returns Version instance that initialized to the
// latest supported NeoFS API revision number in SDK.
func Current() (v Version) {
	return New(sdkMjr, sdkMnr)
}

// Major returns major number of the revision.
func (v Version) Major() uint32 {
	return v.mjr
}

// SetMajor sets major number of the revision.
func (v *Version) SetMajor(val uint32) {
	v.mjr = val
}

// Minor returns minor number of the revision.
func (v Version) Minor() uint32 {
	return v.mnr
}

// SetMinor sets minor number of the revision.
func (v *Version) SetMinor(val uint32) {
	v.mnr = val
}

// ProtoMessage converts v into message to transmit using the NeoFS API
// protocol.
//
// See also [Version.FromProtoMessage].
func (v Version) ProtoMessage() *refs.Version {
	return &refs.Version{
		Major: v.mjr,
		Minor: v.mnr,
	}
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// v from it.
//
// See also [Version.ProtoMessage].
func (v *Version) FromProtoMessage(m *refs.Version) error {
	v.mjr = m.Major
	v.mnr = m.Minor
	return nil
}

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (v Version) String() string {
	return EncodeToString(v)
}

// DecodeString is the inverse of [String], it parses string representation
// of [Version] and sets the value accordingly.
func (v *Version) DecodeString(s string) error {
	if len(s) < 4 { // v0.0
		return errors.New("wrong length")
	}
	if s[0] != 'v' {
		return errors.New("doesn't start with 'v'")
	}
	majS, minS, found := strings.Cut(s[1:], ".")
	if !found {
		return errors.New("dot not found")
	}
	if (len(majS) > 1 && majS[0] == '0') || (len(minS) > 1 && minS[0] == '0') {
		return errors.New("leading zero in number")
	}
	majV, err := strconv.ParseUint(majS, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid major: %w", err)
	}
	minV, err := strconv.ParseUint(minS, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid minor: %w", err)
	}
	v.mjr = uint32(majV)
	v.mnr = uint32(minV)
	return nil
}

// EncodeToString encodes version according to format from specification:
// semver formatted value without patch and with v prefix, e.g. 'v2.1'.
func EncodeToString(v Version) string {
	return fmt.Sprintf("v%d.%d", v.Major(), v.Minor())
}

// MarshalJSON encodes Version into a JSON format of the NeoFS API
// protocol (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (v Version) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(v)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Version
// (Protocol Buffers JSON). Returns an error describing a format violation.
// Use [UnmarshalJSON] to decode data into a new Version.
//
// See also MarshalJSON.
func (v *Version) UnmarshalJSON(data []byte) error {
	var m refs.Version

	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	return v.FromProtoMessage(&m)
}
