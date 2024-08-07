package version

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
)

// Version represents revision number in SemVer scheme.
//
// ID implements built-in comparable interface.
//
// Version is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/refs.Version
// message. See ReadFromV2 / WriteToV2 methods.
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

const sdkMjr, sdkMnr = 2, 16

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

// WriteToV2 writes Version to the refs.Version message.
// The message must not be nil.
//
// See also ReadFromV2.
func (v Version) WriteToV2(m *refs.Version) {
	m.SetMajor(v.mjr)
	m.SetMinor(v.mnr)
}

// ReadFromV2 reads Version from the refs.Version message. Checks if the message
// conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (v *Version) ReadFromV2(m refs.Version) error {
	v.mjr = m.GetMajor()
	v.mnr = m.GetMinor()
	return nil
}

// String implements fmt.Stringer.
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

// Equal returns true if versions are identical.
// Deprecated: Version is comparable.
func (v Version) Equal(v2 Version) bool {
	return v.Major() == v2.Major() &&
		v.Minor() == v2.Minor()
}

// MarshalJSON encodes Version into a JSON format of the NeoFS API
// protocol (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (v Version) MarshalJSON() ([]byte, error) {
	var m refs.Version
	v.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Version
// (Protocol Buffers JSON). Returns an error describing a format violation.
// Use [UnmarshalJSON] to decode data into a new Version.
//
// See also MarshalJSON.
func (v *Version) UnmarshalJSON(data []byte) error {
	var m refs.Version

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return v.ReadFromV2(m)
}
