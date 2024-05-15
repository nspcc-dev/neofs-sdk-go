package version

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
)

// Version represents revision number in SemVer scheme.
//
// Version implements built-in comparable interface.
//
// Version is mutually compatible with [refs.Version] message. See
// [Version.ReadFromV2] / [Version.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type Version struct{ mjr, mnr uint32 }

const sdkMjr, sdkMnr = 2, 16

// Current is the latest NeoFS API protocol version supported by this library.
var Current = Version{sdkMjr, sdkMnr}

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

// WriteToV2 writes Version to the [refs.Version] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Version.ReadFromV2].
func (v Version) WriteToV2(m *refs.Version) {
	m.Major, m.Minor = v.mjr, v.mnr
}

// ReadFromV2 reads Version from the [refs.Version] message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Version.WriteToV2].
func (v *Version) ReadFromV2(m *refs.Version) error {
	v.mjr, v.mnr = m.Major, m.Minor
	return nil
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (v Version) String() string {
	return EncodeToString(v)
}

// EncodeToString encodes version according to format from specification:
// semver formatted value without patch and with v prefix, e.g. 'v2.1'.
func EncodeToString(v Version) string {
	return fmt.Sprintf("v%d.%d", v.Major(), v.Minor())
}
